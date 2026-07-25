package redisstream 

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"famcs-ui/backend/internal/queue"
)

const streamKey = "famcs:jobs"

type Queue struct {
	rdb *redis.Client
}

func New(addr string) *Queue {
	return &Queue{rdb: redis.NewClient(&redis.Options{Addr: addr})}
}

func (q *Queue) Enqueue(ctx context.Context, j queue.Job) error {
	b, err := json.Marshal(j)
	if err != nil {
		return err
	}
	return q.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		Values: map[string]any{"job": b},
	}).Err()
}

func (q *Queue) Consume(ctx context.Context, group string, fn queue.Handler) error {
	if err := q.rdb.XGroupCreateMkStream(ctx, streamKey, group, "0").Err(); err != nil {
		if !strings.Contains(err.Error(), "BUSYGROUP") {
			return err
		}
	}
	consumer := "worker-" + time.Now().Format("150405.000000")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		res, err := q.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: consumer,
			Streams:  []string{streamKey, ">"},
			Count:    1,
			Block:    2 * time.Second,
		}).Result()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			continue
		}
		for _, stream := range res {
			for _, msg := range stream.Messages {
				raw, _ := msg.Values["job"].(string)
				var j queue.Job
				if err := json.Unmarshal([]byte(raw), &j); err == nil {
					_ = fn(ctx, j)
				}
				q.rdb.XAck(ctx, streamKey, group, msg.ID)
			}
		}
	}
}

func (q *Queue) Depth(ctx context.Context) (int64, error) {
	return q.rdb.XLen(ctx, streamKey).Result()
}

func (q *Queue) Close() error {
	return q.rdb.Close()
}
