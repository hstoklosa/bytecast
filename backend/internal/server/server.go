package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
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
	// Infrastructure
	router  *gin.Engine
	db      *database.Connection
	cfg     *configs.Config
	server  *http.Server
	logger  *log.Logger
	
	// Services
	videoService     *services.VideoService
	pubsubService    *services.PubSubService
	watchlistService *services.WatchlistService
	authService      *services.AuthService
}

// New creates a new server instance with all dependencies injected
func New(cfg *configs.Config, db *database.Connection, logger *log.Logger) (*Server, error) {
	// Set mode based on environment
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize the router with middleware
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger(logger))
	
	// Create the server instance
	s := &Server{
		router: router,
		db:     db,
		cfg:    cfg,
		logger: logger,
	}

	// Initialize services and handlers
	if err := s.initServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}
	
	s.setupRoutes()
	
	// Configure the HTTP server with timeouts
	s.server = &http.Server{
		Addr:              fmt.Sprintf("0.0.0.0:%s", cfg.Server.Port),
		Handler:           router,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}
	
	return s, nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Printf("Starting server on port %s", s.cfg.Server.Port)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Println("Shutting down server...")
	return s.server.Shutdown(ctx)
}

// initServices initializes all the application services
func (s *Server) initServices() error {
	db := s.db.DB()
	
	s.videoService = services.NewVideoService(db)
	s.pubsubService = services.NewPubSubService(db, s.cfg, s.videoService)
	s.watchlistService = services.NewWatchlistService(db, s.cfg)
	s.watchlistService.SetPubSubService(s.pubsubService)
	s.authService = services.NewAuthService(db, s.watchlistService, s.cfg.JWT.Secret)
	
	return nil
}

// Factory methods for handlers
func (s *Server) newHealthHandler() *handler.HealthHandler {
	return handler.NewHealthHandler(s.db, s.cfg, s.pubsubService)
}

func (s *Server) newAuthHandler() *handler.AuthHandler {
	return handler.NewAuthHandler(s.authService, s.cfg)
}

func (s *Server) newWatchlistHandler() *handler.WatchlistHandler {
	return handler.NewWatchlistHandler(s.watchlistService)
}

func (s *Server) newPubSubHandler() *handler.YouTubePubSubHandler {
	return handler.NewYouTubePubSubHandler(s.pubsubService)
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Auth middleware
	authMiddleware := middleware.AuthMiddleware([]byte(s.cfg.JWT.Secret))

	// Health routes (no auth required)
	healthHandler := s.newHealthHandler()
	healthHandler.RegisterRoutes(s.router)
	
	// Auth routes (no auth required)
	authHandler := s.newAuthHandler()
	authHandler.RegisterRoutes(s.router)
	
	// Routes requiring authentication
	watchlistHandler := s.newWatchlistHandler()
	watchlistHandler.RegisterRoutes(s.router, authMiddleware)
	
	// PubSub webhook handler (no auth required)
	if s.pubsubService != nil {
		pubsubHandler := s.newPubSubHandler()
		pubsubHandler.RegisterRoutes(s.router)
	}
}
