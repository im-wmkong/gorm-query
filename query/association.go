package query

import "fmt"

// Association represents a GORM association (relationship) name on a model,
// such as "Profile", "Orders", or nested forms like "Profile.Address".
//
// It is typically produced by the generated schema (e.g. schema.User.Profile)
// and consumed by Builder.Preload / Builder.Joins for type-safe usage.
//
// Example:
//
//	qb := query.New().Preload(schema.User.Profile)
//	_ = qb
type Association string

var _ fmt.Stringer = Association("")

// String returns the association as string.
func (a Association) String() string {
	return string(a)
}

// Nested returns a nested association path joined by ".".
//
// Example:
//
//	// Preload "Profile.Address"
//	qb := query.New().Preload(schema.User.Profile.Nested(schema.Profile.Address))
//	_ = qb
func (a Association) Nested(sub Association) Association {
	if a == "" {
		return sub
	}
	if sub == "" {
		return a
	}
	return a + "." + sub
}
