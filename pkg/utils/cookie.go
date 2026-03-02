package utils

import (
	"admin-panel/internal/config"

	"github.com/gin-gonic/gin"
)

func SetTokenCookies(c *gin.Context, cfg *config.Config, tokenPair *TokenPair) {
	c.SetCookie(
		"access_token",
		tokenPair.AccessToken,
		int(cfg.JWT.ExpiresIn.Seconds()),
		"/",
		cfg.Cookie.Domain,
		cfg.Cookie.Secure,
		cfg.Cookie.HTTPOnly,
	)

	c.SetCookie(
		"refresh_token",
		tokenPair.RefreshToken,
		int(cfg.JWT.RefreshExpiresIn.Seconds()),
		"/",
		cfg.Cookie.Domain,
		cfg.Cookie.Secure,
		cfg.Cookie.HTTPOnly,
	)
}

func ClearTokenCookie(c *gin.Context, cfg *config.Config){
	c.SetCookie(
		"access_token",
		"",
		-1,
		"/",
		cfg.Cookie.Domain,
		cfg.Cookie.Secure,
		cfg.Cookie.HTTPOnly,
	)

	c.SetCookie(
		"refresh_token",
		"",
		-1,
		"/",
		cfg.Cookie.Domain,
		cfg.Cookie.Secure,
		cfg.Cookie.HTTPOnly,
	)
}

func GetTokenFromCookie(c *gin.Context, cookieName string)(string, error){
	return c.Cookie(cookieName)
}