package handlers

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"

	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/gin-gonic/gin"
)

const (
	maxDocumentBytes    = 2 << 20 // the generated API reference is about 1.1 MB
	maxDocSearchResults = 100
)

var errDocumentForbidden = errors.New("document path contains a symlink")

// DocsHandler serves the public Markdown documentation. Only the documents in
// the allowlist (docs/public-docs.json) are listed, searched, or returned; any
// other path is reported as not found even when the file exists. docs is nil
// when no documentation source is available.
type DocsHandler struct {
	docs         fs.FS
	public       []string
	publicSet    map[string]bool
	githubOwner  string
	githubRepo   string
	githubBranch string
}

type SearchResult struct {
	Path    string   `json:"path"`
	Name    string   `json:"name"`
	Title   string   `json:"title"`
	Matches []string `json:"matches"`
	Score   int      `json:"score"`
}

// NewDocsHandler serves the allowlisted documents from docsPath. Symlinks
// cannot escape the directory. A missing directory leaves the handler without
// a source.
func NewDocsHandler(docsPath string, public []string, githubOwner, githubRepo, githubBranch string) *DocsHandler {
	var docs fs.FS
	if root, err := os.OpenRoot(docsPath); err == nil {
		docs = root.FS()
	} else {
		utils.GetLogger().Warn("Documentation directory is unavailable", map[string]interface{}{"path": docsPath, "error": err.Error()})
	}
	return NewDocsHandlerFS(docs, public, githubOwner, githubRepo, githubBranch)
}

// NewDocsHandlerFS serves the allowlisted documents from docs, such as the
// copy embedded in the binary by the docscontent package. public holds paths
// relative to docs ending in ".md"; an empty allowlist publishes nothing.
func NewDocsHandlerFS(docs fs.FS, public []string, githubOwner, githubRepo, githubBranch string) *DocsHandler {
	h := &DocsHandler{
		docs:         docs,
		publicSet:    make(map[string]bool, len(public)),
		githubOwner:  githubOwner,
		githubRepo:   githubRepo,
		githubBranch: githubBranch,
	}
	for _, document := range public {
		if fs.ValidPath(document) && path.Ext(document) == ".md" && !h.publicSet[document] {
			h.publicSet[document] = true
			h.public = append(h.public, document)
		}
	}
	sort.Strings(h.public)
	return h
}

// GetDocsList returns the tree of published documentation files
// GET /api/v1/docs
func (h *DocsHandler) GetDocsList(c *gin.Context) {
	accept := c.GetHeader("Accept")
	if accept != "" && !strings.Contains(accept, "application/json") && !strings.Contains(accept, "*/*") {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": "Only application/json is available"})
		return
	}
	docs := []*DocNode{}
	if h.docs == nil {
		utils.GetLogger().Warn("Documentation list requested but no documentation source is configured")
	} else {
		docs = h.buildDocsTree()
	}

	c.JSON(http.StatusOK, gin.H{
		"docs": docs,
	})
}

// GetDoc returns the content of a published documentation file
// GET /api/v1/docs/:path and /api/v1/docs/content/*path
func (h *DocsHandler) GetDoc(c *gin.Context) {
	docPath := strings.TrimPrefix(c.Param("path"), "/")

	// Security: prevent directory traversal
	if strings.Contains(docPath, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document path"})
		return
	}

	// Ensure .md extension
	if !strings.HasSuffix(docPath, ".md") {
		docPath += ".md"
	}
	if !fs.ValidPath(docPath) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document path"})
		return
	}

	// Fail closed: unpublished documents are indistinguishable from missing ones.
	if !h.publicSet[docPath] {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	if h.docs == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Documentation is unavailable"})
		return
	}

	info, err := h.statDocument(docPath)
	if err != nil {
		switch {
		case errors.Is(err, fs.ErrNotExist):
			c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		case errors.Is(err, errDocumentForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		default:
			utils.GetLogger().Error("Failed to resolve document", err, map[string]interface{}{"path": docPath})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve document"})
		}
		return
	}
	if info.Size() > maxDocumentBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Document exceeds 2 MiB"})
		return
	}
	content, err := fs.ReadFile(h.docs, docPath)
	if err != nil {
		utils.GetLogger().Error("Failed to read document", err, map[string]interface{}{"path": docPath})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read document"})
		return
	}

	// Generate GitHub edit URL
	githubURL := ""
	if h.githubOwner != "" && h.githubRepo != "" {
		githubURL = h.generateGitHubURL(docPath)
	}

	c.JSON(http.StatusOK, gin.H{
		"path":       docPath,
		"title":      DocumentTitle(content, docPath),
		"content":    string(content),
		"github_url": githubURL,
	})
}

