package models

import (
	"gorm.io/gorm"
)

// Watchlist represents a collection of YouTube channels created by a user
type Watchlist struct {
	gorm.Model
	UserID      uint   `gorm:"index;not null"` // Foreign key to User
	Name        string `gorm:"size:255;not null"`
	Description string `gorm:"type:text"`
	Channels    []*Channel `gorm:"many2many:watchlist_channels;"`
}

// TableName specifies the table name for the Watchlist model
func (Watchlist) TableName() string {
	return "watchlists"
}
