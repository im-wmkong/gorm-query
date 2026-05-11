// Package query provides the core of a type-safe GORM query builder.
//
// It builds queries dynamically via the functional Condition pattern and supports
// deep-copying (Clone) to avoid condition slice aliasing across derived queries.
// It is typically used together with the generated schema (e.g. schema.User) to
// provide a smooth, type-safe, chainable SQL building experience.
package query

import (
	"github.com/im-wmkong/gorm-query/internal/gormx"
	"gorm.io/gorm"
)

// Condition is the core abstraction: a function that transforms *gorm.DB.
// All WHERE-like conditions are eventually represented as a Condition.
type Condition func(db *gorm.DB) *gorm.DB

// Builder accumulates query conditions for a specific entity type T.
// Builder is NOT concurrency-safe; do not use the same instance in multiple
// goroutines. If you need to reuse conditions concurrently, call Clone to get
// an independent copy.
//
// The type parameter T ties Preload to associations whose Parent is T, so that
// schema.Order.Items cannot be preloaded through a Builder[User].
//
// Example:
//
//	qb := query.New[User]().Where(
//	    schema.User.Status.Eq(1),
//	    schema.User.Age.Gte(18),
//	).Order(schema.User.CreatedAt.Desc()).Page(1, 20)
//
//	session := qb.Apply(db.Model(&User{}))
//	_ = session
type Builder[T any] struct {
	conditions []Condition
}

// New creates a new query builder bound to entity type T.
//
// Example:
//
//	qb := query.New[User]()
//	_ = qb
func New[T any]() *Builder[T] {
	return &Builder[T]{}
}

// Apply applies all accumulated conditions to the given gorm.DB session.
func (b *Builder[T]) Apply(db *gorm.DB) *gorm.DB {
	for _, cond := range b.conditions {
		db = cond(db)
	}
	return db
}

// Clone makes a deep copy of the builder so common conditions can be reused safely.
func (b *Builder[T]) Clone() *Builder[T] {
	conditions := make([]Condition, len(b.conditions))
	copy(conditions, b.conditions)
	return &Builder[T]{conditions: conditions}
}

// Where appends one or more conditions.
func (b *Builder[T]) Where(conds ...Condition) *Builder[T] {
	return b.bind(conds...)
}

// Or appends nested OR conditions.
func (b *Builder[T]) Or(conds ...Condition) *Builder[T] {
	return b.nested(conds, func(db, nested *gorm.DB) *gorm.DB {
		return db.Or(nested)
	})
}

// Not appends nested NOT conditions.
func (b *Builder[T]) Not(conds ...Condition) *Builder[T] {
	return b.nested(conds, func(db, nested *gorm.DB) *gorm.DB {
		return db.Not(nested)
	})
}

// Select sets the SELECT clause. Accepts any mixture of typed columns and
// SQL fragments (Distinct, As, aggregates) because they all satisfy SQLFragment.
func (b *Builder[T]) Select(cols ...SQLFragment) *Builder[T] {
	if len(cols) == 0 {
		return b
	}
	names := SQLFragments(cols).Strings()
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Select(names)
	})
}

// Omit omits columns.
func (b *Builder[T]) Omit(cols ...SQLFragment) *Builder[T] {
	names := SQLFragments(cols).Strings()
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Omit(names...)
	})
}

// Distinct adds DISTINCT to the query.
func (b *Builder[T]) Distinct(cols ...SQLFragment) *Builder[T] {
	args := SQLFragments(cols).Anys()
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Distinct(args...)
	})
}

// Joins adds a raw JOIN clause. args are bound to ? placeholders in sql.
func (b *Builder[T]) Joins(sql string, args ...any) *Builder[T] {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Joins(sql, args...)
	})
}

// Preload preloads an association. The association's Parent must be T; the
// compiler rejects Preload(schema.Order.Items) on a Builder[User].
// Extra conditions (if provided) are applied to the preload query as a nested scope.
//
// Example:
//
//	qb := query.New[User]().Preload(schema.User.Profile)
//	qb = query.New[User]().Preload(
//	    schema.User.Profile.Nested(schema.Profile.Address),
//	    schema.Address.City.Eq("SF"),
//	)
//	_ = qb
func (b *Builder[T]) Preload(assoc nestable[T], conds ...Condition) *Builder[T] {
	path := assoc.Path()
	return b.bind(func(db *gorm.DB) *gorm.DB {
		if len(conds) == 0 {
			return db.Preload(path)
		}
		return db.Preload(path, func(tx *gorm.DB) *gorm.DB {
			for _, cond := range conds {
				tx = cond(tx)
			}
			return tx
		})
	})
}

// Group adds GROUP BY.
func (b *Builder[T]) Group(col SQLFragment) *Builder[T] {
	expr := col.SQL()
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Group(expr)
	})
}

// Having adds HAVING. args are bound to ? placeholders in expr.
func (b *Builder[T]) Having(expr string, args ...any) *Builder[T] {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Having(expr, args...)
	})
}

// Order adds ORDER BY.
func (b *Builder[T]) Order(col SQLFragment) *Builder[T] {
	expr := col.SQL()
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Order(expr)
	})
}

// Page applies pagination (page starts from 1; pageSize defaults to 10).
func (b *Builder[T]) Page(page, pageSize int) *Builder[T] {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Limit(pageSize).Offset(offset)
	})
}

// Limit sets LIMIT.
func (b *Builder[T]) Limit(limit int) *Builder[T] {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Limit(limit)
	})
}

// Offset sets OFFSET.
func (b *Builder[T]) Offset(offset int) *Builder[T] {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset)
	})
}

// Unscoped disables default scopes such as soft deletes.
func (b *Builder[T]) Unscoped() *Builder[T] {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Unscoped()
	})
}

// Scope applies one or more GORM scopes.
func (b *Builder[T]) Scope(funcs ...func(*gorm.DB) *gorm.DB) *Builder[T] {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Scopes(funcs...)
	})
}

func (b *Builder[T]) nested(conds []Condition, applier func(db, nested *gorm.DB) *gorm.DB) *Builder[T] {
	if len(conds) == 0 {
		return b
	}
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return applier(db, gormx.BuildNested(db, conds))
	})
}

func (b *Builder[T]) bind(conds ...Condition) *Builder[T] {
	b.conditions = append(b.conditions, conds...)
	return b
}
