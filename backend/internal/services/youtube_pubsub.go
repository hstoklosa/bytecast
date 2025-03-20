package services

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
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

// ErrPubSubHubError is returned when there's an error communicating with the PubSubHubbub hub
var ErrPubSubHubError = errors.New("error communicating with PubSubHubbub hub")

// PubSubService handles YouTube channel subscriptions via PubSubHubbub
type PubSubService struct {
	db           *gorm.DB
	config       *configs.Config
	client       *http.Client
	videoService *VideoService
}

// NewPubSubService creates a new PubSub service instance
func NewPubSubService(db *gorm.DB, config *configs.Config, videoService *VideoService) *PubSubService {
	return &PubSubService{
		db:           db,
		config:       config,
		client:       &http.Client{Timeout: 10 * time.Second},
		videoService: videoService,
	}
}

// SubscribeToChannel subscribes to a YouTube channel's notifications
func (s *PubSubService) SubscribeToChannel(channelID string) error {
	var existing models.YouTubeSubscription
	err := s.db.Where("channel_id = ?", channelID).First(&existing).Error
	
	// Errors other than "not found"
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("database error checking for existing subscription: %w", err)
	}
	
	var subscription models.YouTubeSubscription
	
	if errors.Is(err, gorm.ErrRecordNotFound) {
		subscription = models.YouTubeSubscription{
			ChannelID:     channelID,
			LeaseSeconds:  s.config.YouTube.LeaseSeconds,
			ExpiresAt:     time.Now().Add(time.Duration(s.config.YouTube.LeaseSeconds) * time.Second),
			Secret:        generateSecret(),
			IsActive:      true,
		}
		
		if err := s.db.Create(&subscription).Error; err != nil {
			return fmt.Errorf("failed to create subscription record: %w", err)
		}
		
		log.Printf("Created new subscription record for channel %s", channelID)
	} else {
		// Subscription exists, update it
		existing.LeaseSeconds = s.config.YouTube.LeaseSeconds
		existing.ExpiresAt = time.Now().Add(time.Duration(s.config.YouTube.LeaseSeconds) * time.Second)
		existing.IsActive = true
		
		if err := s.db.Save(&existing).Error; err != nil {
			return fmt.Errorf("failed to update subscription record: %w", err)
		}
		
		subscription = existing
		log.Printf("Updated existing subscription record for channel %s", channelID)
	}

	// Subscribe to the hub (if fails then keep the subscription record in the database)
	err = s.sendSubscriptionRequest(channelID, "subscribe")
	if err != nil {
		log.Printf("Failed to subscribe to hub for channel %s: %v", channelID, err)
		return nil
	}
	
	log.Printf("Successfully sent hub subscription request for channel %s", channelID)
	return nil
}

// UnsubscribeFromChannel unsubscribes from a YouTube channel's notifications
func (s *PubSubService) UnsubscribeFromChannel(channelID string) error {
	var subscription models.YouTubeSubscription
	if err := s.db.Where("channel_id = ?", channelID).First(&subscription).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("No subscription record found for channel %s", channelID)
			return nil
		}

		return fmt.Errorf("database error when finding subscription: %w", err)
	}

	subscription.IsActive = false
	if err := s.db.Save(&subscription).Error; err != nil {
		return fmt.Errorf("failed to mark subscription as inactive: %w", err)
	}
	
	// Send unsubscribe request to hub
	err := s.sendSubscriptionRequest(channelID, "unsubscribe")
	if err != nil {
		log.Printf("Failed to unsubscribe from hub for channel %s: %v", channelID, err)
	} 

	if err := s.db.Delete(&subscription).Error; err != nil {
		return fmt.Errorf("failed to delete subscription record: %w", err)
	}
	
	return nil
}

