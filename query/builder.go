// Package query provides the core of a type-safe GORM query builder.
//
// It builds queries dynamically via the functional Condition pattern and supports
// deep-copying (Clone) to avoid condition slice aliasing across derived queries.
// It is typically used together with the generated props (e.g. UserProps) to
// provide a smooth, type-safe, chainable SQL building experience.
package query

import (
	"github.com/im-wmkong/gorm-query/internal/column"
	"github.com/im-wmkong/gorm-query/internal/gormx"
	"gorm.io/gorm"
)

// Condition is the core abstraction: a function that transforms *gorm.DB.
// All WHERE-like conditions are eventually represented as a Condition.
type Condition func(db *gorm.DB) *gorm.DB

// Builder accumulates query conditions.
// Builder is NOT concurrency-safe; do not use the same instance in multiple goroutines.
// If you need to reuse conditions concurrently, call Clone to get an independent copy.
//
// Example:
//
//	// Build conditions with generated props, then apply to *gorm.DB.
//	qb := query.New().Where(
//	    UserProps.Status.Eq(1),
//	    UserProps.Age.Gte(18),
//	).Order(UserProps.CreatedAt.Desc()).Page(1, 20)
//
//	session := qb.Apply(db.Model(&User{}))
//	_ = session
type Builder struct {
	conditions []Condition
}

// New creates a new query builder.
//
// Example:
//
//	qb := query.New()
//	_ = qb
func New() *Builder {
	return &Builder{}
}

// Apply applies all accumulated conditions to the given gorm.DB session.
//
// Example:
//
//	session := query.New().Where(UserProps.Age.Gt(18)).Apply(db.Model(&User{}))
//	_ = session
func (b *Builder) Apply(db *gorm.DB) *gorm.DB {
	for _, cond := range b.conditions {
		db = cond(db)
	}
	return db
}

// Clone makes a deep copy of the builder so common conditions can be reused safely.
//
// Example:
//
//	base := query.New().Where(UserProps.Status.Eq(1))
//	q1 := base.Clone().Where(UserProps.Age.Gte(18))
//	q2 := base.Clone().Where(UserProps.Age.Lt(18))
//	_, _ = q1, q2
func (b *Builder) Clone() *Builder {
	conditions := make([]Condition, len(b.conditions))
	copy(conditions, b.conditions)
	return &Builder{conditions: conditions}
}

// Where appends one or more conditions.
//
// Example:
//
//	qb := query.New().Where(UserProps.Email.Like("%example.com%"))
//	_ = qb
func (b *Builder) Where(conds ...Condition) *Builder {
	return b.bind(conds...)
}

// Or appends nested OR conditions.
//
// Example:
//
//	// WHERE (status = 1) OR (status = 2)
//	qb := query.New().Or(UserProps.Status.Eq(1), UserProps.Status.Eq(2))
//	_ = qb
func (b *Builder) Or(conds ...Condition) *Builder {
	return b.nested(conds, func(db, nested *gorm.DB) *gorm.DB {
		return db.Or(nested)
	})
}

// Not appends nested NOT conditions.
//
// Example:
//
//	// WHERE NOT (status = 0)
//	qb := query.New().Not(UserProps.Status.Eq(0))
//	_ = qb
func (b *Builder) Not(conds ...Condition) *Builder {
	return b.nested(conds, func(db, nested *gorm.DB) *gorm.DB {
		return db.Not(nested)
	})
}

// Select sets the SELECT clause.
//
// Example:
//
//	// SELECT user_name, email
//	qb := query.New().Select(UserProps.UserName, UserProps.Email)
//	_ = qb
func (b *Builder) Select(query any, args ...any) *Builder {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Select(column.Value(query), column.Values(args)...)
	})
}

// Omit omits columns.
//
// Example:
//
//	// Omit columns by Column (generated props)
//	qb := query.New().Omit(UserProps.UpdatedAt, UserProps.DeletedAt)
//	_ = qb
func (b *Builder) Omit(columns ...any) *Builder {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Omit(column.ValuesTo[string](columns)...)
	})
}

// Distinct adds DISTINCT to the query.
//
// Example:
//
//	// SELECT DISTINCT email
//	qb := query.New().Distinct(UserProps.Email)
//	_ = qb
func (b *Builder) Distinct(args ...any) *Builder {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Distinct(column.Values(args)...)
	})
}

// Joins adds JOIN clauses.
//
// Example:
//
//	qb := query.New().Joins("JOIN profiles ON profiles.user_id = users.id")
//	_ = qb
func (b *Builder) Joins(query string, args ...any) *Builder {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Joins(query, column.Values(args)...)
	})
}

// Preload preloads associations.
//
// Example:
//
//	qb := query.New().Preload("Profile")
//	_ = qb
func (b *Builder) Preload(query string, args ...any) *Builder {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Preload(query, column.Values(args)...)
	})
}

// Group adds GROUP BY.
//
// Example:
//
//	qb := query.New().Group(UserProps.Status)
//	_ = qb
func (b *Builder) Group(name any) *Builder {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Group(column.ValueTo[string](name))
	})
}

// Having adds HAVING.
//
// Example:
//
//	qb := query.New().Group(UserProps.Status).Having("COUNT(*) > ?", 10)
//	_ = qb
func (b *Builder) Having(query any, args ...any) *Builder {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Having(column.Value(query), column.Values(args)...)
	})
}

// Order adds ORDER BY.
//
// Example:
//
//	qb := query.New().Order(UserProps.CreatedAt.Desc())
//	_ = qb
func (b *Builder) Order(col any) *Builder {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Order(column.Value(col))
	})
}

// Page applies pagination (page starts from 1; pageSize defaults to 10).
//
// Example:
//
//	qb := query.New().Page(2, 50) // page 2, 50 items per page
//	_ = qb
func (b *Builder) Page(page, pageSize int) *Builder {
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
//	qb := query.New().Limit(100)
//	_ = qb
func (b *Builder) Limit(limit int) *Builder {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Limit(limit)
	})
}

// Offset sets OFFSET.
//
// Example:
//
//	qb := query.New().Offset(200)
//	_ = qb
func (b *Builder) Offset(offset int) *Builder {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset)
	})
}

// Unscoped disables default scopes such as soft deletes.
//
// Example:
//
//	qb := query.New().Unscoped()
//	_ = qb
func (b *Builder) Unscoped() *Builder {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Unscoped()
	})
}

// Scope applies one or more GORM scopes.
//
// Example:
//
//	qb := query.New().Scope(func(db *gorm.DB) *gorm.DB { return db.Where("status = ?", 1) })
//	_ = qb
func (b *Builder) Scope(funcs ...func(*gorm.DB) *gorm.DB) *Builder {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Scopes(funcs...)
	})
}

func (b *Builder) nested(conds []Condition, applier func(db, nested *gorm.DB) *gorm.DB) *Builder {
	if len(conds) == 0 {
		return b
	}

	return b.bind(func(db *gorm.DB) *gorm.DB {
		return applier(db, gormx.BuildNested(db, conds))
	})
}

func (b *Builder) bind(conds ...Condition) *Builder {
	b.conditions = append(b.conditions, conds...)
	return b
}
