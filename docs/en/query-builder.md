# Query Builder

`query.Builder[T]` is an **immutable**, chainable SQL builder. Chained calls leave the receiver unchanged; no-op calls may return the original Builder.

```go
base   := schema.User.Query().Where(schema.User.Status.Eq(1))
adults := base.Where(schema.User.Age.Gte(18)) // does NOT mutate base
minors := base.Where(schema.User.Age.Lt(18))  // does NOT mutate base
```

A constructed Builder can be reused concurrently when custom Conditions and Scopes support concurrent calls and callers synchronize changes to referenced values and captured state.

## 1. Creating a Builder

```go
qb := schema.User.Query()      // produced by schemagen
qb := query.New[model.User]()  // equivalent
```

## 2. WHERE family

| Method | Semantics | Notes |
| :--- | :--- | :--- |
| `Where(conds...)` | conds are AND-ed | preferred entrypoint |
| `Or(conds...)` | wraps as `OR (...)` nested group | conds inside are still AND |
| `Not(conds...)` | wraps as `NOT (...)` nested group | conds inside are still AND |

```go
qb := schema.User.Query().
    Where(schema.User.Status.Eq(1)).
    Or(
        schema.User.Age.Lt(18),
        schema.User.Age.Gt(60),
    ).
    Not(schema.User.Email.Eq(""))
```

## 3. Column operator cheat sheet

Aligned with `query/column.go`, grouped by column type:

### Available on every column
| Method | Description |
| :--- | :--- |
| `Eq(v)` / `Neq(v)` | `=` / `<>` |
| `In([]v)` / `NotIn([]v)` | `IN` / `NOT IN` |
| `IsNull()` / `IsNotNull()` | `IS NULL` / `IS NOT NULL` |
| `Asc()` / `Desc()` | order fragment |
| `As(alias)` | aliased select fragment |
| `Distinct()` | `DISTINCT <col>` |
| `Sum() / Count() / Avg() / Max() / Min()` | aggregate, returns `AggFragment`, `.As(...)` chainable |
| `Set(v)` | builds an update `Assignment` |
| `WithTable(alias)` | qualified copy for joins / self-references |

### Comparable columns (numeric, string, time.Time)
| Method | Description |
| :--- | :--- |
| `Gt(v)` / `Gte(v)` / `Lt(v)` / `Lte(v)` | `>` / `>=` / `<` / `<=` |
| `Between(lo, hi)` / `NotBetween(lo, hi)` | `BETWEEN` / `NOT BETWEEN` |

### String columns
| Method | Description |
| :--- | :--- |
| `Like(p)` / `NotLike(p)` | raw pattern, `%` is your responsibility |
| `Contains(v)` / `NotContains(v)` | `LIKE %v%` / `NOT LIKE %v%` |
| `HasPrefix(v)` | `LIKE v%` |
| `HasSuffix(v)` | `LIKE %v` |

### Boolean columns
| Method | Description |
| :--- | :--- |
| `IsTrue()` / `IsFalse()` | sugar for `Eq(true)` / `Eq(false)` |

## 4. SELECT / OMIT / DISTINCT

`Select / Omit / Distinct` accept `SQLFragment`, so plain columns and aggregate / aliased fragments are interchangeable.

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

> ⚠️ `Having(expr string, args ...any)` is currently a **string** with no compile-time column safety—double-check spelling.

## 6. ORDER / LIMIT / OFFSET / Page

```go
schema.User.Query().
    Order(schema.User.CreatedAt.Desc()).
    Limit(50).
    Offset(100)

// Page convenience: page < 1 becomes 1, pageSize < 1 becomes 10.
schema.User.Query().Page(2, 20) // LIMIT 20 OFFSET 20
```

## 7. Associations: Preload / Joins / InnerJoins

`Preload / Joins / InnerJoins` operate on `Association[Parent, Child]` values. `Parent` must equal the Builder's entity type `T`—**checked at compile time**:

```go
schema.User.Query().Preload(schema.Order.Items) // ❌ does not compile
```

### Preload (avoid N+1)

```go
schema.User.Query().Preload(schema.User.Profile)

// Multi-level
schema.User.Query().Preload(
    schema.User.Profile.Nested(schema.Profile.Address),
)

// Conditional preload
schema.User.Query().Preload(
    schema.User.Profile,
    schema.Profile.Bio.Eq("SF"),
)
```

### Joins / InnerJoins (SQL JOIN)

```go
schema.User.Query().Joins(schema.User.Profile)        // LEFT JOIN
schema.User.Query().InnerJoins(schema.User.Profile)   // INNER JOIN

// ON conditions use GORM's association alias
schema.User.Query().Joins(
    schema.User.Profile,
    schema.Profile.Bio.WithTable("Profile").Eq("SF"),
)
```

## 8. Unscoped / Scope (escape hatches)

```go
// Disable default scopes (e.g. soft delete)
schema.User.Query().Unscoped().Where(schema.User.ID.Eq(1))

// Inject raw GORM logic
activeOnly := func(db *gorm.DB) *gorm.DB { return db.Where("status = ?", 1) }
schema.User.Query().Scope(activeOnly)
```

## 9. RawFragment: typed columns can't express it

`RawFragment` renders the string verbatim—**no parameter binding**:

```go
qb.Order(query.RawFragment("FIELD(status, 1, 0, 2)"))
qb.Select(query.RawFragment("COUNT(DISTINCT email) AS uniq"))
```

> ⚠️ Never concatenate untrusted input—SQL injection risk.

## 10. Apply: lower a Builder onto `*gorm.DB`

```go
var users []model.User
err := qb.Apply(db.Model(&model.User{})).Find(&users).Error
```

> 💡 Prefer handing the Builder directly to `repo.Repository[T]`; the repo handles `Model(...)` for you. See [Repository](repository.md).

## 11. Composition and SQL fragment boundaries

Builders snapshot the input containers for IN/NOT IN, Or/Not/association conditions, Scope lists and Having arguments. Mutating the original slice does not change a built query. Pointed-to objects, maps and captured closure state are not deep-copied; callers own the concurrency safety of custom functions. Empty operations may return the original Builder: immutability guarantees behavior, not pointer identity.

`SQLFragment.SQL()` returns raw SQL text. Select, Order and Group evaluate it when accepting the fragment and pass the text to GORM. Column As/Asc/Desc/Distinct and aggregate helpers only compose SQL text; they do not automatically quote identifiers for different dialects.

`WithTable(alias)` retains the concrete column type and its operators. JOIN conditions must use GORM's association alias, for example `schema.Profile.Bio.WithTable("Profile").Eq("SF")`; Preload conditions use the target physical table. The Builder passes association paths and conditions directly to GORM without inferring aliases, rewriting columns or adding parent JOINs. Call Joins separately when a nested JOIN needs independent parent and child conditions; use GORM through Scope for complex associations.

`AggFragment` remains a string type. `AggFragment("SUM(age)")`, `string(agg)` and `column.Sum().As("total")` retain their existing behavior. Use native GORM expressions through Scope for reserved identifiers or special SQL expressions:

```go
qb := schema.User.Query().Scope(func(db *gorm.DB) *gorm.DB {
    column := clause.Column{Table: schema.User.Age.Table(), Name: schema.User.Age.Name()}
    return db.Select("SUM(?) AS ?", column, clause.Column{Name: "total"})
})
```

GORM quotes `clause.Column` using its dialect. The SQL fragment layer adds no rendering protocol.
