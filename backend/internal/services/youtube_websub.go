package services

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"bytecast/configs"
	"bytecast/internal/models"

	"gorm.io/gorm"
)

const (
	hubURL  = "https://pubsubhubbub.appspot.com/subscribe"
	feedURL = "https://www.youtube.com/xml/feeds/videos.xml?channel_id=%s"
)

var ErrPubSubHubError = errors.New("error communicating with PubSubHubbub hub")

type PubSubService struct {
	db             *gorm.DB
	config         *configs.Config
	client         *http.Client
	videoService   *VideoService
	youtubeService *YouTubeService
}

func NewPubSubService(db *gorm.DB, config *configs.Config, videoService *VideoService) *PubSubService {
	youtubeService, err := NewYouTubeService(config)
	if err != nil {
		log.Printf("Failed to create YouTube service for PubSub: %v", err)
		return nil
	}

	return &PubSubService{
		db:             db,
		config:         config,
		client:         &http.Client{Timeout: 10 * time.Second},
		videoService:   videoService,
		youtubeService: youtubeService,
	}
}

func (s *PubSubService) SubscribeToChannel(channelID string) error {
	var existingSub models.HubSubscription
	err := s.db.Where("channel_id = ?", channelID).First(&existingSub).Error
	
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("database error checking for existing subscription: %w", err)
	}
	
	leaseSeconds := s.config.YouTube.LeaseSeconds
	expiresAt := time.Time{} // 0 = permanent subscription
	
	if leaseSeconds > 0 {
		expiresAt = time.Now().Add(time.Duration(leaseSeconds) * time.Second)
	}
	
	if errors.Is(err, gorm.ErrRecordNotFound) {
		subscription := models.HubSubscription{
			ChannelID:     channelID,
			LeaseSeconds:  leaseSeconds,
			ExpiresAt:     expiresAt,
			Secret:        generateSecret(),
			IsActive:      true,
		}
		
		if err := s.db.Create(&subscription).Error; err != nil {
			return fmt.Errorf("failed to create subscription record: %w", err)
		}
		
		existingSub = subscription
	} else {
		existingSub.LeaseSeconds = leaseSeconds
		existingSub.ExpiresAt = expiresAt
		existingSub.IsActive = true
		
		if err := s.db.Save(&existingSub).Error; err != nil {
			return fmt.Errorf("failed to update subscription record: %w", err)
		}
	}

	if err := s.sendSubscriptionRequest(channelID, existingSub.Secret, "subscribe"); err != nil {
		log.Printf("Failed to subscribe to hub for channel %s: %v", channelID, err)
	}
	
	return nil
}

func (s *PubSubService) UnsubscribeFromChannel(channelID string) error {
	var subscription models.HubSubscription
	if err := s.db.Where("channel_id = ?", channelID).First(&subscription).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("No subscription record found for channel %s", channelID)
			return nil
		}
		
		return fmt.Errorf("database error when finding subscription: %w", err)
	}

	// Mark subscription as inactive
	subscription.IsActive = false
	if err := s.db.Save(&subscription).Error; err != nil {
		return fmt.Errorf("failed to mark subscription as inactive: %w", err)
	}
	
	// Send unsubscribe request to hub
	if err := s.sendSubscriptionRequest(channelID, subscription.Secret, "unsubscribe"); err != nil {
		log.Printf("Failed to unsubscribe from hub for channel %s: %v", channelID, err)
	}

	// Delete the subscription record
	// if err := s.db.Delete(&subscription).Error; err != nil {
	// 	return fmt.Errorf("failed to delete subscription record: %w", err)
	// }
	
	return nil
}

