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
