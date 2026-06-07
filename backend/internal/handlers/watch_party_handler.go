package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"golang.org/x/crypto/bcrypt"
)

// WatchPartyHandler handles watch party requests
type WatchPartyHandler struct {
	watchPartyService *services.WatchPartyService
	hubManager        *services.WatchPartyHubManager
	watchPartyRepo    *repository.WatchPartyRepository
	analyticsRepo     *repository.AnalyticsRepository
	upgrader          websocket.Upgrader
	allowedOrigins    map[string]bool
}

// NewWatchPartyHandler creates a new WatchPartyHandler
func NewWatchPartyHandler(
	watchPartyService *services.WatchPartyService,
	hubManager *services.WatchPartyHubManager,
	watchPartyRepo *repository.WatchPartyRepository,
	analyticsRepo *repository.AnalyticsRepository,
	cfg *config.Config,
) *WatchPartyHandler {
	// Parse allowed origins from config
	allowedOrigins := make(map[string]bool)
	if cfg.CORS.AllowedOrigins != "" {
		origins := strings.Split(cfg.CORS.AllowedOrigins, ",")
		for _, origin := range origins {
			allowedOrigins[strings.TrimSpace(origin)] = true
		}
	}

	handler := &WatchPartyHandler{
		watchPartyService: watchPartyService,
		hubManager:        hubManager,
		watchPartyRepo:    watchPartyRepo,
		analyticsRepo:     analyticsRepo,
		allowedOrigins:    allowedOrigins,
	}

	// Initialize upgrader with proper origin checking
	handler.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			// Allow requests with no origin (e.g., same-origin requests)
			if origin == "" {
				return true
			}
			// Check if origin is in allowed list
			return handler.allowedOrigins[origin]
		},
	}

	return handler
}

// CreateWatchParty handles POST /api/v1/watch-parties
func (h *WatchPartyHandler) CreateWatchParty(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "UNAUTHORIZED",
				Message: "Authentication required",
			},
		})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	// Parse request body
	var req models.CreateWatchPartyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	// Create watch party
	party, err := h.watchPartyService.CreateWatchParty(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to create watch party",
			},
		})
		return
	}

	// Generate invite URL
	inviteURL := h.watchPartyService.GetInviteURL(party.InviteCode)

	c.JSON(http.StatusCreated, StandardResponse{
		Success: true,
		Data: gin.H{
			"id":          party.ID,
			"invite_code": party.InviteCode,
			"invite_url":  inviteURL,
			"party":       party,
		},
	})
}

// JoinWatchParty handles POST /api/v1/watch-parties/:code/join
func (h *WatchPartyHandler) JoinWatchParty(c *gin.Context) {
	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "UNAUTHORIZED",
				Message: "Authentication required",
			},
		})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	// Get invite code from URL (matches :id in route)
	inviteCode := c.Param("id")
	if inviteCode == "" {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invite code is required",
			},
		})
		return
	}

	// Parse optional request body for password
	var joinReq models.JoinWatchPartyRequest
	_ = c.ShouldBindJSON(&joinReq) // Ignore error as body is optional

	// Get party by invite code to check for password protection
	party, err := h.watchPartyRepo.GetByInviteCode(c.Request.Context(), inviteCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to verify watch party",
			},
		})
		return
	}

	if party == nil {
		c.JSON(http.StatusNotFound, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Watch party not found or has ended",
			},
		})
		return
	}

	// Check password if party is password-protected
	if party.Password != nil && *party.Password != "" {
		if joinReq.Password == nil || *joinReq.Password == "" {
			c.JSON(http.StatusForbidden, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "PASSWORD_REQUIRED",
					Message: "This watch party requires a password",
				},
			})
			return
		}

		err := bcrypt.CompareHashAndPassword([]byte(*party.Password), []byte(*joinReq.Password))
		if err != nil {
			c.JSON(http.StatusForbidden, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "INVALID_PASSWORD",
					Message: "Incorrect password",
				},
			})
			return
		}
	}

	// Join watch party
	party, err = h.watchPartyService.JoinWatchParty(c.Request.Context(), inviteCode, userID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorCode := "INTERNAL_ERROR"
		errorMessage := "Failed to join watch party"

		if err.Error() == "watch party not found or has ended" {
			statusCode = http.StatusNotFound
			errorCode = "NOT_FOUND"
			errorMessage = err.Error()
		} else if strings.HasPrefix(err.Error(), "party is full") {
			statusCode = http.StatusForbidden
			errorCode = "PARTY_FULL"
			errorMessage = err.Error()
		}

		c.JSON(statusCode, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    errorCode,
				Message: errorMessage,
			},
		})
		return
	}

	// Track join event (non-blocking, log errors but don't fail request)
	if err := h.analyticsRepo.TrackWatchPartyEvent(c.Request.Context(), party.ID, &userID, "join", nil); err != nil {
		log.Printf("Failed to track join event for party %s, user %s: %v", party.ID, userID, err)
	}

	inviteURL := h.watchPartyService.GetInviteURL(party.InviteCode)

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"party":      party,
			"invite_url": inviteURL,
		},
	})
}

