package query

import "github.com/im-wmkong/gorm-query/internal/slicex"

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
func (r RawFragment) SQL() string {
	return string(r)
}

// AggFragment is an aggregate expression (SUM/COUNT/...). It is also a SQLFragment
// and can be further aliased with As.
type AggFragment string

// SQL returns the aggregate expression.
func (a AggFragment) SQL() string {
	return string(a)
}

// As gives the aggregate an alias, producing e.g. "SUM(age) AS age_sum".
func (a AggFragment) As(alias string) SQLFragment {
	return RawFragment(string(a) + " AS " + alias)
}

// SQLFragments is a slice of SQLFragment providing batch rendering helpers.
type SQLFragments []SQLFragment

// Strings renders every fragment to its SQL string form.
func (fs SQLFragments) Strings() []string {
	return slicex.Map(fs, func(f SQLFragment) string { return f.SQL() })
}

// Anys renders every fragment to its SQL string form, boxed as any.
// This shape matches gorm APIs (e.g. db.Distinct) that accept ...any.
func (fs SQLFragments) Anys() []any {
	return slicex.Map(fs, func(f SQLFragment) any { return f.SQL() })
}
