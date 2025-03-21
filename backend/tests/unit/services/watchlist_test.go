package services_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"bytecast/configs"
	"bytecast/internal/models"
	"bytecast/internal/services"
)

// Helper function to find substring index
func findStringIndex(s, substr string) int {
	return strings.Index(s, substr)
}

// MockYouTubeService is a mock implementation of the YouTube service for testing
type MockYouTubeService struct{}

// GetChannelInfo is a mock implementation that returns predefined channel info
func (m *MockYouTubeService) GetChannelInfo(channelID string) (*services.ChannelInfo, error) {
	// Return error for invalid channel IDs
	if channelID == "invalid/format" || channelID == "" {
		return nil, services.ErrInvalidYouTubeURL
	}
	
	// For testing, extract the ID from URLs
	id := channelID
	if len(channelID) > 24 && channelID[:24] == "https://www.youtube.com/" {
		parts := []string{"channel/", "c/", "user/"}
		for _, part := range parts {
			if idx := findStringIndex(channelID, part); idx != -1 {
				start := idx + len(part)
				end := len(channelID)
				if slashIdx := findStringIndex(channelID[start:], "/"); slashIdx != -1 {
					end = start + slashIdx
				}
				if queryIdx := findStringIndex(channelID[start:], "?"); queryIdx != -1 {
					if end > start+queryIdx {
						end = start + queryIdx
					}
				}
				id = channelID[start:end]
				break
			}
		}
	}
	
	// Return mock channel info
	return &services.ChannelInfo{
		ID:          id,
		Title:       "Mock Channel " + id,
		Description: "This is a mock channel for testing",
		Thumbnail:   "https://example.com/thumbnail.jpg",
	}, nil
}

// TestWatchlist is a test-specific version of the Watchlist model without the check constraint
type TestWatchlist struct {
	gorm.Model
	UserID      uint   `gorm:"index;not null"`
	Name        string `gorm:"size:255;not null"`
	Description string `gorm:"type:text"`
	Color       string `gorm:"size:7;not null"`
	Channels    []*models.Channel `gorm:"many2many:watchlist_channels;foreignKey:ID;joinForeignKey:WatchlistID;References:ID;joinReferences:ChannelID"`
}

// TableName specifies the table name for the TestWatchlist model
func (TestWatchlist) TableName() string {
	return "watchlists"
}

func setupWatchlistTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations for User, TestWatchlist, and Channel models
	err = db.AutoMigrate(&models.User{}, &TestWatchlist{}, &models.Channel{})
	require.NoError(t, err)

	return db
}

func createTestUser(t *testing.T, db *gorm.DB) *models.User {
	user := &models.User{
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashedpassword",
	}
	result := db.Create(user)
	require.NoError(t, result.Error)
	require.NotZero(t, user.ID)
	return user
}

func createMockConfig() *configs.Config {
	return &configs.Config{
		YouTube: configs.YouTube{
			APIKey: "mock-api-key",
		},
		JWT: configs.JWT{
			Secret: "mock-jwt-secret-that-is-at-least-32-chars",
		},
		Server: configs.Server{
			Port:        "8080",
			Environment: "development",
			Domain:      "localhost",
		},
		Database: configs.Database{
			Host:     "localhost",
			User:     "postgres",
			Password: "password",
			Name:     "bytecast_test",
			Port:     "5432",
		},
		Superuser: configs.Superuser{
			Username: "admin",
			Email:    "admin@example.com",
			Password: "password",
		},
	}
}

// Create a watchlist service with a mock YouTube service for testing
func createWatchlistService(db *gorm.DB, config *configs.Config) *services.WatchlistService {
	service := services.NewWatchlistService(db, config)
	
	// Replace the YouTube service with our mock
	mockYouTube := &MockYouTubeService{}
	service.SetYouTubeService(mockYouTube)
	
	return service
}

func TestCreateWatchlist(t *testing.T) {
	db := setupWatchlistTestDB(t)
	config := createMockConfig()
	watchlistService := createWatchlistService(db, config)
	user := createTestUser(t, db)

	tests := []struct {
		name        string
		userID      uint
		wlName      string
		description string
		color       string
		wantErr     bool
	}{
		{
			name:        "Valid watchlist",
			userID:      user.ID,
			wlName:      "My Watchlist",
			description: "A collection of my favorite channels",
			color:       "#FF5733",
			wantErr:     false,
		},
		{
			name:        "Invalid user ID",
			userID:      0,
			wlName:      "Invalid Watchlist",
			description: "This should fail",
			color:       "#FF5733",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watchlist, err := watchlistService.CreateWatchlist(tt.userID, tt.wlName, tt.description, tt.color)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, watchlist)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, watchlist)
				assert.Equal(t, tt.userID, watchlist.UserID)
				assert.Equal(t, tt.wlName, watchlist.Name)
				assert.Equal(t, tt.description, watchlist.Description)
				assert.Equal(t, tt.color, watchlist.Color)
			}
		})
	}
}

