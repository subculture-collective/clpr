package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/tagtaxonomy"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/google/uuid"
)

// AutoTagService handles automatic tag generation for clips
type AutoTagService struct {
	tagRepo *repository.TagRepository
}

// NewAutoTagService creates a new AutoTagService
func NewAutoTagService(tagRepo *repository.TagRepository) *AutoTagService {
	return &AutoTagService{
		tagRepo: tagRepo,
	}
}

// TagPattern represents a pattern to match for auto-tagging
type TagPattern struct {
	Pattern *regexp.Regexp
	TagName string
	TagSlug string
	Color   *string
}

type canonicalTagCandidate struct {
	name, slug, parent string
	color              *string
}

var (
	// Common gaming action patterns
	tagPatterns = []TagPattern{
		{
			Pattern: regexp.MustCompile(`(?i)\b(ace|5k|team wipe|team kill)\b`),
			TagName: "Ace",
			TagSlug: "ace",
			Color:   stringPtr("#FF6B6B"),
		},
		{
			Pattern: regexp.MustCompile(`(?i)\b(clutch|1v[2-5])\b`),
			TagName: "Clutch",
			TagSlug: "clutch",
			Color:   stringPtr("#FFA500"),
		},
		{
			Pattern: regexp.MustCompile(`(?i)\b(fail|fails|failed|epic fail)\b`),
			TagName: "Fail",
			TagSlug: "fail",
			Color:   stringPtr("#8B4513"),
		},
		{
			Pattern: regexp.MustCompile(`(?i)\b(rage|raging|angry|mad)\b`),
			TagName: "Rage",
			TagSlug: "rage",
			Color:   stringPtr("#DC143C"),
		},
		{
			Pattern: regexp.MustCompile(`(?i)\b(funny|lol|lmao|hilarious|comedy)\b`),
			TagName: "Funny",
			TagSlug: "funny",
			Color:   stringPtr("#FFD700"),
		},
		{
			Pattern: regexp.MustCompile(`(?i)\b(insane|crazy|amazing|incredible)\b`),
			TagName: "Insane",
			TagSlug: "insane",
			Color:   stringPtr("#9370DB"),
		},
		{
			Pattern: regexp.MustCompile(`(?i)\b(lucky|luck|rng)\b`),
			TagName: "Lucky",
			TagSlug: "lucky",
			Color:   stringPtr("#32CD32"),
		},
		{
			Pattern: regexp.MustCompile(`(?i)\b(bug|glitch|broken)\b`),
			TagName: "Bug",
			TagSlug: "bug",
			Color:   stringPtr("#696969"),
		},
		{
			Pattern: regexp.MustCompile(`(?i)\b(toxic|salt|salty)\b`),
			TagName: "Toxic",
			TagSlug: "toxic",
			Color:   stringPtr("#8B008B"),
		},
		{
			Pattern: regexp.MustCompile(`(?i)\b(epic|legendary|godlike)\b`),
			TagName: "Epic",
			TagSlug: "epic",
			Color:   stringPtr("#FF1493"),
		},
		{
			Pattern: regexp.MustCompile(`(?i)\b(noob|newbie|beginner)\b`),
			TagName: "Noob",
			TagSlug: "noob",
			Color:   stringPtr("#A9A9A9"),
		},
		{
			Pattern: regexp.MustCompile(`(?i)\b(pro|professional|skilled)\b`),
			TagName: "Pro",
			TagSlug: "pro",
			Color:   stringPtr("#4169E1"),
		},
		{
			Pattern: regexp.MustCompile(`(?i)\b(highlight|best|top)\b`),
			TagName: "Highlight",
			TagSlug: "highlight",
			Color:   stringPtr("#00CED1"),
		},
		{
			Pattern: regexp.MustCompile(`(?i)\b(speedrun|speed run|wr|world record)\b`),
			TagName: "Speedrun",
			TagSlug: "speedrun",
			Color:   stringPtr("#FF4500"),
		},
		{
			Pattern: regexp.MustCompile(`(?i)\b(tutorial|guide|how to)\b`),
			TagName: "Tutorial",
			TagSlug: "tutorial",
			Color:   stringPtr("#4682B4"),
		},
	}
)

// GenerateTagsForClip automatically generates tags for a clip
func (s *AutoTagService) GenerateTagsForClip(ctx context.Context, clip *models.Clip) ([]string, error) {
	candidates := canonicalTagCandidates(clip)
	seen := map[string]bool{}
	result := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.slug == "" || seen[candidate.slug] {
			continue
		}
		seen[candidate.slug] = true
		blacklisted, err := s.tagRepo.IsBlacklisted(ctx, candidate.slug)
		if err != nil {
			return nil, fmt.Errorf("checking blacklist for %q: %w", candidate.slug, err)
		}
		if blacklisted {
			continue
		}
		if _, err := s.tagRepo.GetOrCreateTag(ctx, canonicalTagName(candidate.parent), candidate.parent, nil); err != nil {
			return nil, fmt.Errorf("ensuring tag parent %q: %w", candidate.parent, err)
		}
		parent := candidate.parent
		if _, err := s.tagRepo.GetOrCreateTagWithParent(ctx, candidate.name, candidate.slug, &parent, candidate.color); err != nil {
			return nil, fmt.Errorf("ensuring canonical tag %q: %w", candidate.slug, err)
		}
		result = append(result, candidate.slug)
	}
	return result, nil
}

