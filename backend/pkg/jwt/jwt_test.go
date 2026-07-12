package jwt

import (
	"errors"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestGenerateRSAKeyPair(t *testing.T) {
	privateKey, publicKey, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate RSA key pair: %v", err)
	}

	if privateKey == "" {
		t.Error("Private key is empty")
	}

	if publicKey == "" {
		t.Error("Public key is empty")
	}

	// Verify key format
	if len(privateKey) < 100 {
		t.Error("Private key seems too short")
	}

	if len(publicKey) < 100 {
		t.Error("Public key seems too short")
	}
}

func TestNewManager(t *testing.T) {
	privateKey, _, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	manager, err := NewManager(privateKey)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if manager == nil || manager.privateKey == nil {
		t.Error("Manager or private key is nil")
		return
	}

	if manager.publicKey == nil {
		t.Error("Public key is nil")
	}
}

func TestGenerateAccessToken(t *testing.T) {
	privateKey, _, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	manager, err := NewManager(privateKey)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	userID := uuid.New()
	role := "user"

	token, err := manager.GenerateAccessToken(userID, role)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	if token == "" {
		t.Error("Token is empty")
	}

	// Validate the token
	claims, err := manager.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, claims.UserID)
	}

	if claims.Role != role {
		t.Errorf("Expected role %s, got %s", role, claims.Role)
	}

	if claims.JTI == "" {
		t.Error("JTI is empty")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	privateKey, _, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	manager, err := NewManager(privateKey)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	userID := uuid.New()

	token, err := manager.GenerateRefreshToken(userID)
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	if token == "" {
		t.Error("Token is empty")
	}

	// Validate the token
	claims, err := manager.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, claims.UserID)
	}

	if claims.Role != "" {
		t.Errorf("Expected empty role for refresh token, got %s", claims.Role)
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	privateKey, _, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	manager, err := NewManager(privateKey)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	_, err = manager.ValidateToken("invalid.token.here")
	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	privateKey, _, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair() error = %v", err)
	}
	manager, err := NewManager(privateKey)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	now := time.Now()
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodRS256, Claims{
		UserID: uuid.New(),
		Role:   "user",
		JTI:    uuid.NewString(),
		RegisteredClaims: jwtlib.RegisteredClaims{
			IssuedAt:  jwtlib.NewNumericDate(now.Add(-2 * time.Hour)),
			NotBefore: jwtlib.NewNumericDate(now.Add(-2 * time.Hour)),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(-time.Hour)),
		},
	})
	signed, err := token.SignedString(manager.privateKey)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	if _, err := manager.ValidateToken(signed); !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("ValidateToken() error = %v, want %v", err, ErrTokenExpired)
	}
}

func TestHashToken(t *testing.T) {
	token := "test-token"
	hash1 := HashToken(token)

	if hash1 == "" {
		t.Error("Hash is empty")
	}

	// Same token should produce same hash
	hash2 := HashToken(token)
	if hash1 != hash2 {
		t.Error("Same token produced different hashes")
	}

	// Different token should produce different hash
	hash3 := HashToken("different-token")
	if hash1 == hash3 {
		t.Error("Different tokens produced same hash")
	}

	// Hash should be hex string of consistent length (64 chars for SHA256)
	if len(hash1) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(hash1))
	}
}

func TestExtractClaims(t *testing.T) {
	privateKey, _, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	manager, err := NewManager(privateKey)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	userID := uuid.New()
	role := "admin"

	token, err := manager.GenerateAccessToken(userID, role)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Extract claims without validation
	claims, err := manager.ExtractClaims(token)
	if err != nil {
		t.Fatalf("Failed to extract claims: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, claims.UserID)
	}

	if claims.Role != role {
		t.Errorf("Expected role %s, got %s", role, claims.Role)
	}
}

func TestTokenExpiration(t *testing.T) {
	privateKey, _, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	manager, err := NewManager(privateKey)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	userID := uuid.New()

	// Generate access token (15 min)
	accessToken, err := manager.GenerateAccessToken(userID, "user")
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	claims, err := manager.ValidateToken(accessToken)
	if err != nil {
		t.Fatalf("Failed to validate access token: %v", err)
	}

	// Check expiration is approximately 15 minutes from now
	expectedExpiry := time.Now().Add(15 * time.Minute)
	actualExpiry := claims.ExpiresAt.Time

	diff := actualExpiry.Sub(expectedExpiry)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("Access token expiration not as expected. Diff: %v", diff)
	}

	// Generate refresh token (7 days)
	refreshToken, err := manager.GenerateRefreshToken(userID)
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	claims, err = manager.ValidateToken(refreshToken)
	if err != nil {
		t.Fatalf("Failed to validate refresh token: %v", err)
	}

	// Check expiration is approximately 7 days from now
	expectedExpiry = time.Now().Add(7 * 24 * time.Hour)
	actualExpiry = claims.ExpiresAt.Time

	diff = actualExpiry.Sub(expectedExpiry)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("Refresh token expiration not as expected. Diff: %v", diff)
	}
}
