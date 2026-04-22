package model

import "gorm.io/gorm"

// User defines the demo user model.
type User struct {
	gorm.Model
	UserName string `gorm:"column:user_name;size:255;not null"`
	Email    string `gorm:"column:email;size:255;unique"`
	Age      int    `gorm:"column:age"`
	Status   int    `gorm:"column:status;default:1"` // 1: active, 0: inactive
}

// TableName overrides the default table name to `users`.
func (u *User) TableName() string {
	return "users"
}
