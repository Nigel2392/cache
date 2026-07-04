package buffer

import (
	"context"
	"errors"
	"time"

	"github.com/Nigel2392/cache/internal"
)

var _ internal.TypedTransaction[any] = (*CacheBuffer[any])(nil)

// CacheBufferItem tracks the state and operation of a single item in the buffer.
type CacheBufferItem[T any] struct {
	Op       internal.CacheOperation // Tracks OP_SET, OP_DEL, etc.
	Value    T
	LifeTime time.Time
	TTL      time.Duration
}

// CacheBuffer acts as a staging area for cache operations.
// It fully implements internal.TypedTransaction[T].
type CacheBuffer[T any] struct {
	parent        internal.TypedCache[T]
	pending       map[string]CacheBufferItem[T] // Unified state tracker
	cleared       bool
	inTransaction bool
}

func NewCacheBuffer[T any](parent internal.TypedCache[T]) *CacheBuffer[T] {
	return &CacheBuffer[T]{
		parent:        parent,
		pending:       make(map[string]CacheBufferItem[T]),
		inTransaction: true,
	}
}

func (b *CacheBuffer[T]) InTransaction() bool {
	return b.inTransaction
}

func (b *CacheBuffer[T]) Get(ctx context.Context, key string) (T, error) {
	if item, ok := b.pending[key]; ok {
		if item.Op == internal.OP_DEL {
			return *new(T), internal.ErrItemNotFound
		}
		if item.LifeTime.Before(time.Now()) {
			return *new(T), internal.ErrItemNotFound
		}
		return item.Value, nil
	}

	if b.cleared {
		return *new(T), internal.ErrItemNotFound
	}

	return b.parent.Get(ctx, key)
}

func (b *CacheBuffer[T]) GetDefault(ctx context.Context, key string, defaultValue T) (T, error) {
	val, err := b.Get(ctx, key)
	if errors.Is(err, internal.ErrItemNotFound) {
		return defaultValue, nil
	}
	return val, err
}

func (b *CacheBuffer[T]) Set(ctx context.Context, key string, value T, ttl time.Duration) error {
	ttl = internal.GetDefaultTTL(ttl)
	b.pending[key] = CacheBufferItem[T]{
		Op:       internal.OP_SET,
		Value:    value,
		TTL:      ttl,
		LifeTime: time.Now().Add(ttl),
	}
	return nil
}

func (b *CacheBuffer[T]) Delete(ctx context.Context, key string) error {
	b.pending[key] = CacheBufferItem[T]{
		Op: internal.OP_DEL,
	}
	return nil
}

func (b *CacheBuffer[T]) Increment(ctx context.Context, key string, amount int64) (int64, error) {
	var currentVal int64
	var currentTTL = internal.Infinity

	val, err := b.Get(ctx, key)
	if err == nil {
		v, ok := any(val).(int64)
		if !ok {
			return 0, internal.ErrInvalidType
		}
		currentVal = v
		currentTTL = b.TTL(ctx, key)
	} else if !errors.Is(err, internal.ErrItemNotFound) {
		return 0, err
	}

	newVal := currentVal + amount
	if setErr := b.Set(ctx, key, any(newVal).(T), currentTTL); setErr != nil {
		return 0, setErr
	}

	return newVal, nil
}

func (b *CacheBuffer[T]) Decrement(ctx context.Context, key string, amount int64) (int64, error) {
	return b.Increment(ctx, key, -amount)
}

func (b *CacheBuffer[T]) CounterValue(ctx context.Context, key string) (int64, error) {
	val, err := b.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	v, ok := any(val).(int64)
	if !ok {
		return 0, internal.ErrInvalidType
	}
	return v, nil
}

func (b *CacheBuffer[T]) Expire(ctx context.Context, key string, ttl time.Duration) error {
	val, err := b.Get(ctx, key)
	if err != nil {
		return err
	}
	return b.Set(ctx, key, val, ttl) // Resolves as an OP_SET
}

func (b *CacheBuffer[T]) TTL(ctx context.Context, key string) time.Duration {
	if item, ok := b.pending[key]; ok {
		if item.Op == internal.OP_DEL || item.LifeTime.Before(time.Now()) {
			return 0
		}
		return time.Until(item.LifeTime)
	}

	if b.cleared {
		return 0
	}

	return b.parent.TTL(ctx, key)
}

func (b *CacheBuffer[T]) Has(ctx context.Context, key string) bool {
	if item, ok := b.pending[key]; ok {
		return item.Op != internal.OP_DEL && !item.LifeTime.Before(time.Now())
	}

	if b.cleared {
		return false
	}

	return b.parent.Has(ctx, key)
}

func (b *CacheBuffer[T]) Keys(ctx context.Context) ([]string, error) {
	keySet := make(map[string]struct{})

	if !b.cleared {
		parentKeys, err := b.parent.Keys(ctx)
		if err != nil {
			return nil, err
		}
		for _, k := range parentKeys {
			if item, exists := b.pending[k]; !exists || item.Op != internal.OP_DEL {
				keySet[k] = struct{}{}
			}
		}
	}

	for k, item := range b.pending {
		if item.Op == internal.OP_SET && !item.LifeTime.Before(time.Now()) {
			keySet[k] = struct{}{}
		}
	}

	result := make([]string, 0, len(keySet))
	for k := range keySet {
		result = append(result, k)
	}
	return result, nil
}

func (b *CacheBuffer[T]) Clear(ctx context.Context) error {
	b.cleared = true
	b.pending = make(map[string]CacheBufferItem[T])
	return nil
}

func (b *CacheBuffer[T]) Close(ctx context.Context) error {
	b.inTransaction = false
	return nil
}

// Flush returns the tracked operations to be applied by the parent.
func (b *CacheBuffer[T]) Flush() (map[string]CacheBufferItem[T], bool) {
	b.inTransaction = false
	return b.pending, b.cleared
}
