package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetClipCaptions handles GET /clips/:id/captions.
// It serves the optional captions of a clip imported from an archive handoff
// as WebVTT, for the player's <track> element.
func (h *ClipHandler) GetClipCaptions(c *gin.Context) {
	clipIDParam := c.Param("id")
	if clipIDParam == "" || len(clipIDParam) > 128 {
		c.JSON(http.StatusBadRequest, StandardResponse{Success: false, Error: &ErrorInfo{Code: "INVALID_CLIP_ID", Message: "Invalid clip ID"}})
		return
	}

	var clip *services.ClipWithUserData
	var err error
	if clipID, parseErr := uuid.Parse(clipIDParam); parseErr == nil {
		clip, err = h.clipService.GetClip(c.Request.Context(), clipID, nil)
	} else {
		clip, err = h.clipService.GetClipByTwitchID(c.Request.Context(), clipIDParam, nil)
	}
	if err != nil || clip == nil {
		c.JSON(http.StatusNotFound, StandardResponse{Success: false, Error: &ErrorInfo{Code: "CLIP_NOT_FOUND", Message: "Clip not found or has been removed"}})
		return
	}

	record := services.ArchiveProvenanceFromMetadata(clip.SourceMetadata)
	if record == nil || len(record.Captions.Cues) == 0 {
		c.JSON(http.StatusNotFound, StandardResponse{Success: false, Error: &ErrorInfo{Code: "CAPTIONS_NOT_AVAILABLE", Message: "This clip has no captions"}})
		return
	}

	c.Header("Cache-Control", "public, max-age=300")
	c.Data(http.StatusOK, "text/vtt; charset=utf-8", []byte(renderArchiveCaptionsVTT(record.Captions.Cues, record.Passage.DurationMs)))
}

// renderArchiveCaptionsVTT writes cues as WebVTT. A cue without an end time
// is shown until the next cue starts, or until the clip ends.
func renderArchiveCaptionsVTT(cues []services.ArchiveCaptionCue, durationMs int64) string {
	var b strings.Builder
	b.WriteString("WEBVTT\n")
	for i, cue := range cues {
		end := durationMs
		if cue.EndMs != nil {
			end = *cue.EndMs
		} else if i+1 < len(cues) && cues[i+1].StartMs > cue.StartMs {
			end = cues[i+1].StartMs
		}
		if end <= cue.StartMs {
			continue
		}
		text := vttCueText(cue.Text)
		if text == "" {
			continue
		}
		fmt.Fprintf(&b, "\n%d\n%s --> %s\n%s\n", i+1, vttTimestamp(cue.StartMs), vttTimestamp(end), text)
	}
	return b.String()
}

// vttCueText escapes markup and removes what would end a cue early: blank
// lines and the "-->" separator.
func vttCueText(text string) string {
	text = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\r\n", "\n", "\r", "\n").Replace(text)
	lines := strings.Split(text, "\n")
	kept := lines[:0]
	for _, line := range lines {
		if line = strings.TrimSpace(line); line != "" {
			kept = append(kept, line)
		}
	}
	return strings.Join(kept, "\n")
}

func vttTimestamp(ms int64) string {
	hours := ms / 3_600_000
	minutes := ms / 60_000 % 60
	seconds := ms / 1000 % 60
	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, ms%1000)
}
