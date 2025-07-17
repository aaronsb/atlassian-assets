package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// LIST command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List asset objects",
	Long: `List asset objects from the specified schema.
	
Optionally filter by object type or use AQL (Assets Query Language) for advanced filtering.`,
	Example: `  # List all objects in schema
  assets list --schema computers
  
  # List objects by type
  assets list --schema computers --type laptop
  
  # List with AQL filter
  assets list --schema computers --filter "Name like 'MacBook%'"`,
	RunE: runListCmd,
}

var (
	listSchema string
	listType   string
	listFilter string
)

func init() {
	listCmd.Flags().StringVar(&listSchema, "schema", "", "Schema ID or name (required)")
	listCmd.Flags().StringVar(&listType, "type", "", "Object type filter")
	listCmd.Flags().StringVar(&listFilter, "filter", "", "AQL filter query")
	
	listCmd.MarkFlagRequired("schema")
}

func runListCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	
	// Default limit
	limit := 50
	
	response, err := client.ListObjects(ctx, listSchema, limit)
	if err != nil {
		return fmt.Errorf("failed to list objects: %w", err)
	}

	return outputResult(response)
}

// GET command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a specific asset object",
	Long: `Get details of a specific asset object by its ID.`,
	Example: `  # Get asset by ID
  assets get --id OBJ-123`,
	RunE: runGetCmd,
}

var getID string

func init() {
	getCmd.Flags().StringVar(&getID, "id", "", "Object ID (required)")
	getCmd.MarkFlagRequired("id")
}

func runGetCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	response, err := client.GetObject(ctx, getID)
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}

	return outputResult(response)
}

// UPDATE command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing asset object",
	Long: `Update an existing asset object with new data.
	
The data should be provided as a JSON string containing the attributes to update.`,
	Example: `  # Update asset owner
  assets update --id OBJ-123 --data '{"owner":"jane.doe"}'`,
	RunE: runUpdateCmd,
}

var (
	updateID   string
	updateData string
)

func init() {
	updateCmd.Flags().StringVar(&updateID, "id", "", "Object ID (required)")
	updateCmd.Flags().StringVar(&updateData, "data", "", "Updated data as JSON string (required)")
	
	updateCmd.MarkFlagRequired("id")
	updateCmd.MarkFlagRequired("data")
}

func runUpdateCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	response := NewSuccessResponse(map[string]interface{}{
		"action": "update",
		"id":     updateID,
		"data":   updateData,
		"status": "not_implemented",
		"message": "Asset update will be implemented using go-atlassian SDK",
	})

	return outputResult(response)
}

// DELETE command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an asset object",
	Long: `Delete an asset object by its ID.
	
This operation cannot be undone.`,
	Example: `  # Delete asset
  assets delete --id OBJ-123`,
	RunE: runDeleteCmd,
}

var deleteID string

func init() {
	deleteCmd.Flags().StringVar(&deleteID, "id", "", "Object ID (required)")
	deleteCmd.MarkFlagRequired("id")
}

func runDeleteCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	response := NewSuccessResponse(map[string]interface{}{
		"action": "delete",
		"id":     deleteID,
		"status": "not_implemented",
		"message": "Asset deletion will be implemented using go-atlassian SDK",
	})

	return outputResult(response)
}

// SEARCH command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for asset objects",
	Long: `Search for asset objects using AQL (Assets Query Language).
	
AQL allows for complex queries across multiple schemas and object types.`,
	Example: `  # Search for assets
  assets search --query "Name like 'MacBook%' AND Owner = 'john.doe'"`,
	RunE: runSearchCmd,
}

var searchQuery string

func init() {
	searchCmd.Flags().StringVar(&searchQuery, "query", "", "AQL search query (required)")
	searchCmd.MarkFlagRequired("query")
}

func runSearchCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	limit := 50
	
	response, err := client.SearchObjects(ctx, searchQuery, limit)
	if err != nil {
		return fmt.Errorf("failed to search objects: %w", err)
	}

	return outputResult(response)
}