func TestGetWatchlist(t *testing.T) {
	db := setupWatchlistTestDB(t)
	config := createMockConfig()
	watchlistService := createWatchlistService(db, config)
	user := createTestUser(t, db)
	
	// Create a test watchlist
	watchlist := &TestWatchlist{
		UserID:      user.ID,
		Name:        "Test Watchlist",
		Description: "Test Description",
		Color:       "#FF5733",
	}
	result := db.Create(watchlist)
	require.NoError(t, result.Error)
	require.NotZero(t, watchlist.ID)

	tests := []struct {
		name        string
		watchlistID uint
		userID      uint
		wantErr     error
	}{
		{
			name:        "Valid watchlist",
			watchlistID: watchlist.ID,
			userID:      user.ID,
			wantErr:     nil,
		},
		{
			name:        "Non-existent watchlist",
			watchlistID: 999,
			userID:      user.ID,
			wantErr:     services.ErrWatchlistNotFound,
		},
		{
			name:        "Unauthorized user",
			watchlistID: watchlist.ID,
			userID:      999,
			wantErr:     services.ErrWatchlistNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := watchlistService.GetWatchlist(tt.watchlistID, tt.userID)
			
			if tt.wantErr != nil {
				assert.True(t, errors.Is(err, tt.wantErr))
				return
			}
			
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, watchlist.ID, result.ID)
			assert.Equal(t, watchlist.Name, result.Name)
			assert.Equal(t, watchlist.Description, result.Description)
			assert.Equal(t, watchlist.UserID, result.UserID)
		})
	}
}

func TestGetUserWatchlists(t *testing.T) {
	db := setupWatchlistTestDB(t)
	config := createMockConfig()
	watchlistService := createWatchlistService(db, config)
	user := createTestUser(t, db)
	
	// Create another user
	user2 := &models.User{
		Email:        "test2@example.com",
		Username:     "testuser2",
		PasswordHash: "hashedpassword",
	}
	result := db.Create(user2)
	require.NoError(t, result.Error)
	require.NotZero(t, user2.ID)
	
	// Create watchlists for user1
	watchlists := []TestWatchlist{
		{UserID: user.ID, Name: "Watchlist 1", Description: "Description 1", Color: "#FF5733"},
		{UserID: user.ID, Name: "Watchlist 2", Description: "Description 2", Color: "#3366FF"},
	}
	for i := range watchlists {
		result := db.Create(&watchlists[i])
		require.NoError(t, result.Error)
		require.NotZero(t, watchlists[i].ID)
	}
	
	// Create a watchlist for user2
	watchlist := TestWatchlist{UserID: user2.ID, Name: "User2 Watchlist", Description: "User2 Description", Color: "#33FF57"}
	result = db.Create(&watchlist)
	require.NoError(t, result.Error)
	require.NotZero(t, watchlist.ID)

	tests := []struct {
		name    string
		userID  uint
		want    int
		wantErr bool
	}{
		{
			name:    "User with watchlists",
			userID:  user.ID,
			want:    2,
			wantErr: false,
		},
		{
			name:    "User with one watchlist",
			userID:  user2.ID,
			want:    1,
			wantErr: false,
		},
		{
			name:    "User with no watchlists",
			userID:  999,
			want:    0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := watchlistService.GetUserWatchlists(tt.userID)
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			
			assert.NoError(t, err)
			assert.Len(t, result, tt.want)
			
			if tt.userID == user.ID {
				assert.Equal(t, watchlists[0].Name, result[0].Name)
				assert.Equal(t, watchlists[1].Name, result[1].Name)
			} else if tt.userID == user2.ID {
				assert.Equal(t, watchlist.Name, result[0].Name)
			}
		})
	}
}

