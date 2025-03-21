package models

import (
	"time"

	"gorm.io/gorm"
)

// YouTubeSubscription represents a subscription to a YouTube channel
type YouTubeSubscription struct {
	gorm.Model
	ChannelID      string    `gorm:"uniqueIndex;not null" json:"channel_id"`
	SubscribedAt   time.Time `json:"subscribed_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	IsActive       bool      `json:"is_active"`
	LeaseSeconds   int       `json:"lease_seconds"`
	Secret         string    `json:"-"`
}

// BeforeCreate is a GORM hook that sets the SubscribedAt timestamp
func (y *YouTubeSubscription) BeforeCreate(tx *gorm.DB) error {
	y.SubscribedAt = time.Now()
	y.IsActive = true
	return nil
}

// TableName specifies the table name for the YouTubeSubscription model
func (YouTubeSubscription) TableName() string {
	return "you_tube_subscriptions"
} 