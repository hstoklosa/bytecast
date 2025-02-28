package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"bytecast/internal/middleware"
	"bytecast/internal/models"
	"bytecast/internal/routes"
	"bytecast/internal/services"
)

func getDatabaseURL() string {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "bytecast"
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "bytecast"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port,
	)
}

func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// In production, you should ensure JWT_SECRET is set
		// For development, we'll use a default secret
		secret = "your-256-bit-secret" // Minimum 32 characters for HS256
	}
	return secret
}

func main() {
	// Initialize Gin
	r := gin.Default()

	// Database connection string from environment variables
	dsn := getDatabaseURL()

	// Initialize database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize JWT secret
	jwtSecret := getJWTSecret()

	// Initialize services
	authService := services.NewAuthService(db, jwtSecret)

	// Initialize route handlers
	authHandler := routes.NewAuthHandler(authService)

	// Register routes
	authHandler.RegisterRoutes(r)

	// Protected routes group
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware([]byte(jwtSecret)))
	{
		// Add protected routes here
		protected.GET("/me", func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			c.JSON(200, gin.H{"user_id": userID})
		})
	}

	// Health check route
	r.GET("/health", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(500, gin.H{"status": "Database error"})
			return
		}
		if err := sqlDB.Ping(); err != nil {
			c.JSON(500, gin.H{"status": "Database connection lost"})
			return
		}
		c.JSON(200, gin.H{"status": "OK"})
	})

	// Get port from environment variable, default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
