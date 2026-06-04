# Query Builder

`query.Builder[T]` 是一个**不可变（immutable）**、可链式调用的 SQL 构建器。每一次 `Where / Or / Select / ...` 都会返回一个**新的** Builder，原对象保持不变。

```go
base   := schema.User.Query().Where(schema.User.Status.Eq(1))
adults := base.Where(schema.User.Age.Gte(18)) // 不会修改 base
minors := base.Where(schema.User.Age.Lt(18))  // 不会修改 base
```

因此一个完成构建的 Builder 也可以在多个 goroutine 间并发只读。

## 1. 创建 Builder

```go
qb := schema.User.Query()      // 由 schemagen 生成
qb := query.New[model.User]()  // 等价
```

## 2. WHERE 系列

| 方法 | SQL 语义 | 备注 |
| :--- | :--- | :--- |
| `Where(conds...)` | 多个条件之间为 AND | 推荐主入口 |
| `Or(conds...)` | 整体作为 `OR (...)` 嵌套 | conds 之间仍为 AND |
| `Not(conds...)` | 整体作为 `NOT (...)` 嵌套 | conds 之间仍为 AND |

```go
qb := schema.User.Query().
    Where(schema.User.Status.Eq(1)).
    Or(
        schema.User.Age.Lt(18),
        schema.User.Age.Gt(60),
    ).
    Not(schema.User.Email.Eq(""))
```

## 3. 列运算符速查

下表对照 `query/column.go`，按列类型分组列出全部已实现的运算符：

### 通用（所有列）
| 方法 | 说明 |
| :--- | :--- |
| `Eq(v)` / `Neq(v)` | `=` / `<>` |
| `In([]v)` / `NotIn([]v)` | `IN` / `NOT IN` |
| `IsNull()` / `IsNotNull()` | `IS NULL` / `IS NOT NULL` |
| `Asc()` / `Desc()` | 排序片段 |
| `As(alias)` | 列别名 `<col> AS <alias>` |
| `Distinct()` | `DISTINCT <col>` |
| `Sum() / Count() / Avg() / Max() / Min()` | 聚合，返回 `AggFragment`，可继续 `.As(...)` |
| `Set(v)` | 生成更新用的 `Assignment` |
| `WithTable(alias)` | 返回带表/别名限定的列副本（用于 JOIN/自连接） |

### 可比较列（数字、字符串、time.Time）
| 方法 | 说明 |
| :--- | :--- |
| `Gt(v)` / `Gte(v)` / `Lt(v)` / `Lte(v)` | `>` / `>=` / `<` / `<=` |
| `Between(lo, hi)` / `NotBetween(lo, hi)` | `BETWEEN` / `NOT BETWEEN` |

### 字符串列
| 方法 | 说明 |
| :--- | :--- |
| `Like(p)` / `NotLike(p)` | 原始模式串，自行加 `%` |
| `Contains(v)` / `NotContains(v)` | `LIKE %v%` / `NOT LIKE %v%` |
| `HasPrefix(v)` | `LIKE v%` |
| `HasSuffix(v)` | `LIKE %v` |

### 布尔列
| 方法 | 说明 |
| :--- | :--- |
| `IsTrue()` / `IsFalse()` | `Eq(true)` / `Eq(false)` 的语义糖 |

## 4. SELECT / OMIT / DISTINCT

`Select` / `Omit` / `Distinct` 都接收 `SQLFragment`，因此既能传普通列，也能传聚合 / 别名片段。

```go
schema.User.Query().Select(
    schema.User.ID,
    schema.User.UserName.As("name"),
    schema.User.Age.Sum().As("total_age"),
)

schema.User.Query().Omit(schema.User.Password)
schema.User.Query().Distinct(schema.User.Email)
```

## 5. GROUP BY / HAVING

```go
schema.User.Query().
    Select(schema.User.Status, schema.User.ID.Count().As("cnt")).
    Group(schema.User.Status).
    Having("COUNT(*) > ?", 10)
```

> ⚠️ `Having(expr string, args ...any)` 当前是**字符串**，列名无类型保护，自行确保拼写正确。

## 6. ORDER / LIMIT / OFFSET / Page

```go
schema.User.Query().
    Order(schema.User.CreatedAt.Desc()).
    Limit(50).
    Offset(100)

// 分页便捷方法：page < 1 视为 1，pageSize < 1 视为 10
schema.User.Query().Page(2, 20) // LIMIT 20 OFFSET 20
```

## 7. 关联：Preload / Joins / InnerJoins

`Preload / Joins / InnerJoins` 都基于 `Association[Parent, Child]`。`Parent` 必须等于 Builder 的实体类型 `T`，**编译期校验**：

```go
schema.User.Query().Preload(schema.Order.Items) // ❌ 编译报错
```

### Preload（避免 N+1）

```go
schema.User.Query().Preload(schema.User.Profile)

// 多级嵌套
schema.User.Query().Preload(
    schema.User.Profile.Nested(schema.Profile.Address),
)

// 携带过滤条件的 Preload
schema.User.Query().Preload(
    schema.User.Profile,
    schema.Profile.City.Eq("SF"),
)
```

### Joins / InnerJoins（SQL JOIN）

```go
schema.User.Query().Joins(schema.User.Profile)        // LEFT JOIN
schema.User.Query().InnerJoins(schema.User.Profile)   // INNER JOIN

// 携带 ON 条件
schema.User.Query().Joins(
    schema.User.Profile,
    schema.Profile.City.Eq("SF"),
)
```

## 8. Unscoped / Scope（兜底入口）

```go
// 关闭软删除等默认 scope
schema.User.Query().Unscoped().Where(schema.User.ID.Eq(1))

// 接入原生 GORM 表达
activeOnly := func(db *gorm.DB) *gorm.DB { return db.Where("status = ?", 1) }
schema.User.Query().Scope(activeOnly)
```

## 9. RawFragment：当类型化无法表达时

`RawFragment` 直接渲染字符串，**不做参数绑定**：

```go
qb.Order(query.RawFragment("FIELD(status, 1, 0, 2)"))
qb.Select(query.RawFragment("COUNT(DISTINCT email) AS uniq"))
```

> ⚠️ 严禁拼接来自外部输入的字符串，存在 SQL 注入风险。

## 10. Apply：把 Builder 落到 `*gorm.DB`

```go
var users []model.User
err := qb.Apply(db.Model(&model.User{})).Find(&users).Error
```

> 💡 推荐把 Builder 直接交给 `repo.Repository[T]`，由 repo 自动处理 `Model(...)`，参见 [Repository](repository.md)。
