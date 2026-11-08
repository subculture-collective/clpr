package countdownrender

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Video sizes of the outputs.
const (
	countdownWidth, countdownHeight = 1920, 1080
	shortWidth, shortHeight         = 1080, 1920
	thumbWidth, thumbHeight         = 1280, 720
)

// ErrToolMissing means ffmpeg or ffprobe is not installed.
var ErrToolMissing = errors.New("ffmpeg or ffprobe is missing")

// Clip is one ranked clip with its fetched file. Rank 1 is the top clip.
type Clip struct {
	Rank        int
	Path        string
	Title       string
	Broadcaster string
	// Words, when set, are burned into the clip's Short as captions (#81).
	Words []Word
}

// Word is one transcribed word with its time in the clip, in seconds.
type Word struct {
	Start, End float64
	Text       string
}

// Caption is one line shown on a Short from Start to End.
type Caption struct {
	Start, End float64
	Text       string
}

// Captions groups words into lines of at most maxChars columns. A line shows
// from its first word until the next line starts, or briefly after its last
// word when speech pauses.
func Captions(words []Word, maxChars int) []Caption {
	const linger, minimum = 0.4, 0.5
	var out []Caption
	var line []string
	var start, end float64
	flush := func() {
		if len(line) > 0 {
			out = append(out, Caption{Start: start, End: end + linger, Text: strings.Join(line, " ")})
			line = nil
		}
	}
	for _, w := range words {
		text := cleanText(w.Text)
		if text == "" || w.End <= w.Start {
			continue
		}
		text = truncate(text, maxChars)
		joined := strings.Join(append(append([]string{}, line...), text), " ")
		if len(line) > 0 && (textWidth(joined) > maxChars || w.Start-end > 1.0) {
			flush()
		}
		if len(line) == 0 {
			start = w.Start
		}
		line = append(line, text)
		end = w.End
	}
	flush()
	for i := range out {
		if i+1 < len(out) && out[i].End > out[i+1].Start {
			out[i].End = out[i+1].Start
		}
		if out[i].End-out[i].Start < minimum {
			out[i].End = out[i].Start + minimum
			if i+1 < len(out) && out[i].End > out[i+1].Start {
				out[i+1].Start = out[i].End
			}
		}
	}
	return out
}

// Edition is what a countdown video shows. Clips are in rank order, 1 first.
type Edition struct {
	SeriesName string
	Date       string
	Clips      []Clip
}

// Renderer runs ffmpeg with a Template.
type Renderer struct {
	FFmpegPath  string
	FFprobePath string
	Template    Template
	// WorkDir holds intermediate segments and text files; each render uses
	// and removes its own subdirectory.
	WorkDir string
}

func (r *Renderer) ffmpeg() string {
	if r.FFmpegPath != "" {
		return r.FFmpegPath
	}
	return "ffmpeg"
}

func (r *Renderer) ffprobe() string {
	if r.FFprobePath != "" {
		return r.FFprobePath
	}
	return "ffprobe"
}

func (r *Renderer) run(ctx context.Context, args ...string) error {
	base := []string{"-nostdin", "-hide_banner", "-loglevel", "error", "-y"}
	cmd := exec.CommandContext(ctx, r.ffmpeg(), append(base, args...)...) // #nosec G204 -- fixed binary, arguments built here.
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		return nil
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return fmt.Errorf("%w: %v", ErrToolMissing, err)
	}
	message := strings.TrimSpace(stderr.String())
	if len(message) > 400 {
		message = message[len(message)-400:]
	}
	return fmt.Errorf("ffmpeg failed: %v: %s", err, message)
}

// probe returns a file's duration and whether it has an audio stream.
func (r *Renderer) probe(ctx context.Context, path string) (float64, bool, error) {
	out, err := exec.CommandContext(ctx, r.ffprobe(), "-v", "error", "-show_entries", // #nosec G204 -- fixed binary.
		"format=duration:stream=codec_type", "-of", "default=noprint_wrappers=1", "--", path).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			return 0, false, fmt.Errorf("%w: %v", ErrToolMissing, err)
		}
		return 0, false, fmt.Errorf("ffprobe could not read %s", filepath.Base(path))
	}
	duration, audio := 0.0, false
	for _, line := range strings.Split(string(out), "\n") {
		key, value, _ := strings.Cut(strings.TrimSpace(line), "=")
		switch {
		case key == "duration":
			duration, _ = strconv.ParseFloat(value, 64)
		case key == "codec_type" && value == "audio":
			audio = true
		}
	}
	if duration <= 0 || math.IsNaN(duration) || math.IsInf(duration, 0) {
		return 0, false, fmt.Errorf("%s has no duration", filepath.Base(path))
	}
	return duration, audio, nil
}

// session is one render's scratch directory for text files and segments.
type session struct {
	dir   string
	count int
}

