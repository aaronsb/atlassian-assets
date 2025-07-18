package resolver

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ctreminiom/go-atlassian/v2/pkg/infra/models"
	"github.com/aaronsb/atlassian-assets/internal/client"
	"github.com/aaronsb/atlassian-assets/internal/logger"
)

// IDType represents the type of entity being resolved
type IDType string

const (
	IDTypeSchema     IDType = "schema"
	IDTypeObjectType IDType = "object_type"
	IDTypeObject     IDType = "object"
)

// EntityInfo holds information about a resolved entity
type EntityInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name,omitempty"`
	Type        IDType    `json:"type"`
	SchemaID    string    `json:"schema_id,omitempty"`
	ParentID    string    `json:"parent_id,omitempty"`
	LastUpdated time.Time `json:"last_updated"`
}

// ResolverCache holds cached resolution data
type ResolverCache struct {
	mu            sync.RWMutex
	schemas       map[string]*EntityInfo  // schema_id -> EntityInfo
	schemasByName map[string]*EntityInfo  // schema_name -> EntityInfo
	objectTypes   map[string]*EntityInfo  // object_type_id -> EntityInfo
	typesByName   map[string]*EntityInfo  // "schema_name/object_type_name" -> EntityInfo
	objects       map[string]*EntityInfo  // object_id -> EntityInfo
	objectsByKey  map[string]*EntityInfo  // "schema_name/object_key" -> EntityInfo
	lastRefresh   time.Time
	ttl           time.Duration
}

// Resolver provides bidirectional ID resolution between human names and internal IDs
type Resolver struct {
	client    *client.AssetsClient
	cache     *ResolverCache
	diskCache *DiskCache
}

// NewResolver creates a new ID resolver
func NewResolver(client *client.AssetsClient) *Resolver {
	config := client.GetConfig()
	
	// Initialize disk cache
	var diskCache *DiskCache
	if cacheDir, err := config.GetCacheDir(); err == nil {
		if dc, err := NewDiskCache(cacheDir, config.GetCacheTTL()); err == nil {
			diskCache = dc
		}
	}

	return &Resolver{
		client:    client,
		diskCache: diskCache,
		cache: &ResolverCache{
			schemas:       make(map[string]*EntityInfo),
			schemasByName: make(map[string]*EntityInfo),
			objectTypes:   make(map[string]*EntityInfo),
			typesByName:   make(map[string]*EntityInfo),
			objects:       make(map[string]*EntityInfo),
			objectsByKey:  make(map[string]*EntityInfo),
			ttl:           5 * time.Minute, // In-memory cache for 5 minutes
		},
	}
}

// RefreshCache refreshes the resolver cache with latest data from API
func (r *Resolver) RefreshCache(ctx context.Context) error {
	// Try to load from disk cache first
	if r.diskCache != nil {
		if err := r.loadFromDiskCache(); err == nil {
			return nil // Successfully loaded from disk
		}
		// If disk cache failed, continue with API refresh
	}

	// Add timeout to prevent hanging
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	r.cache.mu.Lock()
	defer r.cache.mu.Unlock()

	// Clear existing cache
	r.cache.schemas = make(map[string]*EntityInfo)
	r.cache.schemasByName = make(map[string]*EntityInfo)
	r.cache.objectTypes = make(map[string]*EntityInfo)
	r.cache.typesByName = make(map[string]*EntityInfo)

	// Load schemas
	if err := r.loadSchemas(timeoutCtx); err != nil {
		return fmt.Errorf("failed to load schemas: %w", err)
	}

	// Load object types for all schemas
	if err := r.loadObjectTypes(timeoutCtx); err != nil {
		return fmt.Errorf("failed to load object types: %w", err)
	}

	r.cache.lastRefresh = time.Now()
	
	// Save to disk cache
	if r.diskCache != nil {
		r.saveToDiskCache()
	}

	return nil
}

// loadSchemas loads all schemas into cache
func (r *Resolver) loadSchemas(ctx context.Context) error {
	response, err := r.client.ListSchemas(ctx)
	if err != nil {
		return err
	}

	if !response.Success {
		return fmt.Errorf("API error: %s", response.Error)
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected response format, got type: %T", response.Data)
	}

	schemasValue := data["schemas"]
	schemas, ok := schemasValue.([]*models.ObjectSchemaScheme)
	if !ok {
		return fmt.Errorf("schemas type mismatch: expected []*models.ObjectSchemaScheme, got %T", schemasValue)
	}

	for _, schema := range schemas {
		if schema == nil {
			continue
		}

		id := schema.ID
		
		name := schema.Name
		
		entity := &EntityInfo{
			ID:          id,
			Name:        name,
			Type:        IDTypeSchema,
			LastUpdated: time.Now(),
		}

		r.cache.schemas[id] = entity
		r.cache.schemasByName[strings.ToLower(name)] = entity
	}

	return nil
}

