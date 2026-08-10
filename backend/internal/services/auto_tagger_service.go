package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"github.com/google/uuid"
)

// gameToGenres maps Twitch game/category IDs to genre tag slugs.
// Each entry produces structural tags that link a clip to its game
// and genre categories. Slugs use the "game/..." prefix consistent
// with the seed tag hierarchy (parent_slug = "game").
var GameToGenres = map[string][]string{
	// Just Chatting / IRL
	"509658": {"game/just-chatting", "game/irl"},

	// FPS / Tactical Shooters
	"516575": {"game/valorant", "game/tactical-shooter", "game/fps"},
	"514790": {"game/counter-strike-2", "game/tactical-shooter", "game/fps"},
	"32399":  {"game/call-of-duty", "game/fps", "game/shooter"},
	"498566": {"game/call-of-duty-warzone", "game/battle-royale", "game/fps"},
	"493057": {"game/overwatch-2", "game/fps", "game/hero-shooter"},
	"488552": {"game/rainbow-six-siege", "game/tactical-shooter", "game/fps"},
	"491931": {"game/escape-from-tarkov", "game/fps", "game/survival"},

	// Battle Royale
	"33214":  {"game/fortnite", "game/battle-royale", "game/shooter"},
	"511224": {"game/apex-legends", "game/battle-royale", "game/shooter"},

	// MOBA
	"21779":  {"game/league-of-legends", "game/moba"},
	"29595":  {"game/dota-2", "game/moba"},

	// Sandbox / Survival / Open World
	"27471":  {"game/minecraft", "game/sandbox"},
	"32982":  {"game/grand-theft-auto-v", "game/open-world"},
	"512710": {"game/rust", "game/survival"},
	"489401": {"game/elden-ring", "game/action-rpg", "game/open-world"},
	"490100": {"game/red-dead-redemption-2", "game/open-world"},

	// MMO / RPG
	"506442": {"game/world-of-warcraft", "game/mmorpg"},
	"18122":  {"game/runescape", "game/mmorpg"},
	"513143": {"game/baldurs-gate-3", "game/rpg"},
	"490422": {"game/cyberpunk-2077", "game/rpg", "game/open-world"},

	// Sports / Racing
	"491487": {"game/ea-sports-fc", "game/sports"},
	"513181": {"game/nba-2k", "game/sports"},
	"518088": {"game/rocket-league", "game/sports", "game/racing"},

	// Strategy / Card / Auto-Battler
	"138585": {"game/hearthstone", "game/card-game"},
	"509670": {"game/teamfight-tactics", "game/auto-battler"},

	// Non-gaming categories
	"1469308723": {"game/software", "game/programming"},
	"26936":      {"game/music"},
	"509672":     {"game/art"},

	// Art / Creative / ASMR
	"509671": {"game/asmr"},
	"488191": {"game/creative"},
}

// AutoTaggerService assigns structural tags (duration, language, game) to
// clips based on clip metadata. These tags are deterministic and free — no
// AI needed.
type AutoTaggerService struct {
	tagRepo  *repository.TagRepository
	clipRepo *repository.ClipRepository
}

// NewAutoTaggerService creates a new AutoTaggerService.
func NewAutoTaggerService(tagRepo *repository.TagRepository, clipRepo *repository.ClipRepository) *AutoTaggerService {
	return &AutoTaggerService{tagRepo: tagRepo, clipRepo: clipRepo}
}

// TagClip assigns structural tags to a clip based on its metadata.
// Returns the list of tag slugs that were applied.
// Tagging is best-effort: failures for individual tags are logged but do not
// cause the method to return an error.
func (s *AutoTaggerService) TagClip(ctx context.Context, clip *models.Clip) ([]string, error) {
	var slugs []string

	// Game / genre tags from game_id
	if clip.GameID != nil && *clip.GameID != "" {
		if tags, ok := GameToGenres[*clip.GameID]; ok {
			slugs = append(slugs, tags...)
		} else if clip.GameName != nil && *clip.GameName != "" {
			// Auto-generate a game tag from the game name when not in the map
			slug := slugify(*clip.GameName)
			fullSlug := "game/" + slug
			if err := s.ensureTag(ctx, *clip.GameName, fullSlug, "game"); err != nil {
				log.Printf("failed to ensure game tag: game=%q slug=%q error=%v",
					*clip.GameName, fullSlug, err)
			} else {
				slugs = append(slugs, fullSlug)
			}
		}
	}

	// Duration tag — deterministic, always applies
	if clip.Duration != nil {
		slugs = append(slugs, durationTag(*clip.Duration))
	} else {
		slugs = append(slugs, durationTag(0))
	}

	// Language tag — only when we have a language code
	if clip.Language != nil && *clip.Language != "" {
		slugs = append(slugs, fmt.Sprintf("lang/%s", normalizeLanguage(*clip.Language)))
	}

	// Persist tags on the clip (best-effort)
	for _, slug := range slugs {
		if err := s.clipRepo.AddTagBySlug(ctx, clip.ID, slug); err != nil {
			log.Printf("failed to add structural tag: clip_id=%s tag=%s error=%v",
				clip.ID.String(), slug, err)
		}
	}

	return slugs, nil
}

// durationTag maps clip duration in seconds to a hierarchical tag slug.
//   - ≤30s  → "duration/short"
//   - ≤90s  → "duration/medium"
//   - >90s → "duration/long"
func durationTag(seconds float64) string {
	switch {
	case seconds <= 30:
		return "duration/short"
	case seconds <= 90:
		return "duration/medium"
	default:
		return "duration/long"
	}
}

// normalizeLanguage maps a Twitch language code to our lang slug suffix.
// Unrecognised codes fall back to "other".
func normalizeLanguage(lang string) string {
	langMap := map[string]string{
		"en": "en", "es": "es", "pt": "pt", "fr": "fr",
		"de": "de", "ru": "ru", "ja": "ja", "ko": "ko",
		"zh": "zh", "it": "it", "tr": "tr", "ar": "ar",
	}
	if mapped, ok := langMap[lang]; ok {
		return mapped
	}
	return "other"
}

// ensureTag ensures a tag with the given slug exists, creating it with
// parentSlug as needed. The parent tag is also created if it doesn't exist.
// This is idempotent — safe to call multiple times.
func (s *AutoTaggerService) ensureTag(ctx context.Context, name, slug, parentSlug string) error {
	// Ensure the parent tag exists first
	if _, err := s.tagRepo.GetOrCreateTag(ctx, parentSlug, parentSlug, nil); err != nil {
		return fmt.Errorf("ensuring parent tag %q: %w", parentSlug, err)
	}

	// Check if child tag already exists
	if _, err := s.tagRepo.GetBySlug(ctx, slug); err == nil {
		return nil
	}

	// Create the tag with parent_slug
	newTag := &models.Tag{
		ID:         uuid.New(),
		Name:       name,
		Slug:       slug,
		ParentSlug: &parentSlug,
		UsageCount: 0,
		CreatedAt:  time.Now(),
	}

	if err := s.tagRepo.Create(ctx, newTag); err != nil {
		// Race condition — tag may have been created between check and insert
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return nil
		}
		return fmt.Errorf("creating tag %q: %w", slug, err)
	}
	return nil
}