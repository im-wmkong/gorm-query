package fragment

import "github.com/im-wmkong/gorm-query/internal/slicex"

// Fragment is the local view of "anything that renders itself as SQL".
// query.SQLFragment automatically satisfies this interface, so query/repo
// can pass their typed fragments directly into the helpers below without
// internal/* importing the query package.
type Fragment interface {
	SQL() string
}

// RenderAll renders every fragment to its SQL string form.
func RenderAll[F Fragment](frags []F) []string {
	return slicex.Map(frags, func(f F) string { return f.SQL() })
}

// RenderAllAny renders every fragment to its SQL string form, boxed as any.
// This shape matches gorm APIs (e.g. db.Distinct, db.Select) that accept ...any.
func RenderAllAny[F Fragment](frags []F) []any {
	return slicex.Map(RenderAll(frags), func(s string) any { return s })
}

// Suffix returns "<expr><suffix>". Use it for ORDER BY decorations and
// aliasing, e.g. Suffix(col, " DESC") or Suffix(col, " AS "+alias).
func Suffix(expr, suffix string) string {
	return expr + suffix
}

// Prefix returns "<prefix><expr>". Use it for leading keywords such as
// Prefix("DISTINCT ", col).
func Prefix(prefix, expr string) string {
	return prefix + expr
}

// Call wraps expr in a SQL function call: Call("SUM", col) -> "SUM(col)".
// The function name is emitted verbatim; no quoting or validation is done.
func Call(fn, expr string) string {
	return fn + "(" + expr + ")"
}

// JoinPath joins two non-empty path segments with sep. If either side is
// empty, the other side is returned verbatim, which is exactly the policy
// query.Association.Nested needs.
func JoinPath(a, b, sep string) string {
	switch {
	case a == "":
		return b
	case b == "":
		return a
	default:
		return a + sep + b
	}
}
