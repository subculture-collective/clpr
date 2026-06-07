package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	ws "git.subcult.tv/subculture-collective/clpr/internal/websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type StreamerClipRoomHandler struct {
	service         *services.StreamerClipRoomService
	listener        *services.TwitchChatListenerManager
	twitchAuthRepo  *repository.TwitchAuthRepository
	websocketServer *ws.Server
}

func NewStreamerClipRoomHandler(service *services.StreamerClipRoomService, listener *services.TwitchChatListenerManager, twitchAuthRepo *repository.TwitchAuthRepository, websocketServer *ws.Server) *StreamerClipRoomHandler {
	return &StreamerClipRoomHandler{
		service:         service,
		listener:        listener,
		twitchAuthRepo:  twitchAuthRepo,
		websocketServer: websocketServer,
	}
}

func (h *StreamerClipRoomHandler) GetRoom(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	channel := strings.TrimSpace(c.Param("channel"))
	if channel == "" {
		c.JSON(http.StatusBadRequest, StandardResponse{Success: false, Error: &ErrorInfo{Code: "INVALID_REQUEST", Message: "Channel is required"}})
		return
	}

	room, err := h.service.GetOrCreateRoom(c.Request.Context(), userID, channel)
	if err != nil {
		handleStreamerClipRoomError(c, err, "Failed to load streamer clip room")
		return
	}

	c.JSON(http.StatusOK, StandardResponse{Success: true, Data: room})
}

func (h *StreamerClipRoomHandler) StartRoom(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	channel := strings.TrimSpace(c.Param("channel"))
	if channel == "" {
		c.JSON(http.StatusBadRequest, StandardResponse{Success: false, Error: &ErrorInfo{Code: "INVALID_REQUEST", Message: "Channel is required"}})
		return
	}

	if h.listener == nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{Success: false, Error: &ErrorInfo{Code: "INTERNAL_ERROR", Message: "Listener manager unavailable"}})
		return
	}

	var twitchUsername string
	var accessToken string
	if h.twitchAuthRepo != nil {
		auth, err := h.twitchAuthRepo.GetTwitchAuth(c.Request.Context(), userID)
		if err != nil {
			handleStreamerClipRoomError(c, err, "Failed to start streamer clip room")
			return
		}
		if auth != nil {
			twitchUsername = strings.TrimSpace(auth.TwitchUsername)
			accessToken = strings.TrimSpace(auth.AccessToken)
		}
	}
	if accessToken == "" {
		c.JSON(http.StatusBadRequest, StandardResponse{Success: false, Error: &ErrorInfo{Code: "TWITCH_AUTH_REQUIRED", Message: "Connect Twitch chat before starting the listener"}})
		return
	}

	room, err := h.service.StartRoom(c.Request.Context(), userID, channel)
	if err != nil {
		handleStreamerClipRoomError(c, err, "Failed to start streamer clip room")
		return
	}

	if twitchUsername != "" {
		err = h.listener.StartWithUsername(c.Request.Context(), room.ID, room.TwitchChannel, twitchUsername, accessToken)
	} else {
		err = h.listener.Start(c.Request.Context(), room.ID, room.TwitchChannel, accessToken)
	}
	if err != nil {
		_, _ = h.service.StopRoom(c.Request.Context(), userID, channel)
		c.JSON(http.StatusInternalServerError, StandardResponse{Success: false, Error: &ErrorInfo{Code: "INTERNAL_ERROR", Message: "Failed to start Twitch listener"}})
		return
	}

	h.broadcastRoomEvent(room.ID, "room_status_changed", gin.H{"room": room})
	c.JSON(http.StatusOK, StandardResponse{Success: true, Data: room})
}

func (h *StreamerClipRoomHandler) StopRoom(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	channel := strings.TrimSpace(c.Param("channel"))
	if channel == "" {
		c.JSON(http.StatusBadRequest, StandardResponse{Success: false, Error: &ErrorInfo{Code: "INVALID_REQUEST", Message: "Channel is required"}})
		return
	}

	room, err := h.service.StopRoom(c.Request.Context(), userID, channel)
	if err != nil {
		handleStreamerClipRoomError(c, err, "Failed to stop streamer clip room")
		return
	}

	if h.listener != nil {
		h.listener.Stop(room.ID)
	}
	h.broadcastRoomEvent(room.ID, "room_status_changed", gin.H{"room": room})
	c.JSON(http.StatusOK, StandardResponse{Success: true, Data: room})
}

