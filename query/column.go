package query

import (
	"time"

	"github.com/im-wmkong/gorm-query/internal/fragment"
	"gorm.io/gorm"
)

// SQLFragment is any value that can render itself as a SQL fragment.
// Columns, ordered/aliased/aggregated expressions and raw fragments all
// implement it, which is how Builder accepts heterogeneous inputs for
// Select / Order / Group without losing type safety on WHERE operands.
type SQLFragment interface {
	SQL() string
}

// RawFragment is a ready-to-use SQL fragment. It bypasses type checking;
// use it only for expressions that the typed columns cannot express.
type RawFragment string

// SQL returns the raw SQL fragment.
func (r RawFragment) SQL() string { return string(r) }

// Numeric constrains numeric column value types.
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// Ordered constrains types that support ordering comparisons.
type Ordered interface {
	Numeric | ~string
}

// baseColumn carries table + column name shared by all typed columns.
// The table qualifier is optional; when empty the column renders as bare name.
type baseColumn struct {
	table string
	name  string
}

// SQL returns the qualified column name, e.g. "users.id" or "id".
func (c baseColumn) SQL() string {
	if c.table == "" {
		return c.name
	}
	return c.table + "." + c.name
}

// Name returns the raw column name (without table qualifier).
func (c baseColumn) Name() string { return c.name }

// Table returns the table qualifier (may be empty).
func (c baseColumn) Table() string { return c.table }

// WithTable returns a copy of the column qualified with the given table/alias.
// Use it for joins or self-references without mutating the generated schema.
func (c baseColumn) WithTable(table string) baseColumn {
	return baseColumn{table: table, name: c.name}
}

// IsNull builds an IS NULL condition.
func (c baseColumn) IsNull() Condition { return c.clause("IS NULL") }

// IsNotNull builds an IS NOT NULL condition.
func (c baseColumn) IsNotNull() Condition { return c.clause("IS NOT NULL") }

// Desc returns an ORDER BY fragment: "<col> DESC".
func (c baseColumn) Desc() SQLFragment { return RawFragment(fragment.Suffix(c.SQL(), " DESC")) }

// Asc returns an ORDER BY fragment: "<col> ASC".
func (c baseColumn) Asc() SQLFragment { return RawFragment(fragment.Suffix(c.SQL(), " ASC")) }

// As returns an aliased SELECT fragment: "<col> AS <alias>".
func (c baseColumn) As(alias string) SQLFragment {
	return RawFragment(fragment.Suffix(c.SQL(), " AS "+alias))
}

// Distinct returns a DISTINCT SELECT fragment: "DISTINCT <col>".
func (c baseColumn) Distinct() SQLFragment { return RawFragment(fragment.Prefix("DISTINCT ", c.SQL())) }

// Sum returns aggregate "SUM(<col>)".
func (c baseColumn) Sum() AggFragment { return c.agg("SUM") }

// Count returns aggregate "COUNT(<col>)".
func (c baseColumn) Count() AggFragment { return c.agg("COUNT") }

// Avg returns aggregate "AVG(<col>)".
func (c baseColumn) Avg() AggFragment { return c.agg("AVG") }

// Max returns aggregate "MAX(<col>)".
func (c baseColumn) Max() AggFragment { return c.agg("MAX") }

// Min returns aggregate "MIN(<col>)".
func (c baseColumn) Min() AggFragment { return c.agg("MIN") }

// AggFragment is an aggregate expression (SUM/COUNT/...). It is also a SQLFragment
// and can be further aliased with As.
type AggFragment string

// SQL returns the aggregate expression.
func (a AggFragment) SQL() string { return string(a) }

// As gives the aggregate an alias, producing e.g. "SUM(age) AS age_sum".
func (a AggFragment) As(alias string) SQLFragment {
	return RawFragment(fragment.Suffix(string(a), " AS "+alias))
}

// ValueColumn is a typed column without extra operators beyond equality / in / null.
// It is the base for StringColumn / NumericColumn / TimeColumn / BoolColumn via
// embedding and is also used as the fallback for unknown custom types.
type ValueColumn[T any] struct {
	baseColumn
}

// NewValueColumn constructs a ValueColumn[T].
func NewValueColumn[T any](table, name string) ValueColumn[T] {
	return ValueColumn[T]{baseColumn: baseColumn{table: table, name: name}}
}

// Eq builds "<col> = ?".
func (c ValueColumn[T]) Eq(v T) Condition { return c.compare("=", v) }

// Neq builds "<col> <> ?".
func (c ValueColumn[T]) Neq(v T) Condition { return c.compare("<>", v) }

// In builds "<col> IN ?".
func (c ValueColumn[T]) In(vs []T) Condition { return c.clause("IN ?", vs) }

