package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	qrcode "github.com/skip2/go-qrcode"
	"golang.org/x/crypto/bcrypt"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
)

var (
	// ErrMFANotEnabled is returned when MFA is not enabled for a user
	ErrMFANotEnabled = errors.New("MFA is not enabled for this user")
	// ErrInvalidMFACode is returned when an MFA code is invalid
	ErrInvalidMFACode = errors.New("invalid MFA code")
	// ErrMFAAlreadyEnabled is returned when trying to enable MFA when it's already enabled
	ErrMFAAlreadyEnabled = errors.New("MFA is already enabled")
	// ErrTooManyFailedAttempts is returned when there are too many failed MFA attempts
	ErrTooManyFailedAttempts = errors.New("too many failed attempts, account temporarily locked")
	// ErrNoBackupCodesRemaining is returned when all backup codes have been used
	ErrNoBackupCodesRemaining = errors.New("no backup codes remaining")
)

const (
	// TOTP configuration
	totpPeriod    = 30
	totpSkew      = 1 // Allow ±1 time period (±30 seconds)
	totpDigits    = 6
	totpAlgorithm = otp.AlgorithmSHA1

	// Backup codes configuration
	backupCodeCount  = 10
	backupCodeLength = 8

	// Security configuration
	// maxFailedAttempts = 5
	// lockoutDuration   = 1 * time.Hour
	trustedDeviceTTL = 30 * 24 * time.Hour // 30 days

	// Rate limiting
	rateLimitWindow = 15 * time.Minute
	rateLimitMax    = 5
)

// MFAService handles MFA operations
type MFAService struct {
	cfg           *config.Config
	mfaRepo       *repository.MFARepository
	userRepo      *repository.UserRepository
	emailSvc      *EmailService
	encryptionKey []byte
}

// NewMFAService creates a new MFA service
func NewMFAService(
	cfg *config.Config,
	mfaRepo *repository.MFARepository,
	userRepo *repository.UserRepository,
	emailSvc *EmailService,
) (*MFAService, error) {
	// Get encryption key from config. Accept a raw 32-byte string or a base64-encoded value that decodes to 32 bytes (AES-256).
	var encryptionKey []byte
	rawKey := cfg.Security.MFAEncryptionKey

	// Validate encryption key presence
	if len(rawKey) == 0 {
		return nil, errors.New("MFA_ENCRYPTION_KEY environment variable must be set to enable MFA functionality. Provide a raw 32-byte string or a base64-encoded value that decodes to 32 bytes.")
	}

	// Try to detect and decode base64; otherwise treat as raw
	if strings.ContainsAny(rawKey, "+/=") || len(rawKey) >= 43 {
		decoded, err := base64.StdEncoding.DecodeString(rawKey)
		if err != nil {
			return nil, fmt.Errorf("MFA_ENCRYPTION_KEY must be a valid base64 string encoding 32 bytes; decode error: %v", err)
		}
		encryptionKey = decoded
	} else {
		encryptionKey = []byte(rawKey)
	}

	// Validate length after optional decoding
	if len(encryptionKey) != 32 {
		return nil, fmt.Errorf("MFA_ENCRYPTION_KEY must be exactly 32 bytes for AES-256 after optional base64 decoding, got %d bytes.", len(encryptionKey))
	}

	return &MFAService{
		cfg:           cfg,
		mfaRepo:       mfaRepo,
		userRepo:      userRepo,
		emailSvc:      emailSvc,
		encryptionKey: encryptionKey,
	}, nil
}

