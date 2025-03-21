package models

import (
	"time"

	"gorm.io/gorm"
)

/*
 * Video represents a video from a YouTube channel.
 */
type Video struct {
	gorm.Model
	YoutubeID    string `gorm:"uniqueIndex;size:255;not null" json:"youtube_id"`
	ChannelID    uint `json:"channel_id"`
	Channel      Channel `gorm:"foreignKey:ChannelID" json:"channel"` 
	Title        string `gorm:"size:255;not null" json:"title"`
	Description  string `gorm:"type:text" json:"description"`
	ThumbnailURL string `gorm:"size:512" json:"thumbnail_url"`
	Duration     string `gorm:"size:32" json:"duration"` // in ISO 8601 format (e.g., "PT1H2M3S")
	PublishedAt  time.Time `json:"published_at"`
	Watchlists   []*Watchlist `gorm:"many2many:watchlist_videos;" json:"watchlists"`
}

func (Video) TableName() string {
	return "youtube_videos"
}

func (v *Video) BeforeCreate(tx *gorm.DB) error {
	if v.PublishedAt.IsZero() {
		v.PublishedAt = time.Now()
	}
	return nil
} 