package services

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"bytecast/internal/models"
)

var (
	ErrWatchlistNotFound = errors.New("watchlist not found")
	ErrChannelNotFound   = errors.New("channel not found")
	ErrNotAuthorized     = errors.New("not authorized to access this watchlist")
	ErrInvalidYouTubeID  = errors.New("invalid YouTube channel ID or URL")
)

// WatchlistService handles operations related to watchlists and channels
type WatchlistService struct {
	db *gorm.DB
}

// NewWatchlistService creates a new watchlist service
func NewWatchlistService(db *gorm.DB) *WatchlistService {
	return &WatchlistService{
		db: db,
	}
}

// CreateWatchlist creates a new watchlist for a user
func (s *WatchlistService) CreateWatchlist(userID uint, name, description string) (*models.Watchlist, error) {
	watchlist := models.Watchlist{
		UserID:      userID,
		Name:        name,
		Description: description,
	}

	if err := s.db.Create(&watchlist).Error; err != nil {
		return nil, err
	}

	return &watchlist, nil
}

// GetWatchlist retrieves a watchlist by ID, ensuring it belongs to the specified user
func (s *WatchlistService) GetWatchlist(watchlistID, userID uint) (*models.Watchlist, error) {
	var watchlist models.Watchlist
	if err := s.db.Preload("Channels").Where("id = ? AND user_id = ?", watchlistID, userID).First(&watchlist).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWatchlistNotFound
		}
		return nil, err
	}

	return &watchlist, nil
}

// GetUserWatchlists retrieves all watchlists for a user
func (s *WatchlistService) GetUserWatchlists(userID uint) ([]models.Watchlist, error) {
	var watchlists []models.Watchlist
	if err := s.db.Where("user_id = ?", userID).Find(&watchlists).Error; err != nil {
		return nil, err
	}

	return watchlists, nil
}

// UpdateWatchlist updates a watchlist's name and description
func (s *WatchlistService) UpdateWatchlist(watchlistID, userID uint, name, description string) (*models.Watchlist, error) {
	watchlist, err := s.GetWatchlist(watchlistID, userID)
	if err != nil {
		return nil, err
	}

	watchlist.Name = name
	watchlist.Description = description

	if err := s.db.Save(watchlist).Error; err != nil {
		return nil, err
	}

	return watchlist, nil
}

// DeleteWatchlist deletes a watchlist
func (s *WatchlistService) DeleteWatchlist(watchlistID, userID uint) error {
	result := s.db.Where("id = ? AND user_id = ?", watchlistID, userID).Delete(&models.Watchlist{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrWatchlistNotFound
	}

	return nil
}

// AddChannelToWatchlist adds a channel to a watchlist
func (s *WatchlistService) AddChannelToWatchlist(watchlistID, userID uint, channelID string) error {
	// Verify watchlist belongs to user
	watchlist, err := s.GetWatchlist(watchlistID, userID)
	if err != nil {
		return err
	}

	// Extract YouTube channel ID from URL if needed
	youtubeID := extractYouTubeChannelID(channelID)
	if youtubeID == "" {
		return ErrInvalidYouTubeID
	}

	// Check if channel exists, create if not
	var channel models.Channel
	err = s.db.Where("you_tube_id = ?", youtubeID).First(&channel).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// In a real implementation, we would fetch channel details from YouTube API
		// For now, we'll create a placeholder
		channel = models.Channel{
			YouTubeID: youtubeID,
			Title:     "Channel " + youtubeID, // Placeholder, would be fetched from YouTube API
		}
		if err := s.db.Create(&channel).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// Add channel to watchlist
	return s.db.Model(watchlist).Association("Channels").Append(&channel)
}

// RemoveChannelFromWatchlist removes a channel from a watchlist
func (s *WatchlistService) RemoveChannelFromWatchlist(watchlistID, userID uint, channelID string) error {
	// Verify watchlist belongs to user
	watchlist, err := s.GetWatchlist(watchlistID, userID)
	if err != nil {
		return err
	}

	// Find channel
	var channel models.Channel
	if err := s.db.Where("you_tube_id = ?", channelID).First(&channel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrChannelNotFound
		}
		return err
	}

	// Remove channel from watchlist
	return s.db.Model(watchlist).Association("Channels").Delete(&channel)
}

// GetChannelsInWatchlist retrieves all channels in a watchlist
func (s *WatchlistService) GetChannelsInWatchlist(watchlistID, userID uint) ([]models.Channel, error) {
	watchlist, err := s.GetWatchlist(watchlistID, userID)
	if err != nil {
		return nil, err
	}

	var channels []models.Channel
	if err := s.db.Model(watchlist).Association("Channels").Find(&channels); err != nil {
		return nil, err
	}

	return channels, nil
}

// Helper function to extract YouTube channel ID from various formats
func extractYouTubeChannelID(input string) string {
	// If it's already just an ID, return it
	if !strings.Contains(input, "/") && !strings.Contains(input, ".") {
		return input
	}

	// Handle URLs like https://www.youtube.com/channel/UC_x5XG1OV2P6uZZ5FSM9Ttw
	if strings.Contains(input, "/channel/") {
		parts := strings.Split(input, "/channel/")
		if len(parts) > 1 {
			id := parts[1]
			// Remove any query parameters or hash
			id = strings.Split(id, "?")[0]
			id = strings.Split(id, "#")[0]
			// Remove trailing slash if present
			id = strings.TrimSuffix(id, "/")
			return id
		}
	}

	// In a real implementation, we would handle more URL formats
	// and potentially use the YouTube API to resolve custom URLs

	return ""
}
