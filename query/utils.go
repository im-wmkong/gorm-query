package query

import "gorm.io/gorm"

// unwrap 解析参数
func unwrap(v any) any {
	if e, ok := v.(Column); ok {
		return string(e)
	}
	return v
}

// to T
func to[T any](v any) T {
	if res, ok := unwrap(v).(T); ok {
		return res
	}
	var zero T
	return zero
}

// unwraps 解析参数列表
func unwraps(args []any) []any {
	return mapSlice(args, unwrap)
}

// tos []T
func tos[T any](args []any) []T {
	return mapSlice(args, to[T])
}

func mapSlice[I any, O any](args []I, transform func(I) O) []O {
	if args == nil {
		return nil
	}
	result := make([]O, len(args))
	for i, a := range args {
		result[i] = transform(a)
	}
	return result
}

func buildNested(db *gorm.DB, conds []Condition) *gorm.DB {
	nestedDB := db.Session(&gorm.Session{NewDB: true})
	for _, cond := range conds {
		nestedDB = cond(nestedDB)
	}
	return nestedDB
}
