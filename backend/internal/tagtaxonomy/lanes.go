package tagtaxonomy

import (
	"strings"
	"unicode"
)

// Lane names the process that produced a tag. Clients colour and order tags by
// lane instead of by per-tag colour.
type Lane string

const (
	// LaneCategory tags come from the clip's Twitch category and its genre
	// lookup (game/<slug>).
	LaneCategory Lane = "category"
	// LaneDetected tags are catalog content tags (content/<slug>) chosen by
	// the vision tagger or the title rules.
	LaneDetected Lane = "detected"
	// LaneStreamer tags are the broadcaster's own Twitch channel tags, copied
	// during clip sync. Older rows were stored under content/ with slugs that
	// are not in ContentCatalog; newer rows use streamer/<slug>.
	LaneStreamer Lane = "streamer"
	// LaneCommunity tags are unnamespaced or under community/. They include
	// tags clpr users add and older tags created before namespaces existed.
	LaneCommunity Lane = "community"
	// LaneDuration tags bucket clip length (duration/<slug>).
	LaneDuration Lane = "duration"
	// LaneLanguage tags record the clip language (lang/<code>).
	LaneLanguage Lane = "language"
	// LaneRoot is a namespace root such as "content" or "game".
	LaneRoot Lane = "root"
)

// Namespace roots created by migrations 000142 and later.
const (
	RootContent   = "content"
	RootGame      = "game"
	RootDuration  = "duration"
	RootLanguage  = "lang"
	RootCommunity = "community"
	RootStreamer  = "streamer"
)

var roots = map[string]bool{
	RootContent: true, RootGame: true, RootDuration: true,
	RootLanguage: true, RootCommunity: true, RootStreamer: true,
}

// Labels is the client-facing interpretation of a stored tag.
type Labels struct {
	Lane        Lane
	DisplayName string
	// Evidence is set only for catalog content tags. It describes what the
	// tag definition requires, not a verdict about any particular clip.
	Evidence Evidence
}

// Classify derives the lane, a clean display name, and catalog evidence for a
// stored tag slug and name.
func Classify(slug, name string) Labels {
	if roots[slug] {
		return Labels{Lane: LaneRoot, DisplayName: cleanName(name, slug)}
	}
	namespace, rest, namespaced := strings.Cut(slug, "/")
	if !namespaced {
		return Labels{Lane: LaneCommunity, DisplayName: cleanName(name, slug)}
	}
	switch namespace {
	case RootContent:
		if tag, ok := LookupContent(rest); ok {
			return Labels{Lane: LaneDetected, DisplayName: tag.Name, Evidence: tag.Evidence}
		}
		return Labels{Lane: LaneStreamer, DisplayName: cleanName(name, rest)}
	case RootStreamer:
		return Labels{Lane: LaneStreamer, DisplayName: cleanName(name, rest)}
	case RootGame:
		return Labels{Lane: LaneCategory, DisplayName: cleanName(name, rest)}
	case RootDuration:
		return Labels{Lane: LaneDuration, DisplayName: cleanName(name, rest)}
	case RootLanguage:
		return Labels{Lane: LaneLanguage, DisplayName: cleanName(name, rest)}
	case RootCommunity:
		return Labels{Lane: LaneCommunity, DisplayName: cleanName(name, rest)}
	default:
		return Labels{Lane: LaneCommunity, DisplayName: cleanName(name, rest)}
	}
}

// storedPrefixes are the machine prefixes migrations and taggers put in front
// of tag names ("Content: boss-fight", "Language: English").
var storedPrefixes = []string{
	"taxonomy:", "content:", "game:", "language:", "lang:",
	"duration:", "community:", "streamer:",
}

// cleanName strips a stored namespace prefix and, when what remains is just
// the slug, turns it into words ("boss-fight" -> "Boss Fight").
func cleanName(name, slugTail string) string {
	trimmed := strings.TrimSpace(name)
	lower := strings.ToLower(trimmed)
	for _, prefix := range storedPrefixes {
		if strings.HasPrefix(lower, prefix) {
			trimmed = strings.TrimSpace(trimmed[len(prefix):])
			lower = strings.ToLower(trimmed)
		}
	}
	if trimmed == "" {
		trimmed = slugTail
	}
	if trimmed == slugTail || lower == strings.ReplaceAll(slugTail, "-", " ") {
		return titleWords(strings.ReplaceAll(trimmed, "-", " "))
	}
	return trimmed
}

// initialisms keeps common lowercase slugs readable when they become labels.
var initialisms = map[string]string{
	"irl": "IRL", "asmr": "ASMR", "adhd": "ADHD", "vtuber": "VTuber",
	"mmorpg": "MMORPG", "rpg": "RPG", "fps": "FPS", "moba": "MOBA",
	"pvp": "PvP", "pve": "PvE", "lgbtq": "LGBTQ", "lgbtqia": "LGBTQIA",
	"diy": "DIY", "ai": "AI", "ttrpg": "TTRPG", "dj": "DJ",
}

func titleWords(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if fixed, ok := initialisms[strings.ToLower(word)]; ok {
			words[i] = fixed
			continue
		}
		runes := []rune(word)
		runes[0] = unicode.ToUpper(runes[0])
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

// SplitLabel turns a Twitch channel tag such as "CommunityOriented" into
// "Community Oriented" so streamer tags keep readable names when stored.
func SplitLabel(label string) string {
	runes := []rune(strings.TrimSpace(label))
	var out []rune
	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) && unicode.IsLower(runes[i-1]) {
			out = append(out, ' ')
		}
		out = append(out, r)
	}
	return string(out)
}

// DetectedSlugs lists the full content/<slug> slugs of every catalog tag.
func DetectedSlugs() []string {
	slugs := make([]string, 0, len(ContentCatalog))
	for _, tag := range ContentCatalog {
		slugs = append(slugs, RootContent+"/"+tag.Slug)
	}
	return slugs
}
