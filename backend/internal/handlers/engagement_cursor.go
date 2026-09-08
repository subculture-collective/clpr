package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"math"
	"net/url"
)

type engagementCursor struct {
	Generation uuid.UUID `json:"generation"`
	ClipID     uuid.UUID `json:"clip_id"`
	Score      float64   `json:"score"`
	Filters    string    `json:"filters"`
}

func engagementFilterKey(values url.Values) string {
	canonical := url.Values{}
	for k, v := range values {
		if k != "cursor" && k != "offset" && k != "limit" {
			canonical[k] = v
		}
	}
	if canonical.Get("sort") == "" {
		canonical.Set("sort", "trending")
	}
	if canonical.Get("timeframe") == "" {
		canonical.Set("timeframe", "day")
	}
	sum := sha256.Sum256([]byte(canonical.Encode()))
	return hex.EncodeToString(sum[:])
}
func encodeEngagementCursor(generation, clipID uuid.UUID, score float64, values url.Values) string {
	data, _ := json.Marshal(engagementCursor{generation, clipID, score, engagementFilterKey(values)})
	return base64.RawURLEncoding.EncodeToString(data)
}
func decodeEngagementCursor(raw string, values url.Values) (engagementCursor, error) {
	var cursor engagementCursor
	data, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return cursor, err
	}
	if err = json.Unmarshal(data, &cursor); err != nil {
		return cursor, err
	}
	if cursor.Generation == uuid.Nil || cursor.ClipID == uuid.Nil || cursor.Score <= 0 || math.IsNaN(cursor.Score) || math.IsInf(cursor.Score, 0) || cursor.Filters != engagementFilterKey(values) {
		return cursor, fmt.Errorf("invalid ranking cursor")
	}
	return cursor, nil
}
