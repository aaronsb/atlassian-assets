package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/aaronsb/atlassian-assets/internal/client"
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
	listLimit  int
	listOffset int
)

func init() {
	listCmd.Flags().StringVar(&listSchema, "schema", "", "Schema ID or name (required)")
	listCmd.Flags().StringVar(&listType, "type", "", "Object type filter")
	listCmd.Flags().StringVar(&listFilter, "filter", "", "AQL filter query")
	listCmd.Flags().IntVar(&listLimit, "limit", 50, "Maximum number of results to return (1-1000)")
	listCmd.Flags().IntVar(&listOffset, "offset", 0, "Number of results to skip (for pagination)")
	
	listCmd.MarkFlagRequired("schema")
}

func runListCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	
	// Use ListObjectsWithPagination if available, otherwise fallback to ListObjects
	response, err := client.ListObjectsWithPagination(ctx, listSchema, listLimit, listOffset)
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
	Long: `Search for asset objects using either simple search terms or AQL (Assets Query Language).

Simple search provides exact matching for Name and Key fields.
AQL allows for complex queries across multiple schemas and object types.

Simple search patterns (exact matches only due to AQL limitations):
  term       - Exact match: Name or Key equals "term"
  ^exact$    - Exact match: Name or Key equals "exact"
  =value     - Exact match: Name or Key equals "value"
  *          - Wildcard match all objects in schema/filters

Note: Partial matching (contains/starts with/ends with) is not available 
due to AQL LIKE query limitations in the current environment.`,
	Example: `  # Simple searches (exact matches only)
  assets search --simple "Blue Barn #2"              # Exact name match
  assets search --simple "COMPUTE-1020"              # Exact key match
  assets search --simple "^Blue Barn #2$"            # Exact match with anchors
  assets search --simple "=Red Barn #1"              # Explicit exact match  
  assets search --simple "*" --schema 3              # All objects in schema
  assets search --simple "*" --type "Barns"          # All objects of type
  assets search --simple "john.doe" --schema 3       # Exact owner match
  
  # Advanced AQL searches
  assets search --query "Name like \"MacBook%\" AND Owner = \"john.doe\""
  assets search --query "objectSchemaId = 3 AND Status = \"Active\""
  
  # Pagination examples
  assets search --simple "*" --schema 7 --limit 100        # Get first 100 results
  assets search --simple "*" --schema 7 --limit 50 --offset 50  # Get next 50 results`,
	RunE: runSearchCmd,
}

var (
	searchQuery  string
	searchSimple string
	searchSchema string
	searchType   string
	searchStatus string
	searchOwner  string
	searchLimit  int
	searchOffset int
)

func init() {
	searchCmd.Flags().StringVar(&searchQuery, "query", "", "AQL search query")
	searchCmd.Flags().StringVar(&searchSimple, "simple", "", "Simple search term (searches name, key, and description)")
	searchCmd.Flags().StringVar(&searchSchema, "schema", "", "Limit search to specific schema (ID or name)")
	searchCmd.Flags().StringVar(&searchType, "type", "", "Limit search to specific object type")
	searchCmd.Flags().StringVar(&searchStatus, "status", "", "Filter by status (Active, Inactive, etc.)")
	searchCmd.Flags().StringVar(&searchOwner, "owner", "", "Filter by owner/assignee")
	searchCmd.Flags().IntVar(&searchLimit, "limit", 50, "Maximum number of results to return (1-1000)")
	searchCmd.Flags().IntVar(&searchOffset, "offset", 0, "Number of results to skip (for pagination)")
	
	// Make search mutually exclusive - either query OR simple + optional filters
	searchCmd.MarkFlagsMutuallyExclusive("query", "simple")
}

