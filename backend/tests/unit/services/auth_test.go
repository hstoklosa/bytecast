package services_test

import (
    "testing"
    "time"

    "github.com/golang-jwt/jwt"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    "bytecast/internal/models"
    "bytecast/internal/services"
)

func setupTestDB(t *testing.T) *gorm.DB {
    dsn := "host=localhost user=postgres password=bytecast dbname=bytecast_test port=5433 sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        t.Fatalf("Failed to connect to test database: %v", err)
    }

    db.Exec("DROP TABLE IF EXISTS revoked_tokens, users CASCADE")

    if err := db.AutoMigrate(&models.User{}, &models.RevokedToken{}); err != nil {
        t.Fatalf("Failed to run migrations: %v", err)
    }

    return db
}

func TestAuthService_RegisterUser(t *testing.T) {
    db := setupTestDB(t)
    authService := services.NewAuthService(db, "test-secret")

    tests := []struct {
        name     string
        email    string
        username string
        password string
        wantErr  bool
        errType  error
    }{
        {
            name:     "Valid registration",
            email:    "test@example.com",
            username: "testuser",
            password: "password123",
            wantErr:  false,
        },
        {
            name:     "Duplicate email",
            email:    "test@example.com",
            username: "testuser2",
            password: "password123",
            wantErr:  true,
            errType:  services.ErrUserExists,
        },
        {
            name:     "Duplicate username",
            email:    "test2@example.com",
            username: "testuser",
            password: "password123",
            wantErr:  true,
            errType:  services.ErrUsernameTaken,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := authService.RegisterUser(tt.email, tt.username, tt.password)
            if (err != nil) != tt.wantErr {
                t.Errorf("RegisterUser() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}

func TestAuthService_RevokeToken(t *testing.T) {
    db := setupTestDB(t)
    authService := services.NewAuthService(db, "test-secret")

    email := "test@example.com"
    password := "password123"
    if err := authService.RegisterUser(email, "testuser", password); err != nil {
        t.Fatalf("Failed to register test user: %v", err)
    }

    tokens, _, err := authService.LoginUser(email, password)
    if err != nil {
        t.Fatalf("Failed to login test user: %v", err)
    }

    tests := []struct {
        name         string
        refreshToken string
        wantErr      bool
    }{
        {
            name:         "Valid token revocation",
            refreshToken: tokens.RefreshToken,
            wantErr:     false,
        },
        {
            name:         "Invalid token revocation",
            refreshToken: "invalid-token",
            wantErr:     true,
        },
        {
            name:         "Already revoked token",
            refreshToken: tokens.RefreshToken,
            wantErr:     true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := authService.RevokeToken(tt.refreshToken)
            if (err != nil) != tt.wantErr {
                t.Errorf("RevokeToken() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if !tt.wantErr {
                _, _, err := authService.RefreshTokens(tt.refreshToken)
                if err == nil {
                    t.Error("RefreshTokens() succeeded with revoked token")
                }
            }
        })
    }
}

func TestAuthService_RefreshTokens_WithRevokedToken(t *testing.T) {
    db := setupTestDB(t)
    authService := services.NewAuthService(db, "test-secret")

    email := "test@example.com"
    password := "password123"
    if err := authService.RegisterUser(email, "testuser", password); err != nil {
        t.Fatalf("Failed to register test user: %v", err)
    }

    tokens, _, err := authService.LoginUser(email, password)
    if err != nil {
        t.Fatalf("Failed to login test user: %v", err)
    }

    if err := authService.RevokeToken(tokens.RefreshToken); err != nil {
        t.Fatalf("Failed to revoke token: %v", err)
    }

    _, _, err = authService.RefreshTokens(tokens.RefreshToken)
    if err == nil {
        t.Error("RefreshTokens() succeeded with revoked token")
    }

    newTokens, _, err := authService.LoginUser(email, password)
    if err != nil {
        t.Fatalf("Failed to get new tokens: %v", err)
    }

    refreshedTokens, _, err := authService.RefreshTokens(newTokens.RefreshToken)
    if err != nil {
        t.Errorf("RefreshTokens() failed with new token: %v", err)
    }
    if refreshedTokens == nil {
        t.Error("RefreshTokens() returned nil tokens")
    }
}

func TestAuthService_LoginUser(t *testing.T) {
    db := setupTestDB(t)
    authService := services.NewAuthService(db, "test-secret")

    email := "test@example.com"
    password := "password123"
    if err := authService.RegisterUser(email, "testuser", password); err != nil {
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
            tokens, exp, err := authService.LoginUser(tt.email, tt.password)
            if (err != nil) != tt.wantErr {
                t.Errorf("LoginUser() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if !tt.wantErr {
                if tokens == nil {
                    t.Error("LoginUser() returned nil tokens")
                    return
                }

                if exp.Before(time.Now()) {
                    t.Error("Token expiration time is in the past")
                }

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
            }
        })
    }
}

func TestAuthService_RefreshTokens(t *testing.T) {
    db := setupTestDB(t)
    authService := services.NewAuthService(db, "test-secret")

    email := "test@example.com"
    password := "password123"
    if err := authService.RegisterUser(email, "testuser", password); err != nil {
        t.Fatalf("Failed to register test user: %v", err)
    }

    tokens, _, err := authService.LoginUser(email, password)
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
            newTokens, exp, err := authService.RefreshTokens(tt.refreshToken)
            if (err != nil) != tt.wantErr {
                t.Errorf("RefreshTokens() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if !tt.wantErr {
                if newTokens == nil {
                    t.Error("RefreshTokens() returned nil tokens")
                }

                if exp.Before(time.Now()) {
                    t.Error("Token expiration time is in the past")
                }
            }
        })
    }
}
