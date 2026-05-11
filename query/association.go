package query

import "github.com/im-wmkong/gorm-query/internal/fragment"

// Association represents a GORM association between a Parent entity P and a
// Child entity C (e.g. User has-one Profile yields Association[User, Profile]).
// The path is a GORM association field name like "Profile", possibly joined
// with dots for nesting: "Profile.Address".
//
// Example:
//
//	// Generated schema field:
//	//   Profile query.Association[model.User, model.Profile]
//	qb := query.New[model.User]().Preload(schema.User.Profile)
//	_ = qb
type Association[P, C any] struct {
	path string
}

// NewAssociation constructs an Association with the given field name.
// It is intended for use by schemagen's generated code.
func NewAssociation[P, C any](name string) Association[P, C] {
	return Association[P, C]{path: name}
}

// Path returns the dotted association path (e.g. "Profile" or "Profile.Address").
func (a Association[P, C]) Path() string { return a.path }

// parentOf is a tiny marker method that encodes the Parent type in the method
// set. It makes Association satisfy nestable[P] below, which is how we perform
// compile-time Parent matching without introducing method-level type params.
func (a Association[P, C]) parentOf(P) {}

// nestable[P] is the interface "an association whose Parent is P".
// Any Association[P, *] satisfies nestable[P] (via parentOf(P)). It is used
// by Association.Nested and Builder.Preload to check Parent compatibility.
//
// Child is intentionally erased: methods cannot introduce new type parameters,
// so Association.Nested returns Association[P, any]. Parent checking at every
// level still works because each nestable[P] explicitly binds a concrete P.
type nestable[P any] interface {
	parentOf(P)
	Path() string
}

// Nested chains a child association onto the current one. The sub must have
// Parent == C (this association's Child), otherwise the call fails to compile.
// The returned association has Child erased to any; Preload only needs the
// path + Parent, so this erasure is invisible in practice.
//
// Example:
//
//	// Preload "Profile.Address"
//	rel := schema.User.Profile.Nested(schema.Profile.Address)
//	qb := query.New[model.User]().Preload(rel)
//	_ = qb
func (a Association[P, C]) Nested(sub nestable[C]) Association[P, any] {
	return Association[P, any]{path: fragment.JoinPath(a.path, sub.Path(), ".")}
}
