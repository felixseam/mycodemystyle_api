package tarantool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/tarantool/go-tarantool/v2"
)

const spaceName = "cache_entries"

var ErrInvalidCacheEntry = errors.New("invalid cache entry")

type Cache struct {
	conn *tarantool.Connection
}

func New(conn *tarantool.Connection) *Cache {
	return &Cache{conn: conn}
}

// Set stores a value. A non-positive TTL means no expiration.
func (c *Cache) Set(
	ctx context.Context,
	key string,
	value any,
	ttl time.Duration,
) error {
	if key == "" {
		return errors.New("cache key cannot be empty")
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal cache value: %w", err)
	}

	var expiresAt uint64
	if ttl > 0 {
		expiresAt = uint64(time.Now().Add(ttl).Unix())
	}

	_, err = c.conn.Do(
		tarantool.NewReplaceRequest(spaceName).
			Tuple([]interface{}{
				key,
				string(data),
				expiresAt,
			}),
	).Get()
	if err != nil {
		return fmt.Errorf("set cache key %q: %w", key, err)
	}

	return nil
}

// Get returns false for a missing or expired entry.
func (c *Cache) Get(
	ctx context.Context,
	key string,
	dest any,
) (bool, error) {
	result, err := c.conn.Do(
		tarantool.NewSelectRequest(spaceName).
			Key([]interface{}{key}),
	).Get()
	if err != nil {
		return false, fmt.Errorf("get cache key %q: %w", key, err)
	}

	if len(result) == 0 {
		return false, nil
	}

	row, ok := result[0].([]interface{})
	if !ok || len(row) < 3 {
		return false, ErrInvalidCacheEntry
	}

	value, ok := row[1].(string)
	if !ok {
		return false, ErrInvalidCacheEntry
	}

	expiresAt, err := toInt64(row[2])
	if err != nil {
		return false, ErrInvalidCacheEntry
	}

	if expiresAt > 0 && time.Now().Unix() >= expiresAt {
		if err := c.Delete(ctx, key); err != nil {
			return false, err
		}
		return false, nil
	}

	if err := json.Unmarshal([]byte(value), dest); err != nil {
		return false, fmt.Errorf("decode cache key %q: %w", key, err)
	}

	return true, nil
}

func (c *Cache) Delete(
	ctx context.Context,
	key string,
) error {
	_, err := c.conn.Do(
		tarantool.NewDeleteRequest(spaceName).
			Key([]interface{}{key}),
	).Get()
	if err != nil {
		return fmt.Errorf("delete cache key %q: %w", key, err)
	}

	return nil
}
