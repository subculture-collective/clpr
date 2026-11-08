// Package countdownrender turns the fetched clips of a countdown edition into
// a 16:9 countdown video, 9:16 Shorts and a thumbnail with ffmpeg (#78,
// ADR-002). Layout, colours, fonts and timings come from a Template, so they
// change as data rather than code.
package countdownrender

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

//go:embed template.json
var defaultTemplateJSON []byte

// Template holds every visual and timing choice of the renders.
type Template struct {
	FontFile           string  `json:"font_file"`
	BoldFontFile       string  `json:"bold_font_file"`
	Background         string  `json:"background"`
	Accent             string  `json:"accent"`
	Text               string  `json:"text"`
	Muted              string  `json:"muted"`
	FrameRate          int     `json:"frame_rate"`
	RankCardSeconds    float64 `json:"rank_card_seconds"`
	OutroSeconds       float64 `json:"outro_seconds"`
	LowerThirdSeconds  float64 `json:"lower_third_seconds"`
	TitleMaxChars      int     `json:"title_max_chars"`
	ShortTitleMaxChars int     `json:"short_title_max_chars"`
	CaptionMaxChars    int     `json:"caption_max_chars"`
	SiteLabel          string  `json:"site_label"`
	OutroHeading       string  `json:"outro_heading"`
	OutroFooter        string  `json:"outro_footer"`
	VideoCRF           int     `json:"video_crf"`
	VideoPreset        string  `json:"video_preset"`
	AudioBitrate       string  `json:"audio_bitrate"`
	LoudnessLUFS       float64 `json:"loudness_lufs"`
}

var (
	colorPattern   = regexp.MustCompile(`^0x[0-9A-Fa-f]{6}$`)
	presetPattern  = regexp.MustCompile(`^(ultrafast|superfast|veryfast|faster|fast|medium|slow|slower|veryslow)$`)
	bitratePattern = regexp.MustCompile(`^[0-9]{2,4}k$`)
	// Font paths go into ffmpeg filter graphs unescaped.
	fontPathPattern = regexp.MustCompile(`^/[A-Za-z0-9_./-]+$`)
)

// DefaultTemplate returns the built-in template.
func DefaultTemplate() Template {
	var t Template
	if err := json.Unmarshal(defaultTemplateJSON, &t); err != nil {
		panic(fmt.Sprintf("countdownrender: built-in template: %v", err))
	}
	return t
}

// LoadTemplate reads a template file over the defaults; fields missing from
// the file keep their default value.
func LoadTemplate(path string) (Template, error) {
	t := DefaultTemplate()
	if path == "" {
		return t, t.Validate()
	}
	data, err := os.ReadFile(path) // #nosec G304 -- operator-chosen template path.
	if err != nil {
		return t, err
	}
	if err := json.Unmarshal(data, &t); err != nil {
		return t, fmt.Errorf("template %s: %w", path, err)
	}
	return t, t.Validate()
}

// Validate rejects values that would break or inject into the ffmpeg filter
// graphs.
func (t Template) Validate() error {
	for name, c := range map[string]string{"background": t.Background, "accent": t.Accent, "text": t.Text, "muted": t.Muted} {
		if !colorPattern.MatchString(c) {
			return fmt.Errorf("template %s must look like 0xRRGGBB", name)
		}
	}
	switch {
	case !fontPathPattern.MatchString(t.FontFile) || !fontPathPattern.MatchString(t.BoldFontFile):
		return fmt.Errorf("template font files must be absolute paths of letters, digits, '.', '_', '-' and '/'")
	case t.FrameRate < 15 || t.FrameRate > 60:
		return fmt.Errorf("template frame_rate must be 15 to 60")
	case t.RankCardSeconds < 0.5 || t.RankCardSeconds > 10 || t.OutroSeconds < 1 || t.OutroSeconds > 30:
		return fmt.Errorf("template card durations are out of range")
	case t.LowerThirdSeconds < 0 || t.LowerThirdSeconds > 60:
		return fmt.Errorf("template lower_third_seconds must be 0 to 60")
	case t.TitleMaxChars < 10 || t.ShortTitleMaxChars < 10:
		return fmt.Errorf("template title widths must be at least 10")
	case t.CaptionMaxChars < 8 || t.CaptionMaxChars > 40:
		return fmt.Errorf("template caption_max_chars must be 8 to 40")
	case t.VideoCRF < 10 || t.VideoCRF > 35 || !presetPattern.MatchString(t.VideoPreset) || !bitratePattern.MatchString(t.AudioBitrate):
		return fmt.Errorf("template encoder settings are out of range")
	case t.LoudnessLUFS < -30 || t.LoudnessLUFS > -8:
		return fmt.Errorf("template loudness_lufs must be -30 to -8")
	}
	return nil
}
