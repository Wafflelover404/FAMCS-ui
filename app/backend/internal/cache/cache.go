package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	AnalyzeTTL = time.Hour
	SolveTTL   = 24 * time.Hour
)

type Cache struct {
	rdb *redis.Client
}

func Open(addr string) *Cache {
	return &Cache{rdb: redis.NewClient(&redis.Options{Addr: addr})}
}

func (c *Cache) Client() *redis.Client {
	return c.rdb
}

func (c *Cache) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return c.rdb.Ping(ctx).Err()
}

func (c *Cache) Close() error {
	return c.rdb.Close()
}

func Key(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		h.Write([]byte(p))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))[:24]
}

func Get[T any](ctx context.Context, c *Cache, key string) (T, bool) {
	var zero T
	raw, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return zero, false
	}
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		return zero, false
	}
	return v, true
}

func Set(ctx context.Context, c *Cache, key string, v any, ttl time.Duration) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, b, ttl).Err()
}