func TestUpdateWatchlist(t *testing.T) {
	db := setupWatchlistTestDB(t)
	config := createMockConfig()
	watchlistService := createWatchlistService(db, config)
	user := createTestUser(t, db)

	// Create a test watchlist
	watchlist := TestWatchlist{
		UserID:      user.ID,
		Name:        "Original Name",
		Description: "Original Description",
		Color:       "#FF5733",
	}
	result := db.Create(&watchlist)
	assert.NoError(t, result.Error)
	assert.NotZero(t, watchlist.ID)

	tests := []struct {
		name        string
		watchlistID uint
		userID      uint
		newName     string
		newDesc     string
		newColor    string
		wantErr     bool
	}{
		{
			name:        "Valid update",
			watchlistID: watchlist.ID,
			userID:      user.ID,
			newName:     "Updated Name",
			newDesc:     "Updated Description",
			newColor:    "#3366FF",
			wantErr:     false,
		},
		{
			name:        "Non-existent watchlist",
			watchlistID: 9999,
			userID:      user.ID,
			newName:     "Updated Name",
			newDesc:     "Updated Description",
			newColor:    "#3366FF",
			wantErr:     true,
		},
		{
			name:        "Wrong user",
			watchlistID: watchlist.ID,
			userID:      9999,
			newName:     "Updated Name",
			newDesc:     "Updated Description",
			newColor:    "#3366FF",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, err := watchlistService.UpdateWatchlist(tt.watchlistID, tt.userID, tt.newName, tt.newDesc, tt.newColor)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.newName, updated.Name)
			assert.Equal(t, tt.newDesc, updated.Description)
			assert.Equal(t, tt.newColor, updated.Color)
		})
	}
}

func TestDeleteWatchlist(t *testing.T) {
	db := setupWatchlistTestDB(t)
	config := createMockConfig()
	watchlistService := createWatchlistService(db, config)
	user := createTestUser(t, db)
	
	// Create a test watchlist
	watchlist := &TestWatchlist{
		UserID:      user.ID,
		Name:        "Test Watchlist",
		Description: "Test Description",
		Color:       "#FF5733",
	}
	result := db.Create(watchlist)
	require.NoError(t, result.Error)
	require.NotZero(t, watchlist.ID)

	tests := []struct {
		name        string
		watchlistID uint
		userID      uint
		wantErr     error
	}{
		{
			name:        "Valid delete",
			watchlistID: watchlist.ID,
			userID:      user.ID,
			wantErr:     nil,
		},
		{
			name:        "Non-existent watchlist",
			watchlistID: 999,
			userID:      user.ID,
			wantErr:     services.ErrWatchlistNotFound,
		},
		{
			name:        "Unauthorized user",
			watchlistID: watchlist.ID,
			userID:      999,
			wantErr:     services.ErrWatchlistNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := watchlistService.DeleteWatchlist(tt.watchlistID, tt.userID)
			
			if tt.wantErr != nil {
				assert.True(t, errors.Is(err, tt.wantErr))
				return
			}
			
			assert.NoError(t, err)
			
			// Verify deletion in database
			var count int64
			db.Model(&TestWatchlist{}).Where("id = ?", tt.watchlistID).Count(&count)
			assert.Equal(t, int64(0), count)
		})
	}
}

func TestExtractYouTubeChannelID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Plain channel ID",
			input: "UC_x5XG1OV2P6uZZ5FSM9Ttw",
			want:  "UC_x5XG1OV2P6uZZ5FSM9Ttw",
		},
		{
			name:  "Full URL",
			input: "https://www.youtube.com/channel/UC_x5XG1OV2P6uZZ5FSM9Ttw",
			want:  "UC_x5XG1OV2P6uZZ5FSM9Ttw",
		},
		{
			name:  "URL with trailing slash",
			input: "https://www.youtube.com/channel/UC_x5XG1OV2P6uZZ5FSM9Ttw/",
			want:  "UC_x5XG1OV2P6uZZ5FSM9Ttw",
		},
		{
			name:  "URL with query parameters",
			input: "https://www.youtube.com/channel/UC_x5XG1OV2P6uZZ5FSM9Ttw?view=videos",
			want:  "UC_x5XG1OV2P6uZZ5FSM9Ttw",
		},
		{
			name:  "Invalid URL format",
			input: "invalid/format",
			want:  "", // Not supported in current implementation
		},
		{
			name:  "Empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Since extractYouTubeChannelID is private, we need to test it indirectly
			// We'll use AddChannelToWatchlist with a mock DB that returns ErrInvalidYouTubeID
			// if the extracted ID is empty
			
			db := setupWatchlistTestDB(t)
			config := createMockConfig()
			watchlistService := createWatchlistService(db, config)
			
			// Create a test user and watchlist
			user := createTestUser(t, db)
			watchlist := &TestWatchlist{
				UserID:      user.ID,
				Name:        "Test Watchlist",
				Description: "Test Description",
				Color:       "#FF5733",
			}
			result := db.Create(watchlist)
			require.NoError(t, result.Error)
			
			// Test the extraction
			err := watchlistService.AddChannelToWatchlist(watchlist.ID, user.ID, tt.input)
			
			if tt.want == "" {
				assert.ErrorIs(t, err, services.ErrInvalidYouTubeID)
			} else {
				// If valid ID, should create a channel with that ID
				if assert.NoError(t, err) {
					var channel models.Channel
					err = db.Where("youtube_id = ?", tt.want).First(&channel).Error
					assert.NoError(t, err)
					assert.Equal(t, tt.want, channel.YoutubeID)
				}
			}
		})
	}
}

