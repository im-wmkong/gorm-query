package reflectx

import (
	"sync"
	"testing"

	"github.com/im-wmkong/gorm-query/example/model"
	"github.com/stretchr/testify/assert"
)

type localReflectModel struct {
	ID uint
}

func TestReflectxPackageName(t *testing.T) {
	t.Run("struct value", func(t *testing.T) {
		name, ok := PackageName(localReflectModel{})
		assert.True(t, ok)
		assert.Equal(t, "reflectx", name)
	})

	t.Run("struct pointer", func(t *testing.T) {
		name, ok := PackageName(&localReflectModel{})
		assert.True(t, ok)
		assert.Equal(t, "reflectx", name)
	})

	t.Run("double pointer", func(t *testing.T) {
		ptr := &localReflectModel{}
		name, ok := PackageName(&ptr)
		assert.True(t, ok)
		assert.Equal(t, "reflectx", name)
	})

	t.Run("external package", func(t *testing.T) {
		name, ok := PackageName(&model.User{})
		assert.True(t, ok)
		assert.Equal(t, "model", name)
	})

	t.Run("nil model", func(t *testing.T) {
		name, ok := PackageName(nil)
		assert.False(t, ok)
		assert.Empty(t, name)
	})

	t.Run("anonymous struct has no pkg path", func(t *testing.T) {
		name, ok := PackageName(struct{ ID uint }{})
		assert.False(t, ok)
		assert.Empty(t, name)
	})

	t.Run("built-in type has no pkg path", func(t *testing.T) {
		name, ok := PackageName(123)
		assert.False(t, ok)
		assert.Empty(t, name)
	})

	t.Run("top-level package without slash", func(t *testing.T) {
		// sync.Mutex has PkgPath() == "sync", which does not contain '/'.
		name, ok := PackageName(sync.Mutex{})
		assert.True(t, ok)
		assert.Equal(t, "sync", name)
	})
}
