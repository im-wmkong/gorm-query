# Repository

`repo.BaseRepository[T]` is a generic CRUD implementation built on top of `db.DBProvider`. It picks up the active transaction from `context.Context` automatically, so callers don't have to know whether they're inside a transaction.

## 1. Creating a repository

```go
// dbClient implements both db.DBProvider and db.Transactor.
client := db.NewClient(gormDB)

// Generic repo bound to model.User.
users := repo.New[model.User](client)
```

## 2. Full method set

`Repository[T]` is defined in `repo/repository.go`:

| Method | Description |
| :--- | :--- |
| `DB(ctx)` | Returns the ctx-bound `*gorm.DB` (escape hatch). |
| `Save(ctx, entity)` | Upsert based on the primary key. |
| `Create(ctx, entity)` | INSERT only. |
| `CreateInBatches(ctx, entities, batchSize) (int64, error)` | Batched INSERT, returns rows affected. |
| `Update(ctx, qb, assigns...) (int64, error)` | UPDATE matched rows; `len(assigns) == 0` returns `(0, query.ErrNoAssignment)`. |
| `Delete(ctx, qb) (int64, error)` | DELETE matched rows. |
| `Find(ctx, qb) ([]*T, error)` | List query. |
| `First(ctx, qb) (*T, error)` | First by PK; not found → `(nil, gorm.ErrRecordNotFound)`. |
| `Take(ctx, qb) (*T, error)` | Any single row; same not-found behavior. |
| `Last(ctx, qb) (*T, error)` | Last by PK descending. |
| `Count(ctx, qb) (int64, error)` | Count rows. |
| `Exists(ctx, qb) (bool, error)` | `SELECT 1 LIMIT 1`. |
| `Pluck(ctx, qb, col, dest)` | Pluck a single column into `dest` (slice or scalar). |

## 3. Writes: Save / Create / CreateInBatches

```go
// INSERT only
users.Create(ctx, &model.User{UserName: "Alice"})

// Upsert by PK
users.Save(ctx, &model.User{ID: 1, UserName: "Alice"})

// Batch insert
list := []*model.User{{UserName: "A"}, {UserName: "B"}}
rows, err := users.CreateInBatches(ctx, list, 100)
```

## 4. Updates: `Update + Set`

Every column exposes `Set(v)`, returning a typed `Assignment`:

```go
qb := schema.User.Query().Where(schema.User.ID.Eq(1))
rows, err := users.Update(ctx, qb,
    schema.User.Status.Set(2),
    schema.User.Email.Set("a@b.com"),
)
```

> 💡 `Update` calls GORM's `Updates(map[string]any)`, so zero values are written explicitly when you ask for them via `Set(0)` / `Set("")`.

## 5. Reads: First / Take / Last / Find / Count / Exists / Pluck

```go
// Single row
u, err := users.First(ctx, schema.User.Query().Where(schema.User.Email.Eq("a@b.com")))
if errors.Is(err, gorm.ErrRecordNotFound) { ... }

// List
list, err := users.Find(ctx,
    schema.User.Query().
        Where(schema.User.Status.Eq(1)).
        Order(schema.User.CreatedAt.Desc()),
)

// Count / Exists
n, _   := users.Count(ctx, qb)
ok, _  := users.Exists(ctx, qb)

// Pluck a column
var emails []string
_ = users.Pluck(ctx,
    schema.User.Query().Where(schema.User.Status.Eq(1)),
    schema.User.Email,
    &emails,
)

// Aggregate scalar
var total int64
_ = users.Pluck(ctx,
    schema.User.Query(),
    schema.User.Age.Sum(),
    &total,
)
```

## 6. Delete

```go
rows, err := users.Delete(ctx, schema.User.Query().Where(schema.User.Status.Eq(0)))
```

If the model has `gorm.DeletedAt` this performs a **soft delete**. Use `Unscoped()` on the builder for a hard delete:

```go
users.Delete(ctx, schema.User.Query().Unscoped().Where(schema.User.ID.Eq(1)))
```

## 7. Custom repository pattern

`BaseRepository` is a regular struct—embed it to extend:

```go
type UserRepository interface {
    repo.Repository[model.User]
    FindActiveByEmail(ctx context.Context, email string) (*model.User, error)
}

type userRepository struct {
    *repo.BaseRepository[model.User]
}

func NewUserRepository(p db.DBProvider) UserRepository {
    return &userRepository{BaseRepository: repo.New[model.User](p)}
}

func (r *userRepository) FindActiveByEmail(ctx context.Context, email string) (*model.User, error) {
    qb := schema.User.Query().Where(
        schema.User.Status.Eq(1),
        schema.User.Email.Eq(email),
    )
    return r.First(ctx, qb)
}
```

Full example: [`example/repository/user_repo.go`](../../example/repository/user_repo.go).

## 8. Escape hatch: `DB(ctx)`

When the Builder doesn't yet cover something (subqueries, upsert, `FOR UPDATE`, …), grab the GORM handle directly:

```go
err := r.DB(ctx).
    Clauses(clause.OnConflict{DoNothing: true}).
    Create(&user).Error
```

It still picks up any active transaction from `ctx`.
