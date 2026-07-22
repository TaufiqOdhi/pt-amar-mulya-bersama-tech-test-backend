package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"todo-backend/internal/domain"

	"github.com/redis/go-redis/v9"
)

type cacheRepo struct {
	client *redis.Client
}

func NewCacheRepo(client *redis.Client) domain.CacheRepository {
	return &cacheRepo{client: client}
}

func (r *cacheRepo) Get(ctx context.Context, key string, dest interface{}) error {
	if r.client == nil {
		return fmt.Errorf("redis client is nil")
	}

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

func (r *cacheRepo) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if r.client == nil {
		return nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	return r.client.Set(ctx, key, data, expiration).Err()
}

func (r *cacheRepo) DeletePattern(ctx context.Context, pattern string) error {
	if r.client == nil {
		return nil
	}

	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys for deletion: %w", err)
	}

	if len(keys) > 0 {
		slog.Debug("Invalidating cache keys", "count", len(keys), "pattern", pattern)
		return r.client.Del(ctx, keys...).Err()
	}

	return nil
}
