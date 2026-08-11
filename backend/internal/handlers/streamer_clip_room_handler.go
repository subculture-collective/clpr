package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
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
	if !validTwitchChannel(channel) {
		c.JSON(http.StatusBadRequest, StandardResponse{Success: false, Error: &ErrorInfo{Code: "INVALID_REQUEST", Message: "Invalid Twitch channel"}})
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
	if !validTwitchChannel(channel) {
		c.JSON(http.StatusBadRequest, StandardResponse{Success: false, Error: &ErrorInfo{Code: "INVALID_REQUEST", Message: "Invalid Twitch channel"}})
		return
	}

	if h.listener == nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{Success: false, Error: &ErrorInfo{Code: "INTERNAL_ERROR", Message: "Listener manager unavailable"}})
		return
	}

	var broadcasterUsername string
	if h.twitchAuthRepo != nil {
		auth, err := h.twitchAuthRepo.GetTwitchAuth(c.Request.Context(), userID)
		if err != nil {
			handleStreamerClipRoomError(c, err, "Failed to start streamer clip room")
			return
		}
		if auth != nil {
			broadcasterUsername = strings.TrimSpace(auth.TwitchUsername)
			if !hasTwitchScope(auth.Scopes, "channel:bot") {
				c.JSON(http.StatusForbidden, StandardResponse{Success: false, Error: &ErrorInfo{Code: "TWITCH_BOT_AUTH_REQUIRED", Message: "Authorize the Clpr bot for your Twitch channel before starting the listener"}})
				return
			}
			if !hasTwitchScope(auth.Scopes, "channel:manage:clips") {
				c.JSON(http.StatusForbidden, StandardResponse{Success: false, Error: &ErrorInfo{Code: "TWITCH_CLIP_DOWNLOAD_AUTH_REQUIRED", Message: "Authorize Clpr to download clips from your Twitch channel before starting the listener"}})
				return
			}
		}
	}
	if broadcasterUsername == "" {
		c.JSON(http.StatusBadRequest, StandardResponse{Success: false, Error: &ErrorInfo{Code: "TWITCH_AUTH_REQUIRED", Message: "Connect Twitch and authorize the Clpr bot before starting the listener"}})
		return
	}
	if !strings.EqualFold(broadcasterUsername, channel) {
		c.JSON(http.StatusForbidden, StandardResponse{Success: false, Error: &ErrorInfo{Code: "TWITCH_CHANNEL_MISMATCH", Message: "You can only start the bot in the Twitch channel connected to your account"}})
		return
	}

	botUsername := strings.TrimSpace(os.Getenv("TWITCH_BOT_USERNAME"))
	botOAuthToken := strings.TrimSpace(strings.TrimPrefix(os.Getenv("TWITCH_BOT_OAUTH_TOKEN"), "oauth:"))
	if botUsername == "" || botOAuthToken == "" {
		c.JSON(http.StatusServiceUnavailable, StandardResponse{Success: false, Error: &ErrorInfo{Code: "TWITCH_BOT_NOT_CONFIGURED", Message: "The Clpr Twitch bot is not configured"}})
		return
	}

	room, err := h.service.StartRoom(c.Request.Context(), userID, channel)
	if err != nil {
		handleStreamerClipRoomError(c, err, "Failed to start streamer clip room")
		return
	}

	err = h.listener.StartWithUsername(c.Request.Context(), room.ID, room.TwitchChannel, botUsername, botOAuthToken)
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
	if !validTwitchChannel(channel) {
		c.JSON(http.StatusBadRequest, StandardResponse{Success: false, Error: &ErrorInfo{Code: "INVALID_REQUEST", Message: "Invalid Twitch channel"}})
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

func (h *StreamerClipRoomHandler) UpdateSubmissions(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	channel := strings.TrimSpace(c.Param("channel"))
	if !validTwitchChannel(channel) {
		c.JSON(http.StatusBadRequest, StandardResponse{Success: false, Error: &ErrorInfo{Code: "INVALID_REQUEST", Message: "Invalid Twitch channel"}})
		return
	}

	var req models.UpdateStreamerClipRoomSubmissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Enabled == nil {
		c.JSON(http.StatusBadRequest, StandardResponse{Success: false, Error: &ErrorInfo{Code: "INVALID_REQUEST", Message: "enabled is required and duration_minutes must be between 1 and 1440"}})
		return
	}
	if !*req.Enabled {
		req.DurationMinutes = nil
	}

	room, err := h.service.SetSubmissions(c.Request.Context(), userID, channel, *req.Enabled, req.DurationMinutes)
	if err != nil {
		handleStreamerClipRoomError(c, err, "Failed to update clip submissions")
		return
	}
	h.broadcastRoomEvent(room.ID, "submissions_changed", gin.H{"room": room})
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
	seen := make(map[uuid.UUID]struct{}, len(req.ItemIDs))
	for _, id := range req.ItemIDs {
		parsedID, parseErr := uuid.Parse(strings.TrimSpace(id))
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, StandardResponse{Success: false, Error: &ErrorInfo{Code: "INVALID_REQUEST", Message: "One or more item IDs are invalid"}})
			return
		}
		if _, exists := seen[parsedID]; exists {
			c.JSON(http.StatusBadRequest, StandardResponse{Success: false, Error: &ErrorInfo{Code: "INVALID_REQUEST", Message: "Item IDs must be unique"}})
			return
		}
		seen[parsedID] = struct{}{}
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

func validTwitchChannel(channel string) bool {
	if len(channel) < 1 || len(channel) > 25 {
		return false
	}
	for _, char := range channel {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') && char != '_' {
			return false
		}
	}
	return true
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
	case errors.Is(err, services.ErrStreamerClipRoomInactive):
		c.JSON(http.StatusConflict, StandardResponse{Success: false, Error: &ErrorInfo{Code: "LISTENER_INACTIVE", Message: "Start the Twitch listener before opening submissions"}})
		return
	case errors.Is(err, services.ErrStreamerClipRoomForbidden):
		status = http.StatusForbidden
		code = "FORBIDDEN"
	case errors.Is(err, services.ErrStreamerClipRoomNotFound), strings.Contains(lowerErr, "not found"):
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
