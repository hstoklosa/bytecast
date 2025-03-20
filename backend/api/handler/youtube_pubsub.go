package handler

import (
	"bytecast/internal/services"
	"io"
	"log"
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
	callbackPath := "/api/v1/pubsub/callback"
	r.GET(callbackPath, h.HandleVerification)
	r.POST(callbackPath, h.HandleNotification)
}

// HandleVerification handles the initial subscription verification request
func (h *YouTubePubSubHandler) HandleVerification(c *gin.Context) {
	// Get verification parameters
	mode := c.Query("hub.mode")
	topic := c.Query("hub.topic")
	challenge := c.Query("hub.challenge")

	log.Printf("PubSub verification request received: mode=%s, topic=%s", mode, topic)

	// Validate required parameters
	if mode == "" || topic == "" || challenge == "" {
		log.Printf("PubSub verification failed: missing required parameters")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required parameters",
		})
		return
	}

	switch mode {
	case "subscribe":
		log.Printf("PubSub subscribe verification success: echoing challenge")
		c.String(http.StatusOK, challenge)
	case "unsubscribe":
		log.Printf("PubSub unsubscribe verification success: echoing challenge")
		c.String(http.StatusOK, challenge)
	default:
		log.Printf("PubSub verification failed: invalid mode '%s'", mode)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid hub.mode",
		})
	}
}

// HandleNotification processes incoming video notifications
func (h *YouTubePubSubHandler) HandleNotification(c *gin.Context) {
	// Get the signature from the X-Hub-Signature header
	signature := c.GetHeader("X-Hub-Signature")
	if signature == "" {
		log.Printf("PubSub notification rejected: missing X-Hub-Signature header")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing X-Hub-Signature header",
		})
		return
	}

	log.Printf("PubSub notification received with signature: %s", signature)

	// Read the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("PubSub notification failed: unable to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read request body",
		})
		return
	}

	log.Printf("PubSub notification body size: %d bytes", len(body))

	if err := h.pubsubService.ProcessVideoNotification(body, signature); err != nil {
		log.Printf("PubSub notification processing failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process notification",
		})
		return
	}

	log.Printf("PubSub notification processed successfully")

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
} 