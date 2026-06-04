# 事务模型

`gorm-query` 的事务模型设计目标：**Service 不依赖 `*gorm.DB`，Repository 不感知事务的存在**，事务通过 `context.Context` 隐式传递。

## 1. 三个核心抽象

```go
// 提供 ctx 对应的 DB 句柄（事务感知）
type DBProvider interface {
    DB(ctx context.Context) *gorm.DB
}

// 在 ctx 上开启事务
type Transactor interface {
    Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// 同时实现两个接口的默认实现
type Client struct{ /* ... */ }
```

| 角色 | 依赖 | 原因 |
| :--- | :--- | :--- |
| Repository | `DBProvider` | 只关心拿连接，不关心事务边界 |
| Service | `Transactor` | 协调事务边界 |

> 💡 因为 `*db.Client` 同时实现两个接口，依赖注入时可以传同一个实例。

## 2. 标准用法

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
        // 同一个 txCtx 自动复用同一个 tx
        return s.profiles.Create(txCtx, p)
    })
}
```

## 3. 传播契约

`db.Client` 内部使用一个**私有**的 ctx key，外部无法手动构造事务 ctx：

- 必须经由 `Transactor.Transaction(...)` 进入。
- 进入后 fn 拿到的 `txCtx` 已经携带事务连接，其下游所有 `DBProvider.DB(txCtx)` / `Repository.*(txCtx, ...)` 都会复用同一个 `*gorm.DB`。
- fn 返回 `error` → GORM 自动回滚；返回 `nil` → 提交。
- 如果 ctx 中**不带**事务 key，`DB(ctx)` 直接走 `db.WithContext(ctx)`，这就是非事务的常规调用。

## 4. 嵌套事务

`Transaction` 内部直接复用 GORM 的事务能力：在外层事务的 `txCtx` 上再次调用 `Transaction(...)` 时，GORM 会基于 SAVEPOINT 实现"嵌套事务"。内层失败只会回滚到 SAVEPOINT，不会终止外层。

```go
client.Transaction(ctx, func(outer context.Context) error {
    _ = client.Transaction(outer, func(inner context.Context) error {
        // 仅回滚到 inner 起点
        return errors.New("rollback inner")
    })
    // 外层仍可以继续提交
    return nil
})
```

## 5. 上下文取消 / 超时

`DB(ctx)` 内部已经 `WithContext(ctx)`，因此可以通过 `context.WithTimeout` / `context.WithCancel` 控制查询：

```go
ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
defer cancel()
users, err := s.users.Find(ctx, qb)
```

## 6. 何时直接拿 `*gorm.DB`

某些 GORM 能力当前 Builder 还未封装（子查询、Upsert、Locking、Returning 等）。需要时可以走 `Repository.DB(ctx)` 拿到事务感知的句柄：

```go
err := r.DB(ctx).
    Clauses(clause.Locking{Strength: "UPDATE"}).
    Where("id = ?", id).
    First(&u).Error
```

事务隔离仍然成立。
