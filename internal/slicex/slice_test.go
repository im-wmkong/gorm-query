package slicex

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlicexMap(t *testing.T) {
	assert.Equal(t, []string{"2", "4", "6"}, Map([]int{1, 2, 3}, func(v int) string {
		return fmt.Sprint(v * 2)
	}))

	assert.Nil(t, Map[int, string]([]int(nil), func(v int) string {
		return fmt.Sprint(v)
	}))
}

func TestSlicexFilter(t *testing.T) {
	assert.Equal(t, []int{1, 2, 3}, Filter([]int{1, 2, 3}, func(v int) bool {
		return v > 0
	}))

	assert.Equal(t, []int{2, 3}, Filter([]int{1, 2, 3}, func(v int) bool {
		return v > 1
	}))

	assert.Nil(t, Filter([]int(nil), func(v int) bool {
		return v > 0
	}))
}

func TestSlicexToMap(t *testing.T) {
	t.Run("nil slice returns nil", func(t *testing.T) {
		got := ToMap[int, int, string]([]int(nil), func(v int) (int, string) {
			return v, fmt.Sprint(v)
		})
		assert.Nil(t, got)
	})

	t.Run("empty slice returns nil", func(t *testing.T) {
		got := ToMap([]int{}, func(v int) (int, string) {
			return v, fmt.Sprint(v)
		})
		assert.Nil(t, got)
	})

	t.Run("maps each element", func(t *testing.T) {
		got := ToMap([]int{1, 2, 3}, func(v int) (string, int) {
			return fmt.Sprint(v), v * 10
		})
		assert.Equal(t, map[string]int{"1": 10, "2": 20, "3": 30}, got)
	})

	t.Run("later duplicates override earlier ones", func(t *testing.T) {
		got := ToMap([]int{1, 2, 3, 2}, func(v int) (int, string) {
			return v, fmt.Sprintf("v=%d", v)
		})
		assert.Equal(t, map[int]string{1: "v=1", 2: "v=2", 3: "v=3"}, got)

		// Confirm the "later wins" path actually executes when two items
		// share a key but disagree on the value.
		dup := ToMap([]struct{ k, v string }{{"k", "first"}, {"k", "second"}}, func(it struct{ k, v string }) (string, string) {
			return it.k, it.v
		})
		assert.Equal(t, map[string]string{"k": "second"}, dup)
	})
}
