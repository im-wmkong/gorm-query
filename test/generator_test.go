package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/im-wmkong/gorm-query/schemagen"
	"github.com/im-wmkong/gorm-query/test/model"
	other "github.com/im-wmkong/gorm-query/test/packageconflict/model"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm/schema"
)

func TestContract_GeneratorValidatesBeforeWriting(t *testing.T) {
	for _, tc := range []struct {
		name    string
		models  []any
		opts    []schemagen.Option
		message string
	}{
		{"full_package_identity", []any{&model.Record{}, &other.Record{}}, nil, "same package"},
		{"output_collision", []any{&model.Record{}, &model.Explicit{}}, []schemagen.Option{schemagen.WithNamingStrategy(schema.NamingStrategy{NameReplacer: strings.NewReplacer("Explicit", "Record")})}, "collision"},
		{"member_collision", []any{&model.Record{}, &model.AmbiguousCity{}}, nil, "collision"},
		{"invalid_package", []any{&model.Record{}}, []schemagen.Option{schemagen.WithPackageName("invalid-package")}, "package name"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			original := []byte("package generated\n// existing content must survive validation errors\n")
			path := filepath.Join(dir, "record_gen.go")
			require.NoError(t, os.WriteFile(path, original, 0644))
			opts := []schemagen.Option{schemagen.WithOutputDir(dir), schemagen.WithPackageName("generated"), schemagen.WithLogger(nil)}
			err := schemagen.New(append(opts, tc.opts...)...).Generate(tc.models...)
			require.ErrorContains(t, err, tc.message)
			actual, err := os.ReadFile(path)
			require.NoError(t, err)
			require.Equal(t, original, actual)
			entries, err := os.ReadDir(dir)
			require.NoError(t, err)
			require.Len(t, entries, 1)
		})
	}
}

func TestContract_GeneratorDryRunDoesNotCreateDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "missing")
	err := schemagen.New(schemagen.WithOutputDir(dir), schemagen.WithDryRun(true), schemagen.WithLogger(nil)).Generate(&model.Record{})
	require.Error(t, err)
	_, err = os.Stat(dir)
	require.ErrorIs(t, err, os.ErrNotExist)
}
