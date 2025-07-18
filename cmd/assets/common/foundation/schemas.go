package foundation

import (
	"context"
	"fmt"

	"github.com/aaronsb/atlassian-assets/cmd/assets/common"
)

// ListSchemas lists all available schemas
func ListSchemas(client common.ClientInterface) (*common.Response, error) {
	ctx := context.Background()
	response, err := client.ListSchemas(ctx)
	if err != nil {
		return common.NewErrorResponse(fmt.Errorf("failed to list schemas: %w", err)), nil
	}

	// Add metadata
	if response.Success {
		responseData := response.Data.(map[string]interface{})
		responseData["operation"] = "list_schemas"
		
		return common.NewSuccessResponse(responseData), nil
	}

	return common.NewErrorResponse(fmt.Errorf(response.Error)), nil
}

// GetSchema retrieves details of a specific schema
func GetSchema(client common.ClientInterface, params common.GetSchemaParams) (*common.Response, error) {
	// Validate parameters
	if params.SchemaID == "" {
		return common.NewErrorResponse(fmt.Errorf("schema ID is required")), nil
	}

	ctx := context.Background()
	response, err := client.GetSchema(ctx, params.SchemaID)
	if err != nil {
		return common.NewErrorResponse(fmt.Errorf("failed to get schema: %w", err)), nil
	}

	// Add metadata
	if response.Success {
		responseData := map[string]interface{}{
			"schema": response.Data,
			"schema_id": params.SchemaID,
			"operation": "get_schema",
		}
		return common.NewSuccessResponse(responseData), nil
	}

	return common.NewErrorResponse(fmt.Errorf(response.Error)), nil
}

// GetObjectTypes retrieves object types for a schema
func GetObjectTypes(client common.ClientInterface, schemaID string) (*common.Response, error) {
	// Validate parameters
	if schemaID == "" {
		return common.NewErrorResponse(fmt.Errorf("schema ID is required")), nil
	}

	ctx := context.Background()
	response, err := client.GetObjectTypes(ctx, schemaID)
	if err != nil {
		return common.NewErrorResponse(fmt.Errorf("failed to get object types: %w", err)), nil
	}

	// Add metadata
	if response.Success {
		responseData := response.Data.(map[string]interface{})
		responseData["schema_id"] = schemaID
		responseData["operation"] = "get_object_types"
		
		return common.NewSuccessResponse(responseData), nil
	}

	return common.NewErrorResponse(fmt.Errorf(response.Error)), nil
}

// CreateObjectType creates a new object type
func CreateObjectType(client common.ClientInterface, params common.CreateObjectTypeParams) (*common.Response, error) {
	// Validate parameters
	if params.Schema == "" {
		return common.NewErrorResponse(fmt.Errorf("schema is required")), nil
	}
	if params.Name == "" {
		return common.NewErrorResponse(fmt.Errorf("name is required")), nil
	}

	ctx := context.Background()
	response, err := client.CreateObjectType(ctx, params.Schema, params.Name, params.Description, params.Icon, params.Parent)
	if err != nil {
		return common.NewErrorResponse(fmt.Errorf("failed to create object type: %w", err)), nil
	}

	// Add metadata
	if response.Success {
		responseData := response.Data.(map[string]interface{})
		responseData["schema"] = params.Schema
		responseData["operation"] = "create_object_type"
		
		return common.NewSuccessResponse(responseData), nil
	}

	return common.NewErrorResponse(fmt.Errorf(response.Error)), nil
}

// GetObjectTypeAttributes retrieves attributes for an object type
func GetObjectTypeAttributes(client common.ClientInterface, params common.GetObjectTypeAttributesParams) (*common.Response, error) {
	// Validate parameters
	if params.ObjectTypeID == "" {
		return common.NewErrorResponse(fmt.Errorf("object type ID is required")), nil
	}

	ctx := context.Background()
	response, err := client.GetObjectTypeAttributes(ctx, params.ObjectTypeID)
	if err != nil {
		return common.NewErrorResponse(fmt.Errorf("failed to get object type attributes: %w", err)), nil
	}

	// Add metadata
	if response.Success {
		responseData := response.Data.(map[string]interface{})
		responseData["object_type_id"] = params.ObjectTypeID
		responseData["operation"] = "get_object_type_attributes"
		
		return common.NewSuccessResponse(responseData), nil
	}

	return common.NewErrorResponse(fmt.Errorf(response.Error)), nil
}