func TestAddChannelToWatchlist(t *testing.T) {
	db := setupWatchlistTestDB(t)
	config := createMockConfig()
	watchlistService := createWatchlistService(db, config)
	user := createTestUser(t, db)
	
	// Create a test watchlist
	watchlist := &TestWatchlist{
		UserID:      user.ID,
		Name:        "Test Watchlist",
		Description: "Test Description",
		Color:       "#FF5733",
	}
	result := db.Create(watchlist)
	require.NoError(t, result.Error)
	
	// Create a test channel
	channel := &models.Channel{
		YoutubeID: "UC_existing",
		Title:     "Existing Channel",
	}
	result = db.Create(channel)
	require.NoError(t, result.Error)

	tests := []struct {
		name        string
		watchlistID uint
		userID      uint
		channelID   string
		wantErr     error
	}{
		{
			name:        "Add new channel",
			watchlistID: watchlist.ID,
			userID:      user.ID,
			channelID:   "UC_new_channel",
			wantErr:     nil,
		},
		{
			name:        "Add existing channel",
			watchlistID: watchlist.ID,
			userID:      user.ID,
			channelID:   "UC_existing",
			wantErr:     nil,
		},
		{
			name:        "Invalid channel ID",
			watchlistID: watchlist.ID,
			userID:      user.ID,
			channelID:   "invalid/format",
			wantErr:     services.ErrInvalidYouTubeID,
		},
		{
			name:        "Non-existent watchlist",
			watchlistID: 999,
			userID:      user.ID,
			channelID:   "UC_new_channel",
			wantErr:     services.ErrWatchlistNotFound,
		},
		{
			name:        "Unauthorized user",
			watchlistID: watchlist.ID,
			userID:      999,
			channelID:   "UC_new_channel",
			wantErr:     services.ErrWatchlistNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := watchlistService.AddChannelToWatchlist(tt.watchlistID, tt.userID, tt.channelID)
			
			if tt.wantErr != nil {
				assert.True(t, errors.Is(err, tt.wantErr))
				return
			}
			
			assert.NoError(t, err)
			
			// Verify channel was added to watchlist
			var count int64
			var wl TestWatchlist
			wl.ID = tt.watchlistID
			count = db.Model(&wl).Association("Channels").Count()
			assert.Greater(t, count, int64(0))
			
			// var channels []models.Channel
			// err = db.Model(&TestWatchlist{ID: tt.watchlistID}).Association("Channels").Find(&channels)
			// assert.NoError(t, err)

			assert.NoError(t, err)
			assert.Greater(t, count, int64(0))
			
			// Verify channel exists in database
			var channel models.Channel
			err = db.Where("youtube_id = ?", tt.channelID).First(&channel).Error
			assert.NoError(t, err)
		})
	}
}

