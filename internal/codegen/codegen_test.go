package codegen

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type stringAlias string

type customStruct struct {
	Field string
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
