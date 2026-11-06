package services

import (
	"context"
	"testing"

	"github.com/google/uuid"

	jwtpkg "git.subcult.tv/subculture-collective/clpr/pkg/jwt"
)

func newTokenTypeTestAuthService(t *testing.T) (*AuthService, *jwtpkg.Manager) {
	t.Helper()
	privateKey, _, err := jwtpkg.GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair() error = %v", err)
	}
	manager, err := jwtpkg.NewManager(privateKey)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	// Repositories are nil: a token of the wrong type must be rejected before
	// any repository lookup.
	return NewAuthService(nil, nil, nil, nil, manager), manager
}

func TestGetUserFromTokenRejectsRefreshToken(t *testing.T) {
	service, manager := newTokenTypeTestAuthService(t)
	refresh, err := manager.GenerateRefreshToken(uuid.New())
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}
	if user, err := service.GetUserFromToken(context.Background(), refresh); err == nil {
		t.Fatalf("GetUserFromToken(refresh token) = %v, want error", user)
	}
}

func TestRefreshAccessTokenRejectsAccessToken(t *testing.T) {
	service, manager := newTokenTypeTestAuthService(t)
	access, err := manager.GenerateAccessToken(uuid.New(), "user")
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}
	if _, _, err := service.RefreshAccessToken(context.Background(), access); err == nil {
		t.Fatal("RefreshAccessToken(access token) succeeded, want error")
	}
}