// GetWatchParty handles GET /api/v1/watch-parties/:id
func (h *WatchPartyHandler) GetWatchParty(c *gin.Context) {
	// Parse party ID
	partyIDStr := c.Param("id")
	partyID, err := uuid.Parse(partyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid party ID",
			},
		})
		return
	}

	// Get optional user ID from context
	var userID *uuid.UUID
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(uuid.UUID); ok {
			userID = &uid
		}
	}

	// Get watch party
	party, err := h.watchPartyService.GetWatchParty(c.Request.Context(), partyID, userID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorCode := "INTERNAL_ERROR"
		errorMessage := "Failed to get watch party"

		if err.Error() == "watch party not found" {
			statusCode = http.StatusNotFound
			errorCode = "NOT_FOUND"
			errorMessage = err.Error()
		} else if strings.HasPrefix(err.Error(), "unauthorized:") {
			statusCode = http.StatusForbidden
			errorCode = "FORBIDDEN"
			errorMessage = err.Error()
		}

		c.JSON(statusCode, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    errorCode,
				Message: errorMessage,
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    party,
	})
}

// GetParticipants handles GET /api/v1/watch-parties/:id/participants
func (h *WatchPartyHandler) GetParticipants(c *gin.Context) {
	// Parse party ID
	partyIDStr := c.Param("id")
	partyID, err := uuid.Parse(partyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid party ID",
			},
		})
		return
	}

	// Get optional user ID from context for visibility check
	var userID *uuid.UUID
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(uuid.UUID); ok {
			userID = &uid
		}
	}

	// Get party to check visibility
	party, err := h.watchPartyRepo.GetByID(c.Request.Context(), partyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get watch party",
			},
		})
		return
	}

	if party == nil {
		c.JSON(http.StatusNotFound, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Watch party not found",
			},
		})
		return
	}

	// Check visibility permissions for private parties
	if party.Visibility == "private" {
		if userID == nil {
			c.JSON(http.StatusForbidden, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "FORBIDDEN",
					Message: "Cannot view participants of private party without authentication",
				},
			})
			return
		}
		// Check if user is a participant
		participant, err := h.watchPartyRepo.GetParticipant(c.Request.Context(), partyID, *userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "INTERNAL_ERROR",
					Message: "Failed to verify participant status",
				},
			})
			return
		}
		if participant == nil || participant.LeftAt != nil {
			c.JSON(http.StatusForbidden, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "FORBIDDEN",
					Message: "Not a participant of this private party",
				},
			})
			return
		}
	}

	// Get participants
	participants, err := h.watchPartyRepo.GetActiveParticipants(c.Request.Context(), partyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get participants",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"participants": participants,
			"count":        len(participants),
		},
	})
}

// LeaveWatchParty handles DELETE /api/v1/watch-parties/:id/leave
func (h *WatchPartyHandler) LeaveWatchParty(c *gin.Context) {
	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "UNAUTHORIZED",
				Message: "Authentication required",
			},
		})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	// Parse party ID
	partyIDStr := c.Param("id")
	partyID, err := uuid.Parse(partyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid party ID",
			},
		})
		return
	}

	// Leave watch party
	err = h.watchPartyService.LeaveWatchParty(c.Request.Context(), partyID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to leave watch party",
			},
		})
		return
	}

	// Track leave event (non-blocking, log errors but don't fail request)
	if err := h.analyticsRepo.TrackWatchPartyEvent(c.Request.Context(), partyID, &userID, "leave", nil); err != nil {
		log.Printf("Failed to track leave event for party %s, user %s: %v", partyID, userID, err)
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"message": "Successfully left watch party",
		},
	})
}