func (r *Renderer) newSession() (*session, error) {
	if err := os.MkdirAll(r.WorkDir, 0o700); err != nil {
		return nil, err
	}
	dir, err := os.MkdirTemp(r.WorkDir, "render-")
	if err != nil {
		return nil, err
	}
	return &session{dir: dir}, nil
}

// textFile writes text for drawtext's textfile option, so titles never pass
// through filter-graph escaping.
func (s *session) textFile(lines ...string) (string, error) {
	s.count++
	path := filepath.Join(s.dir, fmt.Sprintf("text-%03d.txt", s.count))
	text := strings.Join(lines, "\n")
	if strings.TrimSpace(text) == "" {
		text = " " // drawtext refuses an empty text
	}
	return path, os.WriteFile(path, []byte(text), 0o600)
}

func (s *session) file(name string) string {
	s.count++
	return filepath.Join(s.dir, fmt.Sprintf("%03d-%s", s.count, name))
}

// drawtext builds one drawtext filter. Every value is generated here or
// validated by Template.Validate, so no filter-graph escaping is needed.
func drawtext(font, textFile string, size int, color, x, y string, extra ...string) string {
	parts := []string{"fontfile=" + font, "textfile=" + textFile, "expansion=none",
		"fontsize=" + strconv.Itoa(size), "fontcolor=" + color, "x=" + x, "y=" + y}
	return "drawtext=" + strings.Join(append(parts, extra...), ":")
}

func (r *Renderer) encodeArgs() []string {
	t := r.Template
	return []string{
		"-c:v", "libx264", "-preset", t.VideoPreset, "-crf", strconv.Itoa(t.VideoCRF), "-pix_fmt", "yuv420p",
		"-profile:v", "high", "-r", strconv.Itoa(t.FrameRate), "-g", strconv.Itoa(2 * t.FrameRate),
		"-c:a", "aac", "-b:a", t.AudioBitrate, "-ar", "48000", "-ac", "2",
		"-fflags", "+bitexact", "-flags:v", "+bitexact", "-flags:a", "+bitexact", "-map_metadata", "-1",
		"-movflags", "+faststart",
	}
}

func seconds(v float64) string { return strconv.FormatFloat(v, 'f', 3, 64) }

// card renders a still card of the given length over the background colour
// with silent audio.
func (r *Renderer) card(ctx context.Context, out string, width, height int, duration float64, filters []string) error {
	t := r.Template
	args := []string{
		"-f", "lavfi", "-i", fmt.Sprintf("color=c=%s:s=%dx%d:r=%d:d=%s", t.Background, width, height, t.FrameRate, seconds(duration)),
		"-f", "lavfi", "-i", "anullsrc=r=48000:cl=stereo",
		"-filter_complex", "[0:v]" + strings.Join(append(filters, "format=yuv420p"), ",") + "[v]",
		"-map", "[v]", "-map", "1:a", "-t", seconds(duration),
	}
	return r.run(ctx, append(append(args, r.encodeArgs()...), out)...)
}

// audioChain normalizes loudness and format so segments concatenate cleanly.
func (r *Renderer) audioChain(in string) string {
	return fmt.Sprintf("[%s]aresample=48000,aformat=channel_layouts=stereo,loudnorm=I=%s:TP=-1.5:LRA=11,aresample=48000[a]",
		in, strconv.FormatFloat(r.Template.LoudnessLUFS, 'f', 1, 64))
}

// video renders one input clip through a video filter chain, normalizing
// its audio or adding silence when it has none.
func (r *Renderer) video(ctx context.Context, in, out, videoChain string) error {
	duration, audio, err := r.probe(ctx, in)
	if err != nil {
		return err
	}
	args := []string{"-i", in}
	graph := "[0:v]" + videoChain + "[v];"
	if audio {
		graph += r.audioChain("0:a")
	} else {
		args = append(args, "-f", "lavfi", "-i", "anullsrc=r=48000:cl=stereo")
		graph += "[1:a]anull[a]"
	}
	args = append(args, "-filter_complex", graph, "-map", "[v]", "-map", "[a]", "-t", seconds(duration))
	return r.run(ctx, append(append(args, r.encodeArgs()...), out)...)
}

func (r *Renderer) rankCard(ctx context.Context, s *session, c Clip, out string) error {
	t := r.Template
	rank, err := s.textFile(fmt.Sprintf("#%d", c.Rank))
	if err != nil {
		return err
	}
	title, err := s.textFile(wrap(c.Title, t.TitleMaxChars, 2)...)
	if err != nil {
		return err
	}
	broadcaster, err := s.textFile(truncate(cleanText(c.Broadcaster), 40))
	if err != nil {
		return err
	}
	return r.card(ctx, out, countdownWidth, countdownHeight, t.RankCardSeconds, []string{
		drawtext(t.BoldFontFile, rank, 260, t.Accent, "(w-text_w)/2", "h*0.14"),
		drawtext(t.BoldFontFile, title, 64, t.Text, "(w-text_w)/2", "h*0.54", "line_spacing=14", "text_align=C"),
		drawtext(t.FontFile, broadcaster, 46, t.Muted, "(w-text_w)/2", "h*0.80"),
	})
}

