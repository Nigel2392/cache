package redis

import (
	"context"
	"errors"
	"time"

	"github.com/Nigel2392/cache"
	"github.com/Nigel2392/cache/buffer"
	"github.com/Nigel2392/cache/internal"
	"github.com/redis/go-redis/v9"
)

var _ cache.TransactionalCache = (*Cache)(nil)

// Cache wraps a go-redis UniversalClient.
// Using redis.UniversalClient allows the user to pass in a standard *redis.Client,
// a *redis.ClusterClient, or a *redis.Ring depending on their infrastructure.
type Cache struct {
	Client redis.UniversalClient
}

func (c *Cache) Connect() error {
	if c.Client == nil {
		return errors.New("redis client is not initialized")
	}
	return c.Client.Ping(context.Background()).Err()
}

func (c *Cache) HasConnected() bool {
	if c.Client == nil {
		return false
	}
	err := c.Client.Ping(context.Background()).Err()
	return err == nil
}

// Get retrieves a value from the cache.
//
// If the key does not exist, Get returns nil and ErrItemNotFound.
func (c *Cache) Get(ctx context.Context, key string) (interface{}, error) {
	val, err := c.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, cache.ErrItemNotFound
	}
	if err != nil {
		return nil, err
	}
	return val, nil
}

// GetDefault retrieves a value from the cache.
//
// If the key does not exist, GetDefault returns the defaultValue.
//
// It may return an error if the key exists but the cache itself returns an error.
func (c *Cache) GetDefault(ctx context.Context, key string, defaultValue interface{}) (interface{}, error) {
	val, err := c.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return defaultValue, nil
	}
	if err != nil {
		return nil, err
	}
	return val, nil
}

// Set sets a value in the cache.
//
// The value is stored in the cache with the specified key.
// The value will expire after the specified ttl.
//
// If the TTL is 0, or Infinity, the value will never expire.
func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.Client.Set(ctx, key, value, time.Duration(ttl)).Err()
}

// Increment atomically increments a numeric key by the given amount.
// If the key does not exist, it initializes it to the amount with an infinite TTL.
// It does NOT reset the TTL of an existing key.
func (c *Cache) Increment(ctx context.Context, key string, amount int64) (int64, error) {
	return c.Client.IncrBy(ctx, key, amount).Result()
}

// Decrement atomically decrements a numeric key by the given amount.
// If the key does not exist, it initializes it to -amount with an infinite TTL.
// It does NOT reset the TTL of an existing key.
func (c *Cache) Decrement(ctx context.Context, key string, amount int64) (int64, error) {
	return c.Client.DecrBy(ctx, key, amount).Result()
}

// CounterValue retrieves the counter for the specified key.
// If the key does not exist in the cache, [ErrItemNotFound] is returned.
func (c *Cache) CounterValue(ctx context.Context, key string) (int64, error) {
	val, err := c.Client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, cache.ErrItemNotFound
	}
	return val, err
}

// Expire sets the TTL for a given key.
// If the key does not exist in the cache, [ErrItemNotFound] is returned.
func (c *Cache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	ok, err := c.Client.Expire(ctx, key, time.Duration(ttl)).Result()
	if err != nil {
		return err
	}
	if !ok {
		return cache.ErrItemNotFound
	}
	return nil
}

// TTL returns the time to live for a key.
//
// If the key does not exist, TTL returns 0.
//
// If any error occurs, TTL returns 0.
func (c *Cache) TTL(ctx context.Context, key string) time.Duration {
	duration, err := c.Client.TTL(ctx, key).Result()
	// Redis returns -2 for non-existent keys and -1 for keys with no expiry.
	if err != nil || duration < 0 {
		return 0
	}
	return duration
}

// Has returns true if the key exists in the cache.
//
// If any error occurs, Has returns false.
func (c *Cache) Has(ctx context.Context, key string) bool {
	count, err := c.Client.Exists(ctx, key).Result()
	return err == nil && count > 0
}

// Delete removes a key from the cache.
//
// If the key does not exist, Delete should return ErrItemNotFound.
func (c *Cache) Delete(ctx context.Context, key string) error {
	deleted, err := c.Client.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	if deleted == 0 {
		return cache.ErrItemNotFound
	}
	return nil
}

// Keys returns all keys in the cache.
//
// If any error occurs, Keys returns an empty slice and the error.
func (c *Cache) Keys(ctx context.Context) ([]string, error) {
	// Note: In large production environments, Keys("*") can block Redis.
	// We use it here to satisfy the exact interface requirement cleanly.
	return c.Client.Keys(ctx, "*").Result()
}

// Clear removes all keys from the cache.
//
// If any error occurs, Clear should return the error.
func (c *Cache) Clear(ctx context.Context) error {
	return c.Client.FlushDB(ctx).Err()
}

// Close closes the cache.
//
// If any error occurs, Close should return the error.
func (c *Cache) Close(ctx context.Context) error {
	if c.Client != nil {
		return c.Client.Close()
	}
	return nil
}

func (c *Cache) RunInTx(ctx context.Context, fn func(ctx context.Context, txCache internal.TypedTransaction[any]) error) error {
	buf := buffer.NewCacheBuffer[any](c)

	if err := fn(ctx, buf); err != nil {
		return err // Rollback
	}

	pendingOps, cleared := buf.Flush()

	_, err := c.Client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		if cleared {
			pipe.FlushDB(ctx)
		}

		for key, item := range pendingOps {
			switch item.Op {
			case cache.OP_SET:
				pipe.Set(ctx, key, item.Value, item.TTL)
			case cache.OP_DEL:
				pipe.Del(ctx, key)
			}
		}
		return nil
	})

	return err
}
