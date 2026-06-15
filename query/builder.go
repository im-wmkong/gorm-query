// Package query provides the core of a type-safe GORM query builder.
//
// It builds queries dynamically via the functional Condition pattern. Builders
// are immutable: every chained method returns a new Builder, so derived
// queries never share state with their base. It is typically used together
// with the generated schema (e.g. schema.User) to provide a smooth, type-safe,
// chainable SQL building experience.
package query

import (
	"github.com/im-wmkong/gorm-query/internal/gormx"
	"gorm.io/gorm"
)

// Condition is the core abstraction: a function that transforms *gorm.DB.
// All WHERE-like conditions are eventually represented as a Condition.
type Condition func(db *gorm.DB) *gorm.DB

// Builder accumulates query conditions for a specific entity type T.
//
// Builder is immutable: every chained method (Where, Or, Select, ...) returns
// a NEW Builder and leaves the receiver unchanged. Deriving multiple queries
// from the same base builder is therefore safe without an explicit copy:
//
//	base    := schema.User.Query().Where(schema.User.Status.Eq(1))
//	adults  := base.Where(schema.User.Age.Gte(18)) // does NOT mutate base
//	minors  := base.Where(schema.User.Age.Lt(18))  // does NOT mutate base
//
// Because the receiver is never mutated, a Builder is also safe to be read
// concurrently from multiple goroutines once it is fully constructed.
//
// The type parameter T ties Preload to associations whose Parent is T, so that
// schema.Order.Items cannot be preloaded through a Builder[User].
//
// Example:
//
//	qb := schema.User.Query().Where(
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
//	qb := schema.User.Query()
//	_ = qb
func New[T any]() *Builder[T] {
	return &Builder[T]{}
}

// Apply applies all accumulated conditions to the given gorm.DB session.
// It runs against an independent session so the passed-in db is left untouched,
// keeping repeated Apply calls on the same base mutually isolated.
//
// Example:
//
//	qb := schema.User.Query().Where(schema.User.Status.Eq(1))
//	var users []User
//	_ = qb.Apply(db.Model(&User{})).Find(&users).Error
func (b *Builder[T]) Apply(db *gorm.DB) *gorm.DB {
	return gormx.Apply(db.Session(&gorm.Session{}), b.conditions)
}

// Where appends one or more conditions.
//
// Example:
//
//	qb := schema.User.Query().Where(
//	    schema.User.Status.Eq(1),
//	    schema.User.Age.Gte(18),
//	)
//	_ = qb
func (b *Builder[T]) Where(conds ...Condition) *Builder[T] {
	return b.bind(conds...)
}

// Or appends nested OR conditions.
//
// Example:
//
//	qb := schema.User.Query().Where(schema.User.Status.Eq(1)).Or(
//	    schema.User.Age.Lt(18),
//	    schema.User.Age.Gt(60),
//	)
//	_ = qb
func (b *Builder[T]) Or(conds ...Condition) *Builder[T] {
	return b.nested(conds, func(db, nested *gorm.DB) *gorm.DB {
		return db.Or(nested)
	})
}

// Not appends nested NOT conditions.
//
// Example:
//
//	qb := schema.User.Query().Not(
//	    schema.User.Status.Eq(0),
//	)
//	_ = qb
func (b *Builder[T]) Not(conds ...Condition) *Builder[T] {
	return b.nested(conds, func(db, nested *gorm.DB) *gorm.DB {
		return db.Not(nested)
	})
}

// Select sets the SELECT clause. Accepts any mixture of typed columns and
// SQL fragments (Distinct, As, aggregates) because they all satisfy SQLFragment.
//
// Example:
//
//	qb := schema.User.Query().Select(
//	    schema.User.ID,
//	    schema.User.UserName,
//	)
//	_ = qb
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
//
// Example:
//
//	qb := schema.User.Query().Omit(schema.User.Password)
//	_ = qb
func (b *Builder[T]) Omit(cols ...SQLFragment) *Builder[T] {
	names := SQLFragments(cols).Strings()
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Omit(names...)
	})
}

// Distinct adds DISTINCT to the query.
//
// Example:
//
//	qb := schema.User.Query().Distinct(schema.User.Email)
//	_ = qb
func (b *Builder[T]) Distinct(cols ...SQLFragment) *Builder[T] {
	args := SQLFragments(cols).Anys()
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Distinct(args...)
	})
}

