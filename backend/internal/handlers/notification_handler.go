package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type notificationService interface {
	GetUserNotifications(context.Context, uuid.UUID, string, int, int) ([]models.NotificationWithSource, error)
	GetUnreadCount(context.Context, uuid.UUID) (int, error)
	MarkAsRead(context.Context, uuid.UUID, uuid.UUID) error
	MarkAllAsRead(context.Context, uuid.UUID) error
	DeleteNotification(context.Context, uuid.UUID, uuid.UUID) error
	GetPreferences(context.Context, uuid.UUID) (*models.NotificationPreferences, error)
	UpdatePreferences(context.Context, *models.NotificationPreferences) error
	ResetPreferences(context.Context, uuid.UUID) (*models.NotificationPreferences, error)
	RegisterDeviceToken(context.Context, uuid.UUID, string, string) error
	UnregisterDeviceToken(context.Context, uuid.UUID, string) error
}

// NotificationHandler handles notification-related HTTP requests
type NotificationHandler struct {
	notificationService notificationService
	emailService        *services.EmailService
}

// NewNotificationHandler creates a new NotificationHandler
func NewNotificationHandler(notificationService notificationService, emailService *services.EmailService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		emailService:        emailService,
	}
}

func notificationUserID(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get("user_id")
	userID, valid := value.(uuid.UUID)
	if !exists || !valid || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return uuid.Nil, false
	}
	return userID, true
}

// ListNotifications handles GET /notifications
func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	userID, ok := notificationUserID(c)
	if !ok {
		return
	}

	// Parse query parameters
	filter := c.DefaultQuery("filter", "all") // all, unread, read
	limitStr := c.DefaultQuery("limit", "50")
	pageStr := c.DefaultQuery("page", "1")

	if filter != "all" && filter != "unread" && filter != "read" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter"})
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit"})
		return
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 || page > 1000000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page"})
		return
	}

	offset := (page - 1) * limit

	// Get notifications
	notifications, err := h.notificationService.GetUserNotifications(
		c.Request.Context(),
		userID,
		filter,
		limit+1,
		offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve notifications",
		})
		return
	}

	// Get unread count
	unreadCount, err := h.notificationService.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve unread count",
		})
		return
	}

	hasMore := len(notifications) > limit
	if hasMore {
		notifications = notifications[:limit]
	}
	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"unread_count":  unreadCount,
		"page":          page,
		"limit":         limit,
		"has_more":      hasMore,
	})
}

// GetUnreadCount handles GET /notifications/count
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID, ok := notificationUserID(c)
	if !ok {
		return
	}

	// Get unread count
	count, err := h.notificationService.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve unread count",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"unread_count": count,
	})
}

// MarkAsRead handles PUT /notifications/:id/read
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, ok := notificationUserID(c)
	if !ok {
		return
	}

	// Parse notification ID
	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid notification ID",
		})
		return
	}

	// Mark as read
	err = h.notificationService.MarkAsRead(c.Request.Context(), notificationID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotificationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to mark notification as read",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification marked as read",
	})
}

// MarkAllAsRead handles PUT /notifications/read-all
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, ok := notificationUserID(c)
	if !ok {
		return
	}

	// Mark all as read
	err := h.notificationService.MarkAllAsRead(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to mark all notifications as read",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All notifications marked as read",
	})
}

// DeleteNotification handles DELETE /notifications/:id
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	userID, ok := notificationUserID(c)
	if !ok {
		return
	}

	// Parse notification ID
	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid notification ID",
		})
		return
	}

	// Delete notification
	err = h.notificationService.DeleteNotification(c.Request.Context(), notificationID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotificationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete notification",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification deleted",
	})
}

// GetPreferences handles GET /notifications/preferences
func (h *NotificationHandler) GetPreferences(c *gin.Context) {
	userID, ok := notificationUserID(c)
	if !ok {
		return
	}

	// Get preferences
	prefs, err := h.notificationService.GetPreferences(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve notification preferences",
		})
		return
	}

	c.JSON(http.StatusOK, prefs)
}

// UpdatePreferences handles PUT /notifications/preferences
func (h *NotificationHandler) UpdatePreferences(c *gin.Context) {
	userID, ok := notificationUserID(c)
	if !ok {
		return
	}

	// Parse request body
	var prefs models.NotificationPreferences
	if err := c.ShouldBindJSON(&prefs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}
	if prefs.EmailDigest != "immediate" && prefs.EmailDigest != "daily" && prefs.EmailDigest != "weekly" && prefs.EmailDigest != "never" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email_digest"})
		return
	}

	// Set user ID
	prefs.UserID = userID

	// Update preferences
	err := h.notificationService.UpdatePreferences(c.Request.Context(), &prefs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update notification preferences",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Notification preferences updated",
		"preferences": prefs,
	})
}

// ResetPreferences handles POST /notifications/preferences/reset
func (h *NotificationHandler) ResetPreferences(c *gin.Context) {
	userID, ok := notificationUserID(c)
	if !ok {
		return
	}

	// Reset preferences to defaults
	prefs, err := h.notificationService.ResetPreferences(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to reset notification preferences",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Notification preferences reset to defaults",
		"preferences": prefs,
	})
}

// Unsubscribe handles GET /notifications/unsubscribe (email unsubscribe)
func (h *NotificationHandler) Unsubscribe(c *gin.Context) {
	token := c.Query("token")
	if token == "" || len(token) > 256 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing unsubscribe token",
		})
		return
	}

	if h.emailService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Email service not configured",
		})
		return
	}

	err := h.emailService.Unsubscribe(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid or expired unsubscribe token",
		})
		return
	}

	// Return success HTML page for email links
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Unsubscribed - Clipper</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 40px 20px;
            background: #f5f5f5;
        }
        .container {
            background: white;
            padding: 40px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            text-align: center;
        }
        h1 {
            color: #2d3748;
            margin-bottom: 20px;
        }
        .success-icon {
            font-size: 48px;
            margin-bottom: 20px;
        }
        p {
            color: #4a5568;
            margin-bottom: 30px;
        }
        a {
            display: inline-block;
            background: #667eea;
            color: white;
            padding: 12px 30px;
            text-decoration: none;
            border-radius: 5px;
            font-weight: 600;
        }
        a:hover {
            background: #5568d3;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="success-icon">✓</div>
        <h1>Successfully Unsubscribed</h1>
        <p>You have been unsubscribed from email notifications. You can still manage your notification preferences in your account settings.</p>
        <a href="/">Return to Clipper</a>
    </div>
</body>
</html>`
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// RegisterDeviceToken handles POST /notifications/register
func (h *NotificationHandler) RegisterDeviceToken(c *gin.Context) {
	userID, ok := notificationUserID(c)
	if !ok {
		return
	}

	// Parse request body
	var req models.RegisterDeviceTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Update user's device token
	err := h.notificationService.RegisterDeviceToken(c.Request.Context(), userID, req.DeviceToken, req.DevicePlatform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to register device token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Device token registered successfully",
	})
}

// UnregisterDeviceToken handles DELETE /notifications/unregister
func (h *NotificationHandler) UnregisterDeviceToken(c *gin.Context) {
	userID, ok := notificationUserID(c)
	if !ok {
		return
	}

	// Parse request body
	var req models.UnregisterDeviceTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Unregister device token
	err := h.notificationService.UnregisterDeviceToken(c.Request.Context(), userID, req.DeviceToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to unregister device token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Device token unregistered successfully",
	})
}
