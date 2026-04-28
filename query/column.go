package query

import (
	"fmt"

	"gorm.io/gorm"
)

// Column represents a database column name.
//
// Example:
//
//	// WHERE age >= 18 AND email LIKE "%@example.com%"
//	qb := query.New().Where(
//	    columns.User.Age.Gte(18),
//	    columns.User.Email.Like("%@example.com%"),
//	)
//
//	// ORDER BY created_at DESC
//	qb = qb.Order(columns.User.CreatedAt.Desc())
//	_ = qb
type Column string

var _ fmt.Stringer = (*Column)(nil)

// String returns the column as string.
//
// Example:
//
//	s := columns.User.Email.String()
//	_ = s
func (c Column) String() string {
	return string(c)
}

// Eq builds an equality condition: column = value.
//
// Example:
//
//	qb := query.New().Where(columns.User.ID.Eq(1))
//	_ = qb
func (c Column) Eq(val any) Condition {
	return c.compare("=", val)
}

// Neq builds an inequality condition: column <> value.
//
// Example:
//
//	qb := query.New().Where(columns.User.Status.Neq(0))
//	_ = qb
func (c Column) Neq(val any) Condition {
	return c.compare("<>", val)
}

// Gt builds a greater-than condition: column > value.
//
// Example:
//
//	qb := query.New().Where(columns.User.Age.Gt(18))
//	_ = qb
func (c Column) Gt(val any) Condition {
	return c.compare(">", val)
}

// Gte builds a greater-than-or-equal condition: column >= value.
//
// Example:
//
//	qb := query.New().Where(columns.User.Age.Gte(18))
//	_ = qb
func (c Column) Gte(val any) Condition {
	return c.compare(">=", val)
}

// Lt builds a less-than condition: column < value.
//
// Example:
//
//	qb := query.New().Where(columns.User.Age.Lt(65))
//	_ = qb
func (c Column) Lt(val any) Condition {
	return c.compare("<", val)
}

// Lte builds a less-than-or-equal condition: column <= value.
//
// Example:
//
//	qb := query.New().Where(columns.User.Age.Lte(65))
//	_ = qb
func (c Column) Lte(val any) Condition {
	return c.compare("<=", val)
}

// Like builds a LIKE condition: column LIKE value.
//
// Example:
//
//	qb := query.New().Where(columns.User.Email.Like("%@example.com%"))
//	_ = qb
func (c Column) Like(val any) Condition {
	return c.compare("LIKE", val)
}

// NotLike builds a NOT LIKE condition: column NOT LIKE value.
//
// Example:
//
//	qb := query.New().Where(columns.User.Email.NotLike("%@example.com%"))
//	_ = qb
func (c Column) NotLike(val any) Condition {
	return c.compare("NOT LIKE", val)
}

// Contains builds a LIKE condition: column LIKE %value%.
//
// Example:
//
//	qb := query.New().Where(columns.User.UserName.Contains("li"))
//	_ = qb
func (c Column) Contains(val string) Condition {
	return c.compare("LIKE", "%"+val+"%")
}

// NotContains builds a NOT LIKE condition: column NOT LIKE %value%.
//
// Example:
//
//	qb := query.New().Where(columns.User.UserName.NotContains("admin"))
//	_ = qb
func (c Column) NotContains(val string) Condition {
	return c.compare("NOT LIKE", "%"+val+"%")
}

// HasPrefix builds a LIKE condition: column LIKE value%.
//
// Example:
//
//	qb := query.New().Where(columns.User.UserName.HasPrefix("Al"))
//	_ = qb
func (c Column) HasPrefix(val string) Condition {
	return c.compare("LIKE", val+"%")
}

// HasSuffix builds a LIKE condition: column LIKE %value.
//
// Example:
//
//	qb := query.New().Where(columns.User.UserName.HasSuffix("son"))
//	_ = qb
func (c Column) HasSuffix(val string) Condition {
	return c.compare("LIKE", "%"+val)
}

// In builds an IN condition: column IN (values).
//
// Example:
//
//	qb := query.New().Where(columns.User.UserName.In([]string{"Alice", "Bob"}))
//	_ = qb
func (c Column) In(vals any) Condition {
	return c.compare("IN", vals)
}

// NotIn builds a NOT IN condition: column NOT IN (values).
//
// Example:
//
//	qb := query.New().Where(columns.User.UserName.NotIn([]string{"admin", "root"}))
//	_ = qb
func (c Column) NotIn(vals any) Condition {
	return c.compare("NOT IN", vals)
}

