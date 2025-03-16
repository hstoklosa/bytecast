package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"bytecast/api/handler"
	"bytecast/api/middleware"
	"bytecast/configs"
	"bytecast/internal/database"
	"bytecast/internal/services"
)

// Server represents the HTTP server and its dependencies
type Server struct {
    router  *gin.Engine
    db      *database.Connection
    cfg     *configs.Config
    server  *http.Server
}

// New creates a new server instance
func New(cfg *configs.Config, db *database.Connection) *Server {
    router := gin.Default()

    return &Server{
        router: router,
        db:     db,
        cfg:    cfg,
    }
}

// Start initializes and starts the server with graceful shutdown
func (s *Server) Start() error {
    // Get the underlying gorm.DB instance
    db := s.db.DB()

    // Initialize services
    watchlistService := services.NewWatchlistService(db, s.cfg)
    authService := services.NewAuthService(db, watchlistService, s.cfg.JWT.Secret)

    // Initialize route handlers
    authHandler := handler.NewAuthHandler(authService, s.cfg)
    watchlistHandler := handler.NewWatchlistHandler(watchlistService)

    // Register routes
    authHandler.RegisterRoutes(s.router)
    
    // Auth middleware for protected routes
    authMiddleware := middleware.AuthMiddleware([]byte(s.cfg.JWT.Secret))
    
    // Register watchlist routes with auth middleware
    watchlistHandler.RegisterRoutes(s.router, authMiddleware)

    // Protected routes group
    protected := s.router.Group("/api/v1")
    protected.Use(middleware.AuthMiddleware([]byte(s.cfg.JWT.Secret)))
    {
        protected.GET("/me", func(c *gin.Context) {
            userID, _ := c.Get("user_id")
            c.JSON(200, gin.H{"user_id": userID})
        })
    }

    // Health check route
    s.router.GET("/health", func(c *gin.Context) {
        if err := s.db.Ping(); err != nil {
            c.JSON(500, gin.H{"status": "Database error", "error": err.Error()})
            return
        }
        c.JSON(200, gin.H{"status": "OK"})
    })

    // Configure HTTP server

    s.server = &http.Server{
        Addr:    "0.0.0.0:" + s.cfg.Server.Port,
        Handler: s.router,
    }

    // Channel to listen for errors coming from the listener.
    serverErrors := make(chan error, 1)

    // Start the server
    go func() {
        log.Printf("Server is running on port %s", s.cfg.Server.Port)
        serverErrors <- s.server.ListenAndServe()
    }()

    // Channel to listen for an interrupt or terminate signal from the OS.
    shutdown := make(chan os.Signal, 1)
    signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

    // Blocking select for shutdown signal or server error
    select {
    case err := <-serverErrors:
        return fmt.Errorf("server error: %w", err)

    case sig := <-shutdown:
        log.Printf("Start shutdown... Signal: %v", sig)

        // Give outstanding requests a deadline for completion.
        ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
        defer cancel()

        // Asking listener to shut down and shed load.
        if err := s.server.Shutdown(ctx); err != nil {
            // Error from closing listeners, or context timeout.
            s.server.Close()
            return fmt.Errorf("could not stop server gracefully: %w", err)
        }
    }

	    return nil
}
