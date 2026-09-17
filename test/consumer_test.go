package test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/im-wmkong/gorm-query/schemagen"
	"github.com/im-wmkong/gorm-query/test/model"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm/schema"
)

// consumerDir creates an isolated module using the current checkout and its
// dependency versions. Compiler failures are reported by the calling subtest.
func consumerDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	root := filepath.Dir(wd)
	dir := t.TempDir()
	mod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	require.NoError(t, err)
	content := strings.Replace(string(mod), "module github.com/im-wmkong/gorm-query", "module example.com/consumer", 1)
	content += fmt.Sprintf("\nrequire github.com/im-wmkong/gorm-query v0.0.0\nreplace github.com/im-wmkong/gorm-query => %q\n", root)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0644))
	sums, err := os.ReadFile(filepath.Join(root, "go.sum"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.sum"), sums, 0644))
	return dir
}

func runConsumer(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local", "GOPROXY=off", "GOFLAGS=")
	output, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("consumer command timed out: %s", output)
	}
	return string(output), err
}

func TestContract_PublicAPICompilation(t *testing.T) {
	const imports = `package consumer

import (
	"github.com/im-wmkong/gorm-query/example/model"
	"github.com/im-wmkong/gorm-query/example/model/schema"
	"github.com/im-wmkong/gorm-query/query"
)

var _ = query.New[model.User]
var _ = schema.User

func use() {
`
	cases := []struct {
		name    string
		body    string
		compile bool
	}{
		{"valid_api_control", `_ = schema.User.Query().Where(schema.User.Age.Gt(18)).Preload(schema.User.Profile)`, true},
		{"aggregate_string_compatibility", `agg := query.AggFragment("SUM(age)"); _ = string(agg); _ = schema.User.Query().Select(agg.As("total"))`, true},
		{"numeric_alias", `_ = schema.User.Age.WithTable("u").Gt(18)`, true},
		{"string_alias", `_ = schema.User.Email.WithTable("u").Like("a%")`, true},
		{"time_alias", `_ = schema.User.CreatedAt.WithTable("u").Set(model.User{}.CreatedAt)`, true},
		{"bool_alias", `_ = query.NewBoolColumn("users", "active").WithTable("u").IsTrue()`, true},
		{"value_alias", `_ = schema.User.DeletedAt.WithTable("u").Eq(model.User{}.DeletedAt)`, true},
		{"reject_alias_wrong_value", `_ = schema.User.Age.WithTable("u").Eq("eighteen")`, false},
		{"reject_alias_wrong_assignment", `_ = schema.User.Age.WithTable("u").Set("eighteen")`, false},
		{"reject_wrong_value_type", `_ = schema.User.Age.Eq("eighteen")`, false},
		{"reject_wrong_assignment_type", `_ = schema.User.Age.Set("eighteen")`, false},
		{"reject_wrong_association_parent", `_ = schema.Profile.Query().Preload(schema.User.Profile)`, false},
		{"reject_wrong_nested_parent", `_ = schema.User.Profile.Nested(schema.User.Profile)`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := consumerDir(t)
			require.NoError(t, os.WriteFile(filepath.Join(dir, "consumer.go"), []byte(imports+tc.body+"\n}\n"), 0644))
			output, err := runConsumer(t, dir, "build", "-mod=readonly", ".")
			if tc.compile {
				require.NoError(t, err, "public API must compile:\n%s", output)
			} else {
				require.Error(t, err, "invalid API use must fail compilation")
				require.Contains(t, output, "cannot use", "must fail type checking, not dependency setup")
			}
		})
	}
}

func TestContract_GeneratedSchemaCompilesAndRuns(t *testing.T) {
	cases := []struct {
		name    string
		model   any
		naming  *schema.NamingStrategy
		fixture string
	}{
		{"default_naming", &model.Record{}, nil, "record"},
		{"explicit_singular", &model.Record{}, &schema.NamingStrategy{SingularTable: true}, "record"},
		{"custom_prefix", &model.Record{}, &schema.NamingStrategy{TablePrefix: "app_"}, "record"},
		{"explicit_table", &model.Explicit{}, nil, "explicit"},
		{"column_types", &model.TypedRecord{}, &schema.NamingStrategy{}, "typed"},
		{"embedded_fields", &model.Shipment{}, nil, "embedded"},
		{"reserved_identifiers_with_scope", &model.Reserved{}, nil, "reserved"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := consumerDir(t)
			opts := []schemagen.Option{
				schemagen.WithOutputDir(filepath.Join(dir, "schema")),
				schemagen.WithLogger(nil),
			}
			if tc.naming != nil {
				opts = append(opts, schemagen.WithNamingStrategy(*tc.naming))
			}
			require.NoError(t, schemagen.New(opts...).Generate(tc.model))
			if tc.name == "embedded_fields" {
				require.NoError(t, schemagen.New(opts...).Generate(&model.ReorderedShipment{}, &model.CityCollision{}, &model.CityOverride{}))
			}
			source, err := os.ReadFile(filepath.Join("consumer", tc.fixture+"_test.go.tmpl"))
			require.NoError(t, err)
			var naming schema.NamingStrategy
			if tc.naming != nil {
				naming = *tc.naming
			}
			config := fmt.Sprintf("schema.NamingStrategy{SingularTable: %t, TablePrefix: %q}", naming.SingularTable, naming.TablePrefix)
			source = []byte(strings.ReplaceAll(string(source), "NAMING_STRATEGY", config))
			require.NoError(t, os.WriteFile(filepath.Join(dir, "consumer_test.go"), source, 0644))
			// Subprocesses do not inherit the parent go test's race instrumentation.
			output, err := runConsumer(t, dir, "test", "-race", "-mod=readonly", "-count=1", "-timeout=30s", ".")
			require.NoError(t, err, "generated schema must compile and execute:\n%s", output)
		})
	}
}
