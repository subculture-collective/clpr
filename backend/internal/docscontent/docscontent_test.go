package docscontent

import (
	"io/fs"
	"path/filepath"
	"testing"
)

func TestEmbeddedIncludesDocumentationIndex(t *testing.T) {
	content, err := fs.ReadFile(Embedded(), "index.md")
	if err != nil {
		t.Fatalf("embedded index.md: %v", err)
	}
	if len(content) == 0 {
		t.Fatal("embedded index.md is empty")
	}
}

func TestOpenFallsBackToEmbedded(t *testing.T) {
	if _, source := Open(filepath.Join(t.TempDir(), "missing")); source != "embedded" {
		t.Fatalf("source = %q, want embedded", source)
	}
	if _, source := Open(t.TempDir()); source != "directory" {
		t.Fatalf("source = %q, want directory", source)
	}
}
