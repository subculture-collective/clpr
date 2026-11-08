package countdownrender

import (
	"crypto/sha256"
	"math"
	"reflect"
	"strings"
	"testing"
)

func sha256Sum(data []byte) [32]byte { return sha256.Sum256(data) }

func TestCleanText(t *testing.T) {
	cases := map[string]string{
		"  🔥 Clutch\t1v5\n ace 🔥 ": "Clutch 1v5 ace",
		"a\u200db\ufe0fc\u0007d":   "abcd",
		"日本語  のタイトル":               "日本語 のタイトル",
		"ok ✅ done":                "ok done",
	}
	for in, want := range cases {
		if got := cleanText(in); got != want {
			t.Errorf("cleanText(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestWrap(t *testing.T) {
	cases := []struct {
		in    string
		width int
		lines int
		want  []string
	}{
		{"short title", 20, 2, []string{"short title"}},
		{"the quick brown fox jumps over the lazy dog", 16, 2, []string{"the quick brown", "fox jumps over…"}},
		{"日本語のタイトルとても長い", 10, 2, []string{"日本語のタ", "イトルと…"}},
		{"", 10, 2, nil},
		{"one two three", 40, 1, []string{"one two three"}},
		{"averyveryverylongwordwithoutspaces", 10, 2, []string{"averyveryv", "erylongwo…"}},
	}
	for _, tc := range cases {
		got := wrap(tc.in, tc.width, tc.lines)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("wrap(%q, %d, %d) = %q, want %q", tc.in, tc.width, tc.lines, got, tc.want)
		}
		for _, line := range got {
			if textWidth(line) > tc.width {
				t.Errorf("wrap(%q) line %q is wider than %d", tc.in, line, tc.width)
			}
		}
	}
}

func TestTemplateValidation(t *testing.T) {
	if err := DefaultTemplate().Validate(); err != nil {
		t.Fatalf("default template: %v", err)
	}
	bad := []func(*Template){
		func(t *Template) { t.Accent = "red" },
		func(t *Template) { t.FontFile = "/fonts/a.ttf:fontcolor=red" },
		func(t *Template) { t.BoldFontFile = "relative.ttf" },
		func(t *Template) { t.VideoPreset = "fast,drawtext" },
		func(t *Template) { t.AudioBitrate = "160" },
		func(t *Template) { t.FrameRate = 240 },
		func(t *Template) { t.LoudnessLUFS = 0 },
	}
	for i, mutate := range bad {
		tmpl := DefaultTemplate()
		mutate(&tmpl)
		if err := tmpl.Validate(); err == nil {
			t.Errorf("case %d: invalid template accepted", i)
		}
	}
	if !strings.HasPrefix(DefaultTemplate().FontFile, "/usr/share/fonts/") {
		t.Fatal("default font moved; update Dockerfile.render")
	}
}

func TestCaptions(t *testing.T) {
	words := []Word{
		{0.2, 0.5, "No"}, {0.5, 0.8, "way"}, {0.8, 1.1, "he"}, {1.1, 1.4, "actually"}, {1.4, 1.9, "hit"},
		{1.9, 2.2, "that"}, {2.2, 2.3, "🔥"}, {4.0, 4.2, "what"}, {4.2, 4.21, "?"},
		{5.0, 5.4, "Supercalifragilisticexpialidocious-and-more"},
	}
	got := Captions(words, 16)
	want := []Caption{
		{0.2, 1.1, "No way he"},        // "No way he actually" is 18 columns
		{1.1, 1.9, "actually hit"},     // "actually hit that" is 17
		{1.9, 2.6, "that"},             // the emoji is dropped; lingers 0.4 s
		{4.0, 4.61, "what ?"},          // a pause over a second starts a line
		{5.0, 5.8, "Supercalifragil…"}, // a word wider than a line is cut
	}
	if len(got) != len(want) {
		t.Fatalf("captions = %+v", got)
	}
	for i := range want {
		if got[i].Text != want[i].Text || math.Abs(got[i].Start-want[i].Start) > 1e-9 || math.Abs(got[i].End-want[i].End) > 1e-9 {
			t.Errorf("caption %d = %+v, want %+v", i, got[i], want[i])
		}
	}
	if Captions(nil, 16) != nil {
		t.Error("no words should give no captions")
	}
	// A short line is held for at least half a second, pushing the next one.
	short := Captions([]Word{{1.0, 1.05, "a"}, {1.06, 1.1, "b"}, {1.2, 1.3, "verylongwordhere"}}, 8)
	if len(short) != 2 || short[0].End != 1.5 || short[1].Start != 1.5 || short[1].End < short[1].Start {
		t.Fatalf("short captions = %+v", short)
	}
}
