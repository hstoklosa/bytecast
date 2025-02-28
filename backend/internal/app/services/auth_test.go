package services

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"bytecast/internal/app/models"
)

func setupTestDB(t *testing.T) *gorm.DB {
	// Use test database
	dsn := "host=localhost user=postgres password=bytecast dbname=bytecast_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Clear test data
	db.Exec("TRUNCATE TABLE users RESTART IDENTITY")

	return db
}

func TestAuthService_RegisterUser(t *testing.T) {
	db := setupTestDB(t)
	authService := NewAuthService(db, "test-secret")

	tests := []struct {
		name     string
		email    string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid registration",
			email:    "test@example.com",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "Duplicate email",
			email:    "test@example.com",
			password: "password123",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authService.RegisterUser(tt.email, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthService_LoginUser(t *testing.T) {
	db := setupTestDB(t)
	authService := NewAuthService(db, "test-secret")

	// Register a test user
	email := "test@example.com"
	password := "password123"
	if err := authService.RegisterUser(email, password); err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	tests := []struct {
		name     string
		email    string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid credentials",
			email:    email,
			password: password,
			wantErr:  false,
		},
		{
			name:     "Invalid password",
			email:    email,
			password: "wrongpassword",
			wantErr:  true,
		},
		{
			name:     "Non-existent user",
			email:    "nonexistent@example.com",
			password: password,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := authService.LoginUser(tt.email, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoginUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tokens == nil {
					t.Error("LoginUser() returned nil tokens")
					return
				}

				// Verify access token
				token, err := jwt.Parse(tokens.AccessToken, func(token *jwt.Token) (interface{}, error) {
					return []byte("test-secret"), nil
				})

				if err != nil || !token.Valid {
					t.Errorf("Invalid access token: %v", err)
				}

				claims, ok := token.Claims.(jwt.MapClaims)
				if !ok {
					t.Error("Failed to parse token claims")
					return
				}

				if claims["type"] != "access" {
					t.Error("Invalid token type")
				}

				exp, ok := claims["exp"].(float64)
				if !ok {
					t.Error("Missing expiration time")
					return
				}

				if time.Unix(int64(exp), 0).Before(time.Now()) {
					t.Error("Token already expired")
				}
			}
		})
	}
}

func TestAuthService_RefreshTokens(t *testing.T) {
	db := setupTestDB(t)
	authService := NewAuthService(db, "test-secret")

	// Register and login a test user
	email := "test@example.com"
	password := "password123"
	if err := authService.RegisterUser(email, password); err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	tokens, err := authService.LoginUser(email, password)
	if err != nil {
		t.Fatalf("Failed to login test user: %v", err)
	}

	tests := []struct {
		name         string
		refreshToken string
		wantErr      bool
	}{
		{
			name:         "Valid refresh token",
			refreshToken: tokens.RefreshToken,
			wantErr:     false,
		},
		{
			name:         "Invalid refresh token",
			refreshToken: "invalid-token",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newTokens, err := authService.RefreshTokens(tt.refreshToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("RefreshTokens() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && newTokens == nil {
				t.Error("RefreshTokens() returned nil tokens")
			}
		})
	}
}
