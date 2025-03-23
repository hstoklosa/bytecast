package handler

import (
	"bytecast/internal/services"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type YouTubePubSubHandler struct {
	pubsubService *services.PubSubService
}

func NewYouTubePubSubHandler(pubsubService *services.PubSubService) *YouTubePubSubHandler {
	if pubsubService == nil {
		log.Printf("PubSub service is nil")
	}

	return &YouTubePubSubHandler{
		pubsubService: pubsubService,
	}
}

func (h *YouTubePubSubHandler) RegisterRoutes(r *gin.Engine) {
	callbackPath := "/pubsub/callback"
	r.GET(callbackPath, h.HandleVerification)
	r.POST(callbackPath, h.HandleNotification)
}

func (h *YouTubePubSubHandler) HandleVerification(c *gin.Context) {
	mode := c.Query("hub.mode")
	challenge := c.Query("hub.challenge")
	// verifyToken := c.Query("hub.verify_token")

	if mode == "" || challenge == "" {
		c.String(http.StatusBadRequest, "Missing required parameters (hub.mode, hub.challenge).")
		return
	}

	switch mode {
	case "subscribe":
		c.String(http.StatusOK, challenge)
	case "unsubscribe":
		c.String(http.StatusOK, challenge)
	default:
		c.String(http.StatusBadRequest, "Invalid hub.mode")
	}
}

func (h *YouTubePubSubHandler) HandleNotification(c *gin.Context) {
	signature := c.GetHeader("X-Hub-Signature")
	if signature == "" {
		c.String(http.StatusBadRequest, "Missing X-Hub-Signature header")
		return
	}

	// Read the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusBadRequest, "Failed to read request body") // http.StatusInternalServerError
		return
	}

	if err := h.pubsubService.ProcessVideoNotification(body, signature); err != nil {
		c.String(http.StatusInternalServerError, "Failed to process notification")
		return
	}

	c.String(http.StatusOK, "Notification processed successfully")
} 