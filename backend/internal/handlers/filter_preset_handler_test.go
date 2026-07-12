package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestFilterPresetRejectsMismatchedRouteUser(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users/"+uuid.NewString()+"/filter-presets", nil)
	c.Params = gin.Params{{Key: "id", Value: uuid.NewString()}}
	c.Set("user_id", uuid.New())
	(&FilterPresetHandler{}).GetUserPresets(c)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestFilterPresetRejectsMalformedIdentityWithoutPanic(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users/"+uuid.NewString()+"/filter-presets", nil)
	c.Params = gin.Params{{Key: "id", Value: uuid.NewString()}}
	c.Set("user_id", "invalid")
	(&FilterPresetHandler{}).GetUserPresets(c)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestFilterPresetRejectsEmptyUpdate(t *testing.T) {
	userID := uuid.New()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/users/"+userID.String()+"/filter-presets/"+uuid.NewString(), bytes.NewBufferString(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: userID.String()}, {Key: "presetId", Value: uuid.NewString()}}
	c.Set("user_id", userID)
	(&FilterPresetHandler{}).UpdatePreset(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestFilterPresetRejectsInvalidNestedFilters(t *testing.T) {
	userID := uuid.New()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/users/"+userID.String()+"/filter-presets", bytes.NewBufferString(`{"name":"bad","filters":{"sort":"unknown"}}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: userID.String()}}
	c.Set("user_id", userID)
	(&FilterPresetHandler{}).CreatePreset(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
