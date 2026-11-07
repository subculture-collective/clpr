package services

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// Rekolekt clip handoff, contract version 1. The producer documents it in
// rekolekt docs/product/clip-handoff.md; its fixture is copied to
// testdata/rekolekt-clip-handoff.v1.json.
const (
	ArchiveHandoffSchema        = "rekolekt.clip-handoff"
	ArchiveHandoffSchemaVersion = 1

	// MaxArchiveHandoffBytes bounds the raw handoff document.
	MaxArchiveHandoffBytes = 512 << 10
	// MaxArchiveCaptionCues and MaxArchiveCueChars match the producer's bounds.
	MaxArchiveCaptionCues = 200
	MaxArchiveCueChars    = 2000
	// maxArchiveCaptionTextBytes bounds the total caption text kept per clip,
	// because source_metadata is returned with every clip.
	maxArchiveCaptionTextBytes = 64 << 10
	// maxArchivePassageMs is a sanity bound; the producer limits exports to two
	// minutes and clpr's own upload duration limit still applies to the file.
	maxArchivePassageMs = 60 * 60 * 1000
	maxArchiveURLLength = 2048
	maxArchiveIDLength  = 128
	maxArchiveTextField = 200
)

var archiveSHA256Pattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

// ArchiveCaptionCue is one caption cue, timed relative to the clip start.
// A nil EndMs means the archive had no usable end time for that segment.
type ArchiveCaptionCue struct {
	StartMs       int64  `json:"start_ms"`
	EndMs         *int64 `json:"end_ms"`
	Text          string `json:"text"`
	TextTruncated bool   `json:"text_truncated,omitempty"`
}

// ArchiveHandoff is the decoded producer document. Unknown fields are ignored
// and are not stored.
type ArchiveHandoff struct {
	Schema        string  `json:"schema"`
	SchemaVersion int     `json:"schema_version"`
	ExportID      string  `json:"export_id"`
	GeneratedAt   float64 `json:"generated_at"`
	Archive       struct {
		Name   string `json:"name"`
		Origin string `json:"origin"`
	} `json:"archive"`
	Recording  ArchiveRecording  `json:"recording"`
	Passage    ArchivePassage    `json:"passage"`
	Transcript ArchiveTranscript `json:"transcript"`
	Captions   struct {
		Timing        string              `json:"timing"`
		Cues          []ArchiveCaptionCue `json:"cues"`
		CuesTruncated bool                `json:"cues_truncated"`
		TextSHA256    *string             `json:"text_sha256"`
	} `json:"captions"`
	Media      ArchiveMedia      `json:"media"`
	Provenance ArchiveProvenance `json:"provenance"`
}

type ArchiveRecording struct {
	VideoID   string  `json:"video_id"`
	YoutubeID *string `json:"youtube_id"`
}

type ArchivePassage struct {
	URL        string `json:"url"`
	PlayerURL  string `json:"player_url"`
	StartMs    int64  `json:"start_ms"`
	EndMs      int64  `json:"end_ms"`
	DurationMs int64  `json:"duration_ms"`
}

type ArchiveTranscript struct {
	Status         string   `json:"status"`
	Source         *string  `json:"source,omitempty"`
	TranscriptID   *string  `json:"transcript_id,omitempty"`
	Language       *string  `json:"language,omitempty"`
	Kind           *string  `json:"kind,omitempty"`
	SegmentCount   *int     `json:"segment_count,omitempty"`
	RevisionSHA256 *string  `json:"revision_sha256,omitempty"`
	Timing         *string  `json:"timing,omitempty"`
	CapturedAt     *float64 `json:"captured_at,omitempty"`
}

type ArchiveMedia struct {
	Type     string `json:"type"`
	Filename string `json:"filename"`
	Bytes    int64  `json:"bytes"`
	SHA256   string `json:"sha256"`
}

type ArchiveProvenance struct {
	Basis       string `json:"basis"`
	Statement   string `json:"statement"`
	MediaSource struct {
		Kind            string  `json:"kind"`
		YoutubeID       *string `json:"youtube_id,omitempty"`
		OriginalSHA256  *string `json:"original_sha256,omitempty"`
		TimelineAligned *bool   `json:"timeline_aligned,omitempty"`
	} `json:"media_source"`
}

