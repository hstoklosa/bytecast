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
	"strings"
	"time"

	"bytecast/configs"
	"bytecast/internal/models"

	"gorm.io/gorm"
)

const (
	pubsubHubURL = "https://pubsubhubbub.appspot.com/subscribe"
	feedURL      = "https://www.youtube.com/feeds/videos.xml?channel_id=%s"
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
			ChannelTitle:  "", // updated when we receive notifications
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
		existing.LastVerifiedAt = time.Now()
		
		if err := s.db.Save(&existing).Error; err != nil {
			return fmt.Errorf("failed to update subscription record: %w", err)
		}
		
		subscription = existing
		log.Printf("Updated existing subscription record for channel %s", channelID)
	}

	// Subscribe to the hub (if fails then keep the subscription record in the database)
	err = s.subscribeToHub(channelID)
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

	// Mark as inactive but don't delete yet
	subscription.IsActive = false
	if err := s.db.Save(&subscription).Error; err != nil {
		return fmt.Errorf("failed to mark subscription as inactive: %w", err)
	}
	
	log.Printf("Marked subscription for channel %s as inactive", channelID)

	// Send unsubscribe request to hub
	err := s.unsubscribeFromHub(channelID)
	if err != nil {
		log.Printf("Failed to unsubscribe from hub for channel %s: %v", channelID, err)
	} else {
		log.Printf("Successfully sent hub unsubscription request for channel %s", channelID)
	}

	// Now we can delete the subscription
	if err := s.db.Delete(&subscription).Error; err != nil {
		return fmt.Errorf("failed to delete subscription record: %w", err)
	}
	
	log.Printf("Deleted subscription record for channel %s", channelID)
	return nil
}

// subscribeToHub sends a subscription request to the PubSubHubbub hub
func (s *PubSubService) subscribeToHub(channelID string) error {
	feedURL := fmt.Sprintf(feedURL, channelID)
	
	req, err := http.NewRequest("POST", pubsubHubURL, nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("hub.callback", s.config.YouTube.CallbackURL)
	q.Add("hub.topic", feedURL)
	q.Add("hub.verify", "sync")
	q.Add("hub.mode", "subscribe")
	q.Add("hub.lease_seconds", fmt.Sprintf("%d", s.config.YouTube.LeaseSeconds))
	req.URL.RawQuery = q.Encode()

	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 204 No Content is a success response for PubSubHubbub
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
		log.Printf("Successful subscription request for channel %s (status: %d)", channelID, resp.StatusCode)
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%w: subscription failed with status %d: %s", ErrPubSubHubError, resp.StatusCode, string(body))
	}

	return nil
}

// unsubscribeFromHub sends an unsubscribe request to the PubSubHubbub hub
func (s *PubSubService) unsubscribeFromHub(channelID string) error {
	feedURL := fmt.Sprintf(feedURL, channelID)
	
	// Prepare unsubscribe request
	req, err := http.NewRequest("POST", pubsubHubURL, nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("hub.callback", s.config.YouTube.CallbackURL)
	q.Add("hub.topic", feedURL)
	q.Add("hub.verify", "sync")
	q.Add("hub.mode", "unsubscribe")
	req.URL.RawQuery = q.Encode()

	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 204 No Content is a success response for PubSubHubbub
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
		log.Printf("Successful unsubscription request for channel %s (status: %d)", channelID, resp.StatusCode)
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%w: unsubscribe failed with status %d: %s", ErrPubSubHubError, resp.StatusCode, string(body))
	}

	return nil
}

// generateSecret generates a random secret for HMAC verification
func generateSecret() string {
	// Generate a cryptographically secure random 32-byte secret
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		// Fallback to less secure method if crypto/rand fails
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
	Published string `xml:"published"`
	Updated   string `xml:"updated"`
	Author    Author `xml:"author"`
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
	// Verify notification signature
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
			// Log error but continue processing other entries
			log.Printf("Error processing entry: %v", err)
		}
	}

	return nil
}

// verifySignature verifies the HMAC signature of a notification and returns the channel ID
func (s *PubSubService) verifySignature(body []byte, signature string) (string, error) {
	// Extract channel ID from the feed
	var feed Feed
	if err := xml.Unmarshal(body, &feed); err != nil || len(feed.Entries) == 0 {
		return "", fmt.Errorf("failed to parse feed: %w", err)
	}
	
	channelID := feed.Entries[0].Author.ChannelID
	if channelID == "" {
		return "", errors.New("channel ID not found in feed")
	}

	// Get subscription secret for this channel
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
	if err := s.db.Where("you_tube_id = ?", entry.Author.ChannelID).First(&channel).Error; err != nil {
		return fmt.Errorf("channel not found: %w", err)
	}

	// Parse published date
	publishedAt, err := time.Parse(time.RFC3339, entry.Published)
	if err != nil {
		publishedAt = time.Now()
	}

	// Create a new YouTube video
	video := &models.YouTubeVideo{
		YouTubeID:    videoID,
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