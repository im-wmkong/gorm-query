package fsx

import (
	"go/token"
	"path/filepath"
)

// InferPackageNameFromPath infers a Go package name from the last segment of
// the given directory path without reading any files. It returns an empty
// string if the base name is not a valid Go identifier.
func InferPackageNameFromPath(dir string) string {
	base := filepath.Base(filepath.Clean(dir))
	if token.IsIdentifier(base) {
		return base
	}
	return ""
}
