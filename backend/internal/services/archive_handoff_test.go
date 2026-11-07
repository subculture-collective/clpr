package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

// The fixture is the producer's docs/api/examples/clip-handoff.v1.json from
// rekolekt a4d37c304e54ecfbf8822d2c6162d4f88b414483, copied byte for byte.
const archiveHandoffFixture = "testdata/rekolekt-clip-handoff.v1.json"

func loadArchiveHandoffFixture(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile(archiveHandoffFixture)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return raw
}

// mutateArchiveHandoff decodes the fixture into a generic map, applies fn and
// re-encodes it, so each case changes exactly one thing.
func mutateArchiveHandoff(t *testing.T, fn func(map[string]any)) []byte {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal(loadArchiveHandoffFixture(t), &doc); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	fn(doc)
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	return raw
}

func section(doc map[string]any, key string) map[string]any {
	return doc[key].(map[string]any)
}

func TestParseArchiveHandoffAcceptsProducerFixture(t *testing.T) {
	handoff, err := ParseArchiveHandoff(loadArchiveHandoffFixture(t))
	if err != nil {
		t.Fatalf("fixture rejected: %v", err)
	}
	if handoff.Passage.StartMs != 754000 || handoff.Passage.EndMs != 763000 || handoff.Passage.DurationMs != 9000 {
		t.Fatalf("passage = %+v", handoff.Passage)
	}
	if got := len(handoff.Captions.Cues); got != 4 {
		t.Fatalf("cues = %d, want 4", got)
	}
	if handoff.Captions.Cues[2].EndMs != nil {
		t.Fatalf("untimed cue end = %v, want nil", *handoff.Captions.Cues[2].EndMs)
	}
	if handoff.Transcript.RevisionSHA256 == nil || *handoff.Transcript.Source != "whisper" {
		t.Fatalf("transcript = %+v", handoff.Transcript)
	}
}

