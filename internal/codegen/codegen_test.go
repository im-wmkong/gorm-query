package codegen

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/dave/jennifer/jen"
	"github.com/stretchr/testify/assert"
)

type stringAlias string

type customStruct struct {
	Field string
}

func render(s *jen.Statement) string {
	f := jen.NewFile("dummy")
	f.Add(jen.Var().Id("_").Op("=").Add(s))
	return f.GoString()
}

func TestClassifyColumn(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want ColumnKind
	}{
		{"time", time.Time{}, KindTime},
		{"time pointer", &time.Time{}, KindTime},
		{"bool", true, KindBool},
		{"string", "x", KindString},
		{"string alias", stringAlias("x"), KindString},
		{"int", 1, KindNumeric},
		{"int8", int8(1), KindNumeric},
		{"uint64", uint64(1), KindNumeric},
		{"float64", 1.0, KindNumeric},
		{"struct", customStruct{}, KindValue},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kind, _ := ClassifyColumn(reflect.TypeOf(tc.in))
			assert.Equal(t, tc.want, kind)
		})
	}
}

func TestColumnTypeCode(t *testing.T) {
	t.Run("time", func(t *testing.T) {
		got := render(ColumnTypeCode(reflect.TypeOf(time.Time{})))
		assert.Contains(t, got, "TimeColumn")
		assert.NotContains(t, got, "TimeColumn[")
	})
	t.Run("bool", func(t *testing.T) {
		got := render(ColumnTypeCode(reflect.TypeOf(true)))
		assert.Contains(t, got, "BoolColumn")
		assert.NotContains(t, got, "BoolColumn[")
	})
	t.Run("string", func(t *testing.T) {
		got := render(ColumnTypeCode(reflect.TypeOf("")))
		assert.Contains(t, got, "StringColumn[string]")
	})
	t.Run("numeric", func(t *testing.T) {
		got := render(ColumnTypeCode(reflect.TypeOf(int64(0))))
		assert.Contains(t, got, "NumericColumn[int64]")
	})
	t.Run("value", func(t *testing.T) {
		got := render(ColumnTypeCode(reflect.TypeOf(customStruct{})))
		assert.Contains(t, got, "ValueColumn[")
	})
}

func TestColumnCtorCode(t *testing.T) {
	t.Run("time", func(t *testing.T) {
		got := render(ColumnCtorCode(reflect.TypeOf(time.Time{}), "users", "created_at"))
		assert.Contains(t, got, `NewTimeColumn("users", "created_at")`)
	})
	t.Run("bool", func(t *testing.T) {
		got := render(ColumnCtorCode(reflect.TypeOf(true), "users", "active"))
		assert.Contains(t, got, `NewBoolColumn("users", "active")`)
	})
	t.Run("string", func(t *testing.T) {
		got := render(ColumnCtorCode(reflect.TypeOf(""), "users", "name"))
		assert.Contains(t, got, `NewStringColumn[string]("users", "name")`)
	})
	t.Run("numeric", func(t *testing.T) {
		got := render(ColumnCtorCode(reflect.TypeOf(int64(0)), "users", "id"))
		assert.Contains(t, got, `NewNumericColumn[int64]("users", "id")`)
	})
	t.Run("value", func(t *testing.T) {
		got := render(ColumnCtorCode(reflect.TypeOf(customStruct{}), "users", "data"))
		assert.Contains(t, got, `NewValueColumn[`)
		assert.Contains(t, got, `("users", "data")`)
	})
}

func TestAssociationCode(t *testing.T) {
	parent := reflect.TypeOf(customStruct{})
	child := reflect.TypeOf(customStruct{})

	typeCode := render(AssociationTypeCode(parent, child))
	assert.Contains(t, typeCode, "Association[")

	ctorCode := render(AssociationCtorCode(parent, child, "Profile"))
	assert.Contains(t, ctorCode, `NewAssociation[`)
	assert.Contains(t, ctorCode, `]("Profile")`)
}

func TestGoTypeCode(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		assert.Contains(t, render(GoTypeCode(nil)), "any")
	})
	t.Run("primitive", func(t *testing.T) {
		assert.Contains(t, render(GoTypeCode(reflect.TypeOf(0))), "int")
	})
	t.Run("named package type", func(t *testing.T) {
		got := render(GoTypeCode(reflect.TypeOf(customStruct{})))
		assert.Contains(t, got, "customStruct")
		assert.Contains(t, got, "internal/codegen")
	})
}

func TestUnexportName(t *testing.T) {
	assert.Equal(t, "", UnexportName(""))
	assert.Equal(t, "user", UnexportName("User"))
	assert.Equal(t, "userProfile", UnexportName("UserProfile"))
	assert.Equal(t, "u", UnexportName("U"))
	// already lowercase stays the same
	assert.Equal(t, "user", UnexportName("user"))
	// non-ascii first rune is lowered if possible
	assert.True(t, strings.HasPrefix(UnexportName("ÄField"), "ä"))
}