func (s *PubSubService) sendSubscriptionRequest(channelID, secret, mode string) error {
	feedURL := fmt.Sprintf(feedURL, channelID)
	callbackURL := s.config.YouTube.CallbackURL
	
	form := url.Values{}
	form.Set("hub.callback", callbackURL)
	form.Set("hub.mode", mode)
	form.Set("hub.topic", feedURL)
	form.Set("hub.secret", secret)
	form.Set("hub.verify", "async")
	
	if s.config.YouTube.LeaseSeconds > 0 {
		form.Set("hub.lease_seconds", fmt.Sprintf("%d", s.config.YouTube.LeaseSeconds))
	}

	resp, err := s.client.PostForm(hubURL, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%w: subscription failed with status %d: %s", ErrPubSubHubError, resp.StatusCode, string(body))
	}

	log.Printf("Successfully sent hub %s request for channel %s", mode, channelID)

	return nil
}

// generateSecret generates a random secret for HMAC verification 
// with a fallback method if crypto/rand fails
func generateSecret() string {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		for i := range secret {
			secret[i] = byte(i * 7)
		}
	}

	return base64.StdEncoding.EncodeToString(secret)
}

// The structs represent a YouTube video, watchlist, and channel XML structure
type Feed struct {
	XMLName xml.Name `xml:"feed"`
	Entries []Entry  `xml:"entry"`
}

type Entry struct {
	ID        string `xml:"id"`
	VideoID   string `xml:"videoId"`
	ChannelID string `xml:"channelId"`
	Title     string `xml:"title"`
	Link      Link   `xml:"link"`
	Author    Author `xml:"author"`
	Published string `xml:"published"`
	Updated   string `xml:"updated"`
}

type Link struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

type Author struct {
	Name string `xml:"name"`
	URI  string `xml:"uri"`
}

func (s *PubSubService) ProcessVideoNotification(body []byte, signature string) error {
	channelID, err := s.verifySignature(body, signature)
	if err != nil {
		return fmt.Errorf("invalid notification signature: %w", err)
	}

	log.Printf("X Signature verified for channel: %s", channelID)

	// Parse notification XML
	var feed Feed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return fmt.Errorf("failed to parse notification: %w", err)
	}

	// Process each entry
	for _, entry := range feed.Entries {
		if err := s.processEntry(entry); err != nil {
			log.Printf("X Error processing entry: %v", err)
		}
	}

	return nil
}

func (s *PubSubService) verifySignature(body []byte, signature string) (string, error) {
	actualSignature := signature
	if strings.HasPrefix(signature, "sha1=") {
		hexSignature := signature[5:]
		decodedBytes, err := hex.DecodeString(hexSignature) // convert hex to bytes
		if err != nil {
			return "", fmt.Errorf("invalid signature format: %w", err)
		}
		
		// Encode bytes to base64 to match our format
		actualSignature = base64.StdEncoding.EncodeToString(decodedBytes)
	}

	// Parse feed to get channelID - needed to find the right secret
	var feed Feed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return "", fmt.Errorf("failed to parse feed: %w", err)
	}
	
	if len(feed.Entries) == 0 {
		return "", fmt.Errorf("no entries found in feed")
	}

	// Use ChannelID from the entry directly instead of from Author
	channelID := feed.Entries[0].ChannelID
	if channelID == "" {
		return "", errors.New("channel ID not found in feed")
	}

	// Get subscription secret for the ChannelID
	var subscription models.HubSubscription
	if err := s.db.Where("channel_id = ?", channelID).First(&subscription).Error; err != nil {
		return "", fmt.Errorf("subscription not found: %w", err)
	}

	// Calculate HMAC
	mac := hmac.New(sha1.New, []byte(subscription.Secret))
	mac.Write(body)
	expectedSignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if actualSignature != expectedSignature {
		log.Printf("Signature mismatch. Expected: %s, Got: %s", expectedSignature, actualSignature)
		return "", errors.New("signature mismatch")
	}

	return channelID, nil
}

// generateThumbnailURL generates the URL for the video thumbnail
func (s *PubSubService) generateThumbnailURL(videoID string) string {
	return fmt.Sprintf("https://img.youtube.com/vi/%s/maxresdefault.jpg", videoID)
}

