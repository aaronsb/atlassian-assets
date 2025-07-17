package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/aaronsb/atlassian-assets/internal/client"
	"github.com/aaronsb/atlassian-assets/internal/hints"
)

// CREATE command with subcommands for streamlined workflow
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create assets with guided workflow",
	Long: `Create new object types and instances with intelligent guidance.
	
This command provides a streamlined workflow for creating assets with contextual hints
for next steps and automatic validation.`,
}

// CREATE OBJECT-TYPE subcommand
var createObjectTypeCmd = &cobra.Command{
	Use:   "object-type",
	Short: "Create a new object type with guided setup",
	Long: `Create a new object type in a schema with intelligent defaults and next-step suggestions.
	
This command creates the basic object type structure and provides contextual hints
for enhancement, attribute assignment, and instance creation.`,
	Example: `  # Create a new top-level object type
  assets create object-type --schema 3 --name "Data Centers"
  
  # Create with description and parent
  assets create object-type --schema 3 --name "Servers" --parent "Infrastructure" --description "Physical and virtual servers"`,
	RunE: runCreateObjectTypeCmd,
}

var (
	createObjectTypeSchema      string
	createObjectTypeName        string
	createObjectTypeDescription string
	createObjectTypeParent      string
	createObjectTypeIcon        string
)

func init() {
	createObjectTypeCmd.Flags().StringVar(&createObjectTypeSchema, "schema", "", "Schema ID or name (required)")
	createObjectTypeCmd.Flags().StringVar(&createObjectTypeName, "name", "", "Object type name (required)")
	createObjectTypeCmd.Flags().StringVar(&createObjectTypeDescription, "description", "", "Description of the object type")
	createObjectTypeCmd.Flags().StringVar(&createObjectTypeParent, "parent", "", "Parent object type ID or name")
	createObjectTypeCmd.Flags().StringVar(&createObjectTypeIcon, "icon", "", "Icon ID or name")
	
	createObjectTypeCmd.MarkFlagRequired("schema")
	createObjectTypeCmd.MarkFlagRequired("name")
}

func runCreateObjectTypeCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	
	// Resolve parent name to ID if needed
	var parentPtr *string
	if createObjectTypeParent != "" {
		parentPtr = &createObjectTypeParent
	}

	// Create the object type
	response, err := client.CreateObjectType(ctx, createObjectTypeSchema, createObjectTypeName, createObjectTypeDescription, createObjectTypeIcon, parentPtr)
	if err != nil {
		return fmt.Errorf("failed to create object type: %w", err)
	}

	// Add contextual hints
	if response.Success {
		enhancedResponse := addNextStepHints(response, "create_object_type", map[string]interface{}{
			"object_type_name": createObjectTypeName,
			"schema_id":        createObjectTypeSchema,
			"has_parent":       parentPtr != nil,
			"has_description":  createObjectTypeDescription != "",
			"has_custom_icon":  createObjectTypeIcon != "",
		})
		return outputResult(enhancedResponse)
	}

	return outputResult(response)
}

// CREATE INSTANCE subcommand (legacy compatibility)
var createInstanceCmd = &cobra.Command{
	Use:   "instance",
	Short: "Create a new object instance (legacy - use 'complete' for better experience)",
	Long: `Create a new object instance with basic attributes.
	
For a better experience with intelligent completion and validation,
use 'assets complete' instead of this command.`,
	Example: `  # Create instance (legacy)
  assets create instance --schema 3 --type 42 --data '{"name":"SERVER-001"}'
  
  # Better approach (recommended)
  assets complete --type 42 --data '{"name":"SERVER-001"}'`,
	RunE: runCreateInstanceCmd,
}

var (
	createInstanceSchema string
	createInstanceType   string
	createInstanceData   string
)

func init() {
	createInstanceCmd.Flags().StringVar(&createInstanceSchema, "schema", "", "Schema ID or name (required)")
	createInstanceCmd.Flags().StringVar(&createInstanceType, "type", "", "Object type ID (required)")
	createInstanceCmd.Flags().StringVar(&createInstanceData, "data", "", "Object data as JSON string (required)")
	
	createInstanceCmd.MarkFlagRequired("schema")
	createInstanceCmd.MarkFlagRequired("type")
	createInstanceCmd.MarkFlagRequired("data")
}

func runCreateInstanceCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	return createAsset(ctx, client, createInstanceSchema, createInstanceType, createInstanceData)
}

// createAsset creates a new asset object (legacy function)
func createAsset(ctx context.Context, client *client.AssetsClient, schema, objectType, data string) error {
	// 1. Parse the JSON data
	var attributes map[string]interface{}
	if err := json.Unmarshal([]byte(data), &attributes); err != nil {
		return fmt.Errorf("failed to parse JSON data: %w", err)
	}

	// 2. Create the object using our client
	response, err := client.CreateObject(ctx, objectType, attributes)
	if err != nil {
		return fmt.Errorf("failed to create object: %w", err)
	}

	// 3. Add hint suggesting better approach
	if response.Success {
		enhancedResponse := addNextStepHints(response, "create_instance_legacy", map[string]interface{}{
			"object_type_id": objectType,
		})
		return outputResult(enhancedResponse)
	}

	return outputResult(response)
}

// Helper function to add contextual hints using centralized system
func addNextStepHints(response interface{}, commandType string, context map[string]interface{}) interface{} {
	// Convert response to map for modification
	responseMap := make(map[string]interface{})
	
	// Handle different response types
	switch r := response.(type) {
	case *client.Response:
		responseMap["success"] = r.Success
		responseMap["data"] = r.Data
		if r.Error != "" {
			responseMap["error"] = r.Error
		}
		// Add success to context for hint evaluation
		context["success"] = r.Success
	case map[string]interface{}:
		responseMap = r
		// Add success to context for hint evaluation
		if success, ok := r["success"].(bool); ok {
			context["success"] = success
		}
	default:
		return response // Return as-is if we can't parse
	}
	
	// Get contextual hints from centralized system
	contextualHints := hints.GetContextualHints(commandType, context)
	
	if len(contextualHints) > 0 {
		responseMap["next_steps"] = contextualHints
	}
	
	return responseMap
}

func init() {
	// Add subcommands to create command
	createCmd.AddCommand(createObjectTypeCmd)
	createCmd.AddCommand(createInstanceCmd)
}