func (r *Renderer) clipSegment(ctx context.Context, s *session, c Clip, out string) error {
	t := r.Template
	rank, err := s.textFile(fmt.Sprintf("#%d", c.Rank))
	if err != nil {
		return err
	}
	broadcaster, err := s.textFile(truncate(cleanText(c.Broadcaster), 40))
	if err != nil {
		return err
	}
	title, err := s.textFile(truncate(cleanText(c.Title), 70))
	if err != nil {
		return err
	}
	site, err := s.textFile(t.SiteLabel)
	if err != nil {
		return err
	}
	enable := []string{}
	if t.LowerThirdSeconds > 0 {
		enable = []string{"enable=lt(t\\," + seconds(t.LowerThirdSeconds) + ")"}
	}
	chain := []string{
		fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease", countdownWidth, countdownHeight),
		fmt.Sprintf("pad=%d:%d:(ow-iw)/2:(oh-ih)/2:color=black", countdownWidth, countdownHeight),
		"setsar=1", "fps=" + strconv.Itoa(t.FrameRate),
		"drawbox=" + strings.Join(append([]string{"x=60", "y=ih-250", "w=iw-120", "h=170", "color=black@0.62", "t=fill"}, enable...), ":"),
		drawtext(t.BoldFontFile, rank, 80, t.Accent, "100", "h-238", enable...),
		drawtext(t.BoldFontFile, broadcaster, 50, t.Text, "290", "h-232", enable...),
		drawtext(t.FontFile, title, 36, t.Muted, "290", "h-158", enable...),
		drawtext(t.FontFile, site, 34, t.Text, "w-text_w-60", "48", "box=1", "boxcolor=black@0.45", "boxborderw=12"),
		"format=yuv420p",
	}
	return r.video(ctx, c.Path, out, strings.Join(chain, ","))
}

func (r *Renderer) outro(ctx context.Context, s *session, ed Edition, out string) error {
	t := r.Template
	heading, err := s.textFile(cleanText(t.OutroHeading))
	if err != nil {
		return err
	}
	lines := make([]string, 0, len(ed.Clips))
	for _, c := range ed.Clips {
		lines = append(lines, fmt.Sprintf("#%d  %s — %s", c.Rank, truncate(cleanText(c.Title), 50), truncate(cleanText(c.Broadcaster), 24)))
	}
	list, err := s.textFile(lines...)
	if err != nil {
		return err
	}
	footer, err := s.textFile(wrap(t.OutroFooter, 90, 2)...)
	if err != nil {
		return err
	}
	site, err := s.textFile(t.SiteLabel)
	if err != nil {
		return err
	}
	size := int(math.Min(38, math.Max(14, 700/(float64(len(lines))*1.6))))
	return r.card(ctx, out, countdownWidth, countdownHeight, t.OutroSeconds, []string{
		drawtext(t.BoldFontFile, heading, 72, t.Accent, "160", "90"),
		drawtext(t.FontFile, list, size, t.Text, "160", "220", "line_spacing="+strconv.Itoa(size*6/10)),
		drawtext(t.FontFile, footer, 30, t.Muted, "(w-text_w)/2", "h-150", "line_spacing=10", "text_align=C"),
		drawtext(t.BoldFontFile, site, 40, t.Text, "w-text_w-80", "90"),
	})
}

// RenderCountdown writes the 16:9 countdown video: for each clip from the
// lowest rank to #1, a rank card and the clip with a lower third; then an
// outro listing every clip.
func (r *Renderer) RenderCountdown(ctx context.Context, ed Edition, out string) error {
	if err := r.Template.Validate(); err != nil {
		return err
	}
	if len(ed.Clips) == 0 {
		return fmt.Errorf("edition has no clips")
	}
	s, err := r.newSession()
	if err != nil {
		return err
	}
	defer os.RemoveAll(s.dir)

	var segments []string
	for i := len(ed.Clips) - 1; i >= 0; i-- {
		c := ed.Clips[i]
		card, clip := s.file("card.mp4"), s.file("clip.mp4")
		if err := r.rankCard(ctx, s, c, card); err != nil {
			return fmt.Errorf("rank card #%d: %w", c.Rank, err)
		}
		if err := r.clipSegment(ctx, s, c, clip); err != nil {
			return fmt.Errorf("clip #%d: %w", c.Rank, err)
		}
		segments = append(segments, card, clip)
	}
	outro := s.file("outro.mp4")
	if err := r.outro(ctx, s, ed, outro); err != nil {
		return fmt.Errorf("outro: %w", err)
	}
	segments = append(segments, outro)

	var list strings.Builder
	for _, seg := range segments {
		fmt.Fprintf(&list, "file '%s'\n", seg)
	}
	listPath := filepath.Join(s.dir, "concat.txt")
	if err := os.WriteFile(listPath, []byte(list.String()), 0o600); err != nil {
		return err
	}
	return r.run(ctx, "-f", "concat", "-safe", "0", "-i", listPath, "-c", "copy",
		"-fflags", "+bitexact", "-map_metadata", "-1", "-movflags", "+faststart", out)
}

