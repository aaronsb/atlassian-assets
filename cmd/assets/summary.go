package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/aaronsb/atlassian-assets/internal/validation"
)

// SUMMARY command - high-level analysis and summary tools  
var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "High-level analysis and summary tools",
	Long: `Composite commands that provide friendly summaries of complex operations.
	
These commands encapsulate common analysis patterns instead of requiring
complex jq pipelines to extract meaningful information.`,
}

// SUMMARY COMPLETION subcommand
var summaryCompletionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Analyze object completion results in friendly format",
	Long: `Take object data and show what completion would achieve.
	
Instead of complex jq to analyze completion results, this provides
a structured summary of what would be applied, suggested, and completed.`,
	Example: `  # Analyze what completion would do for a workstation
  assets summary completion --type 141 --data '{"name":"AI Workstation"}'`,
	RunE: runSummaryCompletionCmd,
}

var (
	summaryType string
	summaryData string
)

func init() {
	summaryCompletionCmd.Flags().StringVar(&summaryType, "type", "", "Object type ID (required)")
	summaryCompletionCmd.Flags().StringVar(&summaryData, "data", "", "Object data as JSON string (required)")
	summaryCompletionCmd.MarkFlagRequired("type")
	summaryCompletionCmd.MarkFlagRequired("data")
}

func runSummaryCompletionCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Parse the JSON data
	var properties map[string]interface{}
	if err := json.Unmarshal([]byte(summaryData), &properties); err != nil {
		return fmt.Errorf("failed to parse JSON data: %w", err)
	}

	// Use our validation system to get completion results
	validator := validation.NewObjectValidator(client)
	result, err := validator.CompleteObject(ctx, summaryType, properties)
	if err != nil {
		return fmt.Errorf("completion analysis failed: %w", err)
	}

	// Instead of requiring jq, provide structured summary
	summary := map[string]interface{}{
		"action":                 "completion_summary",
		"object_type_id":         summaryType,
		"success":                result.Success,
		"provided_properties":    len(result.OriginalProperties),
		"completed_properties":   len(result.CompletedProperties),
		"applied_defaults_count": len(result.AppliedDefaults),
		"suggestions_count":      len(result.Suggestions),
		"missing_critical_count": len(result.MissingCritical),
	}

	// Add details sections
	if len(result.AppliedDefaults) > 0 {
		defaults := make([]map[string]interface{}, 0)
		for _, def := range result.AppliedDefaults {
			defaults = append(defaults, map[string]interface{}{
				"field":      def.Field,
				"value":      def.Value,
				"reason":     def.Reason,
				"confidence": def.Confidence,
			})
		}
		summary["applied_defaults"] = defaults
	}

	if len(result.MissingCritical) > 0 {
		summary["missing_critical"] = result.MissingCritical
	}

	// Add readable status
	if result.Success {
		summary["status"] = "✅ Ready to create - all requirements met"
	} else {
		summary["status"] = fmt.Sprintf("⚠️ Missing %d critical fields", len(result.MissingCritical))
	}

	return outputResult(NewSuccessResponse(summary))
}

// SUMMARY SCHEMA subcommand
var summarySchemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Show schema overview with counts and structure",
	Long: `Provide a high-level overview of a schema without requiring jq filtering.
	
Shows object type counts, hierarchy depth, and key statistics.`,
	Example: `  # Get IT Employee Assets overview
  assets summary schema --id 7`,
	RunE: runSummarySchemaCmd,
}

var summarySchemaID string

func init() {
	summarySchemaCmd.Flags().StringVar(&summarySchemaID, "id", "", "Schema ID (required)")
	summarySchemaCmd.MarkFlagRequired("id")
}

func runSummarySchemaCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Get schema info
	schemaResponse, err := client.GetSchema(ctx, summarySchemaID)
	if err != nil {
		return fmt.Errorf("failed to get schema: %w", err)
	}

	// Get object types
	typesResponse, err := client.GetObjectTypes(ctx, summarySchemaID)
	if err != nil {
		return fmt.Errorf("failed to get object types: %w", err)
	}

	if !schemaResponse.Success || !typesResponse.Success {
		return fmt.Errorf("API error getting schema data")
	}

	schemaData := schemaResponse.Data
	typesData := typesResponse.Data.(map[string]interface{})
	objectTypesData := typesData["object_types"]
	
	// Parse object types to analyze structure
	jsonBytes, _ := json.Marshal(objectTypesData)
	var objectTypes []map[string]interface{}
	json.Unmarshal(jsonBytes, &objectTypes)

	// Analyze structure without complex jq
	parentCounts := make(map[string]int)
	rootTypes := 0
	
	for _, objType := range objectTypes {
		if parent, exists := objType["parentObjectTypeId"]; exists && parent != nil {
			parentStr := parent.(string)
			parentCounts[parentStr]++
		} else {
			rootTypes++
		}
	}

	summary := map[string]interface{}{
		"action":            "schema_summary",
		"schema":            schemaData,
		"total_object_types": len(objectTypes),
		"root_types":        rootTypes,
		"parent_types":      len(parentCounts),
		"deepest_hierarchy": len(parentCounts), // Simplified for now
	}

	// Add top parents by child count
	if len(parentCounts) > 0 {
		topParents := make([]map[string]interface{}, 0)
		for parentID, count := range parentCounts {
			// Find the parent name
			parentName := parentID
			for _, objType := range objectTypes {
				if objType["id"] == parentID {
					parentName = objType["name"].(string)
					break
				}
			}
			
			topParents = append(topParents, map[string]interface{}{
				"id":           parentID,
				"name":         parentName,
				"child_count":  count,
			})
		}
		summary["top_parents"] = topParents
	}

	return outputResult(NewSuccessResponse(summary))
}

func init() {
	summaryCmd.AddCommand(summaryCompletionCmd)
	summaryCmd.AddCommand(summarySchemaCmd)
}