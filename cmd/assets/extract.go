package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ctreminiom/go-atlassian/v2/pkg/infra/models"
	"github.com/aaronsb/atlassian-assets/internal/client"
)

// EXTRACT command with subcommands for attribute extraction
var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract attributes from objects and object types",
	Long: `Universal attribute extraction tools for building attribute libraries.
	
Extract attributes from any source (object instances, object types, schemas)
and prepare them for universal application to other targets.`,
}

// EXTRACT ATTRIBUTES subcommand
var extractAttributesCmd = &cobra.Command{
	Use:   "attributes",
	Short: "Extract attributes from an object or object type",
	Long: `Extract all attributes with their values and references from an object instance,
or extract the attribute schema from an object type.
	
This creates a portable attribute set that can be applied to other objects.`,
	Example: `  # Extract attributes from object instance
  assets extract attributes --from-object 991
  
  # Extract attribute schema from object type
  assets extract attributes --from-object-type 65
  
  # Extract with reference resolution
  assets extract attributes --from-object 991 --resolve-references`,
	RunE: runExtractAttributesCmd,
}

var (
	extractFromObject     string
	extractFromObjectType string
	extractResolveRefs    bool
	extractIncludeSystem  bool
)

func init() {
	extractAttributesCmd.Flags().StringVar(&extractFromObject, "from-object", "", "Object ID to extract attributes from")
	extractAttributesCmd.Flags().StringVar(&extractFromObjectType, "from-object-type", "", "Object type ID to extract attribute schema from")
	extractAttributesCmd.Flags().BoolVar(&extractResolveRefs, "resolve-references", false, "Resolve reference targets")
	extractAttributesCmd.Flags().BoolVar(&extractIncludeSystem, "include-system", false, "Include system attributes (Created, Updated, Key)")
	
	extractAttributesCmd.MarkFlagsOneRequired("from-object", "from-object-type")
}

func runExtractAttributesCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	
	if extractFromObject != "" {
		return extractFromObjectInstance(ctx, client, extractFromObject)
	}
	
	if extractFromObjectType != "" {
		return extractFromObjectTypeSchema(ctx, client, extractFromObjectType)
	}
	
	return fmt.Errorf("no extraction source specified")
}

// extractFromObjectInstance extracts attributes with values from an object instance
func extractFromObjectInstance(ctx context.Context, client *client.AssetsClient, objectID string) error {
	// Get the object instance
	response, err := client.GetObject(ctx, objectID)
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("failed to get object: %s", response.Error)
	}

	// Handle the structured response from GetObject
	objectData := response.Data.(*models.ObjectScheme)
	
	// Extract object metadata
	objectInfo := map[string]interface{}{
		"object_id":   objectData.ID,
		"object_key":  objectData.ObjectKey,
		"label":       objectData.Label,
		"object_type": map[string]interface{}{
			"id":   objectData.ObjectType.ID,
			"name": objectData.ObjectType.Name,
		},
		"created": objectData.Created,
		"updated": objectData.Updated,
	}
	
	// Extract attributes
	var extractedAttributes []map[string]interface{}
	
	if objectData.Attributes != nil {
		for _, attr := range objectData.Attributes {
			attrInfo := extractAttributeFromInstanceStruct(attr)
			
			// Skip system attributes unless requested
			if !extractIncludeSystem && isSystemAttribute(attrInfo) {
				continue
			}
			
			// Resolve references if requested
			if extractResolveRefs && isReferenceAttribute(attrInfo) {
				resolvedRef, err := resolveAttributeReference(ctx, client, attrInfo)
				if err == nil {
					attrInfo["resolved_reference"] = resolvedRef
				} else {
					attrInfo["reference_error"] = err.Error()
				}
			}
			
			extractedAttributes = append(extractedAttributes, attrInfo)
		}
	}
	
	result := map[string]interface{}{
		"action":           "extract_attributes",
		"source_type":      "object_instance",
		"source_object":    objectInfo,
		"attribute_count":  len(extractedAttributes),
		"attributes":       extractedAttributes,
		"extraction_config": map[string]interface{}{
			"resolve_references": extractResolveRefs,
			"include_system":     extractIncludeSystem,
		},
	}
	
	successResponse := NewSuccessResponse(result)
	
	// Add contextual hints
	enhancedResponse := addNextStepHints(successResponse, "extract_attributes", map[string]interface{}{
		"source_type":      "object_instance",
		"success":          successResponse.Success,
		"has_references":   extractResolveRefs,
		"attribute_count":  len(extractedAttributes),
	})
	
	return outputResult(enhancedResponse)
}

