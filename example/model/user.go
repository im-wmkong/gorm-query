package model

import "gorm.io/gorm"

// User 定义了用户模型
type User struct {
	gorm.Model
	UserName string `gorm:"column:user_name;size:255;not null"`
	Email    string `gorm:"column:email;size:255;unique"`
	Age      int    `gorm:"column:age"`
	Status   int    `gorm:"column:status;default:1"` // 1: 活跃, 0: 非活跃
}

// TableName 将 User 使用的表名覆盖为 `users`
func (u *User) TableName() string {
	return "users"
}
