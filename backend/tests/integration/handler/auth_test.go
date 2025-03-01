package handler_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    "bytecast/api/handler"
    "bytecast/internal/models"
    "bytecast/internal/services"
)

type testServer struct {
    db          *gorm.DB
    engine      *gin.Engine
    authHandler *handler.AuthHandler
}

func setupTestServer(t *testing.T) *testServer {
    // Use test database
    dsn := "host=localhost user=postgres password=bytecast dbname=bytecast_test port=5433 sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        t.Fatalf("Failed to connect to test database: %v", err)
    }

    // Run migrations
    // Drop existing tables if they exist
    db.Exec("DROP TABLE IF EXISTS revoked_tokens, users CASCADE")

    // Run migrations
    if err := db.AutoMigrate(&models.User{}, &models.RevokedToken{}); err != nil {
        t.Fatalf("Failed to run migrations: %v", err)
    }

    // Setup Gin
    gin.SetMode(gin.TestMode)
    engine := gin.New()

    // Setup auth handler
    authService := services.NewAuthService(db, "test-secret")
    authHandler := handler.NewAuthHandler(authService)
    authHandler.RegisterRoutes(engine)

    return &testServer{
        db:          db,
        engine:      engine,
        authHandler: authHandler,
    }
}

func TestAuthHandler_Logout(t *testing.T) {
    server := setupTestServer(t)

    // Register and login a test user first
    registerData := map[string]string{
        "email":    "test@example.com",
        "password": "password123",
    }

    registerBody, _ := json.Marshal(registerData)
    req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(registerBody))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    server.engine.ServeHTTP(w, req)
    assert.Equal(t, http.StatusCreated, w.Code)

    // Login to get tokens
    loginBody, _ := json.Marshal(registerData)
    req = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(loginBody))
    req.Header.Set("Content-Type", "application/json")
    w = httptest.NewRecorder()
    server.engine.ServeHTTP(w, req)
    assert.Equal(t, http.StatusOK, w.Code)

    var loginResponse struct {
        RefreshToken string `json:"refresh_token"`
    }
    err := json.Unmarshal(w.Body.Bytes(), &loginResponse)
    assert.NoError(t, err)

    tests := []struct {
        name       string
        token      string
        wantStatus int
    }{
        {
            name:       "Valid token logout",
            token:      loginResponse.RefreshToken,
            wantStatus: http.StatusNoContent,
        },
        {
            name:       "Invalid token logout",
            token:      "invalid-token",
            wantStatus: http.StatusUnauthorized,
        },
        {
            name:       "Already revoked token",
            token:      loginResponse.RefreshToken,
            wantStatus: http.StatusUnauthorized,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            logoutData := map[string]string{
                "refresh_token": tt.token,
            }
            logoutBody, _ := json.Marshal(logoutData)

            req := httptest.NewRequest(http.MethodPost, "/auth/logout", bytes.NewBuffer(logoutBody))
            req.Header.Set("Content-Type", "application/json")
            w := httptest.NewRecorder()
            server.engine.ServeHTTP(w, req)

            assert.Equal(t, tt.wantStatus, w.Code)
        })
    }
}

func TestAuthHandler_RefreshWithRevokedToken(t *testing.T) {
    server := setupTestServer(t)

    // Register and login a test user
    userData := map[string]string{
        "email":    "test@example.com",
        "password": "password123",
    }

    // Register
    registerBody, _ := json.Marshal(userData)
    req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(registerBody))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    server.engine.ServeHTTP(w, req)
    assert.Equal(t, http.StatusCreated, w.Code)

    // Login to get tokens
    loginBody, _ := json.Marshal(userData)
    req = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(loginBody))
    req.Header.Set("Content-Type", "application/json")
    w = httptest.NewRecorder()
    server.engine.ServeHTTP(w, req)
    assert.Equal(t, http.StatusOK, w.Code)

    var loginResponse struct {
        RefreshToken string `json:"refresh_token"`
    }
    err := json.Unmarshal(w.Body.Bytes(), &loginResponse)
    assert.NoError(t, err)

    // Logout to revoke the token
    logoutData := map[string]string{
        "refresh_token": loginResponse.RefreshToken,
    }
    logoutBody, _ := json.Marshal(logoutData)
    req = httptest.NewRequest(http.MethodPost, "/auth/logout", bytes.NewBuffer(logoutBody))
    req.Header.Set("Content-Type", "application/json")
    w = httptest.NewRecorder()
    server.engine.ServeHTTP(w, req)
    assert.Equal(t, http.StatusNoContent, w.Code)

    // Attempt to refresh with revoked token
    refreshData := map[string]string{
        "refresh_token": loginResponse.RefreshToken,
    }
    refreshBody, _ := json.Marshal(refreshData)
    req = httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(refreshBody))
    req.Header.Set("Content-Type", "application/json")
    w = httptest.NewRecorder()
    server.engine.ServeHTTP(w, req)

    // Should fail with unauthorized
    assert.Equal(t, http.StatusUnauthorized, w.Code)
}
