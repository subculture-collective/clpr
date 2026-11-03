package main

import (
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

func versionInfo(active int, versions ...int) *services.IndexVersionInfo {
	info := &services.IndexVersionInfo{}
	for _, v := range versions {
		info.AllVersions = append(info.AllVersions, services.IndexVersion{Version: v, IsActive: v == active})
	}
	for i := range info.AllVersions {
		if info.AllVersions[i].IsActive {
			info.ActiveVersion = &info.AllVersions[i]
		}
	}
	return info
}

func TestPreviousVersion(t *testing.T) {
	tests := []struct {
		name    string
		info    *services.IndexVersionInfo
		want    int
		wantErr bool
	}{
		{"active is newest", versionInfo(5, 5, 4, 3), 4, false},
		// A rebuild without swap leaves a newer inactive version; the old
		// AllVersions[1] choice would have picked the active version itself.
		{"newer unswapped version exists", versionInfo(4, 5, 4, 3), 3, false},
		{"gap in versions", versionInfo(7, 7, 2), 2, false},
		{"active is oldest", versionInfo(3, 5, 4, 3), 0, true},
		{"no active version", versionInfo(0, 2, 1), 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := previousVersion(tt.info)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %t", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("version = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestRebuildAllSucceeded(t *testing.T) {
	if rebuildAllSucceeded(nil) {
		t.Fatal("nil result must not count as success")
	}
	if rebuildAllSucceeded(&services.RebuildAllResult{Success: false, Errors: []string{"clips failed"}}) {
		t.Fatal("unsuccessful result must not count as success")
	}
	if !rebuildAllSucceeded(&services.RebuildAllResult{Success: true}) {
		t.Fatal("successful result must count as success")
	}
}
