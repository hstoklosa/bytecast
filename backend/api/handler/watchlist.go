package handler

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"bytecast/api/utils"
	apperrors "bytecast/internal/errors"
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
	YoutubeID   string `json:"youtube_id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Thumbnail   string `json:"thumbnail_url,omitempty"`
	CustomName  string `json:"custom_name,omitempty"`
}

type WatchlistHandler struct {
	watchlistService *services.WatchlistService
}

func NewWatchlistHandler(watchlistService *services.WatchlistService) *WatchlistHandler {
	return &WatchlistHandler{
		watchlistService: watchlistService,
	}
}

func (h *WatchlistHandler) RegisterRoutes(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	watchlists := r.Group("/api/v1/watchlists")
	watchlists.Use(authMiddleware) // Require authentication for all watchlist routes

	watchlists.POST("", h.createWatchlist)
	watchlists.GET("", h.getUserWatchlists)
	watchlists.GET("/:id", h.getWatchlist)
	watchlists.PUT("/:id", h.updateWatchlist)
	watchlists.DELETE("/:id", h.deleteWatchlist)

	watchlists.POST("/:id/channels", h.addChannel)
	watchlists.GET("/:id/channels", h.getChannels)
	watchlists.DELETE("/:id/channels/:channel_id", h.removeChannel)
}

func (h *WatchlistHandler) getUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.HandleError(c, apperrors.NewUnauthorized("User not authenticated", nil))
		return 0, false
	}

	return userID.(uint), true
}

func (h *WatchlistHandler) createWatchlist(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	var req createWatchlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, err, "Invalid request format")
		return
	}

	watchlist, err := h.watchlistService.CreateWatchlist(userID, req.Name, req.Description, req.Color)
	if err != nil {
		utils.HandleError(c, apperrors.NewInternal("Failed to create watchlist", err))
		return
	}

	c.JSON(http.StatusCreated, watchlistToResponse(watchlist))
}

func (h *WatchlistHandler) getUserWatchlists(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	watchlists, err := h.watchlistService.GetUserWatchlists(userID)
	if err != nil {
		utils.HandleError(c, apperrors.NewInternal("Failed to retrieve watchlists", err))
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

func (h *WatchlistHandler) getWatchlist(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	watchlistID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.HandleError(c, apperrors.NewBadRequest("Invalid watchlist ID", err))
		return
	}

	watchlist, err := h.watchlistService.GetWatchlist(uint(watchlistID), userID)
	if err != nil {
		switch err {
		case services.ErrWatchlistNotFound:
			utils.HandleError(c, apperrors.NewNotFound("Watchlist not found", err))
		default:
			utils.HandleError(c, apperrors.NewInternal("Failed to retrieve watchlist", err))
		}
		return
	}

	c.JSON(http.StatusOK, watchlistToResponse(watchlist))
}

func (h *WatchlistHandler) updateWatchlist(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	watchlistID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.HandleError(c, apperrors.NewBadRequest("Invalid watchlist ID", err))
		return
	}

	var req updateWatchlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, err, "Invalid request format")
		return
	}

	watchlist, err := h.watchlistService.UpdateWatchlist(uint(watchlistID), userID, req.Name, req.Description, req.Color)
	if err != nil {
		switch err {
		case services.ErrWatchlistNotFound:
			utils.HandleError(c, apperrors.NewNotFound("Watchlist not found", err))
		default:
			utils.HandleError(c, apperrors.NewInternal("Failed to update watchlist", err))
		}
		return
	}

	c.JSON(http.StatusOK, watchlistToResponse(watchlist))
}

func (h *WatchlistHandler) deleteWatchlist(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	watchlistID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.HandleError(c, apperrors.NewBadRequest("Invalid watchlist ID", err))
		return
	}

	err = h.watchlistService.DeleteWatchlist(uint(watchlistID), userID)
	if err != nil {
		switch err {
		case services.ErrWatchlistNotFound:
			utils.HandleError(c, apperrors.NewNotFound("Watchlist not found", err))
		default:
			utils.HandleError(c, apperrors.NewInternal("Failed to delete watchlist", err))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *WatchlistHandler) addChannel(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	watchlistID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.HandleError(c, apperrors.NewBadRequest("Invalid watchlist ID", err))
		return
	}

	var req addChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, err, "Invalid request format")
		return
	}

	err = h.watchlistService.AddChannelToWatchlist(uint(watchlistID), userID, req.ChannelID)
	if err != nil {
		switch err {
		case services.ErrWatchlistNotFound:
			utils.HandleError(c, apperrors.NewNotFound("Watchlist not found", err))
		case services.ErrInvalidYouTubeID:
			utils.HandleError(c, apperrors.NewBadRequest("Invalid YouTube channel ID or URL", err))
		case services.ErrYouTubeAPIError:
			utils.HandleError(c, apperrors.NewServiceUnavailable("YouTube API service is currently unavailable", err))
		case services.ErrMissingAPIKey:
			utils.HandleError(c, apperrors.NewServiceUnavailable("YouTube API key is not configured. Please add a YouTube API key to your environment variables.", err))
		default:
			utils.HandleError(c, apperrors.NewInternal(fmt.Sprintf("Failed to add channel to watchlist: %v", err), err))
		}
		return
	}

	c.Status(http.StatusOK)
}

func (h *WatchlistHandler) getChannels(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	watchlistID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.HandleError(c, apperrors.NewBadRequest("Invalid watchlist ID", err))
		return
	}

	channels, err := h.watchlistService.GetChannelsInWatchlist(uint(watchlistID), userID)
	if err != nil {
		switch err {
		case services.ErrWatchlistNotFound:
			utils.HandleError(c, apperrors.NewNotFound("Watchlist not found", err))
		default:
			utils.HandleError(c, apperrors.NewInternal("Failed to retrieve channels", err))
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

func (h *WatchlistHandler) removeChannel(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	watchlistID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.HandleError(c, apperrors.NewBadRequest("Invalid watchlist ID", err))
		return
	}

	channelID := c.Param("channel_id")
	if channelID == "" {
		utils.HandleError(c, apperrors.NewBadRequest("Channel ID is required", nil))
		return
	}

	err = h.watchlistService.RemoveChannelFromWatchlist(uint(watchlistID), userID, channelID)
	if err != nil {
		switch err {
		case services.ErrWatchlistNotFound:
			utils.HandleError(c, apperrors.NewNotFound("Watchlist not found", err))
		case services.ErrChannelNotFound:
			utils.HandleError(c, apperrors.NewNotFound("Channel not found in watchlist", err))
		default:
			utils.HandleError(c, apperrors.NewInternal("Failed to remove channel from watchlist", err))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

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
		YoutubeID:   channel.YoutubeID,
		Title:       channel.Title,
		Description: channel.Description,
		Thumbnail:   channel.ThumbnailURL,
		CustomName:  channel.CustomName,
	}
}
