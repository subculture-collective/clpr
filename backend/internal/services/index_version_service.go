package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"git.subcult.tv/subculture-collective/clpr/pkg/opensearch"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

// IndexVersionService handles versioned index management for zero-downtime rebuilds
type IndexVersionService struct {
	osClient *opensearch.Client
}

// IndexVersion represents a version of an index
type IndexVersion struct {
	Name      string    `json:"name"`
	Version   int       `json:"version"`
	Alias     string    `json:"alias"`
	CreatedAt time.Time `json:"created_at"`
	DocCount  int64     `json:"doc_count"`
	SizeBytes int64     `json:"size_bytes"`
	IsActive  bool      `json:"is_active"`
}

// IndexVersionInfo contains detailed information about index versions
type IndexVersionInfo struct {
	BaseIndex     string         `json:"base_index"`
	ActiveVersion *IndexVersion  `json:"active_version,omitempty"`
	AllVersions   []IndexVersion `json:"all_versions"`
	TotalVersions int            `json:"total_versions"`
	LatestVersion int            `json:"latest_version"`
}

// NewIndexVersionService creates a new IndexVersionService
func NewIndexVersionService(osClient *opensearch.Client) *IndexVersionService {
	return &IndexVersionService{
		osClient: osClient,
	}
}

// getVersionedIndexName returns the versioned index name (e.g., "clips_v1")
func getVersionedIndexName(baseIndex string, version int) string {
	return fmt.Sprintf("%s_v%d", baseIndex, version)
}

// parseVersionFromIndexName extracts the version number from an index name
func parseVersionFromIndexName(baseIndex, indexName string) (int, bool) {
	prefix := baseIndex + "_v"
	if !strings.HasPrefix(indexName, prefix) {
		return 0, false
	}
	var version int
	_, err := fmt.Sscanf(indexName, prefix+"%d", &version)
	return version, err == nil
}

// GetIndexVersionInfo returns information about all versions of an index
func (s *IndexVersionService) GetIndexVersionInfo(ctx context.Context, baseIndex string) (*IndexVersionInfo, error) {
	info := &IndexVersionInfo{
		BaseIndex:   baseIndex,
		AllVersions: []IndexVersion{},
	}

	// Get all indices matching the pattern
	pattern := baseIndex + "_v*"
	req := opensearchapi.CatIndicesRequest{
		Index:  []string{pattern},
		Format: "json",
	}

	res, err := req.Do(ctx, s.osClient.GetClient())
	if err != nil {
		return nil, fmt.Errorf("failed to list indices: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		// If 404, no versioned indices exist yet
		if res.StatusCode == 404 {
			return info, nil
		}
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("failed to list indices: %s - %s", res.Status(), string(body))
	}

	var indices []struct {
		Index      string `json:"index"`
		DocsCount  string `json:"docs.count"`
		StoreSize  string `json:"store.size"`
		StoreBytes string `json:"pri.store.size"`
	}

	if err := json.NewDecoder(res.Body).Decode(&indices); err != nil {
		return nil, fmt.Errorf("failed to parse indices response: %w", err)
	}

	// Get alias information
	aliasReq := opensearchapi.CatAliasesRequest{
		Name:   []string{baseIndex},
		Format: "json",
	}

	aliasRes, err := aliasReq.Do(ctx, s.osClient.GetClient())
	if err != nil {
		return nil, fmt.Errorf("failed to get aliases: %w", err)
	}
	defer aliasRes.Body.Close()

	activeIndex := ""
	if !aliasRes.IsError() {
		var aliases []struct {
			Alias string `json:"alias"`
			Index string `json:"index"`
		}
		if err := json.NewDecoder(aliasRes.Body).Decode(&aliases); err == nil {
			for _, alias := range aliases {
				if alias.Alias == baseIndex {
					activeIndex = alias.Index
					break
				}
			}
		}
	}

	// Parse index info
	for _, idx := range indices {
		version, ok := parseVersionFromIndexName(baseIndex, idx.Index)
		if !ok {
			continue
		}

		var docCount int64
		fmt.Sscanf(idx.DocsCount, "%d", &docCount)

		var sizeBytes int64
		fmt.Sscanf(idx.StoreBytes, "%d", &sizeBytes)

		indexVersion := IndexVersion{
			Name:      idx.Index,
			Version:   version,
			Alias:     baseIndex,
			DocCount:  docCount,
			SizeBytes: sizeBytes,
			IsActive:  idx.Index == activeIndex,
		}

		info.AllVersions = append(info.AllVersions, indexVersion)

		if version > info.LatestVersion {
			info.LatestVersion = version
		}
	}

	// Sort versions by version number (descending)
	sort.Slice(info.AllVersions, func(i, j int) bool {
		return info.AllVersions[i].Version > info.AllVersions[j].Version
	})

	// Set ActiveVersion after sorting to point to the correct slice element
	for i := range info.AllVersions {
		if info.AllVersions[i].IsActive {
			info.ActiveVersion = &info.AllVersions[i]
			break
		}
	}

	info.TotalVersions = len(info.AllVersions)

	return info, nil
}

// CreateVersionedIndex creates a new versioned index with mapping
func (s *IndexVersionService) CreateVersionedIndex(ctx context.Context, baseIndex string, version int, mapping string) error {
	indexName := getVersionedIndexName(baseIndex, version)

	// Check if index already exists
	existsReq := opensearchapi.IndicesExistsRequest{
		Index: []string{indexName},
	}

	existsRes, err := existsReq.Do(ctx, s.osClient.GetClient())
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}
	defer existsRes.Body.Close()

	if existsRes.StatusCode == 200 {
		return fmt.Errorf("index %s already exists", indexName)
	}

	// Create the index
	createReq := opensearchapi.IndicesCreateRequest{
		Index: indexName,
		Body:  strings.NewReader(mapping),
	}

	createRes, err := createReq.Do(ctx, s.osClient.GetClient())
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer createRes.Body.Close()

	if createRes.IsError() {
		body, _ := io.ReadAll(createRes.Body)
		return fmt.Errorf("failed to create index: %s - %s", createRes.Status(), string(body))
	}

	utils.Info("Created versioned index", map[string]interface{}{"index": indexName})
	return nil
}

