package configs

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"

	apperrors "bytecast/internal/errors"
)

type Config struct {
    Database  Database  `validate:"required"`
    JWT       JWT       `validate:"required"`
    Server    Server    `validate:"required"`
    Superuser Superuser `validate:"required"`
    YouTube   YouTube
}

type Superuser struct {
    Username string `validate:"required"`
    Email    string `validate:"required,email"`
    Password string `validate:"required,min=8"`
}

type Database struct {
    Host     string `validate:"required"`
    User     string `validate:"required"`
    Password string `validate:"required"`
    Name     string `validate:"required"`
    Port     string `validate:"required,numeric"`
}

type JWT struct {
    Secret string `validate:"required,min=32"`
}

type Server struct {
    Port        string `validate:"required,numeric"`
    Environment string `validate:"required,oneof=development production"`
    Domain      string `validate:"required"`
}

type YouTube struct {
   APIKey       string
    CallbackURL  string
    LeaseSeconds int
}

func Load() (*Config, error) {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("Note: .env file not found, using environment variables")
	}

    // Map environment variables to match docker-compose naming
    cfg := &Config{
        Superuser: Superuser{
            Username: getEnvWithDefault("SUPERUSER_USERNAME", "admin"),
            Email:    getEnvWithDefault("SUPERUSER_EMAIL", "admin@example.com"),
            Password: getEnvWithDefault("SUPERUSER_PASSWORD", "password"),
        },
        Database: Database{
            Host:     getEnvWithDefault("DB_HOST", getEnvWithDefault("POSTGRES_HOST", "localhost")),
            User:     getEnvWithDefault("DB_USER", getEnvWithDefault("POSTGRES_USER", "postgres")),
            Password: getEnvWithDefault("DB_PASSWORD", getEnvWithDefault("POSTGRES_PASSWORD", "bytecast")),
            Name:     getEnvWithDefault("DB_NAME", getEnvWithDefault("POSTGRES_DB", "bytecast")),
            Port:     getEnvWithDefault("DB_PORT", getEnvWithDefault("POSTGRES_PORT", "5432")),
        },
        JWT: JWT{
            Secret: getEnvWithDefault("JWT_SECRET", ""),
        },
        Server: Server{
            Port:        getEnvWithDefault("PORT", "8080"),
            Environment: getEnvWithDefault("APP_ENV", "development"),
            Domain:      getEnvWithDefault("APP_DOMAIN", "localhost"),
        },
        YouTube: YouTube{
            APIKey:       getEnvWithDefault("YOUTUBE_API_KEY", ""),
            CallbackURL:  getEnvWithDefault("YOUTUBE_WEBSUB_CALLBACK_URL", ""),
            LeaseSeconds: getEnvInt("YOUTUBE_WEBSUB_LEASE_SECONDS", 432000), // Default 5 days (max 10 days)
        },
    }

    if err := validateConfig(cfg); err != nil {
        return nil, fmt.Errorf("config validation error: %w", err)
    }

    logConfigStatus(cfg)
    return cfg, nil
}

// GetDSN returns the formatted database connection string
func (d *Database) GetDSN() string {
    return fmt.Sprintf(
        "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
        d.Host, d.User, d.Password, d.Name, d.Port,
    )
}

func getEnvWithDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        intValue, err := strconv.Atoi(value)
        if err == nil {
            return intValue
        }
    }
    return defaultValue
}

func validateConfig(cfg *Config) error {
    validate := validator.New()
    
    if err := validate.Struct(cfg); err != nil {
        return err
    }
    
    if len(cfg.JWT.Secret) < 32 {
        return apperrors.NewInvalidConfigError("JWT_SECRET", "must be at least 32 characters long", nil)
    }
    
    if cfg.YouTube.APIKey != "" {
        // api key is provided but callback URL is missing
        if cfg.YouTube.LeaseSeconds > 0 && cfg.YouTube.CallbackURL == "" {
            return apperrors.NewInvalidConfigError("YOUTUBE_WEBSUB_CALLBACK_URL", 
                "required when YOUTUBE_API_KEY and YOUTUBE_WEBSUB_LEASE_SECONDS are provided", nil)
        }
    }
    
    return nil
}

func logConfigStatus(cfg *Config) {
    log.Printf("Database connection: %s:%s", cfg.Database.Host, cfg.Database.Port)
    log.Printf("Server port: %s, Environment: %s", cfg.Server.Port, cfg.Server.Environment)
    
    youtubeAPIEnabled := cfg.YouTube.APIKey != ""
    pubSubEnabled := youtubeAPIEnabled && cfg.YouTube.CallbackURL != ""
    
    if youtubeAPIEnabled {
        log.Printf("YouTube API: Enabled")
    } else {
        log.Printf("Warning: YouTube API disabled (YOUTUBE_API_KEY not set)")
    }
    
    if pubSubEnabled {
        log.Printf("YouTube PubSub: Enabled (lease seconds: %d)", cfg.YouTube.LeaseSeconds)
    } else if youtubeAPIEnabled {
        log.Printf("Warning: YouTube PubSub disabled (YOUTUBE_WEBSUB_CALLBACK_URL not set)")
    }
}

type ConfigStatus struct {
    YouTubeAPIEnabled bool
    PubSubEnabled     bool
}

func (c *Config) GetConfigStatus() *ConfigStatus {
    status := &ConfigStatus{
        YouTubeAPIEnabled: c.YouTube.APIKey != "",
    }
    status.PubSubEnabled = status.YouTubeAPIEnabled && c.YouTube.CallbackURL != ""
    return status
}