func (h *StreamerClipRoomHandler) ListItems(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	roomID, err := parseRoomIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{Success: false, Error: &ErrorInfo{Code: "INVALID_REQUEST", Message: "Invalid room ID"}})
		return
	}

	status := strings.ToLower(strings.TrimSpace(c.DefaultQuery("status", "all")))
	if status == "" {
		status = "all"
	}
	switch status {
	case "pending", "approved", "rejected", "skipped", "all":
	default:
		c.JSON(http.StatusBadRequest, StandardResponse{Success: false, Error: &ErrorInfo{Code: "INVALID_REQUEST", Message: "Invalid status filter"}})
		return
	}

	items, err := h.service.ListItems(c.Request.Context(), userID, roomID, status)
	if err != nil {
		handleStreamerClipRoomError(c, err, "Failed to list streamer clip room items")
		return
	}

	c.JSON(http.StatusOK, StandardResponse{Success: true, Data: items})
}

func (h *StreamerClipRoomHandler) ApproveItem(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	roomID, itemID, ok := parseRoomAndItemIDs(c)
	if !ok {
		return
	}

	item, err := h.service.ApproveItem(c.Request.Context(), userID, roomID, itemID)
	if err != nil {
		handleStreamerClipRoomError(c, err, "Failed to approve streamer clip room item")
		return
	}

	h.broadcastRoomEvent(roomID, "item_approved", gin.H{"item": item})
	c.JSON(http.StatusOK, StandardResponse{Success: true, Data: item})
}

func (h *StreamerClipRoomHandler) RejectItem(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	roomID, itemID, ok := parseRoomAndItemIDs(c)
	if !ok {
		return
	}

	item, err := h.service.RejectItem(c.Request.Context(), userID, roomID, itemID)
	if err != nil {
		handleStreamerClipRoomError(c, err, "Failed to reject streamer clip room item")
		return
	}

	h.broadcastRoomEvent(roomID, "item_rejected", gin.H{"item": item})
	c.JSON(http.StatusOK, StandardResponse{Success: true, Data: item})
}

func (h *StreamerClipRoomHandler) ReorderItems(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	roomID, err := parseRoomIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{Success: false, Error: &ErrorInfo{Code: "INVALID_REQUEST", Message: "Invalid room ID"}})
		return
	}

	var req models.ReorderStreamerClipRoomItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{Success: false, Error: &ErrorInfo{Code: "INVALID_REQUEST", Message: err.Error()}})
		return
	}

	itemIDs := make([]uuid.UUID, 0, len(req.ItemIDs))
	for _, id := range req.ItemIDs {
		parsedID, parseErr := uuid.Parse(strings.TrimSpace(id))
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, StandardResponse{Success: false, Error: &ErrorInfo{Code: "INVALID_REQUEST", Message: "One or more item IDs are invalid"}})
			return
		}
		itemIDs = append(itemIDs, parsedID)
	}

	if err := h.service.ReorderApprovedItems(c.Request.Context(), userID, roomID, itemIDs); err != nil {
		handleStreamerClipRoomError(c, err, "Failed to reorder streamer clip room items")
		return
	}

	h.broadcastRoomEvent(roomID, "items_reordered", gin.H{"item_ids": req.ItemIDs})
	c.JSON(http.StatusOK, StandardResponse{Success: true, Data: gin.H{"message": "Items reordered successfully"}})
}

