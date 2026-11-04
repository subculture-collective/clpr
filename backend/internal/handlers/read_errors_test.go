package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestWriteReadErrorMapsCancellationAndTimeouts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	statementTimeout := fmt.Errorf("failed to list tags: %w", &pgconn.PgError{Code: pgStatementTimeoutCode, Message: "canceling statement due to statement timeout"})

	tests := []struct {
		name       string
		requestCtx context.Context
		err        error
		wantStatus int
		wantBody   bool
	}{
		{"wrapped context.Canceled", context.Background(), fmt.Errorf("failed to list tags: %w", context.Canceled), StatusClientClosedRequest, false},
		{"request context cancelled", cancelled, errors.New("conn closed"), StatusClientClosedRequest, false},
		{"context deadline", context.Background(), fmt.Errorf("x: %w", context.DeadlineExceeded), http.StatusGatewayTimeout, true},
		{"statement timeout", context.Background(), statementTimeout, http.StatusGatewayTimeout, true},
		{"other error", context.Background(), errors.New("boom"), http.StatusInternalServerError, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/tags", http.NoBody).WithContext(tt.requestCtx)

			writeReadError(ctx, tt.err, "Failed to fetch tags")

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if !ctx.IsAborted() {
				t.Fatal("expected the handler chain to be aborted")
			}
			if got := recorder.Body.Len() > 0; got != tt.wantBody {
				t.Fatalf("body present = %v, want %v (body %q)", got, tt.wantBody, recorder.Body.String())
			}
			if tt.wantStatus == StatusClientClosedRequest && len(ctx.Errors) != 0 {
				t.Fatalf("abandoned requests must not be logged as errors, got %v", ctx.Errors)
			}
			if tt.wantStatus != StatusClientClosedRequest && len(ctx.Errors) == 0 {
				t.Fatal("failed reads must keep the error on the context for the request log")
			}
		})
	}
}
