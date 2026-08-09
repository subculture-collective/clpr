package handlers

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	maxDocumentBytes    = 1 << 20
	maxDocSearchResults = 100
)

type DocsHandler struct {
	docsPath     string
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

func NewDocsHandler(docsPath, githubOwner, githubRepo, githubBranch string) *DocsHandler {
	return &DocsHandler{
		docsPath:     docsPath,
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
	docs, err := h.buildDocsTree(h.docsPath, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list documentation"})
		return
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

	rootPath, err := filepath.EvalSymlinks(h.docsPath)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Documentation is unavailable"})
		return
	}
	filePath, err := filepath.EvalSymlinks(filepath.Join(rootPath, filepath.FromSlash(docPath)))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve document"})
		}
		return
	}
	relative, err := filepath.Rel(rootPath, filePath)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}
	info, err := os.Stat(filePath)
	if err != nil || !info.Mode().IsRegular() {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}
	if info.Size() > maxDocumentBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Document exceeds 1 MiB"})
		return
	}
	content, err := os.ReadFile(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read document"})
		return
	}

	// Generate GitHub edit URL
	githubURL := ""
	if h.githubOwner != "" && h.githubRepo != "" {
		// Convert relative path to GitHub URL
		cleanPath := strings.TrimPrefix(docPath, "/")
		if !strings.HasSuffix(cleanPath, ".md") {
			cleanPath += ".md"
		}
		githubURL = h.generateGitHubURL(cleanPath)
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
	if _, err := os.Stat(h.docsPath); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Documentation is unavailable"})
		return
	}

	query = strings.ToLower(query)
	results := h.searchDocuments(h.docsPath, "", query)
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

func (h *DocsHandler) searchDocuments(basePath, relativePath, query string) []SearchResult {
	currentPath := filepath.Join(basePath, relativePath)
	entries, err := os.ReadDir(currentPath)
	if err != nil {
		return nil
	}

	var results []SearchResult

	for _, entry := range entries {
		// Skip hidden files, archive, and vault
		if strings.HasPrefix(entry.Name(), ".") || entry.Name() == "archive" || entry.Name() == "vault" || entry.Type()&os.ModeSymlink != 0 {
			continue
		}

		name := entry.Name()
		path := filepath.Join(relativePath, name)

		if entry.IsDir() {
			// Recursively search subdirectories
			subResults := h.searchDocuments(basePath, path, query)
			results = append(results, subResults...)
		} else if strings.HasSuffix(name, ".md") {
			// Search in markdown file
			fullPath := filepath.Join(basePath, path)
			info, err := entry.Info()
			if err != nil || info.Size() > maxDocumentBytes {
				continue
			}
			content, err := os.ReadFile(fullPath)
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
					Path:    strings.TrimSuffix(path, ".md"),
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
	return fmt.Sprintf("https://github.com/%s/%s/edit/%s/docs/%s", h.githubOwner, h.githubRepo, h.githubBranch, filepath.ToSlash(docPath))
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

func (h *DocsHandler) buildDocsTree(basePath, relativePath string) ([]*DocNode, error) {
	currentPath := filepath.Join(basePath, relativePath)
	entries, err := os.ReadDir(currentPath)
	if err != nil {
		return nil, err
	}

	var nodes []*DocNode

	for _, entry := range entries {
		// Skip hidden files and directories
		if strings.HasPrefix(entry.Name(), ".") || entry.Type()&os.ModeSymlink != 0 {
			continue
		}

		// Skip archive and vault directories
		if entry.Name() == "archive" || entry.Name() == "vault" {
			continue
		}

		name := entry.Name()
		path := filepath.Join(relativePath, name)

		if entry.IsDir() {
			// Recursively build tree for subdirectories
			children, err := h.buildDocsTree(basePath, path)
			if err != nil {
				continue
			}

			nodes = append(nodes, &DocNode{
				Name:     name,
				Path:     path,
				Type:     "directory",
				Children: children,
			})
		} else if strings.HasSuffix(name, ".md") {
			// Add markdown files
			nodes = append(nodes, &DocNode{
				Name: strings.TrimSuffix(name, ".md"),
				Path: strings.TrimSuffix(path, ".md"),
				Type: "file",
			})
		}
	}

	return nodes, nil
}
