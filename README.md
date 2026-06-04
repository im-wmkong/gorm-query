# GORM Query 🚀

*[Read this in Chinese / 中文版](README_CN.md)*

[![Go Reference](https://pkg.go.dev/badge/github.com/im-wmkong/gorm-query.svg)](https://pkg.go.dev/github.com/im-wmkong/gorm-query)
[![Go Report Card](https://goreportcard.com/badge/github.com/im-wmkong/gorm-query)](https://goreportcard.com/report/github.com/im-wmkong/gorm-query)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

**GORM Query** is a strongly-typed query builder and generic repository library built on top of GORM.

It eliminates fragile "magic strings" in GORM queries through **code generation**, providing a smooth fluent API. It also includes generic repositories and context-aware transaction management for cleaner service and data layers.

## ✨ Core Features

- 🛡️ **Strongly-typed Query Building** — say goodbye to `db.Where("age > ?", 18)`, embrace `schema.User.Age.Gt(18)`. Field-name typos fail at compile time.
- 📦 **Generic Repository** — `repo.BaseRepository[T]` gives you full CRUD in one line.
- 🎯 **No more bloated repositories** — compose dynamic queries with `query.Builder` instead of writing dozens of `FindByXxx` methods.
- 🔄 **Implicit context transactions** — pass transactions via `context.Context`. Service and Repository layers stay decoupled from `*gorm.DB`.
- 🧊 **Immutable, concurrent-safe Builder** — every chained call returns a new Builder; derived queries never share state.

## 📦 Installation

```bash
go get github.com/im-wmkong/gorm-query
```

## 🗺️ Capability Map

| Module | Responsibility | Highlights |
| :--- | :--- | :--- |
| **`schemagen`** | Code generation | Parses GORM models into typed schema dictionaries (columns + associations). Options: `WithOutputDir`, `WithPackageName`, `WithNamingStrategy`, `WithDryRun`, `WithLogger`. |
| **`query`** | Dynamic query builder | **Builder**: `Where / Or / Not / Select / Omit / Distinct / Preload / Joins / InnerJoins / Group / Having / Order / Page / Limit / Offset / Unscoped / Scope / Apply` <br>**Column**: `Eq / Neq / Gt / Gte / Lt / Lte / In / NotIn / Between / NotBetween / Like / Contains / HasPrefix / HasSuffix / IsNull / IsNotNull / Sum / Count / Avg / Max / Min / As / Distinct / Set / Asc / Desc / WithTable` <br>**Association**: `Preload / Joins / Nested` |
| **`repo`** | Generic repository | `Save / Create / CreateInBatches / Update / Delete / Find / First / Take / Last / Count / Exists / Pluck / DB(ctx)` |
| **`db`** | Context transactions | `Client` implements `DBProvider` + `Transactor`; transactions flow through `context.Context`. |

→ Full API reference on [pkg.go.dev](https://pkg.go.dev/github.com/im-wmkong/gorm-query). Module deep-dives live in [`docs/`](docs/README.md).

## 🚀 Quick Start

### 1. Define your model

```go
package model

import "gorm.io/gorm"

type User struct {
    gorm.Model
    UserName string `gorm:"column:user_name"`
    Email    string `gorm:"column:email"`
    Age      int    `gorm:"column:age"`
    Status   int    `gorm:"column:status"`
}
```

### 2. Generate the schema

Create a tiny generation script (e.g. `cmd/gen/main.go`):

```go
package main

import (
    "log"

    "your_project_name/model"
    "github.com/im-wmkong/gorm-query/schemagen"
)

func main() {
    if err := schemagen.New().Generate(&model.User{}); err != nil {
        log.Fatalf("generate failed: %v", err)
    }
}
```

```bash
go run cmd/gen/main.go
```

This emits `model/schema/user_gen.go` with a `schema.User` variable.

> **💡 Pro tip**: add `//go:generate go run cmd/gen/main.go` to a Go file and trigger generation via `go generate ./...`.

→ See [docs/en/schemagen.md](docs/en/schemagen.md) for all options, naming rules, and limitations.

### 3. Build type-safe queries

```go
import (
    "your_project_name/model"
    "your_project_name/model/schema"
)

qb := schema.User.Query().
    Where(
        schema.User.Age.Gte(18),
        schema.User.UserName.Contains("wmkong"),
    ).
    Order(schema.User.CreatedAt.Desc()).
    Page(1, 20)

// Apply onto *gorm.DB...
var users []model.User
err := qb.Apply(db.Model(&model.User{})).Find(&users).Error
```

→ Full operator reference: [docs/en/query-builder.md](docs/en/query-builder.md).

### 4. Or hand the Builder to a Repository

```go
client := db.NewClient(gormDB)
users  := repo.New[model.User](client)

list, err := users.Find(ctx,
    schema.User.Query().Where(schema.User.Status.Eq(1)),
)
```

→ All repository methods in [docs/en/repository.md](docs/en/repository.md).

## 💡 Patterns at a glance

### Context-aware transactions

```go
type UserService struct {
    users repo.Repository[model.User]
    tx    db.Transactor
}

func (s *UserService) Register(ctx context.Context, u *model.User, p *model.Profile) error {
    return s.tx.Transaction(ctx, func(txCtx context.Context) error {
        if err := s.users.Create(txCtx, u); err != nil {
            return err
        }
        p.UserID = u.ID
        return s.profiles.Create(txCtx, p)
    })
}
```

→ Propagation contract, nested transactions, cancellation: [docs/en/transaction.md](docs/en/transaction.md).

### Dynamic queries without bloating Repositories

```go
func (s *UserService) GetUsers(ctx context.Context, name string, minAge int) ([]*model.User, error) {
    qb := schema.User.Query().Where(schema.User.Status.Eq(1))
    if name != "" {
        qb = qb.Where(schema.User.UserName.Contains(name))
    }
    if minAge > 0 {
        qb = qb.Where(schema.User.Age.Gte(minAge))
    }
    return s.users.Find(ctx, qb)
}
```

### Updates with type-safe assignments

```go
qb := schema.User.Query().Where(schema.User.ID.Eq(1))
rows, err := s.users.Update(ctx, qb,
    schema.User.Status.Set(2),
    schema.User.Email.Set("a@b.com"),
)
```

### Preload + nested associations

```go
schema.User.Query().Preload(
    schema.User.Profile.Nested(schema.Profile.Address),
    schema.Address.City.Eq("SF"),
)
```

### Escape hatches

```go
// Builder: inject raw GORM logic
qb.Scope(func(db *gorm.DB) *gorm.DB { return db.Where("status = ?", 1) })

// Repository: grab the transaction-aware *gorm.DB
err := r.DB(ctx).
    Clauses(clause.OnConflict{DoNothing: true}).
    Create(&u).Error
```

## 📚 Documentation

A full runnable demo (SQLite + schemagen + repository + service + tests) lives in [`example/`](example).

| Topic | Doc |
| :--- | :--- |
| Query Builder API | [docs/en/query-builder.md](docs/en/query-builder.md) |
| Repository methods | [docs/en/repository.md](docs/en/repository.md) |
| Transaction model | [docs/en/transaction.md](docs/en/transaction.md) |
| Code generator | [docs/en/schemagen.md](docs/en/schemagen.md) |
| FAQ & pitfalls | [docs/en/faq.md](docs/en/faq.md) |

→ Index for both English and Chinese: [docs/README.md](docs/README.md).

## 🚧 Known limitations

- No type-safe subqueries / `EXISTS` / `IN (SELECT ...)`.
- No typed `OnConflict` / `Upsert` / `Returning` / `FOR UPDATE`.
- `Having(...)` and `RawFragment` are still string-based — drop into `repo.DB(ctx)` when needed.
- Integration tests cover SQLite only; verify MySQL / Postgres specifics yourself.

## 🤝 Contributing

Issues and PRs welcome. Before submitting:

```bash
make tidy
make generate
make test
```

> When you change `README.md`, please keep `README_CN.md` in sync (and vice versa).

## 📄 License

[MIT License](LICENSE).
