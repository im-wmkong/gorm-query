// Package fsx provides filesystem helpers for Go source packages.
package fsx

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// ReadPackageNameFromDir reads non-test .go files under dir and returns the
// declared package name. It returns an empty string (with nil error) when no
// valid Go source file is found.
//
// It returns an error when:
//   - the directory cannot be read;
//   - two different package names are declared in the same directory.
//
// Unparseable .go files are skipped silently to tolerate work-in-progress files.
func ReadPackageNameFromDir(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	fset := token.NewFileSet()

	var pkg string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		path := filepath.Join(dir, name)
		file, err := parser.ParseFile(fset, path, nil, parser.PackageClauseOnly)
		if err != nil {
			// Ignore unparseable files (e.g. work-in-progress source).
			continue
		}

		detected := file.Name.Name
		if pkg != "" && pkg != detected {
			return "", fmt.Errorf("multiple packages found in %s: %q and %q", dir, pkg, detected)
		}
		pkg = detected
	}

	return pkg, nil
}
