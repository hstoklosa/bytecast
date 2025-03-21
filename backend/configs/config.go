package configs

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
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

// Load returns a validated configuration struct
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

    if cfg.JWT.Secret == "" {
        log.Printf("Warning: JWT_SECRET not set")
    }
    
    if cfg.YouTube.APIKey == "" {
        log.Printf("Warning: YOUTUBE_API_KEY not set, YouTube API features will be disabled")
    }

    if err := validateConfig(cfg); err != nil {
        return nil, fmt.Errorf("config validation error: %w", err)
    }

    log.Printf("Configuration loaded successfully")
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
    return nil
}
