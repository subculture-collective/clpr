package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"unicode"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type TopicClassificationService struct {
	repo *repository.ClipTopicRepository
}

func NewTopicClassificationService(repo *repository.ClipTopicRepository) *TopicClassificationService {
	return &TopicClassificationService{repo: repo}
}

var topicKeywords = map[string][]string{
	"news-politics":        {"election", "president", "congress", "government", "policy", "breaking news", "politics"},
	"irl-travel":           {"travel", "airport", "street", "restaurant", "irl", "city"},
	"reactions-commentary": {"react", "reaction", "commentary", "responds", "take"},
	"music-performance":    {"song", "music", "concert", "singing", "performance"},
	"sports":               {"football", "basketball", "baseball", "soccer", "match", "sports"},
	"creative-making":      {"art", "drawing", "painting", "cooking", "build", "making"},
	"tech":                 {"technology", "software", "coding", "computer", "ai", "hardware"},
	"culture-drama":        {"drama", "controversy", "culture", "viral"},
	"gaming":               {"game", "gaming", "speedrun", "boss", "ranked"},
}

func (s *TopicClassificationService) ClassifyClip(ctx context.Context, clipID uuid.UUID) error {
	input, err := s.repo.GetClassificationInput(ctx, clipID)
	if err != nil {
		return err
	}
	candidates := map[uuid.UUID]models.TopicCandidate{}
	for _, topic := range input.MappedTopics {
		evidence, _ := json.Marshal(map[string]string{"twitch_category_id": input.TwitchCategoryID})
		candidates[topic.ID] = models.TopicCandidate{TopicID: topic.ID, Source: "twitch_category", Confidence: .55, Evidence: evidence}
	}
	title := strings.ToLower(input.Title)
	tags := strings.ToLower(strings.Join(input.Tags, " "))
	transcript := strings.ToLower(input.Transcript)
	for slug, keywords := range topicKeywords {
		confidence, source, matched := 0.0, "metadata", ""
		for _, keyword := range keywords {
			if containsTopicKeyword(transcript, keyword) {
				confidence, source, matched = .82, "transcript", keyword
				break
			}
			if containsTopicKeyword(tags, keyword) && confidence < .76 {
				confidence, source, matched = .76, "tag", keyword
			}
			if containsTopicKeyword(title, keyword) && confidence < .72 {
				confidence, source, matched = .72, "metadata", keyword
			}
		}
		if confidence == 0 {
			continue
		}
		// Politics is a sensitive topic. Only assign it from the higher-confidence
		// transcript signal; Twitch category mappings remain independently visible.
		if slug == "news-politics" && source != "transcript" {
			continue
		}
		topicID, findErr := s.repo.FindActiveTopicIDBySlug(ctx, slug)
		if findErr != nil {
			if errors.Is(findErr, pgx.ErrNoRows) {
				continue
			}
			return findErr
		}
		evidence, _ := json.Marshal(map[string]string{"matched": matched})
		if existing, ok := candidates[topicID]; !ok || confidence > existing.Confidence {
			candidates[topicID] = models.TopicCandidate{TopicID: topicID, Source: source, Confidence: confidence, Evidence: evidence}
		}
	}
	list := make([]models.TopicCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		list = append(list, candidate)
	}
	return s.repo.UpsertCandidates(ctx, clipID, list)
}

func containsTopicKeyword(text, keyword string) bool {
	if strings.ContainsRune(keyword, ' ') {
		return strings.Contains(text, keyword)
	}
	for _, token := range strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	}) {
		if token == keyword {
			return true
		}
	}
	return false
}

func (s *TopicClassificationService) Backfill(ctx context.Context, limit int) (int, error) {
	ids, err := s.repo.ListUnclassifiedClipIDs(ctx, limit)
	if err != nil {
		return 0, err
	}
	processed := 0
	for _, id := range ids {
		if err := s.ClassifyClip(ctx, id); err != nil {
			return processed, err
		}
		processed++
	}
	return processed, nil
}
