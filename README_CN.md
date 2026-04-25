# GORM Query 🚀

[![Go Reference](https://pkg.go.dev/badge/github.com/im-wmkong/gorm-query.svg)](https://pkg.go.dev/github.com/im-wmkong/gorm-query)
[![Go Report Card](https://goreportcard.com/badge/github.com/im-wmkong/gorm-query)](https://goreportcard.com/report/github.com/im-wmkong/gorm-query)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

**GORM Query** 是一个基于 GORM 的强类型查询构建器与通用仓储（Repository）库。

它通过**代码生成**消除了 GORM 查询中脆弱的“魔法字符串”，提供丝滑的链式调用体验，并内置了企业级的泛型 Repository 与基于 Context 的事务管理方案。

## ✨ 核心特性

- 🛡️ **强类型查询构建**：告别 `db.Where("age > ?", 18)`，拥抱 `columns.User.Age.Gt(18)`，在编译期拦截字段名拼写错误。
- 📦 **开箱即用的泛型仓储**：提供 `repo.BaseRepository[T]`，一行代码拥有完整的 CRUD 能力。
- 🎯 **告别臃肿的 Repository**：结合通用 builder 查询构建器，按需动态组合查询条件，无需再为不同业务编写数十个 `FindByXxx` 方法。
- 🔄 **隐式上下文事务**：基于 `context.Context` 传递事务，Service 层与 Repo 层彻底解耦，再也不用把 `*gorm.DB` 传来传去。

## 📦 安装

```bash
go get github.com/im-wmkong/gorm-query
```

## 🚀 快速开始

### 1. 定义你的模型

先像平常一样定义 GORM 模型：

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

### 2. 配置并执行代码生成

创建一个简单的生成脚本（例如放在 `cmd/gen/main.go`），将你需要生成查询属性的模型传入生成器：

```go
package main

import (
    "log"

    "your_project_name/model" // 替换为你的实际项目路径
    "github.com/im-wmkong/gorm-query/colgen"
)

func main() {
    // 实例化生成器并传入模型
   err := colgen.New().Generate(&model.User{})
    
    if err != nil {
        log.Fatalf("generate failed: %v", err)
    }
}
```

在终端直接运行该脚本：

```bash
go run cmd/gen/main.go
```
*这会自动生成一个代码文件（例如 `model/columns/user_gen.go`），其中包含 `columns.User` 变量。*

> **💡 进阶提示：** 你也可以在任意 Go 文件的头部添加 `//go:generate go run cmd/gen/main.go`，之后就能通过在项目根目录运行 `go generate ./...` 将其无缝接入你的标准工作流。

### 3. 享受丝滑的强类型查询

现在，你可以使用生成的 `columns.User` 配合 Query Builder 进行类型安全的查询了：

```go
import (
    "your_project_name/model/columns"
    "your_project_name/model"
    "github.com/im-wmkong/gorm-query/query"
)

// 1. 丝滑地构建查询条件
qb := query.New().
    Where(
        columns.User.Age.Gte(18),
        columns.User.UserName.Contains("wmkong"),
    ).
    Page(1, 20).
    Order(columns.User.ID.Desc())

// 2. 应用到 gorm.DB
var users []model.User
err := qb.Apply(db).Find(&users).Error
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

**初始化与注入：**
```go
import (
    "github.com/im-wmkong/gorm-query/db"
    "github.com/im-wmkong/gorm-query/repo"
)

// 1. 初始化 DB Client
// 它同时实现了 repo 需要的 db.DBProvider 和 service 需要的 db.Transactor 接口
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
    return s.transactor.Transaction(ctx, func(txCtx context.Context) error {
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
        columns.User.Status.Eq(1), // 默认条件
    )

    // 2. 动态追加条件
    if name != "" {
        qb = qb.Where(columns.User.UserName.Contains(name))
    }
    if minAge > 0 {
        qb = qb.Where(columns.User.Age.Gte(minAge))
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
baseQuery := query.New().Where(columns.User.Status.Eq(1))

// 派生查询 A
adultsQuery := baseQuery.Clone().Where(columns.User.Age.Gte(18))

// 派生查询 B (不会包含 Age >= 18 的条件)
minorsQuery := baseQuery.Clone().Where(columns.User.Age.Lt(18))
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
