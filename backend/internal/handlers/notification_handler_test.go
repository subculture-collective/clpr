package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type notificationServiceStub struct {
	listFn    func(context.Context, uuid.UUID, string, int, int) ([]models.NotificationWithSource, error)
	markErr   error
	deleteErr error
}

func (s notificationServiceStub) GetUserNotifications(ctx context.Context, id uuid.UUID, filter string, limit, offset int) ([]models.NotificationWithSource, error) {
	return s.listFn(ctx, id, filter, limit, offset)
}
func (notificationServiceStub) GetUnreadCount(context.Context, uuid.UUID) (int, error) { return 3, nil }
func (s notificationServiceStub) MarkAsRead(context.Context, uuid.UUID, uuid.UUID) error {
	return s.markErr
}
func (notificationServiceStub) MarkAllAsRead(context.Context, uuid.UUID) error { return nil }
func (s notificationServiceStub) DeleteNotification(context.Context, uuid.UUID, uuid.UUID) error {
	return s.deleteErr
}
func (notificationServiceStub) GetPreferences(context.Context, uuid.UUID) (*models.NotificationPreferences, error) {
	return nil, errors.New("unused")
}
func (notificationServiceStub) UpdatePreferences(context.Context, *models.NotificationPreferences) error {
	return nil
}
func (notificationServiceStub) ResetPreferences(context.Context, uuid.UUID) (*models.NotificationPreferences, error) {
	return nil, errors.New("unused")
}
func (notificationServiceStub) RegisterDeviceToken(context.Context, uuid.UUID, string, string) error {
	return nil
}
func (notificationServiceStub) UnregisterDeviceToken(context.Context, uuid.UUID, string) error {
	return nil
}

func notificationListContext(target string, identity interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, target, nil)
	if identity != nil {
		c.Set("user_id", identity)
	}
	return c, w
}

func TestListNotificationsRejectsMalformedIdentityAndQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewNotificationHandler(notificationServiceStub{listFn: func(context.Context, uuid.UUID, string, int, int) ([]models.NotificationWithSource, error) {
		t.Fatal("service must not be called")
		return nil, nil
	}}, nil)

	for _, tc := range []struct {
		name, target string
		identity     interface{}
		status       int
	}{
		{"identity", "/api/v1/notifications", "bad", http.StatusUnauthorized},
		{"filter", "/api/v1/notifications?filter=bogus", uuid.New(), http.StatusBadRequest},
		{"limit", "/api/v1/notifications?limit=101", uuid.New(), http.StatusBadRequest},
		{"page", "/api/v1/notifications?page=1000001", uuid.New(), http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, w := notificationListContext(tc.target, tc.identity)
			h.ListNotifications(c)
			if w.Code != tc.status {
				t.Fatalf("expected %d, got %d", tc.status, w.Code)
			}
		})
	}
}

func TestListNotificationsComputesHasMoreFromExtraRow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	h := NewNotificationHandler(notificationServiceStub{listFn: func(_ context.Context, got uuid.UUID, filter string, limit, offset int) ([]models.NotificationWithSource, error) {
		if got != userID || filter != "unread" || limit != 3 || offset != 2 {
			t.Fatalf("unexpected query: %s %s %d %d", got, filter, limit, offset)
		}
		return []models.NotificationWithSource{{}, {}, {}}, nil
	}}, nil)
	c, w := notificationListContext("/api/v1/notifications?filter=unread&limit=2&page=2", userID)
	h.ListNotifications(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if body := w.Body.String(); body == "" || !containsAll(body, `"has_more":true`, `"limit":2`, `"unread_count":3`) {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestNotificationMutationsReturnNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uuid.New()
	for _, tc := range []struct {
		name, method string
		service      notificationServiceStub
		invoke       func(*NotificationHandler, *gin.Context)
	}{
		{"read", http.MethodPut, notificationServiceStub{markErr: repository.ErrNotificationNotFound}, func(h *NotificationHandler, c *gin.Context) { h.MarkAsRead(c) }},
		{"delete", http.MethodDelete, notificationServiceStub{deleteErr: repository.ErrNotificationNotFound}, func(h *NotificationHandler, c *gin.Context) { h.DeleteNotification(c) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(tc.method, "/api/v1/notifications/"+id.String(), nil)
			c.Params = gin.Params{{Key: "id", Value: id.String()}}
			c.Set("user_id", uuid.New())
			tc.invoke(NewNotificationHandler(tc.service, nil), c)
			if w.Code != http.StatusNotFound {
				t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestNotificationPreferenceAndDeviceInputsFailClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, method, target, body string
		invoke                     func(*NotificationHandler, *gin.Context)
	}{
		{"digest", http.MethodPut, "/api/v1/notifications/preferences", `{"email_digest":"sometimes"}`, func(h *NotificationHandler, c *gin.Context) { h.UpdatePreferences(c) }},
		{"platform", http.MethodPost, "/api/v1/notifications/register", `{"device_token":"token","device_platform":"desktop"}`, func(h *NotificationHandler, c *gin.Context) { h.RegisterDeviceToken(c) }},
		{"oversized token", http.MethodDelete, "/api/v1/notifications/unregister", `{"device_token":"` + strings.Repeat("x", 4097) + `"}`, func(h *NotificationHandler, c *gin.Context) { h.UnregisterDeviceToken(c) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(tc.method, tc.target, bytes.NewBufferString(tc.body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Set("user_id", uuid.New())
			tc.invoke(NewNotificationHandler(notificationServiceStub{}, nil), c)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
			}
		})
	}
}

func containsAll(value string, needles ...string) bool {
	for _, needle := range needles {
		if !strings.Contains(value, needle) {
			return false
		}
	}
	return true
}
