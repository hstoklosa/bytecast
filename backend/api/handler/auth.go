package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"bytecast/internal/services"
)

type registerRequest struct {
Email    string `json:"email" binding:"required,email"`
Username string `json:"username" binding:"required,min=3,max=24,alphanum"`
Password string `json:"password" binding:"required,min=6"`
}

type loginRequest struct {
Email    string `json:"email" binding:"required,email"`
Password string `json:"password" binding:"required,min=6"`
}

type refreshRequest struct {
    RefreshToken string `json:"refresh_token" binding:"required"`
}

type logoutRequest struct {
    RefreshToken string `json:"refresh_token" binding:"required"`
}

type AuthHandler struct {
authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) RegisterRoutes(r *gin.Engine) {
    auth := r.Group("/auth")
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

c.JSON(http.StatusCreated, gin.H{"message": "user registered successfully"})
}

func (h *AuthHandler) login(c *gin.Context) {
var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := h.authService.LoginUser(req.Email, req.Password)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

func (h *AuthHandler) logout(c *gin.Context) {
    var req logoutRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := h.authService.RevokeToken(req.RefreshToken); err != nil {
        switch err {
        case services.ErrTokenInvalid:
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
        }
        return
    }

    c.Status(http.StatusNoContent)
}

func (h *AuthHandler) refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := h.authService.RefreshTokens(req.RefreshToken)
	if err != nil {
		if err == services.ErrTokenInvalid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to refresh tokens"})
		return
	}

	c.JSON(http.StatusOK, tokens)
}
