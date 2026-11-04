package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// authenticatedUserID returns the identity established by AuthMiddleware.
// Treating a missing or malformed value as unauthenticated keeps handlers from
// accidentally passing interface{} values to SQL authorization checks.
func authenticatedUserID(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get("user_id")
	userID, ok := value.(uuid.UUID)
	if !exists || !ok || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return uuid.Nil, false
	}
	return userID, true
}
