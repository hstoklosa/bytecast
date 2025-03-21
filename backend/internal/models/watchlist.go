package models

import (
	"fmt"
	"regexp"

	"gorm.io/gorm"
)

/*
 * Watchlist represents a collection of YouTube channels created by a user
 */
type Watchlist struct {
	gorm.Model
	UserID      uint   `gorm:"index;not null"` // Foreign key to User
	Name        string `gorm:"size:255;not null"`
	Description string `gorm:"type:text"`
	Color       string `gorm:"size:7;not null;check:color ~ '^#[a-fA-F0-9]{6}$'"`
	Channels    []*Channel `gorm:"many2many:watchlist_channels;"`
	Videos      []*Video `gorm:"many2many:watchlist_videos;"`
}

func (Watchlist) TableName() string {
	return "watchlists"
}

// BeforeSave Hook to validate the watchlist before saving
func (w *Watchlist) BeforeSave(tx *gorm.DB) error {
	if err := w.ValidateColor(); err != nil {
		return err
	}
	return nil
}

// ValidateColor checks if the color is a valid hex code
func (w *Watchlist) ValidateColor() error {
	if w.Color == "" {
		return fmt.Errorf("color is required")
	}
	
	matched, err := regexp.MatchString(`^#[a-fA-F0-9]{6}$`, w.Color)
	if err != nil {
		return fmt.Errorf("invalid color format")
	}
	if !matched {
		return fmt.Errorf("color must be a valid 6-digit hex code (e.g. #FF0000)")
	}
	
	return nil
}