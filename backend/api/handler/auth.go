package handler

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"

    "bytecast/configs"
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
    authService *services.AuthService
    config      *configs.Config
}

func NewAuthHandler(authService *services.AuthService, config *configs.Config) *AuthHandler {
    return &AuthHandler{
        authService: authService,
        config:      config,
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
}

func (h *AuthHandler) errorResponse(c *gin.Context, status int, message string) {
    c.JSON(status, gin.H{
        "message": message,
        "status": status,
    })
}

func (h *AuthHandler) register(c *gin.Context) {
    var req registerRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.errorResponse(c, http.StatusBadRequest, "Please check your input and try again")
        return
    }

    if err := h.authService.RegisterUser(req.Email, req.Username, req.Password); err != nil {
        switch err {
        case services.ErrUserExists:
            h.errorResponse(c, http.StatusConflict, "This email is already registered")
            return
        case services.ErrUsernameTaken:
            h.errorResponse(c, http.StatusConflict, "This username is already taken")
            return
        default:
            h.errorResponse(c, http.StatusInternalServerError, "Failed to create account. Please try again")
            return
        }
    }

    // After successful registration, perform login to generate tokens
    tokens, exp, err := h.authService.LoginUser(req.Email, req.Password)
    if err != nil {
        h.errorResponse(c, http.StatusInternalServerError, "Failed to complete registration. Please try again")
        return
    }

    h.setRefreshCookie(c, tokens.RefreshToken, exp)
    
    c.JSON(http.StatusCreated, gin.H{
        "access_token": tokens.AccessToken,
    })
}

func (h *AuthHandler) login(c *gin.Context) {
    var req loginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.errorResponse(c, http.StatusBadRequest, "Please check your input and try again")
        return
    }

    tokens, exp, err := h.authService.LoginUser(req.Identifier, req.Password)
    if err != nil {
        if err == services.ErrInvalidCredentials {
            h.errorResponse(c, http.StatusUnauthorized, "Incorrect email or password")
            return
        }
        h.errorResponse(c, http.StatusInternalServerError, "Failed to log in. Please try again")
        return
    }

    h.setRefreshCookie(c, tokens.RefreshToken, exp)

    c.JSON(http.StatusOK, gin.H{
        "access_token": tokens.AccessToken,
    })
}

func (h *AuthHandler) refresh(c *gin.Context) {
    refreshToken, err := c.Cookie("refresh_token")
    if err != nil {
        h.errorResponse(c, http.StatusUnauthorized, "Session expired. Please log in again")
        return
    }

    tokens, exp, err := h.authService.RefreshTokens(refreshToken)
    if err != nil {
        if err == services.ErrTokenInvalid {
            h.errorResponse(c, http.StatusUnauthorized, "Invalid session. Please log in again")
            return
        }
        h.errorResponse(c, http.StatusInternalServerError, "Failed to refresh session. Please try again")
        return
    }

    h.setRefreshCookie(c, tokens.RefreshToken, exp)

    c.JSON(http.StatusOK, gin.H{
        "access_token": tokens.AccessToken,
    })
}

func (h *AuthHandler) logout(c *gin.Context) {
    refreshToken, err := c.Cookie("refresh_token")
    if err != nil {
        h.errorResponse(c, http.StatusUnauthorized, "No active session found")
        return
    }

    if err := h.authService.RevokeToken(refreshToken); err != nil {
        switch err {
        case services.ErrTokenInvalid:
            h.errorResponse(c, http.StatusUnauthorized, "Invalid session")
        default:
            h.errorResponse(c, http.StatusInternalServerError, "Failed to log out. Please try again")
        }
        return
    }

    // Clear the refresh token cookie
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

    c.Status(http.StatusNoContent)
}