// extractFromObjectTypeSchema extracts attribute schema from an object type
func extractFromObjectTypeSchema(ctx context.Context, client *client.AssetsClient, objectTypeID string) error {
	// Get object type attributes
	response, err := client.GetObjectTypeAttributes(ctx, objectTypeID)
	if err != nil {
		return fmt.Errorf("failed to get object type attributes: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("failed to get object type attributes: %s", response.Error)
	}

	attrData := response.Data.(map[string]interface{})
	attributesRaw := attrData["attributes"]
	
	var extractedAttributes []map[string]interface{}
	
	switch attrs := attributesRaw.(type) {
	case []*models.ObjectTypeAttributeScheme:
		for _, attr := range attrs {
			attrInfo := extractAttributeFromSchema(attr)
			
			// Skip system attributes unless requested
			if !extractIncludeSystem && attr.System {
				continue
			}
			
			// Resolve references if requested
			if extractResolveRefs && attr.Type == 1 && attr.ReferenceObjectTypeID != "" {
				resolvedRef, err := discoverObjectTypeInfo(ctx, client, attr.ReferenceObjectTypeID)
				if err == nil {
					attrInfo["resolved_reference"] = resolvedRef
				} else {
					attrInfo["reference_error"] = err.Error()
				}
			}
			
			extractedAttributes = append(extractedAttributes, attrInfo)
		}
	default:
		return fmt.Errorf("unexpected attributes type: %T", attributesRaw)
	}
	
	result := map[string]interface{}{
		"action":           "extract_attributes",
		"source_type":      "object_type_schema",
		"source_object_type": objectTypeID,
		"attribute_count":  len(extractedAttributes),
		"attributes":       extractedAttributes,
		"extraction_config": map[string]interface{}{
			"resolve_references": extractResolveRefs,
			"include_system":     extractIncludeSystem,
		},
	}
	
	successResponse := NewSuccessResponse(result)
	
	// Add contextual hints
	enhancedResponse := addNextStepHints(successResponse, "extract_attributes", map[string]interface{}{
		"source_type":      "object_type_schema",
		"success":          successResponse.Success,
		"has_references":   extractResolveRefs,
		"attribute_count":  len(extractedAttributes),
	})
	
	return outputResult(enhancedResponse)
}

// extractAttributeFromInstanceStruct extracts attribute info from a structured object instance attribute
func extractAttributeFromInstanceStruct(attr *models.ObjectAttributeScheme) map[string]interface{} {
	attrInfo := map[string]interface{}{
		"id":   attr.ID,
		"name": attr.ObjectTypeAttribute.Name,
		"system": attr.ObjectTypeAttribute.System,
		"editable": attr.ObjectTypeAttribute.Editable,
		"required": attr.ObjectTypeAttribute.MinimumCardinality > 0,
	}
	
	// Data type information
	if attr.ObjectTypeAttribute.DefaultType != nil {
		attrInfo["data_type"] = attr.ObjectTypeAttribute.DefaultType.Name
		if attr.ObjectTypeAttribute.DefaultType.ID != 0 {
			attrInfo["data_type_id"] = attr.ObjectTypeAttribute.DefaultType.ID
		}
	}
	
	// Check for reference type
	if attr.ObjectTypeAttribute.Type == 1 {
		attrInfo["is_reference"] = true
		attrInfo["attribute_type"] = 1
		if attr.ObjectTypeAttribute.ReferenceObjectTypeID != "" {
			attrInfo["reference_object_type_id"] = attr.ObjectTypeAttribute.ReferenceObjectTypeID
		}
	} else {
		attrInfo["attribute_type"] = attr.ObjectTypeAttribute.Type
	}
	
	// Extract values
	if len(attr.ObjectAttributeValues) > 0 {
		firstValue := attr.ObjectAttributeValues[0]
		attrInfo["value"] = firstValue.Value
		attrInfo["display_value"] = firstValue.DisplayValue
		
		// For multiple values
		if len(attr.ObjectAttributeValues) > 1 {
			var allValues []string
			for _, val := range attr.ObjectAttributeValues {
				allValues = append(allValues, val.Value)
			}
			attrInfo["all_values"] = allValues
			attrInfo["multiple_values"] = true
		}
	}
	
	return attrInfo
}


// extractAttributeFromSchema extracts attribute info from an object type schema
func extractAttributeFromSchema(attr *models.ObjectTypeAttributeScheme) map[string]interface{} {
	attrInfo := map[string]interface{}{
		"id":                        attr.ID,
		"name":                      attr.Name,
		"description":               attr.Description,
		"system":                    attr.System,
		"editable":                  attr.Editable,
		"required":                  attr.MinimumCardinality > 0,
		"minimum_cardinality":       attr.MinimumCardinality,
		"maximum_cardinality":       attr.MaximumCardinality,
		"attribute_type":            attr.Type,
		"summable":                  attr.Summable,
		"unique":                    attr.UniqueAttribute,
	}
	
	if attr.DefaultType != nil {
		attrInfo["data_type"] = attr.DefaultType.Name
		if attr.DefaultType.ID != 0 {
			attrInfo["data_type_id"] = attr.DefaultType.ID
		}
	}
	
	// Reference information
	if attr.Type == 1 && attr.ReferenceObjectTypeID != "" {
		attrInfo["is_reference"] = true
		attrInfo["reference_object_type_id"] = attr.ReferenceObjectTypeID
		if attr.ReferenceType != nil {
			attrInfo["reference_type"] = attr.ReferenceType.Name
		}
	}
	
	return attrInfo
}

// Helper functions
func isSystemAttribute(attrInfo map[string]interface{}) bool {
	if system, exists := attrInfo["system"]; exists && system != nil {
		return system.(bool)
	}
	return false
}

func isReferenceAttribute(attrInfo map[string]interface{}) bool {
	if isRef, exists := attrInfo["is_reference"]; exists {
		return isRef.(bool)
	}
	return false
}

func resolveAttributeReference(ctx context.Context, client *client.AssetsClient, attrInfo map[string]interface{}) (map[string]interface{}, error) {
	if refObjTypeID, exists := attrInfo["reference_object_type_id"]; exists {
		return discoverObjectTypeInfo(ctx, client, refObjTypeID.(string))
	}
	return nil, fmt.Errorf("no reference object type ID found")
}


func init() {
	extractCmd.AddCommand(extractAttributesCmd)
}