package handler

import (
	"bytecast/internal/services"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// YouTubePubSubHandler handles PubSubHubbub notifications and verification requests
type YouTubePubSubHandler struct {
	pubsubService *services.PubSubService
}

// NewYouTubePubSubHandler creates a new PubSubHubbub handler
func NewYouTubePubSubHandler(pubsubService *services.PubSubService) *YouTubePubSubHandler {
	return &YouTubePubSubHandler{
		pubsubService: pubsubService,
	}
}

// RegisterRoutes registers all PubSubHubbub routes
func (h *YouTubePubSubHandler) RegisterRoutes(r *gin.Engine) {
	callbackPath := "/pubsub/callback"
	r.GET(callbackPath, h.HandleVerification)
	r.POST(callbackPath, h.HandleNotification)
}

// HandleVerification handles the initial subscription verification request
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

// HandleNotification processes incoming video notifications
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