package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBillingRetirementPreservesUnrelatedWebhooks(t *testing.T) {
	router := supportedRouter()
	for _, endpoint := range []struct{ method, path string }{
		{http.MethodPost, "/api/v1/webhooks/stripe"},
		{http.MethodGet, "/api/v1/subscriptions/me"},
		{http.MethodGet, "/api/v1/subscriptions/invoices"},
		{http.MethodPost, "/api/v1/subscriptions/cancel"},
		{http.MethodPost, "/api/v1/subscriptions/checkout"},
	} {
		r := httptest.NewRecorder()
		router.ServeHTTP(r, httptest.NewRequest(endpoint.method, endpoint.path, nil))
		if r.Code != http.StatusNotFound {
			t.Errorf("retired endpoint %s %s returned %d", endpoint.method, endpoint.path, r.Code)
		}
	}
	for _, endpoint := range []string{
		"POST /api/v1/webhooks/sendgrid", "GET /api/v1/webhooks/events",
		"POST /api/v1/webhooks", "GET /api/v1/webhooks", "GET /internal/operations/webhooks",
	} {
		found := false
		for _, route := range router.Routes() {
			if route.Method+" "+route.Path == endpoint {
				found = true
			}
		}
		if !found {
			t.Errorf("supported route disappeared: %s", endpoint)
		}
	}
}