// SearchDocs performs full-text search across published documentation files
// GET /api/v1/docs/search?q=query
func (h *DocsHandler) SearchDocs(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if len(query) < 2 || len(query) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query must be between 2 and 100 characters"})
		return
	}
	if h.docs == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Documentation is unavailable"})
		return
	}

	query = strings.ToLower(query)
	results := h.searchDocuments(query)
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].Path < results[j].Path
		}
		return results[i].Score > results[j].Score
	})
	if len(results) > maxDocSearchResults {
		results = results[:maxDocSearchResults]
	}

	c.JSON(http.StatusOK, gin.H{
		"query":   query,
		"results": results,
		"count":   len(results),
	})
}

// statDocument resolves a published document, rejecting symlinks anywhere on
// the path and anything that is not a regular file.
func (h *DocsHandler) statDocument(docPath string) (fs.FileInfo, error) {
	var info fs.FileInfo
	segments := strings.Split(docPath, "/")
	for i := range segments {
		var err error
		info, err = fs.Lstat(h.docs, strings.Join(segments[:i+1], "/"))
		if err != nil {
			return nil, err
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			return nil, errDocumentForbidden
		}
	}
	if !info.Mode().IsRegular() {
		return nil, fs.ErrNotExist
	}
	return info, nil
}

// readPublished returns the content of every published document that can be
// served, in path order. Missing, oversized, or unsafe entries are skipped.
func (h *DocsHandler) readPublished(visit func(docPath string, content []byte)) {
	for _, docPath := range h.public {
		info, err := h.statDocument(docPath)
		if err != nil || info.Size() > maxDocumentBytes {
			continue
		}
		content, err := fs.ReadFile(h.docs, docPath)
		if err != nil {
			continue
		}
		visit(docPath, content)
	}
}

func (h *DocsHandler) searchDocuments(query string) []SearchResult {
	var results []SearchResult
	h.readPublished(func(docPath string, content []byte) {
		if !strings.Contains(strings.ToLower(string(content)), query) {
			return
		}
		matches := []string{}
		score := 0
		// Find matching lines (up to 3)
		for _, line := range strings.Split(string(content), "\n") {
			if len(matches) >= 3 {
				break
			}
			if strings.Contains(strings.ToLower(line), query) {
				// Trim and add context
				trimmed := strings.TrimSpace(line)
				if len(trimmed) > 100 {
					idx := strings.Index(strings.ToLower(trimmed), query)
					start := max(0, idx-40)
					end := min(len(trimmed), idx+len(query)+40)
					trimmed = "..." + trimmed[start:end] + "..."
				}
				matches = append(matches, trimmed)
				score++
			}
		}

		name := strings.TrimSuffix(path.Base(docPath), ".md")
		title := DocumentTitle(content, docPath)
		// Boost score for name and title matches
		if strings.Contains(strings.ToLower(name), query) || strings.Contains(strings.ToLower(title), query) {
			score += 5
		}

		results = append(results, SearchResult{
			Path:    strings.TrimSuffix(docPath, ".md"),
			Name:    name,
			Title:   title,
			Matches: matches,
			Score:   score,
		})
	})
	return results
}

