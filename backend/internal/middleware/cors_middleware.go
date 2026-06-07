package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"git.subcult.tv/subculture-collective/clpr/config"
)

// CORSMiddleware creates CORS middleware
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	allowedOrigins := strings.Split(cfg.CORS.AllowedOrigins, ",")
	originsMap := make(map[string]bool)
	for _, origin := range allowedOrigins {
		originsMap[strings.TrimSpace(origin)] = true
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		if originsMap[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, Authorization, accept, origin, Cache-Control, X-Requested-With, X-CSRF-Token")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
			c.Writer.Header().Set("Access-Control-Expose-Headers", "X-CSRF-Token")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
