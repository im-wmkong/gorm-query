# GORM Query 🚀

[![Go Reference](https://pkg.go.dev/badge/github.com/im-wmkong/gorm-query.svg)](https://pkg.go.dev/github.com/im-wmkong/gorm-query)
[![Go Report Card](https://goreportcard.com/badge/github.com/im-wmkong/gorm-query)](https://goreportcard.com/report/github.com/im-wmkong/gorm-query)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

**GORM Query** 是一个基于 GORM 的强类型查询构建器与通用仓储（Repository）库。

它通过**代码生成**消除了 GORM 查询中脆弱的“魔法字符串”，提供丝滑的链式调用体验。它还内置了泛型 Repository 与基于 Context 的事务管理方案，帮助你写出更整洁的 Service 和数据访问层代码。

## ✨ 核心特性

- 🛡️ **强类型查询构建**：告别 `db.Where("age > ?", 18)`，拥抱 `columns.User.Age.Gt(18)`，在编译期拦截字段名拼写错误。
- 📦 **开箱即用的泛型仓储**：提供 `repo.BaseRepository[T]`，一行代码拥有完整的 CRUD 能力。
- 🎯 **告别臃肿的 Repository**：结合通用 builder 查询构建器，按需动态组合查询条件，无需再为不同业务编写数十个 `FindByXxx` 方法。
- 🔄 **隐式上下文事务**：基于 `context.Context` 传递事务，Service 层与 Repo 层彻底解耦，再也不用把 `*gorm.DB` 传来传去。

## 📦 安装

```bash
go get github.com/im-wmkong/gorm-query
```

## 🗺️ 能力总览

GORM Query 的核心能力由以下 4 个模块构成，它们各司其职，又完美配合：

| 模块 | 核心职责 | 典型能力 / 方法 |
| :--- | :--- | :--- |
| **`colgen`** | **代码生成** | 解析 GORM 模型生成强类型列定义。支持自定义输出目录、包名及 Dry-run 校验。 |
| **`query`** | **动态查询构建** | **Builder**: `Where`, `Select`, `Joins`, `Preload`, `Page`, `Clone`, `Apply`... <br>**Column**: `Eq`, `Gt`, `Like`, `In`, `Between`, `Sum`, `Asc`... |
| **`repo`** | **泛型仓储** | 提供通用 CRUD：`Create`, `Find`, `First`, `Update`, `Delete`, `Count`, `Pluck`... |
| **`db`** | **上下文事务管理** | 提供 `db.Client`，支持通过 `context.Context` 无感传递事务连接，避免手动透传 DB。 |

**最佳上手路径：** `colgen` 生成列定义 ➔ 使用 `columns.Xxx` 构建 `query` ➔ 传入 `repo` 执行。

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

创建一个简单的生成脚本（例如放在 `cmd/gen/main.go`），将你需要生成查询列定义的模型传入生成器：

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

默认情况下，`colgen` 会把生成结果写入独立的 `columns` 包中。

> **💡 进阶提示：** 你也可以在任意 Go 文件的头部添加 `//go:generate go run cmd/gen/main.go`，之后就能通过在项目根目录运行 `go generate ./...` 将其无缝接入你的标准工作流。

### 3. 享受丝滑的强类型查询

现在，你可以使用生成的 `columns.User` 配合 Query Builder 进行类型安全的查询了：

```go
import (
    "gorm.io/gorm"

    "your_project_name/model/columns"
    "your_project_name/model"
    "github.com/im-wmkong/gorm-query/query"
)

var db *gorm.DB

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

`db.Client` 是 Service 层和 Repository 层之间的连接点：

- `db.Client` 同时实现了 `db.DBProvider` 和 `db.Transactor`
- Repository 只依赖 `db.DBProvider`
- Service 只依赖 `db.Transactor`
- 当前事务通过 `context.Context` 向下传递，因此你不需要手动传递 `*gorm.DB`

为了让示例更紧凑，下面的代码直接接收 `*db.Client`，因为它本身就同时实现了这两个接口。

**在 Service 中直接使用泛型 Repository：**
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

**把业务逻辑包进一个事务：**
```go
// Service 只需要在这里开启一次事务。
// Repository 会自动从 txCtx 中拿到同一个事务连接。
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

由于两个 Repository 调用拿到的是同一个 `txCtx`，它们会自动运行在同一个事务里。只要其中任意一步返回错误，GORM 就会帮你整体回滚。

如果后续你需要自定义 Repository 方法，也可以再把 `repo.New[T](dbClient)` 包装成你自己的 Repository 类型，而不需要改变这套事务模型。

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

当 Service 已经依赖泛型 Repository 时，`query.Builder` 就成了补齐“动态过滤能力”的那块拼图。

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
    users, err := s.users.Find(ctx, qb)
    if err != nil {
        return nil, err
    }
    
    return users, nil
}
```

### 3. 查询条件的复用 (防污染)

当你需要从同一个基础 builder 派生多个查询时，请使用 `.Clone()`。

它可以防止底层切片共享带来的条件污染：

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
