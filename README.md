# GORM Query 🚀

*[Read this in Chinese / 中文版](README_CN.md)*

[![Go Reference](https://pkg.go.dev/badge/github.com/im-wmkong/gorm-query.svg)](https://pkg.go.dev/github.com/im-wmkong/gorm-query)
[![Go Report Card](https://goreportcard.com/badge/github.com/im-wmkong/gorm-query)](https://goreportcard.com/report/github.com/im-wmkong/gorm-query)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

**GORM Query** is a strongly-typed query builder and generic repository library based on GORM.

It eliminates the fragile "magic strings" in GORM queries through **code generation**, providing a silky-smooth method-chaining experience. It also features built-in enterprise-grade generic repositories and a context-based transaction management solution.

## ✨ Core Features

- 🛡️ **Strongly-Typed Query Building**: Say goodbye to `db.Where("age > ?", 18)` and embrace `UserProps.Age.Gte(18)`. Catch field spelling errors at compile time.
- 📦 **Out-of-the-Box Generic Repository**: Provides `repo.BaseRepository[T]`, granting you full CRUD capabilities with just one line of code.
- 🎯 **Farewell to Bloated Repositories**: Combined with the generic query builder, you can dynamically compose query conditions on demand. No need to write dozens of `FindByXxx` methods for different business scenarios.
- 🔄 **Implicit Contextual Transactions**: Pass transactions via `context.Context`, completely decoupling the Service layer from the Repo layer. No more passing `*gorm.DB` around everywhere.

## 📦 Installation

```bash
go get github.com/im-wmkong/gorm-query
```

## 🚀 Quick Start

### 1. Define your model and add the generation directive

Import the package in your entity model file and add the `//go:generate` comment above your struct:

```go
package model

import "gorm.io/gorm"

//go:generate go run github.com/im-wmkong/gorm-query/cmd/gen-props@latest -type=User
type User struct {
    gorm.Model
    Name   string `gorm:"column:user_name"`
    Age    int
    Status int
}
```

### 2. Generate strongly-typed property code

Run the following command in your project root (or the directory where the model is located):

```bash
go generate ./...
```
*This will automatically generate a `user_gen.go` file containing the `UserProps` variable.*

### 3. Use the Query Builder

```go
import (
    "github.com/im-wmkong/gorm-query/query"
    // Import your model package
)

// Build query conditions
qb := query.New().
    Where(
        model.UserProps.Age.Gte(18),
        model.UserProps.Name.Contains("wmkong"),
    ).
    Page(1, 20).
    Order(model.UserProps.ID.Desc())

// Apply to gorm.DB
var users []model.User
err := qb.Apply(db.Model(&model.User{})).Find(&users).Error
```

## 💡 Advanced Usage

### 1. Generic Repository and Context-Aware TX

By combining the provided `db.Client` and `repo.BaseRepository`, you can build an extremely clean architecture:

**Define Repository and Service:**
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
    tm          db.TransactionManager
}

func NewUserService(userRepo *UserRepository, profileRepo *ProfileRepository, tm db.TransactionManager) *UserService {
    return &UserService{
        userRepo:    userRepo,
        profileRepo: profileRepo,
        tm:          tm,
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
// It implements both the db.Connector interface required by the repo and the db.TransactionManager interface required by the service
dbClient := db.NewClient(gormDB)

// 2. Instantiate UserRepository
userRepo := NewUserRepository(dbClient)

// 3. Instantiate ProfileRepository
profileRepo := NewProfileRepository(dbClient)

// 4. Inject into Service
userService := NewUserService(userRepo, profileRepo, dbClient)
```

**Elegantly use transactions in the Service layer:**
```go
// The business code doesn't need to know about the underlying gorm.DB at all
func (s *UserService) CreateUserAndProfile(ctx context.Context, user *model.User, profile *model.Profile) error {
    // Start transaction
    return s.tm.Transaction(ctx, func(txCtx context.Context) error {
        // Automatically uses the transaction connection from txCtx
        if err := s.userRepo.Create(txCtx, user); err != nil {
            return err 
        }

        profile.UserID = user.ID
        // If this fails, the Create above will automatically roll back
        if err := s.profileRepo.Create(txCtx, profile); err != nil {
            return err
        }
        // All operations are within the same transaction, no manual Commit/Rollback needed
        return nil
    })
}
```

### 2. Say Goodbye to Bloat: Dynamic Repository Queries

In traditional development patterns, to meet various business query requirements, Repository interfaces tend to expand infinitely:

```go
// ❌ Bloated Repository interface in traditional mode
// type UserRepository interface {
//     FindByName(ctx context.Context, name string) ([]*model.User, error)
//     FindByAgeGt(ctx context.Context, age int) ([]*model.User, error)
//     FindByStatusWithPage(ctx context.Context, status, page, size int) ([]*model.User, error)
//     // ... and dozens of other similar methods
// }
```

**GORM Query**'s `query.Builder` completely solves this problem. Leveraging generic query building capabilities, developers can freely customize query conditions in the Service layer and pass them directly to the generic repository. This keeps the Repository layer extremely minimalist, eliminating the need to define redundant methods.

```go
// ✅ Modern mode: Minimalist Repository + Powerful Builder
func (s *UserService) GetUsersByDynamicConditions(ctx context.Context, name string, minAge int) ([]*model.User, error) {
    // 1. Use the builder to freely compose query conditions
    qb := query.New().Where(
        model.UserProps.Status.Eq(1), // Default condition
    )

    // 2. Dynamically append conditions
    if name != "" {
        qb = qb.Where(model.UserProps.Name.Contains(name))
    }
    if minAge > 0 {
        qb = qb.Where(model.UserProps.Age.Gte(minAge))
    }

    // 3. Pass the builder directly to the BaseRepository's Find method; no need to add any methods in the repo!
    users, err := s.userRepo.Find(ctx, qb)
    if err != nil {
        return nil, err
    }
    
    return users, nil
}
```

### 3. Reusing Query Conditions (Preventing Pollution)

If you need to derive different queries from a base query, please use the `.Clone()` method to prevent condition pollution of the underlying slice array:

```go
baseQuery := query.New().Where(UserProps.Status.Eq(1))

// Derived query A
adultsQuery := baseQuery.Clone().Where(UserProps.Age.Gte(18))

// Derived query B (will NOT contain the Age >= 18 condition)
minorsQuery := baseQuery.Clone().Where(UserProps.Age.Lt(18))
```

## 🤝 Contributing

Issues and Pull Requests are welcome!

After modifying the code, please ensure you run the following commands to check code quality:
```bash
make tidy
make generate
make test
```

## 📄 License

This project is open-sourced under the [MIT License](LICENSE).