// RenderShort writes a 9:16 Short of one clip: the clip over a blurred,
// cropped copy of itself, with its title, broadcaster and the site label.
func (r *Renderer) RenderShort(ctx context.Context, c Clip, out string) error {
	t := r.Template
	if err := t.Validate(); err != nil {
		return err
	}
	s, err := r.newSession()
	if err != nil {
		return err
	}
	defer os.RemoveAll(s.dir)
	title, err := s.textFile(wrap(c.Title, t.ShortTitleMaxChars, 2)...)
	if err != nil {
		return err
	}
	broadcaster, err := s.textFile(truncate(cleanText(c.Broadcaster), 30))
	if err != nil {
		return err
	}
	site, err := s.textFile(t.SiteLabel)
	if err != nil {
		return err
	}
	chain := strings.Join([]string{
		"split=2[bgsrc][fgsrc];" +
			// The background is blurred at quarter size and scaled up: the
			// same look at a fraction of the cost of a full-size blur.
			fmt.Sprintf("[bgsrc]scale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d,boxblur=6:2,scale=%d:%d[bg];",
				shortWidth/4, shortHeight/4, shortWidth/4, shortHeight/4, shortWidth, shortHeight) +
			fmt.Sprintf("[fgsrc]scale=%d:-2[fg];", shortWidth) +
			"[bg][fg]overlay=(W-w)/2:(H-h)/2",
		"setsar=1", "fps=" + strconv.Itoa(t.FrameRate),
		"drawbox=x=0:y=150:w=iw:h=300:color=black@0.55:t=fill",
		drawtext(t.BoldFontFile, title, 66, t.Text, "(w-text_w)/2", "190", "line_spacing=16", "text_align=C"),
		drawtext(t.BoldFontFile, broadcaster, 52, t.Accent, "(w-text_w)/2", "h-360"),
		drawtext(t.FontFile, site, 42, t.Muted, "(w-text_w)/2", "h-280"),
	}, ",")
	for _, caption := range Captions(c.Words, t.CaptionMaxChars) {
		text, err := s.textFile(caption.Text)
		if err != nil {
			return err
		}
		// Below a landscape clip, which ends at y=1264; over a portrait one.
		chain += "," + drawtext(t.BoldFontFile, text, 64, t.Text, "(w-text_w)/2", "1330",
			"borderw=4", "bordercolor=black",
			"enable=between(t\\,"+seconds(caption.Start)+"\\,"+seconds(caption.End)+")")
	}
	return r.video(ctx, c.Path, out, chain+",format=yuv420p")
}

// RenderThumbnail writes a 1280x720 JPEG from the middle of the #1 clip with
// the countdown size, series and date.
func (r *Renderer) RenderThumbnail(ctx context.Context, ed Edition, out string) error {
	t := r.Template
	if err := t.Validate(); err != nil {
		return err
	}
	if len(ed.Clips) == 0 {
		return fmt.Errorf("edition has no clips")
	}
	top := ed.Clips[0]
	duration, _, err := r.probe(ctx, top.Path)
	if err != nil {
		return err
	}
	s, err := r.newSession()
	if err != nil {
		return err
	}
	defer os.RemoveAll(s.dir)
	headline, err := s.textFile(fmt.Sprintf("TOP %d", len(ed.Clips)))
	if err != nil {
		return err
	}
	subtitle, err := s.textFile(truncate(cleanText(ed.SeriesName+" · "+ed.Date), 44))
	if err != nil {
		return err
	}
	chain := strings.Join([]string{
		fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d", thumbWidth, thumbHeight, thumbWidth, thumbHeight),
		"setsar=1",
		"drawbox=x=0:y=ih-230:w=iw:h=230:color=black@0.65:t=fill",
		drawtext(t.BoldFontFile, headline, 104, t.Accent, "50", "h-215"),
		drawtext(t.BoldFontFile, subtitle, 46, t.Text, "54", "h-82"),
	}, ",")
	return r.run(ctx, "-ss", seconds(duration/2), "-i", top.Path, "-vf", chain, "-frames:v", "1", "-q:v", "3",
		"-fflags", "+bitexact", "-map_metadata", "-1", out)
}