// ArchiveProvenanceRecord is what clpr stores under
// source_metadata.archive_provenance on the submission and the approved clip.
// It is supplied provenance: the archive's declaration of what was selected,
// not verified rights or origin.
type ArchiveProvenanceRecord struct {
	Schema        string `json:"schema"`
	SchemaVersion int    `json:"schema_version"`
	ExportID      string `json:"export_id"`
	Archive       struct {
		Name   string `json:"name"`
		Origin string `json:"origin"`
	} `json:"archive"`
	Recording  ArchiveRecording         `json:"recording"`
	Passage    ArchivePassage           `json:"passage"`
	Transcript ArchiveTranscript        `json:"transcript"`
	Captions   ArchiveCaptionsRecord    `json:"captions"`
	Media      ArchiveMedia             `json:"media"`
	Provenance ArchiveProvenance        `json:"provenance"`
	Import     ArchiveImportCheckRecord `json:"import"`
}

// ArchiveCaptionsRecord keeps the archive's cues apart from the submitter's edits.
type ArchiveCaptionsRecord struct {
	Timing           string              `json:"timing"`
	SourceCues       []ArchiveCaptionCue `json:"source_cues"`
	SourceTextSHA256 *string             `json:"source_text_sha256"`
	CuesTruncated    bool                `json:"cues_truncated"`
	Cues             []ArchiveCaptionCue `json:"cues"`
	Edited           bool                `json:"edited"`
	Language         *string             `json:"language,omitempty"`
}

// ArchiveImportCheckRecord records what clpr itself checked at import.
type ArchiveImportCheckRecord struct {
	MediaDigestMatched bool      `json:"media_digest_matched"`
	ImportedAt         time.Time `json:"imported_at"`
}

// ParseArchiveHandoff decodes and validates a version 1 handoff document.
// Errors are *ValidationError with a message the submitter can act on.
func ParseArchiveHandoff(raw []byte) (*ArchiveHandoff, error) {
	if len(raw) == 0 {
		return nil, archiveErr("The archive handoff is empty")
	}
	if len(raw) > MaxArchiveHandoffBytes {
		return nil, archiveErr(fmt.Sprintf("The archive handoff is larger than %d KiB", MaxArchiveHandoffBytes>>10))
	}

	var header struct {
		Schema        *string          `json:"schema"`
		SchemaVersion *json.RawMessage `json:"schema_version"`
	}
	if err := json.Unmarshal(raw, &header); err != nil {
		return nil, archiveErr("The archive handoff is not valid JSON")
	}
	if header.Schema == nil || *header.Schema != ArchiveHandoffSchema {
		return nil, archiveErr(fmt.Sprintf("This file is not a %s document", ArchiveHandoffSchema))
	}
	if header.SchemaVersion == nil || strings.TrimSpace(string(*header.SchemaVersion)) != strconv.Itoa(ArchiveHandoffSchemaVersion) {
		return nil, archiveErr(fmt.Sprintf("Unsupported archive handoff version; clpr reads version %d. Export the clip again from the archive.", ArchiveHandoffSchemaVersion))
	}

	var handoff ArchiveHandoff
	if err := json.Unmarshal(raw, &handoff); err != nil {
		return nil, archiveErr("The archive handoff has a field of the wrong type")
	}
	if err := validateArchiveHandoff(&handoff); err != nil {
		return nil, err
	}
	return &handoff, nil
}

