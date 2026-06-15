package query

import (
	"errors"

	"github.com/im-wmkong/gorm-query/internal/slicex"
)

// ErrNoAssignment is returned by update operations when no assignment is
// provided. An empty assignment list is treated as a programming error rather
// than a silent no-op, so callers can catch it with errors.Is.
var ErrNoAssignment = errors.New("query: no assignment provided")

// Assignment represents a single "<column> = value" pair for update statements.
// It is produced by a typed column's Set method, so the value type is checked
// against the column's declared type at compile time.
//
// Example:
//
//	r.Updates(ctx, qb, schema.User.Age.Set(31), schema.User.Status.Set(2))
type Assignment struct {
	Column string
	Value  any
}

// Assignments is a shorthand builder for multiple Assignment values.
// It can be passed directly to Repository.Updates via the variadic form.
type Assignments []Assignment

// ToMap converts assignments into a map[string]any form expected by GORM.
// Later assignments to the same column override earlier ones.
func (as Assignments) ToMap() map[string]any {
	return slicex.ToMap(as, func(a Assignment) (string, any) {
		return a.Column, a.Value
	})
}
