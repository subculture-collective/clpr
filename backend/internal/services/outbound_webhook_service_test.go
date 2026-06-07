package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

func TestGenerateSignature(t *testing.T) {
	service := &OutboundWebhookService{}

	payload := `{"event":"clip.submitted","data":{"submission_id":"123"}}`
	secret := "test-secret"

	signature := service.generateSignature(payload, secret)

	assert.NotEmpty(t, signature)
	assert.Len(t, signature, 64) // SHA256 produces 64 hex characters

	// Test that the same inputs produce the same signature
	signature2 := service.generateSignature(payload, secret)
	assert.Equal(t, signature, signature2)

	// Test that different payloads produce different signatures
	differentPayload := `{"event":"clip.approved","data":{"submission_id":"456"}}`
	signature3 := service.generateSignature(differentPayload, secret)
	assert.NotEqual(t, signature, signature3)
}

func TestOutboundWebhookCalculateNextRetry(t *testing.T) {
	service := &OutboundWebhookService{}

	// Test exponential backoff
	baseTime := time.Now()

	// First retry (attempt 1): 30 * 2^1 = 60 seconds
	nextRetry1 := service.calculateNextRetry(1)
	assert.True(t, nextRetry1.After(baseTime.Add(55*time.Second)))
	assert.True(t, nextRetry1.Before(baseTime.Add(65*time.Second)))

	// Second retry (attempt 2): 30 * 2^2 = 120 seconds = 2 minutes
	nextRetry2 := service.calculateNextRetry(2)
	assert.True(t, nextRetry2.After(baseTime.Add(115*time.Second)))
	assert.True(t, nextRetry2.Before(baseTime.Add(125*time.Second)))

	// Large retry count should be capped at max delay (1 hour)
	nextRetry10 := service.calculateNextRetry(10)
	assert.True(t, nextRetry10.After(baseTime.Add(55*time.Minute)))
	assert.True(t, nextRetry10.Before(baseTime.Add(65*time.Minute)))
}

func TestValidateEvents(t *testing.T) {
	service := &OutboundWebhookService{}

	// Test valid events
	validEvents := []string{models.WebhookEventClipSubmitted, models.WebhookEventClipApproved}
	err := service.validateEvents(validEvents)
	assert.NoError(t, err)

	// Test invalid events
	invalidEvents := []string{models.WebhookEventClipSubmitted, "invalid.event"}
	err = service.validateEvents(invalidEvents)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported event")
}

func TestGenerateSecret(t *testing.T) {
	service := &OutboundWebhookService{}

	secret1, err := service.generateSecret()
	assert.NoError(t, err)
	assert.NotEmpty(t, secret1)
	assert.Len(t, secret1, 64) // 32 bytes = 64 hex characters

	// Test that multiple calls produce different secrets
	secret2, err := service.generateSecret()
	assert.NoError(t, err)
	assert.NotEmpty(t, secret2)
	assert.NotEqual(t, secret1, secret2)
}
