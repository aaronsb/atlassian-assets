package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ctreminiom/go-atlassian/v2/pkg/infra/models"
	"github.com/aaronsb/atlassian-assets/internal/client"
)

// TRACE command with subcommands for reference discovery
var traceCmd = &cobra.Command{
	Use:   "trace",
	Short: "Trace and discover references across schemas",
	Long: `Tools for discovering where references point across the entire workspace.
	
Maintains schema integrity by following actual reference chains instead of
degrading to text fields. Essential for proper cross-schema attribute copying.`,
}

// TRACE REFERENCE subcommand
var traceReferenceCmd = &cobra.Command{
	Use:   "reference",
	Short: "Trace where a reference attribute points",
	Long: `Follow a reference attribute back to its source object type.
	
This discovers the target schema and object type for reference attributes,
enabling proper cross-schema reference resolution.`,
	Example: `  # Trace where Manufacturer attribute points
  assets trace reference --attribute-id 697 --source-schema 7
  
  # Trace reference by attribute name and object type
  assets trace reference --attribute-name "Manufacturer" --object-type 65`,
	RunE: runTraceReferenceCmd,
}

var (
	traceAttributeID   string
	traceAttributeName string
	traceObjectType    string
	traceSourceSchema  string
)

func init() {
	traceReferenceCmd.Flags().StringVar(&traceAttributeID, "attribute-id", "", "Attribute ID to trace")
	traceReferenceCmd.Flags().StringVar(&traceAttributeName, "attribute-name", "", "Attribute name to trace")
	traceReferenceCmd.Flags().StringVar(&traceObjectType, "object-type", "", "Object type ID containing the attribute")
	traceReferenceCmd.Flags().StringVar(&traceSourceSchema, "source-schema", "", "Source schema ID")
	
	// Either attribute-id OR (attribute-name + object-type) is required
	traceReferenceCmd.MarkFlagsOneRequired("attribute-id", "attribute-name")
	traceReferenceCmd.MarkFlagsRequiredTogether("attribute-name", "object-type")
}

func runTraceReferenceCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	
	// If we have attribute name but not ID, resolve it first
	if traceAttributeID == "" && traceAttributeName != "" {
		attributeID, err := resolveAttributeNameToID(ctx, client, traceObjectType, traceAttributeName)
		if err != nil {
			return fmt.Errorf("failed to resolve attribute name to ID: %w", err)
		}
		traceAttributeID = attributeID
	}
	
	return traceReference(ctx, client, traceAttributeID, traceSourceSchema)
}

// resolveAttributeNameToID converts an attribute name to its ID within an object type
func resolveAttributeNameToID(ctx context.Context, client *client.AssetsClient, objectTypeID, attributeName string) (string, error) {
	response, err := client.GetObjectTypeAttributes(ctx, objectTypeID)
	if err != nil {
		return "", fmt.Errorf("failed to get object type attributes: %w", err)
	}

	if !response.Success {
		return "", fmt.Errorf("failed to get object type attributes: %s", response.Error)
	}

	attrData := response.Data.(map[string]interface{})
	attributesRaw := attrData["attributes"]
	
	switch attrs := attributesRaw.(type) {
	case []*models.ObjectTypeAttributeScheme:
		for _, attr := range attrs {
			if attr.Name == attributeName {
				return attr.ID, nil
			}
		}
	default:
		return "", fmt.Errorf("unexpected attributes type: %T", attributesRaw)
	}
	
	return "", fmt.Errorf("attribute '%s' not found in object type %s", attributeName, objectTypeID)
}

