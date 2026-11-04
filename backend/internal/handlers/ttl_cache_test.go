package handlers

import (
	"errors"
	"testing"
	"time"
)

func TestTTLCacheReusesValuesUntilExpiryAndSkipsFailures(t *testing.T) {
	clock := time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC)
	cache := newTTLCache[string, int](time.Minute)
	cache.now = func() time.Time { return clock }

	loads := 0
	load := func() (int, error) {
		loads++
		return loads * 10, nil
	}

	for i := 0; i < 3; i++ {
		got, err := cache.get("a", load)
		if err != nil || got != 10 {
			t.Fatalf("call %d: got %d, %v", i, got, err)
		}
	}
	if loads != 1 {
		t.Fatalf("expected one load within the window, got %d", loads)
	}

	if got, _ := cache.get("b", load); got != 20 {
		t.Fatalf("keys must be cached independently, got %d", got)
	}

	clock = clock.Add(time.Minute)
	if got, _ := cache.get("a", load); got != 30 {
		t.Fatalf("expected a reload after expiry, got %d", got)
	}

	failing := errors.New("db down")
	clock = clock.Add(time.Minute)
	if _, err := cache.get("a", func() (int, error) { return 0, failing }); !errors.Is(err, failing) {
		t.Fatalf("expected the load error, got %v", err)
	}
	if got, _ := cache.get("a", load); got != 40 {
		t.Fatalf("a failed load must not be cached; got %d", got)
	}
}
