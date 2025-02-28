package services

import (
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
	ErrTokenInvalid     = errors.New("invalid token")
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

func (s *AuthService) RegisterUser(email, password string) error {
	var existingUser models.User
	if err := s.db.Where("email = ?", email).First(&existingUser).Error; err == nil {
		return ErrUserExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := models.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	return s.db.Create(&user).Error
}

func (s *AuthService) LoginUser(email, password string) (*TokenPair, error) {
	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateTokenPair(user.ID)
}

func (s *AuthService) RefreshTokens(refreshToken string) (*TokenPair, error) {
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrTokenInvalid
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, ErrTokenInvalid
	}

	return s.generateTokenPair(uint(userID))
}

func (s *AuthService) generateTokenPair(userID uint) (*TokenPair, error) {
	// Generate Access Token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(s.accessExp).Unix(),
		"type":    "access",
	})

	accessTokenString, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	// Generate Refresh Token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(s.refreshExp).Unix(),
		"type":    "refresh",
	})

	refreshTokenString, err := refreshToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}