func validateArchiveHandoff(h *ArchiveHandoff) error {
	if _, err := uuid.Parse(h.ExportID); err != nil {
		return archiveErr("The archive handoff has an invalid export_id")
	}
	if err := boundedText("archive.name", h.Archive.Name, maxArchiveTextField, true); err != nil {
		return err
	}
	origin, err := parseArchiveOrigin(h.Archive.Origin)
	if err != nil {
		return err
	}
	if err := boundedText("recording.video_id", h.Recording.VideoID, maxArchiveIDLength, true); err != nil {
		return err
	}
	if h.Recording.YoutubeID != nil {
		if err := boundedText("recording.youtube_id", *h.Recording.YoutubeID, maxArchiveIDLength, false); err != nil {
			return err
		}
	}

	p := h.Passage
	if p.StartMs < 0 || p.EndMs <= p.StartMs {
		return archiveErr("The archive handoff has an invalid passage range: end_ms must be after start_ms, and start_ms cannot be negative")
	}
	if p.DurationMs != p.EndMs-p.StartMs {
		return archiveErr("The archive handoff passage duration does not match its start and end")
	}
	if p.DurationMs > maxArchivePassageMs {
		return archiveErr("The archive handoff passage is longer than one hour")
	}
	if err := sameOriginArchiveURL("passage.url", p.URL, origin); err != nil {
		return err
	}
	if err := sameOriginArchiveURL("passage.player_url", p.PlayerURL, origin); err != nil {
		return err
	}

	if err := validateArchiveTranscript(&h.Transcript); err != nil {
		return err
	}

	switch h.Captions.Timing {
	case "clip-relative", "unavailable":
	default:
		return archiveErr("The archive handoff has unsupported caption timing")
	}
	if h.Captions.Timing == "unavailable" && len(h.Captions.Cues) > 0 {
		return archiveErr("The archive handoff lists cues but says caption timing is unavailable")
	}
	if err := ValidateArchiveCaptionCues(h.Captions.Cues, p.DurationMs); err != nil {
		return err
	}
	if len(h.Captions.Cues) > 0 {
		if h.Captions.TextSHA256 == nil || !archiveSHA256Pattern.MatchString(*h.Captions.TextSHA256) {
			return archiveErr("The archive handoff captions are missing their text_sha256 digest")
		}
		if got := ArchiveCueTextSHA256(h.Captions.Cues); got != *h.Captions.TextSHA256 {
			return archiveErr("The archive handoff captions do not match their text_sha256 digest; the file may have been edited. Download the handoff again.")
		}
	} else if h.Captions.TextSHA256 != nil && !archiveSHA256Pattern.MatchString(*h.Captions.TextSHA256) {
		return archiveErr("The archive handoff has an invalid captions text_sha256")
	}

	m := h.Media
	if m.Type != "video/mp4" {
		return archiveErr("The archive handoff media type must be video/mp4")
	}
	if m.Bytes <= 0 {
		return archiveErr("The archive handoff media size must be positive")
	}
	if !archiveSHA256Pattern.MatchString(m.SHA256) {
		return archiveErr("The archive handoff media sha256 must be 64 lowercase hex characters")
	}
	if err := boundedText("media.filename", m.Filename, 255, true); err != nil {
		return err
	}

	pv := h.Provenance
	if pv.Basis != "declared" {
		return archiveErr("The archive handoff provenance basis must be \"declared\"")
	}
	if err := boundedText("provenance.statement", pv.Statement, 1000, false); err != nil {
		return err
	}
	switch pv.MediaSource.Kind {
	case "youtube-range":
		if pv.MediaSource.YoutubeID == nil || strings.TrimSpace(*pv.MediaSource.YoutubeID) == "" {
			return archiveErr("The archive handoff youtube-range source is missing its youtube_id")
		}
		if err := boundedText("provenance.media_source.youtube_id", *pv.MediaSource.YoutubeID, maxArchiveIDLength, true); err != nil {
			return err
		}
	case "registered-original":
		if pv.MediaSource.OriginalSHA256 == nil || !archiveSHA256Pattern.MatchString(*pv.MediaSource.OriginalSHA256) {
			return archiveErr("The archive handoff registered-original source is missing a valid original_sha256")
		}
	default:
		return archiveErr("The archive handoff has an unsupported media source kind")
	}
	return nil
}

func validateArchiveTranscript(t *ArchiveTranscript) error {
	switch t.Status {
	case "unavailable":
		return nil
	case "captured":
	default:
		return archiveErr("The archive handoff transcript status must be captured or unavailable")
	}
	if t.Source == nil || (*t.Source != "whisper" && *t.Source != "youtube") {
		return archiveErr("The archive handoff transcript source must be whisper or youtube")
	}
	if t.Kind == nil || (*t.Kind != "machine" && *t.Kind != "caption-auto" && *t.Kind != "caption-manual") {
		return archiveErr("The archive handoff transcript kind is not supported")
	}
	if t.RevisionSHA256 == nil || !archiveSHA256Pattern.MatchString(*t.RevisionSHA256) {
		return archiveErr("The archive handoff transcript is missing a valid revision_sha256")
	}
	if t.SegmentCount == nil || *t.SegmentCount < 0 {
		return archiveErr("The archive handoff transcript segment_count is invalid")
	}
	if t.Timing == nil || *t.Timing != "segment" {
		return archiveErr("The archive handoff transcript timing is not supported")
	}
	if t.TranscriptID != nil {
		if err := boundedText("transcript.transcript_id", *t.TranscriptID, maxArchiveIDLength, false); err != nil {
			return err
		}
	}
	if t.Language != nil {
		if err := boundedText("transcript.language", *t.Language, 35, false); err != nil {
			return err
		}
	}
	return nil
}

