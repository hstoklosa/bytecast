package services

import (
	"errors"
	"fmt"
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

type YouTubeServiceInterface interface {
	GetChannelInfo(channelID string) (*ChannelInfo, error)
}

type PubSubServiceInterface interface {
	SubscribeToChannel(channelID string) error
	UnsubscribeFromChannel(channelID string) error
}

type WatchlistService struct {
	db             *gorm.DB
	config         *configs.Config
	youtubeService YouTubeServiceInterface
	pubsubService  PubSubServiceInterface
}

func NewWatchlistService(db *gorm.DB, config *configs.Config, youtubeService YouTubeServiceInterface) *WatchlistService {
	return &WatchlistService{
		db:             db,
		config:         config,
		youtubeService: youtubeService,
		pubsubService:  nil, // later via SetPubSubService if available
	}
}

func (s *WatchlistService) CreateDefaultWatchlist(userID uint) error {
	watchlist := models.Watchlist{
		UserID:      userID,
		Name:        "Default",
		Description: "Your default watchlist",
		Color:       "#3b82f6",
	}
	return s.db.Create(&watchlist).Error
}

func (s *WatchlistService) CreateWatchlist(userID uint, name, description, color string) (*models.Watchlist, error) {
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

func (s *WatchlistService) GetUserWatchlists(userID uint) ([]models.Watchlist, error) {
	var watchlists []models.Watchlist
	if err := s.db.Where("user_id = ?", userID).Find(&watchlists).Error; err != nil {
		return nil, err
	}

	return watchlists, nil
}

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

func (s *WatchlistService) AddChannelToWatchlist(watchlistID, userID uint, channelID string) error {
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Verify watchlist belongs to user
	var watchlist models.Watchlist
	if err := tx.Where("id = ? AND user_id = ?", watchlistID, userID).First(&watchlist).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrWatchlistNotFound
		}
		return err
	}

	if s.youtubeService == nil {
		tx.Rollback()
		return ErrMissingAPIKey
	}

	channelInfo, err := s.youtubeService.GetChannelInfo(channelID)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, ErrInvalidYouTubeURL) || errors.Is(err, ErrChannelNotFoundAPI) {
			return ErrInvalidYouTubeID
		}
		return err
	}

	var channel models.Channel
	err = tx.Where("youtube_id = ?", channelInfo.ID).First(&channel).Error
	
	needsSubscription := false
	if errors.Is(err, gorm.ErrRecordNotFound) {
		var softDeletedChannel models.Channel
		err = tx.Unscoped().Where("youtube_id = ?", channelInfo.ID).First(&softDeletedChannel).Error
		
		if err == nil {
			softDeletedChannel.DeletedAt = gorm.DeletedAt{}
			softDeletedChannel.Title = channelInfo.Title
			softDeletedChannel.Description = channelInfo.Description
			softDeletedChannel.ThumbnailURL = channelInfo.Thumbnail
			
			if err := tx.Unscoped().Save(&softDeletedChannel).Error; err != nil {
				tx.Rollback()
				return err
			}
			
			channel = softDeletedChannel
		} else {
			channel = models.Channel{
				YoutubeID:     channelInfo.ID,
				Title:         channelInfo.Title,
				Description:   channelInfo.Description,
				ThumbnailURL:  channelInfo.Thumbnail,
			}
			if err := tx.Create(&channel).Error; err != nil {
				tx.Rollback()
				return err
			}
			
			needsSubscription = true
		}
	} else if err != nil {
		tx.Rollback()
		return err
	} else {
		channel.Title = channelInfo.Title
		channel.Description = channelInfo.Description
		channel.ThumbnailURL = channelInfo.Thumbnail
		if err := tx.Save(&channel).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	var exists int64
	err = tx.Model(&models.Watchlist{}).
		Joins("JOIN watchlist_channels ON watchlist_channels.watchlist_id = watchlists.id").
		Where("watchlists.id = ? AND watchlist_channels.channel_id = ?", watchlistID, channel.ID).
		Count(&exists).Error
	
	if err != nil {
		tx.Rollback()
		return err
	}
	
	if exists == 0 {
		if err := tx.Exec("INSERT INTO watchlist_channels (watchlist_id, channel_id) VALUES (?, ?)", 
			watchlistID, channel.ID).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	
	if needsSubscription && s.pubsubService != nil {
		if err := s.pubsubService.SubscribeToChannel(channelInfo.ID); err != nil {
			log.Printf("Warning: Failed to subscribe to YouTube PubSubHubbub for channel %s: %v", channelInfo.ID, err)
		} else {
			log.Printf("Successfully subscribed to YouTube PubSubHubbub for channel %s", channelInfo.ID)
		}
	}

	return nil
}

func (s *WatchlistService) RemoveChannelFromWatchlist(watchlistID, userID uint, channelID string) error {
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var watchlist models.Watchlist
	if err := tx.Where("id = ? AND user_id = ?", watchlistID, userID).First(&watchlist).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrWatchlistNotFound
		}
		return err
	}

	extractedID := channelID
	if s.youtubeService != nil && (strings.Contains(channelID, "/") || strings.Contains(channelID, "@")) {
		channelInfo, err := s.youtubeService.GetChannelInfo(channelID)
		if err == nil && channelInfo != nil {
			extractedID = channelInfo.ID
		}
	}

	var channel models.Channel
	if err := tx.Where("youtube_id = ?", extractedID).First(&channel).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrChannelNotFound
		}
		return err
	}

	if err := tx.Exec(`
		DELETE FROM watchlist_videos
		WHERE watchlist_id = ? AND video_id IN (
			SELECT id FROM youtube_videos WHERE channel_id = ?
		)
	`, watchlistID, channel.ID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove videos from watchlist: %w", err)
	}

	if err := tx.Exec("DELETE FROM watchlist_channels WHERE watchlist_id = ? AND channel_id = ?", 
		watchlistID, channel.ID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove channel from watchlist: %w", err)
	}
	
	var usageCount int64
	if err := tx.Model(&models.Watchlist{}).
		Joins("JOIN watchlist_channels ON watchlist_channels.watchlist_id = watchlists.id").
		Where("watchlist_channels.channel_id = ?", channel.ID).
		Count(&usageCount).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to check channel usage: %w", err)
	}
	
	needsCleanup := usageCount == 0
	
	if err := tx.Commit().Error; err != nil {
		return err
	}
	
	if needsCleanup {
		if s.pubsubService != nil {
			if err := s.pubsubService.UnsubscribeFromChannel(extractedID); err != nil {
				log.Printf("Warning: Failed to unsubscribe from YouTube PubSubHubbub for channel %s: %v", extractedID, err)
			}
		}
		
		cleanupTx := s.db.Begin()
		if cleanupTx.Error != nil {
			log.Printf("Warning: Failed to start cleanup transaction: %v", cleanupTx.Error)
			return nil
		}
		
		// Soft-delete videos related to this channel
		if err := cleanupTx.Where("channel_id = ?", channel.ID).Delete(&models.Video{}).Error; err != nil {
			cleanupTx.Rollback()
			log.Printf("Warning: Failed to delete channel videos: %v", err)
			return nil
		}
		
		if err := cleanupTx.Delete(&channel).Error; err != nil {
			cleanupTx.Rollback()
			log.Printf("Warning: Failed to delete channel: %v", err)
			return nil
		}
		
		if err := cleanupTx.Commit().Error; err != nil {
			log.Printf("Warning: Failed to commit cleanup transaction: %v", err)
		}
	}
	
	return nil
}

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

// SetYouTubeService sets the YouTube service (used for testing)
func (s *WatchlistService) SetYouTubeService(youtubeService YouTubeServiceInterface) {
	s.youtubeService = youtubeService
}

func (s *WatchlistService) SetPubSubService(pubsubService PubSubServiceInterface) {
	s.pubsubService = pubsubService
}