func TestRemoveChannelFromWatchlist(t *testing.T) {
	db := setupWatchlistTestDB(t)
	config := createMockConfig()
	watchlistService := createWatchlistService(db, config)
	user := createTestUser(t, db)
	
	// Create a test watchlist
	watchlist := &TestWatchlist{
		UserID:      user.ID,
		Name:        "Test Watchlist",
		Description: "Test Description",
		Color:       "#FF5733",
	}
	result := db.Create(watchlist)
	require.NoError(t, result.Error)
	
	// Create test channels
	channel1 := &models.Channel{
		YoutubeID: "UC_channel1",
		Title:     "Channel 1",
	}
	result = db.Create(channel1)
	require.NoError(t, result.Error)
	
	channel2 := &models.Channel{
		YoutubeID: "UC_channel2",
		Title:     "Channel 2",
	}
	result = db.Create(channel2)
	require.NoError(t, result.Error)
	
	// Add channels to watchlist
	err := db.Model(watchlist).Association("Channels").Append([]*models.Channel{channel1, channel2})
	require.NoError(t, err)

	tests := []struct {
		name        string
		watchlistID uint
		userID      uint
		channelID   string
		wantErr     error
	}{
		{
			name:        "Remove existing channel",
			watchlistID: watchlist.ID,
			userID:      user.ID,
			channelID:   "UC_channel1",
			wantErr:     nil,
		},
		{
			name:        "Remove non-existent channel",
			watchlistID: watchlist.ID,
			userID:      user.ID,
			channelID:   "UC_nonexistent",
			wantErr:     services.ErrChannelNotFound,
		},
		{
			name:        "Non-existent watchlist",
			watchlistID: 999,
			userID:      user.ID,
			channelID:   "UC_channel2",
			wantErr:     services.ErrWatchlistNotFound,
		},
		{
			name:        "Unauthorized user",
			watchlistID: watchlist.ID,
			userID:      999,
			channelID:   "UC_channel2",
			wantErr:     services.ErrWatchlistNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := watchlistService.RemoveChannelFromWatchlist(tt.watchlistID, tt.userID, tt.channelID)
			
			if tt.wantErr != nil {
				assert.True(t, errors.Is(err, tt.wantErr))
				return
			}
			
			assert.NoError(t, err)
			
			// Verify channel was removed from watchlist
			// Verify channel was removed from watchlist
			var channels []models.Channel
			var wl TestWatchlist
			wl.ID = tt.watchlistID
			err = db.Model(&wl).Association("Channels").Find(&channels)
			assert.NoError(t, err)
			for _, ch := range channels {
				assert.NotEqual(t, tt.channelID, ch.YoutubeID)
			}
		})
	}
}

func TestGetChannelsInWatchlist(t *testing.T) {
	db := setupWatchlistTestDB(t)
	config := createMockConfig()
	watchlistService := createWatchlistService(db, config)
	user := createTestUser(t, db)
	
	// Create a test watchlist
	watchlist := &TestWatchlist{
		UserID:      user.ID,
		Name:        "Test Watchlist",
		Description: "Test Description",
		Color:       "#FF5733",
	}
	result := db.Create(watchlist)
	require.NoError(t, result.Error)
	
	// Create another watchlist
	watchlist2 := &TestWatchlist{
		UserID:      user.ID,
		Name:        "Test Watchlist 2",
		Description: "Test Description 2",
		Color:       "#3366FF",
	}
	result = db.Create(watchlist2)
	require.NoError(t, result.Error)
	
	// Create test channels
	channels := []*models.Channel{
		{YoutubeID: "UC_channel1", Title: "Channel 1"},
		{YoutubeID: "UC_channel2", Title: "Channel 2"},
		{YoutubeID: "UC_channel3", Title: "Channel 3"},
	}
	
	for _, ch := range channels {
		result = db.Create(ch)
		require.NoError(t, result.Error)
	}
	
	// Add channels to watchlists
	err := db.Model(watchlist).Association("Channels").Append([]*models.Channel{channels[0], channels[1]})
	require.NoError(t, err)
	
	err = db.Model(watchlist2).Association("Channels").Append([]*models.Channel{channels[2]})
	require.NoError(t, err)

	tests := []struct {
		name        string
		watchlistID uint
		userID      uint
		want        int
		wantErr     error
	}{
		{
			name:        "Get channels from watchlist",
			watchlistID: watchlist.ID,
			userID:      user.ID,
			want:        2,
			wantErr:     nil,
		},
		{
			name:        "Get channels from watchlist 2",
			watchlistID: watchlist2.ID,
			userID:      user.ID,
			want:        1,
			wantErr:     nil,
		},
		{
			name:        "Non-existent watchlist",
			watchlistID: 999,
			userID:      user.ID,
			want:        0,
			wantErr:     services.ErrWatchlistNotFound,
		},
		{
			name:        "Unauthorized user",
			watchlistID: watchlist.ID,
			userID:      999,
			want:        0,
			wantErr:     services.ErrWatchlistNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channels, err := watchlistService.GetChannelsInWatchlist(tt.watchlistID, tt.userID)
			
			if tt.wantErr != nil {
				assert.True(t, errors.Is(err, tt.wantErr))
				return
			}
			
			assert.NoError(t, err)
			assert.Len(t, channels, tt.want)
			
			if tt.watchlistID == watchlist.ID {
				assert.Equal(t, "UC_channel1", channels[0].YoutubeID)
				assert.Equal(t, "UC_channel2", channels[1].YoutubeID)
			} else if tt.watchlistID == watchlist2.ID {
				assert.Equal(t, "UC_channel3", channels[0].YoutubeID)
			}
		})
	}
}
