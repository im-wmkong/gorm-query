package test

import (
	"testing"

	"github.com/im-wmkong/gorm-query/example/model"
	"github.com/im-wmkong/gorm-query/internal/reflectx"
	"github.com/stretchr/testify/assert"
)

type localReflectModel struct {
	ID uint
}

func TestInternalReflectxPackageName(t *testing.T) {
	assert.Equal(t, "test", reflectx.PackageName(localReflectModel{}))
	assert.Equal(t, "test", reflectx.PackageName(&localReflectModel{}))
	assert.Equal(t, "model", reflectx.PackageName(&model.User{}))
}
