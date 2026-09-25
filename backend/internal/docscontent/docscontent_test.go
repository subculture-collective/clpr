package docscontent

import (
	"io/fs"
	"path/filepath"
	"testing"
)

// The embedded tree must hold exactly the allowlisted documents: nothing
// internal may be compiled into the binary.
func TestEmbeddedMatchesPublicManifest(t *testing.T) {
	documents, err := PublicDocuments()
	if err != nil {
		t.Fatalf("PublicDocuments: %v", err)
	}
	if len(documents) == 0 {
		t.Fatal("public documentation manifest lists no documents")
	}
	want := make(map[string]bool, len(documents))
	for _, document := range documents {
		want[document] = true
		content, err := fs.ReadFile(Embedded(), document)
		if err != nil {
			t.Fatalf("embedded %s: %v", document, err)
		}
		if len(content) == 0 {
			t.Fatalf("embedded %s is empty", document)
		}
	}
	err = fs.WalkDir(Embedded(), ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && !want[name] {
			t.Errorf("embedded %s is not listed in docs/public-docs.json", name)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestParsePublicManifestRejectsUnsafeEntries(t *testing.T) {
	for _, manifest := range []string{
		`{"schema_version":2,"documents":[]}`,
		`{"schema_version":1,"documents":["../secret.md"]}`,
		`{"schema_version":1,"documents":["/etc/passwd.md"]}`,
		`{"schema_version":1,"documents":["notes.txt"]}`,
		`{"schema_version":1,"documents":["ops/.hidden.md"]}`,
		`{"schema_version":1,"documents":["a.md","a.md"]}`,
		`not json`,
	} {
		if _, err := ParsePublicManifest([]byte(manifest)); err == nil {
			t.Errorf("ParsePublicManifest(%s) succeeded", manifest)
		}
	}
	documents, err := ParsePublicManifest([]byte(`{"schema_version":1,"documents":["users/guide.md","index.md"]}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(documents) != 2 || documents[0] != "index.md" || documents[1] != "users/guide.md" {
		t.Fatalf("documents = %v", documents)
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
