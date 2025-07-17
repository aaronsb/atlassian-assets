package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/aaronsb/atlassian-assets/internal/hints"
)

// BROWSE command - high-level schema browsing tools
var browseCmd = &cobra.Command{
	Use:   "browse",
	Short: "High-level browsing and exploration tools",
	Long: `Composite commands that combine fundamental tools for common schema exploration patterns.
	
These are "glue" commands that string together lower-level tools to provide
friendly interfaces for common tasks like exploring hierarchies and comparing types.`,
}

// BROWSE HIERARCHY subcommand
var browseHierarchyCmd = &cobra.Command{
	Use:   "hierarchy",
	Short: "Show object type hierarchy for a schema",
	Long: `Display the parent-child relationships of object types in a schema.
	
This combines schema types + jq filtering to show the hierarchical structure.`,
	Example: `  # Show IT Employee Assets hierarchy
  assets browse hierarchy --schema 7
  
  # Show specific parent's children
  assets browse hierarchy --schema 7 --parent 28`,
	RunE: runBrowseHierarchyCmd,
}

var (
	browseSchema string
	browseParent string
)

func init() {
	browseHierarchyCmd.Flags().StringVar(&browseSchema, "schema", "", "Schema ID to browse (required)")
	browseHierarchyCmd.Flags().StringVar(&browseParent, "parent", "", "Show children of specific parent type")
	browseHierarchyCmd.MarkFlagRequired("schema")
}

func runBrowseHierarchyCmd(cmd *cobra.Command, args []string) error {
	// This is a "glue" command - it calls the fundamental schema types command
	// then processes the output with structured logic instead of complex jq
	
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	
	// Get all object types using our fundamental command
	response, err := client.GetObjectTypes(ctx, browseSchema)
	if err != nil {
		return fmt.Errorf("failed to get object types: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("API error: %s", response.Error)
	}

	data := response.Data.(map[string]interface{})
	objectTypesData := data["object_types"]
	
	// Parse into a structure we can work with
	jsonBytes, _ := json.Marshal(objectTypesData)
	var objectTypes []map[string]interface{}
	json.Unmarshal(jsonBytes, &objectTypes)

	// Build hierarchy structure
	hierarchy := buildHierarchy(objectTypes, browseParent)
	
	result := map[string]interface{}{
		"action":     "browse_hierarchy",
		"schema_id":  browseSchema,
		"hierarchy":  hierarchy,
		"total_types": len(objectTypes),
	}
	
	if browseParent != "" {
		result["filtered_parent"] = browseParent
	}

	response := NewSuccessResponse(result)
	
	// Add contextual hints
	enhancedResponse := addNextStepHints(response, "browse_hierarchy", map[string]interface{}{
		"schema_id":      browseSchema,
		"success":        response.Success,
		"has_children":   len(hierarchy) > 0,
		"has_empty_types": false, // TODO: Check if types have no instances
	})

	return outputResult(enhancedResponse)
}

// BROWSE CHILDREN subcommand  
var browseChildrenCmd = &cobra.Command{
	Use:   "children",
	Short: "Show children of a specific object type",
	Long: `List all child object types of a parent type.
	
This is a focused view showing what inherits from a specific parent.`,
	Example: `  # Show all IT Hardware children
  assets browse children --parent 28 --schema 7`,
	RunE: runBrowseChildrenCmd,
}

func runBrowseChildrenCmd(cmd *cobra.Command, args []string) error {
	// Another glue command - simpler than complex jq filtering
	return runBrowseHierarchyCmd(cmd, args) // Reuse hierarchy logic with parent filter
}

// BROWSE ATTRIBUTES subcommand
var browseAttributesCmd = &cobra.Command{
	Use:   "attrs",
	Short: "Compare attributes between object types",
	Long: `Show attribute comparison between object types in a friendly format.
	
Instead of complex jq, this provides structured attribute analysis.`,
	Example: `  # Compare laptop and workstation attributes
  assets browse attrs --types 69,141`,
	RunE: runBrowseAttributesCmd,
}

var browseTypes string

func init() {
	browseAttributesCmd.Flags().StringVar(&browseTypes, "types", "", "Comma-separated object type IDs to compare")
	browseAttributesCmd.MarkFlagRequired("types")
}

func runBrowseAttributesCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	typeIDs := strings.Split(browseTypes, ",")
	
	comparison := make(map[string]interface{})
	
	for _, typeID := range typeIDs {
		typeID = strings.TrimSpace(typeID)
		response, err := client.GetObjectTypeAttributes(ctx, typeID)
		if err != nil {
			continue
		}
		
		if response.Success {
			data := response.Data.(map[string]interface{})
			attributes := data["attributes"]
			
			// Extract just the essential info without complex jq
			jsonBytes, _ := json.Marshal(attributes)
			var attrs []map[string]interface{}
			json.Unmarshal(jsonBytes, &attrs)
			
			simplified := make([]map[string]interface{}, 0)
			for _, attr := range attrs {
				simplified = append(simplified, map[string]interface{}{
					"name":     attr["name"],
					"required": attr["minimumCardinality"] == float64(1),
					"editable": attr["editable"],
					"system":   attr["system"],
				})
			}
			
			comparison[typeID] = map[string]interface{}{
				"attributes": simplified,
				"count":      len(simplified),
			}
		}
	}
	
	result := map[string]interface{}{
		"action":     "browse_attributes", 
		"comparison": comparison,
		"type_ids":   typeIDs,
	}

	return outputResult(NewSuccessResponse(result))
}

// Helper function to build hierarchy without complex jq
func buildHierarchy(objectTypes []map[string]interface{}, parentFilter string) []map[string]interface{} {
	var result []map[string]interface{}
	
	for _, objType := range objectTypes {
		item := map[string]interface{}{
			"id":   objType["id"],
			"name": objType["name"],
		}
		
		if parent, exists := objType["parentObjectTypeId"]; exists {
			item["parent"] = parent
			
			// If filtering by parent, only include matching children
			if parentFilter != "" && parent != parentFilter {
				continue
			}
		} else {
			item["parent"] = nil
			
			// If filtering by parent, skip root items
			if parentFilter != "" {
				continue
			}
		}
		
		result = append(result, item)
	}
	
	return result
}

// Helper function to add contextual hints using centralized system
func addNextStepHints(response interface{}, commandType string, context map[string]interface{}) interface{} {
	// Convert response to map for modification
	responseMap := make(map[string]interface{})
	
	// Handle different response types
	switch r := response.(type) {
	case *Response:
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
	// Set up browse subcommands - flags already defined above
	browseChildrenCmd.Flags().StringVar(&browseSchema, "schema", "", "Schema ID to browse (required)")
	browseChildrenCmd.Flags().StringVar(&browseParent, "parent", "", "Parent object type ID (required)")
	browseChildrenCmd.MarkFlagRequired("schema")
	browseChildrenCmd.MarkFlagRequired("parent")
	
	browseCmd.AddCommand(browseHierarchyCmd)
	browseCmd.AddCommand(browseChildrenCmd)
	browseCmd.AddCommand(browseAttributesCmd)
}