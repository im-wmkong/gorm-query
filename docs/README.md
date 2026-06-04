# GORM Query Documentation

Complete documentation for `gorm-query`. The README focuses on getting you started; this directory dives into each module in depth.

> *[中文文档 / Chinese version](#中文文档)*

## English

| Topic | Description |
| :--- | :--- |
| [Query Builder](en/query-builder.md) | All Builder methods, column operators, Preload / Joins, raw fragments |
| [Repository](en/repository.md) | Generic repository methods, assignments, custom repository pattern |
| [Transaction](en/transaction.md) | `db.Client`, context propagation, `DBProvider` / `Transactor` |
| [Schemagen](en/schemagen.md) | Code generator options, generation rules, limitations |
| [FAQ & Pitfalls](en/faq.md) | Common pitfalls, current limitations, interop with raw GORM |

## 中文文档

| 主题 | 说明 |
| :--- | :--- |
| [查询构建器](zh/query-builder.md) | Builder 全量方法、列运算符、Preload / Joins、Raw 片段 |
| [泛型 Repository](zh/repository.md) | Repository 全量方法、Assignment、自定义扩展 |
| [事务模型](zh/transaction.md) | `db.Client`、Context 传递、`DBProvider` / `Transactor` |
| [代码生成器](zh/schemagen.md) | schemagen 选项、生成规则与限制 |
| [常见问题](zh/faq.md) | 常见陷阱、当前限制、与原生 GORM 混用 |

## Runnable example

A full runnable example lives in [`example/`](../example): SQLite + schemagen + repository + service + tests.

```bash
go run ./example
```
