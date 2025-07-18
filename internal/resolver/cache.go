package resolver

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aaronsb/atlassian-assets/internal/logger"
)

// PersistentCacheEntry represents a cached resolver entry with metadata
type PersistentCacheEntry struct {
	WorkspaceID string                `json:"workspace_id"`
	SiteURL     string                `json:"site_url"`
	Schemas     map[string]*EntityInfo `json:"schemas"`
	SchemasByName map[string]*EntityInfo `json:"schemas_by_name"`
	ObjectTypes map[string]*EntityInfo `json:"object_types"`
	TypesByName map[string]*EntityInfo `json:"types_by_name"`
	CachedAt    time.Time             `json:"cached_at"`
	ExpiresAt   time.Time             `json:"expires_at"`
	Version     int                   `json:"version"`
}

// DiskCache handles persistent caching of resolver data
type DiskCache struct {
	cacheDir string
	ttl      time.Duration
}

// NewDiskCache creates a new disk cache instance
func NewDiskCache(baseCacheDir string, ttl time.Duration) (*DiskCache, error) {
	// Create resolver-specific subdirectory
	cacheDir := filepath.Join(baseCacheDir, "resolver")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &DiskCache{
		cacheDir: cacheDir,
		ttl:      ttl,
	}, nil
}

// getCacheKey generates a consistent cache key for a workspace
func (dc *DiskCache) getCacheKey(workspaceID, siteURL string) string {
	// Create a hash of workspace ID and site URL for the cache key
	hasher := sha256.New()
	hasher.Write([]byte(workspaceID + "|" + siteURL))
	return fmt.Sprintf("workspace_%x.json", hasher.Sum(nil))
}

// getCacheFilePath returns the full path to the cache file
func (dc *DiskCache) getCacheFilePath(workspaceID, siteURL string) string {
	return filepath.Join(dc.cacheDir, dc.getCacheKey(workspaceID, siteURL))
}

// LoadCache loads cached resolver data from disk
func (dc *DiskCache) LoadCache(workspaceID, siteURL string) (*PersistentCacheEntry, error) {
	filePath := dc.getCacheFilePath(workspaceID, siteURL)
	
	// Check if cache file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("cache file not found")
	}

	// Read cache file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	// Parse JSON
	var entry PersistentCacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to parse cache file: %w", err)
	}

	// Check if cache has expired
	if time.Now().After(entry.ExpiresAt) {
		return nil, fmt.Errorf("cache has expired")
	}

	// Validate workspace ID matches
	if entry.WorkspaceID != workspaceID {
		return nil, fmt.Errorf("workspace ID mismatch in cache")
	}

	return &entry, nil
}

// SaveCache saves resolver data to disk cache
func (dc *DiskCache) SaveCache(workspaceID, siteURL string, cache *ResolverCache) error {
	entry := &PersistentCacheEntry{
		WorkspaceID:   workspaceID,
		SiteURL:       siteURL,
		Schemas:       make(map[string]*EntityInfo),
		SchemasByName: make(map[string]*EntityInfo),
		ObjectTypes:   make(map[string]*EntityInfo),
		TypesByName:   make(map[string]*EntityInfo),
		CachedAt:      time.Now(),
		ExpiresAt:     time.Now().Add(dc.ttl),
		Version:       1,
	}

	// Copy cache data (thread-safe copy)
	for k, v := range cache.schemas {
		entry.Schemas[k] = v
	}
	for k, v := range cache.schemasByName {
		entry.SchemasByName[k] = v
	}
	for k, v := range cache.objectTypes {
		entry.ObjectTypes[k] = v
	}
	for k, v := range cache.typesByName {
		entry.TypesByName[k] = v
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	// Write to file atomically
	filePath := dc.getCacheFilePath(workspaceID, siteURL)
	tempPath := filePath + ".tmp"
	
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	if err := os.Rename(tempPath, filePath); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return fmt.Errorf("failed to move cache file: %w", err)
	}

	return nil
}

// ListCachedWorkspaces returns information about all cached workspaces
func (dc *DiskCache) ListCachedWorkspaces() ([]*CacheInfo, error) {
	files, err := os.ReadDir(dc.cacheDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache directory: %w", err)
	}

	var cacheInfos []*CacheInfo
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			info, err := dc.getCacheInfo(filepath.Join(dc.cacheDir, file.Name()))
			if err != nil {
				// Log error but continue with other files
				logger.Warning("failed to read cache info for %s: %v", file.Name(), err)
				continue
			}
			cacheInfos = append(cacheInfos, info)
		}
	}

	return cacheInfos, nil
}

// CacheInfo holds summary information about a cached workspace
type CacheInfo struct {
	WorkspaceID   string    `json:"workspace_id"`
	SiteURL       string    `json:"site_url"`
	SchemaCount   int       `json:"schema_count"`
	ObjectTypeCount int     `json:"object_type_count"`
	CachedAt      time.Time `json:"cached_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	IsExpired     bool      `json:"is_expired"`
	SizeBytes     int64     `json:"size_bytes"`
	FileName      string    `json:"file_name"`
}

// getCacheInfo reads cache file header information
func (dc *DiskCache) getCacheInfo(filePath string) (*CacheInfo, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var entry PersistentCacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}

	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	return &CacheInfo{
		WorkspaceID:     entry.WorkspaceID,
		SiteURL:         entry.SiteURL,
		SchemaCount:     len(entry.Schemas),
		ObjectTypeCount: len(entry.ObjectTypes),
		CachedAt:        entry.CachedAt,
		ExpiresAt:       entry.ExpiresAt,
		IsExpired:       time.Now().After(entry.ExpiresAt),
		SizeBytes:       stat.Size(),
		FileName:        filepath.Base(filePath),
	}, nil
}

// ClearExpiredCache removes expired cache files
func (dc *DiskCache) ClearExpiredCache() error {
	cacheInfos, err := dc.ListCachedWorkspaces()
	if err != nil {
		return err
	}

	var removedCount int
	for _, info := range cacheInfos {
		if info.IsExpired {
			filePath := filepath.Join(dc.cacheDir, info.FileName)
			if err := os.Remove(filePath); err != nil {
				logger.Warning("failed to remove expired cache file %s: %v", info.FileName, err)
			} else {
				removedCount++
			}
		}
	}

	if removedCount > 0 {
		logger.Info("Removed %d expired cache files", removedCount)
	}

	return nil
}

// ClearAllCache removes all cache files
func (dc *DiskCache) ClearAllCache() error {
	return os.RemoveAll(dc.cacheDir)
}