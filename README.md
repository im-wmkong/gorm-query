# GORM Query 🚀

*[Read this in Chinese / 中文版](README_CN.md)*

[![Go Reference](https://pkg.go.dev/badge/github.com/im-wmkong/gorm-query.svg)](https://pkg.go.dev/github.com/im-wmkong/gorm-query)
[![Go Report Card](https://goreportcard.com/badge/github.com/im-wmkong/gorm-query)](https://goreportcard.com/report/github.com/im-wmkong/gorm-query)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

**GORM Query** is a strongly-typed query builder and generic repository library built on top of GORM.

It eliminates fragile "magic strings" in GORM queries through **code generation**, providing a smooth fluent API experience. It also includes generic repositories and context-aware transaction management for cleaner service and data layers.

## ✨ Core Features

- 🛡️ **Strongly-typed Query Building**: Say goodbye to `db.Where("age > ?", 18)` and embrace `columns.User.Age.Gt(18)`. Catch field name typos at compile time.
- 📦 **Out-of-the-box Generic Repository**: Use `repo.BaseRepository[T]` to gain full CRUD capabilities with a single line of code.
- 🎯 **Stop Bloating Repositories**: Combine the universal query builder to compose dynamic queries on the fly—no more writing dozens of `FindByXxx` methods.
- 🔄 **Implicit Context Transactions**: Pass transactions via `context.Context`. Decouple your Service layer from the Repo layer without passing `*gorm.DB` everywhere.

## 📦 Installation

```bash
go get github.com/im-wmkong/gorm-query
```

## 🗺️ Capability Map

GORM Query consists of 4 core modules:

| Module | Core Responsibility | Key Capabilities / Methods |
| :--- | :--- | :--- |
| **`colgen`** | **Code Generation** | Generates column definitions like `columns.User` from GORM models. Supports custom output, package names, and dry-run validation. |
| **`query`** | **Dynamic Query Builder** | **Builder**: `Where`, `Or`, `Select`, `Joins`, `Preload`, `Page`, `Apply`... <br>**Column**: `Eq`, `Gt`, `Like`, `In`, `Between`, `Sum`, `Asc`... |
| **`repo`** | **Generic Repository** | Provides common CRUD methods: `Create`, `Save`, `Find`, `First`, `Update`, `Delete`, `Pluck`... |
| **`db`** | **Context Transaction** | `db.Client` implements both `DBProvider` and `Transactor`. Repositories automatically reuse the same transaction via `ctx`. |

**Best Path to Start:** `colgen` generation ➔ `columns.Xxx` definitions ➔ `query` builder ➔ `repo` execution.

## 🚀 Quick Start

### 1. Define Your Model

Define your GORM model as usual:

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

### 2. Configure and Run Code Generation

Create a simple generation script (e.g., at `cmd/gen/main.go`) and pass the models you want to generate query columns for:

```go
package main

import (
    "log"

    "your_project_name/model" // Replace with your actual project path
    "github.com/im-wmkong/gorm-query/colgen"
)

func main() {
    // Initialize the generator and provide your models
    err := colgen.New().Generate(&model.User{})

    if err != nil {
        log.Fatalf("generate failed: %v", err)
    }
}
```

Run the script from your terminal:

```bash
go run cmd/gen/main.go
```
*This will automatically create a code file (e.g., `model/columns/user_gen.go`) containing the generated `columns.User` variable.*

By default, `colgen` writes generated columns into a dedicated `columns` package.

> **💡 Pro Tip:** You can add `//go:generate go run cmd/gen/main.go` to the top of any Go file and trigger generation using `go generate ./...` in your standard workflow.

### 3. Enjoy Smooth Strongly-typed Queries

Now you can use the generated `columns.User` with the Query Builder for type-safe queries:

```go
import (
    "gorm.io/gorm"

    "your_project_name/model/columns"
    "your_project_name/model"
    "github.com/im-wmkong/gorm-query/query"
)

var db *gorm.DB

// 1. Build queries fluently
qb := query.New().
    Where(
        columns.User.Age.Gte(18),
        columns.User.UserName.Contains("wmkong"),
    ).
    Page(1, 20).
    Order(columns.User.ID.Desc())

// 2. Apply to gorm.DB
var users []model.User
err := qb.Apply(db).Find(&users).Error
```

## 💡 Advanced Usage

### 1. Generic Repository & Context-Aware Transactions

`db.Client` is the glue between your Service and Repository layers:

- `db.Client` implements both `db.DBProvider` and `db.Transactor`
- Repositories depend only on `db.DBProvider`
- Services depend only on `db.Transactor`
- The active transaction flows through `context.Context`, so you never need to pass `*gorm.DB` around manually

For brevity, the example below accepts `*db.Client` directly, since it already implements both interfaces.

**Define a Service with Generic Repositories:**
```go
type UserService struct {
    users    repo.Repository[model.User]
    profiles repo.Repository[model.Profile]
    tx       db.Transactor
}

func NewUserService(dbClient *db.Client) *UserService {
    return &UserService{
        users:    repo.New[model.User](dbClient),
        profiles: repo.New[model.Profile](dbClient),
        tx:       dbClient,
    }
}
```

**Wrap Business Logic in One Transaction:**
```go
// The service coordinates the transaction once.
// Repositories automatically pick it up from txCtx.
func (s *UserService) CreateUserAndProfile(ctx context.Context, user *model.User, profile *model.Profile) error {
    return s.tx.Transaction(ctx, func(txCtx context.Context) error {
        if err := s.users.Create(txCtx, user); err != nil {
            return err
        }

        profile.UserID = user.ID
        if err := s.profiles.Create(txCtx, profile); err != nil {
            return err
        }

        return nil
    })
}
```

Both repository calls receive the same `txCtx`, so they run inside the same transaction automatically. If any step returns an error, GORM rolls everything back for you.

When you need custom repository methods later, you can still wrap `repo.New[T](dbClient)` inside your own repository types without changing this transaction model.

### 2. Dynamic Repository Queries

Stop inflating your Repository interfaces with dozens of specific methods like `FindByNameAndAge`. Use the `query.Builder` to handle dynamic conditions in the Service layer while keeping your Repository clean.

Once your service already depends on a generic repository, `query.Builder` becomes the missing piece for dynamic filtering.

```go
func (s *UserService) GetUsers(ctx context.Context, name string, minAge int) ([]*model.User, error) {
    // 1. Build dynamic conditions
    qb := query.New().Where(columns.User.Status.Eq(1))

    if name != "" {
        qb = qb.Where(columns.User.UserName.Contains(name))
    }
    if minAge > 0 {
        qb = qb.Where(columns.User.Age.Gte(minAge))
    }

    // 2. Pass the builder directly to the generic Find method
    return s.users.Find(ctx, qb)
}
```

### 3. Query Reuse (Cloning)

Use `.Clone()` when multiple derived queries start from the same base builder.

It helps you derive new queries from a base query without polluting the original:

```go
baseQuery := query.New().Where(columns.User.Status.Eq(1))

// Derived Query A
adults := baseQuery.Clone().Where(columns.User.Age.Gte(18))

// Derived Query B (Will NOT include Age >= 18 condition)
minors := baseQuery.Clone().Where(columns.User.Age.Lt(18))
```

## 🤝 Contributing

Issues and Pull Requests are welcome!

Before submitting, please run:
```bash
make tidy
make generate
make test
```

## 📄 License

This project is licensed under the [MIT License](LICENSE).