func runSearchCmd(cmd *cobra.Command, args []string) error {
	// Validate that either query or simple is provided
	if searchQuery == "" && searchSimple == "" {
		return fmt.Errorf("either --query or --simple must be provided")
	}

	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	
	// Validate pagination parameters
	if searchLimit < 1 || searchLimit > 1000 {
		return fmt.Errorf("limit must be between 1 and 1000, got %d", searchLimit)
	}
	if searchOffset < 0 {
		return fmt.Errorf("offset must be 0 or greater, got %d", searchOffset)
	}
	
	// Build AQL query based on input type
	var finalQuery string
	var queryType string
	
	if searchQuery != "" {
		// Use direct AQL query
		finalQuery = searchQuery
		queryType = "aql"
	} else {
		// Build AQL from simple search terms
		finalQuery, err = buildSimpleSearchQuery(searchSimple, searchSchema, searchType, searchStatus, searchOwner)
		if err != nil {
			return fmt.Errorf("failed to build search query: %w", err)
		}
		queryType = "simple"
	}
	
	response, err := client.SearchObjectsWithPagination(ctx, finalQuery, searchLimit, searchOffset)
	if err != nil {
		return fmt.Errorf("failed to search objects: %w", err)
	}

	// Add contextual hints
	enhancedResponse := addNextStepHints(response, "search_objects", map[string]interface{}{
		"search_query":     finalQuery,
		"query_type":       queryType,
		"simple_term":      searchSimple,
		"search_filters":   buildFilterSummary(),
		"success":          response.Success,
		"has_results":      getResultCount(response) > 0,
	})

	return outputResult(enhancedResponse)
}

// buildSimpleSearchQuery converts simple search terms into AQL with regex-inspired filters
func buildSimpleSearchQuery(term, schema, objectType, status, owner string) (string, error) {
	var conditions []string
	
	// Add schema filter if provided
	if schema != "" {
		conditions = append(conditions, fmt.Sprintf("objectSchemaId = %s", schema))
	}
	
	// Add object type filter if provided
	if objectType != "" {
		conditions = append(conditions, fmt.Sprintf("objectType = \"%s\"", objectType))
	}
	
	// Add status filter if provided
	if status != "" {
		conditions = append(conditions, fmt.Sprintf("Status = \"%s\"", status))
	}
	
	// Add owner filter if provided
	if owner != "" {
		conditions = append(conditions, fmt.Sprintf("Owner = \"%s\"", owner))
	}
	
	// Add simple term search with regex-inspired patterns
	if term != "" {
		termCondition := buildTermSearchCondition(term)
		conditions = append(conditions, termCondition)
	}
	
	// If no conditions, return error
	if len(conditions) == 0 {
		return "", fmt.Errorf("no search conditions provided")
	}
	
	// Combine all conditions with AND
	query := conditions[0]
	for _, cond := range conditions[1:] {
		query += " AND " + cond
	}
	
	return query, nil
}

// buildTermSearchCondition creates AQL search condition with basic patterns (LIKE queries are non-functional)
func buildTermSearchCondition(term string) string {
	// Determine search pattern based on term format
	var nameCondition, keyCondition string
	
	switch {
	case term == "*":
		// Wildcard - match all non-empty values
		return "(Name != \"\" OR Key != \"\")"
		
	case strings.HasPrefix(term, "^") && strings.HasSuffix(term, "$"):
		// Exact match: ^exact$ -> exact match
		exactTerm := strings.TrimSuffix(strings.TrimPrefix(term, "^"), "$")
		nameCondition = fmt.Sprintf("Name = \"%s\"", exactTerm)
		keyCondition = fmt.Sprintf("Key = \"%s\"", exactTerm)
		
	case strings.HasPrefix(term, "="):
		// Exact match: =value -> exact match
		exactTerm := strings.TrimPrefix(term, "=")
		nameCondition = fmt.Sprintf("Name = \"%s\"", exactTerm)
		keyCondition = fmt.Sprintf("Key = \"%s\"", exactTerm)
		
	default:
		// Default: exact match only (LIKE queries don't work in this AQL implementation)
		// This is a limitation - partial matches are not supported
		nameCondition = fmt.Sprintf("Name = \"%s\"", term)
		keyCondition = fmt.Sprintf("Key = \"%s\"", term)
	}
	
	// Return combined condition
	return fmt.Sprintf("(%s OR %s)", nameCondition, keyCondition)
}

// buildFilterSummary creates a summary of active filters for hints
func buildFilterSummary() map[string]string {
	filters := make(map[string]string)
	
	if searchSchema != "" {
		filters["schema"] = searchSchema
	}
	if searchType != "" {
		filters["type"] = searchType
	}
	if searchStatus != "" {
		filters["status"] = searchStatus
	}
	if searchOwner != "" {
		filters["owner"] = searchOwner
	}
	
	return filters
}

// getResultCount extracts result count from response  
func getResultCount(response *client.Response) int {
	if !response.Success || response.Data == nil {
		return 0
	}
	
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return 0
	}
	
	total, ok := data["total"].(int)
	if !ok {
		return 0
	}
	
	return total
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

