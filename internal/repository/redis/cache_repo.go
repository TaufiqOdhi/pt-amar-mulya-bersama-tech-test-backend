package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
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

	if err := r.client.Set(ctx, key, data, expiration).Err(); err != nil {
		return err
	}

	// Track key in user-specific key set for fast O(1) invalidation
	if userIDStr := extractUserIDFromKey(key); userIDStr != "" {
		setKey := fmt.Sprintf("user_keys:%s", userIDStr)
		_ = r.client.SAdd(ctx, setKey, key).Err()
		_ = r.client.Expire(ctx, setKey, expiration*2).Err()
	}

	return nil
}

func extractUserIDFromKey(key string) string {
	parts := strings.Split(key, ":")
	if len(parts) >= 2 && parts[0] == "user" && parts[1] != "" {
		return parts[1]
	}
	return ""
}

func (r *cacheRepo) DeletePattern(ctx context.Context, pattern string) error {
	if r.client == nil {
		return nil
	}

	if userIDStr := extractUserIDFromKey(pattern); userIDStr != "" {
		setKey := fmt.Sprintf("user_keys:%s", userIDStr)
		keys, err := r.client.SMembers(ctx, setKey).Result()
		if err == nil && len(keys) > 0 {
			keysToDelete := append(keys, setKey)
			slog.Debug("Invalidating user cache via key set", "count", len(keys), "userID", userIDStr)
			return r.client.Del(ctx, keysToDelete...).Err()
		}
	}

	// Fallback: SCAN with count hint = 100 for efficient iteration
	var cursor uint64
	var keys []string
	for {
		var scanned []string
		var err error
		scanned, cursor, err = r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan keys for deletion: %w", err)
		}
		keys = append(keys, scanned...)
		if cursor == 0 {
			break
		}
	}

	if len(keys) > 0 {
		slog.Debug("Invalidating cache keys via scan", "count", len(keys), "pattern", pattern)
		// Delete in batches of 500
		for i := 0; i < len(keys); i += 500 {
			end := i + 500
			if end > len(keys) {
				end = len(keys)
			}
			if err := r.client.Del(ctx, keys[i:end]...).Err(); err != nil {
				return err
			}
		}
	}

	return nil
}
