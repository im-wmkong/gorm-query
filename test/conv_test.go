package test

import (
	"testing"

	"github.com/im-wmkong/gorm-query/internal/conv"
	"github.com/im-wmkong/gorm-query/query"
	"github.com/stretchr/testify/assert"
)

func TestInternalConv(t *testing.T) {
	t.Run("ToStringMap converts map to string values", func(t *testing.T) {
		values := map[query.Column]int{
			query.Column("age"):    18,
			query.Column("status"): 1,
		}

		assert.Equal(t, map[string]int{"age": 18, "status": 1}, conv.ToStringMap(values))
		assert.Empty(t, conv.ToStringMap(map[query.Column]int(nil)))
	})
}
