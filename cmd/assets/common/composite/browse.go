package composite

import (
	"fmt"

	"github.com/aaronsb/atlassian-assets/cmd/assets/common"
	"github.com/aaronsb/atlassian-assets/cmd/assets/common/foundation"
)

// BrowseSchema provides a comprehensive overview of a schema including structure and asset distribution
func BrowseSchema(client common.ClientInterface, params common.BrowseSchemaParams) (*common.Response, error) {
	// Validate parameters
	if params.SchemaID == "" {
		return common.NewErrorResponse(fmt.Errorf("schema ID is required")), nil
	}

	// Step 1: Get schema details
	schemaResponse, err := foundation.GetSchema(client, common.GetSchemaParams{
		SchemaID: params.SchemaID,
	})
	if err != nil {
		return common.NewErrorResponse(fmt.Errorf("failed to get schema details: %w", err)), nil
	}

	if !schemaResponse.Success {
		return schemaResponse, nil
	}

	// Step 2: Get object types in the schema
	objectTypesResponse, err := foundation.GetObjectTypes(client, params.SchemaID)
	if err != nil {
		return common.NewErrorResponse(fmt.Errorf("failed to get object types: %w", err)), nil
	}

	if !objectTypesResponse.Success {
		return objectTypesResponse, nil
	}

	// Step 3: Get a sample of objects to understand asset distribution
	sampleObjectsResponse, err := foundation.ListObjects(client, common.ListParams{
		Schema: params.SchemaID,
		Limit:  25, // Small sample for overview
		Offset: 0,
	})
	if err != nil {
		return common.NewErrorResponse(fmt.Errorf("failed to get sample objects: %w", err)), nil
	}

	// Extract data from responses
	schemaData := schemaResponse.Data.(map[string]interface{})
	objectTypesData := objectTypesResponse.Data.(map[string]interface{})
	
	var sampleObjects []interface{}
	var totalObjects int
	if sampleObjectsResponse.Success {
		sampleData := sampleObjectsResponse.Data.(map[string]interface{})
		if objects, ok := sampleData["objects"].([]interface{}); ok {
			sampleObjects = objects
		}
		if total, ok := sampleData["total"].(int); ok {
			totalObjects = total
		}
	}

	// Build comprehensive response
	browseData := map[string]interface{}{
		"schema":             schemaData["schema"],
		"schema_id":          params.SchemaID,
		"object_types":       objectTypesData["object_types"],
		"object_type_count":  objectTypesData["count"],
		"sample_objects":     sampleObjects,
		"total_objects":      totalObjects,
		"sample_size":        len(sampleObjects),
		"operation":          "browse_schema",
		"workflow_context": map[string]interface{}{
			"current_state":         "schema_explored",
			"completion_percentage": 75,
			"suggested_next_steps": []string{
				"search_specific_objects",
				"examine_object_types",
				"analyze_asset_distribution",
			},
		},
	}

	return common.NewSuccessResponse(browseData), nil
}

// BrowseObjectType provides detailed information about a specific object type
func BrowseObjectType(client common.ClientInterface, objectTypeID string) (*common.Response, error) {
	// Validate parameters
	if objectTypeID == "" {
		return common.NewErrorResponse(fmt.Errorf("object type ID is required")), nil
	}

	// Get object type attributes
	attributesResponse, err := foundation.GetObjectTypeAttributes(client, common.GetObjectTypeAttributesParams{
		ObjectTypeID: objectTypeID,
	})
	if err != nil {
		return common.NewErrorResponse(fmt.Errorf("failed to get object type attributes: %w", err)), nil
	}

	if !attributesResponse.Success {
		return attributesResponse, nil
	}

	// Add metadata for object type browsing
	responseData := attributesResponse.Data.(map[string]interface{})
	responseData["operation"] = "browse_object_type"
	responseData["workflow_context"] = map[string]interface{}{
		"current_state":         "object_type_analyzed",
		"completion_percentage": 50,
		"suggested_next_steps": []string{
			"create_object_instance",
			"validate_object_structure",
			"enhance_object_type",
		},
	}

	return common.NewSuccessResponse(responseData), nil
}