//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
)

// ip_address is an inet column; reading it into *string used to fail, so
// every trusted-device lookup (and therefore "trust this browser") errored.
func TestMFATrustedDevicesRoundTripInetAddress(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	testutil.TruncateTables(t, pool, "mfa_trusted_devices", "users")
	ctx := context.Background()
	repo := NewMFARepository(pool)
	userID := insertAuditTestUser(t, pool)

	ip, agent, name := "203.0.113.7", "test-agent", "Test browser"
	now := time.Now()
	device := &models.MFATrustedDevice{
		UserID: userID, DeviceFingerprint: "fingerprint-1", DeviceName: &name,
		IPAddress: &ip, UserAgent: &agent, TrustedAt: now, ExpiresAt: now.Add(time.Hour), LastUsedAt: now,
	}
	if err := repo.CreateTrustedDevice(ctx, device); err != nil {
		t.Fatalf("create: %v", err)
	}

	found, err := repo.GetTrustedDeviceByFingerprint(ctx, userID, "fingerprint-1")
	if err != nil || found == nil {
		t.Fatalf("lookup = %v, %v", found, err)
	}
	if found.IPAddress == nil || *found.IPAddress != ip {
		t.Fatalf("IPAddress = %v, want %s", found.IPAddress, ip)
	}

	devices, err := repo.GetTrustedDevices(ctx, userID)
	if err != nil || len(devices) != 1 || *devices[0].IPAddress != ip {
		t.Fatalf("list = %+v, %v", devices, err)
	}

	logIP := "198.51.100.4"
	if err := repo.CreateAuditLog(ctx, &models.MFAAuditLog{UserID: &userID, Action: models.MFAActionLoginSuccess, Success: true, IPAddress: &logIP}); err != nil {
		t.Fatalf("audit log: %v", err)
	}
	logs, err := repo.GetAuditLogs(ctx, userID, 10)
	if err != nil || len(logs) != 1 || logs[0].IPAddress == nil || *logs[0].IPAddress != logIP {
		t.Fatalf("audit logs = %+v, %v", logs, err)
	}
}
