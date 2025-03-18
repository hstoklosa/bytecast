package services

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
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
	db            *gorm.DB
	callbackURL   string
	leaseSeconds  int
	client        *http.Client
}

// NewPubSubService creates a new PubSub service instance
func NewPubSubService(db *gorm.DB, config *configs.Config) (*PubSubService, error) {
	if config == nil || config.YouTube.CallbackURL == "" {
		return nil, errors.New("config with callback URL is required")
	}

	leaseSeconds := config.YouTube.LeaseSeconds
	if leaseSeconds <= 0 {
		// Default to 5 days if not specified (YouTube's maximum is 10 days)
		leaseSeconds = 432000
	}

	return &PubSubService{
		db:           db,
		callbackURL:  config.YouTube.CallbackURL,
		leaseSeconds: leaseSeconds,
		client:       &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// SubscribeToChannel subscribes to a YouTube channel's notifications
func (s *PubSubService) SubscribeToChannel(channelID string) error {
	// Check if subscription already exists
	var existing models.YouTubeSubscription
	err := s.db.Where("channel_id = ?", channelID).First(&existing).Error
	
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		// Database error other than "not found"
		return fmt.Errorf("database error checking for existing subscription: %w", err)
	}
	
	var subscription models.YouTubeSubscription
	
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new subscription
		subscription = models.YouTubeSubscription{
			ChannelID:     channelID,
			ChannelTitle:  "", // Will be updated when we receive notifications
			LeaseSeconds:  s.leaseSeconds,
			ExpiresAt:     time.Now().Add(time.Duration(s.leaseSeconds) * time.Second),
			Secret:        generateSecret(), // Generate a random secret for verification
			IsActive:      true,
		}
		
		if err := s.db.Create(&subscription).Error; err != nil {
			return fmt.Errorf("failed to create subscription record: %w", err)
		}
		
		log.Printf("Created new subscription record for channel %s", channelID)
	} else {
		// Subscription exists, update it
		existing.LeaseSeconds = s.leaseSeconds
		existing.ExpiresAt = time.Now().Add(time.Duration(s.leaseSeconds) * time.Second)
		existing.IsActive = true
		existing.LastVerifiedAt = time.Now()
		
		if err := s.db.Save(&existing).Error; err != nil {
			return fmt.Errorf("failed to update subscription record: %w", err)
		}
		
		subscription = existing
		log.Printf("Updated existing subscription record for channel %s", channelID)
	}

	// Subscribe to the hub
	// If this fails, we still keep the subscription record in the database
	// but we'll log the error
	err = s.subscribeToHub(channelID)
	if err != nil {
		// Log the error but don't return it - we want to continue with the channel addition
		log.Printf("Failed to subscribe to hub for channel %s: %v", channelID, err)
		return nil
	}
	
	log.Printf("Successfully sent hub subscription request for channel %s", channelID)
	return nil
}

// UnsubscribeFromChannel unsubscribes from a YouTube channel's notifications
func (s *PubSubService) UnsubscribeFromChannel(channelID string) error {
	// Find the subscription
	var subscription models.YouTubeSubscription
	if err := s.db.Where("channel_id = ?", channelID).First(&subscription).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Nothing to unsubscribe from
			log.Printf("No subscription record found for channel %s, nothing to unsubscribe", channelID)
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
		// Don't return the error - we want to delete the subscription regardless
		// of whether the unsubscribe request succeeded
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

// RenewAllSubscriptions renews all active subscriptions that are close to expiring
func (s *PubSubService) RenewAllSubscriptions() error {
	// Find subscriptions that need renewal (expiring within 24 hours)
	var subscriptions []models.YouTubeSubscription
	if err := s.db.Where("is_active = ? AND expires_at <= ?", true, time.Now().Add(24*time.Hour)).Find(&subscriptions).Error; err != nil {
		return err
	}

	log.Printf("Renewing %d YouTube PubSubHubbub subscriptions", len(subscriptions))

	// Renew each subscription
	for _, sub := range subscriptions {
		if err := s.renewSubscription(sub); err != nil {
			log.Printf("Error renewing subscription for channel %s: %v", sub.ChannelID, err)
			continue
		}
		log.Printf("Successfully renewed subscription for channel %s", sub.ChannelID)
	}

	return nil
}

// renewSubscription renews a single subscription
func (s *PubSubService) renewSubscription(sub models.YouTubeSubscription) error {
	// Update subscription in database
	sub.LeaseSeconds = s.leaseSeconds
	sub.ExpiresAt = time.Now().Add(time.Duration(s.leaseSeconds) * time.Second)
	if err := s.db.Save(&sub).Error; err != nil {
		return err
	}

	// Resubscribe to the channel
	return s.subscribeToHub(sub.ChannelID)
}

// subscribeToHub sends a subscription request to the PubSubHubbub hub
func (s *PubSubService) subscribeToHub(channelID string) error {
	feedURL := fmt.Sprintf(feedURL, channelID)
	
	// Prepare subscription request
	req, err := http.NewRequest("POST", pubsubHubURL, nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("hub.callback", s.callbackURL)
	q.Add("hub.topic", feedURL)
	q.Add("hub.verify", "sync")
	q.Add("hub.mode", "subscribe")
	q.Add("hub.lease_seconds", fmt.Sprintf("%d", s.leaseSeconds))
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
	q.Add("hub.callback", s.callbackURL)
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
