package countdownrender

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Golden frames are compared at this size; a mean difference per channel
// above the tolerance fails. Regenerate with UPDATE_GOLDEN=1 inside the
// render image (see backend/Dockerfile.render).
const (
	goldenWidth, goldenHeight = 320, 180
	goldenTolerance           = 4.0
)

func requireRenderTools(t *testing.T) Template {
	t.Helper()
	tmpl := DefaultTemplate()
	for _, tool := range []string{"ffmpeg", "ffprobe"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s is not installed", tool)
		}
	}
	for _, font := range []string{tmpl.FontFile, tmpl.BoldFontFile} {
		if _, err := os.Stat(font); err != nil {
			t.Skipf("font %s is not installed", font)
		}
	}
	// Short cards keep the test fast; layout is what the goldens check.
	tmpl.RankCardSeconds, tmpl.OutroSeconds, tmpl.LowerThirdSeconds = 1, 2, 2
	return tmpl
}

func ffmpegFixture(t *testing.T, out string, args ...string) {
	t.Helper()
	cmd := exec.Command("ffmpeg", append(append([]string{"-nostdin", "-loglevel", "error", "-y"}, args...),
		"-c:v", "libx264", "-preset", "ultrafast", "-pix_fmt", "yuv420p", "-fflags", "+bitexact", out)...)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("fixture %s: %v: %s", filepath.Base(out), err, output)
	}
}

func fixtureEdition(t *testing.T, dir string) Edition {
	t.Helper()
	landscape := filepath.Join(dir, "landscape.mp4")
	portrait := filepath.Join(dir, "portrait.mp4")
	bars := filepath.Join(dir, "bars.mp4")
	ffmpegFixture(t, landscape, "-f", "lavfi", "-i", "testsrc2=s=1280x720:r=30:d=4", "-f", "lavfi", "-i", "sine=f=440:d=4",
		"-c:a", "aac", "-shortest")
	ffmpegFixture(t, portrait, "-f", "lavfi", "-i", "testsrc2=s=720x1280:r=30:d=4")
	ffmpegFixture(t, bars, "-f", "lavfi", "-i", "smptebars=s=1920x1080:r=30:d=3", "-f", "lavfi", "-i", "sine=f=220:d=3",
		"-c:a", "aac", "-shortest")
	return Edition{SeriesName: "Overall", Date: "2026-11-07", Clips: []Clip{
		{Rank: 1, Path: landscape, Title: `50% "quoted" \ it's a: test, with {braces} and a very long tail that must be cut`, Broadcaster: "alpha_streams"},
		{Rank: 2, Path: portrait, Title: "日本語のタイトル とても長い説明文がここに入ります 🔥", Broadcaster: "ブラボー"},
		{Rank: 3, Path: bars, Title: "🔥 Clutch 1v5 ace", Broadcaster: "charlie"},
	}}
}

func probeStreams(t *testing.T, path string) string {
	t.Helper()
	out, err := exec.Command("ffprobe", "-v", "error", "-show_entries",
		"stream=codec_type,codec_name,width,height,sample_rate,channels:format=duration", "-of", "compact=p=0:nk=0", path).Output()
	if err != nil {
		t.Fatalf("ffprobe %s: %v", path, err)
	}
	return string(out)
}

func durationOf(t *testing.T, path string) float64 {
	t.Helper()
	out, err := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "csv=p=0", path).Output()
	if err != nil {
		t.Fatal(err)
	}
	var d float64
	if _, err := fmt.Sscanf(strings.TrimSpace(string(out)), "%f", &d); err != nil {
		t.Fatalf("duration %q: %v", out, err)
	}
	return d
}

