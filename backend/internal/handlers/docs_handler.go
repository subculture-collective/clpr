package handlers

import (
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
	maxDocumentBytes    = 1 << 20
	maxDocSearchResults = 100
)

// DocsHandler serves the Markdown documentation tree. docs is nil when no
// documentation source is available.
type DocsHandler struct {
	docs         fs.FS
	githubOwner  string
	githubRepo   string
	githubBranch string
}

type SearchResult struct {
	Path    string   `json:"path"`
	Name    string   `json:"name"`
	Matches []string `json:"matches"`
	Score   int      `json:"score"`
}

// NewDocsHandler serves documentation from docsPath. Symlinks cannot escape
// the directory. A missing directory leaves the handler without a source.
func NewDocsHandler(docsPath, githubOwner, githubRepo, githubBranch string) *DocsHandler {
	var docs fs.FS
	if root, err := os.OpenRoot(docsPath); err == nil {
		docs = root.FS()
	} else {
		utils.GetLogger().Warn("Documentation directory is unavailable", map[string]interface{}{"path": docsPath, "error": err.Error()})
	}
	return NewDocsHandlerFS(docs, githubOwner, githubRepo, githubBranch)
}

// NewDocsHandlerFS serves documentation from docs, such as the copy embedded
// in the binary by the docscontent package.
func NewDocsHandlerFS(docs fs.FS, githubOwner, githubRepo, githubBranch string) *DocsHandler {
	return &DocsHandler{
		docs:         docs,
		githubOwner:  githubOwner,
		githubRepo:   githubRepo,
		githubBranch: githubBranch,
	}
}

// GetDocsList returns the list of all available documentation files
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
		tree, err := h.buildDocsTree(".")
		if err != nil {
			// An unreadable tree is reported as empty so the docs page renders.
			utils.GetLogger().Error("Failed to list documentation", err)
		} else if tree != nil {
			docs = tree
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"docs": docs,
	})
}

// GetDoc returns the content of a specific documentation file
// GET /api/v1/docs/:path
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

	if h.docs == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Documentation is unavailable"})
		return
	}

	// Reject symlinks anywhere on the path, matching the listing, which skips them.
	var info fs.FileInfo
	segments := strings.Split(docPath, "/")
	for i := range segments {
		var err error
		info, err = fs.Lstat(h.docs, strings.Join(segments[:i+1], "/"))
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
			} else {
				utils.GetLogger().Error("Failed to resolve document", err, map[string]interface{}{"path": docPath})
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve document"})
			}
			return
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}
	if !info.Mode().IsRegular() {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}
	if info.Size() > maxDocumentBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Document exceeds 1 MiB"})
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
		"content":    string(content),
		"github_url": githubURL,
	})
}

// SearchDocs performs full-text search across documentation files
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
	results := h.searchDocuments(".", query)
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

func (h *DocsHandler) searchDocuments(dir, query string) []SearchResult {
	entries, err := fs.ReadDir(h.docs, dir)
	if err != nil {
		return nil
	}

	var results []SearchResult

	for _, entry := range entries {
		// Skip hidden files, archive, and vault
		if strings.HasPrefix(entry.Name(), ".") || entry.Name() == "archive" || entry.Name() == "vault" || entry.Type()&fs.ModeSymlink != 0 {
			continue
		}

		name := entry.Name()
		docPath := path.Join(dir, name)

		if entry.IsDir() {
			// Recursively search subdirectories
			subResults := h.searchDocuments(docPath, query)
			results = append(results, subResults...)
		} else if strings.HasSuffix(name, ".md") {
			// Search in markdown file
			info, err := entry.Info()
			if err != nil || info.Size() > maxDocumentBytes {
				continue
			}
			content, err := fs.ReadFile(h.docs, docPath)
			if err != nil {
				continue
			}

			contentLower := strings.ToLower(string(content))
			lines := strings.Split(string(content), "\n")

			// Check if query matches
			if strings.Contains(contentLower, query) {
				matches := []string{}
				score := 0

				// Find matching lines (up to 3)
				for _, line := range lines {
					if len(matches) >= 3 {
						break
					}
					if strings.Contains(strings.ToLower(line), query) {
						// Trim and add context
						trimmed := strings.TrimSpace(line)
						if len(trimmed) > 100 {
							// Find query position and add context
							idx := strings.Index(strings.ToLower(trimmed), query)
							start := max(0, idx-40)
							end := min(len(trimmed), idx+len(query)+40)
							trimmed = "..." + trimmed[start:end] + "..."
						}
						matches = append(matches, trimmed)
						score++
					}
				}

				// Boost score for title matches
				if strings.Contains(strings.ToLower(name), query) {
					score += 5
				}

				results = append(results, SearchResult{
					Path:    strings.TrimSuffix(docPath, ".md"),
					Name:    strings.TrimSuffix(name, ".md"),
					Matches: matches,
					Score:   score,
				})
			}
		}
	}

	return results
}

func (h *DocsHandler) generateGitHubURL(docPath string) string {
	return fmt.Sprintf("https://github.com/%s/%s/edit/%s/docs/%s", h.githubOwner, h.githubRepo, h.githubBranch, docPath)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type DocNode struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	Type     string     `json:"type"` // "file" or "directory"
	Children []*DocNode `json:"children,omitempty"`
}

func (h *DocsHandler) buildDocsTree(dir string) ([]*DocNode, error) {
	entries, err := fs.ReadDir(h.docs, dir)
	if err != nil {
		return nil, err
	}

	var nodes []*DocNode

	for _, entry := range entries {
		// Skip hidden files and directories
		if strings.HasPrefix(entry.Name(), ".") || entry.Type()&fs.ModeSymlink != 0 {
			continue
		}

		// Skip archive and vault directories
		if entry.Name() == "archive" || entry.Name() == "vault" {
			continue
		}

		name := entry.Name()
		docPath := path.Join(dir, name)

		if entry.IsDir() {
			// Recursively build tree for subdirectories
			children, err := h.buildDocsTree(docPath)
			if err != nil {
				continue
			}

			nodes = append(nodes, &DocNode{
				Name:     name,
				Path:     docPath,
				Type:     "directory",
				Children: children,
			})
		} else if strings.HasSuffix(name, ".md") {
			// Add markdown files
			nodes = append(nodes, &DocNode{
				Name: strings.TrimSuffix(name, ".md"),
				Path: strings.TrimSuffix(docPath, ".md"),
				Type: "file",
			})
		}
	}

	return nodes, nil
}
