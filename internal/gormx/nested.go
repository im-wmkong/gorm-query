package gormx

import (
	"gorm.io/gorm"
)

func BuildNested[T ~func(*gorm.DB) *gorm.DB](db *gorm.DB, conds []T) *gorm.DB {
	nested := db.Session(&gorm.Session{NewDB: true})
	for _, cond := range conds {
		nested = cond(nested)
	}
	return nested
}
