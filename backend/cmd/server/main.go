package main

import (
    "log"

    "bytecast/configs"
    "bytecast/internal/database"
    "bytecast/internal/server"
)

func main() {
    // Load configuration
    cfg, err := configs.Load()
    if err != nil {
        log.Fatal("Failed to load configuration:", err)
    }

    // Initialize database connection
    dbConn, err := database.New(cfg)
    if err != nil {
        log.Fatal("Failed to initialize database:", err)
    }
    defer func() {
        if err := dbConn.Close(); err != nil {
            log.Printf("Error closing database connection: %v", err)
        }
    }()

    // Run migrations
    if err := dbConn.RunMigrations(); err != nil {
        log.Fatal("Failed to run migrations:", err)
    }

    // Create and start server
    srv := server.New(cfg, dbConn)
    if err := srv.Start(); err != nil {
        log.Fatal("Server error:", err)
    }
}
