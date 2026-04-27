package fsx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFsxInferPackageNameFromPath(t *testing.T) {
	t.Run("returns base name when it is a valid Go identifier", func(t *testing.T) {
		assert.Equal(t, "columns", InferPackageNameFromPath("./path/to/columns"))
		assert.Equal(t, "mypkg", InferPackageNameFromPath("mypkg"))
		assert.Equal(t, "pkg_v2", InferPackageNameFromPath("/abs/pkg_v2"))
	})

	t.Run("returns empty when base name is not a valid identifier", func(t *testing.T) {
		assert.Empty(t, InferPackageNameFromPath("./invalid-dir"))
		assert.Empty(t, InferPackageNameFromPath("./123start"))
		assert.Empty(t, InferPackageNameFromPath(""))
	})
}
