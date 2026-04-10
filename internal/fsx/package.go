package fsx

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func PackageNameFromDir(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	var pkg string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		file, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			continue
		}

		detected, _ := packageNameFromFile(file)
		_ = file.Close()
		if detected == "" {
			continue
		}
		if pkg != "" && pkg != detected {
			return "", fmt.Errorf("multiple packages found in %s", dir)
		}
		pkg = detected
	}

	return pkg, nil
}

func packageNameFromFile(file *os.File) (string, error) {
	scanner := bufio.NewScanner(file)
	inBlockComment := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		for {
			switch {
			case inBlockComment:
				end := strings.Index(line, "*/")
				if end == -1 {
					line = ""
					break
				}
				line = strings.TrimSpace(line[end+2:])
				inBlockComment = false
				if line == "" {
					break
				}
				continue
			case strings.HasPrefix(line, "/*"):
				end := strings.Index(line, "*/")
				if end == -1 {
					inBlockComment = true
					line = ""
					break
				}
				line = strings.TrimSpace(line[end+2:])
				if line == "" {
					break
				}
				continue
			}
			break
		}

		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		if strings.HasPrefix(line, "package ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
		break
	}

	return "", scanner.Err()
}
