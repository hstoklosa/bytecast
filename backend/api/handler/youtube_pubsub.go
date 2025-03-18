package handler

import (
	"bytecast/internal/services"
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

	// Handle different verification modes
	switch mode {
	case "subscribe":
		log.Printf("PubSub subscribe verification success: echoing challenge")
		// Echo the challenge string
		c.String(http.StatusOK, challenge)
	case "unsubscribe":
		log.Printf("PubSub unsubscribe verification success: echoing challenge")
		// Echo the challenge string
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
	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
} 