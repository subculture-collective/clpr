package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func communityTestContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, w
}

func TestOptionalCommunityUserIDRejectsMalformedIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, w := communityTestContext(http.MethodGet, "/", "")
	c.Set("user_id", "not-a-uuid")

	if _, ok := optionalCommunityUserID(c); ok {
		t.Fatal("expected malformed optional identity to be rejected")
	}
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestCommunityPaginationRejectsMalformedValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, query := range []string{"?page=zero", "?page=0", "?limit=zero", "?limit=101"} {
		t.Run(query, func(t *testing.T) {
			c, w := communityTestContext(http.MethodGet, "/"+query, "")
			if _, _, ok := parseCommunityPagination(c, 20); ok {
				t.Fatal("expected malformed pagination to be rejected")
			}
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestListCommunitiesRejectsUnsupportedSort(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &CommunityHandler{}
	c, w := communityTestContext(http.MethodGet, "/api/v1/communities?sort=popular", "")

	h.ListCommunities(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetMembersRejectsUnsupportedRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &CommunityHandler{}
	communityID := uuid.New()
	c, w := communityTestContext(http.MethodGet, "/api/v1/communities/"+communityID.String()+"/members?role=owner", "")
	c.Params = gin.Params{{Key: "id", Value: communityID.String()}}

	h.GetMembers(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateCommunityRejectsEmptyUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &CommunityHandler{}
	communityID := uuid.New()
	c, w := communityTestContext(http.MethodPut, "/api/v1/communities/"+communityID.String(), `{}`)
	c.Params = gin.Params{{Key: "id", Value: communityID.String()}}
	c.Set("user_id", uuid.New())

	h.UpdateCommunity(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCommunityHandlersContainUnavailablePrivateLifecycle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name   string
		method string
		body   string
		call   func(*CommunityHandler, *gin.Context)
	}{
		{
			name: "create", method: http.MethodPost,
			body: `{"name":"Private group","is_public":false}`,
			call: func(h *CommunityHandler, c *gin.Context) { h.CreateCommunity(c) },
		},
		{
			name: "update", method: http.MethodPut,
			body: `{"is_public":false}`,
			call: func(h *CommunityHandler, c *gin.Context) { h.UpdateCommunity(c) },
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := &CommunityHandler{}
			communityID := uuid.NewString()
			c, w := communityTestContext(test.method, "/api/v1/communities/"+communityID, test.body)
			c.Params = gin.Params{{Key: "id", Value: communityID}}
			c.Set("user_id", uuid.New())

			test.call(h, c)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestGetDiscussionValidatesParentCommunityID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &CommunityHandler{}
	c, w := communityTestContext(http.MethodGet, "/api/v1/communities/not-a-uuid/discussions/"+uuid.NewString(), "")
	c.Params = gin.Params{
		{Key: "id", Value: "not-a-uuid"},
		{Key: "discussionId", Value: uuid.NewString()},
	}

	h.GetDiscussion(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, w.Code)
	}
}