// loadObjectTypes loads object types for all schemas
func (r *Resolver) loadObjectTypes(ctx context.Context) error {
	schemaCount := len(r.cache.schemas)
	logger.Info("Loading object types for %d schemas...", schemaCount)
	
	processed := 0
	for schemaID := range r.cache.schemas {
		processed++
		logger.Info("Loading schema %d/%d (ID: %s)...", processed, schemaCount, schemaID)
		
		if err := r.loadObjectTypesForSchema(ctx, schemaID); err != nil {
			// Log error but continue with other schemas
			logger.Error("failed to load object types for schema %s: %v", schemaID, err)
		} else {
			logger.Info("SUCCESS: loaded object types for schema %s", schemaID)
		}
	}
	logger.Info("Finished loading object types for all schemas")
	return nil
}

// loadObjectTypesForSchema loads object types for a specific schema
func (r *Resolver) loadObjectTypesForSchema(ctx context.Context, schemaID string) error {
	response, err := r.client.GetObjectTypes(ctx, schemaID)
	if err != nil {
		return err
	}

	if !response.Success {
		return fmt.Errorf("API error: %s", response.Error)
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected response format")
	}

	objectTypesValue := data["object_types"]
	objectTypes, ok := objectTypesValue.([]*models.ObjectTypeScheme)
	if !ok {
		return fmt.Errorf("object_types type mismatch: expected []*models.ObjectTypeScheme, got %T", objectTypesValue)
	}

	schemaInfo := r.cache.schemas[schemaID]
	schemaName := ""
	if schemaInfo != nil {
		schemaName = schemaInfo.Name
	}

	for _, objectType := range objectTypes {
		if objectType == nil {
			continue
		}

		id := objectType.ID
		
		name := objectType.Name
		
		entity := &EntityInfo{
			ID:          id,
			Name:        name,
			Type:        IDTypeObjectType,
			SchemaID:    schemaID,
			ParentID:    schemaID,
			LastUpdated: time.Now(),
		}

		r.cache.objectTypes[id] = entity
		
		// Create compound key for name-based lookup: "schema_name/object_type_name"
		if schemaName != "" {
			compoundKey := fmt.Sprintf("%s/%s", strings.ToLower(schemaName), strings.ToLower(name))
			r.cache.typesByName[compoundKey] = entity
		}
	}

	return nil
}

// needsRefresh checks if cache needs refreshing
func (r *Resolver) needsRefresh() bool {
	r.cache.mu.RLock()
	defer r.cache.mu.RUnlock()
	
	return time.Since(r.cache.lastRefresh) > r.cache.ttl
}

// ensureCache ensures cache is fresh
func (r *Resolver) ensureCache(ctx context.Context) error {
	if r.needsRefresh() {
		return r.RefreshCache(ctx)
	}
	return nil
}

// ResolveSchemaID resolves a schema name to its ID
func (r *Resolver) ResolveSchemaID(ctx context.Context, nameOrID string) (string, error) {
	if err := r.ensureCache(ctx); err != nil {
		return "", err
	}

	r.cache.mu.RLock()
	defer r.cache.mu.RUnlock()

	// Check if it's already an ID
	if entity, exists := r.cache.schemas[nameOrID]; exists {
		return entity.ID, nil
	}

	// Try to resolve by name
	if entity, exists := r.cache.schemasByName[strings.ToLower(nameOrID)]; exists {
		return entity.ID, nil
	}

	return "", fmt.Errorf("schema not found: %s", nameOrID)
}

// ResolveSchemaName resolves a schema ID to its name
func (r *Resolver) ResolveSchemaName(ctx context.Context, id string) (string, error) {
	if err := r.ensureCache(ctx); err != nil {
		return "", err
	}

	r.cache.mu.RLock()
	defer r.cache.mu.RUnlock()

	if entity, exists := r.cache.schemas[id]; exists {
		return entity.Name, nil
	}

	return "", fmt.Errorf("schema ID not found: %s", id)
}

