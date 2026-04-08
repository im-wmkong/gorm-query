package genprops

import (
	"bufio"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

func getPkgName(model interface{}) string {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	pkgPath := t.PkgPath()
	parts := strings.Split(pkgPath, "/")
	return parts[len(parts)-1]
}

func detectPackageName(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		filePath := filepath.Join(dir, name)
		f, err := os.Open(filePath)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())

			if strings.HasPrefix(line, "package ") {
				f.Close()
				return strings.TrimPrefix(line, "package "), nil
			}

			if line != "" && !strings.HasPrefix(line, "//") {
				break
			}
		}

		f.Close()
	}

	return "", nil
}
