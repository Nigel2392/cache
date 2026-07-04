package cache

import "github.com/Nigel2392/cache/internal"

type CacheOperation = internal.CacheOperation

const (
	OP_UNKNOWN    = internal.OP_UNKNOWN
	OP_GET        = internal.OP_GET        //  Get(c context.Context, key string) (T, error)
	OP_GETDEFAULT = internal.OP_GETDEFAULT //  GetDefault(c context.Context, key string, defaultValue T) (T, error)
	OP_SET        = internal.OP_SET        //  Set(c context.Context, key string, value T, ttl time.Duration) error
	OP_INCR       = internal.OP_INCR       //  Increment(c context.Context, key string, amount int64) (int64, error)
	OP_DECR       = internal.OP_DECR       //  Decrement(c context.Context, key string, amount int64) (int64, error)
	OP_CVAL       = internal.OP_CVAL       //  CounterValue(c context.Context, key string) (int64, error)
	OP_EXPR       = internal.OP_EXPR       //  Expire(c context.Context, key string, ttl time.Duration) error
	OP_TTL        = internal.OP_TTL        //  TTL(c context.Context, key string) time.Duration
	OP_HAS        = internal.OP_HAS        //  Has(c context.Context, key string) bool
	OP_DEL        = internal.OP_DEL        //  Delete(c context.Context, key string) error
	OP_KEYS       = internal.OP_KEYS       //  Keys(c context.Context) ([]string, error)
	OP_CLEAR      = internal.OP_CLEAR      //  Clear(c context.Context) error
	OP_CLOSE      = internal.OP_CLOSE      //  Close(c context.Context) error
)
