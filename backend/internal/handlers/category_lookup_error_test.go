package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"github.com/gin-gonic/gin"
)

// A statement timeout while loading the category used to be reported as
// "Category not found" (404), hiding the database failure.
func TestRespondCategoryLookupError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := map[string]struct {
		err  error
		want int
	}{
		"missing category": {repository.ErrCategoryNotFound, http.StatusNotFound},
		"database timeout": {fmt.Errorf("failed to get category by slug: %w", errors.New("canceling statement due to statement timeout")), http.StatusInternalServerError},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			respondCategoryLookupError(c, "reactions-commentary", tc.err)
			if w.Code != tc.want {
				t.Fatalf("status = %d, want %d", w.Code, tc.want)
			}
		})
	}
}