func (h *StreamerClipRoomHandler) WebSocket(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	roomID, err := parseRoomIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{Success: false, Error: &ErrorInfo{Code: "INVALID_REQUEST", Message: "Invalid room ID"}})
		return
	}

	room, err := h.service.GetRoomByID(c.Request.Context(), userID, roomID)
	if err != nil {
		handleStreamerClipRoomError(c, err, "Failed to open streamer clip room websocket")
		return
	}

	username := room.TwitchChannel
	if h.twitchAuthRepo != nil {
		if auth, authErr := h.twitchAuthRepo.GetTwitchAuth(c.Request.Context(), userID); authErr == nil && auth != nil {
			if twitchUsername := strings.TrimSpace(auth.TwitchUsername); twitchUsername != "" {
				username = twitchUsername
			}
		}
	}

	if h.websocketServer == nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{Success: false, Error: &ErrorInfo{Code: "INTERNAL_ERROR", Message: "WebSocket server unavailable"}})
		return
	}

	hub := h.websocketServer.GetOrCreateHub(roomID.String())
	responseHeader := http.Header{}
	if subprotocol := c.GetHeader("Sec-WebSocket-Protocol"); subprotocol != "" {
		if strings.HasPrefix(subprotocol, "auth.bearer.") {
			responseHeader.Set("Sec-WebSocket-Protocol", subprotocol)
		}
	}
	conn, err := h.websocketServer.Upgrader.Upgrade(c.Writer, c.Request, responseHeader)
	if err != nil {
		return
	}

	client := ws.NewChatClient(hub, conn, userID, username)
	client.ReadOnly = true
	hub.Register <- client
	go client.WritePump()
	go client.ReadPump()
}

func (h *StreamerClipRoomHandler) broadcastRoomEvent(roomID uuid.UUID, eventType string, data gin.H) {
	if h == nil || h.websocketServer == nil {
		return
	}

	payload, err := json.Marshal(models.StreamerClipRoomEvent{Type: eventType, Data: data})
	if err != nil {
		return
	}

	hub := h.websocketServer.GetOrCreateHub(roomID.String())
	select {
	case hub.Broadcast <- payload:
	default:
	}
}

func currentUserID(c *gin.Context) (uuid.UUID, bool) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, StandardResponse{Success: false, Error: &ErrorInfo{Code: "UNAUTHORIZED", Message: "Authentication required"}})
		return uuid.UUID{}, false
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{Success: false, Error: &ErrorInfo{Code: "INTERNAL_ERROR", Message: "Invalid user ID format"}})
		return uuid.UUID{}, false
	}

	return userID, true
}

func parseRoomAndItemIDs(c *gin.Context) (uuid.UUID, uuid.UUID, bool) {
	roomID, err := parseRoomIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{Success: false, Error: &ErrorInfo{Code: "INVALID_REQUEST", Message: "Invalid room ID"}})
		return uuid.UUID{}, uuid.UUID{}, false
	}

	itemID, err := uuid.Parse(strings.TrimSpace(c.Param("itemId")))
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{Success: false, Error: &ErrorInfo{Code: "INVALID_REQUEST", Message: "Invalid item ID"}})
		return uuid.UUID{}, uuid.UUID{}, false
	}

	return roomID, itemID, true
}

func parseRoomIDParam(c *gin.Context) (uuid.UUID, error) {
	roomIDParam := strings.TrimSpace(c.Param("roomId"))
	if roomIDParam == "" {
		roomIDParam = strings.TrimSpace(c.Param("channel"))
	}
	return uuid.Parse(roomIDParam)
}

func handleStreamerClipRoomError(c *gin.Context, err error, message string) {
	if err == nil {
		return
	}

	status := http.StatusInternalServerError
	code := "INTERNAL_ERROR"
	lowerErr := strings.ToLower(err.Error())
	switch {
	case errors.Is(err, services.ErrStreamerClipRoomForbidden):
		status = http.StatusForbidden
		code = "FORBIDDEN"
	case strings.Contains(lowerErr, "not found"):
		status = http.StatusNotFound
		code = "NOT_FOUND"
	}

	if status == http.StatusInternalServerError {
		c.JSON(status, StandardResponse{Success: false, Error: &ErrorInfo{Code: code, Message: message}})
		return
	}

	errorMessage := "Forbidden"
	if status == http.StatusNotFound {
		errorMessage = "Not found"
	}
	c.JSON(status, StandardResponse{Success: false, Error: &ErrorInfo{Code: code, Message: errorMessage}})
}
