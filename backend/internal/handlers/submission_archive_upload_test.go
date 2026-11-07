package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// archiveHandoffFor returns the producer fixture with its media fields set to
// describe content, as a real export of those bytes would.
func archiveHandoffFor(t *testing.T, content []byte, mutate func(map[string]any)) string {
	t.Helper()
	raw, err := os.ReadFile("../services/testdata/rekolekt-clip-handoff.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(content)
	media := doc["media"].(map[string]any)
	media["bytes"] = len(content)
	media["sha256"] = hex.EncodeToString(sum[:])
	if mutate != nil {
		mutate(doc)
	}
	out, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

type archiveUploadHarness struct {
	handler     *SubmissionHandler
	storage     *fakeClipStorage
	probeCalled *bool
	got         **services.SubmitUploadRequest
}

func newArchiveUploadHarness(t *testing.T) archiveUploadHarness {
	t.Helper()
	gin.SetMode(gin.TestMode)
	probeCalled := false
	var got *services.SubmitUploadRequest
	service := &fakeSubmissionService{
		submitUploadFn: func(ctx context.Context, userID uuid.UUID, req *services.SubmitUploadRequest, ip, fp string) (*models.ClipSubmission, error) {
			got = req
			return &models.ClipSubmission{ID: req.SubmissionID, UserID: userID, Status: "pending", SourceType: "upload", SourcePlatform: "upload"}, nil
		},
	}
	validator := services.NewUploadValidator(config.ClipSourceConfig{
		MaxDurationSeconds:     600,
		MaxUploadBytes:         4096,
		AllowedUploadMimeTypes: []string{"video/mp4"},
	}, func(ctx context.Context, filePath string) (int64, error) {
		probeCalled = true
		return 9, nil
	})
	storageBackend := &fakeClipStorage{}
	return archiveUploadHarness{
		handler:     NewSubmissionHandler(service, validator, storageBackend),
		storage:     storageBackend,
		probeCalled: &probeCalled,
		got:         &got,
	}
}

func (h archiveUploadHarness) post(t *testing.T, fields map[string]string, content []byte) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Set("user_id", uuid.New())
	ctx.Request = newUploadTestRequest(t, "/api/v1/submissions/upload", fields, "file", "passage.mp4", content)
	h.handler.SubmitUpload(ctx)
	return rec
}

func TestSubmitUploadWithMatchingArchiveHandoffStoresProvenance(t *testing.T) {
	h := newArchiveUploadHarness(t)
	content := bytes.Repeat([]byte{0}, 1024)
	edited := `[{"start_ms":0,"end_ms":1840,"text":"and that is the part nobody reads,"},{"start_ms":6900,"end_ms":9000,"text":"So let's actually read it."}]`

	rec := h.post(t, map[string]string{"archive_handoff": archiveHandoffFor(t, content, nil), "caption_cues": edited}, content)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	req := *h.got
	if req == nil || req.ArchiveProvenance == nil {
		t.Fatal("submission request carries no archive provenance")
	}
	record := req.ArchiveProvenance
	if !record.Import.MediaDigestMatched || record.Provenance.Basis != "declared" {
		t.Fatalf("record = %+v", record)
	}
	if !record.Captions.Edited || len(record.Captions.Cues) != 2 || len(record.Captions.SourceCues) != 4 {
		t.Fatalf("captions = %+v", record.Captions)
	}
	if record.Captions.SourceCues[0].Text != "and that's the part nobody reads," {
		t.Fatalf("source cue overwritten: %q", record.Captions.SourceCues[0].Text)
	}
	if !*h.probeCalled || h.storage.putCalls != 1 {
		t.Fatalf("probe=%v puts=%d: the normal validation and storage must still run", *h.probeCalled, h.storage.putCalls)
	}
}

func TestSubmitUploadWithoutArchiveHandoffIsUnchanged(t *testing.T) {
	h := newArchiveUploadHarness(t)
	rec := h.post(t, map[string]string{}, bytes.Repeat([]byte{0}, 512))
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if (*h.got).ArchiveProvenance != nil {
		t.Fatal("plain upload gained archive provenance")
	}
}

func TestSubmitUploadArchiveHandoffRejectionsHappenBeforeStorage(t *testing.T) {
	content := bytes.Repeat([]byte{0}, 1024)
	cases := []struct {
		name   string
		fields func(t *testing.T) map[string]string
		file   []byte
		field  string
		want   string
	}{
		{
			name: "mismatched bytes",
			fields: func(t *testing.T) map[string]string {
				return map[string]string{"archive_handoff": archiveHandoffFor(t, content, nil)}
			},
			file:  bytes.Repeat([]byte{1}, 1024),
			field: "file",
			want:  "does not match the archive handoff",
		},
		{
			name: "mismatched size",
			fields: func(t *testing.T) map[string]string {
				return map[string]string{"archive_handoff": archiveHandoffFor(t, content, nil)}
			},
			file:  bytes.Repeat([]byte{0}, 1000),
			field: "file",
			want:  "does not match the archive handoff",
		},
		{
			name: "unsupported version",
			fields: func(t *testing.T) map[string]string {
				return map[string]string{"archive_handoff": archiveHandoffFor(t, content, func(d map[string]any) { d["schema_version"] = 2 })}
			},
			file:  content,
			field: "archive_handoff",
			want:  "Unsupported archive handoff version",
		},
		{
			name: "unsafe passage url",
			fields: func(t *testing.T) map[string]string {
				return map[string]string{"archive_handoff": archiveHandoffFor(t, content, func(d map[string]any) {
					d["passage"].(map[string]any)["url"] = "javascript:alert(1)"
				})}
			},
			file:  content,
			field: "archive_handoff",
			want:  "https URL",
		},
		{
			name: "edited cue past the clip",
			fields: func(t *testing.T) map[string]string {
				return map[string]string{
					"archive_handoff": archiveHandoffFor(t, content, nil),
					"caption_cues":    `[{"start_ms":0,"end_ms":99999,"text":"x"}]`,
				}
			},
			file:  content,
			field: "archive_handoff",
			want:  "outside the clip",
		},
		{
			name:   "cues without a handoff",
			fields: func(t *testing.T) map[string]string { return map[string]string{"caption_cues": `[]`} },
			file:   content,
			field:  "caption_cues",
			want:   "only be sent with an archive_handoff",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newArchiveUploadHarness(t)
			rec := h.post(t, tc.fields(t), tc.file)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
			}
			var body struct {
				Error string `json:"error"`
				Field string `json:"field"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Field != tc.field || !strings.Contains(body.Error, tc.want) {
				t.Fatalf("error = %q (field %q), want %q in field %q", body.Error, body.Field, tc.want, tc.field)
			}
			if h.storage.putCalls != 0 || *h.probeCalled || *h.got != nil {
				t.Fatalf("puts=%d probe=%v submitted=%v; a rejected import must not reach validation, storage or submission", h.storage.putCalls, *h.probeCalled, *h.got != nil)
			}
		})
	}
}

func TestRenderArchiveCaptionsVTT(t *testing.T) {
	end := func(v int64) *int64 { return &v }
	got := renderArchiveCaptionsVTT([]services.ArchiveCaptionCue{
		{StartMs: 0, EndMs: end(1840), Text: "a <b>bold</b> & --> claim"},
		{StartMs: 5210, EndMs: nil, Text: "[laughter]"},
		{StartMs: 6900, EndMs: end(9000), Text: "line one\n\nline two"},
		{StartMs: 8990, EndMs: nil, Text: "last"},
	}, 9000)
	want := "WEBVTT\n" +
		"\n1\n00:00:00.000 --> 00:00:01.840\na &lt;b&gt;bold&lt;/b&gt; &amp; --&gt; claim\n" +
		"\n2\n00:00:05.210 --> 00:00:06.900\n[laughter]\n" +
		"\n3\n00:00:06.900 --> 00:00:09.000\nline one\nline two\n" +
		"\n4\n00:00:08.990 --> 00:00:09.000\nlast\n"
	if got != want {
		t.Fatalf("vtt =\n%s\nwant\n%s", got, want)
	}
	if ts := vttTimestamp(3_723_004); ts != "01:02:03.004" {
		t.Fatalf("timestamp = %s", ts)
	}
}
