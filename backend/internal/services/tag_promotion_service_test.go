package services_test

import (
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewTagPromotionService verifies the constructor creates a non-nil service.
func TestNewTagPromotionService(t *testing.T) {
	svc := services.NewTagPromotionService(nil)
	require.NotNil(t, svc, "NewTagPromotionService should return a non-nil service")
}

// TestTagPromotionServiceTypes verifies the exported error sentinels and
// service type are accessible.
func TestTagPromotionServiceTypes(t *testing.T) {
	assert.NotNil(t, services.ErrPromotionNotFound)
	assert.NotNil(t, services.ErrPromotionAlreadyPending)
	assert.NotNil(t, services.ErrPromotionNotPending)
	assert.NotNil(t, services.ErrTagNotFound)
}