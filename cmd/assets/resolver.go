package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/aaronsb/atlassian-assets/internal/resolver"
)

// RESOLVE command with subcommands
var resolveCmd = &cobra.Command{
	Use:   "resolve",
	Short: "Resolve between human names and internal IDs",
	Long: `Resolve between human-readable names and internal IDs for schemas, object types, and objects.
	
The resolver provides bidirectional translation between the names you see in the UI 
and the internal IDs used by the API. This solves the critical abstraction problem 
where IDs can represent different things across schemas and time.`,
}

// RESOLVE SCHEMA subcommand
var resolveSchemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Resolve schema names and IDs",
	Long: `Resolve between schema names and IDs.
	
Supports both directions:
- Name to ID: resolve schema "Facilities" -> "6"
- ID to Name: resolve schema "6" -> "Facilities"`,
	Example: `  # Resolve schema name to ID
  assets resolve schema --name "Facilities"
  
  # Resolve schema ID to name
  assets resolve schema --id "6"`,
	RunE: runResolveSchemaCmd,
}

var (
	resolveSchemaName string
	resolveSchemaID   string
)

func init() {
	resolveSchemaCmd.Flags().StringVar(&resolveSchemaName, "name", "", "Schema name to resolve to ID")
	resolveSchemaCmd.Flags().StringVar(&resolveSchemaID, "id", "", "Schema ID to resolve to name")
}

func runResolveSchemaCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	resolver := resolver.NewResolver(client)
	ctx := context.Background()

	var response *Response

	if resolveSchemaName != "" && resolveSchemaID != "" {
		return fmt.Errorf("specify either --name or --id, not both")
	}

	if resolveSchemaName == "" && resolveSchemaID == "" {
		// List all schemas with resolution info
		schemas, err := resolver.ListResolvedSchemas(ctx)
		if err != nil {
			return fmt.Errorf("failed to list schemas: %w", err)
		}

		response = NewSuccessResponse(map[string]interface{}{
			"action":  "list_schemas",
			"schemas": schemas,
			"count":   len(schemas),
		})
	} else if resolveSchemaName != "" {
		// Resolve name to ID
		id, err := resolver.ResolveSchemaID(ctx, resolveSchemaName)
		if err != nil {
			return fmt.Errorf("failed to resolve schema name: %w", err)
		}

		response = NewSuccessResponse(map[string]interface{}{
			"action": "name_to_id",
			"input":  resolveSchemaName,
			"result": id,
			"type":   "schema",
		})
	} else {
		// Resolve ID to name
		name, err := resolver.ResolveSchemaName(ctx, resolveSchemaID)
		if err != nil {
			return fmt.Errorf("failed to resolve schema ID: %w", err)
		}

		response = NewSuccessResponse(map[string]interface{}{
			"action": "id_to_name",
			"input":  resolveSchemaID,
			"result": name,
			"type":   "schema",
		})
	}

	return outputResult(response)
}

// RESOLVE TYPE subcommand
var resolveTypeCmd = &cobra.Command{
	Use:   "type",
	Short: "Resolve object type names and IDs",
	Long: `Resolve between object type names and IDs within a schema.
	
Object types are scoped to schemas, so you must specify the schema context.`,
	Example: `  # Resolve type name to ID
  assets resolve type --schema "Facilities" --name "Bicycles"
  
  # Resolve type ID to name
  assets resolve type --id "52"`,
	RunE: runResolveTypeCmd,
}

var (
	resolveTypeSchema string
	resolveTypeName   string
	resolveTypeID     string
)

func init() {
	resolveTypeCmd.Flags().StringVar(&resolveTypeSchema, "schema", "", "Schema name or ID (required for name resolution)")
	resolveTypeCmd.Flags().StringVar(&resolveTypeName, "name", "", "Object type name to resolve to ID")
	resolveTypeCmd.Flags().StringVar(&resolveTypeID, "id", "", "Object type ID to resolve to name")
}

func runResolveTypeCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	resolver := resolver.NewResolver(client)
	ctx := context.Background()

	var response *Response

	if resolveTypeName != "" && resolveTypeID != "" {
		return fmt.Errorf("specify either --name or --id, not both")
	}

	if resolveTypeName == "" && resolveTypeID == "" {
		// List all object types for schema
		if resolveTypeSchema == "" {
			return fmt.Errorf("--schema is required when listing object types")
		}

		objectTypes, err := resolver.ListResolvedObjectTypes(ctx, resolveTypeSchema)
		if err != nil {
			return fmt.Errorf("failed to list object types: %w", err)
		}

		response = NewSuccessResponse(map[string]interface{}{
			"action":       "list_object_types",
			"schema":       resolveTypeSchema,
			"object_types": objectTypes,
			"count":        len(objectTypes),
		})
	} else if resolveTypeName != "" {
		// Resolve name to ID
		if resolveTypeSchema == "" {
			return fmt.Errorf("--schema is required when resolving object type name")
		}

		id, err := resolver.ResolveObjectTypeID(ctx, resolveTypeSchema, resolveTypeName)
		if err != nil {
			return fmt.Errorf("failed to resolve object type name: %w", err)
		}

		response = NewSuccessResponse(map[string]interface{}{
			"action": "name_to_id",
			"input":  resolveTypeName,
			"result": id,
			"schema": resolveTypeSchema,
			"type":   "object_type",
		})
	} else {
		// Resolve ID to name
		name, schemaName, err := resolver.ResolveObjectTypeName(ctx, resolveTypeID)
		if err != nil {
			return fmt.Errorf("failed to resolve object type ID: %w", err)
		}

		response = NewSuccessResponse(map[string]interface{}{
			"action":      "id_to_name",
			"input":       resolveTypeID,
			"result":      name,
			"schema_name": schemaName,
			"type":        "object_type",
		})
	}

	return outputResult(response)
}

