package services_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"bytecast/internal/models"
	"bytecast/internal/services"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations
	err = db.AutoMigrate(&models.User{}, &models.Watchlist{}, &models.Channel{})
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

func TestCreateWatchlist(t *testing.T) {
	db := setupTestDB(t)
	watchlistService := services.NewWatchlistService(db)
	user := createTestUser(t, db)

	tests := []struct {
		name        string
		userID      uint
		wlName      string
		description string
		wantErr     bool
	}{
		{
			name:        "Valid watchlist",
			userID:      user.ID,
			wlName:      "My Watchlist",
			description: "Test description",
			wantErr:     false,
		},
		{
			name:        "Empty name",
			userID:      user.ID,
			wlName:      "",
			description: "Test description",
			wantErr:     true,
		},
		{
			name:        "Non-existent user",
			userID:      999,
			wlName:      "My Watchlist",
			description: "Test description",
			wantErr:     false, // GORM doesn't validate foreign keys by default in SQLite
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watchlist, err := watchlistService.CreateWatchlist(tt.userID, tt.wlName, tt.description)
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			
			assert.NoError(t, err)
			assert.NotNil(t, watchlist)
			assert.Equal(t, tt.userID, watchlist.UserID)
			assert.Equal(t, tt.wlName, watchlist.Name)
			assert.Equal(t, tt.description, watchlist.Description)
			assert.NotZero(t, watchlist.ID)
			assert.NotZero(t, watchlist.CreatedAt)
		})
	}
}

func TestGetWatchlist(t *testing.T) {
	db := setupTestDB(t)
	watchlistService := services.NewWatchlistService(db)
	user := createTestUser(t, db)
	
	// Create a test watchlist
	watchlist := &models.Watchlist{
		UserID:      user.ID,
		Name:        "Test Watchlist",
		Description: "Test Description",
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
	db := setupTestDB(t)
	watchlistService := services.NewWatchlistService(db)
	user1 := createTestUser(t, db)
	
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
	watchlists := []models.Watchlist{
		{UserID: user1.ID, Name: "Watchlist 1", Description: "Description 1"},
		{UserID: user1.ID, Name: "Watchlist 2", Description: "Description 2"},
	}
	for i := range watchlists {
		result := db.Create(&watchlists[i])
		require.NoError(t, result.Error)
		require.NotZero(t, watchlists[i].ID)
	}
	
	// Create a watchlist for user2
	watchlist := models.Watchlist{UserID: user2.ID, Name: "User2 Watchlist", Description: "User2 Description"}
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
			userID:  user1.ID,
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
			
			if tt.userID == user1.ID {
				assert.Equal(t, watchlists[0].Name, result[0].Name)
				assert.Equal(t, watchlists[1].Name, result[1].Name)
			} else if tt.userID == user2.ID {
				assert.Equal(t, watchlist.Name, result[0].Name)
			}
		})
	}
}

func TestUpdateWatchlist(t *testing.T) {
	db := setupTestDB(t)
	watchlistService := services.NewWatchlistService(db)
	user := createTestUser(t, db)
	
	// Create a test watchlist
	watchlist := &models.Watchlist{
		UserID:      user.ID,
		Name:        "Test Watchlist",
		Description: "Test Description",
	}
	result := db.Create(watchlist)
	require.NoError(t, result.Error)
	require.NotZero(t, watchlist.ID)

	tests := []struct {
		name        string
		watchlistID uint
		userID      uint
		newName     string
		newDesc     string
		wantErr     error
	}{
		{
			name:        "Valid update",
			watchlistID: watchlist.ID,
			userID:      user.ID,
			newName:     "Updated Watchlist",
			newDesc:     "Updated Description",
			wantErr:     nil,
		},
		{
			name:        "Non-existent watchlist",
			watchlistID: 999,
			userID:      user.ID,
			newName:     "Updated Watchlist",
			newDesc:     "Updated Description",
			wantErr:     services.ErrWatchlistNotFound,
		},
		{
			name:        "Unauthorized user",
			watchlistID: watchlist.ID,
			userID:      999,
			newName:     "Updated Watchlist",
			newDesc:     "Updated Description",
			wantErr:     services.ErrWatchlistNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, err := watchlistService.UpdateWatchlist(tt.watchlistID, tt.userID, tt.newName, tt.newDesc)
			
			if tt.wantErr != nil {
				assert.True(t, errors.Is(err, tt.wantErr))
				return
			}
			
			assert.NoError(t, err)
			assert.NotNil(t, updated)
			assert.Equal(t, tt.newName, updated.Name)
			assert.Equal(t, tt.newDesc, updated.Description)
			
			// Verify in database
			var dbWatchlist models.Watchlist
			err = db.First(&dbWatchlist, tt.watchlistID).Error
			assert.NoError(t, err)
			assert.Equal(t, tt.newName, dbWatchlist.Name)
			assert.Equal(t, tt.newDesc, dbWatchlist.Description)
		})
	}
}

