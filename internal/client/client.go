package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ctreminiom/go-atlassian/v2/assets"
	"github.com/ctreminiom/go-atlassian/v2/pkg/infra/models"
	"github.com/aaronsb/atlassian-assets/internal/config"
	"github.com/aaronsb/atlassian-assets/internal/logger"
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

// IsDeleteAllowed checks if delete operations are allowed
func (ac *AssetsClient) IsDeleteAllowed() bool {
	return ac.config.AllowDelete
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

// CreateSchema creates a new object schema
func (ac *AssetsClient) CreateSchema(ctx context.Context, name, description string) (*Response, error) {
	if ac.workspaceID == "" {
		return NewErrorResponse(fmt.Errorf("workspace ID not set")), nil
	}

	// Create schema payload with proper structure
	payload := &models.ObjectSchemaPayloadScheme{
		Name:            name,
		ObjectSchemaKey: generateSchemaKey(name),
	}
	
	if description != "" {
		payload.Description = description
	}

	schema, response, err := ac.assetsAPI.ObjectSchema.Create(ctx, ac.workspaceID, payload)
	if err != nil {
		return NewErrorResponse(fmt.Errorf("failed to create schema: %w", err)), nil
	}

	if response.Code != 201 && response.Code != 200 {
		return NewErrorResponse(fmt.Errorf("API error: %d - %s", response.Code, response.Bytes.String())), nil
	}

	return NewSuccessResponse(schema), nil
}

// generateSchemaKey creates a schema key from the name
func generateSchemaKey(name string) string {
	// Simple key generation - uppercase and replace spaces with underscores
	key := ""
	for _, char := range name {
		if char >= 'a' && char <= 'z' {
			key += string(char - 32) // Convert to uppercase
		} else if char >= 'A' && char <= 'Z' {
			key += string(char)
		} else if char >= '0' && char <= '9' {
			key += string(char)
		} else if char == ' ' || char == '-' {
			key += "_"
		}
	}
	
	// Limit to reasonable length
	if len(key) > 10 {
		key = key[:10]
	}
	
	return key
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
	return ac.ListObjectsWithPagination(ctx, schemaID, limit, 0)
}

// ListObjectsWithPagination lists objects in a specific schema with pagination support
func (ac *AssetsClient) ListObjectsWithPagination(ctx context.Context, schemaID string, limit int, offset int) (*Response, error) {
	if ac.workspaceID == "" {
		return NewErrorResponse(fmt.Errorf("workspace ID not set")), nil
	}

	// Use AQL query to get all objects in schema - now using direct HTTP call
	query := fmt.Sprintf("objectSchemaId = %s", schemaID)
	
	logger.Debug("ListObjectsWithPagination using direct HTTP call - workspaceID=%s, query=%s, limit=%d, offset=%d", ac.workspaceID, query, limit, offset)
	
	// Use direct HTTP call to bypass broken SDK
	objects, err := ac.searchObjectsDirectWithPagination(ctx, query, limit, offset)
	if err != nil {
		return NewErrorResponse(fmt.Errorf("failed to list objects: %w", err)), nil
	}

	logger.Debug("ListObjects direct HTTP call successful - found %d objects", objects.Total)
	return NewSuccessResponse(map[string]interface{}{
		"objects": objects.Values,
		"total":   objects.Total,
		"schema":  schemaID,
		"method":  "direct_http",
		"query":   query,
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
	return ac.SearchObjectsWithPagination(ctx, query, limit, 0)
}

// SearchObjectsWithPagination searches for objects using AQL with pagination support
func (ac *AssetsClient) SearchObjectsWithPagination(ctx context.Context, query string, limit int, offset int) (*Response, error) {
	if ac.workspaceID == "" {
		return NewErrorResponse(fmt.Errorf("workspace ID not set")), nil
	}

	logger.Debug("SearchObjectsWithPagination using direct HTTP call - workspaceID=%s, query=%s, limit=%d, offset=%d", ac.workspaceID, query, limit, offset)
	
	// Use direct HTTP call to bypass broken SDK
	objects, err := ac.searchObjectsDirectWithPagination(ctx, query, limit, offset)
	if err != nil {
		return NewErrorResponse(fmt.Errorf("failed to search objects: %w", err)), nil
	}

	logger.Debug("SearchObjects direct HTTP call successful - found %d objects", objects.Total)
	return NewSuccessResponse(map[string]interface{}{
		"objects": objects.Values,
		"total":   objects.Total,
		"query":   query,
	}), nil
}

// searchObjectsDirect bypasses the broken SDK and makes direct HTTP calls
func (ac *AssetsClient) searchObjectsDirect(ctx context.Context, aqlQuery string, maxResults int) (*models.ObjectListResultScheme, error) {
	return ac.searchObjectsDirectWithPagination(ctx, aqlQuery, maxResults, 0)
}

// searchObjectsDirectWithPagination bypasses the broken SDK with pagination support
func (ac *AssetsClient) searchObjectsDirectWithPagination(ctx context.Context, aqlQuery string, maxResults int, startAt int) (*models.ObjectListResultScheme, error) {
	// Build endpoint URL with pagination parameters in URL (not payload)
	endpoint := fmt.Sprintf("https://api.atlassian.com/jsm/assets/workspace/%s/v1/object/aql?startAt=%d&maxResults=%d&includeAttributes=true", 
		ac.workspaceID, startAt, maxResults)
	
	// Create request payload with only the query (pagination is in URL)
	payload := map[string]interface{}{
		"qlQuery": aqlQuery,
	}
	
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	
	// Set Basic Auth
	auth := base64.StdEncoding.EncodeToString([]byte(ac.config.GetUsername() + ":" + ac.config.GetPassword()))
	req.Header.Set("Authorization", "Basic " + auth)
	
	logger.Debug("Direct HTTP POST to %s", endpoint)
	logger.Debug("Payload: %s", string(payloadBytes))
	
	// Make the request
	resp, err := ac.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	
	logger.Debug("Direct HTTP response status: %d", resp.StatusCode)
	
	if resp.StatusCode != 200 {
		var errorBody bytes.Buffer
		errorBody.ReadFrom(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, errorBody.String())
	}
	
	// Parse response
	var result models.ObjectListResultScheme
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Debug pagination info
	logger.Debug("API Response - Total: %d, MaxResults: %d, StartAt: %d, IsLast: %t", 
		result.Total, result.MaxResults, result.StartAt, result.IsLast)
	
	return &result, nil
}

// CreateObjectType creates a new object type in the specified schema
func (ac *AssetsClient) CreateObjectType(ctx context.Context, schemaID, name, description, iconID string, parentObjectTypeID *string) (*Response, error) {
	if ac.workspaceID == "" {
		return NewErrorResponse(fmt.Errorf("workspace ID not set")), nil
	}

	// Provide default icon if none specified (API now requires iconId)
	if iconID == "" {
		iconID = "1" // Default generic icon
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

// CreateObject creates a new object instance in the specified object type
func (ac *AssetsClient) CreateObject(ctx context.Context, objectTypeID string, attributes map[string]interface{}) (*Response, error) {
	if ac.workspaceID == "" {
		return NewErrorResponse(fmt.Errorf("workspace ID not set")), nil
	}

	// Get object type attributes to resolve names to IDs
	attrResponse, err := ac.GetObjectTypeAttributes(ctx, objectTypeID)
	if err != nil {
		return NewErrorResponse(fmt.Errorf("failed to get object type attributes: %w", err)), nil
	}

	if !attrResponse.Success {
		return NewErrorResponse(fmt.Errorf("failed to get object type attributes: %s", attrResponse.Error)), nil
	}

	// Build name-to-ID mapping using proper type handling
	attrData := attrResponse.Data.(map[string]interface{})
	attributesRaw := attrData["attributes"]
	nameToID := make(map[string]string)
	
	// Handle the attributes as they come from the API
	switch attrs := attributesRaw.(type) {
	case []*models.ObjectTypeAttributeScheme:
		for _, attr := range attrs {
			nameToID[attr.Name] = attr.ID
		}
	case []interface{}:
		for _, attr := range attrs {
			if attrMap, ok := attr.(map[string]interface{}); ok {
				if id, ok := attrMap["id"].(string); ok {
					if name, ok := attrMap["name"].(string); ok {
						nameToID[name] = id
					}
				}
			}
		}
	default:
		return NewErrorResponse(fmt.Errorf("unexpected attributes type: %T", attributesRaw)), nil
	}

	// Convert map[string]interface{} to ObjectPayloadScheme attributes using resolved IDs
	var objectAttributes []*models.ObjectPayloadAttributeScheme
	
	for key, value := range attributes {
		// Try to resolve attribute name to ID
		attributeID := key
		if id, found := nameToID[key]; found {
			attributeID = id
		}
		
		attr := &models.ObjectPayloadAttributeScheme{
			ObjectTypeAttributeID: attributeID,
			ObjectAttributeValues: []*models.ObjectPayloadAttributeValueScheme{
				{
					Value: fmt.Sprintf("%v", value),
				},
			},
		}
		objectAttributes = append(objectAttributes, attr)
	}

	payload := &models.ObjectPayloadScheme{
		ObjectTypeID: objectTypeID,
		Attributes:   objectAttributes,
	}

	// Debug output
	logger.Debug("Creating object with payload: ObjectTypeID=%s, AttributeCount=%d", objectTypeID, len(objectAttributes))
	logger.Debug("NameToID mapping: %+v", nameToID)
	for i, attr := range objectAttributes {
		logger.Debug("Attribute[%d]: ID=%s, ValueCount=%d", i, attr.ObjectTypeAttributeID, len(attr.ObjectAttributeValues))
		if len(attr.ObjectAttributeValues) > 0 {
			logger.Debug("FirstValue=%s", attr.ObjectAttributeValues[0].Value)
		}
	}

	object, response, err := ac.assetsAPI.Object.Create(ctx, ac.workspaceID, payload)
	if err != nil {
		return NewErrorResponse(fmt.Errorf("failed to create object: %w", err)), nil
	}

	if response.Code != 201 && response.Code != 200 {
		return NewErrorResponse(fmt.Errorf("API error: %d - %s", response.Code, response.Bytes.String())), nil
	}

	return NewSuccessResponse(map[string]interface{}{
		"object":      object,
		"object_type": objectTypeID,
		"message":     fmt.Sprintf("Successfully created object in object type %s", objectTypeID),
		"resolved_attributes": nameToID,
	}), nil
}

// CreateObjectTypeAttribute creates a new attribute on an object type
func (ac *AssetsClient) CreateObjectTypeAttribute(ctx context.Context, objectTypeID string, payload *models.ObjectTypeAttributePayloadScheme) (*Response, error) {
	if ac.workspaceID == "" {
		return NewErrorResponse(fmt.Errorf("workspace ID not set")), nil
	}

	attribute, response, err := ac.assetsAPI.ObjectTypeAttribute.Create(ctx, ac.workspaceID, objectTypeID, payload)
	if err != nil {
		return NewErrorResponse(fmt.Errorf("failed to create object type attribute: %w", err)), nil
	}

	if response.Code != 201 && response.Code != 200 {
		return NewErrorResponse(fmt.Errorf("API error: %d - %s", response.Code, response.Bytes.String())), nil
	}

	return NewSuccessResponse(map[string]interface{}{
		"attribute":      attribute,
		"object_type_id": objectTypeID,
		"message":        fmt.Sprintf("Successfully created attribute '%s' on object type %s", payload.Name, objectTypeID),
	}), nil
}

// DeleteObjectType deletes an object type (and all its instances)
func (ac *AssetsClient) DeleteObjectType(ctx context.Context, objectTypeID string) (*Response, error) {
	if ac.workspaceID == "" {
		return NewErrorResponse(fmt.Errorf("workspace ID not set")), nil
	}

	if !ac.config.AllowDelete {
		return NewErrorResponse(fmt.Errorf("delete operations are disabled")), nil
	}

	_, response, err := ac.assetsAPI.ObjectType.Delete(ctx, ac.workspaceID, objectTypeID)
	if err != nil {
		return NewErrorResponse(fmt.Errorf("failed to delete object type: %w", err)), nil
	}

	if response.Code != 204 && response.Code != 200 {
		return NewErrorResponse(fmt.Errorf("API error: %d - %s", response.Code, response.Bytes.String())), nil
	}

	return NewSuccessResponse(map[string]interface{}{
		"object_type_id": objectTypeID,
		"message":        fmt.Sprintf("Successfully deleted object type %s", objectTypeID),
	}), nil
}

// DeleteObject deletes an object instance
func (ac *AssetsClient) DeleteObject(ctx context.Context, objectID string) (*Response, error) {
	if ac.workspaceID == "" {
		return NewErrorResponse(fmt.Errorf("workspace ID not set")), nil
	}

	if !ac.config.AllowDelete {
		return NewErrorResponse(fmt.Errorf("delete operations are disabled")), nil
	}

	response, err := ac.assetsAPI.Object.Delete(ctx, ac.workspaceID, objectID)
	if err != nil {
		return NewErrorResponse(fmt.Errorf("failed to delete object: %w", err)), nil
	}

	if response.Code != 204 && response.Code != 200 {
		return NewErrorResponse(fmt.Errorf("API error: %d - %s", response.Code, response.Bytes.String())), nil
	}

	return NewSuccessResponse(map[string]interface{}{
		"object_id": objectID,
		"message":   fmt.Sprintf("Successfully deleted object %s", objectID),
	}), nil
}

// GetObjectType gets object type details
func (ac *AssetsClient) GetObjectType(ctx context.Context, objectTypeID string) (*Response, error) {
	if ac.workspaceID == "" {
		return NewErrorResponse(fmt.Errorf("workspace ID not set")), nil
	}

	objectType, response, err := ac.assetsAPI.ObjectType.Get(ctx, ac.workspaceID, objectTypeID)
	if err != nil {
		return NewErrorResponse(fmt.Errorf("failed to get object type: %w", err)), nil
	}

	if response.Code != 200 {
		return NewErrorResponse(fmt.Errorf("API error: %d - %s", response.Code, response.Bytes.String())), nil
	}

	return NewSuccessResponse(objectType), nil
}

// RemoveAttribute removes an attribute from an object type
func (ac *AssetsClient) RemoveAttribute(ctx context.Context, objectTypeID, attributeID string) (*Response, error) {
	if ac.workspaceID == "" {
		return NewErrorResponse(fmt.Errorf("workspace ID not set")), nil
	}

	response, err := ac.assetsAPI.ObjectTypeAttribute.Delete(ctx, ac.workspaceID, attributeID)
	if err != nil {
		return NewErrorResponse(fmt.Errorf("failed to remove attribute: %w", err)), nil
	}

	if response.Code != 204 && response.Code != 200 {
		return NewErrorResponse(fmt.Errorf("API error: %d - %s", response.Code, response.Bytes.String())), nil
	}

	return NewSuccessResponse(map[string]interface{}{
		"object_type_id": objectTypeID,
		"attribute_id":   attributeID,
		"message":        fmt.Sprintf("Successfully removed attribute %s from object type %s", attributeID, objectTypeID),
	}), nil
}

// RemoveRelationship removes a relationship from an object
func (ac *AssetsClient) RemoveRelationship(ctx context.Context, objectID, relationshipID string) (*Response, error) {
	if ac.workspaceID == "" {
		return NewErrorResponse(fmt.Errorf("workspace ID not set")), nil
	}

	// Note: This is a placeholder implementation as the exact API endpoint
	// for removing relationships may vary based on the Atlassian Assets API
	return NewErrorResponse(fmt.Errorf("remove relationship not yet implemented")), nil
}

// RemoveRelationshipByType removes a relationship by type and target
func (ac *AssetsClient) RemoveRelationshipByType(ctx context.Context, objectID, relationshipType, targetID string) (*Response, error) {
	if ac.workspaceID == "" {
		return NewErrorResponse(fmt.Errorf("workspace ID not set")), nil
	}

	// Note: This is a placeholder implementation as the exact API endpoint
	// for removing relationships by type may vary based on the Atlassian Assets API
	return NewErrorResponse(fmt.Errorf("remove relationship by type not yet implemented")), nil
}

// RemoveProperty removes a property from an object
func (ac *AssetsClient) RemoveProperty(ctx context.Context, objectID, propertyID string) (*Response, error) {
	if ac.workspaceID == "" {
		return NewErrorResponse(fmt.Errorf("workspace ID not set")), nil
	}

	// Note: This is a placeholder implementation as removing properties
	// typically involves updating the object with null values for the property
	return NewErrorResponse(fmt.Errorf("remove property not yet implemented")), nil
}

// RemovePropertyByName removes a property by name from an object
func (ac *AssetsClient) RemovePropertyByName(ctx context.Context, objectID, propertyName string) (*Response, error) {
	if ac.workspaceID == "" {
		return NewErrorResponse(fmt.Errorf("workspace ID not set")), nil
	}

	// Note: This is a placeholder implementation as removing properties by name
	// typically involves updating the object with null values for the property
	return NewErrorResponse(fmt.Errorf("remove property by name not yet implemented")), nil
}