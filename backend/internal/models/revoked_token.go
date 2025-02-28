package models

import (
	"time"
)

// RevokedToken represents a blacklisted refresh token
type RevokedToken struct {
	ID        uint      `gorm:"primaryKey"`
	TokenHash string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"index;not null"`
	UserID    uint      `gorm:"index;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
