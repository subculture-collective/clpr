package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type contactRepository interface {
	Create(context.Context, *models.ContactMessage) error
	List(context.Context, int, int, string, string) ([]*models.ContactMessage, int, error)
	UpdateStatus(context.Context, uuid.UUID, string) error
}

// ContactHandler handles contact form related HTTP requests
type ContactHandler struct {
	contactRepo contactRepository
}

// NewContactHandler creates a new contact handler
func NewContactHandler(contactRepo contactRepository) *ContactHandler {
	return &ContactHandler{contactRepo: contactRepo}
}

// SubmitContactMessage handles contact form submissions
func (h *ContactHandler) SubmitContactMessage(c *gin.Context) {
	var req models.CreateContactMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID if authenticated (optional)
	var userID *uuid.UUID
	if id, exists := c.Get("user_id"); exists {
		uid, ok := id.(uuid.UUID)
		if !ok || uid == uuid.Nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid authenticated user"})
			return
		}
		userID = &uid
	}

	// Get IP address and user agent for abuse prevention
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Create contact message
	contactMessage := &models.ContactMessage{
		ID:        uuid.New(),
		UserID:    userID,
		Email:     req.Email,
		Category:  req.Category,
		Subject:   req.Subject,
		Message:   req.Message,
		Status:    "pending",
		IPAddress: &ipAddress,
		UserAgent: &userAgent,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to database
	if err := h.contactRepo.Create(c.Request.Context(), contactMessage); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit contact message"})
		return
	}

	// Return success response (don't expose internal ID)
	c.JSON(http.StatusCreated, gin.H{
		"message": "Contact message submitted successfully. We will review your message and get back to you soon.",
		"status":  "success",
	})
}

// GetContactMessages retrieves contact messages (admin only)
func (h *ContactHandler) GetContactMessages(c *gin.Context) {
	// Parse query parameters
	page := 1
	limit := 20
	category := c.Query("category")
	status := c.Query("status")
	if category != "" && category != "abuse" && category != "account" && category != "billing" && category != "feedback" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category parameter"})
		return
	}
	if status != "" && status != "pending" && status != "reviewed" && status != "resolved" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status parameter"})
		return
	}

	if pageStr := c.Query("page"); pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil || p < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page parameter"})
			return
		}
		page = p
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l < 1 || l > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Limit must be between 1 and 100"})
			return
		}
		limit = l
	}

	// Get contact messages from repository
	messages, total, err := h.contactRepo.List(c.Request.Context(), page, limit, category, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve contact messages"})
		return
	}

	// Calculate pagination metadata
	totalPages := (total + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"data": messages,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total_items": total,
			"total_pages": totalPages,
		},
	})
}

// UpdateContactMessageStatus updates the status of a contact message (admin only)
func (h *ContactHandler) UpdateContactMessageStatus(c *gin.Context) {
	// Get message ID from URL
	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	// Parse request body
	var req struct {
		Status string `json:"status" binding:"required,oneof=pending reviewed resolved"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update status in database
	if err := h.contactRepo.UpdateStatus(c.Request.Context(), messageID, req.Status); err != nil {
		if errors.Is(err, repository.ErrContactMessageNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Contact message not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update contact message status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Contact message status updated successfully",
	})
}
