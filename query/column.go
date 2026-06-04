package query

import (
	"time"

	"gorm.io/gorm"
)

// Numeric constrains numeric column value types.
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// Ordered constrains types that support ordering comparisons (numeric,
// string-like and time.Time).
type Ordered interface {
	Numeric | ~string | time.Time
}

// baseColumn carries table + column name shared by all typed columns.
// The table qualifier is optional; when empty the column renders as bare name.
type baseColumn struct {
	table string
	name  string
}

// SQL returns the qualified column name, e.g. "users.id" or "id".
//
// Example:
//
//	s := schema.User.Email.SQL() // "users.email"
//	_ = s
func (c baseColumn) SQL() string {
	if c.table == "" {
		return c.name
	}
	return c.table + "." + c.name
}

// Name returns the raw column name (without table qualifier).
//
// Example:
//
//	n := schema.User.Email.Name() // "email"
//	_ = n
func (c baseColumn) Name() string {
	return c.name
}

// Table returns the table qualifier (may be empty).
//
// Example:
//
//	t := schema.User.Email.Table() // "users"
//	_ = t
func (c baseColumn) Table() string {
	return c.table
}

// WithTable returns a copy of the column qualified with the given table/alias.
// Use it for joins or self-references without mutating the generated schema.
//
// Example:
//
//	col := schema.User.ID.WithTable("u")
//	_ = col
func (c baseColumn) WithTable(table string) baseColumn {
	return baseColumn{table: table, name: c.name}
}

// IsNull builds an IS NULL condition.
//
// Example:
//
//	qb := schema.User.Query().Where(schema.User.DeletedAt.IsNull())
//	_ = qb
func (c baseColumn) IsNull() Condition {
	return c.clause("IS NULL")
}

// IsNotNull builds an IS NOT NULL condition.
//
// Example:
//
//	qb := schema.User.Query().Where(schema.User.DeletedAt.IsNotNull())
//	_ = qb
func (c baseColumn) IsNotNull() Condition {
	return c.clause("IS NOT NULL")
}

// Desc returns an ORDER BY fragment: "<col> DESC".
//
// Example:
//
//	qb := schema.User.Query().Order(schema.User.CreatedAt.Desc())
//	_ = qb
func (c baseColumn) Desc() SQLFragment {
	return RawFragment(c.SQL() + " DESC")
}

// Asc returns an ORDER BY fragment: "<col> ASC".
//
// Example:
//
//	qb := schema.User.Query().Order(schema.User.CreatedAt.Asc())
//	_ = qb
func (c baseColumn) Asc() SQLFragment {
	return RawFragment(c.SQL() + " ASC")
}

// As returns an aliased SELECT fragment: "<col> AS <alias>".
//
// Example:
//
//	qb := schema.User.Query().Select(schema.User.UserName.As("name"))
//	_ = qb
func (c baseColumn) As(alias string) SQLFragment {
	return RawFragment(c.SQL() + " AS " + alias)
}

// Distinct returns a DISTINCT SELECT fragment: "DISTINCT <col>".
//
// Example:
//
//	qb := schema.User.Query().Select(schema.User.Email.Distinct())
//	_ = qb
func (c baseColumn) Distinct() SQLFragment {
	return RawFragment("DISTINCT " + c.SQL())
}

// Sum returns aggregate "SUM(<col>)".
//
// Example:
//
//	qb := schema.User.Query().Select(schema.User.Age.Sum().As("total_age"))
//	_ = qb
func (c baseColumn) Sum() AggFragment {
	return c.agg("SUM")
}

// Count returns aggregate "COUNT(<col>)".
//
// Example:
//
//	qb := schema.User.Query().Select(schema.User.ID.Count().As("user_count"))
//	_ = qb
func (c baseColumn) Count() AggFragment {
	return c.agg("COUNT")
}

// Avg returns aggregate "AVG(<col>)".
//
// Example:
//
//	qb := schema.User.Query().Select(schema.User.Age.Avg().As("avg_age"))
//	_ = qb
func (c baseColumn) Avg() AggFragment {
	return c.agg("AVG")
}

// Max returns aggregate "MAX(<col>)".
//
// Example:
//
//	qb := schema.User.Query().Select(schema.User.Age.Max().As("max_age"))
//	_ = qb
func (c baseColumn) Max() AggFragment {
	return c.agg("MAX")
}

// Min returns aggregate "MIN(<col>)".
//
// Example:
//
//	qb := schema.User.Query().Select(schema.User.Age.Min().As("min_age"))
//	_ = qb
func (c baseColumn) Min() AggFragment {
	return c.agg("MIN")
}

