package queue

import (
	"context"
	"encoding/json"
)

type Job struct {
	ID     string          `json:"id"`
	Kind   string          `json:"kind"`
	Params json.RawMessage `json:"params"`
}

type Handler func(ctx context.Context, j Job) error

type Queue interface {
	Enqueue(ctx context.Context, j Job) error
	Consume(ctx context.Context, group string, fn Handler) error
	Depth(ctx context.Context) (int64, error)
	Close() error
}
