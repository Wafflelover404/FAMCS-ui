package store 

import (
	"context"
	"encoding/json"
	"time"
)

type JobRecord struct {
	ID         string          `json:"id"`
	Kind       string          `json:"kind"`
	Status     string          `json:"status"`
	Params     json.RawMessage `json:"params"`
	Error      *string         `json:"error,omitempty"`
	CreatedAt  time.Time       `json:"createdAt"`
	StartedAt  *time.Time      `json:"startedAt,omitempty"`
	FinishedAt *time.Time      `json:"finishedAt,omitempty"`
	Result     json.RawMessage `json:"result,omitempty"`
}

func (s *Store) CreateJob(ctx context.Context, kind string, params any) (string, error) {
	b, err := json.Marshal(params)
	if err != nil {
		return "", err
	}
	var id string
	err = s.pool.QueryRow(ctx, `INSERT INTO jobs (kind, params) VALUES ($1,$2) RETURNING id`, kind, b).Scan(&id)
	return id, err
}

func (s *Store) SetJobStatus(ctx context.Context, id, status string) error {
	switch status {
	case "running":
		_, err := s.pool.Exec(ctx, `UPDATE jobs SET status=$1, started_at=now() WHERE id=$2`, status, id)
		return err
	case "done":
		_, err := s.pool.Exec(ctx, `UPDATE jobs SET status=$1, finished_at=now() WHERE id=$2`, status, id)
		return err
	default:
		_, err := s.pool.Exec(ctx, `UPDATE jobs SET status=$1 WHERE id=$2`, status, id)
		return err
	}
}

func (s *Store) SetJobError(ctx context.Context, id, msg string) error {
	_, err := s.pool.Exec(ctx, `UPDATE jobs SET status='failed', error=$1, finished_at=now() WHERE id=$2`, msg, id)
	return err
}

func (s *Store) SaveResult(ctx context.Context, jobID string, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx,
		`INSERT INTO results (job_id, payload) VALUES ($1,$2)
		 ON CONFLICT (job_id) DO UPDATE SET payload=$2`, jobID, b)
	return err
}

func (s *Store) GetJob(ctx context.Context, id string) (*JobRecord, error) {
	var r JobRecord
	var payload []byte
	err := s.pool.QueryRow(ctx, `
		SELECT j.id, j.kind, j.status, j.params, j.error, j.created_at, j.started_at, j.finished_at, res.payload
		FROM jobs j LEFT JOIN results res ON res.job_id = j.id
		WHERE j.id = $1`, id,
	).Scan(&r.ID, &r.Kind, &r.Status, &r.Params, &r.Error, &r.CreatedAt, &r.StartedAt, &r.FinishedAt, &payload)
	if err != nil {
		return nil, err
	}
	if payload != nil {
		r.Result = payload
	}
	return &r, nil
}
