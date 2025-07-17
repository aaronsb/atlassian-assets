package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// SCHEMA command with subcommands
var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Manage asset schemas",
	Long: `Manage asset schemas and object types.
	
Schemas define the structure and types of assets that can be created.`,
}

// SCHEMA LIST subcommand
var schemaListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available schemas",
	Long: `List all available asset schemas in the workspace.`,
	Example: `  # List all schemas
  assets schema list`,
	RunE: runSchemaListCmd,
}

func runSchemaListCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	response, err := client.ListSchemas(ctx)
	if err != nil {
		return fmt.Errorf("failed to list schemas: %w", err)
	}

	// Add contextual hints
	enhancedResponse := addNextStepHints(response, "schema_management", map[string]interface{}{
		"success": response.Success,
		"action":  "list_schemas",
	})

	return outputResult(enhancedResponse)
}

// SCHEMA GET subcommand
var schemaGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get details of a specific schema",
	Long: `Get detailed information about a specific schema including its object types.`,
	Example: `  # Get schema details
  assets schema get --id computers`,
	RunE: runSchemaGetCmd,
}

var schemaGetID string

func init() {
	schemaGetCmd.Flags().StringVar(&schemaGetID, "id", "", "Schema ID (required)")
	schemaGetCmd.MarkFlagRequired("id")
}

func runSchemaGetCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	response, err := client.GetSchema(ctx, schemaGetID)
	if err != nil {
		return fmt.Errorf("failed to get schema: %w", err)
	}

	return outputResult(response)
}

// SCHEMA TYPES subcommand
var schemaTypesCmd = &cobra.Command{
	Use:   "types",
	Short: "List object types in a schema",
	Long: `List all object types available in a specific schema.`,
	Example: `  # List types in schema
  assets schema types --schema computers`,
	RunE: runSchemaTypesCmd,
}

var schemaTypesSchema string

func init() {
	schemaTypesCmd.Flags().StringVar(&schemaTypesSchema, "schema", "", "Schema ID (required)")
	schemaTypesCmd.MarkFlagRequired("schema")
}

func runSchemaTypesCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	response, err := client.GetObjectTypes(ctx, schemaTypesSchema)
	if err != nil {
		return fmt.Errorf("failed to get object types: %w", err)
	}

	// Add contextual hints
	enhancedResponse := addNextStepHints(response, "schema_management", map[string]interface{}{
		"success":   response.Success,
		"schema_id": schemaTypesSchema,
		"action":    "list_object_types",
	})

	return outputResult(enhancedResponse)
}

// SCHEMA CREATE-TYPE subcommand
var schemaCreateTypeCmd = &cobra.Command{
	Use:   "create-type",
	Short: "Create a new object type in a schema",
	Long: `Create a new object type in the specified schema.
	
This creates a new object type that can have instances created. Object types
define the structure and attributes for assets of that type.`,
	Example: `  # Create an AI Workstation object type
  assets schema create-type --schema 7 --name "AI Workstation" --description "High-performance workstations for AI/ML development" --icon "143"`,
	RunE: runSchemaCreateTypeCmd,
}

var (
	createTypeSchema      string
	createTypeName        string
	createTypeDescription string
	createTypeIconID      string
	createTypeParent      string
)

func init() {
	schemaCreateTypeCmd.Flags().StringVar(&createTypeSchema, "schema", "", "Schema ID where the object type will be created (required)")
	schemaCreateTypeCmd.Flags().StringVar(&createTypeName, "name", "", "Name of the new object type (required)")
	schemaCreateTypeCmd.Flags().StringVar(&createTypeDescription, "description", "", "Description of the object type")
	schemaCreateTypeCmd.Flags().StringVar(&createTypeIconID, "icon", "", "Icon ID for the object type (optional)")
	schemaCreateTypeCmd.Flags().StringVar(&createTypeParent, "parent", "", "Parent object type ID for inheritance (optional)")
	
	schemaCreateTypeCmd.MarkFlagRequired("schema")
	schemaCreateTypeCmd.MarkFlagRequired("name")
}

func runSchemaCreateTypeCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Convert parent string to pointer if provided
	var parentPtr *string
	if createTypeParent != "" {
		parentPtr = &createTypeParent
	}

	response, err := client.CreateObjectType(ctx, createTypeSchema, createTypeName, createTypeDescription, createTypeIconID, parentPtr)
	if err != nil {
		return fmt.Errorf("failed to create object type: %w", err)
	}

	// Add contextual hints
	enhancedResponse := addNextStepHints(response, "create_object_type", map[string]interface{}{
		"object_type_name": createTypeName,
		"schema_id":        createTypeSchema,
		"has_parent":       parentPtr != nil,
		"has_description":  createTypeDescription != "",
		"has_custom_icon":  createTypeIconID != "",
		"success":          response.Success,
	})

	return outputResult(enhancedResponse)
}


func init() {
	// Add subcommands to schema command
	schemaCmd.AddCommand(schemaListCmd)
	schemaCmd.AddCommand(schemaGetCmd)
	schemaCmd.AddCommand(schemaTypesCmd)
	schemaCmd.AddCommand(schemaCreateTypeCmd)
}