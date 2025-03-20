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

// UserYouTubeSubscription represents the many-to-many relationship between users and YouTube channels
type UserYouTubeSubscription struct {
	gorm.Model
	UserID                uint                `json:"user_id"`
	YouTubeSubscriptionID uint                `json:"youtube_subscription_id"`
	YouTubeSubscription   YouTubeSubscription `gorm:"foreignKey:YouTubeSubscriptionID" json:"youtube_subscription"`
	AddedAt               time.Time           `json:"added_at"`
}

// TableName specifies the table name for UserYouTubeSubscription
func (UserYouTubeSubscription) TableName() string {
	return "user_youtube_subscriptions"
}

// BeforeCreate is a GORM hook that sets the AddedAt timestamp
func (u *UserYouTubeSubscription) BeforeCreate(tx *gorm.DB) error {
	u.AddedAt = time.Now()
	return nil
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