package models

import (
	"gorm.io/gorm"
)

type User struct {
gorm.Model
Email        string `gorm:"uniqueIndex;not null" json:"email"`
Username     string `gorm:"uniqueIndex;not null;size:24" json:"username"`
PasswordHash string `gorm:"not null" json:"-"` // "-" omits from JSON responses
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "users"
}
