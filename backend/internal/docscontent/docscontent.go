// Package docscontent embeds the repository's public Markdown documentation so
// the API can serve /api/v1/docs from images that do not ship the docs
// directory. Only documents listed in docs/public-docs.json are published.
// Regenerate content/ and public-docs.json with `npm run docs:embed` from the
// repository root.
package docscontent

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
)

//go:embed content
var embedded embed.FS

//go:embed public-docs.json
var publicManifest []byte

// Embedded returns the documentation tree compiled into the binary.
func Embedded() fs.FS {
	sub, err := fs.Sub(embedded, "content")
	if err != nil {
		panic(err) // content is a compile-time directory
	}
	return sub
}

// Open returns the documentation at dir when that directory exists, so local
// edits are served without a rebuild, and the embedded copy otherwise. The
// returned source is "directory" or "embedded". Either source is filtered by
// PublicDocuments before anything is served.
func Open(dir string) (fs.FS, string) {
	if dir != "" {
		if root, err := os.OpenRoot(dir); err == nil {
			return root.FS(), "directory"
		}
	}
	return Embedded(), "embedded"
}

// PublicDocuments returns the published document paths, relative to the
// documentation root and ending in ".md", in sorted order.
func PublicDocuments() ([]string, error) {
	return ParsePublicManifest(publicManifest)
}

// ParsePublicManifest validates a docs/public-docs.json document.
func ParsePublicManifest(data []byte) ([]string, error) {
	var manifest struct {
		SchemaVersion int      `json:"schema_version"`
		Documents     []string `json:"documents"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("public documentation manifest: %w", err)
	}
	if manifest.SchemaVersion != 1 {
		return nil, fmt.Errorf("public documentation manifest: unsupported schema_version %d", manifest.SchemaVersion)
	}
	seen := make(map[string]bool, len(manifest.Documents))
	for _, document := range manifest.Documents {
		if !fs.ValidPath(document) || document == "." || path.Ext(document) != ".md" || strings.HasPrefix(path.Base(document), ".") {
			return nil, fmt.Errorf("public documentation manifest: invalid document %q", document)
		}
		if seen[document] {
			return nil, fmt.Errorf("public documentation manifest: duplicate document %q", document)
		}
		seen[document] = true
	}
	documents := append([]string(nil), manifest.Documents...)
	sort.Strings(documents)
	return documents, nil
}
