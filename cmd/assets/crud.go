package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/aaronsb/atlassian-assets/internal/validation"
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

	// Add contextual hints
	enhancedResponse := addNextStepHints(response, "list_objects", map[string]interface{}{
		"schema_id":    listSchema,
		"success":      response.Success,
		"has_results":  true, // TODO: Check actual result count
	})

	return outputResult(enhancedResponse)
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

// DELETE command is now implemented in delete.go

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

	// Add contextual hints
	enhancedResponse := addNextStepHints(response, "search_objects", map[string]interface{}{
		"search_query": searchQuery,
		"success":      response.Success,
		"has_results":  true, // TODO: Check actual result count
	})

	return outputResult(enhancedResponse)
}

// ATTRIBUTES command
var attributesCmd = &cobra.Command{
	Use:   "attributes",
	Short: "Get object type attributes",
	Long: `Get all attributes for a specific object type.
	
This shows the schema definition including required attributes, data types, 
and validation rules needed for object creation and updates.`,
	Example: `  # Get attributes for object type
  assets attributes --type 133
  
  # Get attributes using resolver
  assets attributes --type "laptops" --schema "facilities"`,
	RunE: runAttributesCmd,
}

var (
	attributesType   string
	attributesSchema string
)

func init() {
	attributesCmd.Flags().StringVar(&attributesType, "type", "", "Object type ID or name (required)")
	attributesCmd.Flags().StringVar(&attributesSchema, "schema", "", "Schema ID or name (for name resolution)")
	
	attributesCmd.MarkFlagRequired("type")
}

func runAttributesCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	
	// For now, require direct object type ID to avoid cache timeout issues
	// TODO: Fix resolver cache performance issues before enabling name resolution
	objectTypeID := attributesType
	if attributesSchema != "" {
		return fmt.Errorf("name resolution temporarily disabled due to cache timeout issues. Please use object type ID directly (e.g., --type 133)")
	}
	
	response, err := client.GetObjectTypeAttributes(ctx, objectTypeID)
	if err != nil {
		return fmt.Errorf("failed to get object type attributes: %w", err)
	}

	return outputResult(response)
}

// VALIDATE command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate object properties against schema",
	Long: `Validate object properties against the object type schema.
	
This command checks if the provided properties meet the validation rules
including required fields, data types, and allowed values.`,
	Example: `  # Validate properties for laptops
  assets validate --type 133 --data '{"name":"Test Laptop","device_type":"Physical"}'`,
	RunE: runValidateCmd,
}

var (
	validateType string
	validateData string
)

func init() {
	validateCmd.Flags().StringVar(&validateType, "type", "", "Object type ID (required)")
	validateCmd.Flags().StringVar(&validateData, "data", "", "Property data as JSON string (required)")
	
	validateCmd.MarkFlagRequired("type")
	validateCmd.MarkFlagRequired("data")
}

func runValidateCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Parse the JSON data
	var properties map[string]interface{}
	if err := json.Unmarshal([]byte(validateData), &properties); err != nil {
		return fmt.Errorf("failed to parse JSON data: %w", err)
	}

	// Create validator and validate the object
	validator := validation.NewObjectValidator(client)
	result, err := validator.ValidateForCreate(ctx, validateType, properties)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Add summary to the response
	summary := validator.GetValidationSummary(result)
	
	response := NewSuccessResponse(map[string]interface{}{
		"action":            "validate_properties",
		"object_type_id":    validateType,
		"validation_result": result,
		"summary":           summary,
	})

	// Add contextual hints
	enhancedResponse := addNextStepHints(response, "validate_objects", map[string]interface{}{
		"object_type_id": validateType,
		"success":        response.Success,
		"has_errors":     false, // TODO: Check actual validation errors
	})

	return outputResult(enhancedResponse)
}

// COMPLETE command
var completeCmd = &cobra.Command{
	Use:   "complete",
	Short: "Intelligently complete object properties",
	Long: `Intelligently complete object properties with reasonable defaults and suggestions.
	
This command takes partial object information and attempts to create a complete,
valid object by applying intelligent defaults and providing helpful suggestions
for missing fields. Perfect for AI agents that have semantic understanding
but may not know exact schema requirements.`,
	Example: `  # Complete a laptop object with minimal info
  assets complete --type 133 --data '{"name":"John'\''s MacBook"}'
  
  # AI-friendly: provide what you know, get completion suggestions
  assets complete --type 133 --data '{"name":"Development Laptop","device_type":"Physical"}'`,
	RunE: runCompleteCmd,
}

var (
	completeType string
	completeData string
)

func init() {
	completeCmd.Flags().StringVar(&completeType, "type", "", "Object type ID (required)")
	completeCmd.Flags().StringVar(&completeData, "data", "", "Partial property data as JSON string (required)")
	
	completeCmd.MarkFlagRequired("type")
	completeCmd.MarkFlagRequired("data")
}

func runCompleteCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Parse the JSON data
	var properties map[string]interface{}
	if err := json.Unmarshal([]byte(completeData), &properties); err != nil {
		return fmt.Errorf("failed to parse JSON data: %w", err)
	}

	// Create validator and complete the object
	validator := validation.NewObjectValidator(client)
	result, err := validator.CompleteObject(ctx, completeType, properties)
	if err != nil {
		return fmt.Errorf("completion failed: %w", err)
	}

	// Add summary to the response
	summary := validator.GetCompletionSummary(result)
	
	response := NewSuccessResponse(map[string]interface{}{
		"action":            "complete_object",
		"object_type_id":    completeType,
		"completion_result": result,
		"summary":           summary,
	})

	// Add contextual hints
	enhancedResponse := addNextStepHints(response, "intelligent_completion", map[string]interface{}{
		"object_type_id": completeType,
		"success":        response.Success,
		"has_suggestions": result != nil,
	})

	return outputResult(enhancedResponse)
}