// ValueColumn is a typed column without extra operators beyond equality / in / null.
// It is the base for StringColumn / NumericColumn / TimeColumn / BoolColumn via
// embedding and is also used as the fallback for unknown custom types.
type ValueColumn[T any] struct {
	baseColumn
}

// NewValueColumn constructs a ValueColumn[T].
//
// Example:
//
//	col := query.NewValueColumn[gorm.DeletedAt]("users", "deleted_at")
//	_ = col
func NewValueColumn[T any](table, name string) ValueColumn[T] {
	return ValueColumn[T]{baseColumn: baseColumn{table: table, name: name}}
}

// Eq builds "<col> = ?".
//
// Example:
//
//	qb := schema.User.Query().Where(schema.User.Status.Eq(1))
//	_ = qb
func (c ValueColumn[T]) Eq(v T) Condition {
	return c.compare("=", v)
}

// Neq builds "<col> <> ?".
//
// Example:
//
//	qb := schema.User.Query().Where(schema.User.Status.Neq(0))
//	_ = qb
func (c ValueColumn[T]) Neq(v T) Condition {
	return c.compare("<>", v)
}

// In builds "<col> IN ?".
//
// Example:
//
//	qb := schema.User.Query().Where(schema.User.Status.In([]int{1, 2, 3}))
//	_ = qb
func (c ValueColumn[T]) In(vs []T) Condition {
	return c.compare("IN", vs)
}

// NotIn builds "<col> NOT IN ?".
//
// Example:
//
//	qb := schema.User.Query().Where(schema.User.Status.NotIn([]int{0, 99}))
//	_ = qb
func (c ValueColumn[T]) NotIn(vs []T) Condition {
	return c.compare("NOT IN", vs)
}

// Set produces an assignment "<col> = v" for Repository.Updates.
// Uses the bare column name (without table qualifier) since UPDATE targets
// a single table and GORM does not accept qualified names here.
//
// Example:
//
//	rows, err := r.Update(ctx,
//	    schema.User.Query().Where(schema.User.ID.Eq(1)),
//	    schema.User.Status.Set(2),
//	)
//	_, _ = rows, err
func (c ValueColumn[T]) Set(v T) Assignment {
	return Assignment{Column: c.Name(), Value: v}
}

// orderable[T] adds the ordered comparison operators (Gt / Gte / Lt / Lte /
// Between / NotBetween) on top of ValueColumn[T]. T is constrained by Ordered
// so only numeric, string-like and time.Time columns can embed it.
type orderable[T Ordered] struct {
	ValueColumn[T]
}

// Gt builds "<col> > ?".
//
// Example:
//
//	qb := schema.User.Query().Where(schema.User.Age.Gt(18))
//	_ = qb
func (c orderable[T]) Gt(v T) Condition {
	return c.compare(">", v)
}

// Gte builds "<col> >= ?".
//
// Example:
//
//	qb := schema.User.Query().Where(schema.User.Age.Gte(18))
//	_ = qb
func (c orderable[T]) Gte(v T) Condition {
	return c.compare(">=", v)
}

// Lt builds "<col> < ?".
//
// Example:
//
//	qb := schema.User.Query().Where(schema.User.Age.Lt(60))
//	_ = qb
func (c orderable[T]) Lt(v T) Condition {
	return c.compare("<", v)
}

// Lte builds "<col> <= ?".
//
// Example:
//
//	qb := schema.User.Query().Where(schema.User.Age.Lte(60))
//	_ = qb
func (c orderable[T]) Lte(v T) Condition {
	return c.compare("<=", v)
}

// Between builds "<col> BETWEEN ? AND ?".
//
// Example:
//
//	qb := schema.User.Query().Where(schema.User.Age.Between(18, 60))
//	_ = qb
func (c orderable[T]) Between(lo, hi T) Condition {
	return c.between("BETWEEN", lo, hi)
}

// NotBetween builds "<col> NOT BETWEEN ? AND ?".
//
// Example:
//
//	qb := schema.User.Query().Where(schema.User.Age.NotBetween(18, 60))
//	_ = qb
func (c orderable[T]) NotBetween(lo, hi T) Condition {
	return c.between("NOT BETWEEN", lo, hi)
}

// StringColumn is a typed column for string-like values. It exposes the
// string-only operators (Like / HasPrefix / ...) on top of orderable's API.
type StringColumn[T ~string] struct {
	orderable[T]
}

// NewStringColumn constructs a StringColumn[T].
//
// Example:
//
//	col := query.NewStringColumn[string]("users", "email")
//	_ = col
func NewStringColumn[T ~string](table, name string) StringColumn[T] {
	return StringColumn[T]{orderable: orderable[T]{ValueColumn: NewValueColumn[T](table, name)}}
}

