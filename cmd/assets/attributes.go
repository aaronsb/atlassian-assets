package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ctreminiom/go-atlassian/v2/pkg/infra/models"
	"github.com/aaronsb/atlassian-assets/internal/client"
)

// COPY-ATTRIBUTES command
var copyAttributesCmd = &cobra.Command{
	Use:   "copy-attributes",
	Short: "Copy attributes from one object type to another",
	Long: `Copy attributes from a source object type to a destination object type.
	
This is useful for creating object types with similar attribute structures,
like copying laptop attributes to workstations since they're both computers.`,
	Example: `  # Copy laptop attributes to workstations
  assets copy-attributes --from 69 --to 141
  
  # Copy with confirmation prompt
  assets copy-attributes --from "Laptops" --to "Workstations" --dry-run`,
	RunE: runCopyAttributesCmd,
}

var (
	copyFromType string
	copyToType   string
	dryRun       bool
	skipExisting bool
)

func init() {
	copyAttributesCmd.Flags().StringVar(&copyFromType, "from", "", "Source object type ID or name (required)")
	copyAttributesCmd.Flags().StringVar(&copyToType, "to", "", "Destination object type ID or name (required)")
	copyAttributesCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be copied without making changes")
	copyAttributesCmd.Flags().BoolVar(&skipExisting, "skip-existing", true, "Skip attributes that already exist in destination")
	
	copyAttributesCmd.MarkFlagRequired("from")
	copyAttributesCmd.MarkFlagRequired("to")
}

func runCopyAttributesCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Get source attributes
	sourceResponse, err := client.GetObjectTypeAttributes(ctx, copyFromType)
	if err != nil {
		return fmt.Errorf("failed to get source attributes: %w", err)
	}

	if !sourceResponse.Success {
		return fmt.Errorf("failed to get source attributes: %s", sourceResponse.Error)
	}

	// Get destination attributes  
	destResponse, err := client.GetObjectTypeAttributes(ctx, copyToType)
	if err != nil {
		return fmt.Errorf("failed to get destination attributes: %w", err)
	}

	if !destResponse.Success {
		return fmt.Errorf("failed to get destination attributes: %s", destResponse.Error)
	}

	return copyAttributes(ctx, client, sourceResponse, destResponse, copyFromType, copyToType, dryRun, skipExisting)
}

