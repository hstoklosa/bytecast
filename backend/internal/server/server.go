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

type Server struct {
	/* Infrastructure */ 
	router       *gin.Engine
	db           *database.Connection
	cfg          *configs.Config
	configStatus *configs.ConfigStatus
	server       *http.Server
	logger       *log.Logger
	
	/* Dependencies */
	videoService     *services.VideoService
	youtubeService   *services.YouTubeService
	pubsubService    *services.PubSubService
	watchlistService *services.WatchlistService
	authService      *services.AuthService
}

// New creates a new server instance with all dependencies injected
func New(cfg *configs.Config, db *database.Connection, logger *log.Logger) (*Server, error) {
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS())
	
	s := &Server{
		router:      router,
		db:          db,
		cfg:         cfg,
		configStatus: cfg.GetConfigStatus(),
		logger:      logger,
	}

	if err := s.initServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}
	
	s.setupRoutes()
	
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

func (s *Server) Start() error {
	s.logger.Printf("Starting server on port %s", s.cfg.Server.Port)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Println("Shutting down server...")
	return s.server.Shutdown(ctx)
}

func (s *Server) initServices() error {
	db := s.db.DB()
	s.videoService = services.NewVideoService(db)
	
	if s.configStatus.YouTubeAPIEnabled {
		var err error
		s.youtubeService, err = services.NewYouTubeService(s.cfg)
		if err != nil {
			return fmt.Errorf("failed to initialize YouTube service: %w", err)
		}
		s.logger.Println("YouTube service initialized successfully")
	}
	
	if s.configStatus.PubSubEnabled {
		s.pubsubService = services.NewPubSubService(db, s.cfg, s.videoService)
		s.logger.Println("PubSub service initialized successfully")
	}
	
	s.watchlistService = services.NewWatchlistService(db, s.cfg, s.youtubeService)
	
	if s.pubsubService != nil {
		s.watchlistService.SetPubSubService(s.pubsubService)
	}
	
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

func (s *Server) setupRoutes() {
	healthHandler := s.newHealthHandler()
	healthHandler.RegisterRoutes(s.router)
	
	authMiddleware := middleware.AuthMiddleware([]byte(s.cfg.JWT.Secret))	
	authHandler := s.newAuthHandler()
	authHandler.RegisterRoutes(s.router)

	watchlistHandler := s.newWatchlistHandler()
	watchlistHandler.RegisterRoutes(s.router, authMiddleware)
	
	if s.pubsubService != nil {
		pubsubHandler := s.newPubSubHandler()
		pubsubHandler.RegisterRoutes(s.router)
	}
}
