package test

import (
	"fmt"
	"testing"

	"github.com/im-wmkong/gorm-query/internal/slices"
	"github.com/stretchr/testify/assert"
)

func TestInternalSlicesMap(t *testing.T) {
	assert.Nil(t, slices.Map(nil, func(v int) string {
		return fmt.Sprint(v)
	}))

	assert.Equal(t, []string{"2", "4", "6"}, slices.Map([]int{1, 2, 3}, func(v int) string {
		return fmt.Sprint(v * 2)
	}))
}
