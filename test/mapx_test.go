package test

import (
	"fmt"
	"testing"

	"github.com/im-wmkong/gorm-query/internal/mapx"
	"github.com/stretchr/testify/assert"
)

func TestInternalMapxMap(t *testing.T) {
	t.Run("MapKeys converts keys and keeps values", func(t *testing.T) {
		assert.Nil(t, mapx.MapKeys(map[int]int(nil), func(key int) string {
			return fmt.Sprint(key)
		}))

		assert.Equal(t, map[string]int{"1": 10, "2": 20}, mapx.MapKeys(map[int]int{1: 10, 2: 20}, func(key int) string {
			return fmt.Sprint(key)
		}))
	})
}
