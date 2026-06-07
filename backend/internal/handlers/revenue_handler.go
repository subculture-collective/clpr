package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

// RevenueHandler handles revenue reporting HTTP requests
type RevenueHandler struct {
	revenueService *services.RevenueService
}

// NewRevenueHandler creates a new revenue handler
func NewRevenueHandler(revenueService *services.RevenueService) *RevenueHandler {
	return &RevenueHandler{
		revenueService: revenueService,
	}
}

// GetRevenueMetrics returns comprehensive revenue metrics for admin dashboard
// @Summary Get revenue metrics
// @Description Returns MRR, churn, ARPU, plan distribution, and cohort retention metrics
// @Tags admin
// @Produce json
// @Success 200 {object} models.RevenueMetrics
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/admin/revenue [get]
func (h *RevenueHandler) GetRevenueMetrics(c *gin.Context) {
	metrics, err := h.revenueService.GetRevenueMetrics(c.Request.Context())
	if err != nil {
		// Log full error internally for debugging
		log.Printf("Revenue metrics error: %v", err)
		// Return generic error to client to avoid exposing internal details
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to retrieve metrics at this time",
			"code":  "METRICS_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, metrics)
}
