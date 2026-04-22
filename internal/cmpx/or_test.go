package cmpx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type contextKey string

func TestOr(t *testing.T) {
	t.Run("returns first non-zero int", func(t *testing.T) {
		assert.Equal(t, 3, Or(0, 0, 3, 4))
	})

	t.Run("returns first non-empty string", func(t *testing.T) {
		assert.Equal(t, "gorm-query", Or("", "gorm-query", "fallback"))
	})

	t.Run("returns zero value when all values are zero", func(t *testing.T) {
		assert.Equal(t, 0, Or(0, 0, 0))
		assert.Equal(t, "", Or("", ""))
	})

	t.Run("supports pointer types", func(t *testing.T) {
		first := 1
		second := 2
		assert.Same(t, &first, Or[*int](nil, &first, &second))
	})

	t.Run("supports interface values like context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), contextKey("k"), "v")
		assert.Same(t, ctx, Or[context.Context](nil, ctx, context.TODO()))
	})

	t.Run("returns zero value for empty input", func(t *testing.T) {
		assert.Equal(t, 0, Or[int]())
		assert.Nil(t, Or[*int]())
	})
}
