package services

import (
	"fmt"
	"strings"

	"git.subcult.tv/subculture-collective/clpr/internal/tagtaxonomy"
)

type contentTagEvidence = tagtaxonomy.Evidence

const (
	visibleTag    = tagtaxonomy.Visible
	contextualTag = tagtaxonomy.Contextual
	strongTag     = tagtaxonomy.Strong
)

// contentTagCatalog is the model-facing content vocabulary. It lives in
// tagtaxonomy so API responses can label content tags from the same source.
var contentTagCatalog = tagtaxonomy.ContentCatalog

var contentTagSlugs = func() []string {
	slugs := make([]string, 0, len(contentTagCatalog))
	for _, tag := range contentTagCatalog {
		slugs = append(slugs, tag.Slug)
	}
	return slugs
}()

func contentTagPromptCatalog() string {
	groups := []struct {
		evidence contentTagEvidence
		heading  string
	}{
		{visibleTag, "VISIBLE: may be selected from direct thumbnail evidence"},
		{contextualTag, "CONTEXTUAL: requires explicit metadata or transcript plus compatible visual evidence"},
		{strongTag, "STRONG: requires explicit outcome evidence; title wording or one ambiguous frame is not enough"},
	}
	var builder strings.Builder
	for groupIndex, group := range groups {
		if groupIndex > 0 {
			builder.WriteByte('\n')
		}
		builder.WriteString(group.heading)
		builder.WriteString(":\n")
		for _, tag := range contentTagCatalog {
			if tag.Evidence != group.evidence {
				continue
			}
			fmt.Fprintf(&builder, "- %s: %s\n", tag.Slug, tag.Description)
		}
	}
	return strings.TrimSpace(builder.String())
}
