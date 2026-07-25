package store

import (
	"context"
	"encoding/json"
	"time"
)

type VerificationRecord struct {
	ID        string          `json:"id"`
	Suite     string          `json:"suite"`
	Passed    int             `json:"passed"`
	Failed    int             `json:"failed"`
	Details   json.RawMessage `json:"details,omitempty"`
	CreatedAt time.Time       `json:"createdAt"`
}

func (s *Store) RecordVerification(ctx context.Context, suite string, passed, failed int, details any) (string, error) {
	b, err := json.Marshal(details)
	if err != nil {
		return "", err
	}
	var id string
	err = s.pool.QueryRow(ctx,
		`INSERT INTO verification_runs (suite, passed, failed, details) VALUES ($1,$2,$3,$4) RETURNING id`,
		suite, passed, failed, b,
	).Scan(&id)
	return id, err
}

func (s *Store) ListVerifications(ctx context.Context, limit int) ([]VerificationRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx,
		`SELECT id, suite, passed, failed, details, created_at FROM verification_runs
		 ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []VerificationRecord
	for rows.Next() {
		var r VerificationRecord
		if err := rows.Scan(&r.ID, &r.Suite, &r.Passed, &r.Failed, &r.Details, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
