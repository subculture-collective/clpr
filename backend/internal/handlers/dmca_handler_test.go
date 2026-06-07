package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// ==============================================================================
// Public Endpoint Tests - Takedown Notice Submission
// ==============================================================================

func TestSubmitTakedownNotice_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewDMCAHandler(nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/dmca/takedown", bytes.NewReader([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.SubmitTakedownNotice(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid JSON, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSubmitTakedownNotice_MalformedRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewDMCAHandler(nil, nil)

	tests := []struct {
		name string
		req  models.SubmitDMCANoticeRequest
	}{
		{
			name: "Empty complainant name",
			req: models.SubmitDMCANoticeRequest{
				ComplainantName:            "",
				ComplainantEmail:           "test@example.com",
				ComplainantAddress:         "123 Main St",
				Relationship:               "owner",
				CopyrightedWorkDescription: "Test description",
				InfringingURLs:             []string{"https://example.com/clip/test"},
				GoodFaithStatement:         true,
				AccuracyStatement:          true,
				Signature:                  "Test",
			},
		},
		{
			name: "Invalid email format",
			req: models.SubmitDMCANoticeRequest{
				ComplainantName:            "Test User",
				ComplainantEmail:           "not-an-email",
				ComplainantAddress:         "123 Main St",
				Relationship:               "owner",
				CopyrightedWorkDescription: "Test description",
				InfringingURLs:             []string{"https://example.com/clip/test"},
				GoodFaithStatement:         true,
				AccuracyStatement:          true,
				Signature:                  "Test User",
			},
		},
		{
			name: "Invalid relationship value",
			req: models.SubmitDMCANoticeRequest{
				ComplainantName:            "Test User",
				ComplainantEmail:           "test@example.com",
				ComplainantAddress:         "123 Main St",
				Relationship:               "invalid",
				CopyrightedWorkDescription: "Test description",
				InfringingURLs:             []string{"https://example.com/clip/test"},
				GoodFaithStatement:         true,
				AccuracyStatement:          true,
				Signature:                  "Test User",
			},
		},
		{
			name: "Empty infringing URLs",
			req: models.SubmitDMCANoticeRequest{
				ComplainantName:            "Test User",
				ComplainantEmail:           "test@example.com",
				ComplainantAddress:         "123 Main St",
				Relationship:               "owner",
				CopyrightedWorkDescription: "Test description",
				InfringingURLs:             []string{},
				GoodFaithStatement:         true,
				AccuracyStatement:          true,
				Signature:                  "Test User",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.req)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/dmca/takedown", bytes.NewReader(jsonBody))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.SubmitTakedownNotice(c)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d for malformed request, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

// ==============================================================================
// Public Endpoint Tests - Counter-Notice Submission
// ==============================================================================

func TestSubmitCounterNotice_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewDMCAHandler(nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/dmca/counter-notice", bytes.NewReader([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.SubmitCounterNotice(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid JSON, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSubmitCounterNotice_MalformedRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewDMCAHandler(nil, nil)

	tests := []struct {
		name string
		req  models.SubmitDMCACounterNoticeRequest
	}{
		{
			name: "Empty user name",
			req: models.SubmitDMCACounterNoticeRequest{
				DMCANoticeID:          uuid.New(),
				UserName:              "",
				UserEmail:             "test@example.com",
				UserAddress:           "123 Main St",
				RemovedMaterialURL:    "https://example.com/clip/test",
				GoodFaithStatement:    true,
				ConsentToJurisdiction: true,
				ConsentToService:      true,
				Signature:             "Test",
			},
		},
		{
			name: "Invalid email format",
			req: models.SubmitDMCACounterNoticeRequest{
				DMCANoticeID:          uuid.New(),
				UserName:              "Test User",
				UserEmail:             "not-an-email",
				UserAddress:           "123 Main St",
				RemovedMaterialURL:    "https://example.com/clip/test",
				GoodFaithStatement:    true,
				ConsentToJurisdiction: true,
				ConsentToService:      true,
				Signature:             "Test User",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.req)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/dmca/counter-notice", bytes.NewReader(jsonBody))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.SubmitCounterNotice(c)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d for malformed request, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

// ==============================================================================
// User Strikes Endpoint Tests
// ==============================================================================

func TestGetUserStrikes_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewDMCAHandler(nil, nil)

	userID := uuid.New()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users/"+userID.String()+"/dmca-strikes", nil)
	c.Params = gin.Params{{Key: "id", Value: userID.String()}}
	// Not setting user_id to test authorization

	handler.GetUserStrikes(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d for unauthorized request, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestGetUserStrikes_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewDMCAHandler(nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users/invalid-id/dmca-strikes", nil)
	c.Params = gin.Params{{Key: "id", Value: "invalid-id"}}
	c.Set("user_id", uuid.New())

	handler.GetUserStrikes(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid user ID, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetUserStrikes_Forbidden_DifferentUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewDMCAHandler(nil, nil)

	// User trying to view someone else's strikes (non-admin)
	viewingUserID := uuid.New()
	targetUserID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users/"+targetUserID.String()+"/dmca-strikes", nil)
	c.Params = gin.Params{{Key: "id", Value: targetUserID.String()}}
	c.Set("user_id", viewingUserID)
	c.Set("role", "user") // Regular user, not admin

	handler.GetUserStrikes(c)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d for forbidden access, got %d", http.StatusForbidden, w.Code)
	}
}

// ==============================================================================
// Admin Endpoint Tests - Authorization
// ==============================================================================

func TestReviewNotice_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewDMCAHandler(nil, nil)

	noticeID := uuid.New()
	requestBody := models.UpdateDMCANoticeStatusRequest{
		Status: "valid",
	}
	jsonBody, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/admin/dmca/notices/"+noticeID.String()+"/review", bytes.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: noticeID.String()}}
	// Not setting user_id to test authorization

	handler.ReviewNotice(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d for unauthorized request, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestReviewNotice_InvalidNoticeID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewDMCAHandler(nil, nil)

	requestBody := models.UpdateDMCANoticeStatusRequest{
		Status: "valid",
	}
	jsonBody, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/admin/dmca/notices/invalid-id/review", bytes.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "invalid-id"}}
	c.Set("user_id", uuid.New())

	handler.ReviewNotice(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid notice ID, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestReviewNotice_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewDMCAHandler(nil, nil)

	noticeID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/admin/dmca/notices/"+noticeID.String()+"/review", bytes.NewReader([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: noticeID.String()}}
	c.Set("user_id", uuid.New())

	handler.ReviewNotice(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid JSON, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestProcessTakedown_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewDMCAHandler(nil, nil)

	noticeID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/dmca/notices/"+noticeID.String()+"/process", nil)
	c.Params = gin.Params{{Key: "id", Value: noticeID.String()}}
	// Not setting user_id to test authorization

	handler.ProcessTakedown(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d for unauthorized request, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestProcessTakedown_InvalidNoticeID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewDMCAHandler(nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/dmca/notices/invalid-id/process", nil)
	c.Params = gin.Params{{Key: "id", Value: "invalid-id"}}
	c.Set("user_id", uuid.New())

	handler.ProcessTakedown(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid notice ID, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestForwardCounterNotice_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewDMCAHandler(nil, nil)

	counterNoticeID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/dmca/counter-notices/"+counterNoticeID.String()+"/forward", nil)
	c.Params = gin.Params{{Key: "id", Value: counterNoticeID.String()}}
	// Not setting user_id to test authorization

	handler.ForwardCounterNotice(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d for unauthorized request, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestForwardCounterNotice_InvalidCounterNoticeID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewDMCAHandler(nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/dmca/counter-notices/invalid-id/forward", nil)
	c.Params = gin.Params{{Key: "id", Value: "invalid-id"}}
	c.Set("user_id", uuid.New())

	handler.ForwardCounterNotice(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid counter-notice ID, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestListDMCANotices_ValidPaginationParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewDMCAHandler(nil, nil)

	tests := []struct {
		name   string
		query  string
		status int
	}{
		{
			name:   "Default pagination",
			query:  "",
			status: http.StatusOK,
		},
		{
			name:   "Valid page and limit",
			query:  "?page=2&limit=10",
			status: http.StatusOK,
		},
		{
			name:   "Valid status filter",
			query:  "?status=pending",
			status: http.StatusOK,
		},
		{
			name:   "Max limit boundary",
			query:  "?limit=100",
			status: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/dmca/notices"+tt.query, nil)

			handler.ListDMCANotices(c)

			if w.Code != tt.status {
				t.Errorf("Expected status %d, got %d", tt.status, w.Code)
			}
		})
	}
}
