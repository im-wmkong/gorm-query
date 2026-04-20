package reflectx

import (
	"testing"

	"github.com/im-wmkong/gorm-query/example/model"
	"github.com/stretchr/testify/assert"
)

type localReflectModel struct {
	ID uint
}

func TestReflectxPackageName(t *testing.T) {
	assert.Equal(t, "reflectx", PackageName(localReflectModel{}))
	assert.Equal(t, "reflectx", PackageName(&localReflectModel{}))
	assert.Equal(t, "model", PackageName(&model.User{}))
}
