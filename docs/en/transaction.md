# Transaction model

Goal: **Service code does not depend on `*gorm.DB`, Repository code is unaware of the transaction**. Transactions flow through `context.Context`.

## 1. Three core abstractions

```go
// Hands out the *gorm.DB bound to ctx (transaction-aware).
type DBProvider interface {
    DB(ctx context.Context) *gorm.DB
}

// Opens a transaction on top of ctx.
type Transactor interface {
    Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// Default implementation that satisfies both interfaces.
type Client struct{ /* ... */ }
```

| Role | Depends on | Why |
| :--- | :--- | :--- |
| Repository | `DBProvider` | only needs a connection, not transaction boundaries |
| Service | `Transactor` | owns transaction boundaries |

> 💡 Since `*db.Client` implements both, you can wire the same instance into your DI container.

## 2. Standard usage

```go
type UserService struct {
    users repo.Repository[model.User]
    tx    db.Transactor
}

func NewUserService(client *db.Client) *UserService {
    return &UserService{
        users: repo.New[model.User](client),
        tx:    client,
    }
}

func (s *UserService) Register(ctx context.Context, u *model.User, p *model.Profile) error {
    return s.tx.Transaction(ctx, func(txCtx context.Context) error {
        if err := s.users.Create(txCtx, u); err != nil {
            return err
        }
        // Same txCtx → same transaction
        return s.profiles.Create(txCtx, p)
    })
}
```

## 3. Propagation contract

`db.Client` uses a **private** ctx key, so external code cannot fabricate a transaction context:

- You must enter via `Transactor.Transaction(...)`.
- The `txCtx` passed to `fn` carries the transaction; every `DBProvider.DB(txCtx)` / `Repository.*(txCtx, ...)` downstream reuses the same `*gorm.DB`.
- `fn` returns `error` → GORM rolls back; returns `nil` → commits.
- If a ctx **does not** carry the transaction key, `DB(ctx)` falls back to `db.WithContext(ctx)`—plain non-transactional access.

## 4. Nested transactions

`Transaction` builds on GORM's transaction support: calling `Transaction(...)` again with an outer `txCtx` opens a SAVEPOINT-based nested transaction. Failure inside the inner block rolls back to the SAVEPOINT only.

```go
client.Transaction(ctx, func(outer context.Context) error {
    _ = client.Transaction(outer, func(inner context.Context) error {
        // Rolls back to inner's SAVEPOINT only
        return errors.New("rollback inner")
    })
    // Outer transaction can still commit
    return nil
})
```

## 5. Cancellation / timeouts

`DB(ctx)` already calls `WithContext(ctx)`, so `context.WithTimeout` / `context.WithCancel` work out of the box:

```go
ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
defer cancel()
users, err := s.users.Find(ctx, qb)
```

## 6. When to fall back to raw `*gorm.DB`

Some GORM features (subqueries, upsert, locking, returning) aren't yet typed in the Builder. Reach for `Repository.DB(ctx)`—still transaction-aware:

```go
err := r.DB(ctx).
    Clauses(clause.Locking{Strength: "UPDATE"}).
    Where("id = ?", id).
    First(&u).Error
```

The transactional isolation is preserved.
