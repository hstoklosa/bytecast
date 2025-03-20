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
func (s *VideoService) CreateVideo(video *models.YouTubeVideo) error {
	// Check if video already exists
	var existingVideo models.YouTubeVideo
	if err := s.db.Where("youtube_id = ?", video.YouTubeID).First(&existingVideo).Error; err == nil {
		// Video already exists, update it instead
		return s.UpdateVideo(video)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Database error
		return fmt.Errorf("failed to check for existing video: %w", err)
	}

	// Create new video
	if err := s.db.Create(video).Error; err != nil {
		return fmt.Errorf("failed to create video: %w", err)
	}

	return nil
}

// GetVideoByID retrieves a YouTube video by its ID
func (s *VideoService) GetVideoByID(videoID string) (*models.YouTubeVideo, error) {
	var video models.YouTubeVideo
	if err := s.db.Where("youtube_id = ?", videoID).First(&video).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("video not found")
		}
		return nil, fmt.Errorf("failed to get video: %w", err)
	}

	return &video, nil
}

// GetVideosByChannelID retrieves all videos from a specific channel
func (s *VideoService) GetVideosByChannelID(channelID uint) ([]models.YouTubeVideo, error) {
	var videos []models.YouTubeVideo
	if err := s.db.Where("channel_id = ?", channelID).Find(&videos).Error; err != nil {
		return nil, fmt.Errorf("failed to get videos for channel: %w", err)
	}

	return videos, nil
}

// GetRecentVideos retrieves videos published within a specified time period
func (s *VideoService) GetRecentVideos(since time.Time) ([]models.YouTubeVideo, error) {
	var videos []models.YouTubeVideo
	if err := s.db.Where("published_at > ?", since).Order("published_at DESC").Find(&videos).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent videos: %w", err)
	}

	return videos, nil
}

// UpdateVideo updates an existing YouTube video
func (s *VideoService) UpdateVideo(video *models.YouTubeVideo) error {
	var existingVideo models.YouTubeVideo
	if err := s.db.Where("youtube_id = ?", video.YouTubeID).First(&existingVideo).Error; err != nil {
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
	if err := s.db.Where("youtube_id = ?", videoID).Delete(&models.YouTubeVideo{}).Error; err != nil {
		return fmt.Errorf("failed to delete video: %w", err)
	}

	return nil
}

// AddVideoToWatchlist adds a video to a watchlist
func (s *VideoService) AddVideoToWatchlist(videoID string, watchlistID uint) error {
	// Get the video
	var video models.YouTubeVideo
	if err := s.db.Where("youtube_id = ?", videoID).First(&video).Error; err != nil {
		return fmt.Errorf("video not found: %w", err)
	}

	// Get the watchlist
	var watchlist models.Watchlist
	if err := s.db.First(&watchlist, watchlistID).Error; err != nil {
		return fmt.Errorf("watchlist not found: %w", err)
	}

	// Check if the video is already in the watchlist
	var count int64
	if err := s.db.Table("watchlist_videos").
		Where("youtube_video_id = ? AND watchlist_id = ?", video.ID, watchlistID).
		Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check watchlist: %w", err)
	}

	if count > 0 {
		// Video is already in the watchlist
		return nil
	}

	// Add the video to the watchlist using the many-to-many relationship
	if err := s.db.Model(&watchlist).Association("Videos").Append(&video); err != nil {
		return fmt.Errorf("failed to add video to watchlist: %w", err)
	}

	return nil
}

// RemoveVideoFromWatchlist removes a video from a watchlist
func (s *VideoService) RemoveVideoFromWatchlist(videoID string, watchlistID uint) error {
	// Get the video
	var video models.YouTubeVideo
	if err := s.db.Where("youtube_id = ?", videoID).First(&video).Error; err != nil {
		return fmt.Errorf("video not found: %w", err)
	}

	// Get the watchlist
	var watchlist models.Watchlist
	if err := s.db.First(&watchlist, watchlistID).Error; err != nil {
		return fmt.Errorf("watchlist not found: %w", err)
	}

	// Remove the video from the watchlist
	if err := s.db.Model(&watchlist).Association("Videos").Delete(&video); err != nil {
		return fmt.Errorf("failed to remove video from watchlist: %w", err)
	}

	return nil
}

// GetVideosByWatchlistID retrieves all videos in a specific watchlist
func (s *VideoService) GetVideosByWatchlistID(watchlistID uint) ([]models.YouTubeVideo, error) {
	var videos []models.YouTubeVideo
	if err := s.db.Joins("JOIN watchlist_videos ON watchlist_videos.youtube_video_id = youtube_videos.id").
		Where("watchlist_videos.watchlist_id = ?", watchlistID).
		Find(&videos).Error; err != nil {
		return nil, fmt.Errorf("failed to get videos for watchlist: %w", err)
	}

	return videos, nil
} 