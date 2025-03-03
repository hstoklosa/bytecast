package database

import (
    "fmt"
    "time"

    "golang.org/x/crypto/bcrypt"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    "bytecast/configs"
    "bytecast/internal/models"
)

// Connection holds the database connection and configuration
type Connection struct {
    db     *gorm.DB
    config *configs.Config
}

// New creates a new database connection with proper connection pooling
func New(cfg *configs.Config) (*Connection, error) {
	db, err := gorm.Open(postgres.Open(cfg.Database.GetDSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	conn := &Connection{
		db:     db,
		config: cfg,
	}

	if err := conn.configurePool(); err != nil {
		return nil, fmt.Errorf("failed to configure connection pool: %w", err)
	}

	return conn, nil
}

// DB returns the underlying gorm.DB instance
func (c *Connection) DB() *gorm.DB {
	return c.db
}

// RunMigrations executes all database migrations and ensures superuser exists
func (c *Connection) RunMigrations() error {
    if err := c.db.AutoMigrate(
        &models.User{},
        &models.RevokedToken{},
        &models.Channel{},
        &models.Watchlist{},
    ); err != nil {
        return fmt.Errorf("failed to run migrations: %w", err)
    }

    // Create superuser if it doesn't exist
    if err := c.ensureSuperuser(); err != nil {
        return fmt.Errorf("failed to ensure superuser exists: %w", err)
    }

    return nil
}

// ensureSuperuser creates the superuser if it doesn't already exist
func (c *Connection) ensureSuperuser() error {
    // Check if superuser already exists
    var count int64
    if err := c.db.Model(&models.User{}).Where("email = ? OR username = ?", 
        c.config.Superuser.Email, c.config.Superuser.Username).Count(&count).Error; err != nil {
        return fmt.Errorf("failed to check for existing superuser: %w", err)
    }

    if count > 0 {
        return nil // Superuser already exists
    }

    // Generate password hash
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(c.config.Superuser.Password), bcrypt.DefaultCost)
    if err != nil {
        return fmt.Errorf("failed to hash superuser password: %w", err)
    }

    // Create superuser
    superuser := &models.User{
        Email:        c.config.Superuser.Email,
        Username:     c.config.Superuser.Username,
        PasswordHash: string(hashedPassword),
    }

    if err := c.db.Create(superuser).Error; err != nil {
        return fmt.Errorf("failed to create superuser: %w", err)
    }

    return nil
}

// Close closes the database connection
func (c *Connection) Close() error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying *sql.DB: %w", err)
	}
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}
	return nil
}

// configurePool sets up the connection pool with optimal settings
func (c *Connection) configurePool() error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)           // Maximum number of idle connections
	sqlDB.SetMaxOpenConns(100)          // Maximum number of open connections
	sqlDB.SetConnMaxLifetime(time.Hour) // Maximum amount of time a connection may be reused

	return nil
}

// Ping checks if the database connection is still alive
func (c *Connection) Ping() error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying *sql.DB: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	return nil
}
