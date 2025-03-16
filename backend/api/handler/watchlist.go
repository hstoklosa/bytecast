package handler

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"bytecast/internal/models"
	"bytecast/internal/services"
)

// Request and response structs
type createWatchlistRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Description string `json:"description"`
	Color       string `json:"color" binding:"required,hexcolor"`
}

type updateWatchlistRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Description string `json:"description"`
	Color       string `json:"color" binding:"required,hexcolor"`
}

type addChannelRequest struct {
	ChannelID string `json:"channel_id" binding:"required"` // Can be URL or ID
}

type watchlistResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("hexcolor", validateHexColor)
	}
}

func validateHexColor(fl validator.FieldLevel) bool {
	return regexp.MustCompile(`^#[a-fA-F0-9]{6}$`).MatchString(fl.Field().String())
}

type channelResponse struct {
	ID          uint   `json:"id"`
	YouTubeID   string `json:"youtube_id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Thumbnail   string `json:"thumbnail_url,omitempty"`
	CustomName  string `json:"custom_name,omitempty"`
}

// WatchlistHandler handles HTTP requests related to watchlists
type WatchlistHandler struct {
	watchlistService *services.WatchlistService
}

// NewWatchlistHandler creates a new watchlist handler
func NewWatchlistHandler(watchlistService *services.WatchlistService) *WatchlistHandler {
	return &WatchlistHandler{
		watchlistService: watchlistService,
	}
}

// RegisterRoutes registers the watchlist routes
func (h *WatchlistHandler) RegisterRoutes(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	watchlists := r.Group("/api/v1/watchlists")
	watchlists.Use(authMiddleware) // Require authentication for all watchlist routes

	watchlists.POST("", h.createWatchlist)
	watchlists.GET("", h.getUserWatchlists)
	watchlists.GET("/:id", h.getWatchlist)
	watchlists.PUT("/:id", h.updateWatchlist)
	watchlists.DELETE("/:id", h.deleteWatchlist)

	// Channel management within watchlists
	watchlists.POST("/:id/channels", h.addChannel)
	watchlists.GET("/:id/channels", h.getChannels)
	watchlists.DELETE("/:id/channels/:channel_id", h.removeChannel)
}

// errorResponse returns a standardized error response
func (h *WatchlistHandler) errorResponse(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"message": message,
		"status":  status,
	})
}

// getUserID extracts the user ID from the context
func (h *WatchlistHandler) getUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		h.errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return 0, false
	}

	return userID.(uint), true
}

// createWatchlist handles the creation of a new watchlist
func (h *WatchlistHandler) createWatchlist(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	var req createWatchlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	watchlist, err := h.watchlistService.CreateWatchlist(userID, req.Name, req.Description, req.Color)
	if err != nil {
		h.errorResponse(c, http.StatusInternalServerError, "Failed to create watchlist")
		return
	}

	c.JSON(http.StatusCreated, watchlistToResponse(watchlist))
}

// getUserWatchlists returns all watchlists for the authenticated user
func (h *WatchlistHandler) getUserWatchlists(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	watchlists, err := h.watchlistService.GetUserWatchlists(userID)
	if err != nil {
		h.errorResponse(c, http.StatusInternalServerError, "Failed to retrieve watchlists")
		return
	}

	response := make([]watchlistResponse, len(watchlists))
	for i, watchlist := range watchlists {
		response[i] = watchlistToResponse(&watchlist)
	}

	c.JSON(http.StatusOK, gin.H{
		"watchlists": response,
	})
}

// getWatchlist returns a specific watchlist by ID
func (h *WatchlistHandler) getWatchlist(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	watchlistID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid watchlist ID")
		return
	}

	watchlist, err := h.watchlistService.GetWatchlist(uint(watchlistID), userID)
	if err != nil {
		if err == services.ErrWatchlistNotFound {
			h.errorResponse(c, http.StatusNotFound, "Watchlist not found")
		} else {
			h.errorResponse(c, http.StatusInternalServerError, "Failed to retrieve watchlist")
		}
		return
	}

	c.JSON(http.StatusOK, watchlistToResponse(watchlist))
}

// updateWatchlist updates a watchlist's name and description
func (h *WatchlistHandler) updateWatchlist(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	watchlistID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid watchlist ID")
		return
	}

	var req updateWatchlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	watchlist, err := h.watchlistService.UpdateWatchlist(uint(watchlistID), userID, req.Name, req.Description, req.Color)
	if err != nil {
		if err == services.ErrWatchlistNotFound {
			h.errorResponse(c, http.StatusNotFound, "Watchlist not found")
		} else {
			h.errorResponse(c, http.StatusInternalServerError, "Failed to update watchlist")
		}
		return
	}

	c.JSON(http.StatusOK, watchlistToResponse(watchlist))
}

// deleteWatchlist deletes a watchlist
func (h *WatchlistHandler) deleteWatchlist(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	watchlistID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid watchlist ID")
		return
	}

	err = h.watchlistService.DeleteWatchlist(uint(watchlistID), userID)
	if err != nil {
		if err == services.ErrWatchlistNotFound {
			h.errorResponse(c, http.StatusNotFound, "Watchlist not found")
		} else {
			h.errorResponse(c, http.StatusInternalServerError, "Failed to delete watchlist")
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// addChannel adds a channel to a watchlist
func (h *WatchlistHandler) addChannel(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	watchlistID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid watchlist ID")
		return
	}

	var req addChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	err = h.watchlistService.AddChannelToWatchlist(uint(watchlistID), userID, req.ChannelID)
	if err != nil {
		switch err {
		case services.ErrWatchlistNotFound:
			h.errorResponse(c, http.StatusNotFound, "Watchlist not found")
		case services.ErrInvalidYouTubeID:
			h.errorResponse(c, http.StatusBadRequest, "Invalid YouTube channel ID or URL")
		case services.ErrYouTubeAPIError:
			h.errorResponse(c, http.StatusServiceUnavailable, "YouTube API service is currently unavailable")
		case services.ErrMissingAPIKey:
			h.errorResponse(c, http.StatusServiceUnavailable, "YouTube API key is not configured. Please add a YouTube API key to your environment variables.")
		default:
			h.errorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to add channel to watchlist: %v", err))
		}
		return
	}

	c.Status(http.StatusOK)
}

// getChannels returns all channels in a watchlist
func (h *WatchlistHandler) getChannels(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	watchlistID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid watchlist ID")
		return
	}

	channels, err := h.watchlistService.GetChannelsInWatchlist(uint(watchlistID), userID)
	if err != nil {
		if err == services.ErrWatchlistNotFound {
			h.errorResponse(c, http.StatusNotFound, "Watchlist not found")
		} else {
			h.errorResponse(c, http.StatusInternalServerError, "Failed to retrieve channels")
		}
		return
	}

	response := make([]channelResponse, len(channels))
	for i, channel := range channels {
		response[i] = channelToResponse(&channel)
	}

	c.JSON(http.StatusOK, gin.H{
		"channels": response,
	})
}

// removeChannel removes a channel from a watchlist
func (h *WatchlistHandler) removeChannel(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	watchlistID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.errorResponse(c, http.StatusBadRequest, "Invalid watchlist ID")
		return
	}

	channelID := c.Param("channel_id")
	if channelID == "" {
		h.errorResponse(c, http.StatusBadRequest, "Channel ID is required")
		return
	}

	err = h.watchlistService.RemoveChannelFromWatchlist(uint(watchlistID), userID, channelID)
	if err != nil {
		switch err {
		case services.ErrWatchlistNotFound:
			h.errorResponse(c, http.StatusNotFound, "Watchlist not found")
		case services.ErrChannelNotFound:
			h.errorResponse(c, http.StatusNotFound, "Channel not found in watchlist")
		default:
			h.errorResponse(c, http.StatusInternalServerError, "Failed to remove channel from watchlist")
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// Helper functions to convert models to response structs
func watchlistToResponse(watchlist *models.Watchlist) watchlistResponse {
	return watchlistResponse{
		ID:          watchlist.ID,
		Name:        watchlist.Name,
		Description: watchlist.Description,
		Color:       watchlist.Color,
		CreatedAt:   watchlist.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   watchlist.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func channelToResponse(channel *models.Channel) channelResponse {
	return channelResponse{
		ID:          channel.ID,
		YouTubeID:   channel.YouTubeID,
		Title:       channel.Title,
		Description: channel.Description,
		Thumbnail:   channel.ThumbnailURL,
		CustomName:  channel.CustomName,
	}
}
