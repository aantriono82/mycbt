package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mycbt/backend/internal/config"
	"mycbt/backend/internal/db"
	"mycbt/backend/internal/httpapi"
	"mycbt/backend/internal/repo/userrepo"
	"mycbt/backend/internal/service/authsvc"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()

	var deps httpapi.Deps
	deps.Config = cfg

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

		deps.Users = users
		deps.Auth = auth
		deps.Pool = d.Pool
	} else {
		log.Printf("warning: DATABASE_URL not set; auth endpoints disabled")
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
