package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"bytecast/api/handler"
	"bytecast/api/middleware"
	"bytecast/configs"
	"bytecast/internal/models"
	"bytecast/internal/services"
)

type watchlistTestServer struct {
	db               *gorm.DB
	engine           *gin.Engine
	authService      *services.AuthService
	watchlistService *services.WatchlistService
	testUserID       uint
}

func setupWatchlistTestServer(t *testing.T) *watchlistTestServer {
	dsn := "host=localhost user=postgres password=bytecast dbname=bytecast_test port=5433 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Clean up existing tables
	db.Exec("DROP TABLE IF EXISTS watchlist_channels, channels, watchlists, revoked_tokens, users CASCADE")

	// Run migrations for all models
	if err := db.AutoMigrate(&models.User{}, &models.RevokedToken{}, &models.Channel{}, &models.Watchlist{}); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create test user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	testUser := models.User{
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: string(hashedPassword),
	}
	result := db.Create(&testUser)
	if err := result.Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	gin.SetMode(gin.TestMode)
	engine := gin.New()

	cfg := &configs.Config{
		Server: configs.Server{
			Port:        "8080",
			Environment: "development",
			Domain:      "localhost",
		},
		JWT: configs.JWT{
			Secret: "test-secret",
		},
	}

	// Initialize services
	authService := services.NewAuthService(db, cfg.JWT.Secret)
	watchlistService := services.NewWatchlistService(db)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, cfg)
	watchlistHandler := handler.NewWatchlistHandler(watchlistService)

	// Register routes
	authHandler.RegisterRoutes(engine)
	authMiddleware := middleware.AuthMiddleware([]byte(cfg.JWT.Secret))
	watchlistHandler.RegisterRoutes(engine, authMiddleware)

	return &watchlistTestServer{
		db:               db,
		engine:           engine,
		authService:      authService,
		watchlistService: watchlistService,
		testUserID:       testUser.ID,
	}
}

func getAuthTokenForTest(t *testing.T, server *watchlistTestServer) string {
	// Login to get token
	loginBody := map[string]string{
		"identifier": "testuser",
		"password":   "password123",
	}
	loginJSON, err := json.Marshal(loginBody)
	if err != nil {
		t.Fatalf("Failed to marshal login request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(loginJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var loginResp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("Failed to unmarshal login response: %v", err)
	}

	token, exists := loginResp["access_token"]
	if !exists {
		t.Fatal("Access token not found in login response")
	}

	return token
}

func TestCreateWatchlist(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	server := setupWatchlistTestServer(t)
	token := getAuthTokenForTest(t, server)

	// Test creating a watchlist
	watchlistBody := map[string]string{
		"name":        "My Test Watchlist",
		"description": "A watchlist for testing",
	}
	watchlistJSON, err := json.Marshal(watchlistBody)
	if err != nil {
		t.Fatalf("Failed to marshal watchlist request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/watchlists", bytes.NewBuffer(watchlistJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	w := httptest.NewRecorder()
	server.engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var watchlistResp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &watchlistResp); err != nil {
		t.Fatalf("Failed to unmarshal watchlist response: %v", err)
	}

	assert.Equal(t, "My Test Watchlist", watchlistResp["name"])
	assert.Equal(t, "A watchlist for testing", watchlistResp["description"])
	assert.NotNil(t, watchlistResp["id"])
}

func TestGetWatchlists(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	server := setupWatchlistTestServer(t)
	token := getAuthTokenForTest(t, server)

	// Create a watchlist directly through the service
	watchlist, err := server.watchlistService.CreateWatchlist(server.testUserID, "Test Watchlist", "Test Description")
	if err != nil {
		t.Fatalf("Failed to create test watchlist: %v", err)
	}

	// Test getting watchlists
	req := httptest.NewRequest(http.MethodGet, "/api/v1/watchlists", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	w := httptest.NewRecorder()
	server.engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var watchlistsResp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &watchlistsResp); err != nil {
		t.Fatalf("Failed to unmarshal watchlists response: %v", err)
	}
	
	watchlists, ok := watchlistsResp["watchlists"].([]interface{})
	assert.True(t, ok)
	assert.GreaterOrEqual(t, len(watchlists), 1)
	
	firstWatchlist := watchlists[0].(map[string]interface{})
	assert.Equal(t, float64(watchlist.ID), firstWatchlist["id"])
	assert.Equal(t, watchlist.Name, firstWatchlist["name"])
}

func TestAddChannelToWatchlist(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	server := setupWatchlistTestServer(t)
	token := getAuthTokenForTest(t, server)

	// Create a watchlist directly through the service
	watchlist, err := server.watchlistService.CreateWatchlist(server.testUserID, "Test Watchlist", "Test Description")
	if err != nil {
		t.Fatalf("Failed to create test watchlist: %v", err)
	}

	// Test adding a channel to the watchlist
	channelBody := map[string]string{
		"channel_id": "UC_x5XG1OV2P6uZZ5FSM9Ttw", // Google Developers channel ID
	}
	channelJSON, err := json.Marshal(channelBody)
	if err != nil {
		t.Fatalf("Failed to marshal channel request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/watchlists/%d/channels", watchlist.ID), bytes.NewBuffer(channelJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	w := httptest.NewRecorder()
	server.engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the channel was added by getting the channels in the watchlist
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/watchlists/%d/channels", watchlist.ID), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	w = httptest.NewRecorder()
	server.engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var channelsResp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &channelsResp); err != nil {
		t.Fatalf("Failed to unmarshal channels response: %v", err)
	}
	
	channels, ok := channelsResp["channels"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 1, len(channels))
	
	firstChannel := channels[0].(map[string]interface{})
	assert.Equal(t, "UC_x5XG1OV2P6uZZ5FSM9Ttw", firstChannel["youtube_id"])
}
