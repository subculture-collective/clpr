package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestStripeWebhookRejectsOversizedPayloadBeforeService(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/stripe", bytes.NewBufferString(strings.Repeat("x", (1<<20)+1)))
	c.Request.Header.Set("Stripe-Signature", "valid-shaped")
	(&SubscriptionHandler{}).HandleWebhook(c)
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", w.Code)
	}
}

func TestStripeWebhookRejectsMissingOrOversizedSignature(t *testing.T) {
	for _, signature := range []string{"", strings.Repeat("x", 8193)} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/stripe", bytes.NewBufferString(`{}`))
		if signature != "" {
			c.Request.Header.Set("Stripe-Signature", signature)
		}
		(&SubscriptionHandler{}).HandleWebhook(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	}
}