func (h *DocsHandler) generateGitHubURL(docPath string) string {
	return fmt.Sprintf("https://github.com/%s/%s/edit/%s/docs/%s", h.githubOwner, h.githubRepo, h.githubBranch, docPath)
}

type DocNode struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	Type     string     `json:"type"` // "file" or "directory"
	Title    string     `json:"title,omitempty"`
	Children []*DocNode `json:"children,omitempty"`
}

// buildDocsTree groups the published documents by directory. Each file node
// carries the title of its own content, so documents that share a base name
// (README, index) in different directories remain distinguishable.
func (h *DocsHandler) buildDocsTree() []*DocNode {
	root := &DocNode{Type: "directory"}
	h.readPublished(func(docPath string, content []byte) {
		parent := root
		segments := strings.Split(docPath, "/")
		for i, segment := range segments[:len(segments)-1] {
			var next *DocNode
			for _, child := range parent.Children {
				if child.Type == "directory" && child.Name == segment {
					next = child
					break
				}
			}
			if next == nil {
				next = &DocNode{Name: segment, Path: strings.Join(segments[:i+1], "/"), Type: "directory"}
				parent.Children = append(parent.Children, next)
			}
			parent = next
		}
		parent.Children = append(parent.Children, &DocNode{
			Name:  strings.TrimSuffix(segments[len(segments)-1], ".md"),
			Path:  strings.TrimSuffix(docPath, ".md"),
			Type:  "file",
			Title: DocumentTitle(content, docPath),
		})
	})
	sortDocNodes(root.Children)
	if root.Children == nil {
		return []*DocNode{}
	}
	return root.Children
}

// sortDocNodes orders files before directories, each by name, at every level.
func sortDocNodes(nodes []*DocNode) {
	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].Type != nodes[j].Type {
			return nodes[i].Type == "file"
		}
		return nodes[i].Name < nodes[j].Name
	})
	for _, node := range nodes {
		sortDocNodes(node.Children)
	}
}

// DocumentTitle returns the title of a Markdown document: its first level-one
// heading outside front matter and code fences, then its front matter title,
// then a readable form of the file name.
func DocumentTitle(content []byte, docPath string) string {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	scanner.Buffer(make([]byte, 0, 64*1024), maxDocumentBytes)
	frontMatterTitle := ""
	inFrontMatter := false
	fence := ""
	for lineNumber := 0; scanner.Scan(); lineNumber++ {
		line := strings.TrimRight(scanner.Text(), "\r")
		trimmed := strings.TrimSpace(line)
		if lineNumber == 0 && trimmed == "---" {
			inFrontMatter = true
			continue
		}
		if inFrontMatter {
			if trimmed == "---" || trimmed == "..." {
				inFrontMatter = false
			} else if value, ok := strings.CutPrefix(line, "title:"); ok {
				frontMatterTitle = unquoteYAMLScalar(strings.TrimSpace(value))
			}
			continue
		}
		if fence != "" {
			if strings.HasPrefix(trimmed, fence) {
				fence = ""
			}
			continue
		}
		if strings.HasPrefix(trimmed, "```") {
			fence = "```"
			continue
		}
		if strings.HasPrefix(trimmed, "~~~") {
			fence = "~~~"
			continue
		}
		// ATX headings allow up to three spaces of indentation.
		if len(line)-len(strings.TrimLeft(line, " ")) > 3 {
			continue
		}
		if heading, ok := strings.CutPrefix(trimmed, "# "); ok {
			heading = strings.TrimSpace(strings.TrimRight(strings.TrimSpace(heading), "#"))
			if heading != "" {
				return heading
			}
		}
	}
	if frontMatterTitle != "" {
		return frontMatterTitle
	}
	name := strings.TrimSuffix(path.Base(docPath), path.Ext(docPath))
	return strings.TrimSpace(strings.NewReplacer("-", " ", "_", " ").Replace(name))
}

func unquoteYAMLScalar(value string) string {
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
			return strings.TrimSpace(value[1 : len(value)-1])
		}
	}
	return value
}
