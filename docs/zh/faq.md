# FAQ 与常见陷阱

## 1. 为什么 `Update` 传空 assigns 不报错？

`BaseRepository.Update` 在 `len(assigns) == 0` 时显式返回 `(0, nil)`，是契约性 no-op，避免误触发"GORM 收到空 map → 全表更新"。需要变更时请总是显式 `Set(...)`。

## 2. 为什么 Builder 链式后原对象没变？

`Builder` 设计为**不可变**：每次 `Where/Or/Select/...` 都返回新的 Builder，原对象与已派生的 Builder 互不影响，可在多个 goroutine 并发只读。详见 [Query Builder](query-builder.md)。

## 3. 多级 Preload 怎么写？

```go
schema.User.Query().Preload(
    schema.User.Profile.Nested(schema.Profile.Address),
)
```

`Nested` 在编译期校验：父子关联类型必须能对得上。

## 4. `Joins` 和 `Preload` 有什么区别？

| 维度 | `Joins` / `InnerJoins` | `Preload` |
| :--- | :--- | :--- |
| SQL | 一条 SQL，JOIN 后扁平化 | N+1（实际是 1+N，一次主表 + 一次关联表 IN 查询） |
| 适用 | 需要 JOIN 后过滤 / 排序 | 仅读取关联数据 |
| Slice 关联 | 不推荐（笛卡尔积） | 推荐 |

## 5. 为什么 `Having` 仍然是字符串？

当前实现 `Having(expr string, args ...any)` 直接透传 GORM。后续会引入类型化版本；在那之前请确保表达式不会拼接外部输入。

## 6. `RawFragment` 安全吗？

**不安全**。`RawFragment` 直接把字符串作为 SQL 片段渲染，**没有参数绑定**。仅用于编译期可确定的常量片段（如 `FIELD(status, 1, 0, 2)`），严禁拼接用户输入。

## 7. Builder 暂未覆盖的能力如何使用？

通过 `Repository.DB(ctx)` 拿到事务感知的 `*gorm.DB`：

```go
err := r.DB(ctx).
    Clauses(clause.OnConflict{DoNothing: true}).
    Create(&u).Error

err := r.DB(ctx).
    Clauses(clause.Locking{Strength: "UPDATE"}).
    Where("id = ?", id).
    First(&u).Error
```

或在 Builder 中使用 `Scope(func(*gorm.DB) *gorm.DB)` 嵌入原生表达。

## 8. 当前已知的功能限制

- 不支持类型安全的子查询 / `EXISTS` / `IN (SELECT ...)`。
- 不支持 `OnConflict` / `Upsert` / `Returning` / `FOR UPDATE` 类型化封装。
- `Having` / `RawFragment` 仍是字符串。
- 仅在 SQLite 上做了完整集成测试；MySQL / Postgres 方言相关行为请自行验证。

后续路线图见仓库 Issues。

## 9. schemagen 报 "all models must be in the same package"？

`schemagen.Generate(...)` 单次调用要求所有模型在同一 Go 包内（详见 [schemagen 限制](schemagen.md#6-限制与陷阱)）。如果你的模型分散在多个包，请按包分别调用 `Generate`，或为每个包写一个独立的 `cmd/gen`。

## 10. `repo.First / Take / Last` 找不到记录返回什么？

返回 `(nil, gorm.ErrRecordNotFound)`，与 GORM 一致。建议使用 `errors.Is(err, gorm.ErrRecordNotFound)` 判定。

## 11. 软删除如何"硬删"？

```go
users.Delete(ctx, schema.User.Query().Unscoped().Where(schema.User.ID.Eq(1)))
```

## 12. 中英文文档不一致以哪个为准？

以 [`docs/`](../README.md) 为权威；`README.md` / `README_CN.md` 仅做导航与最小上手示例。
