package main

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/config"
)

func TestInitializeJWTManagerFailsClosedOutsideDevelopment(t *testing.T) {
	profiles := []config.ServerConfig{
		{Environment: "production", GinMode: "release"},
		{Environment: "staging", GinMode: "release"},
		{Environment: "production", GinMode: "debug"},
		{Environment: "development", GinMode: "release"},
	}

	for _, profile := range profiles {
		profile := profile
		t.Run(profile.Environment+"/"+profile.GinMode, func(t *testing.T) {
			cfg := &config.Config{Server: profile}
			manager, err := initializeJWTManager(cfg)
			if err == nil || manager != nil {
				t.Fatal("initializeJWTManager() must reject a missing key outside development")
			}
			if !strings.Contains(err.Error(), "JWT_PRIVATE_KEY is required") {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestInitializeJWTManagerDevelopmentLogDoesNotExposeKey(t *testing.T) {
	var output bytes.Buffer
	originalWriter := log.Writer()
	originalFlags := log.Flags()
	log.SetOutput(&output)
	log.SetFlags(0)
	t.Cleanup(func() {
		log.SetOutput(originalWriter)
		log.SetFlags(originalFlags)
	})

	cfg := &config.Config{Server: config.ServerConfig{Environment: "development", GinMode: "debug"}}
	manager, err := initializeJWTManager(cfg)
	if err != nil {
		t.Fatalf("initializeJWTManager() error = %v", err)
	}
	if manager == nil {
		t.Fatal("initializeJWTManager() returned nil manager")
	}

	logged := output.String()
	if strings.Contains(logged, "PRIVATE KEY") || strings.Contains(logged, "PUBLIC KEY") || strings.Contains(logged, "BEGIN RSA") {
		t.Fatalf("development log exposed key material: %s", logged)
	}
	if !strings.Contains(logged, "fingerprint") {
		t.Fatalf("development log should identify the ephemeral key by fingerprint: %s", logged)
	}
}
