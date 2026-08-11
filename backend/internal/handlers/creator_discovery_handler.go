package handlers

import (
	"net/http"
	"strconv"

	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/gin-gonic/gin"
)

// ListCreatorDiscovery returns creator-first discovery rails.
// GET /api/v1/broadcasters/discover
func (h *BroadcasterHandler) ListCreatorDiscovery(c *gin.Context) {
	limit := 12
	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed < 1 || parsed > 24 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 24"})
			return
		}
		limit = parsed
	}

	rails, err := h.creatorDiscoveryRepo.ListCreatorDiscovery(c.Request.Context(), limit)
	if err != nil {
		utils.GetLogger().Error("Failed to list creator discovery rails", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list creator discovery rails"})
		return
	}

	c.Header("Cache-Control", "public, max-age=60, stale-while-revalidate=300")
	c.JSON(http.StatusOK, gin.H{"success": true, "data": rails})
}