// copyAttributes performs the actual attribute copying logic
func copyAttributes(ctx context.Context, client *client.AssetsClient, sourceResponse, destResponse *client.Response, sourceTypeID, destTypeID string, dryRun, skipExisting bool) error {
	sourceData := sourceResponse.Data.(map[string]interface{})
	destData := destResponse.Data.(map[string]interface{})
	
	// Parse source attributes
	sourceAttrs := sourceData["attributes"]
	var sourceAttributes []*models.ObjectTypeAttributeScheme
	switch attrs := sourceAttrs.(type) {
	case []*models.ObjectTypeAttributeScheme:
		sourceAttributes = attrs
	default:
		return fmt.Errorf("unexpected source attributes type: %T", sourceAttrs)
	}
	
	// Parse destination attributes to build existing names map
	destAttrs := destData["attributes"]
	var destAttributes []*models.ObjectTypeAttributeScheme
	switch attrs := destAttrs.(type) {
	case []*models.ObjectTypeAttributeScheme:
		destAttributes = attrs
	default:
		return fmt.Errorf("unexpected destination attributes type: %T", destAttrs)
	}
	
	// Build map of existing destination attribute names
	existingNames := make(map[string]bool)
	for _, attr := range destAttributes {
		existingNames[attr.Name] = true
	}
	
	// Plan which attributes to copy
	var attributesToCopy []*models.ObjectTypeAttributeScheme
	var skippedAttributes []string
	
	for _, attr := range sourceAttributes {
		// Skip system attributes (Created, Updated, Key)
		if attr.System {
			continue
		}
		
		// Skip if exists and skipExisting is true
		if existingNames[attr.Name] && skipExisting {
			skippedAttributes = append(skippedAttributes, attr.Name)
			continue
		}
		
		attributesToCopy = append(attributesToCopy, attr)
	}
	
	result := map[string]interface{}{
		"action":                "copy_attributes",
		"source_type":           sourceTypeID,
		"destination_type":      destTypeID,
		"source_count":          len(sourceAttributes),
		"destination_count":     len(destAttributes),
		"planned_copies":        len(attributesToCopy),
		"skipped_existing":      len(skippedAttributes),
		"dry_run":              dryRun,
	}
	
	if dryRun {
		// Dry run - show what would be copied
		result["status"] = "dry_run_complete"
		result["message"] = fmt.Sprintf("Would copy %d attributes from type %s to type %s", len(attributesToCopy), sourceTypeID, destTypeID)
		
		if len(attributesToCopy) > 0 {
			copyList := make([]string, 0)
			for _, attr := range attributesToCopy {
				copyList = append(copyList, attr.Name)
			}
			result["attributes_to_copy"] = copyList
		}
		
		if len(skippedAttributes) > 0 {
			result["skipped_attributes"] = skippedAttributes
		}
		
		return outputResult(NewSuccessResponse(result))
	}
	
	// Actual copying
	var createdAttributes []string
	var failedAttributes []map[string]interface{}
	
	for _, sourceAttr := range attributesToCopy {
		// Convert source attribute to payload for destination
		payload := &models.ObjectTypeAttributePayloadScheme{
			Name:                sourceAttr.Name,
			Description:         sourceAttr.Description,
			MinimumCardinality:  &sourceAttr.MinimumCardinality,
			MaximumCardinality:  &sourceAttr.MaximumCardinality,
		}
		
		// Handle reference attributes (Type 1 = Reference)
		if sourceAttr.Type == 1 && sourceAttr.ReferenceObjectTypeID != "" {
			payload.Type = &sourceAttr.Type
			payload.TypeValue = sourceAttr.ReferenceObjectTypeID
		}
		
		// Map the default type - this is critical for attribute creation
		if sourceAttr.DefaultType != nil {
			if sourceAttr.DefaultType.ID != 0 {
				payload.DefaultTypeID = &sourceAttr.DefaultType.ID
			} else {
				// If ID is 0, try to map by name to common type IDs
				switch sourceAttr.DefaultType.Name {
				case "Text":
					textTypeID := 0  // Default text type
					payload.DefaultTypeID = &textTypeID
				case "Float":
					floatTypeID := 3
					payload.DefaultTypeID = &floatTypeID
				case "Boolean":
					boolTypeID := 2
					payload.DefaultTypeID = &boolTypeID
				case "DateTime":
					dateTypeID := 6
					payload.DefaultTypeID = &dateTypeID
				case "URL":
					urlTypeID := 7
					payload.DefaultTypeID = &urlTypeID
				case "Textarea":
					textareaTypeID := 9
					payload.DefaultTypeID = &textareaTypeID
				case "Select":
					selectTypeID := 10
					payload.DefaultTypeID = &selectTypeID
				}
			}
		}
		
		// Copy other properties as available
		if sourceAttr.Summable {
			payload.Summable = sourceAttr.Summable
		}
		if sourceAttr.UniqueAttribute {
			payload.UniqueAttribute = sourceAttr.UniqueAttribute
		}
		
		// Create the attribute on destination object type
		response, err := client.CreateObjectTypeAttribute(ctx, destTypeID, payload)
		if err != nil {
			failedAttributes = append(failedAttributes, map[string]interface{}{
				"attribute": sourceAttr.Name,
				"error":     err.Error(),
			})
			continue
		}
		
		if response.Success {
			createdAttributes = append(createdAttributes, sourceAttr.Name)
		} else {
			failedAttributes = append(failedAttributes, map[string]interface{}{
				"attribute": sourceAttr.Name,
				"error":     response.Error,
			})
		}
	}
	
	result["status"] = "completed"
	result["created_count"] = len(createdAttributes)
	result["failed_count"] = len(failedAttributes)
	result["message"] = fmt.Sprintf("Successfully copied %d attributes from type %s to type %s", len(createdAttributes), sourceTypeID, destTypeID)
	
	if len(createdAttributes) > 0 {
		result["created_attributes"] = createdAttributes
	}
	
	if len(failedAttributes) > 0 {
		result["failed_attributes"] = failedAttributes
	}
	
	return outputResult(NewSuccessResponse(result))
}