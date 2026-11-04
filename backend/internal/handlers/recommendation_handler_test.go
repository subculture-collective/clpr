package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func recommendationTestContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", uuid.New())
	return c, w
}

func TestUpdatePreferencesRejectsEmptyPatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, w := recommendationTestContext(http.MethodPut, "/api/v1/recommendations/preferences", `{}`)
	(&RecommendationHandler{}).UpdatePreferences(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestOnboardingRejectsOversizedPreferenceValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	value := string(bytes.Repeat([]byte("x"), 101))
	c, w := recommendationTestContext(http.MethodPost, "/api/v1/recommendations/onboarding", `{"favorite_games":["`+value+`"]}`)
	(&RecommendationHandler{}).CompleteOnboarding(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
