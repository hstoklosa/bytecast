package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"

	"bytecast/configs"
)

var (
	ErrYouTubeAPIError     = errors.New("error calling YouTube API")
	ErrChannelNotFoundAPI  = errors.New("channel not found on YouTube")
	ErrInvalidYouTubeURL   = errors.New("invalid YouTube channel URL")
	ErrMissingAPIKey       = errors.New("YouTube API key is not configured")
)

// YouTubeService handles interactions with the YouTube API
type YouTubeService struct {
	apiKey     string
	httpClient *http.Client
}

// ChannelInfo contains the essential information about a YouTube channel
type ChannelInfo struct {
	ID          string
	Title       string
	Description string
	Thumbnail   string
}

// VideoDetails contains the essential information about a YouTube video
type VideoDetails struct {
	ID          string
	Title       string
	Description string
	Thumbnail   string
	Duration    string
}

// NewYouTubeService creates a new YouTube service
func NewYouTubeService(config *configs.Config) (*YouTubeService, error) {
	if config == nil {
		return nil, errors.New("config is required")
	}

	if config.YouTube.APIKey == "" {
		return nil, ErrMissingAPIKey
	}

	return &YouTubeService{
		apiKey:     config.YouTube.APIKey,
		httpClient: &http.Client{},
	}, nil
}

// GetChannelInfo retrieves channel information from YouTube API
func (s *YouTubeService) GetChannelInfo(channelID string) (*ChannelInfo, error) {
	ctx := context.Background()
	
	// Create the YouTube service with the API key
	youtubeService, err := youtube.NewService(ctx, option.WithAPIKey(s.apiKey))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrYouTubeAPIError, err)
	}

	// Extract the actual channel ID if a URL was provided
	extractedID, err := s.ExtractChannelID(channelID)
	if err != nil {
		return nil, err
	}

	// First try direct channel ID lookup for standard YouTube channel IDs
	// This is more efficient than search and uses less quota
	if regexp.MustCompile(`^UC[a-zA-Z0-9_-]{22}$`).MatchString(extractedID) {
		call := youtubeService.Channels.List([]string{"snippet"}).Id(extractedID)
		response, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrYouTubeAPIError, err)
		}
		
		if len(response.Items) > 0 {
			return extractChannelInfo(response.Items[0]), nil
		}
	}
	
	// For custom URLs with @ symbol, try to get by handle first
	// This is more efficient than search for custom URLs
	if strings.HasPrefix(extractedID, "@") {
		handle := strings.TrimPrefix(extractedID, "@")
		call := youtubeService.Channels.List([]string{"snippet"}).ForHandle(handle)
		response, err := call.Do()
		if err != nil {
			// Continue to search if this fails
		} else if len(response.Items) > 0 {
			return extractChannelInfo(response.Items[0]), nil
		}
	}
	
	// If direct ID lookup failed or it's not a channel ID, try search
	// Search uses more quota but can find channels by username or custom URL
	searchCall := youtubeService.Search.List([]string{"snippet"}).
		Q(extractedID).
		Type("channel").
		MaxResults(1)
	
	searchResponse, err := searchCall.Do()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrYouTubeAPIError, err)
	}
	
	if len(searchResponse.Items) == 0 {
		return nil, ErrChannelNotFoundAPI
	}
	
	// Get the actual channel ID from search results
	foundChannelID := searchResponse.Items[0].Id.ChannelId
	
	// Now get the full channel details
	call := youtubeService.Channels.List([]string{"snippet"}).Id(foundChannelID)
	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrYouTubeAPIError, err)
	}
	
	if len(response.Items) == 0 {
		return nil, ErrChannelNotFoundAPI
	}
	
	return extractChannelInfo(response.Items[0]), nil
}

// Helper function to extract channel info from a YouTube API response item
func extractChannelInfo(channel *youtube.Channel) *ChannelInfo {
	thumbnailURL := ""
	if channel.Snippet.Thumbnails != nil {
		if channel.Snippet.Thumbnails.High != nil {
			thumbnailURL = channel.Snippet.Thumbnails.High.Url
		} else if channel.Snippet.Thumbnails.Medium != nil {
			thumbnailURL = channel.Snippet.Thumbnails.Medium.Url
		} else if channel.Snippet.Thumbnails.Default != nil {
			thumbnailURL = channel.Snippet.Thumbnails.Default.Url
		}
	}
	
	return &ChannelInfo{
		ID:          channel.Id,
		Title:       channel.Snippet.Title,
		Description: channel.Snippet.Description,
		Thumbnail:   thumbnailURL,
	}
}