// frameAt extracts one frame at seconds, scaled to the golden size.
func frameAt(t *testing.T, path string, at float64, width, height int) image.Image {
	t.Helper()
	out := filepath.Join(t.TempDir(), "frame.png")
	args := []string{"-nostdin", "-loglevel", "error", "-y"}
	if at >= 0 {
		args = append(args, "-ss", fmt.Sprintf("%.3f", at))
	}
	args = append(args, "-i", path, "-frames:v", "1", "-vf", fmt.Sprintf("scale=%d:%d:flags=area", width, height), out)
	if output, err := exec.Command("ffmpeg", args...).CombinedOutput(); err != nil {
		t.Fatalf("extract frame: %v: %s", err, output)
	}
	f, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	return img
}

func meanDifference(a, b image.Image) float64 {
	bounds := a.Bounds()
	if bounds != b.Bounds() {
		return math.Inf(1)
	}
	var sum float64
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r1, g1, b1, _ := a.At(x, y).RGBA()
			r2, g2, b2, _ := b.At(x, y).RGBA()
			sum += math.Abs(float64(r1)-float64(r2)) + math.Abs(float64(g1)-float64(g2)) + math.Abs(float64(b1)-float64(b2))
		}
	}
	return sum / float64(bounds.Dx()*bounds.Dy()*3) / 257
}

func checkGolden(t *testing.T, name string, img image.Image) {
	t.Helper()
	path := filepath.Join("testdata", name+".png")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if err := png.Encode(f, img); err != nil {
			t.Fatal(err)
		}
		return
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("golden %s: %v (run with UPDATE_GOLDEN=1 to create it)", name, err)
	}
	defer f.Close()
	want, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	if diff := meanDifference(img, want); diff > goldenTolerance {
		t.Errorf("frame %s differs from its golden by %.2f per channel (tolerance %.1f)", name, diff, goldenTolerance)
	}
}

func TestRenderCountdownGoldenFrames(t *testing.T) {
	tmpl := requireRenderTools(t)
	dir := t.TempDir()
	ed := fixtureEdition(t, dir)
	r := &Renderer{Template: tmpl, WorkDir: filepath.Join(dir, "work")}
	out := filepath.Join(dir, "countdown.mp4")
	if err := r.RenderCountdown(context.Background(), ed, out); err != nil {
		t.Fatal(err)
	}

	streams := probeStreams(t, out)
	for _, want := range []string{"codec_name=h264|codec_type=video|width=1920|height=1080", "codec_name=aac|codec_type=audio|sample_rate=48000|channels=2"} {
		if !strings.Contains(streams, want) {
			t.Fatalf("streams %q lack %q", streams, want)
		}
	}
	// Countdown order #3, #2, #1: three 1 s cards, clips of 3, 4 and 4 s, a 2 s outro.
	if d := durationOf(t, out); math.Abs(d-16) > 0.35 {
		t.Fatalf("duration %.2f s, want about 16", d)
	}
	for _, f := range []struct {
		name string
		at   float64
	}{
		{"countdown-card-3", 0.5},
		{"countdown-clip-3", 2.0},
		{"countdown-clip-2-portrait", 5.5},
		{"countdown-card-1", 9.5},
		{"countdown-clip-1", 11.0},
		{"countdown-clip-1-after-lower-third", 13.5},
		{"countdown-outro", 15.0},
	} {
		checkGolden(t, f.name, frameAt(t, out, f.at, goldenWidth, goldenHeight))
	}
	if entries, _ := os.ReadDir(r.WorkDir); len(entries) != 0 {
		t.Fatalf("work files left behind: %v", entries)
	}
}

