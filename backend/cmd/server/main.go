package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bytecast/configs"
	"bytecast/internal/database"
	"bytecast/internal/errors"
	"bytecast/internal/server"
)

func main() {
	logger := configureLogger()
	logger.Println("Starting ByteCast API server...")
	
	// Create a base context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Set up graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	
	cfg, err := configs.Load()
	if err != nil {
		if appErr, ok := err.(errors.AppError); ok {
			logger.Fatal("Configuration error: ", appErr.Message)
		} else {
			logger.Fatal("Configuration error: ", err)
		}
	}
	
	logger.Println("Configuration loaded successfully")
	configStatus := cfg.GetConfigStatus()
	logger.Printf("Feature status - YouTube API: %v, PubSub: %v", 
		configStatus.YouTubeAPIEnabled, 
		configStatus.PubSubEnabled)
	
	dbConn, err := setupDatabase(cfg, logger)
	if err != nil {
		logger.Fatal("Database setup failed:", err)
	}
	defer closeDatabase(dbConn, logger)
	
	srv, err := server.New(cfg, dbConn, logger)
	if err != nil {
		logger.Fatal("Failed to create server:", err)
	}
	
	// Start server in a goroutine
	go func() {
		if err := srv.Start(); err != nil {
			logger.Fatal("Server error:", err)
		}
	}()
	
	logger.Printf("ByteCast API server running on port %s", cfg.Server.Port)
	
	<-shutdown
	logger.Println("Shutdown signal received, gracefully shutting down...")
	
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 15*time.Second)
	defer shutdownCancel()
	
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Printf("Server shutdown error: %v", err)
	}
	
	logger.Println("Server gracefully stopped")
}

func configureLogger() *log.Logger {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	return logger
}

func setupDatabase(cfg *configs.Config, logger *log.Logger) (*database.Connection, error) {
	logger.Println("Initializing database connection...")
	dbConn, err := database.New(cfg)
	if err != nil {
		return nil, err
	}
	
	logger.Println("Running database migrations...")
	if err := dbConn.RunMigrations(); err != nil {
		return nil, err
	}
	
	return dbConn, nil
}

func closeDatabase(dbConn *database.Connection, logger *log.Logger) {
	logger.Println("Closing database connection...")
	if err := dbConn.Close(); err != nil {
		logger.Printf("Error closing database connection: %v", err)
	}
}
