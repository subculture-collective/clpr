package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

func TestErrorCorrelationMiddlewareAddsStableHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, testCase := range []struct {
		status int
		code   string
	}{
		{http.StatusBadRequest, "INVALID_REQUEST"},
		{http.StatusUnauthorized, "UNAUTHORIZED"},
		{http.StatusNotFound, "NOT_FOUND"},
		{http.StatusTooManyRequests, "RATE_LIMITED"},
		{http.StatusInternalServerError, "INTERNAL_ERROR"},
		{http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE"},
	} {
		t.Run(testCase.code, func(t *testing.T) {
			router := gin.New()
			router.Use(requestid.New(), ErrorCorrelationMiddleware())
			router.GET("/failure", func(c *gin.Context) {
				c.JSON(testCase.status, gin.H{"error": "safe message"})
			})
			request := httptest.NewRequest(http.MethodGet, "/failure", nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if got := response.Header().Get("X-Error-Code"); got != testCase.code {
				t.Fatalf("expected X-Error-Code %q, got %q", testCase.code, got)
			}
			if response.Header().Get("X-Request-ID") == "" {
				t.Fatal("expected X-Request-ID response header")
			}
		})
	}
}

func TestErrorCorrelationMiddlewarePreservesDomainCode(t *testing.T) {
	router := gin.New()
	router.Use(ErrorCorrelationMiddleware())
	router.GET("/failure", func(c *gin.Context) {
		c.Header("X-Error-Code", "CAMPAIGN_NOT_FOUND")
		c.Status(http.StatusNotFound)
	})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/failure", nil))
	if got := response.Header().Get("X-Error-Code"); got != "CAMPAIGN_NOT_FOUND" {
		t.Fatalf("expected domain code to be preserved, got %q", got)
	}
}