// StartEnrollment initiates MFA enrollment for a user
func (s *MFAService) StartEnrollment(ctx context.Context, userID uuid.UUID, email string) (*models.EnrollMFAResponse, error) {
	// Check if MFA is already enabled
	existingMFA, err := s.mfaRepo.GetMFAByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing MFA: %w", err)
	}

	if existingMFA != nil && existingMFA.Enabled {
		return nil, ErrMFAAlreadyEnabled
	}

	// Generate TOTP secret
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Clipper",
		AccountName: email,
		Period:      totpPeriod,
		Digits:      otp.DigitsSix,
		Algorithm:   totpAlgorithm,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	secret := key.Secret()

	// Encrypt the secret for storage
	encryptedSecret, err := s.encryptSecret(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret: %w", err)
	}

	// Generate backup codes
	backupCodes, hashedCodes, err := s.generateBackupCodes()
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	// Create or update MFA record (not enabled yet)
	now := time.Now()
	mfa := &models.UserMFA{
		UserID:                 userID,
		Secret:                 encryptedSecret,
		Enabled:                false, // Not enabled until verified
		BackupCodes:            hashedCodes,
		BackupCodesGeneratedAt: &now,
	}

	if existingMFA == nil {
		err = s.mfaRepo.CreateMFA(ctx, mfa)
	} else {
		err = s.mfaRepo.UpdateMFA(ctx, mfa)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to save MFA config: %w", err)
	}

	// Generate QR code
	qrCodeURL, err := s.generateQRCode(key.URL())
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Create audit log
	_ = s.createAuditLog(ctx, userID, models.MFAActionEnrollStart, true, nil)

	return &models.EnrollMFAResponse{
		Secret:      secret,
		QRCodeURL:   qrCodeURL,
		BackupCodes: backupCodes,
	}, nil
}

// VerifyEnrollment verifies the TOTP code and enables MFA
func (s *MFAService) VerifyEnrollment(ctx context.Context, userID uuid.UUID, code string) error {
	// Get MFA config
	mfa, err := s.mfaRepo.GetMFAByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get MFA config: %w", err)
	}

	if mfa == nil {
		return errors.New("MFA enrollment not started")
	}

	if mfa.Enabled {
		return ErrMFAAlreadyEnabled
	}

	// Decrypt secret
	secret, err := s.decryptSecret(mfa.Secret)
	if err != nil {
		return fmt.Errorf("failed to decrypt secret: %w", err)
	}

	// Verify TOTP code
	valid := totp.Validate(code, secret)
	if !valid {
		_ = s.createAuditLog(ctx, userID, models.MFAActionEnrollFailed, false, nil)
		return ErrInvalidMFACode
	}

	// Enable MFA
	now := time.Now()
	mfa.Enabled = true
	mfa.EnrolledAt = &now

	err = s.mfaRepo.UpdateMFA(ctx, mfa)
	if err != nil {
		return fmt.Errorf("failed to enable MFA: %w", err)
	}

	// Create audit log
	_ = s.createAuditLog(ctx, userID, models.MFAActionEnrollComplete, true, nil)

	// Send notification email
	user, _ := s.userRepo.GetByID(ctx, userID)
	if user != nil && user.Email != nil {
		_ = s.emailSvc.SendMFAEnabledEmail(ctx, *user.Email, user.Username)
	}

	return nil
}

// VerifyTOTP verifies a TOTP code for login
func (s *MFAService) VerifyTOTP(ctx context.Context, userID uuid.UUID, code string, ipAddress, userAgent *string) error {
	// Check rate limiting
	since := time.Now().Add(-rateLimitWindow)
	failedAttempts, err := s.mfaRepo.GetFailedLoginAttempts(ctx, userID, since)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}

	if failedAttempts >= rateLimitMax {
		_ = s.createAuditLog(ctx, userID, models.MFAActionLoginFailed, false, nil)
		return ErrTooManyFailedAttempts
	}

	// Get MFA config
	mfa, err := s.mfaRepo.GetMFAByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get MFA config: %w", err)
	}

	if mfa == nil || !mfa.Enabled {
		return ErrMFANotEnabled
	}

	// Decrypt secret
	secret, err := s.decryptSecret(mfa.Secret)
	if err != nil {
		return fmt.Errorf("failed to decrypt secret: %w", err)
	}

	// Verify TOTP code with time skew
	valid, err := totp.ValidateCustom(code, secret, time.Now(), totp.ValidateOpts{
		Period:    totpPeriod,
		Skew:      totpSkew,
		Digits:    totpDigits,
		Algorithm: totpAlgorithm,
	})

	if err != nil || !valid {
		_ = s.createAuditLog(ctx, userID, models.MFAActionLoginFailed, false, nil)
		return ErrInvalidMFACode
	}

	// Success - create audit log
	_ = s.createAuditLog(ctx, userID, models.MFAActionLoginSuccess, true, nil)

	return nil
}

