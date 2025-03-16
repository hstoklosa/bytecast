package services

import (
	"errors"
	"log"
	"strings"

	"gorm.io/gorm"

	"bytecast/configs"
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
	db             *gorm.DB
	youtubeService YouTubeServiceInterface
	config         *configs.Config
}

// NewWatchlistService creates a new watchlist service
func NewWatchlistService(db *gorm.DB, config *configs.Config) *WatchlistService {
	var youtubeService YouTubeServiceInterface
	
	// Only try to initialize YouTube service if API key is provided
	if config != nil && config.YouTube.APIKey != "" {
		var err error
		youtubeService, err = NewYouTubeService(config)
		if err != nil {
			// Log the error but continue without YouTube service
			log.Printf("Warning: YouTube service initialization failed: %v", err)
		}
	} else {
		log.Printf("YouTube API key not provided, YouTube features will be disabled")
	}
	
	return &WatchlistService{
		db:             db,
		youtubeService: youtubeService,
		config:         config,
	}
}

// CreateDefaultWatchlist creates a default watchlist for a newly registered user
func (s *WatchlistService) CreateDefaultWatchlist(userID uint) error {
	watchlist := models.Watchlist{
		UserID:      userID,
		Name:        "Default",
		Description: "Your default watchlist",
		Color:       "#3b82f6", // Default blue color
	}
	return s.db.Create(&watchlist).Error
}

// CreateWatchlist creates a new watchlist for a user
func (s *WatchlistService) CreateWatchlist(userID uint, name, description, color string) (*models.Watchlist, error) {
	// Validate userID
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}
	
	watchlist := models.Watchlist{
		UserID:      userID,
		Name:        name,
		Description: description,
		Color:       color,
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

// UpdateWatchlist updates a watchlist's name, description, and color
func (s *WatchlistService) UpdateWatchlist(watchlistID, userID uint, name, description, color string) (*models.Watchlist, error) {
	watchlist, err := s.GetWatchlist(watchlistID, userID)
	if err != nil {
		return nil, err
	}

	watchlist.Name = name
	watchlist.Description = description
	watchlist.Color = color

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

	// Check if YouTube service is available
	if s.youtubeService == nil {
		return ErrMissingAPIKey
	}

	// Use YouTube API to get channel info
	channelInfo, err := s.youtubeService.GetChannelInfo(channelID)
	if err != nil {
		if errors.Is(err, ErrInvalidYouTubeURL) || errors.Is(err, ErrChannelNotFoundAPI) {
			return ErrInvalidYouTubeID
		}
		return err
	}

	// Check if channel already exists in database
	var channel models.Channel
	err = s.db.Where("you_tube_id = ?", channelInfo.ID).First(&channel).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new channel with data from YouTube API
		channel = models.Channel{
			YouTubeID:     channelInfo.ID,
			Title:         channelInfo.Title,
			Description:   channelInfo.Description,
			ThumbnailURL:  channelInfo.Thumbnail,
		}
		if err := s.db.Create(&channel).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		// Update existing channel with latest data from YouTube
		channel.Title = channelInfo.Title
		channel.Description = channelInfo.Description
		channel.ThumbnailURL = channelInfo.Thumbnail
		if err := s.db.Save(&channel).Error; err != nil {
			return err
		}
	}

	// Check if channel is already in the watchlist
	var exists int64
	err = s.db.Model(&models.Watchlist{}).
		Joins("JOIN watchlist_channels ON watchlist_channels.watchlist_id = watchlists.id").
		Joins("JOIN channels ON watchlist_channels.channel_id = channels.id").
		Where("watchlists.id = ? AND channels.you_tube_id = ?", watchlistID, channelInfo.ID).
		Count(&exists).Error
	
	if err != nil {
		return err
	}
	
	if exists > 0 {
		// Channel is already in the watchlist, no need to add it again
		return nil
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

	// Try to extract channel ID if it's a URL and YouTube service is available
	extractedID := channelID
	if s.youtubeService != nil && (strings.Contains(channelID, "/") || strings.Contains(channelID, "@")) {
		// Try to get channel info to extract the proper ID
		channelInfo, err := s.youtubeService.GetChannelInfo(channelID)
		if err == nil && channelInfo != nil {
			extractedID = channelInfo.ID
		}
	}

	// Find channel
	var channel models.Channel
	if err := s.db.Where("you_tube_id = ?", extractedID).First(&channel).Error; err != nil {
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

// YouTubeServiceInterface defines the interface for YouTube service
type YouTubeServiceInterface interface {
	GetChannelInfo(channelID string) (*ChannelInfo, error)
}

// SetYouTubeService sets the YouTube service (used for testing)
func (s *WatchlistService) SetYouTubeService(youtubeService YouTubeServiceInterface) {
	s.youtubeService = youtubeService
}
