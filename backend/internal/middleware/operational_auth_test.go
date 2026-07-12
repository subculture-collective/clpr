package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequireOperationalToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name   string
		config string
		header string
		want   int
	}{
		{name: "missing configuration fails closed", want: http.StatusUnauthorized},
		{name: "missing credential", config: "service-secret", want: http.StatusUnauthorized},
		{name: "wrong credential", config: "service-secret", header: "Bearer wrong", want: http.StatusUnauthorized},
		{name: "valid credential", config: "service-secret", header: "Bearer service-secret", want: http.StatusNoContent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.GET("/internal", RequireOperationalToken(tt.config), func(c *gin.Context) { c.Status(http.StatusNoContent) })
			req := httptest.NewRequest(http.MethodGet, "/internal", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			response := httptest.NewRecorder()
			r.ServeHTTP(response, req)
			if response.Code != tt.want {
				t.Fatalf("status = %d, want %d", response.Code, tt.want)
			}
		})
	}
}
