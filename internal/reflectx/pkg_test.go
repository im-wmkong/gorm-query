package reflectx

import (
	"reflect"
	"sync"
	"testing"

	"github.com/im-wmkong/gorm-query/example/model"
	"github.com/stretchr/testify/assert"
)

type localReflectModel struct {
	ID uint
}

func TestUnwrapPtr(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		assert.Nil(t, UnwrapPtr(nil))
	})

	t.Run("non pointer returns same type", func(t *testing.T) {
		typ := reflect.TypeOf(0)
		assert.Equal(t, typ, UnwrapPtr(typ))
	})

	t.Run("single pointer is unwrapped", func(t *testing.T) {
		typ := reflect.TypeOf(&localReflectModel{})
		assert.Equal(t, reflect.TypeOf(localReflectModel{}), UnwrapPtr(typ))
	})

	t.Run("multiple pointers are unwrapped", func(t *testing.T) {
		ptr := &localReflectModel{}
		typ := reflect.TypeOf(&ptr)
		assert.Equal(t, reflect.TypeOf(localReflectModel{}), UnwrapPtr(typ))
	})
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
