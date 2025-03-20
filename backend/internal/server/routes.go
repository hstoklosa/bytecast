package server

import (
	"github.com/gin-gonic/gin"
)

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	s.router.GET("/health", s.healthCheckHandler)
	s.router.GET("/debug/pubsub", s.debugPubSubHandler)

	s.authHandler.RegisterRoutes(s.router)
	s.watchlistHandler.RegisterRoutes(s.router, s.authMiddleware)
	s.pubsubHandler.RegisterRoutes(s.router)
}

// healthCheckHandler handles the health check endpoint
func (s *Server) healthCheckHandler(c *gin.Context) {
	if err := s.db.Ping(); err != nil {
		c.JSON(500, gin.H{"status": "Database error", "error": err.Error()})
		return
	}
	
	c.JSON(200, gin.H{"status": "OK"})
}

// debugPubSubHandler handles the debug endpoint for PubSub status
func (s *Server) debugPubSubHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "OK",
		"youtube_pubsub": gin.H{
			"callback_url": s.cfg.YouTube.CallbackURL,
			"api_key_configured": s.cfg.YouTube.APIKey != "",
			"pubsub_service_initialized": s.pubsubService != nil,
			"routes_registered": s.pubsubHandler != nil,
		},
	})
} 