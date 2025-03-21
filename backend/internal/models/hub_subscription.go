package models

import (
	"time"

	"gorm.io/gorm"
)

/*
 * HubSubscription represents a subscription to a YouTube's WebSub hub to
 * receive notifications when a new video is uploaded to a channel.
 */
type HubSubscription struct {
	gorm.Model
	ChannelID      string    `gorm:"uniqueIndex;not null" json:"channel_id"`
	SubscribedAt   time.Time `json:"subscribed_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	IsActive       bool      `json:"is_active"`
	LeaseSeconds   int       `json:"lease_seconds"`
	Secret         string    `json:"-"`
}

func (y *HubSubscription) BeforeCreate(tx *gorm.DB) error {
	y.SubscribedAt = time.Now()
	y.IsActive = true
	return nil
}

func (HubSubscription) TableName() string {
	return "hub_subscriptions"
} 