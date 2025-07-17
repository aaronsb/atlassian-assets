package main

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ctreminiom/go-atlassian/v2/pkg/infra/models"
	"github.com/aaronsb/atlassian-assets/internal/client"
	"github.com/aaronsb/atlassian-assets/internal/hints"
)

// CATALOG command with subcommands for browsing global catalogs
var catalogCmd = &cobra.Command{
	Use:   "catalog",
	Short: "Browse global catalogs of attributes, objects, and schemas",
	Long: `Global catalog browsers for discovering reusable components across the workspace.
	
Search and browse attributes, object types, and schemas from across all schemas
to build your universal attribute marketplace.`,
}

// CATALOG ATTRIBUTES subcommand
var catalogAttributesCmd = &cobra.Command{
	Use:   "attributes",
	Short: "Browse all attributes across the workspace with pagination and search",
	Long: `Global attribute catalog with search capabilities.
	
Discovers every attribute across all schemas in the workspace and provides
search and filtering to find reusable attributes for your marketplace.`,
	Example: `  # Browse all attributes (paginated)
  assets catalog attributes
  
  # Search for CPU-related attributes
  assets catalog attributes --pattern "cpu|processor|chip"
  
  # Search for cost/price attributes
  assets catalog attributes --pattern "cost|price|budget"
  
  # Show specific page
  assets catalog attributes --page 2 --per-page 50
  
  # Show all in one schema
  assets catalog attributes --schema 7`,
	RunE: runCatalogAttributesCmd,
}

var (
	catalogPattern  string
	catalogSchema   string
	catalogPage     int
	catalogPerPage  int
	catalogAllPages bool
)

func init() {
	catalogAttributesCmd.Flags().StringVar(&catalogPattern, "pattern", "", "Regex pattern to match attribute names (case-insensitive)")
	catalogAttributesCmd.Flags().StringVar(&catalogSchema, "schema", "", "Limit to specific schema ID")
	catalogAttributesCmd.Flags().IntVar(&catalogPage, "page", 1, "Page number (1-based)")
	catalogAttributesCmd.Flags().IntVar(&catalogPerPage, "per-page", 25, "Results per page")
	catalogAttributesCmd.Flags().BoolVar(&catalogAllPages, "all", false, "Show all results (disable pagination)")
}

func runCatalogAttributesCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	
	if catalogSchema != "" {
		return catalogAttributesInSchema(ctx, client, catalogSchema)
	}
	
	return catalogAttributesAllSchemas(ctx, client)
}

// catalogAttributesInSchema catalogs attributes within a specific schema
func catalogAttributesInSchema(ctx context.Context, client *client.AssetsClient, schemaID string) error {
	// Get all object types in the schema
	response, err := client.GetObjectTypes(ctx, schemaID)
	if err != nil {
		return fmt.Errorf("failed to get object types: %w", err)
	}
	
	if !response.Success {
		return fmt.Errorf("failed to get object types: %s", response.Error)
	}
	
	data := response.Data.(map[string]interface{})
	objectTypesData := data["object_types"]
	
	jsonBytes, _ := json.Marshal(objectTypesData)
	var objectTypes []map[string]interface{}
	json.Unmarshal(jsonBytes, &objectTypes)
	
	// Collect all attributes
	var allAttributes []AttributeCatalogEntry
	
	for _, objType := range objectTypes {
		objTypeID := objType["id"].(string)
		objTypeName := objType["name"].(string)
		
		attrResponse, err := client.GetObjectTypeAttributes(ctx, objTypeID)
		if err != nil {
			continue // Skip failed lookups
		}
		
		if !attrResponse.Success {
			continue
		}
		
		// Process attributes
		attrData := attrResponse.Data.(map[string]interface{})
		attributesRaw := attrData["attributes"]
		
		switch attrs := attributesRaw.(type) {
		case []*models.ObjectTypeAttributeScheme:
			for _, attr := range attrs {
				entry := AttributeCatalogEntry{
					AttributeID:     attr.ID,
					Name:            attr.Name,
					Description:     attr.Description,
					DataType:        getDataTypeName(attr),
					DataTypeID:      getDataTypeID(attr),
					IsReference:     attr.Type == 1,
					IsSystem:        attr.System,
					Required:        attr.MinimumCardinality > 0,
					Editable:        attr.Editable,
					Unique:          attr.UniqueAttribute,
					Summable:        attr.Summable,
					ObjectTypeID:    objTypeID,
					ObjectTypeName:  objTypeName,
					SchemaID:        schemaID,
				}
				
				if attr.Type == 1 && attr.ReferenceObjectTypeID != "" {
					entry.ReferenceObjectTypeID = attr.ReferenceObjectTypeID
					entry.ReferenceType = getReferenceTypeName(attr)
				}
				
				allAttributes = append(allAttributes, entry)
			}
		}
	}
	
	return processAndDisplayCatalog(allAttributes, fmt.Sprintf("Schema %s", schemaID))
}

