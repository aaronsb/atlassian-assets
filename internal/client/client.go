package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ctreminiom/go-atlassian/v2/assets"
	"github.com/ctreminiom/go-atlassian/v2/pkg/infra/models"
	"github.com/aaronsb/atlassian-assets/internal/config"
)

// AssetsClient wraps the go-atlassian client with our configuration
type AssetsClient struct {
	httpClient  *http.Client
	assetsAPI   *assets.Client
	config      *config.Config
	workspaceID string
}

// NewAssetsClient creates a new Assets client with the given configuration
func NewAssetsClient(cfg *config.Config) (*AssetsClient, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create HTTP client for basic auth
	httpClient := &http.Client{}

	// Create the assets API client - use default Assets API URL
	assetsAPI, err := assets.New(httpClient, "https://api.atlassian.com/")
	if err != nil {
		return nil, fmt.Errorf("failed to create assets client: %w", err)
	}

	// Set basic authentication
	assetsAPI.Auth.SetBasicAuth(cfg.GetUsername(), cfg.GetPassword())

	ac := &AssetsClient{
		httpClient: httpClient,
		assetsAPI:  assetsAPI,
		config:     cfg,
	}

	// Discover workspace ID if not provided
	if cfg.WorkspaceID == "" {
		workspaceID, err := ac.discoverWorkspaceID(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to discover workspace ID: %w", err)
		}
		ac.workspaceID = workspaceID
	} else {
		ac.workspaceID = cfg.WorkspaceID
	}

	return ac, nil
}

