package handlers

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

// SendGridWebhookHandler handles incoming SendGrid webhook events
type SendGridWebhookHandler struct {
	emailLogRepo *repository.EmailLogRepository
	publicKey    *ecdsa.PublicKey
	logger       *utils.StructuredLogger
}

// NewSendGridWebhookHandler creates a new SendGrid webhook handler
func NewSendGridWebhookHandler(emailLogRepo *repository.EmailLogRepository, sendgridPublicKey string) *SendGridWebhookHandler {
	logger := utils.GetLogger()

	var publicKey *ecdsa.PublicKey
	if sendgridPublicKey != "" {
		key, err := parseECDSAPublicKey(sendgridPublicKey)
		if err != nil {
			logger.Warn("Failed to parse SendGrid public key, webhook signature verification will be disabled", map[string]interface{}{"error": err.Error()})
		} else {
			publicKey = key
		}
	} else {
		logger.Warn("SendGrid public key not configured, webhook signature verification is disabled")
	}

	return &SendGridWebhookHandler{
		emailLogRepo: emailLogRepo,
		publicKey:    publicKey,
		logger:       logger,
	}
}

// HandleWebhook processes SendGrid webhook events
// @Summary Handle SendGrid webhook events
// @Description Processes SendGrid webhook events for email delivery tracking
// @Tags webhooks
// @Accept json
// @Produce json
// @Param X-Twilio-Email-Event-Webhook-Signature header string false "SendGrid signature"
// @Param X-Twilio-Email-Event-Webhook-Timestamp header string false "SendGrid timestamp"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/webhooks/sendgrid [post]
func (h *SendGridWebhookHandler) HandleWebhook(c *gin.Context) {
	// Read request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read webhook body", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Verify webhook signature if public key is configured
	if h.publicKey != nil {
		signature := c.GetHeader("X-Twilio-Email-Event-Webhook-Signature")
		timestamp := c.GetHeader("X-Twilio-Email-Event-Webhook-Timestamp")

		if signature == "" || timestamp == "" {
			h.logger.Warn("Missing webhook signature or timestamp headers")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing signature headers"})
			return
		}

		if err := h.verifySignature(body, signature, timestamp); err != nil {
			h.logger.Warn("Invalid webhook signature", map[string]interface{}{"error": err.Error()})
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
			return
		}
	}

	// Parse webhook events (SendGrid sends an array of events)
	var events []models.SendGridWebhookEvent
	if err := json.Unmarshal(body, &events); err != nil {
		h.logger.Error("Failed to parse webhook events", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event format"})
		return
	}

	h.logger.Info("Received SendGrid webhook events", map[string]interface{}{"event_count": len(events)})

	// Process each event
	for _, event := range events {
		if err := h.processEvent(c.Request.Context(), &event); err != nil {
			// Log error but continue processing other events
			h.logger.Error("Failed to process event", err, map[string]interface{}{"event_type": event.Event})
		}
	}

	// Return 200 OK immediately as per requirements
	c.JSON(http.StatusOK, gin.H{"message": "Events processed"})
}

// processEvent processes a single SendGrid webhook event
func (h *SendGridWebhookHandler) processEvent(ctx context.Context, event *models.SendGridWebhookEvent) error {
	// Convert timestamp to time.Time
	eventTime := time.Unix(event.Timestamp, 0)

	// Determine status and event type
	status := h.mapEventToStatus(event.Event)

	// Check if this is a new event or an update to an existing log
	var existingLog *models.EmailLog
	var err error

	if event.SgMessageID != "" {
		existingLog, err = h.emailLogRepo.GetEmailLogByMessageID(ctx, event.SgMessageID)
		if err != nil {
			h.logger.Error("Failed to check for existing email log", err, map[string]interface{}{"message_id": event.SgMessageID})
		}
	}

	// Prepare metadata
	metadataBytes, _ := json.Marshal(event)
	metadataStr := string(metadataBytes)

	if existingLog != nil {
		// Update existing log
		h.updateExistingLog(existingLog, event, status, eventTime)
		existingLog.Metadata = &metadataStr
		existingLog.UpdatedAt = time.Now()

		if err := h.emailLogRepo.UpdateEmailLog(ctx, existingLog); err != nil {
			return fmt.Errorf("failed to update email log: %w", err)
		}

		h.logger.Info("Updated email log", map[string]interface{}{"log_id": existingLog.ID, "event_type": event.Event})
	} else {
		// Create new log entry
		log := &models.EmailLog{
			ID:                uuid.New(),
			Recipient:         event.Email,
			Status:            status,
			EventType:         event.Event,
			SendGridMessageID: &event.SgMessageID,
			SendGridEventID:   &event.SgEventID,
			IPAddress:         &event.IP,
			UserAgent:         &event.UserAgent,
			Metadata:          &metadataStr,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		// Set template from category if available
		if len(event.Category) > 0 {
			log.Template = &event.Category[0]
		}

		// Set event-specific fields
		h.setEventSpecificFields(log, event, eventTime)

		if err := h.emailLogRepo.CreateEmailLog(ctx, log); err != nil {
			return fmt.Errorf("failed to create email log: %w", err)
		}

		h.logger.Info("Created email log", map[string]interface{}{"log_id": log.ID, "event_type": event.Event})
	}

	return nil
}

// mapEventToStatus maps SendGrid event types to our status field
func (h *SendGridWebhookHandler) mapEventToStatus(eventType string) string {
	switch eventType {
	case "processed", "delivered":
		return models.EmailLogStatusDelivered
	case "bounce":
		return models.EmailLogStatusBounce
	case "dropped":
		return models.EmailLogStatusDropped
	case "open":
		return models.EmailLogStatusOpen
	case "click":
		return models.EmailLogStatusClick
	case "spamreport":
		return models.EmailLogStatusSpamReport
	case "unsubscribe", "group_unsubscribe", "group_resubscribe":
		return models.EmailLogStatusUnsubscribe
	case "deferred":
		return models.EmailLogStatusDeferred
	default:
		return eventType
	}
}

// setEventSpecificFields sets fields specific to the event type
func (h *SendGridWebhookHandler) setEventSpecificFields(log *models.EmailLog, event *models.SendGridWebhookEvent, eventTime time.Time) {
	switch event.Event {
	case "processed":
		// 'processed' is when SendGrid has received and validated the message
		log.SentAt = &eventTime
	case "delivered":
		// 'delivered' is when the message was successfully delivered to the receiving server
		log.DeliveredAt = &eventTime
	case "bounce":
		log.BouncedAt = &eventTime
		if event.Type != "" {
			bounceType := event.Type // hard, soft, blocked
			log.BounceType = &bounceType
		}
		if event.Reason != "" {
			log.BounceReason = &event.Reason
		}
	case "dropped":
		log.BouncedAt = &eventTime
		if event.Reason != "" {
			log.BounceReason = &event.Reason
		}
	case "open":
		log.OpenedAt = &eventTime
	case "click":
		log.ClickedAt = &eventTime
		if event.URL != "" {
			log.LinkURL = &event.URL
		}
	case "spamreport":
		log.SpamReportedAt = &eventTime
	case "unsubscribe", "group_unsubscribe":
		log.UnsubscribedAt = &eventTime
	}
}

// updateExistingLog updates an existing log with new event data
func (h *SendGridWebhookHandler) updateExistingLog(log *models.EmailLog, event *models.SendGridWebhookEvent, status string, eventTime time.Time) {
	// Update status to the latest event
	log.Status = status

	// Update event-specific timestamps and fields
	switch event.Event {
	case "delivered":
		log.DeliveredAt = &eventTime
	case "bounce":
		log.BouncedAt = &eventTime
		if event.Type != "" {
			bounceType := event.Type
			log.BounceType = &bounceType
		}
		if event.Reason != "" {
			log.BounceReason = &event.Reason
		}
	case "dropped":
		log.BouncedAt = &eventTime
		if event.Reason != "" {
			log.BounceReason = &event.Reason
		}
	case "open":
		// Only update if not already set (track first open)
		if log.OpenedAt == nil {
			log.OpenedAt = &eventTime
		}
	case "click":
		// Only update if not already set (track first click)
		if log.ClickedAt == nil {
			log.ClickedAt = &eventTime
		}
		if event.URL != "" {
			log.LinkURL = &event.URL
		}
	case "spamreport":
		log.SpamReportedAt = &eventTime
	case "unsubscribe", "group_unsubscribe":
		log.UnsubscribedAt = &eventTime
	}
}

// verifySignature verifies the SendGrid webhook signature using ECDSA
func (h *SendGridWebhookHandler) verifySignature(payload []byte, signature, timestamp string) error {
	if signature == "" {
		h.logger.Warn("Webhook signature verification failed: empty signature")
		return fmt.Errorf("empty signature")
	}

	if timestamp == "" {
		h.logger.Warn("Webhook signature verification failed: empty timestamp")
		return fmt.Errorf("empty timestamp")
	}

	// Verify timestamp is not too old (within 5 minutes)
	timestampInt, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		h.logger.Warn("Webhook signature verification failed: invalid timestamp format", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("invalid timestamp format: %w", err)
	}

	timestampTime := time.Unix(timestampInt, 0)
	age := time.Since(timestampTime)
	if age > 5*time.Minute {
		h.logger.Warn("Webhook signature verification failed: timestamp too old", map[string]interface{}{
			"timestamp_age_seconds": age.Seconds(),
			"max_age_seconds":       300,
		})
		return fmt.Errorf("timestamp too old: %v (max 5 minutes)", age)
	}

	// Future timestamps are also invalid
	if age < 0 {
		h.logger.Warn("Webhook signature verification failed: timestamp in future", map[string]interface{}{
			"timestamp": timestamp,
		})
		return fmt.Errorf("timestamp is in the future")
	}

	// Construct the signed payload according to SendGrid spec: timestamp + payload
	signedPayload := timestamp + string(payload)

	// Hash the signed payload with SHA-256
	hash := sha256.Sum256([]byte(signedPayload))

	// Decode the base64-encoded signature
	sigBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		h.logger.Warn("Webhook signature verification failed: invalid base64 encoding", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("invalid signature encoding: %w", err)
	}

	// Parse the DER-encoded ECDSA signature
	r, s, err := parseECDSASignature(sigBytes)
	if err != nil {
		h.logger.Warn("Webhook signature verification failed: invalid signature format", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("invalid signature format: %w", err)
	}

	// Validate r and s are within the curve order [1, n-1]
	// This is a security best practice to prevent malleability attacks
	n := h.publicKey.Curve.Params().N
	if r.Cmp(n) >= 0 || s.Cmp(n) >= 0 {
		h.logger.Warn("Webhook signature verification failed: r or s exceeds curve order")
		return fmt.Errorf("invalid signature: r or s exceeds curve order")
	}

	// Verify the signature using the public key
	if !ecdsa.Verify(h.publicKey, hash[:], r, s) {
		h.logger.Warn("Webhook signature verification failed: signature mismatch")
		return fmt.Errorf("invalid signature")
	}

	h.logger.Info("Webhook signature verification successful", map[string]interface{}{
		"timestamp": timestamp,
	})
	return nil
}

// parseECDSASignature parses an ECDSA signature into r and s components
// SendGrid can send signatures in two formats:
// 1. Raw format: r || s (32 bytes each for P-256 curve) - total 64 bytes
// 2. DER-encoded ASN.1 format: SEQUENCE { INTEGER r, INTEGER s }
func parseECDSASignature(sigBytes []byte) (*big.Int, *big.Int, error) {
	// Check for DER SEQUENCE tag first (more reliable than length check)
	if len(sigBytes) > 0 && sigBytes[0] == 0x30 {
		// DER-encoded signature
		return parseDERSignature(sigBytes)
	}

	// Assume raw r||s format (32 bytes each for P-256)
	if len(sigBytes) != 64 {
		return nil, nil, fmt.Errorf("invalid signature length: expected 64 bytes for raw format, got %d", len(sigBytes))
	}

	r := new(big.Int).SetBytes(sigBytes[:32])
	s := new(big.Int).SetBytes(sigBytes[32:])

	// Validate r and s are positive (non-zero)
	// This prevents malleability attacks and ensures signature validity
	if r.Sign() <= 0 || s.Sign() <= 0 {
		return nil, nil, fmt.Errorf("invalid signature: r and s must be positive")
	}

	return r, s, nil
}

// parseDERSignature parses a DER-encoded ECDSA signature
// DER format: SEQUENCE { INTEGER r, INTEGER s }
// This is a basic DER parser that handles common cases
func parseDERSignature(sigBytes []byte) (*big.Int, *big.Int, error) {
	if len(sigBytes) < 8 {
		return nil, nil, fmt.Errorf("signature too short: minimum 8 bytes required, got %d", len(sigBytes))
	}

	// Check for SEQUENCE tag
	if sigBytes[0] != 0x30 {
		return nil, nil, fmt.Errorf("invalid DER signature: expected SEQUENCE tag (0x30), got 0x%02x", sigBytes[0])
	}

	// Parse sequence length (supporting only short form for simplicity)
	// Long form would have bit 7 set (value >= 0x80)
	seqLen := int(sigBytes[1])
	if seqLen >= 0x80 {
		return nil, nil, fmt.Errorf("DER long form length not supported (complex DER encoding)")
	}

	// Validate sequence length doesn't exceed buffer
	if 2+seqLen > len(sigBytes) {
		return nil, nil, fmt.Errorf("invalid DER signature: sequence length %d exceeds buffer size %d", seqLen, len(sigBytes)-2)
	}

	idx := 2 // Skip SEQUENCE tag and length

	// Parse r
	if idx >= len(sigBytes) || sigBytes[idx] != 0x02 {
		return nil, nil, fmt.Errorf("invalid DER signature: expected INTEGER tag for r at position %d", idx)
	}
	idx++

	if idx >= len(sigBytes) {
		return nil, nil, fmt.Errorf("invalid DER signature: unexpected end of data")
	}

	rLen := int(sigBytes[idx])
	if rLen >= 0x80 {
		return nil, nil, fmt.Errorf("DER long form length not supported for r")
	}
	if rLen == 0 {
		return nil, nil, fmt.Errorf("invalid DER signature: r length is zero")
	}
	idx++

	// Validate r length doesn't exceed buffer
	if idx+rLen > len(sigBytes) {
		return nil, nil, fmt.Errorf("invalid DER signature: r length %d exceeds remaining buffer %d", rLen, len(sigBytes)-idx)
	}

	r := new(big.Int).SetBytes(sigBytes[idx : idx+rLen])
	idx += rLen

	// Parse s
	if idx >= len(sigBytes) || sigBytes[idx] != 0x02 {
		return nil, nil, fmt.Errorf("invalid DER signature: expected INTEGER tag for s at position %d", idx)
	}
	idx++

	if idx >= len(sigBytes) {
		return nil, nil, fmt.Errorf("invalid DER signature: unexpected end of data")
	}

	sLen := int(sigBytes[idx])
	if sLen >= 0x80 {
		return nil, nil, fmt.Errorf("DER long form length not supported for s")
	}
	if sLen == 0 {
		return nil, nil, fmt.Errorf("invalid DER signature: s length is zero")
	}
	idx++

	// Validate s length doesn't exceed buffer
	if idx+sLen > len(sigBytes) {
		return nil, nil, fmt.Errorf("invalid DER signature: s length %d exceeds remaining buffer %d", sLen, len(sigBytes)-idx)
	}

	s := new(big.Int).SetBytes(sigBytes[idx : idx+sLen])

	// Validate r and s are positive (non-zero) as required by ECDSA
	// This prevents malleability attacks and ensures signature validity
	if r.Sign() <= 0 {
		return nil, nil, fmt.Errorf("invalid DER signature: r must be positive")
	}
	if s.Sign() <= 0 {
		return nil, nil, fmt.Errorf("invalid DER signature: s must be positive")
	}

	return r, s, nil
}

// parseECDSAPublicKey parses an ECDSA public key from PEM or raw base64 DER format.
// SendGrid provides verification keys as raw base64-encoded DER, so both formats are supported.
func parseECDSAPublicKey(publicKeyStr string) (*ecdsa.PublicKey, error) {
	var derBytes []byte

	block, _ := pem.Decode([]byte(publicKeyStr))
	if block != nil {
		derBytes = block.Bytes
	} else {
		// Try raw base64 DER (SendGrid format)
		decoded, err := base64.StdEncoding.DecodeString(publicKeyStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse as PEM or base64 DER: %w", err)
		}
		derBytes = decoded
	}

	pub, err := x509.ParsePKIXPublicKey(derBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	ecdsaPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an ECDSA public key")
	}

	return ecdsaPub, nil
}
