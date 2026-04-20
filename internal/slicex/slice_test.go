package slicex

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlicexMap(t *testing.T) {
	assert.Nil(t, Map(nil, func(v int) string {
		return fmt.Sprint(v)
	}))

	assert.Equal(t, []string{"2", "4", "6"}, Map([]int{1, 2, 3}, func(v int) string {
		return fmt.Sprint(v * 2)
	}))
}
