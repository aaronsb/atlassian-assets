package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ctreminiom/go-atlassian/v2/pkg/infra/models"
	"github.com/aaronsb/atlassian-assets/internal/client"
)

// APPLY command with subcommands for applying attributes
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply attributes to object types and objects",
	Long: `Universal attribute application system for the attribute marketplace.
	
Apply any extracted attribute set to any target object type or object instance,
with intelligent reference resolution and type mapping.`,
}

// APPLY ATTRIBUTES subcommand
var applyAttributesCmd = &cobra.Command{
	Use:   "attributes",
	Short: "Apply extracted attributes to an object type",
	Long: `Apply a set of extracted attributes to a target object type.
	
This enables the universal attribute marketplace where you can pick and choose
attributes from any source and apply them to any target with proper type mapping.`,
	Example: `  # Apply all attributes from a file
  assets apply attributes --to-object-type 142 --attributes-file laptop_attrs.json
  
  # Apply selected attributes only
  assets apply attributes --to-object-type 142 --attributes-file laptop_attrs.json --select "CPU,RAM,Cost"
  
  # Apply all except references
  assets apply attributes --to-object-type 142 --attributes-file laptop_attrs.json --skip-references
  
  # Apply with reference mapping
  assets apply attributes --to-object-type 142 --attributes-file laptop_attrs.json --map-references mapping.json`,
	RunE: runApplyAttributesCmd,
}

var (
	applyToObjectType   string
	applyAttributesFile string
	applySelect         string
	applySkipReferences bool
	applyMapReferences  string
	applyDryRun         bool
	applyForceOverwrite bool
)

func init() {
	applyAttributesCmd.Flags().StringVar(&applyToObjectType, "to-object-type", "", "Target object type ID (required)")
	applyAttributesCmd.Flags().StringVar(&applyAttributesFile, "attributes-file", "", "JSON file containing extracted attributes (required)")
	applyAttributesCmd.Flags().StringVar(&applySelect, "select", "", "Comma-separated list of attribute names to apply")
	applyAttributesCmd.Flags().BoolVar(&applySkipReferences, "skip-references", false, "Skip reference attributes")
	applyAttributesCmd.Flags().StringVar(&applyMapReferences, "map-references", "", "JSON file with reference mappings")
	applyAttributesCmd.Flags().BoolVar(&applyDryRun, "dry-run", false, "Show what would be applied without creating")
	applyAttributesCmd.Flags().BoolVar(&applyForceOverwrite, "force-overwrite", false, "Overwrite existing attributes with same name")
	
	applyAttributesCmd.MarkFlagRequired("to-object-type")
	applyAttributesCmd.MarkFlagRequired("attributes-file")
}

func runApplyAttributesCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	
	// Load extracted attributes from file
	attributesData, err := loadExtractedAttributes(applyAttributesFile)
	if err != nil {
		return fmt.Errorf("failed to load attributes file: %w", err)
	}
	
	// Load reference mappings if provided
	var referenceMappings map[string]string
	if applyMapReferences != "" {
		referenceMappings, err = loadReferenceMappings(applyMapReferences)
		if err != nil {
			return fmt.Errorf("failed to load reference mappings: %w", err)
		}
	}
	
	return applyAttributesToObjectType(ctx, client, applyToObjectType, attributesData, referenceMappings)
}

// loadExtractedAttributes loads attributes from an extracted attributes JSON file
func loadExtractedAttributes(filename string) (map[string]interface{}, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	var extractedData map[string]interface{}
	if err := json.Unmarshal(data, &extractedData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	
	return extractedData, nil
}

// loadReferenceMappings loads reference mappings from JSON file
func loadReferenceMappings(filename string) (map[string]string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read mapping file: %w", err)
	}
	
	var mappings map[string]string
	if err := json.Unmarshal(data, &mappings); err != nil {
		return nil, fmt.Errorf("failed to parse mapping JSON: %w", err)
	}
	
	return mappings, nil
}

