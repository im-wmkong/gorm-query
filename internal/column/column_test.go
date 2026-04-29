package column

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type stringerValue string

func (s stringerValue) String() string {
	return string(s)
}

func TestColumn(t *testing.T) {
	t.Run("Value and Values unwrap stringer", func(t *testing.T) {
		args := []any{stringerValue("status"), 42}

		assert.Equal(t, "status", Value(args[0]))
		assert.Equal(t, 42, Value(args[1]))
		assert.Equal(t, []any{"status", 42}, Values(args))
		assert.Nil(t, Values[any](nil))
	})

	t.Run("ValueTo and ValuesTo keep zero values on mismatch", func(t *testing.T) {
		assert.Equal(t, "name", ValueTo[string](stringerValue("name")))
		assert.Equal(t, "123", ValueTo[string](123))
		assert.Equal(t, []string{"status", "10"}, ValuesTo[string]([]any{stringerValue("status"), 10}))
		assert.Nil(t, ValuesTo[string, any](nil))
	})

	t.Run("ToStringMap converts map to string values", func(t *testing.T) {
		values := map[stringerValue]int{
			stringerValue("age"):    18,
			stringerValue("status"): 1,
		}

		assert.Equal(t, map[string]int{"age": 18, "status": 1}, ToStringMap(values))
		assert.Empty(t, ToStringMap(map[stringerValue]int(nil)))
	})
}
