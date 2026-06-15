# GORM Query 🚀

[![Go Reference](https://pkg.go.dev/badge/github.com/im-wmkong/gorm-query.svg)](https://pkg.go.dev/github.com/im-wmkong/gorm-query)
[![Go Report Card](https://goreportcard.com/badge/github.com/im-wmkong/gorm-query)](https://goreportcard.com/report/github.com/im-wmkong/gorm-query)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

**GORM Query** 是一个基于 GORM 的强类型查询构建器与通用仓储（Repository）库。

通过**代码生成**消除了 GORM 查询中脆弱的"魔法字符串"，提供丝滑的链式调用体验。同时内置泛型 Repository 与基于 Context 的事务管理方案，帮助你写出更整洁的 Service 与数据访问层代码。

## ✨ 核心特性

- 🛡️ **强类型查询构建** —— 告别 `db.Where("age > ?", 18)`，拥抱 `schema.User.Age.Gt(18)`，编译期拦截字段名拼写错误。
- 📦 **开箱即用的泛型仓储** —— 一行代码即可拥有完整 CRUD：`repo.BaseRepository[T]`。
- 🎯 **告别臃肿的 Repository** —— 用 `query.Builder` 动态组合查询条件，不再为每个业务场景编写 `FindByXxx`。
- 🔄 **隐式上下文事务** —— 通过 `context.Context` 传递事务，Service 与 Repo 层与 `*gorm.DB` 解耦。
- 🧊 **不可变、并发只读安全的 Builder** —— 每次链式调用都返回新的 Builder，派生查询互不影响。

## 📦 安装

```bash
go get github.com/im-wmkong/gorm-query
```

## 🗺️ 能力总览

| 模块 | 核心职责 | 关键能力 |
| :--- | :--- | :--- |
| **`schemagen`** | 代码生成 | 把 GORM 模型解析为强类型 schema 字典（列 + 关联）。Options：`WithOutputDir` / `WithPackageName` / `WithNamingStrategy` / `WithDryRun` / `WithLogger`。 |
| **`query`** | 动态查询构建 | **Builder**：`Where / Or / Not / Select / Omit / Distinct / Preload / Joins / InnerJoins / Group / Having / Order / Page / Limit / Offset / Unscoped / Scope / Apply` <br>**Column**：`Eq / Neq / Gt / Gte / Lt / Lte / In / NotIn / Between / NotBetween / Like / Contains / HasPrefix / HasSuffix / IsNull / IsNotNull / Sum / Count / Avg / Max / Min / As / Distinct / Set / Asc / Desc / WithTable` <br>**Association**：`Preload / Joins / Nested` |
| **`repo`** | 泛型仓储 | `Save / Create / CreateInBatches / Update / Delete / Find / First / Take / Last / Count / Exists / Pluck / DB(ctx)` |
| **`db`** | 上下文事务 | `Client` 同时实现 `DBProvider` 与 `Transactor`，事务通过 `context.Context` 隐式传递。 |

→ 完整 API 参考：[pkg.go.dev](https://pkg.go.dev/github.com/im-wmkong/gorm-query)；模块深度文档见 [`docs/`](docs/README.md)。

## 🚀 快速开始

### 1. 定义模型

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

### 2. 生成 schema

新建一个生成脚本（例如 `cmd/gen/main.go`）：

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

执行后会在 `model/schema/user_gen.go` 生成 `schema.User` 变量。

> **💡 提示**：可以在任意 Go 文件顶部加 `//go:generate go run cmd/gen/main.go`，之后通过 `go generate ./...` 触发生成。

→ 全部选项、命名规则与限制见 [docs/zh/schemagen.md](docs/zh/schemagen.md)。

### 3. 类型安全地构建查询

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

// 应用到 *gorm.DB
var users []model.User
err := qb.Apply(db.Model(&model.User{})).Find(&users).Error
```

→ 完整运算符速查：[docs/zh/query-builder.md](docs/zh/query-builder.md)。

### 4. 也可以直接交给 Repository

```go
client := db.NewClient(gormDB)
users  := repo.New[model.User](client)

list, err := users.Find(ctx,
    schema.User.Query().Where(schema.User.Status.Eq(1)),
)
```

→ Repository 全量方法见 [docs/zh/repository.md](docs/zh/repository.md)。

## 💡 常用模式

### 基于 Context 的事务

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

→ 传播契约、嵌套事务、超时取消见 [docs/zh/transaction.md](docs/zh/transaction.md)。

### 动态查询：Repository 不再臃肿

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

### 类型安全的 Update

```go
qb := schema.User.Query().Where(schema.User.ID.Eq(1))
rows, err := s.users.Update(ctx, qb,
    schema.User.Status.Set(2),
    schema.User.Email.Set("a@b.com"),
)
```

### Preload + 嵌套关联

```go
schema.User.Query().Preload(
    schema.User.Profile.Nested(schema.Profile.Address),
    schema.Address.City.Eq("SF"),
)
```

### 兜底逃生口

```go
// Builder：嵌入原生 GORM 逻辑
qb.Scope(func(db *gorm.DB) *gorm.DB { return db.Where("status = ?", 1) })

// Repository：拿到事务感知的 *gorm.DB
err := r.DB(ctx).
    Clauses(clause.OnConflict{DoNothing: true}).
    Create(&u).Error
```

## 📚 文档

完整可运行的示例（SQLite + schemagen + repository + service + 测试）见 [`example/`](example)。

| 主题 | 文档 |
| :--- | :--- |
| Query Builder | [docs/zh/query-builder.md](docs/zh/query-builder.md) |
| Repository | [docs/zh/repository.md](docs/zh/repository.md) |
| 事务模型 | [docs/zh/transaction.md](docs/zh/transaction.md) |
| 代码生成器 | [docs/zh/schemagen.md](docs/zh/schemagen.md) |
| 常见问题 | [docs/zh/faq.md](docs/zh/faq.md) |

→ 中英文索引：[docs/README.md](docs/README.md)。

## 🚧 当前限制

- 暂不支持类型安全的子查询 / `EXISTS` / `IN (SELECT ...)`。
- 暂不支持类型安全的 `OnConflict` / `Upsert` / `Returning` / `FOR UPDATE`。
- `Having(...)` 与 `RawFragment` 仍是字符串，必要时请走 `repo.DB(ctx)` 兜底。

## 🤝 参与贡献

欢迎提交 Issue 与 Pull Request。提交前请运行：

```bash
make tidy
make generate
make test
```

> 修改 `README_CN.md` 时请同步更新 `README.md`，反之亦然。

## 📄 开源协议

[MIT License](LICENSE)。