func TestArchiveCueTextSHA256MatchesProducer(t *testing.T) {
	handoff, err := ParseArchiveHandoff(loadArchiveHandoffFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	const want = "d9c7ad084c026a0b90974f17ef70e415b714a2b97690e3f93dded73fbbc3b64a"
	if got := ArchiveCueTextSHA256(handoff.Captions.Cues); got != want {
		t.Fatalf("digest = %s, want %s", got, want)
	}
}

func TestArchiveCueTextSHA256KeepsNonASCIIAndEscapesLikePython(t *testing.T) {
	// Python: json.dumps([0, None, 'café "<b>"  \x01\n'], ensure_ascii=False, separators=(",", ":"))
	var buf bytes.Buffer
	writePythonJSONString(&buf, "café \"<b>\"  \x01\n")
	want := "\"café \\\"<b>\\\"  \\u0001\\n\""
	if buf.String() != want {
		t.Fatalf("encoded = %q, want %q", buf.String(), want)
	}
}

func TestParseArchiveHandoffRejections(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(map[string]any)
		message string
	}{
		{"wrong schema", func(d map[string]any) { d["schema"] = "other.handoff" }, "not a rekolekt.clip-handoff"},
		{"unsupported version", func(d map[string]any) { d["schema_version"] = 2 }, "Unsupported archive handoff version"},
		{"missing version", func(d map[string]any) { delete(d, "schema_version") }, "Unsupported archive handoff version"},
		{"bad export id", func(d map[string]any) { d["export_id"] = "nope" }, "export_id"},
		{"end before start", func(d map[string]any) { section(d, "passage")["end_ms"] = 750000 }, "invalid passage range"},
		{"negative start", func(d map[string]any) {
			p := section(d, "passage")
			p["start_ms"] = -1
		}, "invalid passage range"},
		{"duration mismatch", func(d map[string]any) { section(d, "passage")["duration_ms"] = 1 }, "duration does not match"},
		{"http passage url", func(d map[string]any) {
			section(d, "passage")["url"] = "http://archive.example/api/share/videos/x"
		}, "must be an https URL"},
		{"javascript passage url", func(d map[string]any) {
			section(d, "passage")["url"] = "javascript:alert(1)"
		}, "must be an https URL"},
		{"passage on another origin", func(d map[string]any) {
			section(d, "passage")["url"] = "https://evil.example/api/share/videos/x"
		}, "not on the archive's origin"},
		{"credentials in url", func(d map[string]any) {
			section(d, "passage")["player_url"] = "https://user:pw@archive.example/v/x"
		}, "must not contain credentials"},
		{"origin with path", func(d map[string]any) { section(d, "archive")["origin"] = "https://archive.example/app" }, "only a scheme and host"},
		{"oversize name", func(d map[string]any) {
			section(d, "archive")["name"] = strings.Repeat("a", maxArchiveTextField+1)
		}, "archive.name is longer"},
		{"cue past clip end", func(d map[string]any) {
			cue := section(d, "captions")["cues"].([]any)[3].(map[string]any)
			cue["end_ms"] = 9001
		}, "Caption cue 4 has an end time outside the clip"},
		{"cue text edited", func(d map[string]any) {
			cue := section(d, "captions")["cues"].([]any)[0].(map[string]any)
			cue["text"] = "something else"
		}, "do not match their text_sha256"},
		{"oversize cue", func(d map[string]any) {
			cue := section(d, "captions")["cues"].([]any)[0].(map[string]any)
			cue["text"] = strings.Repeat("x", MaxArchiveCueChars+1)
		}, "longer than 2000 characters"},
		{"media not mp4", func(d map[string]any) { section(d, "media")["type"] = "video/webm" }, "must be video/mp4"},
		{"bad media digest", func(d map[string]any) { section(d, "media")["sha256"] = "ABC" }, "media sha256"},
		{"zero media size", func(d map[string]any) { section(d, "media")["bytes"] = 0 }, "media size must be positive"},
		{"verified basis", func(d map[string]any) { section(d, "provenance")["basis"] = "verified" }, "basis must be \"declared\""},
		{"unknown media source", func(d map[string]any) {
			section(section(d, "provenance"), "media_source")["kind"] = "remote-url"
		}, "unsupported media source kind"},
		{"unknown transcript source", func(d map[string]any) { section(d, "transcript")["source"] = "other" }, "whisper or youtube"},
		{"wrong field type", func(d map[string]any) { section(d, "passage")["start_ms"] = "754000" }, "wrong type"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseArchiveHandoff(mutateArchiveHandoff(t, tc.mutate))
			var valErr *ValidationError
			if !errors.As(err, &valErr) {
				t.Fatalf("err = %v, want ValidationError", err)
			}
			if !strings.Contains(valErr.Message, tc.message) {
				t.Fatalf("message = %q, want it to contain %q", valErr.Message, tc.message)
			}
		})
	}
}

func TestParseArchiveHandoffRejectsMalformedAndOversize(t *testing.T) {
	if _, err := ParseArchiveHandoff([]byte("{not json")); err == nil || !strings.Contains(err.Error(), "not valid JSON") {
		t.Fatalf("malformed err = %v", err)
	}
	if _, err := ParseArchiveHandoff(bytes.Repeat([]byte(" "), MaxArchiveHandoffBytes+1)); err == nil || !strings.Contains(err.Error(), "larger than") {
		t.Fatalf("oversize err = %v", err)
	}
}

func TestParseArchiveHandoffAcceptsUnavailableTranscript(t *testing.T) {
	raw := mutateArchiveHandoff(t, func(d map[string]any) {
		d["transcript"] = map[string]any{"status": "unavailable", "captured_at": 1794066900.0}
		d["captions"] = map[string]any{"timing": "unavailable", "cues": []any{}, "cues_truncated": false, "text_sha256": nil}
	})
	handoff, err := ParseArchiveHandoff(raw)
	if err != nil {
		t.Fatalf("unavailable transcript rejected: %v", err)
	}
	record := BuildArchiveProvenanceRecord(handoff, nil, time.Now())
	if len(record.Captions.Cues) != 0 || record.Captions.SourceCues == nil {
		t.Fatalf("captions = %+v", record.Captions)
	}
}

