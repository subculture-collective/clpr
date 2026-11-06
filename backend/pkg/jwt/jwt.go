package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	// ErrInvalidToken is returned when token validation fails
	ErrInvalidToken = errors.New("invalid token")
	// ErrTokenExpired is returned when token has expired
	ErrTokenExpired = errors.New("token has expired")
	// ErrInvalidSigningMethod is returned when token uses wrong signing method
	ErrInvalidSigningMethod = errors.New("invalid signing method")
)

// Token types carried in the token_type claim.
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// legacyAccessTokenMaxLifetime classifies tokens issued before the
// token_type claim existed: access tokens lived 15 minutes and refresh tokens
// 7 days. Every such token has expired 7 days after deployment.
const legacyAccessTokenMaxLifetime = time.Hour

// Claims represents the JWT claims
type Claims struct {
	UserID    uuid.UUID `json:"sub"`
	Role      string    `json:"role"`
	JTI       string    `json:"jti"` // JWT ID for revocation
	TokenType string    `json:"token_type,omitempty"`
	jwt.RegisteredClaims
}

// tokenType returns the claim's token type, inferring it from the token's
// lifetime when the claim predates the token_type field.
func (c *Claims) tokenType() string {
	if c.TokenType != "" {
		return c.TokenType
	}
	if c.IssuedAt == nil || c.ExpiresAt == nil {
		return ""
	}
	if c.ExpiresAt.Sub(c.IssuedAt.Time) <= legacyAccessTokenMaxLifetime {
		return TokenTypeAccess
	}
	return TokenTypeRefresh
}

// Manager handles JWT token generation and validation
type Manager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// NewManager creates a new JWT manager with RSA keys
func NewManager(privateKeyPEM string) (*Manager, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	var privateKey *rsa.PrivateKey

	// Try PKCS8 first (modern format)
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS1 format as fallback
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	} else {
		// PKCS8 key parsed successfully, convert to RSA
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("key is not RSA private key")
		}
	}

	return &Manager{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}, nil
}

// GenerateAccessToken generates a short-lived access token (15 minutes)
func (m *Manager) GenerateAccessToken(userID uuid.UUID, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		Role:      role,
		JTI:       uuid.New().String(),
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(m.privateKey)
}

// GenerateRefreshToken generates a long-lived refresh token (7 days)
func (m *Manager) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		Role:      "", // Refresh tokens don't need role
		JTI:       uuid.New().String(),
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(m.privateKey)
}

// ValidateToken validates a JWT token and returns the claims
func (m *Manager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, ErrInvalidSigningMethod
		}
		return m.publicKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateAccessToken validates a token and requires it to be an access token.
func (m *Manager) ValidateAccessToken(tokenString string) (*Claims, error) {
	return m.validateTokenType(tokenString, TokenTypeAccess)
}

// ValidateRefreshToken validates a token and requires it to be a refresh token.
func (m *Manager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return m.validateTokenType(tokenString, TokenTypeRefresh)
}

func (m *Manager) validateTokenType(tokenString, tokenType string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.tokenType() != tokenType {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// ExtractClaims extracts claims from token without validation (use carefully)
func (m *Manager) ExtractClaims(tokenString string) (*Claims, error) {
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// HashToken creates a SHA256 hash of the token for storage
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// GenerateRSAKeyPair generates a new RSA key pair for JWT signing
func GenerateRSAKeyPair() (privateKeyPEM, publicKeyPEM string, err error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %w", err)
	}

	// Encode private key to PEM
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	privateKeyPEM = string(pem.EncodeToMemory(privateKeyBlock))

	// Encode public key to PEM
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal public key: %w", err)
	}
	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	publicKeyPEM = string(pem.EncodeToMemory(publicKeyBlock))

	return privateKeyPEM, publicKeyPEM, nil
}