func TestDeleteWatchlist(t *testing.T) {
	db := setupTestDB(t)
	watchlistService := services.NewWatchlistService(db)
	user := createTestUser(t, db)
	
	// Create a test watchlist
	watchlist := &models.Watchlist{
		UserID:      user.ID,
		Name:        "Test Watchlist",
		Description: "Test Description",
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
			db.Model(&models.Watchlist{}).Where("id = ?", tt.watchlistID).Count(&count)
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
			input: "https://www.youtube.com/user/GoogleDevelopers",
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
			
			db := setupTestDB(t)
			watchlistService := services.NewWatchlistService(db)
			
			// Create a test user and watchlist
			user := createTestUser(t, db)
			watchlist := &models.Watchlist{
				UserID:      user.ID,
				Name:        "Test Watchlist",
				Description: "Test Description",
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
					err = db.Where("you_tube_id = ?", tt.want).First(&channel).Error
					assert.NoError(t, err)
					assert.Equal(t, tt.want, channel.YouTubeID)
				}
			}
		})
	}
}

func TestAddChannelToWatchlist(t *testing.T) {
	db := setupTestDB(t)
	watchlistService := services.NewWatchlistService(db)
	user := createTestUser(t, db)
	
	// Create a test watchlist
	watchlist := &models.Watchlist{
		UserID:      user.ID,
		Name:        "Test Watchlist",
		Description: "Test Description",
	}
	result := db.Create(watchlist)
	require.NoError(t, result.Error)
	
	// Create a test channel
	channel := &models.Channel{
		YouTubeID: "UC_existing",
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
			var wl models.Watchlist
			wl.ID = tt.watchlistID
			count = db.Model(&wl).Association("Channels").Count()
			assert.Greater(t, count, int64(0))
			
			// var channels []models.Channel
			// err = db.Model(&models.Watchlist{ID: tt.watchlistID}).Association("Channels").Find(&channels)
			// assert.NoError(t, err)

			assert.NoError(t, err)
			assert.Greater(t, count, int64(0))
			
			// Verify channel exists in database
			var channel models.Channel
			err = db.Where("you_tube_id = ?", tt.channelID).First(&channel).Error
			assert.NoError(t, err)
		})
	}
}

func TestRemoveChannelFromWatchlist(t *testing.T) {
	db := setupTestDB(t)
	watchlistService := services.NewWatchlistService(db)
	user := createTestUser(t, db)
	
	// Create a test watchlist
	watchlist := &models.Watchlist{
		UserID:      user.ID,
		Name:        "Test Watchlist",
		Description: "Test Description",
	}
	result := db.Create(watchlist)
	require.NoError(t, result.Error)
	
	// Create test channels
	channel1 := &models.Channel{
		YouTubeID: "UC_channel1",
		Title:     "Channel 1",
	}
	result = db.Create(channel1)
	require.NoError(t, result.Error)
	
	channel2 := &models.Channel{
		YouTubeID: "UC_channel2",
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
			var wl models.Watchlist
			wl.ID = tt.watchlistID
			err = db.Model(&wl).Association("Channels").Find(&channels)
			assert.NoError(t, err)
			for _, ch := range channels {
				assert.NotEqual(t, tt.channelID, ch.YouTubeID)
			}
		})
	}
}

func TestGetChannelsInWatchlist(t *testing.T) {
	db := setupTestDB(t)
	watchlistService := services.NewWatchlistService(db)
	user := createTestUser(t, db)
	
	// Create a test watchlist
	watchlist := &models.Watchlist{
		UserID:      user.ID,
		Name:        "Test Watchlist",
		Description: "Test Description",
	}
	result := db.Create(watchlist)
	require.NoError(t, result.Error)
	
	// Create another watchlist
	watchlist2 := &models.Watchlist{
		UserID:      user.ID,
		Name:        "Test Watchlist 2",
		Description: "Test Description 2",
	}
	result = db.Create(watchlist2)
	require.NoError(t, result.Error)
	
	// Create test channels
	channels := []*models.Channel{
		{YouTubeID: "UC_channel1", Title: "Channel 1"},
		{YouTubeID: "UC_channel2", Title: "Channel 2"},
		{YouTubeID: "UC_channel3", Title: "Channel 3"},
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
				assert.Equal(t, "UC_channel1", channels[0].YouTubeID)
				assert.Equal(t, "UC_channel2", channels[1].YouTubeID)
			} else if tt.watchlistID == watchlist2.ID {
				assert.Equal(t, "UC_channel3", channels[0].YouTubeID)
			}
		})
	}
}
