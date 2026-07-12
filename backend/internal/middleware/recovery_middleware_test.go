package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

func TestJSONRecoveryMiddleware(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		handler        gin.HandlerFunc
		expectedStatus int
		checkResponse  bool
	}{
		{
			name: "recovers from panic and returns JSON",
			handler: func(c *gin.Context) {
				panic("test panic")
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse:  true,
		},
		{
			name: "allows normal requests to pass through",
			handler: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			},
			expectedStatus: http.StatusOK,
			checkResponse:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router with recovery middleware
			r := gin.New()
			r.Use(requestid.New())
			r.Use(JSONRecoveryMiddleware())
			r.GET("/test", tt.handler)

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			// Serve request
			r.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check that response is valid JSON
			if tt.checkResponse {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("response is not valid JSON: %v", err)
				}

				// Check for required fields in error response
				if _, ok := response["error"]; !ok {
					t.Error("error field missing in response")
				}
				if _, ok := response["code"]; !ok {
					t.Error("code field missing in response")
				}
				if _, ok := response["message"]; !ok {
					t.Error("message field missing in response")
				}
				if response["request_id"] == "" {
					t.Error("request_id field missing in response")
				}

				// Check content type
				contentType := w.Header().Get("Content-Type")
				if contentType != "application/json; charset=utf-8" {
					t.Errorf("expected Content-Type 'application/json; charset=utf-8', got '%s'", contentType)
				}
			}
		})
	}
}