// Between builds a BETWEEN condition: column BETWEEN start AND end.
//
// Example:
//
//	qb := query.New().Where(columns.User.Age.Between(18, 30))
//	_ = qb
func (c Column) Between(start, end any) Condition {
	return c.between("BETWEEN", start, end)
}

// NotBetween builds a NOT BETWEEN condition: column NOT BETWEEN start AND end.
//
// Example:
//
//	qb := query.New().Where(columns.User.Age.NotBetween(18, 30))
//	_ = qb
func (c Column) NotBetween(start, end any) Condition {
	return c.between("NOT BETWEEN", start, end)
}

// IsNull builds an IS NULL condition.
//
// Example:
//
//	qb := query.New().Where(columns.User.DeletedAt.IsNull())
//	_ = qb
func (c Column) IsNull() Condition {
	return c.clause("IS NULL")
}

// IsNotNull builds an IS NOT NULL condition.
//
// Example:
//
//	qb := query.New().Where(columns.User.DeletedAt.IsNotNull())
//	_ = qb
func (c Column) IsNotNull() Condition {
	return c.clause("IS NOT NULL")
}

// Desc returns an ORDER BY fragment: "column DESC".
//
// Example:
//
//	qb := query.New().Order(columns.User.CreatedAt.Desc())
//	_ = qb
func (c Column) Desc() string {
	return c.String() + " DESC"
}

// Asc returns an ORDER BY fragment: "column ASC".
//
// Example:
//
//	qb := query.New().Order(columns.User.CreatedAt.Asc())
//	_ = qb
func (c Column) Asc() string {
	return c.String() + " ASC"
}

// Table qualifies the column with table name: "table.column".
//
// Example:
//
//	col := columns.User.Email.Table("users")
//	_ = col
func (c Column) Table(name string) Column {
	return Column(name + "." + c.String())
}

// As returns an aliased SELECT fragment: "column AS alias".
//
// Example:
//
//	qb := query.New().Select(columns.User.Email.As("user_name"))
//	_ = qb
func (c Column) As(alias string) Column {
	return Column(c.String() + " AS " + alias)
}

// Distinct returns a DISTINCT SELECT fragment: "DISTINCT column".
//
// Example:
//
//	qb := query.New().Select(columns.User.Email.Distinct())
//	_ = qb
func (c Column) Distinct() Column {
	return Column("DISTINCT " + c.String())
}

// Sum returns an aggregate SELECT fragment: "SUM(column)".
//
// Example:
//
//	qb := query.New().Select(columns.User.Age.Sum().As("age_sum"))
//	_ = qb
func (c Column) Sum() Column {
	return Column("SUM(" + c.String() + ")")
}

// Count returns an aggregate SELECT fragment: "COUNT(column)".
//
// Example:
//
//	qb := query.New().Select(columns.User.ID.Count().As("cnt"))
//	_ = qb
func (c Column) Count() Column {
	return Column("COUNT(" + c.String() + ")")
}

// Avg returns an aggregate SELECT fragment: "AVG(column)".
//
// Example:
//
//	qb := query.New().Select(columns.User.Age.Avg().As("age_avg"))
//	_ = qb
func (c Column) Avg() Column {
	return Column("AVG(" + c.String() + ")")
}

// Max returns an aggregate SELECT fragment: "MAX(column)".
//
// Example:
//
//	qb := query.New().Select(columns.User.Age.Max().As("age_max"))
//	_ = qb
func (c Column) Max() Column {
	return Column("MAX(" + c.String() + ")")
}

// Min returns an aggregate SELECT fragment: "MIN(column)".
//
// Example:
//
//	qb := query.New().Select(columns.User.Age.Min().As("age_min"))
//	_ = qb
func (c Column) Min() Column {
	return Column("MIN(" + c.String() + ")")
}

func (c Column) compare(op string, val any) Condition {
	return c.clause(op+" ?", val)
}

func (c Column) between(op string, start, end any) Condition {
	return c.clause(op+" ? AND ?", start, end)
}

func (c Column) clause(suffix string, args ...any) Condition {
	return func(db *gorm.DB) *gorm.DB {
		if len(args) == 0 {
			return db.Where(c.String() + " " + suffix)
		}

		resolved := make([]any, len(args))
		for i, arg := range args {
			if col, ok := arg.(Column); ok {
				resolved[i] = gorm.Expr(col.String())
			} else {
				resolved[i] = arg
			}
		}
		return db.Where(c.String()+" "+suffix, resolved...)
	}
}
