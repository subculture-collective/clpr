package handlers

import (
	"strings"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

func TestSearchValidationAcceptsSaidAlone(t *testing.T) {
	h := &SearchHandler{}
	if err := h.validateAndSetDefaults(&models.SearchRequest{}); err == nil {
		t.Fatal("a request without q or said was accepted")
	}
	blank := "   "
	if err := h.validateAndSetDefaults(&models.SearchRequest{Said: &blank}); err == nil {
		t.Fatal("a blank said was accepted")
	}
	long := strings.Repeat("a", 201)
	if err := h.validateAndSetDefaults(&models.SearchRequest{Said: &long}); err == nil {
		t.Fatal("a 201-character said was accepted")
	}
	said := " no way "
	req := &models.SearchRequest{Type: "all", Said: &said}
	if err := h.validateAndSetDefaults(req); err != nil || *req.Said != "no way" || req.Type != "clips" {
		t.Fatalf("said alone = %+v (said %q), %v", req, *req.Said, err)
	}
	req = &models.SearchRequest{Type: "all", Query: "ace", Said: &blank}
	if err := h.validateAndSetDefaults(req); err != nil || req.Said != nil || req.Type != "all" {
		t.Fatalf("q with blank said = %+v, %v", req, err)
	}
}