// EndWatchParty handles POST /api/v1/watch-parties/:id/end
func (h *WatchPartyHandler) EndWatchParty(c *gin.Context) {
	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "UNAUTHORIZED",
				Message: "Authentication required",
			},
		})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	// Parse party ID
	partyIDStr := c.Param("id")
	partyID, err := uuid.Parse(partyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid party ID",
			},
		})
		return
	}

	// End watch party
	err = h.watchPartyService.EndWatchParty(c.Request.Context(), partyID, userID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorCode := "INTERNAL_ERROR"

		if strings.HasPrefix(err.Error(), "unauthorized:") {
			statusCode = http.StatusForbidden
			errorCode = "FORBIDDEN"
		}

		c.JSON(statusCode, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    errorCode,
				Message: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"message": "Watch party ended successfully",
		},
	})
}

// WatchPartyWebSocket handles WebSocket connections for watch party sync
func (h *WatchPartyHandler) WatchPartyWebSocket(c *gin.Context) {
	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "UNAUTHORIZED",
				Message: "Authentication required",
			},
		})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	// Parse party ID
	partyIDStr := c.Param("id")
	partyID, err := uuid.Parse(partyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid party ID",
			},
		})
		return
	}

	// Verify user is a participant
	participant, err := h.watchPartyRepo.GetParticipant(c.Request.Context(), partyID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to verify participant",
			},
		})
		return
	}

	if participant == nil || participant.LeftAt != nil {
		c.JSON(http.StatusForbidden, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "FORBIDDEN",
				Message: "Not a participant of this watch party",
			},
		})
		return
	}

	// Upgrade to WebSocket with subprotocol support
	// If client sent auth via Sec-WebSocket-Protocol, echo it back to complete handshake
	responseHeader := http.Header{}
	if subprotocol := c.GetHeader("Sec-WebSocket-Protocol"); subprotocol != "" {
		// Extract just the auth protocol part (format: auth.bearer.<token>)
		if strings.HasPrefix(subprotocol, "auth.bearer.") {
			responseHeader.Set("Sec-WebSocket-Protocol", subprotocol)
		}
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, responseHeader)
	if err != nil {
		return
	}

	// Get or create hub for this party
	hub := h.hubManager.GetOrCreateHub(partyID)

	// Create a context for the WebSocket lifetime that's independent of the HTTP request
	// This allows the WebSocket to outlive the initial HTTP request
	wsCtx := context.Background()

	// Create client
	client := &services.WatchPartyClient{
		Hub:    hub,
		Conn:   conn,
		UserID: userID,
		Role:   participant.Role,
		Send:   make(chan []byte, 256),
		User:   participant.User,
	}

	// Register client with hub
	hub.Register <- client

	// Start client pumps with WebSocket context
	go client.WritePump()
	go client.ReadPump(wsCtx)
}

// SendMessage handles POST /api/v1/watch-parties/:id/messages
func (h *WatchPartyHandler) SendMessage(c *gin.Context) {
	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "UNAUTHORIZED",
				Message: "Authentication required",
			},
		})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	// Parse party ID
	partyIDStr := c.Param("id")
	partyID, err := uuid.Parse(partyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid party ID",
			},
		})
		return
	}

	// Verify user is a participant
	participant, err := h.watchPartyRepo.GetParticipant(c.Request.Context(), partyID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to verify participant",
			},
		})
		return
	}

	if participant == nil || participant.LeftAt != nil {
		c.JSON(http.StatusForbidden, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "FORBIDDEN",
				Message: "Not a participant of this watch party",
			},
		})
		return
	}

	// Parse request body
	var req models.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	// Create message
	message := &models.WatchPartyMessage{
		ID:           uuid.New(),
		WatchPartyID: partyID,
		UserID:       userID,
		Message:      req.Message,
	}

	err = h.watchPartyRepo.CreateMessage(c.Request.Context(), message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to create message",
			},
		})
		return
	}

	// Track chat event (non-blocking, log errors but don't fail request)
	if err := h.analyticsRepo.TrackWatchPartyEvent(c.Request.Context(), partyID, &userID, "chat", nil); err != nil {
		log.Printf("Failed to track chat event for party %s, user %s: %v", partyID, userID, err)
	}

	c.JSON(http.StatusCreated, StandardResponse{
		Success: true,
		Data:    message,
	})
}

