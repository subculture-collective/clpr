package services

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

// fakeChallengeKV is an in-memory MFAChallengeKV with Redis' missing-key
// error semantics.
type fakeChallengeKV struct {
	mu      sync.Mutex
	values  map[string]string
	counts  map[string]int64
	ttls    map[string]time.Duration
	failSet bool
}

func newFakeChallengeKV() *fakeChallengeKV {
	return &fakeChallengeKV{values: map[string]string{}, counts: map[string]int64{}, ttls: map[string]time.Duration{}}
}

func (f *fakeChallengeKV) Set(_ context.Context, key string, value interface{}, ttl time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failSet {
		return errors.New("redis down")
	}
	f.values[key] = value.(string)
	f.ttls[key] = ttl
	return nil
}

func (f *fakeChallengeKV) Get(_ context.Context, key string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	v, ok := f.values[key]
	if !ok {
		return "", goredis.Nil
	}
	return v, nil
}

func (f *fakeChallengeKV) GetDel(_ context.Context, key string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	v, ok := f.values[key]
	if !ok {
		return "", goredis.Nil
	}
	delete(f.values, key)
	return v, nil
}

func (f *fakeChallengeKV) Increment(_ context.Context, key string) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.counts[key]++
	return f.counts[key], nil
}

func (f *fakeChallengeKV) Expire(_ context.Context, key string, ttl time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.ttls[key] = ttl
	return nil
}

func (f *fakeChallengeKV) Delete(_ context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.values, key)
	delete(f.counts, key)
	return nil
}

func TestMFALoginChallengeStoreLifecycle(t *testing.T) {
	ctx := context.Background()
	kv := newFakeChallengeKV()
	store := NewMFALoginChallengeStore(kv)
	userID := uuid.New()

	token, err := store.Create(ctx, userID)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if len(token) < 40 {
		t.Fatalf("token %q is too short to be unguessable", token)
	}
	for key, value := range kv.values {
		if strings.Contains(key, token) || value == token {
			t.Fatalf("raw token stored in Redis: %s=%s", key, value)
		}
		if kv.ttls[key] != MFALoginChallengeTTL {
			t.Fatalf("challenge TTL = %v, want %v", kv.ttls[key], MFALoginChallengeTTL)
		}
	}

	got, err := store.Resolve(ctx, token)
	if err != nil || got != userID {
		t.Fatalf("Resolve = %v, %v; want %v", got, err, userID)
	}
	// Resolve does not consume.
	if got, err := store.Resolve(ctx, token); err != nil || got != userID {
		t.Fatalf("second Resolve = %v, %v", got, err)
	}

	got, err = store.Consume(ctx, token)
	if err != nil || got != userID {
		t.Fatalf("Consume = %v, %v; want %v", got, err, userID)
	}
	if _, err := store.Consume(ctx, token); !errors.Is(err, ErrMFAChallengeNotFound) {
		t.Fatalf("second Consume error = %v, want ErrMFAChallengeNotFound", err)
	}
	if _, err := store.Resolve(ctx, token); !errors.Is(err, ErrMFAChallengeNotFound) {
		t.Fatalf("Resolve after Consume error = %v, want ErrMFAChallengeNotFound", err)
	}
}

func TestMFALoginChallengeStoreExhaustsAfterMaxAttempts(t *testing.T) {
	ctx := context.Background()
	store := NewMFALoginChallengeStore(newFakeChallengeKV())
	token, err := store.Create(ctx, uuid.New())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	for i := 1; i < MFALoginChallengeMaxAttempts; i++ {
		exhausted, err := store.RecordFailure(ctx, token)
		if err != nil || exhausted {
			t.Fatalf("attempt %d: exhausted=%v err=%v", i, exhausted, err)
		}
	}
	exhausted, err := store.RecordFailure(ctx, token)
	if err != nil || !exhausted {
		t.Fatalf("final attempt: exhausted=%v err=%v; want exhausted", exhausted, err)
	}
	if _, err := store.Resolve(ctx, token); !errors.Is(err, ErrMFAChallengeNotFound) {
		t.Fatalf("Resolve after exhaustion error = %v", err)
	}
}

func TestMFALoginChallengeStoreRejectsUnknownTokens(t *testing.T) {
	ctx := context.Background()
	store := NewMFALoginChallengeStore(newFakeChallengeKV())
	for _, token := range []string{"", "not-a-real-token"} {
		if _, err := store.Resolve(ctx, token); !errors.Is(err, ErrMFAChallengeNotFound) {
			t.Errorf("Resolve(%q) error = %v", token, err)
		}
		if _, err := store.Consume(ctx, token); !errors.Is(err, ErrMFAChallengeNotFound) {
			t.Errorf("Consume(%q) error = %v", token, err)
		}
	}
}

func TestMFALoginChallengeStoreReportsStorageFailure(t *testing.T) {
	kv := newFakeChallengeKV()
	kv.failSet = true
	if _, err := NewMFALoginChallengeStore(kv).Create(context.Background(), uuid.New()); err == nil {
		t.Fatal("Create succeeded although storage failed")
	}
}

func TestEnrollmentAccountNameIsNeverEmpty(t *testing.T) {
	id := uuid.New()
	if got := enrollmentAccountName("  ", id); got != id.String() {
		t.Fatalf("blank account name = %q, want user ID", got)
	}
	if got := enrollmentAccountName("streamer", id); got != "streamer" {
		t.Fatalf("account name = %q, want streamer", got)
	}
}
