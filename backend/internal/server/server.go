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
    db := s.db.DB()

    videoService := services.NewVideoService(db)
    
    // Initialize PubSub service (making it optional)
    pubsubService, err := services.NewPubSubService(db, s.cfg, videoService)
    if err != nil {
        log.Printf("Warning: Failed to initialize PubSub service: %v", err)
    }
    
    watchlistService := services.NewWatchlistService(db, s.cfg)
    if pubsubService != nil {
        watchlistService.SetPubSubService(pubsubService)
    }
    
    authService := services.NewAuthService(db, watchlistService, s.cfg.JWT.Secret)
    authHandler := handler.NewAuthHandler(authService, s.cfg)
    watchlistHandler := handler.NewWatchlistHandler(watchlistService)
    
    var pubsubHandler *handler.YouTubePubSubHandler
    if pubsubService != nil {
        pubsubHandler = handler.NewYouTubePubSubHandler(pubsubService)
    }
    
    authMiddleware := middleware.AuthMiddleware([]byte(s.cfg.JWT.Secret))
    
    authHandler.RegisterRoutes(s.router)
    watchlistHandler.RegisterRoutes(s.router, authMiddleware)

    if pubsubHandler != nil {
        callbackPath := "/api/v1/pubsub/callback"
        s.router.GET(callbackPath, pubsubHandler.HandleVerification)
        s.router.POST(callbackPath, pubsubHandler.HandleNotification)
        
		// TODO: cron job to renew subscriptions
		go func() {
			for {
				if err := pubsubService.RenewAllSubscriptions(); err != nil {
					log.Printf("Error renewing subscriptions: %v", err)
				}
				
				time.Sleep(12 * time.Hour)
			}
		}()
    }

    protected := s.router.Group("/api/v1")
    protected.Use(middleware.AuthMiddleware([]byte(s.cfg.JWT.Secret)))
    {
        protected.GET("/me", func(c *gin.Context) {
            userID, _ := c.Get("user_id")
            c.JSON(200, gin.H{"user_id": userID})
        })
    }

    s.router.GET("/health", func(c *gin.Context) {
        if err := s.db.Ping(); err != nil {
            c.JSON(500, gin.H{"status": "Database error", "error": err.Error()})
            return
        }
        c.JSON(200, gin.H{"status": "OK"})
    })

    s.router.GET("/debug/pubsub", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status": "OK",
            "youtube_pubsub": gin.H{
                "callback_url": s.cfg.YouTube.CallbackURL,
                "api_key_configured": s.cfg.YouTube.APIKey != "",
                "pubsub_service_initialized": pubsubService != nil,
                "routes_registered": pubsubHandler != nil,
            },
        })
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
