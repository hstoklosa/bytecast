package server

import (
	"bytecast/api/handler"
	"bytecast/api/middleware"
)

// initHandlers initializes all HTTP handlers and middleware
func (s *Server) initHandlers() {
	s.authMiddleware = middleware.AuthMiddleware([]byte(s.cfg.JWT.Secret))
	s.authHandler = handler.NewAuthHandler(s.authService, s.cfg)
	s.watchlistHandler = handler.NewWatchlistHandler(s.watchlistService)
	
	if s.pubsubService != nil {
		s.pubsubHandler = handler.NewYouTubePubSubHandler(s.pubsubService)
	}
} 