// Like builds "<col> LIKE ?".
//
// Example:
//
//	qb := schema.User.Query().Where(schema.User.UserName.Like("A%"))
//	_ = qb
func (c StringColumn[T]) Like(pattern string) Condition {
	return c.compare("LIKE", pattern)
}

// NotLike builds "<col> NOT LIKE ?".
//
// Example:
//
//	qb := schema.User.Query().Where(schema.User.UserName.NotLike("A%"))
//	_ = qb
func (c StringColumn[T]) NotLike(pattern string) Condition {
	return c.compare("NOT LIKE", pattern)
}

// Contains builds "<col> LIKE %v%".
//
// Example:
//
//	qb := schema.User.Query().Where(schema.User.UserName.Contains("ali"))
//	_ = qb
func (c StringColumn[T]) Contains(v string) Condition {
	return c.compare("LIKE", "%"+v+"%")
}

// NotContains builds "<col> NOT LIKE %v%".
//
// Example:
//
//	qb := schema.User.Query().Where(schema.User.UserName.NotContains("admin"))
//	_ = qb
func (c StringColumn[T]) NotContains(v string) Condition {
	return c.compare("NOT LIKE", "%"+v+"%")
}

// HasPrefix builds "<col> LIKE v%".
//
// Example:
//
//	qb := schema.User.Query().Where(schema.User.Email.HasPrefix("admin@"))
//	_ = qb
func (c StringColumn[T]) HasPrefix(v string) Condition {
	return c.compare("LIKE", v+"%")
}

// HasSuffix builds "<col> LIKE %v".
//
// Example:
//
//	qb := schema.User.Query().Where(schema.User.Email.HasSuffix("@example.com"))
//	_ = qb
func (c StringColumn[T]) HasSuffix(v string) Condition {
	return c.compare("LIKE", "%"+v)
}

// NumericColumn is a typed column for numeric values (integers / floats).
// Ordered comparisons + Between / NotBetween come from the embedded orderable.
type NumericColumn[T Numeric] struct {
	orderable[T]
}

// NewNumericColumn constructs a NumericColumn[T].
//
// Example:
//
//	col := query.NewNumericColumn[int]("users", "age")
//	_ = col
func NewNumericColumn[T Numeric](table, name string) NumericColumn[T] {
	return NumericColumn[T]{orderable: orderable[T]{ValueColumn: NewValueColumn[T](table, name)}}
}

// TimeColumn is a typed column for time.Time values.
// Ordered comparisons + Between / NotBetween come from the embedded orderable.
type TimeColumn struct {
	orderable[time.Time]
}

// NewTimeColumn constructs a TimeColumn.
//
// Example:
//
//	col := query.NewTimeColumn("users", "created_at")
//	_ = col
func NewTimeColumn(table, name string) TimeColumn {
	return TimeColumn{orderable: orderable[time.Time]{ValueColumn: NewValueColumn[time.Time](table, name)}}
}

// BoolColumn is a typed column for boolean values.
type BoolColumn struct {
	ValueColumn[bool]
}

// NewBoolColumn constructs a BoolColumn.
//
// Example:
//
//	col := query.NewBoolColumn("users", "active")
//	_ = col
func NewBoolColumn(table, name string) BoolColumn {
	return BoolColumn{ValueColumn: NewValueColumn[bool](table, name)}
}

// IsTrue is a semantic alias for Eq(true).
//
// Example:
//
//	qb := schema.User.Query().Where(schema.User.Active.IsTrue())
//	_ = qb
func (c BoolColumn) IsTrue() Condition {
	return c.compare("=", true)
}

// IsFalse is a semantic alias for Eq(false).
//
// Example:
//
//	qb := schema.User.Query().Where(schema.User.Active.IsFalse())
//	_ = qb
func (c BoolColumn) IsFalse() Condition {
	return c.compare("=", false)
}

// ---- internal helpers ----

// agg is the single source of truth for "<FN>(<col>)" aggregate fragments.
func (c baseColumn) agg(fn string) AggFragment {
	return AggFragment(fn + "(" + c.SQL() + ")")
}

func (c baseColumn) compare(op string, val any) Condition {
	return c.clause(op+" ?", val)
}

func (c baseColumn) between(op string, lo, hi any) Condition {
	return c.clause(op+" ? AND ?", lo, hi)
}

func (c baseColumn) clause(suffix string, args ...any) Condition {
	qualified := c.SQL()
	return func(db *gorm.DB) *gorm.DB {
		if len(args) == 0 {
			return db.Where(qualified + " " + suffix)
		}
		return db.Where(qualified+" "+suffix, args...)
	}
}
