package models

import (
	"gorm.io/gorm"
)

// Channel represents a YouTube channel that can be added to watchlists
type Channel struct {
	gorm.Model
	YouTubeID    string `gorm:"uniqueIndex;size:255;not null"` // YouTube channel ID
	Title        string `gorm:"size:255;not null"`             // Channel title
	Description  string `gorm:"type:text"`                     // Channel description
	ThumbnailURL string `gorm:"size:512"`                      // URL to channel thumbnail
	CustomName   string `gorm:"size:255"`                      // User-defined alias (optional)
	Watchlists   []*Watchlist `gorm:"many2many:watchlist_channels;"`
}

// TableName specifies the table name for the Channel model
func (Channel) TableName() string {
	return "channels"
}
