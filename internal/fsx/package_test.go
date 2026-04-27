package fsx

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFsxReadPackageNameFromDir_SamePackageAcrossFilesAndSubdirs(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "nested"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.go"), []byte("package sample\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.go"), []byte("/* comment */ package sample\n"), 0o644))

	pkg, err := ReadPackageNameFromDir(dir)
	require.NoError(t, err)
	assert.Equal(t, "sample", pkg)
}

func TestFsxReadPackageNameFromDir_SkipsUnsupportedEntries(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "helper_test.go"), []byte("package ignored\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("plain text\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "broken.go"), []byte("var x = 1\n"), 0o644))
	require.NoError(t, os.Symlink(filepath.Join(dir, "missing.go"), filepath.Join(dir, "missing_link.go")))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "real.go"), []byte("package actual\n"), 0o644))

	pkg, err := ReadPackageNameFromDir(dir)
	require.NoError(t, err)
	assert.Equal(t, "actual", pkg)
}

func TestFsxReadPackageNameFromDir_HandlesMultilineBlockComments(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.go"), []byte("/* block\ncomment\n*/\npackage sample\n"), 0o644))

	pkg, err := ReadPackageNameFromDir(dir)
	require.NoError(t, err)
	assert.Equal(t, "sample", pkg)
}

func TestFsxReadPackageNameFromDir_IgnoresStringLiteralsThatLookLikePackage(t *testing.T) {
	dir := t.TempDir()
	// The `package "fake"` text appears only inside a string literal and a block comment in a valid
	// Go source file; go/parser must still resolve the actual declared package.
	src := "/* package fake */\npackage real\n\nvar _ = \"package fake\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.go"), []byte(src), 0o644))

	pkg, err := ReadPackageNameFromDir(dir)
	require.NoError(t, err)
	assert.Equal(t, "real", pkg)
}

func TestFsxReadPackageNameFromDir_ReturnsEmptyForDirWithoutGoFiles(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("plain text\n"), 0o644))

	pkg, err := ReadPackageNameFromDir(dir)
	require.NoError(t, err)
	assert.Empty(t, pkg)
}

func TestFsxReadPackageNameFromDir_ReturnsErrorForMissingDir(t *testing.T) {
	pkg, err := ReadPackageNameFromDir(filepath.Join(t.TempDir(), "missing"))
	require.Error(t, err)
	assert.Empty(t, pkg)
}

func TestFsxReadPackageNameFromDir_ReturnsErrorForMultiplePackages(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.go"), []byte("package first\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.go"), []byte("package second\n"), 0o644))

	pkg, err := ReadPackageNameFromDir(dir)
	require.Error(t, err)
	assert.Empty(t, pkg)
	assert.Contains(t, err.Error(), "multiple packages found")
}
