package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestGetFollowedGamesUsesRegisteredIDParamAndStrictPagination(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users/"+uuid.NewString()+"/games/following?limit=bad", nil)
	c.Params = gin.Params{{Key: "id", Value: uuid.NewString()}}
	(&GameHandler{}).GetFollowedGames(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response["error"] != "limit must be between 1 and 100" {
		t.Fatalf("handler did not consume registered id param: %q", response["error"])
	}
}