// SwapAlias atomically swaps the alias from old index to new index
func (s *IndexVersionService) SwapAlias(ctx context.Context, baseIndex string, newVersion int) error {
	newIndexName := getVersionedIndexName(baseIndex, newVersion)

	// Verify the new index exists
	existsReq := opensearchapi.IndicesExistsRequest{
		Index: []string{newIndexName},
	}

	existsRes, err := existsReq.Do(ctx, s.osClient.GetClient())
	if err != nil {
		return fmt.Errorf("failed to check new index existence: %w", err)
	}
	defer existsRes.Body.Close()

	if existsRes.StatusCode != 200 {
		return fmt.Errorf("new index %s does not exist", newIndexName)
	}

	// Get current alias info to find old index
	info, err := s.GetIndexVersionInfo(ctx, baseIndex)
	if err != nil {
		return fmt.Errorf("failed to get index info: %w", err)
	}

	// Build alias swap actions
	actions := []map[string]interface{}{}

	// Remove alias from old index if exists
	if info.ActiveVersion != nil {
		actions = append(actions, map[string]interface{}{
			"remove": map[string]interface{}{
				"index": info.ActiveVersion.Name,
				"alias": baseIndex,
			},
		})
	}

	// Add alias to new index
	actions = append(actions, map[string]interface{}{
		"add": map[string]interface{}{
			"index": newIndexName,
			"alias": baseIndex,
		},
	})

	aliasBody := map[string]interface{}{
		"actions": actions,
	}

	bodyBytes, err := json.Marshal(aliasBody)
	if err != nil {
		return fmt.Errorf("failed to marshal alias body: %w", err)
	}

	aliasReq := opensearchapi.IndicesUpdateAliasesRequest{
		Body: bytes.NewReader(bodyBytes),
	}

	aliasRes, err := aliasReq.Do(ctx, s.osClient.GetClient())
	if err != nil {
		return fmt.Errorf("failed to update aliases: %w", err)
	}
	defer aliasRes.Body.Close()

	if aliasRes.IsError() {
		body, _ := io.ReadAll(aliasRes.Body)
		return fmt.Errorf("failed to update aliases: %s - %s", aliasRes.Status(), string(body))
	}

	utils.Info("Swapped alias", map[string]interface{}{
		"alias": baseIndex,
		"old_index": func() string {
			if info.ActiveVersion != nil {
				return info.ActiveVersion.Name
			}
			return "none"
		}(),
		"new_index": newIndexName,
	})

	return nil
}