// NotIn builds "<col> NOT IN ?".
func (c ValueColumn[T]) NotIn(vs []T) Condition { return c.clause("NOT IN ?", vs) }

// Set produces an assignment "<col> = v" for Repository.Updates.
// Uses the bare column name (without table qualifier) since UPDATE targets
// a single table and GORM does not accept qualified names here.
func (c ValueColumn[T]) Set(v T) Assignment {
	return Assignment{Column: c.Name(), Value: v}
}

// orderable[T] adds the ordered comparison operators (Gt / Gte / Lt / Lte /
// Between / NotBetween) on top of ValueColumn[T]. It is embedded by every
// typed column whose underlying SQL type supports ordering (numeric, string,
// time). T is intentionally unconstrained here because SQL itself happily
// orders strings, numbers and times alike; the *outer* column types decide
// which T's are allowed (via Numeric / ~string / time.Time).
type orderable[T any] struct {
	ValueColumn[T]
}

func (c orderable[T]) Gt(v T) Condition  { return c.compare(">", v) }
func (c orderable[T]) Gte(v T) Condition { return c.compare(">=", v) }
func (c orderable[T]) Lt(v T) Condition  { return c.compare("<", v) }
func (c orderable[T]) Lte(v T) Condition { return c.compare("<=", v) }
func (c orderable[T]) Between(lo, hi T) Condition {
	return c.clause("BETWEEN ? AND ?", lo, hi)
}
func (c orderable[T]) NotBetween(lo, hi T) Condition {
	return c.clause("NOT BETWEEN ? AND ?", lo, hi)
}

// StringColumn is a typed column for string-like values. It exposes the
// string-only operators (Like / HasPrefix / ...) on top of orderable's API.
type StringColumn[T ~string] struct {
	orderable[T]
}

// NewStringColumn constructs a StringColumn[T].
func NewStringColumn[T ~string](table, name string) StringColumn[T] {
	return StringColumn[T]{orderable: orderable[T]{ValueColumn: NewValueColumn[T](table, name)}}
}

// Like builds "<col> LIKE ?".
func (c StringColumn[T]) Like(pattern string) Condition { return c.compare("LIKE", pattern) }

// NotLike builds "<col> NOT LIKE ?".
func (c StringColumn[T]) NotLike(pattern string) Condition { return c.compare("NOT LIKE", pattern) }

// Contains builds "<col> LIKE %v%".
func (c StringColumn[T]) Contains(v string) Condition { return c.compare("LIKE", "%"+v+"%") }

// NotContains builds "<col> NOT LIKE %v%".
func (c StringColumn[T]) NotContains(v string) Condition {
	return c.compare("NOT LIKE", "%"+v+"%")
}

// HasPrefix builds "<col> LIKE v%".
func (c StringColumn[T]) HasPrefix(v string) Condition { return c.compare("LIKE", v+"%") }

// HasSuffix builds "<col> LIKE %v".
func (c StringColumn[T]) HasSuffix(v string) Condition { return c.compare("LIKE", "%"+v) }

// NumericColumn is a typed column for numeric values (integers / floats).
// Ordered comparisons + Between / NotBetween come from the embedded orderable.
type NumericColumn[T Numeric] struct {
	orderable[T]
}

// NewNumericColumn constructs a NumericColumn[T].
func NewNumericColumn[T Numeric](table, name string) NumericColumn[T] {
	return NumericColumn[T]{orderable: orderable[T]{ValueColumn: NewValueColumn[T](table, name)}}
}

// TimeColumn is a typed column for time.Time values.
// Ordered comparisons + Between / NotBetween come from the embedded orderable.
type TimeColumn struct {
	orderable[time.Time]
}

// NewTimeColumn constructs a TimeColumn.
func NewTimeColumn(table, name string) TimeColumn {
	return TimeColumn{orderable: orderable[time.Time]{ValueColumn: NewValueColumn[time.Time](table, name)}}
}

// BoolColumn is a typed column for boolean values.
type BoolColumn struct {
	ValueColumn[bool]
}

// NewBoolColumn constructs a BoolColumn.
func NewBoolColumn(table, name string) BoolColumn {
	return BoolColumn{ValueColumn: NewValueColumn[bool](table, name)}
}

// IsTrue is a semantic alias for Eq(true).
func (c BoolColumn) IsTrue() Condition { return c.compare("=", true) }

// IsFalse is a semantic alias for Eq(false).
func (c BoolColumn) IsFalse() Condition { return c.compare("=", false) }

// ---- internal helpers ----

// agg is the single source of truth for "<FN>(<col>)" aggregate fragments.
func (c baseColumn) agg(fn string) AggFragment {
	return AggFragment(fragment.Call(fn, c.SQL()))
}

func (c baseColumn) compare(op string, val any) Condition {
	return c.clause(op+" ?", val)
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
