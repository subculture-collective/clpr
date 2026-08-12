package handlers

import (
	"errors"
	"math"
	"net/http"
	"strconv"

	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TagPromotionHandler struct{ service *services.TagPromotionService }

func NewTagPromotionHandler(service *services.TagPromotionService) *TagPromotionHandler {
	return &TagPromotionHandler{service: service}
}

func (h *TagPromotionHandler) List(c *gin.Context) {
	status := c.DefaultQuery("status", "pending")
	if status != "pending" && status != "approved" && status != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status must be pending, approved, or rejected"})
		return
	}
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(400, gin.H{"error": "page must be positive"})
		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(400, gin.H{"error": "limit must be between 1 and 100"})
		return
	}
	items, total, err := h.service.GetQueue(c.Request.Context(), status, limit, (page-1)*limit)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to load promotion queue"})
		return
	}
	c.JSON(200, gin.H{"items": items, "total": total, "page": page, "limit": limit, "total_pages": int(math.Ceil(float64(total) / float64(limit)))})
}

func (h *TagPromotionHandler) Approve(c *gin.Context) { h.review(c, true) }
func (h *TagPromotionHandler) Reject(c *gin.Context)  { h.review(c, false) }

func (h *TagPromotionHandler) review(c *gin.Context, approve bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid promotion ID"})
		return
	}
	reviewer, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	item, err := h.service.GetQueueItemByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrPromotionNotFound) {
			c.JSON(404, gin.H{"error": "Promotion not found"})
		} else {
			c.JSON(500, gin.H{"error": "Failed to load promotion"})
		}
		return
	}
	if approve {
		err = h.service.ApprovePromotion(c.Request.Context(), item.TagSlug, reviewer)
	} else {
		err = h.service.RejectPromotion(c.Request.Context(), item.TagSlug, reviewer)
	}
	if err != nil {
		c.JSON(409, gin.H{"error": "Promotion is no longer pending"})
		return
	}
	c.JSON(200, gin.H{"message": "Promotion reviewed"})
}