// RollbackAlias rolls back the alias to a previous version
func (s *IndexVersionService) RollbackAlias(ctx context.Context, baseIndex string, targetVersion int) error {
	targetIndexName := getVersionedIndexName(baseIndex, targetVersion)

	// Verify the target index exists
	existsReq := opensearchapi.IndicesExistsRequest{
		Index: []string{targetIndexName},
	}

	existsRes, err := existsReq.Do(ctx, s.osClient.GetClient())
	if err != nil {
		return fmt.Errorf("failed to check target index existence: %w", err)
	}
	defer existsRes.Body.Close()

	if existsRes.StatusCode != 200 {
		return fmt.Errorf("target index %s does not exist", targetIndexName)
	}

	// Swap to target version
	return s.SwapAlias(ctx, baseIndex, targetVersion)
}

// DeleteOldVersions deletes old index versions, keeping the specified number of recent versions
func (s *IndexVersionService) DeleteOldVersions(ctx context.Context, baseIndex string, keepCount int) ([]string, error) {
	info, err := s.GetIndexVersionInfo(ctx, baseIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get index info: %w", err)
	}

	// Sort by version (already sorted descending)
	// Skip first keepCount versions and active version
	deletedIndices := []string{}

	versionsToDelete := []IndexVersion{}
	for i, version := range info.AllVersions {
		// Skip if within keepCount
		if i < keepCount {
			continue
		}
		// Skip if active
		if version.IsActive {
			continue
		}
		versionsToDelete = append(versionsToDelete, version)
	}

	// Delete old versions
	for _, version := range versionsToDelete {
		deleteReq := opensearchapi.IndicesDeleteRequest{
			Index: []string{version.Name},
		}

		deleteRes, err := deleteReq.Do(ctx, s.osClient.GetClient())
		if err != nil {
			utils.Warn("Failed to delete index", map[string]interface{}{"index": version.Name, "error": err})
			continue
		}

		// Read body and close immediately (not deferred) since we're in a loop
		if deleteRes.IsError() {
			body, _ := io.ReadAll(deleteRes.Body)
			deleteRes.Body.Close()
			utils.Warn("Failed to delete index", map[string]interface{}{"index": version.Name, "status": deleteRes.Status(), "body": string(body)})
			continue
		}
		deleteRes.Body.Close()

		deletedIndices = append(deletedIndices, version.Name)
		utils.Info("Deleted old index version", map[string]interface{}{"index": version.Name})
	}

	return deletedIndices, nil
}

// GetNextVersion returns the next version number for a base index
func (s *IndexVersionService) GetNextVersion(ctx context.Context, baseIndex string) (int, error) {
	info, err := s.GetIndexVersionInfo(ctx, baseIndex)
	if err != nil {
		return 0, err
	}

	return info.LatestVersion + 1, nil
}

// RefreshIndex forces a refresh of the index to make all operations visible
func (s *IndexVersionService) RefreshIndex(ctx context.Context, indexName string) error {
	req := opensearchapi.IndicesRefreshRequest{
		Index: []string{indexName},
	}

	res, err := req.Do(ctx, s.osClient.GetClient())
	if err != nil {
		return fmt.Errorf("failed to refresh index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to refresh index: %s - %s", res.Status(), string(body))
	}

	return nil
}

// GetVersionedIndexName returns the versioned index name for a given base index and version
func (s *IndexVersionService) GetVersionedIndexName(baseIndex string, version int) string {
	return getVersionedIndexName(baseIndex, version)
}