// catalogAttributesAllSchemas catalogs attributes across all schemas
func catalogAttributesAllSchemas(ctx context.Context, client *client.AssetsClient) error {
	// Get all schemas first
	schemasResponse, err := client.ListSchemas(ctx)
	if err != nil {
		return fmt.Errorf("failed to list schemas: %w", err)
	}
	
	if !schemasResponse.Success {
		return fmt.Errorf("failed to list schemas: %s", schemasResponse.Error)
	}
	
	schemasData := schemasResponse.Data.(map[string]interface{})
	schemas := schemasData["schemas"]
	
	jsonBytes, _ := json.Marshal(schemas)
	var schemasList []map[string]interface{}
	json.Unmarshal(jsonBytes, &schemasList)
	
	var allAttributes []AttributeCatalogEntry
	
	for _, schema := range schemasList {
		schemaID := schema["id"].(string)
		schemaName := schema["name"].(string)
		
		// Get object types in this schema
		response, err := client.GetObjectTypes(ctx, schemaID)
		if err != nil {
			continue // Skip failed schemas
		}
		
		if !response.Success {
			continue
		}
		
		data := response.Data.(map[string]interface{})
		objectTypesData := data["object_types"]
		
		jsonBytes, _ := json.Marshal(objectTypesData)
		var objectTypes []map[string]interface{}
		json.Unmarshal(jsonBytes, &objectTypes)
		
		// Collect attributes from this schema
		for _, objType := range objectTypes {
			objTypeID := objType["id"].(string)
			objTypeName := objType["name"].(string)
			
			attrResponse, err := client.GetObjectTypeAttributes(ctx, objTypeID)
			if err != nil {
				continue
			}
			
			if !attrResponse.Success {
				continue
			}
			
			attrData := attrResponse.Data.(map[string]interface{})
			attributesRaw := attrData["attributes"]
			
			switch attrs := attributesRaw.(type) {
			case []*models.ObjectTypeAttributeScheme:
				for _, attr := range attrs {
					entry := AttributeCatalogEntry{
						AttributeID:     attr.ID,
						Name:            attr.Name,
						Description:     attr.Description,
						DataType:        getDataTypeName(attr),
						DataTypeID:      getDataTypeID(attr),
						IsReference:     attr.Type == 1,
						IsSystem:        attr.System,
						Required:        attr.MinimumCardinality > 0,
						Editable:        attr.Editable,
						Unique:          attr.UniqueAttribute,
						Summable:        attr.Summable,
						ObjectTypeID:    objTypeID,
						ObjectTypeName:  objTypeName,
						SchemaID:        schemaID,
						SchemaName:      schemaName,
					}
					
					if attr.Type == 1 && attr.ReferenceObjectTypeID != "" {
						entry.ReferenceObjectTypeID = attr.ReferenceObjectTypeID
						entry.ReferenceType = getReferenceTypeName(attr)
					}
					
					allAttributes = append(allAttributes, entry)
				}
			}
		}
	}
	
	return processAndDisplayCatalog(allAttributes, "All Schemas")
}

// AttributeCatalogEntry represents a cataloged attribute
type AttributeCatalogEntry struct {
	AttributeID           string `json:"attribute_id"`
	Name                  string `json:"name"`
	Description           string `json:"description"`
	DataType              string `json:"data_type"`
	DataTypeID            int    `json:"data_type_id,omitempty"`
	IsReference           bool   `json:"is_reference"`
	ReferenceObjectTypeID string `json:"reference_object_type_id,omitempty"`
	ReferenceType         string `json:"reference_type,omitempty"`
	IsSystem              bool   `json:"is_system"`
	Required              bool   `json:"required"`
	Editable              bool   `json:"editable"`
	Unique                bool   `json:"unique"`
	Summable              bool   `json:"summable"`
	ObjectTypeID          string `json:"object_type_id"`
	ObjectTypeName        string `json:"object_type_name"`
	SchemaID              string `json:"schema_id"`
	SchemaName            string `json:"schema_name,omitempty"`
}