// ValidateArchiveCaptionCues checks the bounds the archive applies to cues
// and that every cue falls inside a clip of durationMs.
func ValidateArchiveCaptionCues(cues []ArchiveCaptionCue, durationMs int64) error {
	if len(cues) > MaxArchiveCaptionCues {
		return archiveErr(fmt.Sprintf("Captions can have at most %d cues", MaxArchiveCaptionCues))
	}
	total := 0
	var previousStart int64
	for i, cue := range cues {
		n := i + 1
		if cue.StartMs < 0 || cue.StartMs > durationMs {
			return archiveErr(fmt.Sprintf("Caption cue %d starts outside the clip", n))
		}
		if cue.StartMs < previousStart {
			return archiveErr(fmt.Sprintf("Caption cue %d starts before the cue above it", n))
		}
		previousStart = cue.StartMs
		if cue.EndMs != nil && (*cue.EndMs <= cue.StartMs || *cue.EndMs > durationMs) {
			return archiveErr(fmt.Sprintf("Caption cue %d has an end time outside the clip or before its start", n))
		}
		if !utf8.ValidString(cue.Text) {
			return archiveErr(fmt.Sprintf("Caption cue %d is not valid UTF-8 text", n))
		}
		if utf8.RuneCountInString(cue.Text) > MaxArchiveCueChars {
			return archiveErr(fmt.Sprintf("Caption cue %d is longer than %d characters", n, MaxArchiveCueChars))
		}
		total += len(cue.Text)
	}
	if total > maxArchiveCaptionTextBytes {
		return archiveErr(fmt.Sprintf("Captions are longer than %d KiB in total", maxArchiveCaptionTextBytes>>10))
	}
	return nil
}

// ParseEditedArchiveCaptionCues reads the submitter's edited cues. Text is
// trimmed; an empty cue must be removed rather than sent blank.
func ParseEditedArchiveCaptionCues(raw []byte, durationMs int64) ([]ArchiveCaptionCue, error) {
	if len(raw) > MaxArchiveHandoffBytes {
		return nil, archiveErr("The edited captions are too large")
	}
	var cues []ArchiveCaptionCue
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cues); err != nil {
		return nil, archiveErr("The edited captions are not a valid list of cues")
	}
	for i := range cues {
		cues[i].Text = strings.TrimSpace(cues[i].Text)
		if cues[i].Text == "" {
			return nil, archiveErr(fmt.Sprintf("Caption cue %d is empty; remove it instead", i+1))
		}
		cues[i].TextTruncated = false
	}
	if err := ValidateArchiveCaptionCues(cues, durationMs); err != nil {
		return nil, err
	}
	return cues, nil
}

// ArchiveCueTextSHA256 computes the producer's cue digest: SHA-256 over one
// compact JSON line [start_ms, end_ms, text] per cue, non-ASCII kept as-is.
func ArchiveCueTextSHA256(cues []ArchiveCaptionCue) string {
	sum := sha256.New()
	var line bytes.Buffer
	for _, cue := range cues {
		line.Reset()
		line.WriteByte('[')
		line.WriteString(strconv.FormatInt(cue.StartMs, 10))
		line.WriteByte(',')
		if cue.EndMs == nil {
			line.WriteString("null")
		} else {
			line.WriteString(strconv.FormatInt(*cue.EndMs, 10))
		}
		line.WriteByte(',')
		writePythonJSONString(&line, cue.Text)
		line.WriteString("]\n")
		sum.Write(line.Bytes())
	}
	return hex.EncodeToString(sum.Sum(nil))
}

// writePythonJSONString matches Python's json.dumps(..., ensure_ascii=False):
// it escapes only quotes, backslashes and control characters.
func writePythonJSONString(buf *bytes.Buffer, s string) {
	buf.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			buf.WriteString(`\"`)
		case '\\':
			buf.WriteString(`\\`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		case '\b':
			buf.WriteString(`\b`)
		case '\f':
			buf.WriteString(`\f`)
		default:
			if r < 0x20 {
				fmt.Fprintf(buf, `\u%04x`, r)
			} else {
				buf.WriteRune(r)
			}
		}
	}
	buf.WriteByte('"')
}

