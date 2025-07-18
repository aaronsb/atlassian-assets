package foundation

import (
	"context"
	"fmt"
	"strings"

	"github.com/aaronsb/atlassian-assets/cmd/assets/common"
)

// SearchObjects performs asset search using either simple terms or AQL
func SearchObjects(client common.ClientInterface, params common.SearchParams) (*common.Response, error) {
	// Validate parameters
	if params.Query == "" && params.Simple == "" {
		return common.NewErrorResponse(fmt.Errorf("either query or simple search term must be provided")), nil
	}

	// Validate pagination parameters
	if params.Limit < 1 || params.Limit > 1000 {
		return common.NewErrorResponse(fmt.Errorf("limit must be between 1 and 1000, got %d", params.Limit)), nil
	}
	if params.Offset < 0 {
		return common.NewErrorResponse(fmt.Errorf("offset must be 0 or greater, got %d", params.Offset)), nil
	}

	ctx := context.Background()
	var finalQuery string
	var queryType string

	if params.Query != "" {
		// Use direct AQL query
		finalQuery = params.Query
		queryType = "aql"
	} else {
		// Build AQL from simple search terms
		var err error
		finalQuery, err = buildSimpleSearchQuery(params.Simple, params.Schema, params.Type, params.Status, params.Owner)
		if err != nil {
			return common.NewErrorResponse(fmt.Errorf("failed to build search query: %w", err)), nil
		}
		queryType = "simple"
	}

	// Execute search
	response, err := client.SearchObjectsWithPagination(ctx, finalQuery, params.Limit, params.Offset)
	if err != nil {
		return common.NewErrorResponse(fmt.Errorf("failed to search objects: %w", err)), nil
	}

	// Return with metadata
	if response.Success {
		responseData := response.Data.(map[string]interface{})
		responseData["query_type"] = queryType
		responseData["query"] = finalQuery
		responseData["search_filters"] = buildFilterSummary(params)
		
		return common.NewSuccessResponse(responseData), nil
	}

	return common.NewErrorResponse(fmt.Errorf(response.Error)), nil
}

// ListObjects lists objects in a schema with optional filtering
func ListObjects(client common.ClientInterface, params common.ListParams) (*common.Response, error) {
	// Validate parameters
	if params.Schema == "" {
		return common.NewErrorResponse(fmt.Errorf("schema parameter is required")), nil
	}

	// Validate pagination parameters
	if params.Limit < 1 || params.Limit > 1000 {
		return common.NewErrorResponse(fmt.Errorf("limit must be between 1 and 1000, got %d", params.Limit)), nil
	}
	if params.Offset < 0 {
		return common.NewErrorResponse(fmt.Errorf("offset must be 0 or greater, got %d", params.Offset)), nil
	}

	ctx := context.Background()
	
	// Use ListObjectsWithPagination if available
	response, err := client.ListObjectsWithPagination(ctx, params.Schema, params.Limit, params.Offset)
	if err != nil {
		return common.NewErrorResponse(fmt.Errorf("failed to list objects: %w", err)), nil
	}

	// Add metadata
	if response.Success {
		responseData := response.Data.(map[string]interface{})
		responseData["schema"] = params.Schema
		responseData["operation"] = "list_objects"
		
		return common.NewSuccessResponse(responseData), nil
	}

	return common.NewErrorResponse(fmt.Errorf(response.Error)), nil
}

// GetObject retrieves a specific object by ID
func GetObject(client common.ClientInterface, params common.GetParams) (*common.Response, error) {
	// Validate parameters
	if params.ID == "" {
		return common.NewErrorResponse(fmt.Errorf("object ID is required")), nil
	}

	ctx := context.Background()
	response, err := client.GetObject(ctx, params.ID)
	if err != nil {
		return common.NewErrorResponse(fmt.Errorf("failed to get object: %w", err)), nil
	}

	// Add metadata
	if response.Success {
		responseData := map[string]interface{}{
			"object": response.Data,
			"object_id": params.ID,
			"operation": "get_object",
		}
		return common.NewSuccessResponse(responseData), nil
	}

	return common.NewErrorResponse(fmt.Errorf(response.Error)), nil
}

// CreateObject creates a new object instance
func CreateObject(client common.ClientInterface, params common.CreateObjectParams) (*common.Response, error) {
	// Validate parameters
	if params.ObjectTypeID == "" {
		return common.NewErrorResponse(fmt.Errorf("object type ID is required")), nil
	}
	if params.Attributes == nil || len(params.Attributes) == 0 {
		return common.NewErrorResponse(fmt.Errorf("attributes are required")), nil
	}

	ctx := context.Background()
	response, err := client.CreateObject(ctx, params.ObjectTypeID, params.Attributes)
	if err != nil {
		return common.NewErrorResponse(fmt.Errorf("failed to create object: %w", err)), nil
	}

	// Add metadata
	if response.Success {
		responseData := response.Data.(map[string]interface{})
		responseData["object_type_id"] = params.ObjectTypeID
		responseData["operation"] = "create_object"
		
		return common.NewSuccessResponse(responseData), nil
	}

	return common.NewErrorResponse(fmt.Errorf(response.Error)), nil
}