// processAndDisplayCatalog filters, sorts, paginates and displays the catalog
func processAndDisplayCatalog(allAttributes []AttributeCatalogEntry, scope string) error {
	// Filter by pattern if provided
	var filteredAttributes []AttributeCatalogEntry
	
	if catalogPattern != "" {
		regex, err := regexp.Compile("(?i)" + catalogPattern) // Case-insensitive
		if err != nil {
			return fmt.Errorf("invalid pattern: %w", err)
		}
		
		for _, attr := range allAttributes {
			if regex.MatchString(attr.Name) || regex.MatchString(attr.Description) {
				filteredAttributes = append(filteredAttributes, attr)
			}
		}
	} else {
		filteredAttributes = allAttributes
	}
	
	// Sort alphabetically by name
	sort.Slice(filteredAttributes, func(i, j int) bool {
		return strings.ToLower(filteredAttributes[i].Name) < strings.ToLower(filteredAttributes[j].Name)
	})
	
	// Pagination
	totalCount := len(filteredAttributes)
	
	var paginatedAttributes []AttributeCatalogEntry
	var pageInfo map[string]interface{}
	
	if catalogAllPages {
		paginatedAttributes = filteredAttributes
		pageInfo = map[string]interface{}{
			"pagination": "disabled",
			"total":      totalCount,
		}
	} else {
		startIdx := (catalogPage - 1) * catalogPerPage
		endIdx := startIdx + catalogPerPage
		
		if startIdx >= totalCount {
			paginatedAttributes = []AttributeCatalogEntry{}
		} else {
			if endIdx > totalCount {
				endIdx = totalCount
			}
			paginatedAttributes = filteredAttributes[startIdx:endIdx]
		}
		
		totalPages := (totalCount + catalogPerPage - 1) / catalogPerPage
		pageInfo = map[string]interface{}{
			"current_page": catalogPage,
			"per_page":     catalogPerPage,
			"total_pages":  totalPages,
			"total":        totalCount,
			"showing":      fmt.Sprintf("%d-%d of %d", startIdx+1, startIdx+len(paginatedAttributes), totalCount),
		}
	}
	
	result := map[string]interface{}{
		"action":     "catalog_attributes",
		"scope":      scope,
		"attributes": paginatedAttributes,
		"page_info":  pageInfo,
	}
	
	if catalogPattern != "" {
		result["pattern"] = catalogPattern
		result["pattern_matches"] = len(filteredAttributes)
	}
	
	response := NewSuccessResponse(result)
	
	// Add contextual hints
	enhancedResponse := addNextStepHints(response, "catalog_attributes", map[string]interface{}{
		"scope":         scope,
		"success":       response.Success,
		"has_results":   len(paginatedAttributes) > 0,
		"has_references": hasReferences(paginatedAttributes),
		"pattern":       catalogPattern,
		"result_count":  len(paginatedAttributes),
	})
	
	return outputResult(enhancedResponse)
}

// Helper functions
func getDataTypeName(attr *models.ObjectTypeAttributeScheme) string {
	if attr.DefaultType != nil {
		return attr.DefaultType.Name
	}
	return ""
}

func getDataTypeID(attr *models.ObjectTypeAttributeScheme) int {
	if attr.DefaultType != nil {
		return attr.DefaultType.ID
	}
	return 0
}

func getReferenceTypeName(attr *models.ObjectTypeAttributeScheme) string {
	if attr.ReferenceType != nil {
		return attr.ReferenceType.Name
	}
	return ""
}

// hasReferences checks if any attributes in the list are references
func hasReferences(attributes []AttributeCatalogEntry) bool {
	for _, attr := range attributes {
		if attr.IsReference {
			return true
		}
	}
	return false
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
	catalogCmd.AddCommand(catalogAttributesCmd)
}