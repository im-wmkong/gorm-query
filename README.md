# GORM Query 🚀

[![Go Reference](https://pkg.go.dev/badge/github.com/im-wmkong/gorm-query.svg)](https://pkg.go.dev/github.com/im-wmkong/gorm-query)
[![Go Report Card](https://goreportcard.com/badge/github.com/im-wmkong/gorm-query)](https://goreportcard.com/report/github.com/im-wmkong/gorm-query)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

**GORM Query** 是一个基于 GORM 的强类型查询构建器与通用仓储（Repository）库。

它通过**代码生成**消除了 GORM 查询中脆弱的“魔法字符串”，提供丝滑的链式调用体验，并内置了企业级的泛型 Repository 与基于 Context 的事务管理方案。

## ✨ 核心特性

- 🛡️ **强类型查询构建**：告别 `db.Where("age > ?", 18)`，拥抱 `UserProps.Age.Gte(18)`，在编译期拦截字段名拼写错误。
- 📦 **开箱即用的泛型仓储**：提供 `repo.BaseRepository[T]`，一行代码拥有完整的 CRUD 能力。
- 🎯 **告别臃肿的 Repository**：结合通用 builder 查询构建器，按需动态组合查询条件，无需再为不同业务编写数十个 `FindByXxx` 方法。
- 🔄 **隐式上下文事务**：基于 `context.Context` 传递事务，Service 层与 Repo 层彻底解耦，再也不用把 `*gorm.DB` 传来传去。
- 🧬 **零依赖污染**：生成的属性代码轻量纯净，采用私有常量机制，保证重复调用的绝对幂等性。

## 📦 安装

```bash
go get github.com/im-wmkong/gorm-query
```

## 🚀 快速开始

### 1. 定义你的模型并添加生成指令

在你的实体模型（Model）文件中引入包，并在结构体上方添加 `//go:generate` 注释：

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

### 2. 生成强类型属性代码

在项目根目录（或模型所在目录）运行：

```bash
go generate ./...
```
*这将会自动生成 `user_gen.go` 文件，包含 `UserProps` 变量。*

### 3. 使用 Query Builder

```go
import (
    "github.com/im-wmkong/gorm-query/query"
    // 引入 model package
)

// 构建查询条件
qb := query.New().
    Where(
        model.UserProps.Age.Gte(18),
        model.UserProps.Name.Contains("wmkong"),
    ).
    Page(1, 20).
    Order(model.UserProps.ID.Desc())

// 应用到 gorm.DB
var users []model.User
err := qb.Apply(db.Model(&model.User{})).Find(&users).Error
```

## 💡 高级用法

### 1. 泛型 Repository 与上下文事务 (Context-Aware TX)

结合提供的 `db.Client` 和 `repo.BaseRepository`，你可以构建极度整洁的架构：

**定义 Repository 和 Service：**
```go
// 定义 UserRepository
type UserRepository struct {
    repo.BaseRepository[model.User]
}

func NewUserRepository(dbClient db.Client) *UserRepository {
    return &UserRepository{
        repo.New[model.User](dbClient),
    }
}

// 定义 ProfileRepository
type ProfileRepository struct {
    repo.BaseRepository[model.Profile]
}

func NewProfileRepository(dbClient db.Client) *ProfileRepository {
    return &ProfileRepository{
        repo.New[model.Profile](dbClient),
    }
}

// 定义 UserService
type UserService struct {
    userRepo    *UserRepository
    profileRepo *ProfileRepository
    tm          db.TransactionManager
}

func NewUserService(userRepo *UserRepository, profileRepo *ProfileRepository, tm db.TransactionManager) *UserService {
    return &UserService{
        userRepo: userRepo,
        profileRepo: profileRepo,
        tm: tm,
    }
}
```

**初始化与注入：**
```go
import (
    "github.com/im-wmkong/gorm-query/db"
    "github.com/im-wmkong/gorm-query/repo"
)

// 1. 初始化 DB Client
// 它同时实现了 repo 需要的 db.Connector 和 service 需要的 db.TransactionManager 接口
dbClient := db.NewClient(gormDB)
// 2. 实例化 UserRepository
userRepo := NewUserRepository(dbClient)
// 3. 实例化 ProfileRepository
profileRepo := NewProfileRepository(dbClient)
// 4. 注入到 Service
userService := NewUserService(userRepo, profileRepo, dbClient)
```

**在 Service 层优雅地使用事务：**
```go
// 业务代码完全不需要知道底层 gorm.DB 的存在
func (s *UserService) CreateUserAndProfile(ctx context.Context, user *model.User, profile *model.Profile) error {
    // 开启事务
    return s.tm.Transaction(ctx, func(txCtx context.Context) error {
        // 自动使用 txCtx 中的事务连接
        if err := s.userRepo.Create(txCtx, user); err != nil {
            return err 
        }

        profile.UserID = user.ID
        // 如果这里失败，上面的 Create 会自动回滚
        if err := s.profileRepo.Create(txCtx, profile); err != nil {
            return err
        }
        // 所有操作都在同一个事务中，无需手动 Commit/Rollback
        return nil
    })
}
```

### 2. 告别臃肿：基于 Builder 的动态仓储查询

在传统的开发模式中，为了应对各种业务查询需求，Repository 接口往往会无限膨胀：

```go
// ❌ 传统模式下臃肿的 Repository 接口
// type UserRepository interface {
//     FindByName(ctx context.Context, name string) ([]*model.User, error)
//     FindByAgeGt(ctx context.Context, age int) ([]*model.User, error)
//     FindByStatusWithPage(ctx context.Context, status, page, size int) ([]*model.User, error)
//     // ... 以及其他几十个类似的方法
// }
```

**GORM Query** 的 `query.Builder` 彻底解决了这个问题。借助通用的查询构建能力，开发者可以在 Service 层自由定制查询条件，并直接传递给泛型仓储。这使得 Repository 层保持极简，不再需要定义任何多余的方法。

```go
// ✅ 现代模式：极简的 Repository + 强大的 Builder
func (s *UserService) GetUsersByDynamicConditions(ctx context.Context, name string, minAge int) ([]*model.User, error) {
    // 1. 使用 builder 自由组合查询条件
    qb := query.New().Where(
        model.UserProps.Status.Eq(1), // 默认条件
    )

    // 2. 动态追加条件
    if name != "" {
        qb = qb.Where(model.UserProps.Name.Contains(name))
    }
    if minAge > 0 {
        qb = qb.Where(model.UserProps.Age.Gte(minAge))
    }

    // 3. 直接将 builder 传递给 BaseRepository 的 Find 方法，无需在 repo 新增任何方法！
    users, err := s.userRepo.Find(ctx, qb)
    if err != nil {
        return nil, err
    }
    
    return users, nil
}
```

### 3. 查询条件的复用 (防污染)

如果需要基于一个基础查询派生出不同的查询，请使用 `.Clone()` 方法防止切片底层数组的条件污染：

```go
baseQuery := query.New().Where(UserProps.Status.Eq(1))

// 派生查询 A
adultsQuery := baseQuery.Clone().Where(UserProps.Age.Gte(18))

// 派生查询 B (不会包含 Age >= 18 的条件)
minorsQuery := baseQuery.Clone().Where(UserProps.Age.Lt(18))
```

## 🤝 参与贡献

欢迎提交 Issue 和 Pull Request！

修改代码后，请确保运行以下命令检查代码质量：
```bash
make tidy
make generate
make test
```

## 📄 开源协议

本项目采用 [MIT License](LICENSE) 协议进行开源。
