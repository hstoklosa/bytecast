package models

import (
	"time"

	"gorm.io/gorm"
)

// YouTubeVideo represents a video from a YouTube channel
type YouTubeVideo struct {
	gorm.Model
	YoutubeID    string    `gorm:"uniqueIndex;size:255;not null" json:"youtube_id"`  // YouTube video ID
	ChannelID    uint      `json:"channel_id"`                                       // Reference to the associated Channel
	Channel      Channel   `gorm:"foreignKey:ChannelID" json:"channel"`              // The channel that posted this video
	Title        string    `gorm:"size:255;not null" json:"title"`                   // Video title
	Description  string    `gorm:"type:text" json:"description"`                     // Video description
	ThumbnailURL string    `gorm:"size:512" json:"thumbnail_url"`                    // URL to video thumbnail
	Duration     string    `gorm:"size:32" json:"duration"`                          // Video duration in ISO 8601 format (e.g., "PT1H2M3S")
	PublishedAt  time.Time `json:"published_at"`                                     // When the video was published
	Watchlists   []*Watchlist `gorm:"many2many:watchlist_videos;" json:"watchlists"` // Watchlists containing this video
}

// TableName specifies the table name for the YouTubeVideo model
func (YouTubeVideo) TableName() string {
	return "youtube_videos"
}

// BeforeCreate is a GORM hook that sets timestamps
func (v *YouTubeVideo) BeforeCreate(tx *gorm.DB) error {
	if v.PublishedAt.IsZero() {
		v.PublishedAt = time.Now()
	}
	return nil
} 