// discoverWorkspaceID attempts to discover the workspace ID from the site
// This addresses the abstraction issue you mentioned about site name -> workspace UID
func (ac *AssetsClient) discoverWorkspaceID(ctx context.Context) (string, error) {
	// Use the JSM Service Desk API to get workspace information
	// This is the key to bridging site name -> workspace UID
	
	url := fmt.Sprintf("%s/rest/servicedeskapi/insight/workspace", ac.config.GetBaseURL())
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create workspace discovery request: %w", err)
	}

	// Set basic auth
	req.SetBasicAuth(ac.config.GetUsername(), ac.config.GetPassword())
	req.Header.Set("Accept", "application/json")

	resp, err := ac.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to discover workspace: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("workspace discovery failed: HTTP %d", resp.StatusCode)
	}

	var workspaceResp struct {
		Values []struct {
			WorkspaceID string `json:"workspaceId"`
		} `json:"values"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&workspaceResp); err != nil {
		return "", fmt.Errorf("failed to decode workspace response: %w", err)
	}

	if len(workspaceResp.Values) == 0 {
		return "", fmt.Errorf("no workspaces found")
	}

	// Return the first workspace ID
	return workspaceResp.Values[0].WorkspaceID, nil
}

// GetWorkspaceID returns the current workspace ID
func (ac *AssetsClient) GetWorkspaceID() string {
	return ac.workspaceID
}

// TestConnection tests the connection to the Atlassian instance
func (ac *AssetsClient) TestConnection(ctx context.Context) error {
	// Try to make a simple API call to verify connectivity
	// This could be a call to get current user or workspace info
	return fmt.Errorf("connection test not yet implemented")
}

// GetAssetsAPI returns the assets API client
func (ac *AssetsClient) GetAssetsAPI() *assets.Client {
	return ac.assetsAPI
}

// GetConfig returns the current configuration
func (ac *AssetsClient) GetConfig() *config.Config {
	return ac.config
}

// Close closes the client connection
func (ac *AssetsClient) Close() error {
	// Clean up any resources if needed
	return nil
}

// WithWorkspaceID sets the workspace ID for this client
func (ac *AssetsClient) WithWorkspaceID(workspaceID string) *AssetsClient {
	ac.workspaceID = workspaceID
	return ac
}

// Common response wrapper for consistent error handling
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// NewSuccessResponse creates a successful response
func NewSuccessResponse(data interface{}) *Response {
	return &Response{
		Success: true,
		Data:    data,
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(err error) *Response {
	return &Response{
		Success: false,
		Error:   err.Error(),
	}
}

// ListSchemas lists all object schemas in the workspace
func (ac *AssetsClient) ListSchemas(ctx context.Context) (*Response, error) {
	if ac.workspaceID == "" {
		return NewErrorResponse(fmt.Errorf("workspace ID not set")), nil
	}

	schemas, response, err := ac.assetsAPI.ObjectSchema.List(ctx, ac.workspaceID)
	if err != nil {
		return NewErrorResponse(fmt.Errorf("failed to list schemas: %w", err)), nil
	}

	if response.Code != 200 {
		return NewErrorResponse(fmt.Errorf("API error: %d - %s", response.Code, response.Bytes.String())), nil
	}

	return NewSuccessResponse(map[string]interface{}{
		"schemas": schemas.Values,
		"total":   schemas.Total,
	}), nil
}

// GetSchema gets details of a specific schema
func (ac *AssetsClient) GetSchema(ctx context.Context, schemaID string) (*Response, error) {
	if ac.workspaceID == "" {
		return NewErrorResponse(fmt.Errorf("workspace ID not set")), nil
	}

	schema, response, err := ac.assetsAPI.ObjectSchema.Get(ctx, ac.workspaceID, schemaID)
	if err != nil {
		return NewErrorResponse(fmt.Errorf("failed to get schema: %w", err)), nil
	}

	if response.Code != 200 {
		return NewErrorResponse(fmt.Errorf("API error: %d - %s", response.Code, response.Bytes.String())), nil
	}

	return NewSuccessResponse(schema), nil
}

// GetObjectTypes gets object types for a specific schema
func (ac *AssetsClient) GetObjectTypes(ctx context.Context, schemaID string) (*Response, error) {
	if ac.workspaceID == "" {
		return NewErrorResponse(fmt.Errorf("workspace ID not set")), nil
	}

	objectTypes, response, err := ac.assetsAPI.ObjectSchema.ObjectTypes(ctx, ac.workspaceID, schemaID, false)
	if err != nil {
		return NewErrorResponse(fmt.Errorf("failed to get object types: %w", err)), nil
	}

	if response.Code != 200 {
		return NewErrorResponse(fmt.Errorf("API error: %d - %s", response.Code, response.Bytes.String())), nil
	}

	return NewSuccessResponse(map[string]interface{}{
		"object_types": objectTypes,
		"schema":       schemaID,
		"count":        len(objectTypes),
	}), nil
}

// ListObjects lists objects in a specific schema using structured search
func (ac *AssetsClient) ListObjects(ctx context.Context, schemaID string, limit int) (*Response, error) {
	if ac.workspaceID == "" {
		return NewErrorResponse(fmt.Errorf("workspace ID not set")), nil
	}

	// Use structured search without Query to get all objects in schema
	searchParams := &models.ObjectSearchParamsScheme{
		ObjectSchemaID:    schemaID,
		ResultPerPage:     limit,
		Page:              1,
		IncludeAttributes: true,
		// Note: No Query field means get all objects in the schema
	}
	
	fmt.Printf("DEBUG: Calling Search with workspaceID=%s, objectSchemaID=%s, limit=%d\n", ac.workspaceID, schemaID, limit)
	objects, response, err := ac.assetsAPI.Object.Search(ctx, ac.workspaceID, searchParams)
	
	if objects != nil {
		fmt.Printf("DEBUG: Search returned err=%v, response.Code=%d, objects.TotalFilterCount=%d\n", err, response.Code, objects.TotalFilterCount)
	} else {
		fmt.Printf("DEBUG: Search returned err=%v, response.Code=%d, objects=nil\n", err, response.Code)
	}
	
	if err != nil {
		return NewErrorResponse(fmt.Errorf("failed to list objects: %w", err)), nil
	}

	if response.Code != 200 {
		return NewErrorResponse(fmt.Errorf("API error: %d - %s", response.Code, response.Bytes.String())), nil
	}

	if objects == nil {
		return NewSuccessResponse(map[string]interface{}{
			"objects": []interface{}{},
			"total":   0,
			"schema":  schemaID,
			"method":  "structured_search",
			"error":   "objects response is nil",
		}), nil
	}

	return NewSuccessResponse(map[string]interface{}{
		"objects": objects.ObjectEntries,
		"total":   objects.TotalFilterCount,
		"schema":  schemaID,
		"method":  "structured_search",
	}), nil
}

// GetObject gets a specific object by ID
func (ac *AssetsClient) GetObject(ctx context.Context, objectID string) (*Response, error) {
	if ac.workspaceID == "" {
		return NewErrorResponse(fmt.Errorf("workspace ID not set")), nil
	}

	object, response, err := ac.assetsAPI.Object.Get(ctx, ac.workspaceID, objectID)
	if err != nil {
		return NewErrorResponse(fmt.Errorf("failed to get object: %w", err)), nil
	}

	if response.Code != 200 {
		return NewErrorResponse(fmt.Errorf("API error: %d - %s", response.Code, response.Bytes.String())), nil
	}

	return NewSuccessResponse(object), nil
}

// GetObjectTypeAttributes gets all attributes for a specific object type
func (ac *AssetsClient) GetObjectTypeAttributes(ctx context.Context, objectTypeID string) (*Response, error) {
	if ac.workspaceID == "" {
		return NewErrorResponse(fmt.Errorf("workspace ID not set")), nil
	}

	// Get all attributes for this object type
	attributes, response, err := ac.assetsAPI.ObjectType.Attributes(ctx, ac.workspaceID, objectTypeID, &models.ObjectTypeAttributesParamsScheme{
		OrderByName:     true,
		OrderByRequired: true,
	})
	if err != nil {
		return NewErrorResponse(fmt.Errorf("failed to get object type attributes: %w", err)), nil
	}

	if response.Code != 200 {
		return NewErrorResponse(fmt.Errorf("API error: %d - %s", response.Code, response.Bytes.String())), nil
	}

	return NewSuccessResponse(map[string]interface{}{
		"object_type_id": objectTypeID,
		"attributes":     attributes,
		"count":          len(attributes),
	}), nil
}

// SearchObjects searches for objects using AQL with human-readable results
func (ac *AssetsClient) SearchObjects(ctx context.Context, query string, limit int) (*Response, error) {
	if ac.workspaceID == "" {
		return NewErrorResponse(fmt.Errorf("workspace ID not set")), nil
	}

	objects, response, err := ac.assetsAPI.Object.Filter(ctx, ac.workspaceID, query, true, limit, 0)
	if err != nil {
		return NewErrorResponse(fmt.Errorf("failed to search objects: %w", err)), nil
	}

	if response.Code != 200 {
		return NewErrorResponse(fmt.Errorf("API error: %d - %s", response.Code, response.Bytes.String())), nil
	}

	return NewSuccessResponse(map[string]interface{}{
		"objects": objects.Values,
		"total":   objects.Total,
		"query":   query,
	}), nil
}

// CreateObjectType creates a new object type in the specified schema
func (ac *AssetsClient) CreateObjectType(ctx context.Context, schemaID, name, description, iconID string, parentObjectTypeID *string) (*Response, error) {
	if ac.workspaceID == "" {
		return NewErrorResponse(fmt.Errorf("workspace ID not set")), nil
	}

	payload := &models.ObjectTypePayloadScheme{
		Name:               name,
		Description:        description,
		IconID:             iconID,
		ObjectSchemaID:     schemaID,
		Inherited:          false,
		AbstractObjectType: false,
	}

	// Set parent object type if provided
	if parentObjectTypeID != nil {
		payload.ParentObjectTypeID = *parentObjectTypeID
	}

	objectType, response, err := ac.assetsAPI.ObjectType.Create(ctx, ac.workspaceID, payload)
	if err != nil {
		return NewErrorResponse(fmt.Errorf("failed to create object type: %w", err)), nil
	}

	if response.Code != 201 && response.Code != 200 {
		return NewErrorResponse(fmt.Errorf("API error: %d - %s", response.Code, response.Bytes.String())), nil
	}

	return NewSuccessResponse(map[string]interface{}{
		"object_type": objectType,
		"message":     fmt.Sprintf("Successfully created object type '%s' in schema %s", name, schemaID),
	}), nil
}