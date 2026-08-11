package utils

import (
	"testing"

	"github.com/google/uuid"
)

func TestEncodeCursor(t *testing.T) {
	clipID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	createdAt := int64(1609459200) // 2021-01-01 00:00:00 UTC

	tests := []struct {
		name      string
		sortKey   string
		sortValue float64
		clipID    uuid.UUID
		createdAt int64
	}{
		{
			name:      "trending score cursor",
			sortKey:   "trending",
			sortValue: 123.456789,
			clipID:    clipID,
			createdAt: createdAt,
		},
		{
			name:      "new cursor",
			sortKey:   "new",
			sortValue: float64(createdAt),
			clipID:    clipID,
			createdAt: createdAt,
		},
		{
			name:      "top cursor",
			sortKey:   "top",
			sortValue: 0.0,
			clipID:    clipID,
			createdAt: createdAt,
		},
		{
			name:      "discussed cursor",
			sortKey:   "discussed",
			sortValue: -42.5,
			clipID:    clipID,
			createdAt: createdAt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := EncodeCursor(tt.sortKey, tt.sortValue, tt.clipID, tt.createdAt)
			if encoded == "" {
				t.Error("EncodeCursor returned empty string")
			}

			// Verify it can be decoded back
			decoded, err := DecodeCursor(encoded)
			if err != nil {
				t.Errorf("Failed to decode cursor: %v", err)
			}

			if decoded.SortKey != tt.sortKey {
				t.Errorf("SortKey mismatch: got %s, want %s", decoded.SortKey, tt.sortKey)
			}

			if decoded.SortValue != tt.sortValue {
				t.Errorf("SortValue mismatch: got %f, want %f", decoded.SortValue, tt.sortValue)
			}

			if decoded.ClipID != tt.clipID.String() {
				t.Errorf("ClipID mismatch: got %s, want %s", decoded.ClipID, tt.clipID.String())
			}

			if decoded.CreatedAt != tt.createdAt {
				t.Errorf("CreatedAt mismatch: got %d, want %d", decoded.CreatedAt, tt.createdAt)
			}
		})
	}
}

func TestEncodeCursorWithShuffleSeed(t *testing.T) {
	clipID := uuid.New()
	seed := uuid.NewString()
	encoded := EncodeCursorWithShuffleSeed("trending", 0.9123456789012345, clipID, 1609459200, seed)

	decoded, err := DecodeCursor(encoded)
	if err != nil {
		t.Fatalf("decode seeded cursor: %v", err)
	}
	if decoded.ShuffleSeed != seed {
		t.Fatalf("shuffle seed = %q, want %q", decoded.ShuffleSeed, seed)
	}
	if decoded.SortValue != 0.9123456789012345 {
		t.Fatalf("sort value lost precision: got %.17g", decoded.SortValue)
	}
}

func TestDecodeCursor(t *testing.T) {
	clipID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	createdAt := int64(1609459200)

	tests := []struct {
		name      string
		cursor    string
		wantError bool
		want      *Cursor
	}{
		{
			name:      "valid cursor",
			cursor:    EncodeCursor("trending", 123.456, clipID, createdAt),
			wantError: false,
			want: &Cursor{
				SortKey:   "trending",
				SortValue: 123.456,
				ClipID:    clipID.String(),
				CreatedAt: createdAt,
			},
		},
		{
			name:      "empty cursor",
			cursor:    "",
			wantError: false,
			want:      nil,
		},
		{
			name:      "invalid base64",
			cursor:    "not-valid-base64!@#$",
			wantError: true,
			want:      nil,
		},
		{
			name:      "invalid format - wrong number of parts",
			cursor:    "dHJlbmRpbmdfc2NvcmU6MTIzLjQ1Ng==", // "trending_score:123.456" (only 2 parts)
			wantError: true,
			want:      nil,
		},
		{
			name:      "invalid sort value",
			cursor:    "dHJlbmRpbmdfc2NvcmU6aW52YWxpZDo1NTBlODQwMC1lMjliLTQxZDQtYTcxNi00NDY2NTU0NDAwMDA6MTYwOTQ1OTIwMA==", // "trending_score:invalid:uuid:timestamp"
			wantError: true,
			want:      nil,
		},
		{
			name:      "invalid UUID",
			cursor:    "dHJlbmRpbmdfc2NvcmU6MTIzLjQ1NjppbnZhbGlkLXV1aWQ6MTYwOTQ1OTIwMA==", // "trending_score:123.456:invalid-uuid:timestamp"
			wantError: true,
			want:      nil,
		},
		{
			name:      "invalid timestamp",
			cursor:    "dHJlbmRpbmdfc2NvcmU6MTIzLjQ1Njo1NTBlODQwMC1lMjliLTQxZDQtYTcxNi00NDY2NTU0NDAwMDA6aW52YWxpZA==", // "trending_score:123.456:uuid:invalid"
			wantError: true,
			want:      nil,
		},
		{
			name:      "invalid sort key - not in whitelist",
			cursor:    EncodeCursor("malicious_key", 123.456, clipID, createdAt),
			wantError: true,
			want:      nil,
		},
		{
			name:      "invalid sort key - SQL injection attempt",
			cursor:    EncodeCursor("'; DROP TABLE clips; --", 123.456, clipID, createdAt),
			wantError: true,
			want:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeCursor(tt.cursor)

			if tt.wantError {
				if err == nil {
					t.Error("DecodeCursor expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("DecodeCursor unexpected error: %v", err)
				return
			}

			if tt.want == nil {
				if got != nil {
					t.Errorf("DecodeCursor expected nil, got %v", got)
				}
				return
			}

			if got.SortKey != tt.want.SortKey {
				t.Errorf("SortKey mismatch: got %s, want %s", got.SortKey, tt.want.SortKey)
			}

			if got.SortValue != tt.want.SortValue {
				t.Errorf("SortValue mismatch: got %f, want %f", got.SortValue, tt.want.SortValue)
			}

			if got.ClipID != tt.want.ClipID {
				t.Errorf("ClipID mismatch: got %s, want %s", got.ClipID, tt.want.ClipID)
			}

			if got.CreatedAt != tt.want.CreatedAt {
				t.Errorf("CreatedAt mismatch: got %d, want %d", got.CreatedAt, tt.want.CreatedAt)
			}
		})
	}
}
