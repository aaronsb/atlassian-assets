package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// REMOVE command with subcommands - for removing attributes, relationships, etc.
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove attributes, relationships, and properties from assets",
	Long: `Remove specific attributes, relationships, and properties from existing assets.
	
This command modifies existing entities by removing components rather than deleting
the entire entity. For permanent deletion of object types and instances, use 'delete'.

Provides safe removal workflows with validation and contextual hints.`,
	Example: `  # Remove an attribute from an object type
  assets remove attribute --type-id 123 --attribute-id 456
  
  # Remove a relationship from an object
  assets remove relationship --object-id 789 --relationship-id 101
  
  # Remove a property value from an object
  assets remove property --object-id 789 --property-name "Location"`,
}

// REMOVE ATTRIBUTE subcommand
var removeAttributeCmd = &cobra.Command{
	Use:   "attribute",
	Short: "Remove an attribute from an object type",
	Long: `Remove an attribute definition from an object type.
	
This will remove the attribute from the object type schema but will not delete
the object type itself. Existing objects will lose this attribute data.`,
	Example: `  # Remove attribute by ID
  assets remove attribute --type-id 123 --attribute-id 456
  
  # Remove attribute by name
  assets remove attribute --type-id 123 --attribute-name "Old Field"
  
  # Remove with confirmation
  assets remove attribute --type-id 123 --attribute-id 456 --confirm`,
	RunE: runRemoveAttributeCmd,
}

// REMOVE RELATIONSHIP subcommand
var removeRelationshipCmd = &cobra.Command{
	Use:   "relationship",
	Short: "Remove a relationship from an object",
	Long: `Remove a relationship connection between objects.
	
This breaks the relationship link but does not delete the related objects.`,
	Example: `  # Remove relationship by ID
  assets remove relationship --object-id 789 --relationship-id 101
  
  # Remove relationship by type and target
  assets remove relationship --object-id 789 --relationship-type "connects_to" --target-id 202`,
	RunE: runRemoveRelationshipCmd,
}

// REMOVE PROPERTY subcommand
var removePropertyCmd = &cobra.Command{
	Use:   "property",
	Short: "Remove a property value from an object",
	Long: `Remove a specific property value from an object instance.
	
This clears the property value but keeps the property definition in the object type.`,
	Example: `  # Remove property by name
  assets remove property --object-id 789 --property-name "Location"
  
  # Remove property by ID
  assets remove property --object-id 789 --property-id 456
  
  # Remove multiple properties
  assets remove property --object-id 789 --property-name "Location,Status"`,
	RunE: runRemovePropertyCmd,
}

var (
	// Common flags
	removeObjectID     string
	removeTypeID       string
	removeConfirm      bool
	removeForce        bool
	
	// Attribute flags
	removeAttributeID   string
	removeAttributeName string
	
	// Relationship flags
	removeRelationshipID   string
	removeRelationshipType string
	removeTargetID         string
	
	// Property flags
	removePropertyID   string
	removePropertyName string
)

func init() {
	// Attribute flags
	removeAttributeCmd.Flags().StringVar(&removeTypeID, "type-id", "", "Object type ID to remove attribute from")
	removeAttributeCmd.Flags().StringVar(&removeAttributeID, "attribute-id", "", "Attribute ID to remove")
	removeAttributeCmd.Flags().StringVar(&removeAttributeName, "attribute-name", "", "Attribute name to remove")
	removeAttributeCmd.Flags().BoolVar(&removeConfirm, "confirm", false, "Confirm removal")
	removeAttributeCmd.Flags().BoolVar(&removeForce, "force", false, "Force removal without confirmation")
	removeAttributeCmd.MarkFlagRequired("type-id")
	
	// Relationship flags
	removeRelationshipCmd.Flags().StringVar(&removeObjectID, "object-id", "", "Object ID to remove relationship from")
	removeRelationshipCmd.Flags().StringVar(&removeRelationshipID, "relationship-id", "", "Relationship ID to remove")
	removeRelationshipCmd.Flags().StringVar(&removeRelationshipType, "relationship-type", "", "Relationship type to remove")
	removeRelationshipCmd.Flags().StringVar(&removeTargetID, "target-id", "", "Target object ID for relationship removal")
	removeRelationshipCmd.Flags().BoolVar(&removeConfirm, "confirm", false, "Confirm removal")
	removeRelationshipCmd.Flags().BoolVar(&removeForce, "force", false, "Force removal without confirmation")
	removeRelationshipCmd.MarkFlagRequired("object-id")
	
	// Property flags
	removePropertyCmd.Flags().StringVar(&removeObjectID, "object-id", "", "Object ID to remove property from")
	removePropertyCmd.Flags().StringVar(&removePropertyID, "property-id", "", "Property ID to remove")
	removePropertyCmd.Flags().StringVar(&removePropertyName, "property-name", "", "Property name(s) to remove (comma-separated)")
	removePropertyCmd.Flags().BoolVar(&removeConfirm, "confirm", false, "Confirm removal")
	removePropertyCmd.Flags().BoolVar(&removeForce, "force", false, "Force removal without confirmation")
	removePropertyCmd.MarkFlagRequired("object-id")
	
	// Add subcommands
	removeCmd.AddCommand(removeAttributeCmd)
	removeCmd.AddCommand(removeRelationshipCmd)
	removeCmd.AddCommand(removePropertyCmd)
}

func runRemoveAttributeCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	
	// Validate input
	if removeAttributeID == "" && removeAttributeName == "" {
		return fmt.Errorf("must specify either --attribute-id or --attribute-name")
	}
	
	// Safety confirmation
	if !removeForce && !removeConfirm {
		return fmt.Errorf("attribute removal requires --confirm or --force flag for safety")
	}
	
	// Resolve attribute ID if needed
	var attributeID string
	if removeAttributeID != "" {
		attributeID = removeAttributeID
	} else {
		// TODO: Implement name-to-ID resolution
		return fmt.Errorf("removal by attribute name not yet implemented")
	}
	
	// Perform removal
	response, err := client.RemoveAttribute(ctx, removeTypeID, attributeID)
	if err != nil {
		return fmt.Errorf("failed to remove attribute: %w", err)
	}
	
	if !response.Success {
		return fmt.Errorf("remove operation failed: %v", response.Error)
	}
	
	result := map[string]interface{}{
		"action":       "remove_attribute",
		"type_id":      removeTypeID,
		"attribute_id": attributeID,
		"force":        removeForce,
		"confirm":      removeConfirm,
		"removed":      true,
	}
	
	// Add contextual hints
	enhancedResponse := addNextStepHints(NewSuccessResponse(result), "remove_attribute", map[string]interface{}{
		"type_id":      removeTypeID,
		"attribute_id": attributeID,
		"success":      true,
	})
	
	return outputResult(enhancedResponse)
}

func runRemoveRelationshipCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	
	// Validate input
	if removeRelationshipID == "" && (removeRelationshipType == "" || removeTargetID == "") {
		return fmt.Errorf("must specify either --relationship-id or both --relationship-type and --target-id")
	}
	
	// Safety confirmation
	if !removeForce && !removeConfirm {
		return fmt.Errorf("relationship removal requires --confirm or --force flag for safety")
	}
	
	// Get client instance
	assetsClient, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer assetsClient.Close()
	
	// Perform removal - both methods return not implemented error for now
	var removeErr error
	
	if removeRelationshipID != "" {
		_, removeErr = assetsClient.RemoveRelationship(ctx, removeObjectID, removeRelationshipID)
	} else {
		_, removeErr = assetsClient.RemoveRelationshipByType(ctx, removeObjectID, removeRelationshipType, removeTargetID)
	}
	
	if removeErr != nil {
		return fmt.Errorf("failed to remove relationship: %w", removeErr)
	}
	
	result := map[string]interface{}{
		"action":            "remove_relationship",
		"object_id":         removeObjectID,
		"relationship_id":   removeRelationshipID,
		"relationship_type": removeRelationshipType,
		"target_id":         removeTargetID,
		"force":             removeForce,
		"confirm":           removeConfirm,
		"removed":           true,
	}
	
	// Add contextual hints
	enhancedResponse := addNextStepHints(NewSuccessResponse(result), "remove_relationship", map[string]interface{}{
		"object_id":         removeObjectID,
		"relationship_type": removeRelationshipType,
		"success":           true,
	})
	
	return outputResult(enhancedResponse)
}

func runRemovePropertyCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	
	// Validate input
	if removePropertyID == "" && removePropertyName == "" {
		return fmt.Errorf("must specify either --property-id or --property-name")
	}
	
	// Safety confirmation
	if !removeForce && !removeConfirm {
		return fmt.Errorf("property removal requires --confirm or --force flag for safety")
	}
	
	var removedProperties []string
	var errors []string
	
	if removePropertyID != "" {
		// Remove single property by ID
		response, err := client.RemoveProperty(ctx, removeObjectID, removePropertyID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Property %s: %v", removePropertyID, err))
		} else if !response.Success {
			errors = append(errors, fmt.Sprintf("Property %s: %v", removePropertyID, response.Error))
		} else {
			removedProperties = append(removedProperties, removePropertyID)
		}
	} else {
		// Remove properties by name (comma-separated)
		propertyNames := strings.Split(removePropertyName, ",")
		for _, name := range propertyNames {
			name = strings.TrimSpace(name)
			response, err := client.RemovePropertyByName(ctx, removeObjectID, name)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Property %s: %v", name, err))
			} else if !response.Success {
				errors = append(errors, fmt.Sprintf("Property %s: %v", name, response.Error))
			} else {
				removedProperties = append(removedProperties, name)
			}
		}
	}
	
	result := map[string]interface{}{
		"action":             "remove_property",
		"object_id":          removeObjectID,
		"property_id":        removePropertyID,
		"property_name":      removePropertyName,
		"removed_properties": removedProperties,
		"removed_count":      len(removedProperties),
		"force":              removeForce,
		"confirm":            removeConfirm,
	}
	
	if len(errors) > 0 {
		result["errors"] = errors
		result["success"] = false
	} else {
		result["success"] = true
	}
	
	// Add contextual hints
	enhancedResponse := addNextStepHints(NewSuccessResponse(result), "remove_property", map[string]interface{}{
		"object_id":      removeObjectID,
		"removed_count":  len(removedProperties),
		"success":        len(errors) == 0,
		"has_errors":     len(errors) > 0,
	})
	
	return outputResult(enhancedResponse)
}