// applyAttributesToObjectType applies the extracted attributes to a target object type
func applyAttributesToObjectType(ctx context.Context, client *client.AssetsClient, targetObjectTypeID string, attributesData map[string]interface{}, referenceMappings map[string]string) error {
	// Extract the attributes array from the loaded data
	successData, ok := attributesData["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid attributes file format: missing data section")
	}
	
	attributesArray, ok := successData["attributes"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid attributes file format: missing attributes array")
	}
	
	// Get existing attributes in target object type
	existingResponse, err := client.GetObjectTypeAttributes(ctx, targetObjectTypeID)
	if err != nil {
		return fmt.Errorf("failed to get target object type attributes: %w", err)
	}
	
	if !existingResponse.Success {
		return fmt.Errorf("failed to get target object type attributes: %s", existingResponse.Error)
	}
	
	// Build map of existing attribute names
	existingNames := make(map[string]bool)
	existingData := existingResponse.Data.(map[string]interface{})
	if existingAttrs, ok := existingData["attributes"]; ok {
		switch attrs := existingAttrs.(type) {
		case []*models.ObjectTypeAttributeScheme:
			for _, attr := range attrs {
				existingNames[attr.Name] = true
			}
		}
	}
	
	// Parse selection criteria
	var selectedNames map[string]bool
	if applySelect != "" {
		selectedNames = make(map[string]bool)
		for _, name := range strings.Split(applySelect, ",") {
			selectedNames[strings.TrimSpace(name)] = true
		}
	}
	
	// Process each attribute for application
	var attributesToApply []map[string]interface{}
	var skippedAttributes []map[string]interface{}
	var conflictingAttributes []map[string]interface{}
	
	for _, attrInterface := range attributesArray {
		attr := attrInterface.(map[string]interface{})
		attrName := attr["name"].(string)
		
		// Check selection criteria
		if selectedNames != nil && !selectedNames[attrName] {
			skippedAttributes = append(skippedAttributes, map[string]interface{}{
				"name":   attrName,
				"reason": "not_selected",
			})
			continue
		}
		
		// Check if should skip references
		if applySkipReferences && isReferenceFromExtracted(attr) {
			skippedAttributes = append(skippedAttributes, map[string]interface{}{
				"name":   attrName,
				"reason": "reference_skipped",
			})
			continue
		}
		
		// Check for conflicts with existing attributes
		if existingNames[attrName] && !applyForceOverwrite {
			conflictingAttributes = append(conflictingAttributes, map[string]interface{}{
				"name":   attrName,
				"reason": "already_exists",
			})
			continue
		}
		
		// Prepare attribute for application
		preparedAttr, err := prepareAttributeForApplication(attr, referenceMappings)
		if err != nil {
			skippedAttributes = append(skippedAttributes, map[string]interface{}{
				"name":   attrName,
				"reason": "preparation_failed",
				"error":  err.Error(),
			})
			continue
		}
		
		attributesToApply = append(attributesToApply, preparedAttr)
	}
	
	result := map[string]interface{}{
		"action":                "apply_attributes",
		"target_object_type":    targetObjectTypeID,
		"source_attribute_count": len(attributesArray),
		"planned_applications":  len(attributesToApply),
		"skipped_count":         len(skippedAttributes),
		"conflict_count":        len(conflictingAttributes),
		"dry_run":              applyDryRun,
	}
	
	if len(skippedAttributes) > 0 {
		result["skipped_attributes"] = skippedAttributes
	}
	
	if len(conflictingAttributes) > 0 {
		result["conflicting_attributes"] = conflictingAttributes
	}
	
	if applyDryRun {
		// Dry run - show what would be applied
		result["status"] = "dry_run_complete"
		result["message"] = fmt.Sprintf("Would apply %d attributes to object type %s", len(attributesToApply), targetObjectTypeID)
		
		if len(attributesToApply) > 0 {
			applyList := make([]string, 0)
			for _, attr := range attributesToApply {
				applyList = append(applyList, attr["name"].(string))
			}
			result["attributes_to_apply"] = applyList
		}
		
		return outputResult(NewSuccessResponse(result))
	}
	
	// Actual application
	var createdAttributes []string
	var failedAttributes []map[string]interface{}
	
	for _, attr := range attributesToApply {
		// Convert to ObjectTypeAttributePayloadScheme
		payload, err := convertToAttributePayload(attr)
		if err != nil {
			failedAttributes = append(failedAttributes, map[string]interface{}{
				"attribute": attr["name"],
				"error":     fmt.Sprintf("payload conversion failed: %v", err),
			})
			continue
		}
		
		// Create the attribute
		response, err := client.CreateObjectTypeAttribute(ctx, targetObjectTypeID, payload)
		if err != nil {
			failedAttributes = append(failedAttributes, map[string]interface{}{
				"attribute": attr["name"],
				"error":     err.Error(),
			})
			continue
		}
		
		if response.Success {
			createdAttributes = append(createdAttributes, attr["name"].(string))
		} else {
			failedAttributes = append(failedAttributes, map[string]interface{}{
				"attribute": attr["name"],
				"error":     response.Error,
			})
		}
	}
	
	result["status"] = "completed"
	result["created_count"] = len(createdAttributes)
	result["failed_count"] = len(failedAttributes)
	result["message"] = fmt.Sprintf("Successfully applied %d attributes to object type %s", len(createdAttributes), targetObjectTypeID)
	
	if len(createdAttributes) > 0 {
		result["created_attributes"] = createdAttributes
	}
	
	if len(failedAttributes) > 0 {
		result["failed_attributes"] = failedAttributes
	}
	
	return outputResult(NewSuccessResponse(result))
}