// VerifyBackupCode verifies and consumes a backup code
func (s *MFAService) VerifyBackupCode(ctx context.Context, userID uuid.UUID, code string) error {
	// Check rate limiting
	since := time.Now().Add(-rateLimitWindow)
	failedAttempts, err := s.mfaRepo.GetFailedLoginAttempts(ctx, userID, since)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}

	if failedAttempts >= rateLimitMax {
		_ = s.createAuditLog(ctx, userID, models.MFAActionBackupCodeFailed, false, nil)
		return ErrTooManyFailedAttempts
	}

	// Get MFA config
	mfa, err := s.mfaRepo.GetMFAByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get MFA config: %w", err)
	}

	if mfa == nil || !mfa.Enabled {
		return ErrMFANotEnabled
	}

	if len(mfa.BackupCodes) == 0 {
		return ErrNoBackupCodesRemaining
	}

	// Normalize code (remove spaces and convert to uppercase)
	code = strings.ToUpper(strings.ReplaceAll(code, " ", ""))

	// Check each hashed backup code
	var matchedHash string
	for _, hashedCode := range mfa.BackupCodes {
		err = bcrypt.CompareHashAndPassword([]byte(hashedCode), []byte(code))
		if err == nil {
			matchedHash = hashedCode
			break
		}
	}

	if matchedHash == "" {
		_ = s.createAuditLog(ctx, userID, models.MFAActionBackupCodeFailed, false, nil)
		return ErrInvalidMFACode
	}

	// Consume the backup code
	err = s.mfaRepo.ConsumeBackupCode(ctx, userID, matchedHash)
	if err != nil {
		return fmt.Errorf("failed to consume backup code: %w", err)
	}

	// Create audit log
	_ = s.createAuditLog(ctx, userID, models.MFAActionBackupCodeUsed, true, nil)

	return nil
}

// RegenerateBackupCodes generates new backup codes
func (s *MFAService) RegenerateBackupCodes(ctx context.Context, userID uuid.UUID, totpCode string) ([]string, error) {
	// Verify TOTP code first
	err := s.VerifyTOTP(ctx, userID, totpCode, nil, nil)
	if err != nil {
		return nil, err
	}

	// Get MFA config
	mfa, err := s.mfaRepo.GetMFAByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get MFA config: %w", err)
	}

	if mfa == nil || !mfa.Enabled {
		return nil, ErrMFANotEnabled
	}

	// Generate new backup codes
	backupCodes, hashedCodes, err := s.generateBackupCodes()
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	// Update MFA config with new codes
	now := time.Now()
	mfa.BackupCodes = hashedCodes
	mfa.BackupCodesGeneratedAt = &now

	err = s.mfaRepo.UpdateMFA(ctx, mfa)
	if err != nil {
		return nil, fmt.Errorf("failed to update backup codes: %w", err)
	}

	// Create audit log
	_ = s.createAuditLog(ctx, userID, models.MFAActionBackupCodeRegen, true, nil)

	// Send notification email
	user, _ := s.userRepo.GetByID(ctx, userID)
	if user != nil && user.Email != nil {
		_ = s.emailSvc.SendMFABackupCodesRegeneratedEmail(ctx, *user.Email, user.Username)
	}

	return backupCodes, nil
}

// DisableMFA disables MFA for a user after verification
func (s *MFAService) DisableMFA(ctx context.Context, userID uuid.UUID, code string) error {
	// Get user for email notification
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Password verification is not applicable in this OAuth-only authentication system.
	// The system uses Twitch OAuth for authentication - users don't have passwords.
	// Only MFA code verification is required to disable MFA.

	// Verify MFA code
	err = s.VerifyTOTP(ctx, userID, code, nil, nil)
	if err != nil {
		return err
	}

	// Delete MFA config
	err = s.mfaRepo.DeleteMFA(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to disable MFA: %w", err)
	}

	// Delete all trusted devices
	_ = s.mfaRepo.DeleteAllTrustedDevices(ctx, userID)

	// Create audit log
	_ = s.createAuditLog(ctx, userID, models.MFAActionDisabled, true, nil)

	// Send notification email
	if user.Email != nil {
		_ = s.emailSvc.SendMFADisabledEmail(ctx, *user.Email, user.Username)
	}

	return nil
}