// ResolveObjectTypeID resolves an object type name to its ID within a schema
func (r *Resolver) ResolveObjectTypeID(ctx context.Context, schemaNameOrID, typeNameOrID string) (string, error) {
	if err := r.ensureCache(ctx); err != nil {
		return "", err
	}

	r.cache.mu.RLock()
	defer r.cache.mu.RUnlock()

	// First resolve schema name to get consistent schema name
	schemaID, err := r.ResolveSchemaID(ctx, schemaNameOrID)
	if err != nil {
		return "", fmt.Errorf("failed to resolve schema: %w", err)
	}

	schemaEntity := r.cache.schemas[schemaID]
	if schemaEntity == nil {
		return "", fmt.Errorf("schema not found: %s", schemaNameOrID)
	}

	// Check if typeNameOrID is already an object type ID
	if entity, exists := r.cache.objectTypes[typeNameOrID]; exists && entity.SchemaID == schemaID {
		return entity.ID, nil
	}

	// Try to resolve by compound name
	compoundKey := fmt.Sprintf("%s/%s", strings.ToLower(schemaEntity.Name), strings.ToLower(typeNameOrID))
	if entity, exists := r.cache.typesByName[compoundKey]; exists {
		return entity.ID, nil
	}

	return "", fmt.Errorf("object type not found: %s in schema %s", typeNameOrID, schemaEntity.Name)
}

// ResolveObjectTypeName resolves an object type ID to its name
func (r *Resolver) ResolveObjectTypeName(ctx context.Context, typeID string) (string, string, error) {
	if err := r.ensureCache(ctx); err != nil {
		return "", "", err
	}

	r.cache.mu.RLock()
	defer r.cache.mu.RUnlock()

	if entity, exists := r.cache.objectTypes[typeID]; exists {
		schemaName := ""
		if schemaEntity, exists := r.cache.schemas[entity.SchemaID]; exists {
			schemaName = schemaEntity.Name
		}
		return entity.Name, schemaName, nil
	}

	return "", "", fmt.Errorf("object type ID not found: %s", typeID)
}

// GetObjectInfo retrieves and caches object information
func (r *Resolver) GetObjectInfo(ctx context.Context, objectID string) (*EntityInfo, error) {
	r.cache.mu.RLock()
	if entity, exists := r.cache.objects[objectID]; exists {
		// Check if cached entry is still fresh
		if time.Since(entity.LastUpdated) < r.cache.ttl {
			r.cache.mu.RUnlock()
			return entity, nil
		}
	}
	r.cache.mu.RUnlock()

	// Fetch fresh object data
	response, err := r.client.GetObject(ctx, objectID)
	if err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, fmt.Errorf("API error: %s", response.Error)
	}

	object, ok := response.Data.(*models.ObjectScheme)
	if !ok {
		return nil, fmt.Errorf("unexpected response format: expected *models.ObjectScheme, got %T", response.Data)
	}

	// Extract object information
	id := object.ID
	label := object.Label
	objectKey := object.ObjectKey

	// Extract object type and schema info
	var schemaID, objectTypeID string
	if object.ObjectType != nil {
		objectTypeID = object.ObjectType.ID
		schemaID = object.ObjectType.ObjectSchemaID
	}

	entity := &EntityInfo{
		ID:          id,
		Name:        objectKey,
		DisplayName: label,
		Type:        IDTypeObject,
		SchemaID:    schemaID,
		ParentID:    objectTypeID,
		LastUpdated: time.Now(),
	}

	// Cache the object info
	r.cache.mu.Lock()
	r.cache.objects[id] = entity
	if objectKey != "" && schemaID != "" {
		// Resolve schema name for compound key
		if schemaEntity, exists := r.cache.schemas[schemaID]; exists {
			compoundKey := fmt.Sprintf("%s/%s", strings.ToLower(schemaEntity.Name), strings.ToLower(objectKey))
			r.cache.objectsByKey[compoundKey] = entity
		}
	}
	r.cache.mu.Unlock()

	return entity, nil
}