// DeleteObject deletes an object by ID
func DeleteObject(client common.ClientInterface, params common.DeleteObjectParams) (*common.Response, error) {
	// Validate parameters
	if params.ID == "" {
		return common.NewErrorResponse(fmt.Errorf("object ID is required")), nil
	}

	ctx := context.Background()
	response, err := client.DeleteObject(ctx, params.ID)
	if err != nil {
		return common.NewErrorResponse(fmt.Errorf("failed to delete object: %w", err)), nil
	}

	// Add metadata
	if response.Success {
		responseData := map[string]interface{}{
			"object_id": params.ID,
			"operation": "delete_object",
			"message": fmt.Sprintf("Successfully deleted object %s", params.ID),
		}
		return common.NewSuccessResponse(responseData), nil
	}

	return common.NewErrorResponse(fmt.Errorf(response.Error)), nil
}

// UpdateObject updates an existing object (placeholder for future implementation)
func UpdateObject(client common.ClientInterface, params common.UpdateObjectParams) (*common.Response, error) {
	// This would need to be implemented with the actual update functionality
	// For now, return a placeholder response
	return common.NewSuccessResponse(map[string]interface{}{
		"action": "update_object",
		"id":     params.ID,
		"data":   params.Data,
		"status": "not_implemented",
		"message": "Object update functionality will be implemented",
	}), nil
}

// Helper functions

// buildSimpleSearchQuery converts simple search terms into AQL
func buildSimpleSearchQuery(term, schema, objectType, status, owner string) (string, error) {
	var conditions []string

	// Add schema filter if provided
	if schema != "" {
		conditions = append(conditions, fmt.Sprintf("objectSchemaId = %s", schema))
	}

	// Add object type filter if provided
	if objectType != "" {
		conditions = append(conditions, fmt.Sprintf("objectType = \"%s\"", objectType))
	}

	// Add status filter if provided
	if status != "" {
		conditions = append(conditions, fmt.Sprintf("Status = \"%s\"", status))
	}

	// Add owner filter if provided
	if owner != "" {
		conditions = append(conditions, fmt.Sprintf("Owner = \"%s\"", owner))
	}

	// Add simple term search with regex-inspired patterns
	if term != "" {
		termCondition := buildTermSearchCondition(term)
		conditions = append(conditions, termCondition)
	}

	// If no conditions, return error
	if len(conditions) == 0 {
		return "", fmt.Errorf("no search conditions provided")
	}

	// Combine all conditions with AND
	query := conditions[0]
	for _, cond := range conditions[1:] {
		query += " AND " + cond
	}

	return query, nil
}

// buildTermSearchCondition creates AQL search condition with basic patterns
func buildTermSearchCondition(term string) string {
	var nameCondition, keyCondition string

	switch {
	case term == "*":
		// Wildcard - match all non-empty values
		return "(Name != \"\" OR Key != \"\")"

	case strings.HasPrefix(term, "^") && strings.HasSuffix(term, "$"):
		// Exact match: ^exact$ -> exact match
		exactTerm := strings.TrimSuffix(strings.TrimPrefix(term, "^"), "$")
		nameCondition = fmt.Sprintf("Name = \"%s\"", exactTerm)
		keyCondition = fmt.Sprintf("Key = \"%s\"", exactTerm)

	case strings.HasPrefix(term, "="):
		// Exact match: =value -> exact match
		exactTerm := strings.TrimPrefix(term, "=")
		nameCondition = fmt.Sprintf("Name = \"%s\"", exactTerm)
		keyCondition = fmt.Sprintf("Key = \"%s\"", exactTerm)

	default:
		// Default: exact match only (LIKE queries don't work in this AQL implementation)
		nameCondition = fmt.Sprintf("Name = \"%s\"", term)
		keyCondition = fmt.Sprintf("Key = \"%s\"", term)
	}

	// Return combined condition
	return fmt.Sprintf("(%s OR %s)", nameCondition, keyCondition)
}

// buildFilterSummary creates a summary of active filters
func buildFilterSummary(params common.SearchParams) map[string]string {
	filters := make(map[string]string)

	if params.Schema != "" {
		filters["schema"] = params.Schema
	}
	if params.Type != "" {
		filters["type"] = params.Type
	}
	if params.Status != "" {
		filters["status"] = params.Status
	}
	if params.Owner != "" {
		filters["owner"] = params.Owner
	}

	return filters
}