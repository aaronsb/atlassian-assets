package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// DELETE command with subcommands - for permanent deletion of entities
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete assets (object types and instances)",
	Long: `Delete asset object types and instances from Atlassian Assets.
	
This command performs permanent deletion of entities like object types and instances.
For removing attributes, relationships, or other modifications, use the 'remove' command.

Provides safe deletion workflows with confirmation prompts and contextual hints
for cleanup operations.`,
	Example: `  # Delete an object type (and all its instances)
  assets delete object-type --id 123
  
  # Delete an object instance
  assets delete instance --id 456
  
  # Delete multiple instances with confirmation
  assets delete instance --id 456,789 --confirm`,
}

// DELETE OBJECT-TYPE subcommand
var deleteObjectTypeCmd = &cobra.Command{
	Use:   "object-type",
	Short: "Delete an object type",
	Long: `Delete an object type from a schema.
	
WARNING: This will also delete all instances of this object type.
Use with caution as this operation cannot be undone.`,
	Example: `  # Delete object type by ID
  assets delete object-type --id 123
  
  # Delete object type by name within schema
  assets delete object-type --name "Old Servers" --schema 6
  
  # Force delete without confirmation
  assets delete object-type --id 123 --force`,
	RunE: runDeleteObjectTypeCmd,
}

// DELETE INSTANCE subcommand
var deleteInstanceCmd = &cobra.Command{
	Use:   "instance",
	Short: "Delete object instances",
	Long: `Delete one or more object instances.
	
Can delete single instances or multiple instances at once.
Provides confirmation prompts for safety.`,
	Example: `  # Delete single instance
  assets delete instance --id 456
  
  # Delete multiple instances
  assets delete instance --id 456,789
  
  # Delete instances by AQL query
  assets delete instance --query "Name like 'temp%'"
  
  # Force delete without confirmation
  assets delete instance --id 456 --force`,
	RunE: runDeleteInstanceCmd,
}

var (
	// Common flags
	deleteID      string
	deleteForce   bool
	deleteConfirm bool
	
	// Object type specific flags
	deleteObjectTypeName   string
	deleteObjectTypeSchema string
	
	// Instance specific flags
	deleteInstanceQuery string
	deleteInstanceLimit int
)

func init() {
	// Common flags
	deleteObjectTypeCmd.Flags().StringVar(&deleteID, "id", "", "Object type ID to delete")
	deleteObjectTypeCmd.Flags().BoolVar(&deleteForce, "force", false, "Force deletion without confirmation")
	deleteObjectTypeCmd.Flags().BoolVar(&deleteConfirm, "confirm", false, "Confirm deletion")
	
	// Object type specific flags
	deleteObjectTypeCmd.Flags().StringVar(&deleteObjectTypeName, "name", "", "Object type name to delete")
	deleteObjectTypeCmd.Flags().StringVar(&deleteObjectTypeSchema, "schema", "", "Schema ID when deleting by name")
	
	// Instance flags
	deleteInstanceCmd.Flags().StringVar(&deleteID, "id", "", "Instance ID(s) to delete (comma-separated)")
	deleteInstanceCmd.Flags().StringVar(&deleteInstanceQuery, "query", "", "AQL query to select instances for deletion")
	deleteInstanceCmd.Flags().IntVar(&deleteInstanceLimit, "limit", 10, "Maximum number of instances to delete with query")
	deleteInstanceCmd.Flags().BoolVar(&deleteForce, "force", false, "Force deletion without confirmation")
	deleteInstanceCmd.Flags().BoolVar(&deleteConfirm, "confirm", false, "Confirm deletion")
	
	// Add subcommands
	deleteCmd.AddCommand(deleteObjectTypeCmd)
	deleteCmd.AddCommand(deleteInstanceCmd)
}

func runDeleteObjectTypeCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	
	// Check if deletions are allowed
	if !client.IsDeleteAllowed() {
		return fmt.Errorf("delete operations are disabled - set ATLASSIAN_ASSETS_ALLOW_DELETE=true in environment to enable")
	}
	
	// Validate input
	if deleteID == "" && deleteObjectTypeName == "" {
		return fmt.Errorf("must specify either --id or --name")
	}
	
	if deleteObjectTypeName != "" && deleteObjectTypeSchema == "" {
		return fmt.Errorf("must specify --schema when using --name")
	}
	
	// Resolve object type ID if needed
	var objectTypeID string
	if deleteID != "" {
		objectTypeID = deleteID
	} else {
		// TODO: Implement name-to-ID resolution
		return fmt.Errorf("deletion by name not yet implemented")
	}
	
	// Get object type details for confirmation
	objectType, err := client.GetObjectType(ctx, objectTypeID)
	if err != nil {
		return fmt.Errorf("failed to get object type details: %w", err)
	}
	
	if !objectType.Success {
		return fmt.Errorf("failed to get object type: %v", objectType.Error)
	}
	
	// Safety confirmation
	if !deleteForce && !deleteConfirm {
		return fmt.Errorf("deletion requires --confirm or --force flag for safety")
	}
	
	// Perform deletion
	response, err := client.DeleteObjectType(ctx, objectTypeID)
	if err != nil {
		return fmt.Errorf("failed to delete object type: %w", err)
	}
	
	if !response.Success {
		return fmt.Errorf("delete operation failed: %v", response.Error)
	}
	
	result := map[string]interface{}{
		"action":         "delete_object_type",
		"object_type_id": objectTypeID,
		"force":          deleteForce,
		"confirm":        deleteConfirm,
		"deleted":        true,
	}
	
	// Add contextual hints
	enhancedResponse := addNextStepHints(NewSuccessResponse(result), "delete_object_type", map[string]interface{}{
		"object_type_id": objectTypeID,
		"success":        true,
		"force":          deleteForce,
	})
	
	return outputResult(enhancedResponse)
}

func runDeleteInstanceCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	
	// Check if deletions are allowed
	if !client.IsDeleteAllowed() {
		return fmt.Errorf("delete operations are disabled - set ATLASSIAN_ASSETS_ALLOW_DELETE=true in environment to enable")
	}
	
	// Validate input
	if deleteID == "" && deleteInstanceQuery == "" {
		return fmt.Errorf("must specify either --id or --query")
	}
	
	var instanceIDs []string
	var deletedCount int
	
	if deleteID != "" {
		// Parse comma-separated IDs
		instanceIDs = strings.Split(deleteID, ",")
		for i, id := range instanceIDs {
			instanceIDs[i] = strings.TrimSpace(id)
		}
	} else {
		// Query-based deletion
		searchResponse, err := client.SearchObjects(ctx, deleteInstanceQuery, deleteInstanceLimit)
		if err != nil {
			return fmt.Errorf("failed to search instances: %w", err)
		}
		
		if !searchResponse.Success {
			return fmt.Errorf("search failed: %v", searchResponse.Error)
		}
		
		// Extract instance IDs from search results
		data := searchResponse.Data.(map[string]interface{})
		if objects, ok := data["objects"].([]interface{}); ok {
			for _, obj := range objects {
				if objectMap, ok := obj.(map[string]interface{}); ok {
					if id, ok := objectMap["id"]; ok {
						instanceIDs = append(instanceIDs, fmt.Sprintf("%v", id))
					}
				}
			}
		}
	}
	
	if len(instanceIDs) == 0 {
		return fmt.Errorf("no instances found to delete")
	}
	
	// Safety confirmation
	if !deleteForce && !deleteConfirm {
		return fmt.Errorf("deletion of %d instances requires --confirm or --force flag for safety", len(instanceIDs))
	}
	
	// Delete instances
	var errors []string
	var deletedIDs []string
	
	for _, instanceID := range instanceIDs {
		response, err := client.DeleteObject(ctx, instanceID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Instance %s: %v", instanceID, err))
			continue
		}
		
		if !response.Success {
			errors = append(errors, fmt.Sprintf("Instance %s: %v", instanceID, response.Error))
			continue
		}
		
		deletedIDs = append(deletedIDs, instanceID)
		deletedCount++
	}
	
	result := map[string]interface{}{
		"action":         "delete_instances",
		"requested_ids":  instanceIDs,
		"deleted_ids":    deletedIDs,
		"deleted_count":  deletedCount,
		"total_count":    len(instanceIDs),
		"force":          deleteForce,
		"confirm":        deleteConfirm,
		"query":          deleteInstanceQuery,
	}
	
	if len(errors) > 0 {
		result["errors"] = errors
		result["success"] = false
	} else {
		result["success"] = true
	}
	
	// Add contextual hints
	enhancedResponse := addNextStepHints(NewSuccessResponse(result), "delete_instances", map[string]interface{}{
		"deleted_count": deletedCount,
		"total_count":   len(instanceIDs),
		"success":       len(errors) == 0,
		"has_errors":    len(errors) > 0,
	})
	
	return outputResult(enhancedResponse)
}