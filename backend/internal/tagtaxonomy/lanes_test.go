package tagtaxonomy

import "testing"

func TestClassify(t *testing.T) {
	cases := []struct {
		slug, name string
		want       Labels
	}{
		{"content/boss-fight", "Content: boss-fight", Labels{LaneDetected, "Boss Fight", Contextual}},
		{"content/funny", "Content: Funny", Labels{LaneDetected, "Funny", Strong}},
		{"content/noob", "Content: noob", Labels{LaneDetected, "Beginner", Strong}},
		{"content/english", "Content: english", Labels{Lane: LaneStreamer, DisplayName: "English"}},
		{"content/communityoriented", "Content: communityoriented", Labels{Lane: LaneStreamer, DisplayName: "Communityoriented"}},
		{"streamer/community-oriented", "Streamer: Community Oriented", Labels{Lane: LaneStreamer, DisplayName: "Community Oriented"}},
		{"game/just-chatting", "Game: Just Chatting", Labels{Lane: LaneCategory, DisplayName: "Just Chatting"}},
		{"lang/en", "Language: English", Labels{Lane: LaneLanguage, DisplayName: "English"}},
		{"duration/short", "Short (0-30s)", Labels{Lane: LaneDuration, DisplayName: "Short (0-30s)"}},
		{"goosebumps", "goosebumps", Labels{Lane: LaneCommunity, DisplayName: "Goosebumps"}},
		{"community/cozy", "Cozy", Labels{Lane: LaneCommunity, DisplayName: "Cozy"}},
		{"content", "Taxonomy: Content", Labels{Lane: LaneRoot, DisplayName: "Content"}},
		{"game/irl", "Game: irl", Labels{Lane: LaneCategory, DisplayName: "IRL"}},
		{"content/adhd", "Content: adhd", Labels{Lane: LaneStreamer, DisplayName: "ADHD"}},
	}
	for _, tc := range cases {
		if got := Classify(tc.slug, tc.name); got != tc.want {
			t.Errorf("Classify(%q, %q) = %+v, want %+v", tc.slug, tc.name, got, tc.want)
		}
	}
}

func TestSplitLabel(t *testing.T) {
	for in, want := range map[string]string{
		"CommunityOriented": "Community Oriented",
		"English":           "English",
		"ADHD":              "ADHD",
		"DropsEnabled":      "Drops Enabled",
	} {
		if got := SplitLabel(in); got != want {
			t.Errorf("SplitLabel(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDetectedSlugsCoverCatalog(t *testing.T) {
	slugs := DetectedSlugs()
	if len(slugs) != len(ContentCatalog) {
		t.Fatalf("got %d detected slugs for %d catalog entries", len(slugs), len(ContentCatalog))
	}
	for _, slug := range slugs {
		if got := Classify(slug, "").Lane; got != LaneDetected {
			t.Errorf("%s classified as %s", slug, got)
		}
	}
}
