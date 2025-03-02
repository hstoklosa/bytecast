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
    ErrUserExists        = errors.New("user already exists")
    ErrUsernameTaken     = errors.New("username already taken")
    ErrTokenInvalid      = errors.New("invalid token")
    ErrTokenRevoked      = errors.New("token has been revoked")
)

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthService struct {
	db          *gorm.DB
	jwtSecret   []byte
	accessExp   time.Duration
	refreshExp  time.Duration
}

func NewAuthService(db *gorm.DB, jwtSecret string) *AuthService {
	return &AuthService{
		db:         db,
		jwtSecret:  []byte(jwtSecret),
		accessExp:  15 * time.Minute,  // 15 minutes
		refreshExp: 7 * 24 * time.Hour, // 7 days
	}
}

func (s *AuthService) RegisterUser(email, username, password string) error {
    // Check if email exists
    var existingUser models.User
    if err := s.db.Where("email = ?", email).First(&existingUser).Error; err == nil {
        return ErrUserExists
    } else if !errors.Is(err, gorm.ErrRecordNotFound) {
        return err
    }

    // Check if username exists
    if err := s.db.Where("username = ?", username).First(&existingUser).Error; err == nil {
        return ErrUsernameTaken
    } else if !errors.Is(err, gorm.ErrRecordNotFound) {
        return err
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }

    user := models.User{
        Email:        email,
        Username:     username,
        PasswordHash: string(hashedPassword),
    }

    return s.db.Create(&user).Error
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

    // Check if token is revoked
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
    // Generate Access Token
    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(s.accessExp).Unix(),
        "type":    "access",
    })

    accessTokenString, err := accessToken.SignedString(s.jwtSecret)
    if err != nil {
        return nil, time.Time{}, err
    }

    // Generate Refresh Token
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
