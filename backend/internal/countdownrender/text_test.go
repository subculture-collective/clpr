package countdownrender

import (
	"crypto/sha256"
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
