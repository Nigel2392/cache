package cache

import "github.com/Nigel2392/cache/internal"

type (
	CacheConnector                 = internal.CacheConnector
	TypedCache[T any]              = internal.TypedCache[T]
	TypedTransactionalCache[T any] = internal.TypedTransactionalCache[T]
	TypedTransaction[T any]        = internal.TypedTransaction[T]

	// Predefined TypedCaches with interface{} as their Type.
	Cache              = internal.Cache
	TransactionalCache = internal.TransactionalCache
	Transaction        = internal.Transaction
)