func canonicalTagCandidates(clip *models.Clip) []canonicalTagCandidate {
	candidates := make([]canonicalTagCandidate, 0, 12)
	for _, pattern := range tagPatterns {
		if pattern.Pattern.MatchString(clip.Title) {
			candidates = append(candidates, canonicalTagCandidate{"Content: " + pattern.TagName, "content/" + pattern.TagSlug, "content", pattern.Color})
		}
	}
	if clip.GameID != nil {
		for _, slug := range GameToGenres[*clip.GameID] {
			candidates = append(candidates, canonicalTagCandidate{"Game: " + canonicalTagName(slug), slug, "game", stringPtr("#4169E1")})
		}
	}
	if clip.GameName != nil && *clip.GameName != "" && (clip.GameID == nil || len(GameToGenres[*clip.GameID]) == 0) {
		gameSlug := slugify(*clip.GameName)
		if len(gameSlug) > 45 {
			gameSlug = strings.TrimRight(gameSlug[:45], "-")
		}
		if gameSlug != "" {
			gameTagName := "Game: " + *clip.GameName
			if len([]rune(gameTagName)) > 50 {
				gameTagName = "Game:" + gameSlug
			}
			candidates = append(candidates, canonicalTagCandidate{gameTagName, "game/" + gameSlug, "game", stringPtr("#4169E1")})
		}
	}
	duration := 0.0
	if clip.Duration != nil {
		duration = *clip.Duration
	}
	candidates = append(candidates, canonicalTagCandidate{canonicalTagName(durationTag(duration)), durationTag(duration), "duration", nil})
	if clip.Language != nil && *clip.Language != "" {
		slug := "lang/" + normalizeLanguage(*clip.Language)
		candidates = append(candidates, canonicalTagCandidate{"Language: " + canonicalTagName(slug), slug, "lang", nil})
	}
	return candidates
}

// CanonicalTagSlugsForClip previews deterministic structural tags without
// touching the database. The backfill dry-run uses it to report real work.
func CanonicalTagSlugsForClip(clip *models.Clip) []string {
	seen := map[string]bool{}
	result := make([]string, 0, 12)
	for _, candidate := range canonicalTagCandidates(clip) {
		if candidate.slug == "" || seen[candidate.slug] {
			continue
		}
		seen[candidate.slug] = true
		result = append(result, candidate.slug)
	}
	return result
}

func canonicalTagName(slug string) string {
	name := slug
	if idx := strings.IndexByte(name, '/'); idx >= 0 {
		name = name[idx+1:]
	}
	return strings.Title(strings.ReplaceAll(name, "-", " "))
}

// ApplyAutoTags generates and applies tags to a clip
func (s *AutoTagService) ApplyAutoTags(ctx context.Context, clip *models.Clip) error {
	// Generate tag slugs
	tagSlugs, err := s.GenerateTagsForClip(ctx, clip)
	if err != nil {
		return fmt.Errorf("failed to generate tags: %w", err)
	}

	// Apply each tag to the clip
	for _, slug := range tagSlugs {
		tag, err := s.tagRepo.GetBySlug(ctx, slug)
		if err != nil {
			return fmt.Errorf("loading generated tag %q: %w", slug, err)
		}

		// Add tag to clip
		err = s.tagRepo.AddTagToClip(ctx, clip.ID, tag.ID)
		if err != nil {
			return fmt.Errorf("attaching generated tag %q: %w", slug, err)
		}
	}

	return nil
}

func (s *AutoTagService) TagClip(ctx context.Context, clip *models.Clip) ([]string, error) {
	slugs, err := s.GenerateTagsForClip(ctx, clip)
	if err != nil {
		return nil, err
	}
	for _, slug := range slugs {
		tag, err := s.tagRepo.GetBySlug(ctx, slug)
		if err != nil {
			return nil, fmt.Errorf("loading generated tag %q: %w", slug, err)
		}
		if err := s.tagRepo.AddTagToClip(ctx, clip.ID, tag.ID); err != nil {
			return nil, fmt.Errorf("attaching generated tag %q: %w", slug, err)
		}
	}
	return slugs, nil
}

