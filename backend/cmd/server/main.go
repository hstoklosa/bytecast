package main

import (
	"log"
	"os"

	"bytecast/configs"
	"bytecast/internal/database"
	"bytecast/internal/server"
)

func main() {
	// Configure logger
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	log.Println("Starting ByteCast API server...")
	
	// Step 1: Load configuration
	cfg, err := configs.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}
	
	// Step 2: Initialize database connection
	dbConn, err := database.New(cfg)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer func() {
		if err := dbConn.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()
	
	// Step 3: Run migrations
	if err := dbConn.RunMigrations(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}
	
	// Step 4: Create and start server
	srv := server.New(cfg, dbConn)
	if err := srv.Start(); err != nil {
		log.Fatal("Server error:", err)
	}
}