// RESOLVE OBJECT subcommand
var resolveObjectCmd = &cobra.Command{
	Use:   "object",
	Short: "Resolve object references and get detailed info",
	Long: `Resolve object references and get detailed information.
	
Supports multiple input formats:
- Numeric ID: "384"
- Object key: "FAC-384" 
- Compound format: "Facilities/FAC-384"`,
	Example: `  # Resolve object by ID
  assets resolve object --ref "384"
  
  # Resolve object by key
  assets resolve object --ref "FAC-384"
  
  # Resolve with schema context
  assets resolve object --ref "Facilities/FAC-384"`,
	RunE: runResolveObjectCmd,
}

var resolveObjectRef string

func init() {
	resolveObjectCmd.Flags().StringVar(&resolveObjectRef, "ref", "", "Object reference (ID, key, or schema/key)")
	resolveObjectCmd.MarkFlagRequired("ref")
}

func runResolveObjectCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	resolver := resolver.NewResolver(client)
	ctx := context.Background()

	// First resolve the reference to an ID
	objectID, err := resolver.ResolveObjectID(ctx, resolveObjectRef)
	if err != nil {
		return fmt.Errorf("failed to resolve object reference: %w", err)
	}

	// Get detailed object information
	objectInfo, err := resolver.GetObjectInfo(ctx, objectID)
	if err != nil {
		return fmt.Errorf("failed to get object info: %w", err)
	}

	// Get additional context (schema and object type names)
	var schemaName, objectTypeName string
	if objectInfo.SchemaID != "" {
		schemaName, _ = resolver.ResolveSchemaName(ctx, objectInfo.SchemaID)
	}
	if objectInfo.ParentID != "" {
		objectTypeName, _, _ = resolver.ResolveObjectTypeName(ctx, objectInfo.ParentID)
	}

	response := NewSuccessResponse(map[string]interface{}{
		"action":           "resolve_object",
		"input":            resolveObjectRef,
		"object_id":        objectInfo.ID,
		"object_key":       objectInfo.Name,
		"display_name":     objectInfo.DisplayName,
		"schema_id":        objectInfo.SchemaID,
		"schema_name":      schemaName,
		"object_type_id":   objectInfo.ParentID,
		"object_type_name": objectTypeName,
		"last_updated":     objectInfo.LastUpdated,
	})

	return outputResult(response)
}

// RESOLVE STATS subcommand
var resolveStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show resolver cache statistics",
	Long: `Show statistics about the resolver cache including counts and refresh status.`,
	Example: `  # Show cache stats
  assets resolve stats`,
	RunE: runResolveStatsCmd,
}

func runResolveStatsCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	resolver := resolver.NewResolver(client)
	stats := resolver.GetCacheStats()

	response := NewSuccessResponse(map[string]interface{}{
		"action": "cache_stats",
		"stats":  stats,
	})

	return outputResult(response)
}

// RESOLVE REFRESH subcommand
var resolveRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh the resolver cache",
	Long: `Force refresh of the resolver cache by fetching latest data from the API.`,
	Example: `  # Refresh cache
  assets resolve refresh`,
	RunE: runResolveRefreshCmd,
}

func runResolveRefreshCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	resolver := resolver.NewResolver(client)
	ctx := context.Background()

	if err := resolver.RefreshCache(ctx); err != nil {
		return fmt.Errorf("failed to refresh cache: %w", err)
	}

	stats := resolver.GetCacheStats()

	response := NewSuccessResponse(map[string]interface{}{
		"action":  "refresh_cache",
		"message": "Cache refreshed successfully",
		"stats":   stats,
	})

	return outputResult(response)
}

func init() {
	// Add subcommands to resolve command
	resolveCmd.AddCommand(resolveSchemaCmd)
	resolveCmd.AddCommand(resolveTypeCmd)
	resolveCmd.AddCommand(resolveObjectCmd)
	resolveCmd.AddCommand(resolveStatsCmd)
	resolveCmd.AddCommand(resolveRefreshCmd)
}