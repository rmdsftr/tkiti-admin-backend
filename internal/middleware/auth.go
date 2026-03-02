package middleware

import (
	"admin-panel/internal/config"
	"admin-panel/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc{
	return func(ctx *gin.Context) {
		token, err := utils.GetTokenFromCookie(ctx, "access_token")
		if err != nil {
            ctx.JSON(http.StatusUnauthorized, gin.H{
                "status": "error",
                "error":  "Unauthorized - No token provided",
            })
            ctx.Abort()
            return
        }

		claims, err := utils.ValidateToken(token, cfg.JWT.Secret)
		if err != nil {
            ctx.JSON(http.StatusUnauthorized, gin.H{
                "status": "error",
                "error":  "Unauthorized - Invalid token",
            })
            ctx.Abort()
            return
        }

		ctx.Set("nim", claims.Nim)
		ctx.Set("no_aslab", claims.NoAslab)
		ctx.Set("role", claims.Role)
		ctx.Next()
	}
}