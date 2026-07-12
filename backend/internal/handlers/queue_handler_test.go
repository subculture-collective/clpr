package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestGetQueue_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create handler with nil service (not accessed for unauthenticated case)
	handler := &QueueHandler{
		queueService: nil,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/queue", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// Don't set user_id to simulate unauthenticated request

	handler.GetQueue(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	if response.Success {
		t.Error("expected Success to be false")
	}

	if response.Error == nil {
		t.Error("expected Error to be set")
	} else if response.Error.Code != "UNAUTHORIZED" {
		t.Errorf("expected error code UNAUTHORIZED, got %s", response.Error.Code)
	}
}

func TestGetQueue_RejectsInvalidLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, limit := range []string{"abc", "0", "501", "-1"} {
		t.Run(limit, func(t *testing.T) {
			handler := &QueueHandler{queueService: nil}
			req := httptest.NewRequest(http.MethodGet, "/api/v1/queue?limit="+limit, http.NoBody)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("user_id", uuid.New())
			handler.GetQueue(c)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
			}
			var response StandardResponse
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if response.Error == nil || response.Error.Code != "INVALID_REQUEST" {
				t.Fatalf("expected INVALID_REQUEST, got %#v", response.Error)
			}
		})
	}
}

func TestConvertToPlaylist_RejectsDatabaseOversizedTitle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &QueueHandler{queueService: nil}
	body, _ := json.Marshal(map[string]interface{}{"title": string(bytes.Repeat([]byte("x"), 101))})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/queue/convert-to-playlist", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uuid.New())
	handler.ConvertToPlaylist(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAddToQueue_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &QueueHandler{
		queueService: nil,
	}

	clipID := uuid.New().String()
	reqBody := map[string]interface{}{
		"clip_id": clipID,
		"at_end":  true,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/queue", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// Don't set user_id to simulate unauthenticated request

	handler.AddToQueue(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	if response.Success {
		t.Error("expected Success to be false")
	}

	if response.Error == nil {
		t.Error("expected Error to be set")
	} else if response.Error.Code != "UNAUTHORIZED" {
		t.Errorf("expected error code UNAUTHORIZED, got %s", response.Error.Code)
	}
}

func TestAddToQueue_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &QueueHandler{
		queueService: nil,
	}

	// Invalid JSON body
	req := httptest.NewRequest(http.MethodPost, "/api/v1/queue", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uuid.New()) // Set user ID

	handler.AddToQueue(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	if response.Success {
		t.Error("expected Success to be false")
	}

	if response.Error == nil {
		t.Error("expected Error to be set")
	} else if response.Error.Code != "INVALID_REQUEST" {
		t.Errorf("expected error code INVALID_REQUEST, got %s", response.Error.Code)
	}
}

func TestRemoveFromQueue_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &QueueHandler{
		queueService: nil,
	}

	itemID := uuid.New().String()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/queue/"+itemID, http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: itemID}}
	// Don't set user_id to simulate unauthenticated request

	handler.RemoveFromQueue(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	if response.Success {
		t.Error("expected Success to be false")
	}

	if response.Error == nil {
		t.Error("expected Error to be set")
	} else if response.Error.Code != "UNAUTHORIZED" {
		t.Errorf("expected error code UNAUTHORIZED, got %s", response.Error.Code)
	}
}

func TestRemoveFromQueue_InvalidItemID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &QueueHandler{
		queueService: nil,
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/queue/invalid-uuid", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}
	c.Set("user_id", uuid.New()) // Set user ID

	handler.RemoveFromQueue(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	if response.Success {
		t.Error("expected Success to be false")
	}

	if response.Error == nil {
		t.Error("expected Error to be set")
	} else if response.Error.Code != "INVALID_REQUEST" {
		t.Errorf("expected error code INVALID_REQUEST, got %s", response.Error.Code)
	}
}

func TestReorderQueue_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &QueueHandler{
		queueService: nil,
	}

	reqBody := map[string]interface{}{
		"item_id":      uuid.New().String(),
		"new_position": 1,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/queue/reorder", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// Don't set user_id to simulate unauthenticated request

	handler.ReorderQueue(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	if response.Success {
		t.Error("expected Success to be false")
	}

	if response.Error == nil {
		t.Error("expected Error to be set")
	} else if response.Error.Code != "UNAUTHORIZED" {
		t.Errorf("expected error code UNAUTHORIZED, got %s", response.Error.Code)
	}
}

func TestClearQueue_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &QueueHandler{
		queueService: nil,
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/queue", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// Don't set user_id to simulate unauthenticated request

	handler.ClearQueue(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	if response.Success {
		t.Error("expected Success to be false")
	}

	if response.Error == nil {
		t.Error("expected Error to be set")
	} else if response.Error.Code != "UNAUTHORIZED" {
		t.Errorf("expected error code UNAUTHORIZED, got %s", response.Error.Code)
	}
}

func TestGetQueueCount_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &QueueHandler{
		queueService: nil,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/queue/count", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// Don't set user_id to simulate unauthenticated request

	handler.GetQueueCount(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	if response.Success {
		t.Error("expected Success to be false")
	}

	if response.Error == nil {
		t.Error("expected Error to be set")
	} else if response.Error.Code != "UNAUTHORIZED" {
		t.Errorf("expected error code UNAUTHORIZED, got %s", response.Error.Code)
	}
}
