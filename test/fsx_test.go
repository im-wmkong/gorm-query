package test

import (
	"os"
	"path/filepath"
	"testing"
	_ "unsafe"

	"github.com/im-wmkong/gorm-query/internal/fsx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:linkname linkedPackageNameFromFile github.com/im-wmkong/gorm-query/internal/fsx.packageNameFromFile
func linkedPackageNameFromFile(file *os.File) (string, error)

func TestFsxPackageNameFromFile(t *testing.T) {
	t.Run("parses package after inline block comment", func(t *testing.T) {
		filePath := filepath.Join(t.TempDir(), "inline.go")
		require.NoError(t, os.WriteFile(filePath, []byte("/* block */ package sample\n"), 0o644))

		file, err := os.Open(filePath)
		require.NoError(t, err)
		defer file.Close()

		pkg, err := linkedPackageNameFromFile(file)
		require.NoError(t, err)
		assert.Equal(t, "sample", pkg)
	})

	t.Run("parses package after multiline block comment closes on same line", func(t *testing.T) {
		filePath := filepath.Join(t.TempDir(), "multiline.go")
		content := "/* block comment\ncontinued */ package sample\n"
		require.NoError(t, os.WriteFile(filePath, []byte(content), 0o644))

		file, err := os.Open(filePath)
		require.NoError(t, err)
		defer file.Close()

		pkg, err := linkedPackageNameFromFile(file)
		require.NoError(t, err)
		assert.Equal(t, "sample", pkg)
	})

	t.Run("parses package after multiline block comment with standalone closing line", func(t *testing.T) {
		filePath := filepath.Join(t.TempDir(), "closed.go")
		content := "/* block comment\n*/\npackage sample\n"
		require.NoError(t, os.WriteFile(filePath, []byte(content), 0o644))

		file, err := os.Open(filePath)
		require.NoError(t, err)
		defer file.Close()

		pkg, err := linkedPackageNameFromFile(file)
		require.NoError(t, err)
		assert.Equal(t, "sample", pkg)
	})

	t.Run("parses package after long multiline block comment", func(t *testing.T) {
		filePath := filepath.Join(t.TempDir(), "long.go")
		content := "/* block comment\ncontinued\n*/\npackage sample\n"
		require.NoError(t, os.WriteFile(filePath, []byte(content), 0o644))

		file, err := os.Open(filePath)
		require.NoError(t, err)
		defer file.Close()

		pkg, err := linkedPackageNameFromFile(file)
		require.NoError(t, err)
		assert.Equal(t, "sample", pkg)
	})

	t.Run("returns empty after first non package code", func(t *testing.T) {
		filePath := filepath.Join(t.TempDir(), "nopkg.go")
		content := "/* block */\nvar x = 1\npackage ignored\n"
		require.NoError(t, os.WriteFile(filePath, []byte(content), 0o644))

		file, err := os.Open(filePath)
		require.NoError(t, err)
		defer file.Close()

		pkg, err := linkedPackageNameFromFile(file)
		require.NoError(t, err)
		assert.Empty(t, pkg)
	})

	t.Run("returns scanner error for closed file", func(t *testing.T) {
		filePath := filepath.Join(t.TempDir(), "closed-fd.go")
		require.NoError(t, os.WriteFile(filePath, []byte("package sample\n"), 0o644))

		file, err := os.Open(filePath)
		require.NoError(t, err)
		require.NoError(t, file.Close())

		pkg, err := linkedPackageNameFromFile(file)
		require.Error(t, err)
		assert.Empty(t, pkg)
	})
}

func TestFsxPackageNameFromDir_SamePackageAcrossFilesAndSubdirs(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "nested"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.go"), []byte("package sample\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.go"), []byte("/* comment */ package sample\n"), 0o644))

	pkg, err := fsx.PackageNameFromDir(dir)
	require.NoError(t, err)
	assert.Equal(t, "sample", pkg)
}

func TestFsxPackageNameFromDir_SkipsUnsupportedEntries(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "helper_test.go"), []byte("package ignored\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("plain text\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "broken.go"), []byte("var x = 1\n"), 0o644))
	require.NoError(t, os.Symlink(filepath.Join(dir, "missing.go"), filepath.Join(dir, "missing_link.go")))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "real.go"), []byte("package actual\n"), 0o644))

	pkg, err := fsx.PackageNameFromDir(dir)
	require.NoError(t, err)
	assert.Equal(t, "actual", pkg)
}

func TestFsxPackageNameFromDir_ReturnsErrorForMissingDir(t *testing.T) {
	pkg, err := fsx.PackageNameFromDir(filepath.Join(t.TempDir(), "missing"))
	require.Error(t, err)
	assert.Empty(t, pkg)
}

func TestFsxPackageNameFromDir_ReturnsErrorForMultiplePackages(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.go"), []byte("package first\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.go"), []byte("package second\n"), 0o644))

	pkg, err := fsx.PackageNameFromDir(dir)
	require.Error(t, err)
	assert.Empty(t, pkg)
	assert.Contains(t, err.Error(), "multiple packages found")
}
