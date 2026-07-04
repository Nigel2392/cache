# cache

A flexible, thread-safe, and generic caching utility package for Go. Designed to be a drop-in dependency for multiple repositories, it features a robust in-memory implementation, time-to-live (TTL) expiration, atomic counters, and powerful **transactional memory** support with rollback capabilities.

## Features

* **Generics Support:** Type-safe caches using `TypedCache[T]` to avoid endless type assertions.
* **Transactions:** Execute multiple operations in isolation using `RunInTx`. If the function returns an error (or the context cancels), all changes are safely discarded.
* **Context Propagation:** Inject transactions directly into your `context.Context`. Global wrapper functions (`cache.Get`, `cache.Set`, etc.) will automatically route operations to the active transaction.
* **Atomic Counters:** Built-in `Increment` and `Decrement` operations for numeric values, safely preserving TTLs.
* **Extensible Interfaces:** Easily build custom backends (e.g., Redis, Memcached) by implementing the `TransactionalCache` or `Cache` interfaces.
* **In-Memory Engine:** A concurrent `MemoryCache` with an automatic background cleanup worker for expired items.
  * Transactions support for this cache is limited to development or low traffic environments.

---

## Installation

```bash
go get github.com/Nigel2392/cache

```

---

## Usage

### Basic Global Cache

The package initializes a default in-memory cache automatically. You can use the global wrapper functions right out of the box.

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/Nigel2392/cache"
)

func main() {
    ctx := context.Background()
    
    // Set a value with a 5-minute TTL
    err := cache.Set(ctx, "session_id", "user-123", 5*time.Minute)
    if err != nil {
        panic(err)
    }
    
    // Retrieve the value
    val, err := cache.Get(ctx, "session_id")
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Session ID: %s\n", val)
}
```

### Generic Type-Safe Cache

Avoid `interface{}` by instantiating a generic cache for specific data structures.

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/Nigel2392/cache"
)

type User struct {
    ID   int
    Name string
}

func main() {
    ctx := context.Background()
    
    // Create a cache exclusively for User structs
    userCache := cache.NewGenericMemoryCache[User]()
    userCache.Run(1 * time.Second) // Start the background cleanup worker
    
    // Type-safe set
    userCache.Set(ctx, "user:1", User{ID: 1, Name: "Alice"}, 1*time.Hour)
    
    // Type-safe get (no casting required)
    user, err := userCache.Get(ctx, "user:1")
    if err == nil {
        fmt.Printf("Found user: %s\n", user.Name)
    }
}
```

### Transactions & Rollbacks

Transactions allow you to group cache modifications. If the callback returns an error, the state reverts, leaving the main cache completely untouched.

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "github.com/Nigel2392/cache"
)

func main() {
    ctx := context.Background()
    
    err := cache.RunInTx(ctx, func(ctx context.Context, tx cache.Transaction) error {
        // Read a value (isolated)
        tx.Set(ctx, "temp_key", "temporary_value", 0)
        
        // Delete a value
        tx.Delete(ctx, "existing_key")
        
        // Something went wrong!
        return errors.New("abort transaction")
    })
    
    if err != nil {
        fmt.Println("Transaction rolled back!")
    }
    
    // "temp_key" does not exist in the main cache, 
    // and "existing_key" was never actually deleted.
}

```

### Context Propagation

You can inject a transaction into a context. If you pass this context to the global cache functions, they will automatically use the transaction state instead of the global cache.

```go
func processRequest(ctx context.Context) error {
    return cache.RunInTx(ctx, func(ctx context.Context, tx cache.Transaction) error {
        // Bind the transaction to the context
        txCtx := cache.ContextWithTransaction(ctx, tx)
        
        // This calls the GLOBAL cache.Set, but because txCtx is passed, 
        // it is secretly routed to the transaction.
        cache.Set(txCtx, "user_balance", 100, 0)
        
        return nil // Commits the transaction to the main cache
    })
}

```

### Atomic Counters

Counters drop in strictly as `int64` and allow thread-safe math operations without resetting an item's TTL.

```go
ctx := context.Background()

// Initialize or increment a counter
newVal, _ := cache.Increment(ctx, "page_views", 1)
fmt.Println("Views:", newVal)

// Decrement
newVal, _ = cache.Decrement(ctx, "inventory", 5)

```

---

## Core Interfaces

If you want to implement your own backend (like Redis), implement these interfaces:

### `TypedCache[T any]`

The base interface for standard cache operations (`Get`, `Set`, `Delete`, `Increment`, `TTL`, `Has`, `Clear`, `Keys`).

### `TypedTransactionalCache[T any]`

Embeds `TypedCache[T]` and adds the `RunInTx` method for executing isolated changes.

### `TypedTransaction[T any]`

Represents the state *during* a transaction. Embeds `TypedCache[T]` and adds `InTransaction() bool`.
