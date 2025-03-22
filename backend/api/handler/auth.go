package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"bytecast/api/middleware"
	"bytecast/api/utils"
	"bytecast/configs"
	apperrors "bytecast/internal/errors"
	"bytecast/internal/services"
)

type registerRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Username string `json:"username" binding:"required,min=3,max=24,alphanum"`
    Password string `json:"password" binding:"required,min=8"`
}

type loginRequest struct {
    Identifier string `json:"identifier" binding:"required,min=3"`
    Password   string `json:"password" binding:"required,min=6"`
}

type AuthHandler struct {
    authService   *services.AuthService
    config        *configs.Config
    authMiddleware gin.HandlerFunc
}

func NewAuthHandler(authService *services.AuthService, config *configs.Config) *AuthHandler {
    return &AuthHandler{
        authService:   authService,
        config:        config,
        authMiddleware: middleware.AuthMiddleware([]byte(config.JWT.Secret)),
    }
}

func (h *AuthHandler) setRefreshCookie(c *gin.Context, token string, exp time.Time) {
    secure := h.config.Server.Environment == "production"
    sameSite := http.SameSiteLaxMode
    if secure {
        sameSite = http.SameSiteStrictMode
    }

    c.SetSameSite(sameSite)
    c.SetCookie(
        "refresh_token",
        token,
        int(time.Until(exp).Seconds()),
        "/",
        h.config.Server.Domain,
        secure,
        true, // HTTPOnly
    )
}

func (h *AuthHandler) RegisterRoutes(r *gin.Engine) {
    auth := r.Group("/api/v1/auth")
    {
        auth.POST("/register", h.register)
        auth.POST("/login", h.login)
        auth.POST("/refresh", h.refresh)
        auth.POST("/logout", h.logout)
    }

    protected := r.Group("/api/v1/auth")
    protected.Use(h.authMiddleware)
    {
        protected.GET("/me", h.me)
    }
}

func (h *AuthHandler) register(c *gin.Context) {
    var req registerRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.HandleValidationError(c, err, "Invalid input data")
        return
    }

    if err := h.authService.RegisterUser(req.Email, req.Username, req.Password); err != nil {
        var appErr apperrors.AppError
        
        switch err {
        case services.ErrUserExists:
            appErr = apperrors.NewConflict("This email is already registered", err)
        case services.ErrUsernameTaken:
            appErr = apperrors.NewConflict("This username is already taken", err)
        default:
            appErr = utils.LogError("Failed to create an account", err)
        }
        
        utils.HandleError(c, appErr)
        return
    }

	// After successful registration, perform login to generate tokens
	tokens, exp, err := h.authService.LoginUser(req.Email, req.Password)
	if err != nil {
		appErr := apperrors.NewInternal("Failed to create account. Please try again", err)
		utils.HandleError(c, appErr)
		return
	}

    h.setRefreshCookie(c, tokens.RefreshToken, exp)
    
    c.JSON(http.StatusCreated, gin.H{
		"access_token": tokens.AccessToken,
		"expires_at": exp.Unix(),
    })
}

func (h *AuthHandler) login(c *gin.Context) {
    var req loginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.HandleValidationError(c, err, "Invalid input data")
        return
    }

    tokens, exp, err := h.authService.LoginUser(req.Identifier, req.Password)
    if err != nil {
        var appErr apperrors.AppError
        
        if err == services.ErrInvalidCredentials {
            appErr = apperrors.NewUnauthorized("Invalid username/email or password", err)
        } else {
            appErr = utils.LogError("Authentication failed", err)
        }
        
        utils.HandleError(c, appErr)
        return
    }

    user, err := h.authService.FindByIdentifier(req.Identifier)
    if err != nil {
        utils.HandleError(c, utils.LogError("Failed to retrieve user data", err))
        return
    }

    h.setRefreshCookie(c, tokens.RefreshToken, exp)

    c.JSON(http.StatusOK, gin.H{
		"access_token": tokens.AccessToken,
		"expires_at": exp.Unix(),
		"user": gin.H{
			"id": user.ID,
			"username": user.Username,
			"email": user.Email,
		},
    })
}

func (h *AuthHandler) refresh(c *gin.Context) {
    refreshToken, err := c.Cookie("refresh_token")
    if err != nil {
        utils.HandleError(c, apperrors.NewUnauthorized("Session expired. Please log in again", err))
        return
    }

    tokens, exp, err := h.authService.RefreshTokens(refreshToken)
    if err != nil {
        if err == services.ErrTokenInvalid {
            utils.HandleError(c, apperrors.NewUnauthorized("Invalid session. Please log in again", err))
            return
        }

        utils.HandleError(c, utils.LogError("Failed to refresh session", err))
        return
    }

    h.setRefreshCookie(c, tokens.RefreshToken, exp)
    c.JSON(http.StatusOK, gin.H{
        "access_token": tokens.AccessToken,
        "expires_at": exp.Unix(),
    })
}

func (h *AuthHandler) logout(c *gin.Context) {
    refreshToken, err := c.Cookie("refresh_token")
    if err != nil {
        utils.HandleError(c, apperrors.NewUnauthorized("No active session found", err))
        return
    }

    if err := h.authService.RevokeToken(refreshToken); err != nil {
        switch err {
        case services.ErrTokenInvalid:
            utils.HandleError(c, apperrors.NewUnauthorized("Invalid session", err))
        default:
            utils.HandleError(c, utils.LogError("Failed to log out", err))
        }
        return
    }

    secure := h.config.Server.Environment == "production"
    sameSite := http.SameSiteLaxMode
    if secure {
        sameSite = http.SameSiteStrictMode
    }

    c.SetSameSite(sameSite)
    c.SetCookie(
        "refresh_token",
        "",
        -1,
        "/",
        h.config.Server.Domain,
        secure,
        true,
    )

    c.JSON(http.StatusOK, gin.H{
        "status": "success",
        "message": "Logged out successfully",
    })
}

func (h *AuthHandler) me(c *gin.Context) {
    userID, exists := c.Get("userID")
    if !exists {
        utils.HandleError(c, apperrors.NewUnauthorized("Not authenticated", nil))
        return
    }

    id, ok := userID.(uint)
    if !ok {
        utils.HandleError(c, utils.LogError("Invalid user ID format", nil))
        return
    }
    
    user, err := h.authService.GetUserByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            utils.HandleError(c, apperrors.NewNotFound("User not found", err))
        } else {
            utils.HandleError(c, utils.LogError("Failed to retrieve user data", err))
        }
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "user": gin.H{
            "id": user.ID,
            "username": user.Username,
            "email": user.Email,
        },
    })
}