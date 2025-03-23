package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// CookieOptions represents the options for setting a cookie
type CookieOptions struct {
	Name     string
	Value    string
	Path     string
	Domain   string
	MaxAge   int
	Secure   bool
	HTTPOnly bool
	SameSite http.SameSite
}

// SetCookie sets a cookie in the response
func SetCookie(c *gin.Context, options CookieOptions) {
	if options.SameSite == 0 {
		// Default to Lax if not specified
		options.SameSite = http.SameSiteLaxMode
	}

	c.SetSameSite(options.SameSite)
	c.SetCookie(
		options.Name,
		options.Value,
		options.MaxAge,
		options.Path,
		options.Domain,
		options.Secure,
		options.HTTPOnly,
	)
}

// SetRefreshTokenCookie sets a refresh token cookie with appropriate security settings
func SetRefreshTokenCookie(c *gin.Context, token string, exp time.Time, secure bool, domain string) {
	sameSite := http.SameSiteLaxMode
	if secure {
		sameSite = http.SameSiteStrictMode
	}

	SetCookie(c, CookieOptions{
		Name:     "refresh_token",
		Value:    token,
		MaxAge:   int(time.Until(exp).Seconds()),
		Path:     "/",
		Domain:   domain,
		Secure:   secure,
		HTTPOnly: true,
		SameSite: sameSite,
	})
}

// ClearRefreshTokenCookie clears the refresh token cookie
func ClearRefreshTokenCookie(c *gin.Context, secure bool, domain string) {
	sameSite := http.SameSiteLaxMode
	if secure {
		sameSite = http.SameSiteStrictMode
	}

	SetCookie(c, CookieOptions{
		Name:     "refresh_token",
		Value:    "",
		MaxAge:   -1, // Expire immediately
		Path:     "/",
		Domain:   domain,
		Secure:   secure,
		HTTPOnly: true,
		SameSite: sameSite,
	})
} 