// ExtractChannelID extracts a YouTube channel ID from various URL formats
func (s *YouTubeService) ExtractChannelID(input string) (string, error) {
	// Trim whitespace and handle empty input
	input = strings.TrimSpace(input)
	if input == "" {
		return "", ErrInvalidYouTubeURL
	}

	// If it's already just an ID (starts with UC), return it
	if regexp.MustCompile(`^UC[a-zA-Z0-9_-]{22}$`).MatchString(input) {
		return input, nil
	}

	// Handle custom URLs that start with @
	if strings.HasPrefix(input, "@") {
		return input, nil
	}

	// Handle URLs with or without protocol
	if !strings.Contains(input, "://") && !strings.HasPrefix(input, "www.") {
		// Add https:// if it doesn't have a protocol
		if strings.Contains(input, "youtube.com") || strings.Contains(input, "youtu.be") {
			input = "https://" + input
		}
	} else if strings.HasPrefix(input, "www.") {
		input = "https://" + input
	}

	// Handle URLs like https://www.youtube.com/channel/UC_x5XG1OV2P6uZZ5FSM9Ttw
	channelIDRegex := regexp.MustCompile(`(?:youtube\.com/channel/|youtube\.com/c/|youtube\.com/@)([\w-]+)`)
	matches := channelIDRegex.FindStringSubmatch(input)
	if len(matches) > 1 {
		return matches[1], nil
	}

	// Handle URLs like https://www.youtube.com/user/username
	userRegex := regexp.MustCompile(`youtube\.com/user/([\w-]+)`)
	matches = userRegex.FindStringSubmatch(input)
	if len(matches) > 1 {
		return matches[1], nil
	}

	// Handle URLs like https://www.youtube.com/@username
	atRegex := regexp.MustCompile(`youtube\.com/@([\w-]+)`)
	matches = atRegex.FindStringSubmatch(input)
	if len(matches) > 1 {
		return "@" + matches[1], nil
	}

	// Handle short URLs like https://youtu.be/channelname
	shortRegex := regexp.MustCompile(`youtu\.be/([\w-]+)`)
	matches = shortRegex.FindStringSubmatch(input)
	if len(matches) > 1 {
		return matches[1], nil
	}

	return "", ErrInvalidYouTubeURL
}

// GetVideoDetails retrieves video information from YouTube API
func (s *YouTubeService) GetVideoDetails(videoID string) (*VideoDetails, error) {
	ctx := context.Background()
	
	// Create the YouTube service with the API key
	youtubeService, err := youtube.NewService(ctx, option.WithAPIKey(s.apiKey))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrYouTubeAPIError, err)
	}

	// Get video details
	call := youtubeService.Videos.List([]string{"snippet", "contentDetails"}).Id(videoID)
	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrYouTubeAPIError, err)
	}

	if len(response.Items) == 0 {
		return nil, fmt.Errorf("video not found")
	}

	video := response.Items[0]
	
	// Get the highest quality thumbnail available
	thumbnailURL := ""
	if video.Snippet.Thumbnails != nil {
		if video.Snippet.Thumbnails.Maxres != nil {
			thumbnailURL = video.Snippet.Thumbnails.Maxres.Url
		} else if video.Snippet.Thumbnails.High != nil {
			thumbnailURL = video.Snippet.Thumbnails.High.Url
		} else if video.Snippet.Thumbnails.Medium != nil {
			thumbnailURL = video.Snippet.Thumbnails.Medium.Url
		} else if video.Snippet.Thumbnails.Default != nil {
			thumbnailURL = video.Snippet.Thumbnails.Default.Url
		}
	}

	return &VideoDetails{
		ID:          video.Id,
		Title:       video.Snippet.Title,
		Description: video.Snippet.Description,
		Thumbnail:   thumbnailURL,
		Duration:    video.ContentDetails.Duration,
	}, nil
} 