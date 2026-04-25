# GORM Query 🚀

*[Read this in Chinese / 中文版](README_CN.md)*

[![Go Reference](https://pkg.go.dev/badge/github.com/im-wmkong/gorm-query.svg)](https://pkg.go.dev/github.com/im-wmkong/gorm-query)
[![Go Report Card](https://goreportcard.com/badge/github.com/im-wmkong/gorm-query)](https://goreportcard.com/report/github.com/im-wmkong/gorm-query)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

**GORM Query** is a strongly-typed query builder and generic repository library built on top of GORM.

It eliminates fragile "magic strings" in GORM queries through **code generation**, providing a smooth fluent API experience. It also features enterprise-grade generic repositories and context-based transaction management.

## ✨ Core Features

- 🛡️ **Strongly-typed Query Building**: Say goodbye to `db.Where("age > ?", 18)` and embrace `columns.User.Age.Gt(18)`. Catch field name typos at compile time.
- 📦 **Out-of-the-box Generic Repository**: Use `repo.BaseRepository[T]` to gain full CRUD capabilities with a single line of code.
- 🎯 **Stop Bloating Repositories**: Combine the universal query builder to compose dynamic queries on the fly—no more writing dozens of `FindByXxx` methods.
- 🔄 **Implicit Context Transactions**: Pass transactions via `context.Context`. Decouple your Service layer from the Repo layer without passing `*gorm.DB` everywhere.

## 📦 Installation

```bash
go get github.com/im-wmkong/gorm-query
```

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

Create a simple generation script (e.g., at `cmd/gen/main.go`) and pass the models you want to generate properties for:

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

> **💡 Pro Tip:** You can add `//go:generate go run cmd/gen/main.go` to the top of any Go file and trigger generation using `go generate ./...` in your standard workflow.

### 3. Enjoy Smooth Strongly-typed Queries

Now you can use the generated `columns.User` with the Query Builder for type-safe queries:

```go
import (
    "your_project_name/model/columns"
    "your_project_name/model"
    "github.com/im-wmkong/gorm-query/query"
)

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

Combine `db.Client` and `repo.BaseRepository` to build a clean architecture:

**Define Repositories and Service:**
```go
// Define UserRepository
type UserRepository struct {
    repo.BaseRepository[model.User]
}

func NewUserRepository(dbClient db.Client) *UserRepository {
    return &UserRepository{
        repo.New[model.User](dbClient),
    }
}

// Define ProfileRepository
type ProfileRepository struct {
    repo.BaseRepository[model.Profile]
}

func NewProfileRepository(dbClient db.Client) *ProfileRepository {
    return &ProfileRepository{
        repo.New[model.Profile](dbClient),
    }
}

// Define UserService
type UserService struct {
    userRepo    *UserRepository
    profileRepo *ProfileRepository
    transactor  db.Transactor
}

func NewUserService(userRepo *UserRepository, profileRepo *ProfileRepository, transactor db.Transactor) *UserService {
    return &UserService{
        userRepo:    userRepo,
        profileRepo: profileRepo,
        transactor:  transactor,
    }
}
```

**Initialization and Injection:**
```go
import (
    "github.com/im-wmkong/gorm-query/db"
    "github.com/im-wmkong/gorm-query/repo"
)

// 1. Initialize DB Client
dbClient := db.NewClient(gormDB)
// 2. Instantiate Repositories
userRepo := NewUserRepository(dbClient)
profileRepo := NewProfileRepository(dbClient)
// 3. Inject into Service
userService := NewUserService(userRepo, profileRepo, dbClient)
```

**Elegant Transaction Management in Service Layer:**
```go
// Business logic doesn't need to know about gorm.DB
func (s *UserService) CreateUserAndProfile(ctx context.Context, user *model.User, profile *model.Profile) error {
    // Transaction starts here
    return s.transactor.Transaction(ctx, func(txCtx context.Context) error {
        // Automatically uses the transaction stored in txCtx
        if err := s.userRepo.Create(txCtx, user); err != nil {
            return err 
        }

        profile.UserID = user.ID
        // If this fails, the previous Create will automatically roll back
        if err := s.profileRepo.Create(txCtx, profile); err != nil {
            return err
        }
        
        return nil
    })
}
```

### 2. Dynamic Repository Queries

Stop inflating your Repository interfaces with dozens of specific methods like `FindByNameAndAge`. Use the `query.Builder` to handle dynamic conditions in the Service layer while keeping your Repository clean.

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
    return s.userRepo.Find(ctx, qb)
}
```

### 3. Query Reuse (Cloning)

Use `.Clone()` to derive new queries from a base query without polluting the original:

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