// GetMFAStatus returns the current MFA status for a user
func (s *MFAService) GetMFAStatus(ctx context.Context, userID uuid.UUID) (*models.MFAStatusResponse, error) {
	mfa, err := s.mfaRepo.GetMFAByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get MFA config: %w", err)
	}

	status := &models.MFAStatusResponse{
		Enabled:              false,
		BackupCodesRemaining: 0,
		TrustedDevicesCount:  0,
		Required:             false,
		InGracePeriod:        false,
	}

	if mfa != nil {
		status.Enabled = mfa.Enabled
		status.EnrolledAt = mfa.EnrolledAt
		status.BackupCodesRemaining = len(mfa.BackupCodes)
		status.Required = mfa.MFARequired
		status.RequiredAt = mfa.MFARequiredAt
		status.GracePeriodEnd = mfa.GracePeriodEnd

		// Check if in grace period
		if mfa.MFARequired && !mfa.Enabled && mfa.GracePeriodEnd != nil {
			status.InGracePeriod = time.Now().Before(*mfa.GracePeriodEnd)
		}

		// Count trusted devices
		devices, _ := s.mfaRepo.GetTrustedDevices(ctx, userID)
		status.TrustedDevicesCount = len(devices)
	}

	return status, nil
}

// CreateTrustedDevice creates a trusted device entry
func (s *MFAService) CreateTrustedDevice(ctx context.Context, userID uuid.UUID, fingerprint, deviceName string, ipAddress, userAgent *string) error {
	now := time.Now()
	device := &models.MFATrustedDevice{
		UserID:            userID,
		DeviceFingerprint: fingerprint,
		DeviceName:        &deviceName,
		IPAddress:         ipAddress,
		UserAgent:         userAgent,
		TrustedAt:         now,
		ExpiresAt:         now.Add(trustedDeviceTTL),
		LastUsedAt:        now,
	}

	err := s.mfaRepo.CreateTrustedDevice(ctx, device)
	if err != nil {
		return fmt.Errorf("failed to create trusted device: %w", err)
	}

	// Create audit log
	_ = s.createAuditLog(ctx, userID, models.MFAActionTrustedDeviceAdded, true, nil)

	return nil
}

// IsTrustedDevice checks if a device is trusted
func (s *MFAService) IsTrustedDevice(ctx context.Context, userID uuid.UUID, fingerprint string) (bool, error) {
	device, err := s.mfaRepo.GetTrustedDeviceByFingerprint(ctx, userID, fingerprint)
	if err != nil {
		return false, fmt.Errorf("failed to check trusted device: %w", err)
	}

	if device == nil {
		return false, nil
	}

	// Update last used timestamp
	_ = s.mfaRepo.UpdateTrustedDeviceLastUsed(ctx, device.ID)

	return true, nil
}

// GetTrustedDevices retrieves all trusted devices for a user
func (s *MFAService) GetTrustedDevices(ctx context.Context, userID uuid.UUID) ([]*models.MFATrustedDevice, error) {
	return s.mfaRepo.GetTrustedDevices(ctx, userID)
}

// RevokeTrustedDevice removes a trusted device
func (s *MFAService) RevokeTrustedDevice(ctx context.Context, userID uuid.UUID, deviceID int) error {
	err := s.mfaRepo.DeleteTrustedDevice(ctx, userID, deviceID)
	if err != nil {
		return fmt.Errorf("failed to revoke trusted device: %w", err)
	}

	// Create audit log
	_ = s.createAuditLog(ctx, userID, models.MFAActionTrustedDeviceRevoked, true, nil)

	return nil
}

// generateBackupCodes generates random backup codes
func (s *MFAService) generateBackupCodes() ([]string, []string, error) {
	codes := make([]string, backupCodeCount)
	hashed := make([]string, backupCodeCount)

	for i := 0; i < backupCodeCount; i++ {
		code, err := generateRandomCode(backupCodeLength)
		if err != nil {
			return nil, nil, err
		}

		// Hash the code for storage
		hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
		if err != nil {
			return nil, nil, err
		}

		codes[i] = code
		hashed[i] = string(hash)
	}

	return codes, hashed, nil
}

