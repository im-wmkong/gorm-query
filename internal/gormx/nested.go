package gormx

import (
	"gorm.io/gorm"
)

func BuildNested[T ~func(*gorm.DB) *gorm.DB](db *gorm.DB, conds []T) *gorm.DB {
	return Apply(db.Session(&gorm.Session{NewDB: true}), conds)
}

// Apply runs each condition against db sequentially and returns the final *gorm.DB.
func Apply[T ~func(*gorm.DB) *gorm.DB](db *gorm.DB, conds []T) *gorm.DB {
	for _, cond := range conds {
		db = cond(db)
	}
	return db
}
