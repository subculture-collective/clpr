package search

import (
	"net/url"
	"strconv"
	"strings"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// bindSearchQuery fills req from URL query values the way gin's
// c.ShouldBindQuery fills models.SearchRequest from its `form` tags:
//
//   - a missing key leaves the field at its zero value;
//   - a present key sets the field from its first value, and sets a pointer
//     field even when that value is empty;
//   - strings are used as given;
//   - integers are trimmed, an empty value means 0, and a value that does not
//     parse is an error;
//   - tags keeps every value in order.
//
// SearchRequest has no `binding` rules, so gin's validation step never
// rejected a request and there is nothing to validate here.
//
// The handler uses this instead of c.ShouldBindQuery because gin's binding
// package links protobuf, YAML, TOML, BSON and a struct validator into every
// binary that calls it, about 3 MB gzip in the js/wasm build of this package.
// TestBindSearchQueryMatchesGin checks the two stay identical.
func bindSearchQuery(query url.Values, req *models.SearchRequest) error {
	setString(query, "q", &req.Query)
	setString(query, "type", &req.Type)
	setString(query, "sort", &req.Sort)
	setStringPtr(query, "game_id", &req.GameID)
	setStringPtr(query, "twitch_category_id", &req.TwitchCategoryID)
	setStringPtr(query, "creator_id", &req.CreatorID)
	setStringPtr(query, "language", &req.Language)
	if values, ok := query["tags"]; ok && len(values) > 0 {
		req.Tags = append([]string(nil), values...)
	}
	if err := setIntPtr(query, "min_votes", &req.MinVotes); err != nil {
		return err
	}
	setStringPtr(query, "date_from", &req.DateFrom)
	setStringPtr(query, "date_to", &req.DateTo)
	if err := setInt(query, "page", &req.Page); err != nil {
		return err
	}
	if err := setInt(query, "limit", &req.Limit); err != nil {
		return err
	}
	setStringPtr(query, "said", &req.Said)
	return nil
}

// firstValue returns the key's first value and whether the key is present.
func firstValue(query url.Values, key string) (string, bool) {
	values, ok := query[key]
	if !ok || len(values) == 0 {
		return "", ok
	}
	return values[0], true
}

func setString(query url.Values, key string, dst *string) {
	if value, ok := firstValue(query, key); ok {
		*dst = value
	}
}

func setStringPtr(query url.Values, key string, dst **string) {
	if value, ok := firstValue(query, key); ok {
		*dst = &value
	}
}

func setInt(query url.Values, key string, dst *int) error {
	value, ok := firstValue(query, key)
	if !ok {
		return nil
	}
	n, err := parseQueryInt(value)
	if err != nil {
		return err
	}
	*dst = n
	return nil
}

func setIntPtr(query url.Values, key string, dst **int) error {
	value, ok := firstValue(query, key)
	if !ok {
		return nil
	}
	n, err := parseQueryInt(value)
	if err != nil {
		return err
	}
	*dst = &n
	return nil
}

func parseQueryInt(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = "0"
	}
	return strconv.Atoi(value)
}