// GetMessages handles GET /api/v1/watch-parties/:id/messages
func (h *WatchPartyHandler) GetMessages(c *gin.Context) {
	// Parse party ID
	partyIDStr := c.Param("id")
	partyID, err := uuid.Parse(partyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid party ID",
			},
		})
		return
	}

	// Get optional user ID from context for visibility check
	var userID *uuid.UUID
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(uuid.UUID); ok {
			userID = &uid
		}
	}

	// Get party to check visibility
	party, err := h.watchPartyRepo.GetByID(c.Request.Context(), partyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get watch party",
			},
		})
		return
	}

	if party == nil {
		c.JSON(http.StatusNotFound, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Watch party not found",
			},
		})
		return
	}

	// Check visibility permissions for private parties
	if party.Visibility == "private" {
		if userID == nil {
			c.JSON(http.StatusForbidden, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "FORBIDDEN",
					Message: "Cannot view messages of private party without authentication",
				},
			})
			return
		}
		// Check if user is a participant
		participant, err := h.watchPartyRepo.GetParticipant(c.Request.Context(), partyID, *userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "INTERNAL_ERROR",
					Message: "Failed to verify participant status",
				},
			})
			return
		}
		if participant == nil || participant.LeftAt != nil {
			c.JSON(http.StatusForbidden, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "FORBIDDEN",
					Message: "Not a participant of this private party",
				},
			})
			return
		}
	}

	// Get messages (last 100)
	messages, err := h.watchPartyRepo.GetMessages(c.Request.Context(), partyID, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get messages",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"messages": messages,
			"count":    len(messages),
		},
	})
}

// SendReaction handles POST /api/v1/watch-parties/:id/react
func (h *WatchPartyHandler) SendReaction(c *gin.Context) {
	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "UNAUTHORIZED",
				Message: "Authentication required",
			},
		})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	// Parse party ID
	partyIDStr := c.Param("id")
	partyID, err := uuid.Parse(partyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid party ID",
			},
		})
		return
	}

	// Verify user is a participant
	participant, err := h.watchPartyRepo.GetParticipant(c.Request.Context(), partyID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to verify participant",
			},
		})
		return
	}

	if participant == nil || participant.LeftAt != nil {
		c.JSON(http.StatusForbidden, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "FORBIDDEN",
				Message: "Not a participant of this watch party",
			},
		})
		return
	}

	// Parse request body
	var req models.SendReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	// Create reaction
	reaction := &models.WatchPartyReaction{
		ID:             uuid.New(),
		WatchPartyID:   partyID,
		UserID:         userID,
		Emoji:          req.Emoji,
		VideoTimestamp: req.VideoTimestamp,
	}

	err = h.watchPartyRepo.CreateReaction(c.Request.Context(), reaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to create reaction",
			},
		})
		return
	}

	// Track reaction event (non-blocking, log errors but don't fail request)
	if err := h.analyticsRepo.TrackWatchPartyEvent(c.Request.Context(), partyID, &userID, "reaction", nil); err != nil {
		log.Printf("Failed to track reaction event for party %s, user %s: %v", partyID, userID, err)
	}

	c.JSON(http.StatusCreated, StandardResponse{
		Success: true,
		Data:    reaction,
	})
}

