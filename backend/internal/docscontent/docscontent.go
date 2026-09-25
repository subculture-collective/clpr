// Package docscontent embeds the repository's Markdown documentation so the
// API can serve /api/v1/docs from images that do not ship the docs directory.
// Regenerate content/ with `npm run docs:embed` from the repository root.
package docscontent

import (
	"embed"
	"io/fs"
	"os"
)

//go:embed content
var embedded embed.FS

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
// returned source is "directory" or "embedded".
func Open(dir string) (fs.FS, string) {
	if dir != "" {
		if root, err := os.OpenRoot(dir); err == nil {
			return root.FS(), "directory"
		}
	}
	return Embedded(), "embedded"
}
