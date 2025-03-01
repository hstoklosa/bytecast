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
    "bytecast/configs"
    "bytecast/internal/models"
    "bytecast/internal/services"
)

type testServer struct {
    db          *gorm.DB
    engine      *gin.Engine
    authHandler *handler.AuthHandler
}

func setupTestServer(t *testing.T) *testServer {
    dsn := "host=localhost user=postgres password=bytecast dbname=bytecast_test port=5433 sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        t.Fatalf("Failed to connect to test database: %v", err)
    }

    db.Exec("DROP TABLE IF EXISTS revoked_tokens, users CASCADE")

    if err := db.AutoMigrate(&models.User{}, &models.RevokedToken{}); err != nil {
        t.Fatalf("Failed to run migrations: %v", err)
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

    authService := services.NewAuthService(db, cfg.JWT.Secret)
    authHandler := handler.NewAuthHandler(authService, cfg)
    authHandler.RegisterRoutes(engine)

    return &testServer{
        db:          db,
        engine:      engine,
        authHandler: authHandler,
    }
}

func TestAuthHandler_Register(t *testing.T) {
    server := setupTestServer(t)

    tests := []struct {
        name       string
        input     map[string]string
        wantStatus int
    }{
        {
            name: "Valid registration",
            input: map[string]string{
                "email":    "test@example.com",
                "username": "testuser",
                "password": "password123",
            },
            wantStatus: http.StatusCreated,
        },
        {
            name: "Duplicate email",
            input: map[string]string{
                "email":    "test@example.com",
                "username": "testuser2",
                "password": "password123",
            },
            wantStatus: http.StatusConflict,
        },
        {
            name: "Duplicate username",
            input: map[string]string{
                "email":    "test2@example.com",
                "username": "testuser",
                "password": "password123",
            },
            wantStatus: http.StatusConflict,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            body, _ := json.Marshal(tt.input)
            req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
            req.Header.Set("Content-Type", "application/json")
            w := httptest.NewRecorder()
            server.engine.ServeHTTP(w, req)
            assert.Equal(t, tt.wantStatus, w.Code)
        })
    }
}

func TestAuthHandler_Logout(t *testing.T) {
    server := setupTestServer(t)

    // Register and login a test user first
    registerData := map[string]string{
        "email":    "test@example.com",
        "username": "testuser",
        "password": "password123",
    }

    registerBody, _ := json.Marshal(registerData)
    req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(registerBody))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    server.engine.ServeHTTP(w, req)
    assert.Equal(t, http.StatusCreated, w.Code)

    // Login to get tokens and cookie
    loginBody, _ := json.Marshal(registerData)
    req = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(loginBody))
    req.Header.Set("Content-Type", "application/json")
    w = httptest.NewRecorder()
    server.engine.ServeHTTP(w, req)
    assert.Equal(t, http.StatusOK, w.Code)

    // Get refresh token from cookie
    cookies := w.Result().Cookies()
    var refreshCookie *http.Cookie
    for _, cookie := range cookies {
        if cookie.Name == "refresh_token" {
            refreshCookie = cookie
            break
        }
    }
    assert.NotNil(t, refreshCookie, "Refresh token cookie not found")

    tests := []struct {
        name       string
        cookie     *http.Cookie
        wantStatus int
    }{
        {
            name:       "Valid token logout",
            cookie:     refreshCookie,
            wantStatus: http.StatusNoContent,
        },
        {
            name:       "Invalid token logout",
            cookie:     &http.Cookie{Name: "refresh_token", Value: "invalid-token"},
            wantStatus: http.StatusUnauthorized,
        },
        {
            name:       "Already revoked token",
            cookie:     refreshCookie,
            wantStatus: http.StatusUnauthorized,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
            req.Header.Set("Content-Type", "application/json")
            req.AddCookie(tt.cookie)
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
        "username": "testuser",
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

    // Get refresh token from cookie
    cookies := w.Result().Cookies()
    var refreshCookie *http.Cookie
    for _, cookie := range cookies {
        if cookie.Name == "refresh_token" {
            refreshCookie = cookie
            break
        }
    }
    assert.NotNil(t, refreshCookie, "Refresh token cookie not found")

    // Logout to revoke the token
    req = httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
    req.Header.Set("Content-Type", "application/json")
    req.AddCookie(refreshCookie)
    w = httptest.NewRecorder()
    server.engine.ServeHTTP(w, req)
    assert.Equal(t, http.StatusNoContent, w.Code)

    // Attempt to refresh with revoked token
    req = httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
    req.Header.Set("Content-Type", "application/json")
    req.AddCookie(refreshCookie)
    w = httptest.NewRecorder()
    server.engine.ServeHTTP(w, req)

    // Should fail with unauthorized
    assert.Equal(t, http.StatusUnauthorized, w.Code)

    // Verify proper cookie clearing
    cookies = w.Result().Cookies()
    var clearedCookie *http.Cookie
    for _, cookie := range cookies {
        if cookie.Name == "refresh_token" {
            clearedCookie = cookie
            break
        }
    }
    assert.NotNil(t, clearedCookie, "Cookie clearing not found")
    assert.Equal(t, "", clearedCookie.Value, "Cookie not properly cleared")
    assert.True(t, clearedCookie.MaxAge < 0, "Cookie expiration not properly set")
}
