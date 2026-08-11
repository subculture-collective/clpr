package handlers

import (
	"net/http"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TopicHandler struct {
	repo *repository.ClipTopicRepository
}

func NewTopicHandler(repo *repository.ClipTopicRepository) *TopicHandler {
	return &TopicHandler{repo: repo}
}

func (h *TopicHandler) ListClipTopics(c *gin.Context) {
	clipID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid clip ID"})
		return
	}
	topics, err := h.repo.ListForClip(c.Request.Context(), clipID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list clip topics"})
		return
	}
	if topics == nil {
		topics = []models.ClipTopic{}
	}
	c.JSON(http.StatusOK, gin.H{"topics": topics})
}

func (h *TopicHandler) ReplaceClipTopics(c *gin.Context) {
	clipID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid clip ID"})
		return
	}
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	moderatorID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}
	var request models.ReplaceClipTopicsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.ReplaceManual(c.Request.Context(), clipID, moderatorID, request.TopicIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to replace clip topics"})
		return
	}
	h.ListClipTopics(c)
}

func (h *TopicHandler) MergeTopics(c *gin.Context) {
	sourceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic ID"})
		return
	}
	var request models.MergeTopicsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.Merge(c.Request.Context(), sourceID, request.TargetTopicID); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *TopicHandler) SplitTopic(c *gin.Context) {
	sourceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic ID"})
		return
	}
	userValue, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	userID, ok := userValue.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}
	var request models.SplitTopicRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	topic, err := h.repo.Split(c.Request.Context(), sourceID, userID, request)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"topic": topic})
}
