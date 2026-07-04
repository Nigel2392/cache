package internal

type CacheOperation uint

const (
	OP_UNKNOWN    CacheOperation = iota << 1
	OP_GET                       //  Get(c context.Context, key string) (T, error)
	OP_GETDEFAULT                //  GetDefault(c context.Context, key string, defaultValue T) (T, error)
	OP_SET                       //  Set(c context.Context, key string, value T, ttl time.Duration) error
	OP_INCR                      //  Increment(c context.Context, key string, amount int64) (int64, error)
	OP_DECR                      //  Decrement(c context.Context, key string, amount int64) (int64, error)
	OP_CVAL                      //  CounterValue(c context.Context, key string) (int64, error)
	OP_EXPR                      //  Expire(c context.Context, key string, ttl time.Duration) error
	OP_TTL                       //  TTL(c context.Context, key string) time.Duration
	OP_HAS                       //  Has(c context.Context, key string) bool
	OP_DEL                       //  Delete(c context.Context, key string) error
	OP_KEYS                      //  Keys(c context.Context) ([]string, error)
	OP_CLEAR                     //  Clear(c context.Context) error
	OP_CLOSE                     //  Close(c context.Context) error
)