func TestRenderShortAndThumbnailGoldenFrames(t *testing.T) {
	tmpl := requireRenderTools(t)
	dir := t.TempDir()
	ed := fixtureEdition(t, dir)
	r := &Renderer{Template: tmpl, WorkDir: filepath.Join(dir, "work")}

	short := filepath.Join(dir, "short.mp4")
	if err := r.RenderShort(context.Background(), ed.Clips[0], short); err != nil {
		t.Fatal(err)
	}
	if streams := probeStreams(t, short); !strings.Contains(streams, "width=1080|height=1920") {
		t.Fatalf("short streams %q", streams)
	}
	if d := durationOf(t, short); math.Abs(d-4) > 0.2 {
		t.Fatalf("short lasts %.2f s, want 4", d)
	}
	checkGolden(t, "short-landscape", frameAt(t, short, 1.0, goldenHeight, goldenWidth))

	silent := filepath.Join(dir, "short-silent.mp4")
	if err := r.RenderShort(context.Background(), ed.Clips[1], silent); err != nil {
		t.Fatal(err)
	}
	if streams := probeStreams(t, silent); !strings.Contains(streams, "codec_type=audio") {
		t.Fatalf("a clip without audio got no silent track: %q", streams)
	}
	checkGolden(t, "short-portrait", frameAt(t, silent, 1.0, goldenHeight, goldenWidth))

	thumb := filepath.Join(dir, "thumb.jpg")
	if err := r.RenderThumbnail(context.Background(), ed, thumb); err != nil {
		t.Fatal(err)
	}
	if streams := probeStreams(t, thumb); !strings.Contains(streams, "width=1280|height=720") {
		t.Fatalf("thumbnail streams %q", streams)
	}
	checkGolden(t, "thumbnail", frameAt(t, thumb, -1, goldenWidth, goldenHeight))
}

func TestRenderIsDeterministic(t *testing.T) {
	tmpl := requireRenderTools(t)
	dir := t.TempDir()
	ed := fixtureEdition(t, dir)
	ed.Clips = ed.Clips[:1]
	r := &Renderer{Template: tmpl, WorkDir: filepath.Join(dir, "work")}
	var sums []string
	for i := 0; i < 2; i++ {
		out := filepath.Join(dir, fmt.Sprintf("run-%d.mp4", i))
		if err := r.RenderCountdown(context.Background(), ed, out); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(out)
		if err != nil {
			t.Fatal(err)
		}
		sums = append(sums, fmt.Sprintf("%x", sha256Sum(data)))
	}
	if sums[0] != sums[1] {
		t.Fatalf("two renders of the same inputs differ: %s and %s", sums[0], sums[1])
	}
}

// TestRenderTenClipTiming measures a full-size edition: ten 30 s 1080p60
// clips with audio and the default template. It runs only with
// COUNTDOWN_RENDER_TIMING=1; #78 requires under 10 minutes on the render host.
func TestRenderTenClipTiming(t *testing.T) {
	if os.Getenv("COUNTDOWN_RENDER_TIMING") != "1" {
		t.Skip("set COUNTDOWN_RENDER_TIMING=1 to measure")
	}
	requireRenderTools(t)
	dir := t.TempDir()
	ed := Edition{SeriesName: "Overall", Date: "2026-11-07"}
	for i := 1; i <= 10; i++ {
		path := filepath.Join(dir, fmt.Sprintf("clip-%02d.mp4", i))
		ffmpegFixture(t, path, "-f", "lavfi", "-i", "testsrc2=s=1920x1080:r=60:d=30", "-f", "lavfi", "-i",
			fmt.Sprintf("sine=f=%d:d=30", 200+40*i), "-c:a", "aac", "-shortest")
		ed.Clips = append(ed.Clips, Clip{Rank: i, Path: path, Title: fmt.Sprintf("Clip number %d with a typical title", i), Broadcaster: "broadcaster"})
	}
	r := &Renderer{Template: DefaultTemplate(), WorkDir: filepath.Join(dir, "work")}
	start := time.Now()
	if err := r.RenderCountdown(context.Background(), ed, filepath.Join(dir, "countdown.mp4")); err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(start)
	t.Logf("10-clip countdown rendered in %s", elapsed.Round(time.Second))
	if elapsed > 10*time.Minute {
		t.Fatalf("render took %s, over the 10 minute budget", elapsed)
	}
	start = time.Now()
	if err := r.RenderShort(context.Background(), ed.Clips[0], filepath.Join(dir, "short.mp4")); err != nil {
		t.Fatal(err)
	}
	t.Logf("30 s Short rendered in %s", time.Since(start).Round(time.Second))
}
