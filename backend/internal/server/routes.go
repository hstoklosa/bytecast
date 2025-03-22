package server

func (s *Server) setupRoutes() {
	s.healthHandler.RegisterRoutes(s.router)
	s.authHandler.RegisterRoutes(s.router)
	s.watchlistHandler.RegisterRoutes(s.router, s.authMiddleware)
	if s.pubsubHandler != nil {
		s.pubsubHandler.RegisterRoutes(s.router)
	}
} 