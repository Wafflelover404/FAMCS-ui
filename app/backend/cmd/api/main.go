package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"famcs-ui/backend/internal/api"
	"famcs-ui/backend/internal/cache"
	"famcs-ui/backend/internal/queue"
	"famcs-ui/backend/internal/queue/kafka"
	"famcs-ui/backend/internal/queue/redisstream"
	"famcs-ui/backend/internal/store"
)

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func openQueue() queue.Queue {
	driver := os.Getenv("QUEUE_DRIVER")
	switch driver {
	case "kafka":
		brokers := strings.Split(getenv("KAFKA_BROKERS", "localhost:9092"), ",")
		return kafka.New(brokers)
	default:
		return redisstream.New(getenv("REDIS_ADDR", "localhost:6379"))
	}
}

func main() {
	addr := getenv("API_ADDR", ":8080")
	ctx := context.Background()

	deps := api.Deps{}

	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		st, err := store.OpenWithRetry(ctx, dsn, 10, 2*time.Second)
		if err != nil {
			slog.Warn("postgres unavailable, running without persistence", "err", err)
		} else {
			deps.Store = st
			defer st.Close()
			slog.Info("postgres connected")
		}
	}

	if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" {
		c := cache.Open(redisAddr)
		if err := c.Ping(ctx); err != nil {
			slog.Warn("redis unavailable, running without cache", "err", err)
		} else {
			deps.Cache = c
			defer c.Close()
			slog.Info("redis connected")
		}
	}

	if deps.Store != nil {
		deps.Queue = openQueue()
		slog.Info("queue configured", "driver", getenv("QUEUE_DRIVER", "redisstream"))
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      api.NewRouter(deps),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	slog.Info("listening", "addr", addr)
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