// isReferenceFromExtracted checks if an extracted attribute is a reference
func isReferenceFromExtracted(attr map[string]interface{}) bool {
	if isRef, exists := attr["is_reference"]; exists {
		return isRef.(bool)
	}
	if attrType, exists := attr["attribute_type"]; exists {
		return attrType.(float64) == 1
	}
	return false
}

// prepareAttributeForApplication prepares an extracted attribute for application
func prepareAttributeForApplication(attr map[string]interface{}, referenceMappings map[string]string) (map[string]interface{}, error) {
	prepared := make(map[string]interface{})
	
	// Copy basic fields
	for key, value := range attr {
		prepared[key] = value
	}
	
	// Handle reference mapping if it's a reference attribute
	if isReferenceFromExtracted(attr) {
		if referenceMappings != nil {
			if refObjTypeID, exists := attr["reference_object_type_id"]; exists {
				oldRefID := refObjTypeID.(string)
				if newRefID, mapped := referenceMappings[oldRefID]; mapped {
					prepared["reference_object_type_id"] = newRefID
					prepared["reference_mapped"] = true
				} else {
					return nil, fmt.Errorf("no mapping found for reference object type %s", oldRefID)
				}
			}
		} else if !applySkipReferences {
			return nil, fmt.Errorf("reference attribute requires mapping but no mappings provided")
		}
	}
	
	return prepared, nil
}

// convertToAttributePayload converts a prepared attribute to ObjectTypeAttributePayloadScheme
func convertToAttributePayload(attr map[string]interface{}) (*models.ObjectTypeAttributePayloadScheme, error) {
	payload := &models.ObjectTypeAttributePayloadScheme{
		Name:        attr["name"].(string),
		Description: getStringValue(attr, "description"),
	}
	
	// Set cardinalities if available
	if minCard, exists := attr["minimum_cardinality"]; exists {
		if val, ok := minCard.(float64); ok {
			intVal := int(val)
			payload.MinimumCardinality = &intVal
		}
	}
	
	if maxCard, exists := attr["maximum_cardinality"]; exists {
		if val, ok := maxCard.(float64); ok {
			intVal := int(val)
			payload.MaximumCardinality = &intVal
		}
	}
	
	// Handle data type
	if dataTypeID, exists := attr["data_type_id"]; exists {
		if val, ok := dataTypeID.(float64); ok {
			intVal := int(val)
			payload.DefaultTypeID = &intVal
		}
	} else if dataType, exists := attr["data_type"]; exists {
		// Map data type name to ID
		switch dataType.(string) {
		case "Text":
			textTypeID := 0
			payload.DefaultTypeID = &textTypeID
		case "Float":
			floatTypeID := 3
			payload.DefaultTypeID = &floatTypeID
		case "Integer":
			intTypeID := 1
			payload.DefaultTypeID = &intTypeID
		case "Boolean":
			boolTypeID := 2
			payload.DefaultTypeID = &boolTypeID
		case "DateTime":
			dateTypeID := 6
			payload.DefaultTypeID = &dateTypeID
		}
	}
	
	// Handle reference attributes
	if isReferenceFromExtracted(attr) {
		refType := 1
		payload.Type = &refType
		
		if refObjTypeID, exists := attr["reference_object_type_id"]; exists {
			payload.TypeValue = refObjTypeID.(string)
		}
	}
	
	// Handle other properties
	if summable, exists := attr["summable"]; exists {
		if val, ok := summable.(bool); ok {
			payload.Summable = val
		}
	}
	
	if unique, exists := attr["unique"]; exists {
		if val, ok := unique.(bool); ok {
			payload.UniqueAttribute = val
		}
	}
	
	return payload, nil
}

// getStringValue safely gets a string value from map
func getStringValue(attr map[string]interface{}, key string) string {
	if val, exists := attr[key]; exists && val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func init() {
	applyCmd.AddCommand(applyAttributesCmd)
}