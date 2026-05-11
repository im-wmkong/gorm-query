// Package codegen provides jen-based code generation helpers that render
// type-safe column/association declarations for the query package.
package codegen

import (
	"reflect"
	"time"
	"unicode"

	"github.com/dave/jennifer/jen"
	"github.com/im-wmkong/gorm-query/internal/reflectx"
)

// QueryPkg is the import path of the query package that generated code
// references.
const QueryPkg = "github.com/im-wmkong/gorm-query/query"

// ColumnKind classifies a Go type into the column category used by the query
// package.
type ColumnKind int

const (
	KindValue ColumnKind = iota
	KindString
	KindNumeric
	KindTime
	KindBool
)

// ClassifyColumn picks the column kind for a given Go type, unwrapping pointers.
func ClassifyColumn(t reflect.Type) (ColumnKind, reflect.Type) {
	t = reflectx.UnwrapPtr(t)
	if t == reflect.TypeOf(time.Time{}) {
		return KindTime, t
	}
	switch t.Kind() {
	case reflect.Bool:
		return KindBool, t
	case reflect.String:
		return KindString, t
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return KindNumeric, t
	default:
		return KindValue, t
	}
}

// ColumnTypeCode generates "query.XxxColumn[T]" (or "query.TimeColumn"/"query.BoolColumn").
func ColumnTypeCode(t reflect.Type) *jen.Statement {
	kind, inner := ClassifyColumn(t)
	switch kind {
	case KindTime:
		return jen.Qual(QueryPkg, "TimeColumn")
	case KindBool:
		return jen.Qual(QueryPkg, "BoolColumn")
	case KindString:
		return jen.Qual(QueryPkg, "StringColumn").Types(GoTypeCode(inner))
	case KindNumeric:
		return jen.Qual(QueryPkg, "NumericColumn").Types(GoTypeCode(inner))
	default:
		return jen.Qual(QueryPkg, "ValueColumn").Types(GoTypeCode(inner))
	}
}

// ColumnCtorCode generates the constructor call for a column, e.g.
// query.NewStringColumn[model.UserStatus]("users", "status").
func ColumnCtorCode(t reflect.Type, table, name string) *jen.Statement {
	kind, inner := ClassifyColumn(t)
	switch kind {
	case KindTime:
		return jen.Qual(QueryPkg, "NewTimeColumn").Call(jen.Lit(table), jen.Lit(name))
	case KindBool:
		return jen.Qual(QueryPkg, "NewBoolColumn").Call(jen.Lit(table), jen.Lit(name))
	case KindString:
		return jen.Qual(QueryPkg, "NewStringColumn").Types(GoTypeCode(inner)).Call(jen.Lit(table), jen.Lit(name))
	case KindNumeric:
		return jen.Qual(QueryPkg, "NewNumericColumn").Types(GoTypeCode(inner)).Call(jen.Lit(table), jen.Lit(name))
	default:
		return jen.Qual(QueryPkg, "NewValueColumn").Types(GoTypeCode(inner)).Call(jen.Lit(table), jen.Lit(name))
	}
}

// AssociationTypeCode generates "query.Association[Parent, Child]".
func AssociationTypeCode(parent, child reflect.Type) *jen.Statement {
	return jen.Qual(QueryPkg, "Association").Types(GoTypeCode(parent), GoTypeCode(child))
}

// AssociationCtorCode generates "query.NewAssociation[Parent, Child]("Name")".
func AssociationCtorCode(parent, child reflect.Type, name string) *jen.Statement {
	return jen.Qual(QueryPkg, "NewAssociation").Types(GoTypeCode(parent), GoTypeCode(child)).Call(jen.Lit(name))
}

// GoTypeCode renders a reflect.Type as a jen type reference, handling named
// package types (like model.UserStatus) and built-in primitives uniformly.
// Pointer / composite wrappers have been stripped before this call.
func GoTypeCode(t reflect.Type) *jen.Statement {
	if t == nil {
		return jen.Any()
	}
	if name := t.Name(); name != "" && t.PkgPath() != "" {
		return jen.Qual(t.PkgPath(), name)
	}
	// Built-in primitive or anonymous type: fall back to its Go syntax string.
	return jen.Id(t.String())
}

// UnexportName derives the unexported identifier for a name by lower-casing
// its first rune. Empty names are returned verbatim.
func UnexportName(name string) string {
	if name == "" {
		return ""
	}
	runes := []rune(name)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}