// processEntry processes a single feed entry
func (s *PubSubService) processEntry(entry Entry) error {
	videoID := entry.VideoID
	
	if videoID == "" {
		// Try to extract from entry.ID if VideoID is empty
		// Example format: yt:video:VIDEO_ID
		parts := strings.Split(entry.ID, ":")
		if len(parts) == 3 && parts[0] == "yt" && parts[1] == "video" {
			videoID = parts[2]
		}
		
		if videoID == "" {
			return fmt.Errorf("failed to extract video ID from entry")
		}
	}

	// Find the channel in the database by YouTube channel ID
	var channel models.Channel
	if err := s.db.Where("youtube_id = ?", entry.ChannelID).First(&channel).Error; err != nil {
		return fmt.Errorf("channel not found: %w", err)
	}

	// Parse published date
	publishedAt, err := time.Parse(time.RFC3339, entry.Published)
	if err != nil {
		return fmt.Errorf("failed to parse published date: %w", err)
	}

	// Get video details from YouTube API
	var videoDetails *VideoDetails
	if s.youtubeService != nil {
		videoDetails, err = s.youtubeService.GetVideoDetails(videoID)
		if err != nil {
			// Log error but continue with basic info
			log.Printf("Warning: Failed to fetch video details from YouTube API: %v", err)
		}
	}
	
	// Use available info or fallback to basic info
	if videoDetails == nil {
		videoDetails = &VideoDetails{
			ID:        videoID,
			Title:     entry.Title,
			Thumbnail: s.generateThumbnailURL(videoID),
		}
	}

	// Start a transaction for video creation and watchlist association
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create a new YouTube video
	video := &models.Video{
		YoutubeID:    videoID,
		ChannelID:    channel.ID,
		Title:        videoDetails.Title,
		Description:  videoDetails.Description,
		ThumbnailURL: videoDetails.Thumbnail,
		Duration:     videoDetails.Duration,
		PublishedAt:  publishedAt,
	}

	// Create or update video 
	var existingVideo models.Video
	if err := tx.Where("youtube_id = ?", videoID).First(&existingVideo).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return fmt.Errorf("database error: %w", err)
		}
		
		// Create new video
		if err := tx.Create(video).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to save video: %w", err)
		}
	} else {
		// Update existing video
		existingVideo.Title = videoDetails.Title
		existingVideo.Description = videoDetails.Description
		existingVideo.ThumbnailURL = videoDetails.Thumbnail
		existingVideo.Duration = videoDetails.Duration
		if !publishedAt.IsZero() {
			existingVideo.PublishedAt = publishedAt
		}
		
		if err := tx.Save(&existingVideo).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update video: %w", err)
		}
		
		// Use the existing video for watchlist associations
		video = &existingVideo
	}

	// Find all watchlists that contain this channel
	var watchlists []models.Watchlist
	if err := tx.Joins("JOIN watchlist_channels ON watchlist_channels.watchlist_id = watchlists.id").
		Where("watchlist_channels.channel_id = ?", channel.ID).
		Find(&watchlists).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to find watchlists: %w", err)
	}

	// Add video to each watchlist
	for _, watchlist := range watchlists {
		// Check if video is already in watchlist
		var count int64
		if err := tx.Model(&models.Watchlist{}).
			Joins("JOIN watchlist_videos ON watchlist_videos.watchlist_id = watchlists.id").
			Where("watchlists.id = ? AND watchlist_videos.video_id = ?", watchlist.ID, video.ID).
			Count(&count).Error; err != nil {
			log.Printf("Warning: Error checking if video exists in watchlist: %v", err)
			continue
		}
		
		if count == 0 {
			if err := tx.Exec("INSERT INTO watchlist_videos (watchlist_id, video_id) VALUES (?, ?)", 
				watchlist.ID, video.ID).Error; err != nil {
				log.Printf("Error adding video to watchlist %d: %v", watchlist.ID, err)
				continue
			}
		}
	}

	return tx.Commit().Error
} 