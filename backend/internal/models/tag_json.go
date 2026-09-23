package models

import (
	"encoding/json"

	"git.subcult.tv/subculture-collective/clpr/internal/tagtaxonomy"
)

// MarshalJSON adds the derived lane, display_name, and evidence fields so
// every endpoint that returns a tag describes where it came from.
func (t Tag) MarshalJSON() ([]byte, error) {
	type storedTag Tag
	labels := tagtaxonomy.Classify(t.Slug, t.Name)
	return json.Marshal(struct {
		storedTag
		Lane        tagtaxonomy.Lane     `json:"lane"`
		DisplayName string               `json:"display_name"`
		Evidence    tagtaxonomy.Evidence `json:"evidence,omitempty"`
	}{storedTag(t), labels.Lane, labels.DisplayName, labels.Evidence})
}
