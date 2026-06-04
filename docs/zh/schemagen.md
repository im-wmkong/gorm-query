# 代码生成器（schemagen）

`schemagen` 把 GORM 模型解析为强类型 schema 字典，生成 `<table>_gen.go` 文件，供 Builder 直接引用。

## 1. 最小用法

```go
// cmd/gen/main.go
package main

import (
    "log"

    "your_project/model"
    "github.com/im-wmkong/gorm-query/schemagen"
)

func main() {
    if err := schemagen.New().Generate(&model.User{}, &model.Profile{}); err != nil {
        log.Fatal(err)
    }
}
```

```bash
go run cmd/gen/main.go
```

默认行为：
- 输出目录：`./schema`
- 包名：从输出目录名推断（如 `schema` → `package schema`）
- 命名策略：`schema.NamingStrategy{SingularTable: true}`
- 文件名：模型 `User` → `user_gen.go`

## 2. 配合 `go generate`

```go
//go:generate go run ./cmd/gen
package model
```

```bash
go generate ./...
```

## 3. Options 速查

定义见 `schemagen/options.go`：

| Option | 作用 | 默认值 |
| :--- | :--- | :--- |
| `WithOutputDir(dir)` | 输出目录 | `"schema"` |
| `WithPackageName(name)` | 显式指定包名 | 从输出目录推断 |
| `WithNamingStrategy(ns)` | GORM 命名策略 | `{SingularTable: true}` |
| `WithDryRun(true)` | 校验模式：不写文件，只检查现有文件是否与新生成内容一致 | `false` |
| `WithLogger(logger)` | 自定义日志；传 `nil` 等价于 `NopLogger()` | `DefaultLogger()` |

```go
g := schemagen.New(
    schemagen.WithOutputDir("internal/dal/schema"),
    schemagen.WithPackageName("schema"),
    schemagen.WithDryRun(true),
)
```

## 4. 列类型映射

代码生成时会按 Go 字段类型选择最贴近的 Column 类型：

| Go 类型 | 生成的 Column |
| :--- | :--- |
| `string` / 自定义 `~string` | `query.StringColumn[T]` |
| 整型 / 浮点型（`Numeric`） | `query.NumericColumn[T]` |
| `time.Time` | `query.TimeColumn` |
| `bool` | `query.BoolColumn` |
| 其他（`gorm.DeletedAt`、自定义结构体等） | `query.ValueColumn[T]` |

`ValueColumn[T]` 仅暴露通用运算符（`Eq/Neq/In/NotIn/IsNull/...`），可比较 / 字符串特有运算符不可用。

## 5. 关联识别

字段为 GORM 关联（`HasOne` / `HasMany` / `BelongsTo` / `Many2Many`）时，生成 `query.Association[Parent, Child]`，可用于：

- `Builder.Preload(...)`
- `Builder.Joins(...)` / `Builder.InnerJoins(...)`
- 多级嵌套：`schema.User.Profile.Nested(schema.Profile.Address)`

## 6. 限制与陷阱

- **同包要求**：传给一次 `Generate(...)` 的所有模型必须位于同一 Go 包，否则会返回 "all models must be in the same package"。
- **匿名 / 内置类型**：未导出或匿名结构体会被跳过并打印 warning。
- **DBName 为空**：GORM 解析后 `DBName == ""` 的字段（一般是显式 `gorm:"-"`）会跳过。
- **包名一致性**：若 `outputDir` 已经包含 Go 文件且其包名与目标包名不一致，会直接报错；遇到此情况显式 `WithPackageName(...)` 即可。
- **dry-run**：仅与现有文件做字节级比对；CI 中可作为"是否忘记 regen"的检查。

## 7. 生成结果示例

针对：

```go
type User struct {
    gorm.Model
    UserName string         `gorm:"column:user_name"`
    Profile  *model.Profile // HasOne
}
```

生成（节选）：

```go
type user struct {
    ID        query.NumericColumn[uint]
    CreatedAt query.TimeColumn
    UpdatedAt query.TimeColumn
    DeletedAt query.ValueColumn[gorm.DeletedAt]
    UserName  query.StringColumn[string]
    Profile   query.Association[model.User, model.Profile]
}

var User = user{ /* 各列以 ("users", "<col>") 初始化 */ }

func (s *user) Query() *query.Builder[model.User] {
    return query.New[model.User]()
}
```

完整可运行示例：[`example/cmd/schemagen/main.go`](../../example/cmd/schemagen/main.go) 与 [`example/model/schema/`](../../example/model/schema/)。
