package server

import (
	"bytecast/internal/services"
)

func (s *Server) initServices() error {
	db := s.db.DB()
	
	s.videoService = services.NewVideoService(db)
	s.pubsubService = services.NewPubSubService(db, s.cfg, s.videoService)
	s.watchlistService = services.NewWatchlistService(db, s.cfg)
	s.watchlistService.SetPubSubService(s.pubsubService)
	s.authService = services.NewAuthService(db, s.watchlistService, s.cfg.JWT.Secret)
	
	return nil
}