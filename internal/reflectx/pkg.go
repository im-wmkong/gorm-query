// Package reflectx provides reflection helpers used internally by gorm-query.
package reflectx

import (
	"reflect"
	"strings"
)

// UnwrapPtr returns the underlying type by repeatedly stripping pointer
// indirection. It returns t unchanged if it is not a pointer, and nil if t
// is nil.
func UnwrapPtr(t reflect.Type) reflect.Type {
	for t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

// PackageName returns the last segment of the package path of model's type.
//
// It returns ok=false when:
//   - model is nil;
//   - the type has no package path (e.g. anonymous structs, built-in types).
func PackageName(model any) (name string, ok bool) {
	if model == nil {
		return "", false
	}
	t := reflect.TypeOf(model)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	pkgPath := t.PkgPath()
	if pkgPath == "" {
		return "", false
	}
	if idx := strings.LastIndex(pkgPath, "/"); idx >= 0 {
		return pkgPath[idx+1:], true
	}
	return pkgPath, true
}
