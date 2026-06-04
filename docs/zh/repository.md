# Repository

`repo.BaseRepository[T]` 是基于 `db.DBProvider` 的泛型 CRUD 实现。它会从 `context.Context` 自动获取当前的事务连接，因此调用方无需感知是否处于事务中。

## 1. 创建 Repository

```go
// dbClient 同时实现 db.DBProvider / db.Transactor
client := db.NewClient(gormDB)

// 直接得到一个泛型 repo
users := repo.New[model.User](client)
```

## 2. 接口与方法全集

`Repository[T]` 接口定义见 `repo/repository.go`：

| 方法 | 说明 |
| :--- | :--- |
| `DB(ctx)` | 直接拿到当前 ctx 对应的 `*gorm.DB`（兜底逃生口） |
| `Save(ctx, entity)` | upsert：依据主键决定 INSERT 或 UPDATE |
| `Create(ctx, entity)` | 仅 INSERT |
| `CreateInBatches(ctx, entities, batchSize) (int64, error)` | 分批插入，返回实际写入行数 |
| `Update(ctx, qb, assigns...) (int64, error)` | 按 Builder 条件更新；`len(assigns) == 0` 时返回 `(0, nil)` |
| `Delete(ctx, qb) (int64, error)` | 按 Builder 条件删除 |
| `Find(ctx, qb) ([]*T, error)` | 列表查询 |
| `First(ctx, qb) (*T, error)` | 主键升序第一条；未命中返回 `(nil, gorm.ErrRecordNotFound)` |
| `Take(ctx, qb) (*T, error)` | 任意一条；未命中同上 |
| `Last(ctx, qb) (*T, error)` | 主键降序第一条 |
| `Count(ctx, qb) (int64, error)` | 计数 |
| `Exists(ctx, qb) (bool, error)` | `SELECT 1 LIMIT 1` 判断存在 |
| `Pluck(ctx, qb, col, dest)` | 取单列到 `dest`（切片或 scalar） |

## 3. 写入：Save / Create / CreateInBatches

```go
// 仅 INSERT（无主键 → 新建）
users.Create(ctx, &model.User{UserName: "Alice"})

// upsert（有主键 → UPDATE，否则 INSERT）
users.Save(ctx, &model.User{ID: 1, UserName: "Alice"})

// 批量插入
list := []*model.User{{UserName: "A"}, {UserName: "B"}}
rows, err := users.CreateInBatches(ctx, list, 100)
```

## 4. 更新：`Update + Set`

每一列都暴露了 `Set(v)` 方法，返回类型化 `Assignment`：

```go
qb := schema.User.Query().Where(schema.User.ID.Eq(1))
rows, err := users.Update(ctx, qb,
    schema.User.Status.Set(2),
    schema.User.Email.Set("a@b.com"),
)
```

> 💡 `Update` 内部使用 `Updates(map[string]any)`，因此 GORM 不会写入零值的"歧义场景"。需要写零值时显式 `Set(0)` / `Set("")` 即可。

## 5. 查询：First / Take / Last / Find / Count / Exists / Pluck

```go
// 单条
u, err := users.First(ctx, schema.User.Query().Where(schema.User.Email.Eq("a@b.com")))
if errors.Is(err, gorm.ErrRecordNotFound) { ... }

// 列表
list, err := users.Find(ctx,
    schema.User.Query().
        Where(schema.User.Status.Eq(1)).
        Order(schema.User.CreatedAt.Desc()),
)

// 计数 / 存在
n, _   := users.Count(ctx, qb)
ok, _  := users.Exists(ctx, qb)

// 单列：Pluck 接收任何 SQLFragment
var emails []string
_ = users.Pluck(ctx,
    schema.User.Query().Where(schema.User.Status.Eq(1)),
    schema.User.Email,
    &emails,
)

// 聚合（标量）
var total int64
_ = users.Pluck(ctx,
    schema.User.Query(),
    schema.User.Age.Sum(),
    &total,
)
```

## 6. 删除

```go
rows, err := users.Delete(ctx, schema.User.Query().Where(schema.User.Status.Eq(0)))
```

如果模型带 `gorm.DeletedAt`，这会执行**软删除**；需要硬删除请在 Builder 上加 `Unscoped()`：

```go
users.Delete(ctx, schema.User.Query().Unscoped().Where(schema.User.ID.Eq(1)))
```

## 7. 自定义 Repository 扩展

`BaseRepository` 是普通结构体，内嵌即可扩展：

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

完整示例见 [`example/repository/user_repo.go`](../../example/repository/user_repo.go)。

## 8. 兜底逃生：`DB(ctx)`

当遇到 Builder 暂未覆盖的能力（子查询、Upsert、`FOR UPDATE` 等），可以直接拿到 `*gorm.DB`：

```go
err := r.DB(ctx).
    Clauses(clause.OnConflict{DoNothing: true}).
    Create(&user).Error
```

仍然会自动复用 ctx 中的事务连接。