func (s *PubSubService) sendSubscriptionRequest(channelID, mode string) error {
	feedURL := fmt.Sprintf(feedURL, channelID)
	callbackURL := s.config.YouTube.CallbackURL
	
	form := url.Values{}
	form.Set("hub.callback", callbackURL)
	form.Set("hub.mode", mode)
	form.Set("hub.topic", feedURL)
	form.Set("hub.verify", "async")

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

// Feed represents the YouTube video feed XML structure
type Feed struct {
	XMLName xml.Name `xml:"feed"`
	Entries []Entry  `xml:"entry"`
}

// Entry represents a video entry in the feed
type Entry struct {
	ID        string `xml:"id"`
	VideoID   string `xml:"videoId"`
	Title     string `xml:"title"`
	Link      Link   `xml:"link"`
	Author    Author `xml:"author"`
	Published string `xml:"published"`
	Updated   string `xml:"updated"`
}

// Link represents a link in the feed
type Link struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

// Author represents the video author in the feed
type Author struct {
	Name      string `xml:"name"`
	ChannelID string `xml:"channelId"`
}

// ProcessVideoNotification processes an incoming video notification
func (s *PubSubService) ProcessVideoNotification(body []byte, signature string) error {
	channelID, err := s.verifySignature(body, signature)
	if err != nil {
		return fmt.Errorf("invalid notification signature: %w", err)
	}

	// Parse notification XML
	var feed Feed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return fmt.Errorf("failed to parse notification: %w", err)
	}

	log.Printf("Received PubSubHubbub notification for channel %s with %d entries", channelID, len(feed.Entries))

	// Process each entry
	for _, entry := range feed.Entries {
		if err := s.processEntry(entry); err != nil {
			log.Printf("Error processing entry: %v", err)
		}
	}

	return nil
}

// verifySignature verifies the HMAC signature of a notification and returns the channel ID
func (s *PubSubService) verifySignature(body []byte, signature string) (string, error) {
	var feed Feed
	if err := xml.Unmarshal(body, &feed); err != nil || len(feed.Entries) == 0 {
		return "", fmt.Errorf("failed to parse feed: %w", err)
	}
	
	channelID := feed.Entries[0].Author.ChannelID
	if channelID == "" {
		return "", errors.New("channel ID not found in feed")
	}

	// Get subscription secret for the ChannelID
	var subscription models.YouTubeSubscription
	if err := s.db.Where("channel_id = ?", channelID).First(&subscription).Error; err != nil {
		return "", fmt.Errorf("subscription not found: %w", err)
	}

	// Calculate HMAC
	mac := hmac.New(sha1.New, []byte(subscription.Secret))
	mac.Write(body)
	expectedSignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if signature != expectedSignature {
		return "", errors.New("signature mismatch")
	}

	return channelID, nil
}

// extractVideoID extracts the video ID from the entry ID
func (s *PubSubService) extractVideoID(entryID string) string {
	// Entry ID format: yt:video:VIDEO_ID
	parts := strings.Split(entryID, ":")
	if len(parts) < 3 {
		return ""
	}
	return parts[2]
}

// generateThumbnailURL generates the URL for the video thumbnail
func (s *PubSubService) generateThumbnailURL(videoID string) string {
	return fmt.Sprintf("https://img.youtube.com/vi/%s/maxresdefault.jpg", videoID)
}

// processEntry processes a single feed entry
func (s *PubSubService) processEntry(entry Entry) error {
	// Extract video ID from the entry ID
	videoID := s.extractVideoID(entry.ID)
	if videoID == "" {
		return fmt.Errorf("failed to extract video ID from entry ID: %s", entry.ID)
	}

	// Find the channel in the database by YouTube channel ID
	var channel models.Channel
	if err := s.db.Where("youtube_id = ?", entry.Author.ChannelID).First(&channel).Error; err != nil {
		return fmt.Errorf("channel not found: %w", err)
	}

	// Parse published date
	publishedAt, err := time.Parse(time.RFC3339, entry.Published)
	if err != nil {
		publishedAt = time.Now()
	}

	// Create a new YouTube video
	video := &models.YouTubeVideo{
		YoutubeID:    videoID,
		ChannelID:    channel.ID,
		Title:        entry.Title,
		PublishedAt:  publishedAt,
		ThumbnailURL: s.generateThumbnailURL(videoID),
	}

	// Create or update video using the video service
	if err := s.videoService.CreateVideo(video); err != nil {
		return fmt.Errorf("failed to save video: %w", err)
	}

	// Find all watchlists that contain this channel
	var watchlists []models.Watchlist
	if err := s.db.Joins("JOIN watchlist_channels ON watchlist_channels.watchlist_id = watchlists.id").
		Where("watchlist_channels.channel_id = ?", channel.ID).
		Find(&watchlists).Error; err != nil {
		return fmt.Errorf("failed to find watchlists: %w", err)
	}

	// Add video to each watchlist
	for _, watchlist := range watchlists {
		if err := s.db.Model(&watchlist).Association("Videos").Append(video); err != nil {
			// Log error but continue with other watchlists
			log.Printf("Error adding video to watchlist %d: %v", watchlist.ID, err)
		}
	}

	log.Printf("Successfully processed new video: %s (%s)", video.Title, videoID)
	return nil
} 