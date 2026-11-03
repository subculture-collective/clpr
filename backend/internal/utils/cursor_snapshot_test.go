package utils

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestFullCursorRoundTripsSnapshotWithoutShuffleSeed(t *testing.T) {
	snapshot := time.Date(2026, 11, 2, 21, 4, 5, 123456000, time.UTC)
	created := time.Date(2026, 11, 1, 8, 0, 0, 987654000, time.UTC)
	in := Cursor{
		SortKey:   "rising",
		SortValue: 412,
		ClipID:    uuid.NewString(),
		CreatedAt: CursorTimeValue(created),
		Snapshot:  CursorTimeValue(snapshot),
	}
	out, err := DecodeCursor(EncodeFullCursor(in))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if *out != in {
		t.Fatalf("round trip\n got  %+v\n want %+v", *out, in)
	}
	if !CursorTime(out.Snapshot).Equal(snapshot) || !CursorTime(out.CreatedAt).Equal(created) {
		t.Fatalf("timestamps lost precision: snapshot %v created %v", CursorTime(out.Snapshot), CursorTime(out.CreatedAt))
	}
}

func TestFullCursorRoundTripsSeedAndSnapshot(t *testing.T) {
	in := Cursor{SortKey: "hot", SortValue: -3, ClipID: uuid.NewString(), CreatedAt: 1, ShuffleSeed: uuid.NewString(), Snapshot: 1_800_000_000_000_000}
	out, err := DecodeCursor(EncodeFullCursor(in))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if *out != in {
		t.Fatalf("round trip\n got  %+v\n want %+v", *out, in)
	}
}

func TestEncodeCursorWithoutSnapshotKeepsLegacyFormat(t *testing.T) {
	clipID := uuid.New()
	plain := EncodeCursor("top", 5, clipID, 1609459200)
	full := EncodeFullCursor(Cursor{SortKey: "top", SortValue: 5, ClipID: clipID.String(), CreatedAt: 1609459200})
	if plain != full {
		t.Fatalf("cursors without a snapshot changed format: %q vs %q", plain, full)
	}
}

func TestDecodeCursorRejectsBadSnapshot(t *testing.T) {
	clipID := uuid.NewString()
	for _, raw := range []string{
		"hot:1:" + clipID + ":1::notanumber",
		"hot:1:" + clipID + ":1::0",
		"hot:1:" + clipID + ":1::-5",
		"hot:1:" + clipID + ":1:not-a-uuid:5",
		"hot:1:" + clipID + ":1::5:extra",
	} {
		if _, err := DecodeCursor(encodeRaw(raw)); err == nil {
			t.Errorf("DecodeCursor(%q) succeeded, want error", raw)
		}
	}
}

func TestCursorTimeAcceptsLegacySeconds(t *testing.T) {
	seconds := int64(1609459200)
	if got := CursorTime(seconds); !got.Equal(time.Unix(seconds, 0)) || got.Location() != time.UTC {
		t.Fatalf("CursorTime(seconds) = %v", got)
	}
	micro := time.Date(2026, 1, 2, 3, 4, 5, 678901000, time.UTC)
	if got := CursorTime(CursorTimeValue(micro)); !got.Equal(micro) {
		t.Fatalf("CursorTime(micro) = %v, want %v", got, micro)
	}
}

func encodeRaw(raw string) string {
	return base64.URLEncoding.EncodeToString([]byte(raw))
}
