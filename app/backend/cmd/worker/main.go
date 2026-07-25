package main 

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"famcs-ui/backend/internal/queue"
	"famcs-ui/backend/internal/queue/kafka"
	"famcs-ui/backend/internal/queue/redisstream"
	"famcs-ui/backend/internal/store"
	"famcs-ui/backend/internal/verify"
)

type verifyParams struct {
	Suite  string `json:"suite"`
	Trials int    `json:"trials"`
	Seed   int64  `json:"seed"`
	MaxN   int    `json:"maxN"`
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://famcs:famcs@localhost:5432/famcs?sslmode=disable"
	}
	st, err := store.OpenWithRetry(ctx, dsn, 10, 2*time.Second)
	if err != nil {
		slog.Error("store open failed", "err", err)
		os.Exit(1)
	}
	defer st.Close()

	q := openQueue()
	defer q.Close()

	slog.Info("worker started")
	err = q.Consume(ctx, "workers", func(ctx context.Context, j queue.Job) error {
		return handle(ctx, st, j)
	})
	if err != nil && ctx.Err() == nil {
		slog.Error("consume stopped", "err", err)
	}
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

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func handle(ctx context.Context, st *store.Store, j queue.Job) error {
	slog.Info("job received", "id", j.ID, "kind", j.Kind)
	_ = st.SetJobStatus(ctx, j.ID, "running")

	switch j.Kind {
	case "verify":
		var p verifyParams
		_ = json.Unmarshal(j.Params, &p)
		if p.Trials <= 0 {
			p.Trials = 500
		}
		var result verify.Result
		if p.Suite == "matching" {
			if p.MaxN <= 0 {
				p.MaxN = 3
			}
			result = verify.RunMatchingCrossCheck(p.MaxN)
		} else {
			result = verify.RunStressSuite(ctx, p.Trials, p.Seed)
		}
		_, err := st.RecordVerification(ctx, p.Suite, result.Passed, result.Failed, result)
		if err != nil {
			_ = st.SetJobError(ctx, j.ID, err.Error())
			return err
		}
		if err := st.SaveResult(ctx, j.ID, result); err != nil {
			_ = st.SetJobError(ctx, j.ID, err.Error())
			return err
		}
	default:
		_ = st.SetJobError(ctx, j.ID, "unknown job kind: "+j.Kind)
		return nil
	}

	_ = st.SetJobStatus(ctx, j.ID, "done")
	slog.Info("job done", "id", j.ID)
	return nil
}
