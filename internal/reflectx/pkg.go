package reflectx

import (
	"reflect"
	"strings"
)

func PackageName(model interface{}) string {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	parts := strings.Split(t.PkgPath(), "/")
	return parts[len(parts)-1]
}
