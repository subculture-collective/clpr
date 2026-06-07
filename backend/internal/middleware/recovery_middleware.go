package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

// JSONRecoveryMiddleware returns a middleware that recovers from panics and returns JSON errors
func JSONRecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with stack trace using structured logger
				stack := debug.Stack()
				logger := utils.GetLogger()

				fields := map[string]interface{}{
					"panic":      err,
					"stack":      string(stack),
					"request_id": c.GetString("RequestId"),
				}

				logger.Error("PANIC recovered", nil, fields)

				// Always return JSON error response
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":   "Internal server error",
					"code":    "INTERNAL_ERROR",
					"message": "An unexpected error occurred. Please try again later.",
				})
			}
		}()

		c.Next()
	}
}
