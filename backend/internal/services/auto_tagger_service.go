package services

import (
	"context"
	"fmt"
	"log"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
)

// AutoTaggerService assigns structural tags (duration, language) to clips
// based on clip metadata. These tags are deterministic and free — no AI needed.
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