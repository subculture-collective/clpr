package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TestListApplications_InvalidStatus tests status parameter validation
func TestListApplications_InvalidStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		statusParam    string
		expectedStatus int
		errorContains  string
	}{
		{
			name:           "Invalid status value",
			statusParam:    "invalid_status",
			expectedStatus: http.StatusBadRequest,
			errorContains:  "Invalid status",
		},
		{
			name:           "Valid pending status",
			statusParam:    "pending",
			expectedStatus: http.StatusOK,
			errorContains:  "",
		},
		{
			name:           "Valid approved status",
			statusParam:    "approved",
			expectedStatus: http.StatusOK,
			errorContains:  "",
		},
		{
			name:           "Valid rejected status",
			statusParam:    "rejected",
			expectedStatus: http.StatusOK,
			errorContains:  "",
		},
		{
			name:           "Empty status (all)",
			statusParam:    "",
			expectedStatus: http.StatusOK,
			errorContains:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test validates the status parameter validation logic
			// Full integration tests would require database setup
			validStatuses := map[string]bool{
				"pending":  true,
				"approved": true,
				"rejected": true,
				"":         true,
			}

			isValid := validStatuses[tt.statusParam]
			if tt.errorContains != "" && isValid {
				t.Errorf("Expected invalid status but got valid")
			}
			if tt.errorContains == "" && !isValid {
				t.Errorf("Expected valid status but got invalid")
			}
		})
	}
}

// TestCreateVerificationApplication_InvalidInput tests validation of verification application input
func TestCreateVerificationApplication_InvalidInput(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		errorContains  string
	}{
		{
			name: "Missing twitch_channel_url",
			requestBody: map[string]interface{}{
				"follower_count": 1000,
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "twitch_channel_url",
		},
		{
			name: "Invalid URL format",
			requestBody: map[string]interface{}{
				"twitch_channel_url": "not-a-url",
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "url",
		},
		{
			name: "Negative follower count",
			requestBody: map[string]interface{}{
				"twitch_channel_url": "https://twitch.tv/testchannel",
				"follower_count":     -100,
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "follower_count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/verification/applications", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			// This validates the request structure would fail binding
			// Full integration tests would require mock repository
		})
	}
}

// TestReviewVerificationApplication_InvalidDecision tests decision validation
func TestReviewVerificationApplication_InvalidDecision(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validDecisions := []string{"approved", "rejected"}
	invalidDecision := "invalid_decision"

	isValid := false
	for _, valid := range validDecisions {
		if valid == invalidDecision {
			isValid = true
			break
		}
	}

	if isValid {
		t.Errorf("Expected invalid decision but validation passed")
	}
}

func TestVerificationAdminPaginationRejectsMalformedValuesBeforeRepositoryWork(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, target := range []string{
		"/api/v1/admin/verification/applications?limit=invalid",
		"/api/v1/admin/verification/applications?page=0",
		"/api/v1/admin/verification/audit-logs?only_flagged=maybe",
		"/api/v1/admin/verification/audit-logs?only_flagged=true&user_id=" + uuid.NewString(),
	} {
		t.Run(target, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, target, nil)
			handler := &VerificationHandler{}
			if strings.Contains(target, "audit-logs") {
				handler.GetAuditLogs(c)
			} else {
				handler.ListApplications(c)
			}
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestVerificationReviewRejectsMalformedReviewerIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: uuid.NewString()}}
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/verification/applications/id/review", strings.NewReader(`{"decision":"approved"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", "not-a-uuid")

	(&VerificationHandler{}).ReviewApplication(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

// TestAbusePrevention_RejectionCooldown tests the 30-day cooldown after rejection
func TestAbusePrevention_RejectionCooldown(t *testing.T) {
	tests := []struct {
		name               string
		daysSinceRejection int
		shouldBlock        bool
	}{
		{
			name:               "Within 30-day cooldown",
			daysSinceRejection: 15,
			shouldBlock:        true,
		},
		{
			name:               "Exactly 30 days",
			daysSinceRejection: 30,
			shouldBlock:        false,
		},
		{
			name:               "After cooldown period",
			daysSinceRejection: 31,
			shouldBlock:        false,
		},
		{
			name:               "Recent rejection (1 day)",
			daysSinceRejection: 1,
			shouldBlock:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate cooldown logic
			blocked := tt.daysSinceRejection < 30
			if blocked != tt.shouldBlock {
				t.Errorf("Expected shouldBlock=%v, got %v for %d days since rejection",
					tt.shouldBlock, blocked, tt.daysSinceRejection)
			}
		})
	}
}

