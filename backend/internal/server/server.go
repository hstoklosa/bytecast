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
	// Infrastructure
	router  *gin.Engine
	db      *database.Connection
	cfg     *configs.Config
	server  *http.Server
	
	// Services
	videoService     *services.VideoService
	pubsubService    *services.PubSubService
	watchlistService *services.WatchlistService
	authService      *services.AuthService
	
	// Middleware
	authMiddleware gin.HandlerFunc
	
	// Handlers
	authHandler      *handler.AuthHandler
	watchlistHandler *handler.WatchlistHandler
	pubsubHandler    *handler.YouTubePubSubHandler
	healthHandler    *handler.HealthHandler
}

func New(cfg *configs.Config, db *database.Connection) *Server {
	// Set mode based on environment
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	
	// Use custom logger and recovery middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	
	// Use our custom error handling middleware
	router.Use(middleware.ErrorHandler())

	return &Server{
		router: router,
		db:     db,
		cfg:    cfg,
	}
}

// Start initializes and starts the server with graceful shutdown
func (s *Server) Start() error {
	if err := s.initServices(); err != nil {
		return fmt.Errorf("failed to initialize services: %w", err)
	}
	s.initHandlers()
	s.setupRoutes()
	
	s.server = &http.Server{
		Addr:    "0.0.0.0:" + s.cfg.Server.Port,
		Handler: s.router,
	}

	// Channel to listen for errors coming from the listener
	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("Server is running on port %s", s.cfg.Server.Port)
		serverErrors <- s.server.ListenAndServe()
	}()

	// Channel to listen for an interrupt or terminate signal from the OS
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking select for shutdown signal or server error
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Printf("Start shutdown... Signal: %v", sig)

		// Give outstanding requests a deadline for completion
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Asking listener to shut down and shed load
		if err := s.server.Shutdown(ctx); err != nil {
			// Error from closing listeners, or context timeout
			s.server.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}
