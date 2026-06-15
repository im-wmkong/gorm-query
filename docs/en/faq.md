# FAQ & pitfalls

## 1. What happens when `Update` is called with empty assigns?

`BaseRepository.Update` returns `(0, query.ErrNoAssignment)` when `len(assigns) == 0`. An empty assignment list is treated as a programming error rather than a silent no-op: it surfaces caller bugs early and prevents the GORM trap of "empty map → full-table update". Always use `Set(...)` to express the intended change, and check it with `errors.Is(err, query.ErrNoAssignment)`.

## 2. Why does my Builder reference not change after chaining?

`Builder` is **immutable**: every `Where/Or/Select/...` returns a brand-new Builder; the original and any derived builders are independent and safe to read from multiple goroutines. See [Query Builder](query-builder.md).

## 3. How do I write multi-level Preload?

```go
schema.User.Query().Preload(
    schema.User.Profile.Nested(schema.Profile.Address),
)
```

`Nested` is checked at compile time: parent/child association types must align.

## 4. `Joins` vs `Preload`?

| Aspect | `Joins` / `InnerJoins` | `Preload` |
| :--- | :--- | :--- |
| SQL | one statement, JOINed and flattened | 1+N (one for the parent, one IN-query per association batch) |
| Use case | filtering / ordering by joined columns | reading associated data |
| Slice associations | discouraged (cartesian product) | preferred |

## 5. Why is `Having` still string-based?

The current `Having(expr string, args ...any)` forwards directly to GORM. A typed version is on the roadmap. Until then, never concatenate untrusted input into the expression.

## 6. Is `RawFragment` safe?

**No**. `RawFragment` renders the string verbatim with **no parameter binding**. Use it only for compile-time-constant fragments (e.g. `FIELD(status, 1, 0, 2)`) and never combine it with user input.

## 7. How do I use a feature the Builder doesn't cover yet?

Reach for `Repository.DB(ctx)` to get the transaction-aware `*gorm.DB`:

```go
err := r.DB(ctx).
    Clauses(clause.OnConflict{DoNothing: true}).
    Create(&u).Error

err := r.DB(ctx).
    Clauses(clause.Locking{Strength: "UPDATE"}).
    Where("id = ?", id).
    First(&u).Error
```

Or embed raw GORM logic inside a Builder via `Scope(func(*gorm.DB) *gorm.DB)`.

## 8. Known feature gaps

- No type-safe subqueries / `EXISTS` / `IN (SELECT ...)`.
- No typed `OnConflict` / `Upsert` / `Returning` / `FOR UPDATE`.
- `Having` and `RawFragment` are still strings.
- Integration tests run on SQLite only; verify MySQL / Postgres dialect specifics yourself.

See the repository issues for the roadmap.

## 9. `schemagen` says "all models must be in the same package"?

A single `schemagen.Generate(...)` call requires every model to live in one Go package (see the [schemagen limitations](schemagen.md#6-limitations--gotchas)). If your models are spread across packages, call `Generate` per package or write a separate `cmd/gen` for each.

## 10. What does `repo.First / Take / Last` return when no rows match?

`(nil, gorm.ErrRecordNotFound)`, matching GORM. Use `errors.Is(err, gorm.ErrRecordNotFound)` to branch.

## 11. How do I hard-delete a soft-deleted model?

```go
users.Delete(ctx, schema.User.Query().Unscoped().Where(schema.User.ID.Eq(1)))
```

## 12. Which doc set wins when English/Chinese disagree?

[`docs/`](../README.md) is authoritative. `README.md` and `README_CN.md` are intentionally short—just navigation and a minimal getting-started example.