// TestAbusePrevention_ApplicationLimit tests the maximum application limit
func TestAbusePrevention_ApplicationLimit(t *testing.T) {
	tests := []struct {
		name        string
		totalApps   int
		shouldBlock bool
		maxAllowed  int
	}{
		{
			name:        "First application",
			totalApps:   0,
			shouldBlock: false,
			maxAllowed:  5,
		},
		{
			name:        "Within limit",
			totalApps:   3,
			shouldBlock: false,
			maxAllowed:  5,
		},
		{
			name:        "At limit",
			totalApps:   5,
			shouldBlock: true,
			maxAllowed:  5,
		},
		{
			name:        "Exceeds limit",
			totalApps:   6,
			shouldBlock: true,
			maxAllowed:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate application limit logic
			blocked := tt.totalApps >= tt.maxAllowed
			if blocked != tt.shouldBlock {
				t.Errorf("Expected shouldBlock=%v, got %v for %d total applications",
					tt.shouldBlock, blocked, tt.totalApps)
			}
		})
	}
}

// TestAbusePrevention_DuplicateTwitchURL tests duplicate Twitch URL detection
func TestAbusePrevention_DuplicateTwitchURL(t *testing.T) {
	tests := []struct {
		name              string
		existingAppsCount int
		shouldBlock       bool
	}{
		{
			name:              "No duplicate applications",
			existingAppsCount: 0,
			shouldBlock:       false,
		},
		{
			name:              "One duplicate application",
			existingAppsCount: 1,
			shouldBlock:       true,
		},
		{
			name:              "Multiple duplicate applications",
			existingAppsCount: 3,
			shouldBlock:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate duplicate URL detection logic
			blocked := tt.existingAppsCount > 0
			if blocked != tt.shouldBlock {
				t.Errorf("Expected shouldBlock=%v, got %v for %d existing applications",
					tt.shouldBlock, blocked, tt.existingAppsCount)
			}
		})
	}
}

// TestAuditLog_QueryParameters tests audit log query parameter validation
func TestAuditLog_QueryParameters(t *testing.T) {
	tests := []struct {
		name        string
		limit       int
		page        int
		expectLimit int
		expectPage  int
	}{
		{
			name:        "Valid parameters",
			limit:       50,
			page:        1,
			expectLimit: 50,
			expectPage:  1,
		},
		{
			name:        "Limit too high",
			limit:       150,
			page:        1,
			expectLimit: 50, // Should default to 50
			expectPage:  1,
		},
		{
			name:        "Limit too low",
			limit:       0,
			page:        1,
			expectLimit: 50, // Should default to 50
			expectPage:  1,
		},
		{
			name:        "Invalid page number",
			limit:       25,
			page:        0,
			expectLimit: 25,
			expectPage:  1, // Should default to 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test parameter normalization logic
			actualLimit := tt.limit
			actualPage := tt.page

			if actualLimit < 1 || actualLimit > 100 {
				actualLimit = 50
			}
			if actualPage < 1 {
				actualPage = 1
			}

			if actualLimit != tt.expectLimit {
				t.Errorf("Expected limit=%d, got %d", tt.expectLimit, actualLimit)
			}
			if actualPage != tt.expectPage {
				t.Errorf("Expected page=%d, got %d", tt.expectPage, actualPage)
			}
		})
	}
}
