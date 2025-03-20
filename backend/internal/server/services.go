package server

import (
	"log"
	"time"

	"bytecast/internal/services"
)

func (s *Server) initServices() error {
	db := s.db.DB()
	
	s.videoService = services.NewVideoService(db)
	pubsubService, err := services.NewPubSubService(db, s.cfg, s.videoService)
	if err != nil {
		log.Printf("Warning: Failed to initialize PubSub service: %v", err)
	} else {
		s.pubsubService = pubsubService
		go s.runSubscriptionRenewal()
	}
	
	s.watchlistService = services.NewWatchlistService(db, s.cfg)
	if s.pubsubService != nil {
		s.watchlistService.SetPubSubService(s.pubsubService)
	}
	
	s.authService = services.NewAuthService(db, s.watchlistService, s.cfg.JWT.Secret)
	
	return nil
}

// periodically renews YouTube PubSub subscriptions - TODO: cron jobs
func (s *Server) runSubscriptionRenewal() {
	for {
		if err := s.pubsubService.RenewAllSubscriptions(); err != nil {
			log.Printf("Error renewing subscriptions: %v", err)
		}
		
		time.Sleep(12 * time.Hour)
	}
} 