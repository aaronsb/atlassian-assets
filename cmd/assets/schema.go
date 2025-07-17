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

	return outputResult(response)
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

	return outputResult(response)
}

func init() {
	// Add subcommands to schema command
	schemaCmd.AddCommand(schemaListCmd)
	schemaCmd.AddCommand(schemaGetCmd)
	schemaCmd.AddCommand(schemaTypesCmd)
}