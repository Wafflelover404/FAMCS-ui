package store 

import (
	"context"
	"time"
)

type BenchmarkRecord struct {
	ID        string    `json:"id"`
	Mode      string    `json:"mode"`
	N         int       `json:"n"`
	K         int       `json:"k"`
	Nodes     int64     `json:"nodes"`
	ElapsedMs float64   `json:"elapsedMs"`
	Verdict   string    `json:"verdict"`
	CreatedAt time.Time `json:"createdAt"`
}

func (s *Store) RecordBenchmark(ctx context.Context, mode string, n, k int, nodes int64, elapsedMs float64, verdict string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO benchmarks (mode, n, k, nodes, elapsed_ms, verdict) VALUES ($1,$2,$3,$4,$5,$6)`,
		mode, n, k, nodes, elapsedMs, verdict)
	return err
}

func (s *Store) ListBenchmarks(ctx context.Context, mode string, limit int) ([]BenchmarkRecord, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := s.pool.Query(ctx,
		`SELECT id, mode, n, k, nodes, elapsed_ms, verdict, created_at FROM benchmarks
		 WHERE ($1 = '' OR mode = $1) ORDER BY created_at DESC LIMIT $2`, mode, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []BenchmarkRecord
	for rows.Next() {
		var r BenchmarkRecord
		if err := rows.Scan(&r.ID, &r.Mode, &r.N, &r.K, &r.Nodes, &r.ElapsedMs, &r.Verdict, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
