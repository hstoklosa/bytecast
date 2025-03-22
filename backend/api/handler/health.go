package handler

import (
	"bytecast/configs"
	"bytecast/internal/database"
	"bytecast/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db            *database.Connection
	cfg           *configs.Config
	pubsubService *services.PubSubService
}

func NewHealthHandler(db *database.Connection, cfg *configs.Config, pubsubService *services.PubSubService) *HealthHandler {
	return &HealthHandler{
		db:            db,
		cfg:           cfg,
		pubsubService: pubsubService,
	}
}

func (h *HealthHandler) RegisterRoutes(r *gin.Engine) {
	healthGroup := r.Group("/health")
	{
		healthGroup.GET("/", h.HandleServerCheck)
		healthGroup.GET("/websub", h.HandleWebSubCheck)
	}
}

func (h *HealthHandler) HandleServerCheck(c *gin.Context) {
	if err := h.db.Ping(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Database error",
			"error": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func (h *HealthHandler) HandleWebSubCheck(c *gin.Context) {
	var apiKeyConfigured bool
	var callbackURL string
	
	if h.cfg != nil && h.cfg.YouTube.APIKey != "" {
		apiKeyConfigured = true
		callbackURL = h.cfg.YouTube.CallbackURL
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "OK",
		"youtube_websub": gin.H{
			"callback_url": callbackURL,
			"api_key_configured": apiKeyConfigured,
			"pubsub_service_initialized": h.pubsubService != nil,
		},
	})
}