// ResolveObjectID resolves an object reference to its ID
// Supports: numeric ID, object key, "schema/object_key" format
func (r *Resolver) ResolveObjectID(ctx context.Context, reference string) (string, error) {
	// Check if it's a numeric ID
	if _, err := strconv.Atoi(reference); err == nil {
		// Try to get the object info (this will fetch from API if not cached)
		if _, err := r.GetObjectInfo(ctx, reference); err != nil {
			return "", fmt.Errorf("object ID not found: %s", reference)
		}
		return reference, nil
	}

	// Check if it's a compound reference (schema/object_key)
	if strings.Contains(reference, "/") {
		r.cache.mu.RLock()
		if entity, exists := r.cache.objectsByKey[strings.ToLower(reference)]; exists {
			r.cache.mu.RUnlock()
			return entity.ID, nil
		}
		r.cache.mu.RUnlock()
		
		// Parse schema and object key
		parts := strings.SplitN(reference, "/", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid compound reference format: %s", reference)
		}
		
		return "", fmt.Errorf("compound reference '%s' not found in cache. Use numeric ID instead (e.g., '384') or access the object by ID first to cache it", reference)
	}

	// Try to find by object key across all cached objects first
	r.cache.mu.RLock()
	for _, entity := range r.cache.objects {
		if strings.EqualFold(entity.Name, reference) {
			r.cache.mu.RUnlock()
			return entity.ID, nil
		}
	}
	r.cache.mu.RUnlock()

	// Cache miss - provide helpful guidance to user
	return "", fmt.Errorf("object key '%s' not found in cache. Use numeric ID instead (e.g., '384') or first access the object by ID to cache it", reference)
}

// GetCacheStats returns statistics about the resolver cache
func (r *Resolver) GetCacheStats() map[string]interface{} {
	r.cache.mu.RLock()
	defer r.cache.mu.RUnlock()

	return map[string]interface{}{
		"schemas":      len(r.cache.schemas),
		"object_types": len(r.cache.objectTypes),
		"objects":      len(r.cache.objects),
		"last_refresh": r.cache.lastRefresh,
		"ttl_minutes":  r.cache.ttl.Minutes(),
		"needs_refresh": time.Since(r.cache.lastRefresh) > r.cache.ttl,
	}
}

// ListResolvedSchemas returns all cached schemas with their resolution info
func (r *Resolver) ListResolvedSchemas(ctx context.Context) ([]*EntityInfo, error) {
	if err := r.ensureCache(ctx); err != nil {
		return nil, err
	}

	r.cache.mu.RLock()
	defer r.cache.mu.RUnlock()

	var schemas []*EntityInfo
	for _, entity := range r.cache.schemas {
		schemas = append(schemas, entity)
	}

	return schemas, nil
}

// ListResolvedObjectTypes returns all cached object types for a schema
func (r *Resolver) ListResolvedObjectTypes(ctx context.Context, schemaNameOrID string) ([]*EntityInfo, error) {
	if err := r.ensureCache(ctx); err != nil {
		return nil, err
	}

	schemaID, err := r.ResolveSchemaID(ctx, schemaNameOrID)
	if err != nil {
		return nil, err
	}

	r.cache.mu.RLock()
	defer r.cache.mu.RUnlock()

	var objectTypes []*EntityInfo
	for _, entity := range r.cache.objectTypes {
		if entity.SchemaID == schemaID {
			objectTypes = append(objectTypes, entity)
		}
	}

	return objectTypes, nil
}

// loadFromDiskCache loads cache data from disk
func (r *Resolver) loadFromDiskCache() error {
	if r.diskCache == nil {
		return fmt.Errorf("disk cache not available")
	}

	workspaceID := r.client.GetWorkspaceID()
	siteURL := r.client.GetConfig().GetBaseURL()

	entry, err := r.diskCache.LoadCache(workspaceID, siteURL)
	if err != nil {
		return err
	}

	r.cache.mu.Lock()
	defer r.cache.mu.Unlock()

	// Load cached data
	r.cache.schemas = entry.Schemas
	r.cache.schemasByName = entry.SchemasByName
	r.cache.objectTypes = entry.ObjectTypes
	r.cache.typesByName = entry.TypesByName
	r.cache.lastRefresh = entry.CachedAt

	return nil
}

// saveToDiskCache saves current cache data to disk
func (r *Resolver) saveToDiskCache() {
	if r.diskCache == nil {
		return
	}

	workspaceID := r.client.GetWorkspaceID()
	siteURL := r.client.GetConfig().GetBaseURL()

	if err := r.diskCache.SaveCache(workspaceID, siteURL, r.cache); err != nil {
		logger.Warning("failed to save cache to disk: %v", err)
	}
}

