package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"mycbt/backend/internal/cache"
	"mycbt/backend/internal/config"
	"mycbt/backend/internal/db"
	"mycbt/backend/internal/httpapi"
	"mycbt/backend/internal/httpapi/middleware"
	"mycbt/backend/internal/repo/userrepo"
	"mycbt/backend/internal/service/authsvc"
	"mycbt/backend/internal/storage"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()

	var deps httpapi.Deps
	deps.Config = cfg
	deps.ObjectStore = storage.NewLocalObjectStore(cfg.UploadLocalDir)

	if redisClient := newRedisClient(ctx, cfg); redisClient != nil {
		deps.Redis = redisClient
		middleware.UseRedisRateLimiter(redisClient, cfg.RedisPrefix)
		log.Printf("redis enabled: %s", cfg.RedisAddr)
	} else {
		log.Printf("redis disabled: REDIS_ADDR not set or not reachable")
	}

	if objectStore, err := newObjectStore(ctx, cfg); err != nil {
		log.Printf("warning: minio disabled, fallback to local uploads: %v", err)
	} else if objectStore != nil {
		deps.ObjectStore = objectStore
		log.Printf("minio enabled: bucket=%s endpoint=%s", cfg.MinIOBucket, cfg.MinIOEndpoint)
	}

	if cfg.DatabaseURL != "" {
		d, err := db.Open(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("db: %v", err)
		}
		defer d.Pool.Close()

		users := userrepo.New(d.Pool)
		auth, err := authsvc.New(cfg, users)
		if err != nil {
			log.Fatalf("auth: %v", err)
		}
		if deps.Redis != nil {
			auth.SetBlocklist(cache.NewRedisTokenBlocklist(deps.Redis, cfg.RedisPrefix))
		}

		deps.Users = users
		deps.Auth = auth
		deps.Pool = d.Pool
	} else {
		log.Printf("warning: DATABASE_URL not set; auth endpoints disabled")
	}
	if deps.Redis != nil {
		defer deps.Redis.Close()
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpapi.NewHandler(deps),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("api listening on %s (env=%s)", srv.Addr, cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("shutting down...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}

func newRedisClient(ctx context.Context, cfg config.Config) *redis.Client {
	addr := strings.TrimSpace(cfg.RedisAddr)
	if addr == "" {
		return nil
	}
	dbNum := 0
	if v := strings.TrimSpace(cfg.RedisDB); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			dbNum = parsed
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     strings.TrimSpace(cfg.RedisPassword),
		DB:           dbNum,
		PoolSize:     20,
		MinIdleConns: 5,
	})
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil
	}
	return client
}

func newObjectStore(ctx context.Context, cfg config.Config) (storage.ObjectStore, error) {
	provider := strings.ToLower(strings.TrimSpace(cfg.UploadProvider))
	if provider == "" || provider == "local" {
		return nil, nil
	}
	if provider != "minio" {
		return nil, errors.New("unknown UPLOAD_PROVIDER")
	}

	useSSL := strings.EqualFold(strings.TrimSpace(cfg.MinIOUseSSL), "true") || strings.TrimSpace(cfg.MinIOUseSSL) == "1"
	return storage.NewMinIOObjectStore(ctx, storage.MinIOConfig{
		Endpoint:      cfg.MinIOEndpoint,
		AccessKey:     cfg.MinIOAccessKey,
		SecretKey:     cfg.MinIOSecretKey,
		Bucket:        cfg.MinIOBucket,
		UseSSL:        useSSL,
		PublicBaseURL: cfg.MinIOPublicBaseURL,
		KeyPrefix:     cfg.MinIOKeyPrefix,
	})
}
