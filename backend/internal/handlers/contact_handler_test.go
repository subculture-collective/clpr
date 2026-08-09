package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type contactRepositoryStub struct {
	messages []*models.ContactMessage
	total    int
	err      error
	created  *models.ContactMessage
}

func (s *contactRepositoryStub) Create(_ context.Context, message *models.ContactMessage) error {
	s.created = message
	return s.err
}
func (s *contactRepositoryStub) List(context.Context, int, int, string, string) ([]*models.ContactMessage, int, error) {
	return s.messages, s.total, s.err
}
func (s *contactRepositoryStub) UpdateStatus(context.Context, uuid.UUID, string) error {
	return s.err
}

// TestSubmitContactMessage_InvalidInput tests validation of contact form input
func TestSubmitContactMessage_InvalidInput(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		errorContains  string
	}{
		{
			name: "Invalid email format",
			requestBody: map[string]interface{}{
				"email":    "invalid-email",
				"category": "feedback",
				"subject":  "Test subject",
				"message":  "This is a test message with enough characters.",
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "email",
		},
		{
			name: "Missing email",
			requestBody: map[string]interface{}{
				"category": "feedback",
				"subject":  "Test subject",
				"message":  "This is a test message.",
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "error",
		},
		{
			name: "Invalid category",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"category": "invalid_category",
				"subject":  "Test subject",
				"message":  "This is a test message.",
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "category",
		},
		{
			name: "Subject too short",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"category": "feedback",
				"subject":  "ab",
				"message":  "This is a test message with enough characters.",
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "subject",
		},
		{
			name: "Message too short",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"category": "feedback",
				"subject":  "Test subject",
				"message":  "Short",
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "message",
		},
		{
			name: "Missing category",
			requestBody: map[string]interface{}{
				"email":   "test@example.com",
				"subject": "Test subject",
				"message": "This is a test message.",
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler with nil dependencies (validation happens before repo access)
			handler := &ContactHandler{
				contactRepo: nil,
			}

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/contact", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Execute
			handler.SubmitContactMessage(c)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if errorMsg, ok := response["error"].(string); ok {
				if tt.errorContains != "" && len(errorMsg) == 0 {
					t.Errorf("expected error to contain '%s', got empty error", tt.errorContains)
				}
			} else if tt.errorContains != "" {
				t.Errorf("expected error in response, got none")
			}
		})
	}
}

// TestUpdateContactMessageStatus_InvalidInput tests validation of status update input
func TestUpdateContactMessageStatus_InvalidInput(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		messageID      string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name:      "Invalid message ID",
			messageID: "invalid-uuid",
			requestBody: map[string]interface{}{
				"status": "reviewed",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "Invalid status value",
			messageID: uuid.New().String(),
			requestBody: map[string]interface{}{
				"status": "invalid_status",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing status",
			messageID:      uuid.New().String(),
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &ContactHandler{
				contactRepo: nil,
			}

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPut, "/api/v1/admin/contact/"+tt.messageID+"/status", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = []gin.Param{{Key: "id", Value: tt.messageID}}

			handler.UpdateContactMessageStatus(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestContactHandlerLiveResponseContracts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	messageID := uuid.New()
	now := time.Now().UTC()
	repoStub := &contactRepositoryStub{
		messages: []*models.ContactMessage{{
			ID: messageID, UserID: &userID, Email: "viewer@example.com", Category: "feedback",
			Subject: "Release feedback", Message: "The new release workflow is clear.", Status: "pending",
			CreatedAt: now, UpdatedAt: now,
		}},
		total: 1,
	}
	handler := NewContactHandler(repoStub)

	t.Run("submit", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/contact", bytes.NewBufferString(`{"email":"viewer@example.com","category":"feedback","subject":"Release feedback","message":"The new release workflow is clear."}`))
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = request
		ctx.Set("user_id", userID)
		handler.SubmitContactMessage(ctx)
		if recorder.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", recorder.Code, recorder.Body.String())
		}
		if repoStub.created == nil || repoStub.created.UserID == nil || *repoStub.created.UserID != userID {
			t.Fatalf("authenticated submitter was not recorded: %+v", repoStub.created)
		}
		var response struct {
			Message string `json:"message"`
			Status  string `json:"status"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		if response.Status != "success" || response.Message == "" {
			t.Fatalf("unexpected response: %+v", response)
		}
	})

	t.Run("admin list", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/contact?page=1&limit=20&category=feedback&status=pending", nil)
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = request
		handler.GetContactMessages(ctx)
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
		var response struct {
			Data []*models.ContactMessage `json:"data"`
			Meta struct {
				Page       int `json:"page"`
				Limit      int `json:"limit"`
				TotalItems int `json:"total_items"`
				TotalPages int `json:"total_pages"`
			} `json:"meta"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		if len(response.Data) != 1 || response.Data[0].ID != messageID || response.Meta.TotalItems != 1 {
			t.Fatalf("unexpected list response: %+v", response)
		}
	})

	t.Run("status update", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPut, "/api/v1/admin/contact/"+messageID.String()+"/status", bytes.NewBufferString(`{"status":"resolved"}`))
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = request
		ctx.Params = gin.Params{{Key: "id", Value: messageID.String()}}
		handler.UpdateContactMessageStatus(ctx)
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
	})
}

func TestContactHandlerRejectsInvalidContextAndQueries(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewContactHandler(&contactRepositoryStub{})

	t.Run("malformed optional identity does not panic", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/contact", bytes.NewBufferString(`{"email":"viewer@example.com","category":"feedback","subject":"Release feedback","message":"The new release workflow is clear."}`))
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = request
		ctx.Set("user_id", "invalid")
		handler.SubmitContactMessage(ctx)
		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", recorder.Code)
		}
	})

	for _, target := range []string{
		"/api/v1/admin/contact?page=0",
		"/api/v1/admin/contact?limit=101",
		"/api/v1/admin/contact?category=unknown",
		"/api/v1/admin/contact?status=unknown",
	} {
		t.Run(target, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodGet, target, nil)
			handler.GetContactMessages(ctx)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", recorder.Code)
			}
		})
	}

	t.Run("missing status target", func(t *testing.T) {
		missingHandler := NewContactHandler(&contactRepositoryStub{err: repository.ErrContactMessageNotFound})
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/contact/"+uuid.NewString()+"/status", bytes.NewBufferString(`{"status":"resolved"}`))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Params = gin.Params{{Key: "id", Value: uuid.NewString()}}
		missingHandler.UpdateContactMessageStatus(ctx)
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})

}