// traceReference follows a reference attribute to discover its target
func traceReference(ctx context.Context, client *client.AssetsClient, attributeID, sourceSchema string) error {
	// We need to find the attribute across object types in the schema
	// Let's search through all object types to find this attribute
	
	if sourceSchema == "" {
		// Try to discover schema from attribute ID - for now assume schema 7
		sourceSchema = "7"
	}
	
	// Get all object types in the schema
	response, err := client.GetObjectTypes(ctx, sourceSchema)
	if err != nil {
		return fmt.Errorf("failed to get object types: %w", err)
	}
	
	if !response.Success {
		return fmt.Errorf("failed to get object types: %s", response.Error)
	}
	
	data := response.Data.(map[string]interface{})
	objectTypesData := data["object_types"]
	
	var objectTypes []map[string]interface{}
	jsonBytes, _ := json.Marshal(objectTypesData)
	json.Unmarshal(jsonBytes, &objectTypes)
	
	// Search each object type for this attribute
	for _, objType := range objectTypes {
		objTypeID := objType["id"].(string)
		
		attrResponse, err := client.GetObjectTypeAttributes(ctx, objTypeID)
		if err != nil {
			continue // Skip failed lookups
		}
		
		if !attrResponse.Success {
			continue
		}
		
		// Check if this object type contains our attribute
		attrData := attrResponse.Data.(map[string]interface{})
		attributesRaw := attrData["attributes"]
		
		switch attrs := attributesRaw.(type) {
		case []*models.ObjectTypeAttributeScheme:
			for _, attr := range attrs {
				if attr.ID == attributeID {
					// Found the attribute! Now trace its reference
					result := map[string]interface{}{
						"action":             "trace_reference",
						"attribute_id":       attributeID,
						"attribute_name":     attr.Name,
						"source_schema":      sourceSchema,
						"source_object_type": objTypeID,
						"source_object_name": objType["name"],
					}
					
					// Check if it's a reference attribute
					if attr.Type == 1 && attr.ReferenceObjectTypeID != "" {
						// It's a reference! Discover the target
						targetResponse, err := discoverObjectTypeInfo(ctx, client, attr.ReferenceObjectTypeID)
						if err != nil {
							result["error"] = fmt.Sprintf("Failed to discover target: %v", err)
						} else {
							result["reference_target"] = targetResponse
							result["status"] = "reference_resolved"
							result["message"] = fmt.Sprintf("Reference attribute '%s' points to object type %s", attr.Name, attr.ReferenceObjectTypeID)
						}
					} else {
						result["status"] = "not_reference"
						result["message"] = fmt.Sprintf("Attribute '%s' is not a reference (type %d)", attr.Name, attr.Type)
						result["attribute_type"] = attr.Type
						if attr.DefaultType != nil {
							result["data_type"] = attr.DefaultType.Name
						}
					}
					
					return outputResult(NewSuccessResponse(result))
				}
			}
		}
	}
	
	return outputResult(NewErrorResponse(fmt.Errorf("attribute ID %s not found in schema %s", attributeID, sourceSchema)))
}

// discoverObjectTypeInfo discovers information about an object type, potentially across schemas
func discoverObjectTypeInfo(ctx context.Context, client *client.AssetsClient, objectTypeID string) (map[string]interface{}, error) {
	// For now, try to get object type attributes to see if we can access it
	// This will fail if it's in a different schema, but that's valuable information
	
	response, err := client.GetObjectTypeAttributes(ctx, objectTypeID)
	if err != nil {
		return map[string]interface{}{
			"object_type_id": objectTypeID,
			"status":         "cross_schema_reference",
			"error":          err.Error(),
			"message":        "Object type exists in different schema - cross-schema reference detected",
		}, nil
	}
	
	if !response.Success {
		return map[string]interface{}{
			"object_type_id": objectTypeID,
			"status":         "access_denied",
			"error":          response.Error,
		}, nil
	}
	
	// Successfully accessed - same schema
	attrData := response.Data.(map[string]interface{})
	
	return map[string]interface{}{
		"object_type_id":    objectTypeID,
		"status":           "same_schema",
		"attribute_count":  attrData["count"],
		"message":          "Reference target found in same schema",
	}, nil
}

// TRACE DEPENDENCIES subcommand
var traceDependenciesCmd = &cobra.Command{
	Use:   "dependencies",
	Short: "Discover all dependencies for an object type",
	Long: `Recursively discover all reference dependencies for copying an object type.
	
This analyzes all attributes to find cross-schema references and builds
a complete dependency tree for proper reference-aware copying.`,
	Example: `  # Find all dependencies for laptop object type
  assets trace dependencies --object-type 65 --schema 7
  
  # Show dependencies across all schemas
  assets trace dependencies --object-type 65 --all-schemas`,
	RunE: runTraceDependenciesCmd,
}

var (
	depObjectType string
	depSchema     string
	depAllSchemas bool
)

func init() {
	traceDependenciesCmd.Flags().StringVar(&depObjectType, "object-type", "", "Object type ID to analyze (required)")
	traceDependenciesCmd.Flags().StringVar(&depSchema, "schema", "", "Schema ID (required unless --all-schemas)")
	traceDependenciesCmd.Flags().BoolVar(&depAllSchemas, "all-schemas", false, "Search across all schemas for references")
	
	traceDependenciesCmd.MarkFlagRequired("object-type")
	traceDependenciesCmd.MarkFlagsOneRequired("schema", "all-schemas")
}

func runTraceDependenciesCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	
	if depAllSchemas {
		return traceDependenciesAllSchemas(ctx, client, depObjectType)
	}
	
	return traceDependenciesInSchema(ctx, client, depObjectType, depSchema)
}

// traceDependenciesInSchema analyzes dependencies within a specific schema
func traceDependenciesInSchema(ctx context.Context, client *client.AssetsClient, objectTypeID, schemaID string) error {
	// Get all attributes for this object type
	response, err := client.GetObjectTypeAttributes(ctx, objectTypeID)
	if err != nil {
		return fmt.Errorf("failed to get object type attributes: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("failed to get object type attributes: %s", response.Error)
	}

	attrData := response.Data.(map[string]interface{})
	attributesRaw := attrData["attributes"]
	
	var dependencies []map[string]interface{}
	var nonReferences []map[string]interface{}
	
	switch attrs := attributesRaw.(type) {
	case []*models.ObjectTypeAttributeScheme:
		for _, attr := range attrs {
			if attr.System {
				continue // Skip system attributes
			}
			
			attrInfo := map[string]interface{}{
				"name":        attr.Name,
				"id":          attr.ID,
				"type":        attr.Type,
				"required":    attr.MinimumCardinality > 0,
			}
			
			if attr.DefaultType != nil {
				attrInfo["data_type"] = attr.DefaultType.Name
			}
			
			// Check if it's a reference
			if attr.Type == 1 && attr.ReferenceObjectTypeID != "" {
				// It's a reference! Discover the target
				targetInfo, err := discoverObjectTypeInfo(ctx, client, attr.ReferenceObjectTypeID)
				if err == nil {
					attrInfo["reference_target"] = targetInfo
					attrInfo["reference_object_type_id"] = attr.ReferenceObjectTypeID
					dependencies = append(dependencies, attrInfo)
				} else {
					attrInfo["reference_error"] = err.Error()
					dependencies = append(dependencies, attrInfo)
				}
			} else {
				nonReferences = append(nonReferences, attrInfo)
			}
		}
	default:
		return fmt.Errorf("unexpected attributes type: %T", attributesRaw)
	}
	
	result := map[string]interface{}{
		"action":              "trace_dependencies",
		"object_type_id":      objectTypeID,
		"schema_id":          schemaID,
		"total_attributes":   len(dependencies) + len(nonReferences),
		"reference_count":    len(dependencies),
		"non_reference_count": len(nonReferences),
	}
	
	if len(dependencies) > 0 {
		result["dependencies"] = dependencies
		result["status"] = "dependencies_found"
		result["message"] = fmt.Sprintf("Found %d reference dependencies for object type %s", len(dependencies), objectTypeID)
	} else {
		result["status"] = "no_dependencies"
		result["message"] = fmt.Sprintf("Object type %s has no reference dependencies", objectTypeID)
	}
	
	if len(nonReferences) > 0 {
		result["simple_attributes"] = nonReferences
	}
	
	return outputResult(NewSuccessResponse(result))
}

// traceDependenciesAllSchemas discovers dependencies across all schemas
func traceDependenciesAllSchemas(ctx context.Context, client *client.AssetsClient, objectTypeID string) error {
	// First get all schemas
	schemasResponse, err := client.ListSchemas(ctx)
	if err != nil {
		return fmt.Errorf("failed to list schemas: %w", err)
	}
	
	if !schemasResponse.Success {
		return fmt.Errorf("failed to list schemas: %s", schemasResponse.Error)
	}
	
	schemasData := schemasResponse.Data.(map[string]interface{})
	schemas := schemasData["schemas"]
	
	result := map[string]interface{}{
		"action": "trace_dependencies_all_schemas",
		"object_type_id": objectTypeID,
		"workspace_schemas": schemas,
		"status": "cross_schema_analysis",
		"message": "Cross-schema dependency analysis requires implementing schema discovery workflow",
		"next_steps": []string{
			"Implement findObjectTypeInSchemas function",
			"Add cross-schema reference resolution",
			"Build dependency tree visualization",
		},
	}
	
	return outputResult(NewSuccessResponse(result))
}

func init() {
	traceCmd.AddCommand(traceReferenceCmd)
	traceCmd.AddCommand(traceDependenciesCmd)
}