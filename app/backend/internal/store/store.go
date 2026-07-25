package store 

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

const schema = `
CREATE TABLE IF NOT EXISTS patterns (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	slug text UNIQUE NOT NULL,
	n int NOT NULL,
	k int NOT NULL,
	cells jsonb NOT NULL,
	stock jsonb,
	title text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS jobs (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	kind text NOT NULL,
	status text NOT NULL DEFAULT 'queued',
	params jsonb NOT NULL,
	error text,
	created_at timestamptz NOT NULL DEFAULT now(),
	started_at timestamptz,
	finished_at timestamptz
);

CREATE TABLE IF NOT EXISTS results (
	job_id uuid PRIMARY KEY REFERENCES jobs(id) ON DELETE CASCADE,
	payload jsonb NOT NULL
);

CREATE TABLE IF NOT EXISTS benchmarks (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	mode text NOT NULL,
	n int NOT NULL,
	k int NOT NULL,
	nodes bigint NOT NULL,
	elapsed_ms double precision NOT NULL,
	verdict text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS verification_runs (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	suite text NOT NULL,
	passed int NOT NULL,
	failed int NOT NULL,
	details jsonb,
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_jobs_status_created ON jobs(status, created_at);
CREATE INDEX IF NOT EXISTS idx_benchmarks_mode_n ON benchmarks(mode, n, created_at);
`

func Open(ctx context.Context, dsn string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	s := &Store{pool: pool}
	if err := s.migrate(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return s, nil
}

const migrationLockID = 727271

func (s *Store) migrate(ctx context.Context) error {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock($1)`, int64(migrationLockID)); err != nil {
		return err
	}
	defer conn.Exec(ctx, `SELECT pg_advisory_unlock($1)`, int64(migrationLockID))

	_, err = conn.Exec(ctx, schema)
	return err
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return s.pool.Ping(ctx)
}

func OpenWithRetry(ctx context.Context, dsn string, attempts int, delay time.Duration) (*Store, error) {
	var lastErr error
	for i := 0; i < attempts; i++ {
		s, err := Open(ctx, dsn)
		if err == nil {
			return s, nil
		}
		lastErr = err
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}
	return nil, lastErr
}