// Preload preloads an association. The association's Parent must be T; the
// compiler rejects Preload(schema.Order.Items) on a Builder[User].
// Extra conditions (if provided) are applied to the preload query as a nested scope.
//
// Example:
//
//	qb := schema.User.Query().Preload(schema.User.Profile)
//	qb = schema.User.Query().Preload(
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
			return gormx.Apply(tx, conds)
		})
	})
}

// Joins performs a LEFT JOIN on the given association. The association's
// Parent must be T; the compiler rejects Joins(schema.Order.Items) on a
// Builder[User]. Extra conditions (if provided) are applied to the join's
// ON clause as a nested scope.
//
// Example:
//
//	qb := schema.User.Query().Joins(schema.User.Profile)
//	qb = schema.User.Query().Joins(
//	    schema.User.Profile,
//	    schema.Profile.City.Eq("SF"),
//	)
//	_ = qb
func (b *Builder[T]) Joins(assoc nestable[T], conds ...Condition) *Builder[T] {
	return b.joins(assoc, conds, func(db *gorm.DB, path string, args ...any) *gorm.DB {
		return db.Joins(path, args...)
	})
}

// InnerJoins performs an INNER JOIN on the given association. It mirrors
// Joins but produces an inner join instead of a left join.
//
// Example:
//
//	qb := schema.User.Query().InnerJoins(schema.User.Profile)
//	_ = qb
func (b *Builder[T]) InnerJoins(assoc nestable[T], conds ...Condition) *Builder[T] {
	return b.joins(assoc, conds, func(db *gorm.DB, path string, args ...any) *gorm.DB {
		return db.InnerJoins(path, args...)
	})
}

// Group adds GROUP BY.
//
// Example:
//
//	qb := schema.User.Query().Group(schema.User.Status)
//	_ = qb
func (b *Builder[T]) Group(col SQLFragment) *Builder[T] {
	expr := col.SQL()
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Group(expr)
	})
}

// Having adds HAVING. args are bound to ? placeholders in expr.
//
// Example:
//
//	qb := schema.User.Query().
//	    Group(schema.User.Status).
//	    Having("COUNT(*) > ?", 10)
//	_ = qb
func (b *Builder[T]) Having(expr string, args ...any) *Builder[T] {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Having(expr, args...)
	})
}

// Order adds ORDER BY.
//
// Example:
//
//	qb := schema.User.Query().Order(schema.User.CreatedAt.Desc())
//	_ = qb
func (b *Builder[T]) Order(col SQLFragment) *Builder[T] {
	expr := col.SQL()
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Order(expr)
	})
}

// Page applies pagination (page starts from 1; pageSize defaults to 10).
//
// Example:
//
//	qb := schema.User.Query().Page(2, 20)
//	_ = qb
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
//
// Example:
//
//	qb := schema.User.Query().Limit(50)
//	_ = qb
func (b *Builder[T]) Limit(limit int) *Builder[T] {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Limit(limit)
	})
}

// Offset sets OFFSET.
//
// Example:
//
//	qb := schema.User.Query().Offset(100).Limit(20)
//	_ = qb
func (b *Builder[T]) Offset(offset int) *Builder[T] {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset)
	})
}

// Unscoped disables default scopes such as soft deletes.
//
// Example:
//
//	qb := schema.User.Query().Unscoped().Where(schema.User.ID.Eq(1))
//	_ = qb
func (b *Builder[T]) Unscoped() *Builder[T] {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Unscoped()
	})
}

// Scope applies one or more GORM scopes.
//
// Example:
//
//	activeOnly := func(db *gorm.DB) *gorm.DB {
//	    return db.Where("status = ?", 1)
//	}
//	qb := schema.User.Query().Scope(activeOnly)
//	_ = qb
func (b *Builder[T]) Scope(funcs ...func(*gorm.DB) *gorm.DB) *Builder[T] {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Scopes(funcs...)
	})
}

func (b *Builder[T]) joins(assoc nestable[T], conds []Condition, applier func(db *gorm.DB, path string, args ...any) *gorm.DB) *Builder[T] {
	path := assoc.Path()
	return b.bind(func(db *gorm.DB) *gorm.DB {
		if len(conds) == 0 {
			return applier(db, path)
		}
		return applier(db, path, gormx.BuildNested(db, conds))
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
	if len(conds) == 0 {
		return b
	}
	// Allocate a fresh slice with cap == len so that two Builders derived from
	// the same parent can never overwrite each other's conditions through a
	// shared underlying array.
	conditions := make([]Condition, len(b.conditions)+len(conds))
	copy(conditions, b.conditions)
	copy(conditions[len(b.conditions):], conds)
	return &Builder[T]{conditions: conditions}
}
