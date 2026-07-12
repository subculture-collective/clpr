package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestNewHTTPServerHasBoundedTimeouts(t *testing.T) {
	srv := newHTTPServer("18088", http.NewServeMux())
	if srv.Addr != ":18088" {
		t.Fatalf("Addr = %q, want :18088", srv.Addr)
	}
	for name, value := range map[string]time.Duration{
		"ReadHeaderTimeout": srv.ReadHeaderTimeout,
		"ReadTimeout":       srv.ReadTimeout,
		"WriteTimeout":      srv.WriteTimeout,
		"IdleTimeout":       srv.IdleTimeout,
	} {
		if value <= 0 {
			t.Errorf("%s must be bounded, got %v", name, value)
		}
	}
	if srv.MaxHeaderBytes <= 0 || srv.MaxHeaderBytes > 1<<20 {
		t.Errorf("MaxHeaderBytes = %d, want 1..1MiB", srv.MaxHeaderBytes)
	}
}

func TestHTTPServerShutdownDrainsInflightRequest(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		once.Do(func() { close(started) })
		<-release
		w.WriteHeader(http.StatusNoContent)
	})

	server := httptest.NewUnstartedServer(handler)
	server.Config = newHTTPServer("0", handler)
	server.Start()
	defer server.Close()

	requestDone := make(chan error, 1)
	go func() {
		response, err := http.Get(server.URL) // #nosec G107 -- loopback test server URL.
		if err == nil {
			response.Body.Close()
		}
		requestDone <- err
	}()
	<-started

	shutdownDone := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		shutdownDone <- server.Config.Shutdown(ctx)
	}()

	select {
	case err := <-shutdownDone:
		t.Fatalf("Shutdown returned before in-flight request completed: %v", err)
	case <-time.After(25 * time.Millisecond):
	}
	close(release)
	if err := <-requestDone; err != nil {
		t.Fatalf("request failed during graceful shutdown: %v", err)
	}
	if err := <-shutdownDone; err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
}
