package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestTracingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		serviceName string
		path        string
		method      string
	}{
		{
			name:        "GET request with tracing",
			serviceName: "test-service",
			path:        "/api/test",
			method:      "GET",
		},
		{
			name:        "POST request with tracing",
			serviceName: "clpr-backend",
			path:        "/api/v1/clips",
			method:      "POST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test router
			r := gin.New()
			r.Use(TracingMiddleware(tt.serviceName))

			// Add a test handler
			r.Handle(tt.method, tt.path, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Create a test request
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			// Serve the request
			r.ServeHTTP(w, req)

			// Assert the response
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestTracingMiddlewareWithoutTracing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a test router without tracing middleware
	r := gin.New()

	// Add a test handler
	r.GET("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create a test request
	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)
}
