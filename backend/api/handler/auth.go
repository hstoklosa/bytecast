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

func (h *AuthHandler) register(c *gin.Context) {
    var req registerRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := h.authService.RegisterUser(req.Email, req.Username, req.Password); err != nil {
        switch err {
        case services.ErrUserExists:
            c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
            return
        case services.ErrUsernameTaken:
            c.JSON(http.StatusConflict, gin.H{"error": "Username already taken"})
            return
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
            return
        }
    }

    // After successful registration, perform login to generate tokens
    tokens, exp, err := h.authService.LoginUser(req.Email, req.Password)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
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
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    tokens, exp, err := h.authService.LoginUser(req.Identifier, req.Password)
    if err != nil {
        if err == services.ErrInvalidCredentials {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
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
        c.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh token"})
        return
    }

    tokens, exp, err := h.authService.RefreshTokens(refreshToken)
    if err != nil {
        if err == services.ErrTokenInvalid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to refresh tokens"})
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
        c.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh token"})
        return
    }

    if err := h.authService.RevokeToken(refreshToken); err != nil {
        switch err {
        case services.ErrTokenInvalid:
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
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