// UpdateWatchPartySettings handles PATCH /api/v1/watch-parties/:id/settings
func (h *WatchPartyHandler) UpdateWatchPartySettings(c *gin.Context) {
	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "UNAUTHORIZED",
				Message: "Authentication required",
			},
		})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	// Parse party ID
	partyIDStr := c.Param("id")
	partyID, err := uuid.Parse(partyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid party ID",
			},
		})
		return
	}

	// Parse request body
	var req models.UpdateWatchPartySettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	// Get watch party to verify host
	party, err := h.watchPartyRepo.GetByID(c.Request.Context(), partyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get watch party",
			},
		})
		return
	}

	if party == nil {
		c.JSON(http.StatusNotFound, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Watch party not found",
			},
		})
		return
	}

	// Verify user is the host
	if party.HostUserID != userID {
		c.JSON(http.StatusForbidden, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "FORBIDDEN",
				Message: "Only the host can update party settings",
			},
		})
		return
	}

	// Hash password if provided
	var hashedPassword *string
	if req.Password != nil {
		if *req.Password != "" {
			// Non-empty password: hash it
			hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
			if err != nil {
				c.JSON(http.StatusInternalServerError, StandardResponse{
					Success: false,
					Error: &ErrorInfo{
						Code:    "INTERNAL_ERROR",
						Message: "Failed to hash password",
					},
				})
				return
			}
			hashStr := string(hash)
			hashedPassword = &hashStr
		} else {
			// Empty string means remove password (set to NULL in database)
			hashedPassword = nil
		}
	}
	// If req.Password is nil, hashedPassword remains nil and won't be updated

	// Update settings
	err = h.watchPartyRepo.UpdateSettings(c.Request.Context(), partyID, req.Visibility, hashedPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to update settings",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"message": "Settings updated successfully",
		},
	})
}

// GetWatchPartyHistory handles GET /api/v1/watch-parties/history
func (h *WatchPartyHandler) GetWatchPartyHistory(c *gin.Context) {
	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "UNAUTHORIZED",
				Message: "Authentication required",
			},
		})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	// Parse pagination parameters
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := (page - 1) * limit

	// Get history
	history, totalCount, err := h.watchPartyRepo.GetHistory(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get watch party history",
			},
		})
		return
	}

	totalPages := (totalCount + limit - 1) / limit

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"history": history,
			"pagination": gin.H{
				"page":        page,
				"limit":       limit,
				"total_count": totalCount,
				"total_pages": totalPages,
			},
		},
	})
}

// GetWatchPartyAnalytics handles GET /api/v1/watch-parties/:id/analytics
func (h *WatchPartyHandler) GetWatchPartyAnalytics(c *gin.Context) {
	// Parse party ID
	partyIDStr := c.Param("id")
	partyID, err := uuid.Parse(partyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid party ID",
			},
		})
		return
	}

	// Get optional user ID from context for permission check
	var userID *uuid.UUID
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(uuid.UUID); ok {
			userID = &uid
		}
	}

	// Get watch party to verify access
	party, err := h.watchPartyRepo.GetByID(c.Request.Context(), partyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get watch party",
			},
		})
		return
	}

	if party == nil {
		c.JSON(http.StatusNotFound, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Watch party not found",
			},
		})
		return
	}

	// Check if user is the host or has appropriate permissions based on visibility
	if party.Visibility == "private" {
		// Private parties: only host can view analytics
		if userID == nil || party.HostUserID != *userID {
			c.JSON(http.StatusForbidden, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "FORBIDDEN",
					Message: "Only the host can view analytics for private parties",
				},
			})
			return
		}
	} else if party.Visibility == "friends" {
		// Friends-only parties: only host or participants can view analytics
		if userID == nil {
			c.JSON(http.StatusForbidden, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "FORBIDDEN",
					Message: "Authentication required to view analytics for friends-only parties",
				},
			})
			return
		}
		// Check if user is the host or a participant
		if party.HostUserID != *userID {
			participant, err := h.watchPartyRepo.GetParticipant(c.Request.Context(), partyID, *userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, StandardResponse{
					Success: false,
					Error: &ErrorInfo{
						Code:    "INTERNAL_ERROR",
						Message: "Failed to verify participant status",
					},
				})
				return
			}
			if participant == nil {
				c.JSON(http.StatusForbidden, StandardResponse{
					Success: false,
					Error: &ErrorInfo{
						Code:    "FORBIDDEN",
						Message: "Only the host or participants can view analytics for friends-only parties",
					},
				})
				return
			}
		}
	}
	// Public parties: anyone authenticated can view analytics

	// Get analytics
	analytics, err := h.analyticsRepo.GetWatchPartyAnalytics(c.Request.Context(), partyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get analytics",
			},
		})
		return
	}

	// Get realtime viewer count
	currentViewers, err := h.analyticsRepo.GetRealtimeViewerCount(c.Request.Context(), partyID)
	if err != nil {
		currentViewers = 0 // Don't fail if we can't get current viewers
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"party_id":             analytics.PartyID,
			"unique_viewers":       analytics.UniqueViewers,
			"peak_concurrent":      analytics.PeakConcurrentViewers,
			"current_viewers":      currentViewers,
			"avg_duration_seconds": analytics.AvgDurationSeconds,
			"chat_messages":        analytics.ChatMessages,
			"reactions":            analytics.Reactions,
			"total_engagement":     analytics.ChatMessages + analytics.Reactions,
		},
	})
}

