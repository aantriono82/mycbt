package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func getenvInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

func percentile(sortedVals []int64, p float64) int64 {
	if len(sortedVals) == 0 {
		return 0
	}
	if p <= 0 {
		return sortedVals[0]
	}
	if p >= 100 {
		return sortedVals[len(sortedVals)-1]
	}
	idx := int((p / 100.0) * float64(len(sortedVals)-1))
	return sortedVals[idx]
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Println("DATABASE_URL is required")
		os.Exit(1)
	}

	concurrency := getenvInt("BENCH_CONCURRENCY", 100)
	totalRequests := getenvInt("BENCH_REQUESTS", 10000)
	timeoutMs := getenvInt("BENCH_TIMEOUT_MS", 3000)

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		fmt.Printf("parse config error: %v\n", err)
		os.Exit(1)
	}
	maxConns := int32(getenvInt("DB_MAX_CONNS", 32))
	minConns := int32(getenvInt("DB_MIN_CONNS", 4))
	if minConns > maxConns {
		minConns = maxConns
	}
	cfg.MaxConns = maxConns
	cfg.MinConns = minConns
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = 30 * time.Second

	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		fmt.Printf("open pool error: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	var sessionID string
	if err := pool.QueryRow(ctx, "SELECT id::text FROM exam_sessions LIMIT 1").Scan(&sessionID); err != nil {
		fmt.Printf("load sample session_id error: %v\n", err)
		os.Exit(1)
	}

	latencies := make([]int64, 0, totalRequests)
	var latMu sync.Mutex
	var seq int64
	var failures int64

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for {
				n := atomic.AddInt64(&seq, 1)
				if int(n) > totalRequests {
					return
				}

				reqStart := time.Now()
				qctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
				var c int
				err := pool.QueryRow(qctx, "SELECT count(*) FROM exam_attempts WHERE exam_session_id=$1", sessionID).Scan(&c)
				cancel()
				elapsedUs := time.Since(reqStart).Microseconds()

				latMu.Lock()
				latencies = append(latencies, elapsedUs)
				latMu.Unlock()

				if err != nil {
					atomic.AddInt64(&failures, 1)
				}
			}
		}()
	}
	wg.Wait()
	totalElapsed := time.Since(start)

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	var sumUs int64
	for _, v := range latencies {
		sumUs += v
	}
	avgMs := 0.0
	if len(latencies) > 0 {
		avgMs = float64(sumUs) / float64(len(latencies)) / 1000.0
	}

	throughput := float64(len(latencies)) / totalElapsed.Seconds()
	p50 := float64(percentile(latencies, 50)) / 1000.0
	p95 := float64(percentile(latencies, 95)) / 1000.0
	p99 := float64(percentile(latencies, 99)) / 1000.0

	fmt.Printf("pool_max_conns=%d pool_min_conns=%d\n", maxConns, minConns)
	fmt.Printf("concurrency=%d requests=%d failures=%d\n", concurrency, len(latencies), failures)
	fmt.Printf("duration=%.3fs throughput=%.2f req/s\n", totalElapsed.Seconds(), throughput)
	fmt.Printf("latency_ms avg=%.3f p50=%.3f p95=%.3f p99=%.3f\n", avgMs, p50, p95, p99)
}

