package test

import (
	"fmt"
	"testing"

	"github.com/im-wmkong/gorm-query/internal/slicex"
	"github.com/stretchr/testify/assert"
)

func TestInternalSlicexMap(t *testing.T) {
	assert.Nil(t, slicex.Map(nil, func(v int) string {
		return fmt.Sprint(v)
	}))

	assert.Equal(t, []string{"2", "4", "6"}, slicex.Map([]int{1, 2, 3}, func(v int) string {
		return fmt.Sprint(v * 2)
	}))
}
