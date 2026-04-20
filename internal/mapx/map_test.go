package mapx

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapxMap(t *testing.T) {
	t.Run("MapKeys converts keys and keeps values", func(t *testing.T) {
		assert.Nil(t, MapKeys(map[int]int(nil), func(key int) string {
			return fmt.Sprint(key)
		}))

		assert.Equal(t, map[string]int{"1": 10, "2": 20}, MapKeys(map[int]int{1: 10, 2: 20}, func(key int) string {
			return fmt.Sprint(key)
		}))
	})
}