// AttachContentTags canonicalizes vision/classification labels under content/.
func (s *AutoTagService) AttachContentTags(ctx context.Context, clipID uuid.UUID, labels []string) error {
	if _, err := s.tagRepo.GetOrCreateTag(ctx, "Content", "content", nil); err != nil {
		return fmt.Errorf("ensuring content root: %w", err)
	}
	for _, label := range labels {
		slug := utils.Slugify(strings.TrimPrefix(label, "content/"))
		if slug == "" {
			continue
		}
		if len(slug) > 41 {
			slug = strings.TrimRight(slug[:41], "-")
		}
		fullSlug := "content/" + slug
		blacklisted, err := s.tagRepo.IsBlacklisted(ctx, fullSlug)
		if err != nil {
			return fmt.Errorf("checking content tag blacklist: %w", err)
		}
		if blacklisted {
			continue
		}
		parent := "content"
		tag, err := s.tagRepo.GetOrCreateTagWithParent(ctx, "Content: "+slug, fullSlug, &parent, nil)
		if err != nil {
			return fmt.Errorf("ensuring content tag %q: %w", fullSlug, err)
		}
		if err := s.tagRepo.AddTagToClip(ctx, clipID, tag.ID); err != nil {
			return fmt.Errorf("attaching content tag %q: %w", fullSlug, err)
		}
	}
	return nil
}

// AttachStreamerTags stores a broadcaster's own Twitch channel tags under
// streamer/. They describe the channel, not the clip, so they stay out of the
// content/ vocabulary that the vision tagger and title rules maintain.
func (s *AutoTagService) AttachStreamerTags(ctx context.Context, clipID uuid.UUID, labels []string) error {
	if _, err := s.tagRepo.GetOrCreateTag(ctx, "Streamer", tagtaxonomy.RootStreamer, nil); err != nil {
		return fmt.Errorf("ensuring streamer root: %w", err)
	}
	parent := tagtaxonomy.RootStreamer
	for _, label := range labels {
		fullSlug, name, ok := streamerTagIdentity(label)
		if !ok {
			continue
		}
		blacklisted, err := s.tagRepo.IsBlacklisted(ctx, fullSlug)
		if err != nil {
			return fmt.Errorf("checking streamer tag blacklist: %w", err)
		}
		if blacklisted {
			continue
		}
		tag, err := s.tagRepo.GetOrCreateTagWithParent(ctx, name, fullSlug, &parent, nil)
		if err != nil {
			return fmt.Errorf("ensuring streamer tag %q: %w", fullSlug, err)
		}
		if err := s.tagRepo.AddTagToClip(ctx, clipID, tag.ID); err != nil {
			return fmt.Errorf("attaching streamer tag %q: %w", fullSlug, err)
		}
	}
	return nil
}

// streamerTagIdentity maps a Twitch channel tag such as "CommunityOriented" to
// its stored slug and name: streamer/community-oriented, "Streamer: Community
// Oriented".
func streamerTagIdentity(label string) (slug, name string, ok bool) {
	readable := tagtaxonomy.SplitLabel(label)
	tail := utils.Slugify(readable)
	if len(tail) > 40 {
		tail = strings.TrimRight(tail[:40], "-")
	}
	if tail == "" {
		return "", "", false
	}
	return tagtaxonomy.RootStreamer + "/" + tail, "Streamer: " + readable, true
}

// slugify converts a string to a URL-friendly slug for tag creation.
// Delegates to utils.Slugify and truncates to 50 chars for tag DB storage.
func slugify(s string) string {
	s = utils.Slugify(s)
	if len(s) > 50 {
		s = strings.TrimRight(s[:50], "-")
	}
	return s
}

// getLanguageTag converts language code to tag slug
func getLanguageTag(langCode string) string {
	langMap := map[string]string{
		"en": "english",
		"es": "spanish",
		"fr": "french",
		"de": "german",
		"it": "italian",
		"pt": "portuguese",
		"ru": "russian",
		"ja": "japanese",
		"ko": "korean",
		"zh": "chinese",
		"ar": "arabic",
		"hi": "hindi",
		"tr": "turkish",
		"pl": "polish",
		"nl": "dutch",
		"sv": "swedish",
		"no": "norwegian",
		"fi": "finnish",
		"da": "danish",
	}

	if tag, ok := langMap[langCode]; ok {
		return tag
	}
	return ""
}

// getLanguageName converts language code to full name
func getLanguageName(langCode string) string {
	nameMap := map[string]string{
		"en": "English",
		"es": "Spanish",
		"fr": "French",
		"de": "German",
		"it": "Italian",
		"pt": "Portuguese",
		"ru": "Russian",
		"ja": "Japanese",
		"ko": "Korean",
		"zh": "Chinese",
		"ar": "Arabic",
		"hi": "Hindi",
		"tr": "Turkish",
		"pl": "Polish",
		"nl": "Dutch",
		"sv": "Swedish",
		"no": "Norwegian",
		"fi": "Finnish",
		"da": "Danish",
	}

	if name, ok := nameMap[langCode]; ok {
		return name
	}
	return langCode
}