// generateRandomCode generates a random alphanumeric code without modulo bias
func generateRandomCode(length int) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const maxByte = 255
	const charsetLen = len(charset)

	// Calculate the largest multiple of charsetLen that fits in a byte
	// to avoid modulo bias
	maxUsableByte := maxByte - (maxByte % charsetLen)

	result := make([]byte, length)
	randomBytes := make([]byte, length*2) // Buffer to reduce rejection rate

	for i := 0; i < length; {
		// Get random bytes
		if _, err := rand.Read(randomBytes); err != nil {
			return "", err
		}

		// Use rejection sampling to avoid modulo bias
		for _, b := range randomBytes {
			if i >= length {
				break
			}
			// Only use byte values that are evenly divisible by charsetLen
			if b <= byte(maxUsableByte) {
				result[i] = charset[b%byte(charsetLen)]
				i++
			}
		}
	}

	return string(result), nil
}

// generateQRCode generates a QR code as a data URL
func (s *MFAService) generateQRCode(content string) (string, error) {
	png, err := qrcode.Encode(content, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(png)
	return "data:image/png;base64," + encoded, nil
}

// encryptSecret encrypts a TOTP secret using AES-256-GCM
func (s *MFAService) encryptSecret(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptSecret decrypts a TOTP secret using AES-256-GCM
func (s *MFAService) decryptSecret(encrypted string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// createAuditLog creates an MFA audit log entry
func (s *MFAService) createAuditLog(ctx context.Context, userID uuid.UUID, action string, success bool, details *string) error {
	log := &models.MFAAuditLog{
		UserID:  &userID,
		Action:  action,
		Success: success,
		Details: details,
	}

	return s.mfaRepo.CreateAuditLog(ctx, log)
}

// GenerateDeviceFingerprint generates a device fingerprint from user agent and IP
func GenerateDeviceFingerprint(userAgent, ipAddress string) string {
	data := fmt.Sprintf("%s:%s", userAgent, ipAddress)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// SetMFARequired marks MFA as required for a user (e.g., when promoted to admin/moderator)
func (s *MFAService) SetMFARequired(ctx context.Context, userID uuid.UUID) error {
	// Set MFA required with 7-day grace period
	gracePeriodDays := 7
	return s.mfaRepo.SetMFARequired(ctx, userID, gracePeriodDays)
}

// CheckMFARequired checks if MFA is required for a user and returns enforcement status
func (s *MFAService) CheckMFARequired(ctx context.Context, userID uuid.UUID) (required bool, enabled bool, inGracePeriod bool, err error) {
	mfa, err := s.mfaRepo.GetMFAByUserID(ctx, userID)
	if err != nil {
		return false, false, false, fmt.Errorf("failed to check MFA requirement: %w", err)
	}

	if mfa == nil {
		return false, false, false, nil
	}

	required = mfa.MFARequired
	enabled = mfa.Enabled

	// Check if in grace period
	if required && !enabled && mfa.GracePeriodEnd != nil {
		inGracePeriod = time.Now().Before(*mfa.GracePeriodEnd)
	}

	return required, enabled, inGracePeriod, nil
}

// IsAdminActionAllowed checks if an admin/moderator action is allowed based on MFA status
func (s *MFAService) IsAdminActionAllowed(ctx context.Context, userID uuid.UUID) (bool, string, error) {
	required, enabled, inGracePeriod, err := s.CheckMFARequired(ctx, userID)
	if err != nil {
		return false, "", err
	}

	// If MFA is not required, allow action
	if !required {
		return true, "", nil
	}

	// If MFA is enabled, allow action
	if enabled {
		return true, "", nil
	}

	// If in grace period, allow action with warning
	if inGracePeriod {
		return true, "MFA setup required: Please enable MFA soon. Your grace period will expire.", nil
	}

	// Grace period expired and MFA not enabled - block action
	return false, "MFA is required for admin actions. Please enable MFA to continue.", nil
}
