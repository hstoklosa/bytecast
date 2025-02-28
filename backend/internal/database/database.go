package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

"bytecast/configs"
"bytecast/internal/app/models"
)

// Connection holds the database connection and configuration
type Connection struct {
	db     *gorm.DB
	config *configs.Database
}

// New creates a new database connection with proper connection pooling
func New(cfg *configs.Database) (*Connection, error) {
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{})
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

// RunMigrations executes all database migrations
func (c *Connection) RunMigrations() error {
	if err := c.db.AutoMigrate(&models.User{}); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
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
