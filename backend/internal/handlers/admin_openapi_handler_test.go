package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAdminOpenAPIHandlerServesEmbeddedDocument(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/openapi.json", http.NoBody)

	NewAdminOpenAPIHandler().Get(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, recorder.Code)
	}
	var document struct {
		OpenAPI string `json:"openapi"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &document); err != nil {
		t.Fatalf("embedded document is not valid JSON: %v", err)
	}
	if document.OpenAPI == "" {
		t.Fatal("embedded document has no OpenAPI version")
	}
}