func TestParseEditedArchiveCaptionCues(t *testing.T) {
	cues, err := ParseEditedArchiveCaptionCues([]byte(`[{"start_ms":0,"end_ms":1840,"text":"  edited  "},{"start_ms":5210,"end_ms":null,"text":"[laughs]"}]`), 9000)
	if err != nil {
		t.Fatal(err)
	}
	if cues[0].Text != "edited" || cues[1].EndMs != nil {
		t.Fatalf("cues = %+v", cues)
	}

	for name, raw := range map[string]string{
		"empty text":    `[{"start_ms":0,"end_ms":10,"text":"   "}]`,
		"past end":      `[{"start_ms":0,"end_ms":9001,"text":"x"}]`,
		"out of order":  `[{"start_ms":500,"end_ms":900,"text":"a"},{"start_ms":100,"end_ms":200,"text":"b"}]`,
		"unknown field": `[{"start_ms":0,"end_ms":10,"text":"x","style":"bold"}]`,
		"not a list":    `{"start_ms":0}`,
	} {
		if _, err := ParseEditedArchiveCaptionCues([]byte(raw), 9000); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

func TestBuildArchiveProvenanceRecordKeepsSourceCuesApartFromEdits(t *testing.T) {
	handoff, err := ParseArchiveHandoff(loadArchiveHandoffFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	end := int64(1840)
	edited := []ArchiveCaptionCue{{StartMs: 0, EndMs: &end, Text: "and that is the part nobody reads,"}}
	unchanged := BuildArchiveProvenanceRecord(handoff, append([]ArchiveCaptionCue(nil), handoff.Captions.Cues...), time.Now())
	if unchanged.Captions.Edited {
		t.Fatal("cues identical to the archive's were marked as edited")
	}
	record := BuildArchiveProvenanceRecord(handoff, edited, time.Date(2026, 11, 7, 0, 0, 0, 0, time.UTC))

	if !record.Captions.Edited || len(record.Captions.Cues) != 1 || len(record.Captions.SourceCues) != 4 {
		t.Fatalf("captions = %+v", record.Captions)
	}
	if record.Captions.SourceCues[0].Text != "and that's the part nobody reads," {
		t.Fatalf("source cue changed: %q", record.Captions.SourceCues[0].Text)
	}
	if !record.Import.MediaDigestMatched || record.Provenance.Basis != "declared" {
		t.Fatalf("record = %+v", record)
	}

	// The record round-trips through source_metadata.
	metadata, err := json.Marshal(map[string]any{"storage_key": "x", "archive_provenance": record})
	if err != nil {
		t.Fatal(err)
	}
	back := ArchiveProvenanceFromMetadata(metadata)
	if back == nil || back.ExportID != handoff.ExportID || len(back.Captions.SourceCues) != 4 || back.Captions.Cues[0].Text != edited[0].Text {
		t.Fatalf("round trip = %+v", back)
	}
	if ArchiveProvenanceFromMetadata([]byte(`{"storage_key":"x"}`)) != nil {
		t.Fatal("plain upload metadata reported archive provenance")
	}
}

func TestArchiveHandoffFixtureCopiesMatch(t *testing.T) {
	frontendCopy, err := os.ReadFile("../../../frontend/src/test/fixtures/rekolekt-clip-handoff.v1.json")
	if err != nil {
		t.Skipf("frontend fixture not available: %v", err)
	}
	if !bytes.Equal(frontendCopy, loadArchiveHandoffFixture(t)) {
		t.Fatal("frontend and backend copies of the Rekolekt fixture differ")
	}
}
