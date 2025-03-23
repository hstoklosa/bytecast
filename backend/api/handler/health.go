package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"bytecast/api/utils"
	"bytecast/configs"
	"bytecast/internal/database"
	apperrors "bytecast/internal/errors"
	"bytecast/internal/services"
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
		utils.HandleError(c, apperrors.NewServiceUnavailable("Database connection failed", err))
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"message": "Service is healthy",
	})
}

func (h *HealthHandler) HandleWebSubCheck(c *gin.Context) {
	var apiKeyConfigured bool
	var callbackURL string
	
	if h.cfg != nil && h.cfg.YouTube.APIKey != "" {
		apiKeyConfigured = true
		callbackURL = h.cfg.YouTube.CallbackURL
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"message": "WebSub check completed",
		"data": gin.H{
			"callback_url": callbackURL,
			"api_key_configured": apiKeyConfigured,
			"pubsub_service_initialized": h.pubsubService != nil,
		},
	})
}

