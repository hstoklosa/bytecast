package services

import (
	"errors"
	"fmt"
	"time"

	"bytecast/internal/models"

	"gorm.io/gorm"
)

// VideoService handles operations related to YouTube videos
type VideoService struct {
	db *gorm.DB
}

// NewVideoService creates a new YouTube video service
func NewVideoService(db *gorm.DB) *VideoService {
	return &VideoService{
		db: db,
	}
}

// CreateVideo creates a new YouTube video in the database
func (s *VideoService) CreateVideo(video *models.Video) error {
	// Start a transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if video already exists
	var existingVideo models.Video
	if err := tx.Where("youtube_id = ?", video.YoutubeID).First(&existingVideo).Error; err == nil {
		// Video already exists, update it
		existingVideo.Title = video.Title
		existingVideo.Description = video.Description
		existingVideo.ThumbnailURL = video.ThumbnailURL
		existingVideo.Duration = video.Duration
		
		// Only update PublishedAt if the new value is not zero
		if !video.PublishedAt.IsZero() {
			existingVideo.PublishedAt = video.PublishedAt
		}
		
		if err := tx.Save(&existingVideo).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update video: %w", err)
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new video
		if err := tx.Create(video).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create video: %w", err)
		}
	} else {
		// Database error
		tx.Rollback()
		return fmt.Errorf("failed to check for existing video: %w", err)
	}

	return tx.Commit().Error
}

// GetVideoByID retrieves a YouTube video by its ID
func (s *VideoService) GetVideoByID(videoID string) (*models.Video, error) {
	var video models.Video
	if err := s.db.Where("youtube_id = ?", videoID).First(&video).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("video not found")
		}
		return nil, fmt.Errorf("failed to get video: %w", err)
	}

	return &video, nil
}

// GetVideosByChannelID retrieves all videos from a specific channel
func (s *VideoService) GetVideosByChannelID(channelID uint) ([]models.Video, error) {
	var videos []models.Video
	if err := s.db.Where("channel_id = ?", channelID).Find(&videos).Error; err != nil {
		return nil, fmt.Errorf("failed to get videos for channel: %w", err)
	}

	return videos, nil
}

// GetRecentVideos retrieves videos published within a specified time period
func (s *VideoService) GetRecentVideos(since time.Time) ([]models.Video, error) {
	var videos []models.Video
	if err := s.db.Where("published_at > ?", since).Order("published_at DESC").Find(&videos).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent videos: %w", err)
	}

	return videos, nil
}

// UpdateVideo updates an existing YouTube video
func (s *VideoService) UpdateVideo(video *models.Video) error {
	var existingVideo models.Video
	if err := s.db.Where("youtube_id = ?", video.YoutubeID).First(&existingVideo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("video not found")
		}
		return fmt.Errorf("failed to get video: %w", err)
	}

	// Update existing video with new information
	if err := s.db.Model(&existingVideo).Updates(video).Error; err != nil {
		return fmt.Errorf("failed to update video: %w", err)
	}

	return nil
}

// DeleteVideo removes a YouTube video from the database
func (s *VideoService) DeleteVideo(videoID string) error {
	if err := s.db.Where("youtube_id = ?", videoID).Delete(&models.Video{}).Error; err != nil {
		return fmt.Errorf("failed to delete video: %w", err)
	}

	return nil
}

// AddVideoToWatchlist adds a video to a watchlist
func (s *VideoService) AddVideoToWatchlist(videoID string, watchlistID uint) error {
	// Start a transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get the video
	var video models.Video
	if err := tx.Where("youtube_id = ?", videoID).First(&video).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("video not found: %w", err)
	}

	// Get the watchlist
	var watchlist models.Watchlist
	if err := tx.First(&watchlist, watchlistID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("watchlist not found: %w", err)
	}

	// Check if the video is already in the watchlist
	var count int64
	if err := tx.Table("watchlist_videos").
		Where("video_id = ? AND watchlist_id = ?", video.ID, watchlistID).
		Count(&count).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to check watchlist: %w", err)
	}

	if count == 0 {
		// Add the video to the watchlist using direct SQL
		if err := tx.Exec("INSERT INTO watchlist_videos (watchlist_id, video_id) VALUES (?, ?)", 
			watchlistID, video.ID).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to add video to watchlist: %w", err)
		}
	}

	return tx.Commit().Error
}

// RemoveVideoFromWatchlist removes a video from a watchlist
func (s *VideoService) RemoveVideoFromWatchlist(videoID string, watchlistID uint) error {
	// Start a transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get the video
	var video models.Video
	if err := tx.Where("youtube_id = ?", videoID).First(&video).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("video not found: %w", err)
	}

	// Get the watchlist
	var watchlist models.Watchlist
	if err := tx.First(&watchlist, watchlistID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("watchlist not found: %w", err)
	}

	// Remove the video from the watchlist using direct SQL for efficiency
	if err := tx.Exec("DELETE FROM watchlist_videos WHERE watchlist_id = ? AND video_id = ?",
		watchlistID, video.ID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove video from watchlist: %w", err)
	}

	return tx.Commit().Error
}

// GetVideosByWatchlistID retrieves all videos in a specific watchlist
func (s *VideoService) GetVideosByWatchlistID(watchlistID uint) ([]models.Video, error) {
	var videos []models.Video
	if err := s.db.Joins("JOIN watchlist_videos ON watchlist_videos.video_id = youtube_videos.id").
		Where("watchlist_videos.watchlist_id = ?", watchlistID).
		Find(&videos).Error; err != nil {
		return nil, fmt.Errorf("failed to get videos for watchlist: %w", err)
	}

	return videos, nil
} 