package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

const (
	// MFALoginChallengeTTL is how long a user has to enter their MFA code
	// after the first sign-in step.
	MFALoginChallengeTTL = 5 * time.Minute
	// MFALoginChallengeMaxAttempts is the number of wrong codes accepted for
	// one challenge before the user must sign in again.
	MFALoginChallengeMaxAttempts = 5

	mfaLoginChallengeKeyPrefix = "mfa:login:"
)

// ErrMFAChallengeNotFound is returned when a login challenge is missing,
// expired, already used, or exhausted.
var ErrMFAChallengeNotFound = errors.New("MFA login challenge not found or expired")

// MFAChallengeKV is the subset of the Redis client used by login challenges.
type MFAChallengeKV interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	GetDel(ctx context.Context, key string) (string, error)
	Increment(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
}

// MFALoginChallengeStore holds the pending second step of a sign-in for users
// with MFA enabled. The browser keeps only an opaque random token; Redis
// stores the user ID under a hash of it, so a leaked Redis key cannot be
// replayed as a token.
type MFALoginChallengeStore struct {
	kv  MFAChallengeKV
	ttl time.Duration
}

// NewMFALoginChallengeStore creates a challenge store backed by kv.
func NewMFALoginChallengeStore(kv MFAChallengeKV) *MFALoginChallengeStore {
	return &MFALoginChallengeStore{kv: kv, ttl: MFALoginChallengeTTL}
}

func mfaChallengeKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return mfaLoginChallengeKeyPrefix + hex.EncodeToString(sum[:])
}

func mfaChallengeAttemptsKey(token string) string {
	return mfaChallengeKey(token) + ":attempts"
}

// Create starts a challenge for userID and returns its token.
func (s *MFALoginChallengeStore) Create(ctx context.Context, userID uuid.UUID) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate MFA challenge: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	if err := s.kv.Set(ctx, mfaChallengeKey(token), userID.String(), s.ttl); err != nil {
		return "", fmt.Errorf("store MFA challenge: %w", err)
	}
	return token, nil
}

// Resolve returns the user a challenge belongs to without consuming it.
func (s *MFALoginChallengeStore) Resolve(ctx context.Context, token string) (uuid.UUID, error) {
	if token == "" {
		return uuid.Nil, ErrMFAChallengeNotFound
	}
	value, err := s.kv.Get(ctx, mfaChallengeKey(token))
	return parseChallengeValue(value, err)
}

// RecordFailure counts a wrong code. When the attempt limit is reached the
// challenge is removed and exhausted is true.
func (s *MFALoginChallengeStore) RecordFailure(ctx context.Context, token string) (exhausted bool, err error) {
	attemptsKey := mfaChallengeAttemptsKey(token)
	attempts, err := s.kv.Increment(ctx, attemptsKey)
	if err != nil {
		return false, fmt.Errorf("count MFA attempt: %w", err)
	}
	if attempts == 1 {
		_ = s.kv.Expire(ctx, attemptsKey, s.ttl)
	}
	if attempts < MFALoginChallengeMaxAttempts {
		return false, nil
	}
	_ = s.kv.Delete(ctx, mfaChallengeKey(token))
	_ = s.kv.Delete(ctx, attemptsKey)
	return true, nil
}

// Consume atomically removes a challenge and returns its user, so a code can
// complete a given challenge only once.
func (s *MFALoginChallengeStore) Consume(ctx context.Context, token string) (uuid.UUID, error) {
	if token == "" {
		return uuid.Nil, ErrMFAChallengeNotFound
	}
	value, err := s.kv.GetDel(ctx, mfaChallengeKey(token))
	_ = s.kv.Delete(ctx, mfaChallengeAttemptsKey(token))
	return parseChallengeValue(value, err)
}

func parseChallengeValue(value string, err error) (uuid.UUID, error) {
	if errors.Is(err, goredis.Nil) || (err == nil && value == "") {
		return uuid.Nil, ErrMFAChallengeNotFound
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("read MFA challenge: %w", err)
	}
	userID, parseErr := uuid.Parse(value)
	if parseErr != nil {
		return uuid.Nil, ErrMFAChallengeNotFound
	}
	return userID, nil
}

// LoginRequiresMFA reports whether a sign-in must pass an MFA check: the user
// has MFA enabled and this device is not a trusted device.
func (s *MFAService) LoginRequiresMFA(ctx context.Context, userID uuid.UUID, deviceFingerprint string) (bool, error) {
	mfa, err := s.mfaRepo.GetMFAByUserID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check MFA status: %w", err)
	}
	if mfa == nil || !mfa.Enabled {
		return false, nil
	}
	trusted, err := s.IsTrustedDevice(ctx, userID, deviceFingerprint)
	if err != nil {
		return false, err
	}
	return !trusted, nil
}
