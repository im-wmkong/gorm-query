package model

import "gorm.io/gorm"

// Profile defines the demo profile model.
//
// Association:
//   - User has one Profile (User.Profile)
//   - Profile belongs to User via UserID
type Profile struct {
	gorm.Model
	UserID uint   `gorm:"column:user_id;not null;index"`
	Bio    string `gorm:"column:bio;size:1024"`
}

// TableName overrides the default table name to `profiles`.
func (p *Profile) TableName() string {
	return "profiles"
}
