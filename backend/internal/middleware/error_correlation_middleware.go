package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type errorCorrelationWriter struct {
	gin.ResponseWriter
}

func (w *errorCorrelationWriter) WriteHeader(statusCode int) {
	if statusCode >= http.StatusBadRequest && w.Header().Get("X-Error-Code") == "" {
		w.Header().Set("X-Error-Code", defaultHTTPErrorCode(statusCode))
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func defaultHTTPErrorCode(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "INVALID_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusMethodNotAllowed:
		return "METHOD_NOT_ALLOWED"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusRequestEntityTooLarge:
		return "PAYLOAD_TOO_LARGE"
	case http.StatusUnprocessableEntity:
		return "UNPROCESSABLE_ENTITY"
	case http.StatusTooManyRequests:
		return "RATE_LIMITED"
	case http.StatusInternalServerError:
		return "INTERNAL_ERROR"
	case http.StatusBadGateway:
		return "UPSTREAM_ERROR"
	case http.StatusServiceUnavailable:
		return "SERVICE_UNAVAILABLE"
	case http.StatusGatewayTimeout:
		return "UPSTREAM_TIMEOUT"
	default:
		return fmt.Sprintf("HTTP_%d", statusCode)
	}
}

// ErrorCorrelationMiddleware gives every HTTP error a stable, user-safe code.
// A handler may set X-Error-Code before writing to provide a domain-specific code.
func ErrorCorrelationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer = &errorCorrelationWriter{ResponseWriter: c.Writer}
		c.Next()
	}
}
