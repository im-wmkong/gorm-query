package test

import (
	"testing"

	"github.com/im-wmkong/gorm-query/internal/column"
	"github.com/im-wmkong/gorm-query/query"
	"github.com/stretchr/testify/assert"
)

type stringerValue string

func (s stringerValue) String() string {
	return string(s)
}

func TestInternalColumn(t *testing.T) {
	t.Run("Value and Values unwrap stringer", func(t *testing.T) {
		args := []any{query.Column("user_name"), stringerValue("status"), 42}

		assert.Equal(t, "user_name", column.Value(args[0]))
		assert.Equal(t, "status", column.Value(args[1]))
		assert.Equal(t, 42, column.Value(args[2]))
		assert.Equal(t, []any{"user_name", "status", 42}, column.Values(args))
		assert.Nil(t, column.Values(nil))
	})

	t.Run("ValueTo and ValuesTo keep zero values on mismatch", func(t *testing.T) {
		assert.Equal(t, "age", column.ValueTo[string](query.Column("age")))
		assert.Equal(t, "name", column.ValueTo[string](stringerValue("name")))
		assert.Empty(t, column.ValueTo[string](123))
		assert.Equal(t, []string{"id", "status", ""}, column.ValuesTo[string]([]any{query.Column("id"), stringerValue("status"), 10}))
		assert.Nil(t, column.ValuesTo[string](nil))
	})
}
