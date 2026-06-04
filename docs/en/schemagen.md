# Code generator (schemagen)

`schemagen` parses GORM models into a strongly-typed schema dictionary, emitting `<table>_gen.go` files that the Builder consumes directly.

## 1. Minimal usage

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

Defaults:
- output dir: `./schema`
- package name: inferred from the output dir (`schema` → `package schema`)
- naming strategy: `schema.NamingStrategy{SingularTable: true}`
- file name: `User` model → `user_gen.go`

## 2. Integrating with `go generate`

```go
//go:generate go run ./cmd/gen
package model
```

```bash
go generate ./...
```

## 3. Options cheat sheet

Defined in `schemagen/options.go`:

| Option | Effect | Default |
| :--- | :--- | :--- |
| `WithOutputDir(dir)` | Output directory | `"schema"` |
| `WithPackageName(name)` | Explicit package name | inferred from output dir |
| `WithNamingStrategy(ns)` | GORM naming strategy | `{SingularTable: true}` |
| `WithDryRun(true)` | Verify mode: do not write files; check existing files match the new content | `false` |
| `WithLogger(logger)` | Custom logger; `nil` becomes `NopLogger()` | `DefaultLogger()` |

```go
g := schemagen.New(
    schemagen.WithOutputDir("internal/dal/schema"),
    schemagen.WithPackageName("schema"),
    schemagen.WithDryRun(true),
)
```

## 4. Column type mapping

The generator picks the closest Column type for each Go field:

| Go type | Generated Column |
| :--- | :--- |
| `string` / custom `~string` | `query.StringColumn[T]` |
| integers / floats (`Numeric`) | `query.NumericColumn[T]` |
| `time.Time` | `query.TimeColumn` |
| `bool` | `query.BoolColumn` |
| Other (`gorm.DeletedAt`, custom structs, …) | `query.ValueColumn[T]` |

`ValueColumn[T]` only exposes the universal operators (`Eq/Neq/In/NotIn/IsNull/...`); ordered / string operators are unavailable.

## 5. Association detection

GORM associations (`HasOne` / `HasMany` / `BelongsTo` / `Many2Many`) become `query.Association[Parent, Child]`, usable with:

- `Builder.Preload(...)`
- `Builder.Joins(...)` / `Builder.InnerJoins(...)`
- Multi-level: `schema.User.Profile.Nested(schema.Profile.Address)`

## 6. Limitations & gotchas

- **Single-package requirement**: every model passed in one `Generate(...)` call must live in the same Go package, otherwise you'll see `all models must be in the same package`.
- **Anonymous / unexported types**: skipped with a warning.
- **Empty `DBName`**: fields whose GORM `DBName` is empty (typically explicit `gorm:"-"`) are skipped.
- **Package-name consistency**: if `outputDir` already contains Go files whose package differs from the target, generation fails. Set `WithPackageName(...)` explicitly to force it.
- **Dry-run**: byte-level comparison against existing files; useful in CI to detect a missed regeneration.

## 7. Generated output example

For:

```go
type User struct {
    gorm.Model
    UserName string         `gorm:"column:user_name"`
    Profile  *model.Profile // HasOne
}
```

The generator produces (excerpt):

```go
type user struct {
    ID        query.NumericColumn[uint]
    CreatedAt query.TimeColumn
    UpdatedAt query.TimeColumn
    DeletedAt query.ValueColumn[gorm.DeletedAt]
    UserName  query.StringColumn[string]
    Profile   query.Association[model.User, model.Profile]
}

var User = user{ /* each column initialised with ("users", "<col>") */ }

func (s *user) Query() *query.Builder[model.User] {
    return query.New[model.User]()
}
```

Full runnable example: [`example/cmd/schemagen/main.go`](../../example/cmd/schemagen/main.go) and [`example/model/schema/`](../../example/model/schema/).
