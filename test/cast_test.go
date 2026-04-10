package test

import (
	"testing"

	"github.com/im-wmkong/gorm-query/internal/cast"
	"github.com/im-wmkong/gorm-query/query"
	"github.com/stretchr/testify/assert"
)

type stringerValue string

func (s stringerValue) String() string {
	return string(s)
}

func TestInternalCast(t *testing.T) {
	t.Run("Value and Values unwrap stringer", func(t *testing.T) {
		args := []any{stringerValue("user_name"), 42}

		assert.Equal(t, "user_name", cast.Value(args[0]))
		assert.Equal(t, 42, cast.Value(args[1]))
		assert.Equal(t, []any{"user_name", 42}, cast.Values(args))
		assert.Nil(t, cast.Values(nil))
	})

	t.Run("As ValueAs and ValuesAs keep zero values on mismatch", func(t *testing.T) {
		assert.Equal(t, "status", cast.ValueTo[string](stringerValue("status")))
		assert.Empty(t, cast.ValueTo[string](123))
		assert.Equal(t, []string{"id", "name", ""}, cast.ValuesTo[string]([]any{stringerValue("id"), "name", 10}))
		assert.Nil(t, cast.ValuesTo[string](nil))
	})

	t.Run("MapKeys converts named map keys via stringer", func(t *testing.T) {
		values := map[query.Column]int{
			query.Column("age"):    18,
			query.Column("status"): 1,
		}

		assert.Equal(t, map[string]int{"age": 18, "status": 1}, cast.ToStringMap(values))
		assert.Empty(t, cast.ToStringMap(map[query.Column]int(nil)))
	})
}