// BuildArchiveProvenanceRecord combines a validated handoff, the submitter's
// cues (nil when they sent none) and clpr's own digest check into the stored
// record. Cues identical to the archive's are not marked as edited.
func BuildArchiveProvenanceRecord(h *ArchiveHandoff, editedCues []ArchiveCaptionCue, now time.Time) *ArchiveProvenanceRecord {
	record := &ArchiveProvenanceRecord{
		Schema:        h.Schema,
		SchemaVersion: h.SchemaVersion,
		ExportID:      h.ExportID,
		Archive:       h.Archive,
		Recording:     h.Recording,
		Passage:       h.Passage,
		Transcript:    h.Transcript,
		Media:         h.Media,
		Provenance:    h.Provenance,
		Import:        ArchiveImportCheckRecord{MediaDigestMatched: true, ImportedAt: now.UTC()},
	}
	sourceCues := h.Captions.Cues
	if sourceCues == nil {
		sourceCues = []ArchiveCaptionCue{}
	}
	cues := sourceCues
	edited := editedCues != nil && !sameArchiveCues(editedCues, sourceCues)
	if edited {
		cues = editedCues
	}
	record.Captions = ArchiveCaptionsRecord{
		Timing:           h.Captions.Timing,
		SourceCues:       sourceCues,
		SourceTextSHA256: h.Captions.TextSHA256,
		CuesTruncated:    h.Captions.CuesTruncated,
		Cues:             cues,
		Edited:           edited,
		Language:         h.Transcript.Language,
	}
	return record
}

func sameArchiveCues(a, b []ArchiveCaptionCue) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].StartMs != b[i].StartMs || a[i].Text != b[i].Text || (a[i].EndMs == nil) != (b[i].EndMs == nil) {
			return false
		}
		if a[i].EndMs != nil && *a[i].EndMs != *b[i].EndMs {
			return false
		}
	}
	return true
}

// ArchiveProvenanceFromMetadata reads the stored record from source_metadata.
// It returns nil when the clip was not imported from an archive handoff.
func ArchiveProvenanceFromMetadata(metadata json.RawMessage) *ArchiveProvenanceRecord {
	if len(metadata) == 0 {
		return nil
	}
	var wrapper struct {
		ArchiveProvenance *ArchiveProvenanceRecord `json:"archive_provenance"`
	}
	if err := json.Unmarshal(metadata, &wrapper); err != nil {
		return nil
	}
	return wrapper.ArchiveProvenance
}

func parseArchiveOrigin(raw string) (*url.URL, error) {
	parsed, err := parseArchiveURL("archive.origin", raw)
	if err != nil {
		return nil, err
	}
	if (parsed.Path != "" && parsed.Path != "/") || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, archiveErr("The archive handoff origin must be only a scheme and host")
	}
	return parsed, nil
}

func sameOriginArchiveURL(field, raw string, origin *url.URL) error {
	parsed, err := parseArchiveURL(field, raw)
	if err != nil {
		return err
	}
	if !strings.EqualFold(parsed.Host, origin.Host) {
		return archiveErr(fmt.Sprintf("The archive handoff %s is not on the archive's origin", field))
	}
	return nil
}

func parseArchiveURL(field, raw string) (*url.URL, error) {
	if raw == "" || len(raw) > maxArchiveURLLength {
		return nil, archiveErr(fmt.Sprintf("The archive handoff %s is missing or too long", field))
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.Hostname() == "" {
		return nil, archiveErr(fmt.Sprintf("The archive handoff %s must be an https URL", field))
	}
	if parsed.User != nil {
		return nil, archiveErr(fmt.Sprintf("The archive handoff %s must not contain credentials", field))
	}
	return parsed, nil
}

func boundedText(field, value string, max int, required bool) error {
	if required && strings.TrimSpace(value) == "" {
		return archiveErr(fmt.Sprintf("The archive handoff is missing %s", field))
	}
	if utf8.RuneCountInString(value) > max {
		return archiveErr(fmt.Sprintf("The archive handoff %s is longer than %d characters", field, max))
	}
	return nil
}

func archiveErr(message string) error {
	return &ValidationError{Field: "archive_handoff", Message: message}
}
