package services

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"bytecast/internal/models"
)

var (
    ErrInvalidCredentials = errors.New("invalid email or password")
    ErrUserExists         = errors.New("user already exists")
    ErrUsernameTaken      = errors.New("username already taken")
    ErrTokenInvalid       = errors.New("invalid token")
    ErrTokenRevoked       = errors.New("token has been revoked")
)

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthService struct {
    db            *gorm.DB
    watchlistSvc  *WatchlistService
    jwtSecret     []byte
    accessExp     time.Duration
    refreshExp    time.Duration
}

func NewAuthService(db *gorm.DB, watchlistSvc *WatchlistService, jwtSecret string) *AuthService {
    return &AuthService{
        db:           db,
        watchlistSvc: watchlistSvc,
        jwtSecret:    []byte(jwtSecret),
        accessExp:    15 * time.Minute,   // 15 minutes
        refreshExp:   7 * 24 * time.Hour, // 7 days
    }
}

func (s *AuthService) RegisterUser(email, username, password string) error {
    tx := s.db.Begin()
    if tx.Error != nil {
        return tx.Error
    }
    
    // Defer rollback in case of error
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    // Check if email exists
    var existingUser models.User
    if err := tx.Where("email = ?", email).First(&existingUser).Error; err == nil {
        tx.Rollback()
        return ErrUserExists
    } else if !errors.Is(err, gorm.ErrRecordNotFound) {
        tx.Rollback()
        return err
    }

    // Check if username exists
    if err := tx.Where("username = ?", username).First(&existingUser).Error; err == nil {
        tx.Rollback()
        return ErrUsernameTaken
    } else if !errors.Is(err, gorm.ErrRecordNotFound) {
        tx.Rollback()
        return err
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        tx.Rollback()
        return err
    }

    user := models.User{
        Email:        email,
        Username:     username,
        PasswordHash: string(hashedPassword),
    }

    if err := tx.Create(&user).Error; err != nil {
        tx.Rollback()
        return err
    }

    if err := s.watchlistSvc.CreateDefaultWatchlist(user.ID); err != nil {
        tx.Rollback()
        return err
    }

    return tx.Commit().Error
}

func (s *AuthService) FindByIdentifier(identifier string) (*models.User, error) {
    var user models.User
    err := s.db.Where("email = ? OR username = ?", identifier, identifier).First(&user).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrInvalidCredentials
        }
        return nil, err
    }
    return &user, nil
}

func (s *AuthService) LoginUser(identifier, password string) (*TokenPair, time.Time, error) {
    user, err := s.FindByIdentifier(identifier)
    if err != nil {
        return nil, time.Time{}, err
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
        return nil, time.Time{}, ErrInvalidCredentials
    }

    return s.generateTokenPair(user.ID)
}

func (s *AuthService) RefreshTokens(refreshToken string) (*TokenPair, time.Time, error) {
    token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, ErrTokenInvalid
        }
        return s.jwtSecret, nil
    })

    if err != nil || !token.Valid {
        return nil, time.Time{}, ErrTokenInvalid
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return nil, time.Time{}, ErrTokenInvalid
    }

    isRevoked, err := s.IsTokenRevoked(refreshToken)
    if err != nil {
        return nil, time.Time{}, err
    }
    if isRevoked {
        return nil, time.Time{}, ErrTokenRevoked
    }

    userID, ok := claims["user_id"].(float64)
    if !ok {
        return nil, time.Time{}, ErrTokenInvalid
    }

    return s.generateTokenPair(uint(userID))
}

func (s *AuthService) RevokeToken(token string) error {
    parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, ErrTokenInvalid
        }
        return s.jwtSecret, nil
    })

    if err != nil || !parsedToken.Valid {
        return ErrTokenInvalid
    }

    claims, ok := parsedToken.Claims.(jwt.MapClaims)
    if !ok {
        return ErrTokenInvalid
    }

    userID, ok := claims["user_id"].(float64)
    if !ok {
        return ErrTokenInvalid
    }

    exp, ok := claims["exp"].(float64)
    if !ok {
        return ErrTokenInvalid
    }

    revokedToken := models.RevokedToken{
        TokenHash: s.hashToken(token),
        ExpiresAt: time.Unix(int64(exp), 0),
        UserID:    uint(userID),
    }

    return s.db.Create(&revokedToken).Error
}

func (s *AuthService) IsTokenRevoked(token string) (bool, error) {
    var revokedToken models.RevokedToken
    result := s.db.Where("token_hash = ?", s.hashToken(token)).First(&revokedToken)

    if result.Error == nil {
        return true, nil
    }

    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        return false, nil
    }

    return false, result.Error
}

func (s *AuthService) hashToken(token string) string {
    hash := sha256.Sum256([]byte(token))
    return hex.EncodeToString(hash[:])
}

func (s *AuthService) generateTokenPair(userID uint) (*TokenPair, time.Time, error) {
    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(s.accessExp).Unix(),
        "type":    "access",
    })

    accessTokenString, err := accessToken.SignedString(s.jwtSecret)
    if err != nil {
        return nil, time.Time{}, err
    }

    refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(s.refreshExp).Unix(),
        "type":    "refresh",
    })

    refreshTokenString, err := refreshToken.SignedString(s.jwtSecret)
    if err != nil {
        return nil, time.Time{}, err
    }

    exp := time.Now().Add(s.refreshExp)

    return &TokenPair{
        AccessToken:  accessTokenString,
        RefreshToken: refreshTokenString,
    }, exp, nil
}

func (s *AuthService) GetUserByID(id uint) (*models.User, error) {
    var user models.User
    if err := s.db.First(&user, id).Error; err != nil {
        return nil, err
    }
    return &user, nil
}