// GetUserWatchPartyStats handles GET /api/v1/users/:id/watch-party-stats
func (h *WatchPartyHandler) GetUserWatchPartyStats(c *gin.Context) {
	// Parse user ID
	userIDStr := c.Param("id")
	targetUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid user ID",
			},
		})
		return
	}

	// Get requesting user ID from context
	requestingUserIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "UNAUTHORIZED",
				Message: "Authentication required",
			},
		})
		return
	}

	requestingUserID, ok := requestingUserIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	// Only allow users to view their own stats
	if targetUserID != requestingUserID {
		c.JSON(http.StatusForbidden, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "FORBIDDEN",
				Message: "You can only view your own watch party statistics",
			},
		})
		return
	}

	// Get host stats
	stats, err := h.analyticsRepo.GetHostStats(c.Request.Context(), targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get watch party stats",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    stats,
	})
}

// GetPublicWatchParties handles GET /api/v1/watch-parties/public
func (h *WatchPartyHandler) GetPublicWatchParties(c *gin.Context) {
	// Parse pagination parameters
	limit := 20
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get public parties
	parties, totalCount, err := h.watchPartyRepo.GetPublicParties(c.Request.Context(), limit, offset)
	if err != nil {
		log.Printf("Error getting public parties: %v", err)
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get public parties",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"parties":     parties,
			"total_count": totalCount,
			"limit":       limit,
			"offset":      offset,
		},
	})
}

// GetTrendingWatchParties handles GET /api/v1/watch-parties/trending
func (h *WatchPartyHandler) GetTrendingWatchParties(c *gin.Context) {
	// Parse limit parameter (default 10, max 50)
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 50 {
			limit = parsedLimit
		}
	}

	// Get trending parties
	parties, err := h.watchPartyRepo.GetTrendingParties(c.Request.Context(), limit)
	if err != nil {
		log.Printf("Error getting trending parties: %v", err)
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get trending parties",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"parties": parties,
		},
	})
}

// KickParticipant handles POST /api/v1/watch-parties/:id/kick
func (h *WatchPartyHandler) KickParticipant(c *gin.Context) {
	// Get user ID from context (must be authenticated)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "UNAUTHORIZED",
				Message: "Authentication required",
			},
		})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	// Get party ID from URL
	partyIDStr := c.Param("id")
	partyID, err := uuid.Parse(partyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid party ID",
			},
		})
		return
	}

	// Parse request body for participant user ID
	var req struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "User ID is required",
			},
		})
		return
	}

	participantID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid participant user ID",
			},
		})
		return
	}

	// Get party to verify host
	party, err := h.watchPartyRepo.GetByID(c.Request.Context(), partyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get watch party",
			},
		})
		return
	}

	if party == nil {
		c.JSON(http.StatusNotFound, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Watch party not found",
			},
		})
		return
	}

	// Check if user is host
	if party.HostUserID != userID {
		c.JSON(http.StatusForbidden, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "FORBIDDEN",
				Message: "Only the host can kick participants",
			},
		})
		return
	}

	// Cannot kick yourself
	if participantID == userID {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Cannot kick yourself",
			},
		})
		return
	}

	// Remove participant
	err = h.watchPartyRepo.RemoveParticipant(c.Request.Context(), partyID, participantID)
	if err != nil {
		log.Printf("Error kicking participant: %v", err)
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to kick participant",
			},
		})
		return
	}

	// Notify via WebSocket - broadcast participant left event
	hub := h.hubManager.GetOrCreateHub(partyID)
	event := &models.WatchPartySyncEvent{
		Type:            "participant-left",
		PartyID:         partyID.String(),
		ServerTimestamp: time.Now().Unix(),
		UserID:          &participantID,
	}
	hub.Broadcast <- event

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"message": "Participant kicked successfully",
		},
	})
}
