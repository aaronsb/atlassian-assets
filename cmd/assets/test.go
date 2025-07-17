package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/spf13/cobra"
	"github.com/aaronsb/atlassian-assets/internal/client"
)

// TEST command with subcommands for test environment setup
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test environment setup and validation tools",
	Long: `Tools for creating and managing test environments.
	
Create test schemas, populate with sample data, and validate CLI functionality
in a controlled testing environment.`,
}

// TEST CREATE-SCHEMA subcommand
var testCreateSchemaCmd = &cobra.Command{
	Use:   "create-schema",
	Short: "Create a test schema with sample structure",
	Long: `Create a dedicated test schema for CLI testing and validation.
	
This creates a schema with a predictable structure that can be used
for comprehensive CLI testing without affecting production data.`,
	Example: `  # Create test schema with default name
  assets test create-schema
  
  # Create test schema with custom name
  assets test create-schema --name "CLI_Testing_Schema"
  
  # Create test schema with sample data
  assets test create-schema --with-sample-data`,
	RunE: runTestCreateSchemaCmd,
}

var (
	testSchemaName       string
	testWithSampleData   bool
	testSchemaPrefix     string
)

func init() {
	testCreateSchemaCmd.Flags().StringVar(&testSchemaName, "name", "", "Name for the test schema (default: auto-generated)")
	testCreateSchemaCmd.Flags().BoolVar(&testWithSampleData, "with-sample-data", false, "Create sample object types and objects")
	testCreateSchemaCmd.Flags().StringVar(&testSchemaPrefix, "prefix", "TEST", "Prefix for test schema identification")
}

func runTestCreateSchemaCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	
	// Generate schema name if not provided
	if testSchemaName == "" {
		timestamp := time.Now().Format("20060102_150405")
		randomSuffix := rand.Intn(1000)
		testSchemaName = fmt.Sprintf("%s_CLI_Testing_%s_%d", testSchemaPrefix, timestamp, randomSuffix)
	}
	
	// Create the test schema
	response, err := client.CreateSchema(ctx, testSchemaName, fmt.Sprintf("Test schema for CLI validation created at %s", time.Now().Format(time.RFC3339)))
	if err != nil {
		return fmt.Errorf("failed to create test schema: %w", err)
	}
	
	if !response.Success {
		return fmt.Errorf("schema creation failed: %v", response.Error)
	}
	
	result := map[string]interface{}{
		"action":           "create_test_schema",
		"schema_name":      testSchemaName,
		"prefix":           testSchemaPrefix,
		"with_sample_data": testWithSampleData,
		"schema_response":  response.Data,
	}
	
	// Add sample data if requested
	if testWithSampleData {
		// Extract schema ID from response (handle different response types)
		var schemaID string
		if schemaData, ok := response.Data.(map[string]interface{}); ok {
			schemaID = schemaData["id"].(string)
		} else {
			// Handle direct schema object response
			result["sample_data_error"] = "Schema ID extraction not implemented for this response type"
			result["sample_data_status"] = "skipped"
		}
		
		if schemaID != "" {
			sampleDataResult, err := createSampleTestData(ctx, client, schemaID)
			if err != nil {
				result["sample_data_error"] = err.Error()
				result["sample_data_status"] = "failed"
			} else {
				result["sample_data"] = sampleDataResult
				result["sample_data_status"] = "created"
			}
		}
	}
	
	successResponse := NewSuccessResponse(result)
	
	// Add contextual hints
	enhancedResponse := addNextStepHints(successResponse, "test_environment", map[string]interface{}{
		"schema_name":      testSchemaName,
		"has_sample_data":  testWithSampleData,
		"success":          successResponse.Success,
	})
	
	return outputResult(enhancedResponse)
}

// createSampleTestData creates sample object types and objects for testing
func createSampleTestData(ctx context.Context, assetsClient *client.AssetsClient, schemaID string) (map[string]interface{}, error) {
	sampleData := map[string]interface{}{
		"object_types": []string{},
		"objects":      []string{},
	}
	
	// Create sample object types
	objectTypes := []struct {
		name        string
		description string
		icon        string
	}{
		{"Test Servers", "Sample servers for CLI testing", "143"},
		{"Test Applications", "Sample applications for CLI testing", "144"},
		{"Test Locations", "Sample locations for CLI testing", "145"},
	}
	
	for _, ot := range objectTypes {
		response, err := assetsClient.CreateObjectType(ctx, schemaID, ot.name, ot.description, ot.icon, nil)
		if err != nil {
			continue // Skip failed creations but don't fail entirely
		}
		
		if response.Success {
			sampleData["object_types"] = append(sampleData["object_types"].([]string), ot.name)
		}
	}
	
	return sampleData, nil
}

// TEST CLEANUP subcommand
var testCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up test schemas and data",
	Long: `Remove test schemas created by the test setup commands.
	
This helps maintain a clean workspace by removing temporary test data.`,
	Example: `  # Clean up schemas with TEST prefix
  assets test cleanup --prefix TEST
  
  # Clean up specific test schema
  assets test cleanup --schema-name "TEST_CLI_Testing_20240717_123456"`,
	RunE: runTestCleanupCmd,
}

var (
	testCleanupPrefix     string
	testCleanupSchemaName string
	testCleanupDryRun     bool
)

func init() {
	testCleanupCmd.Flags().StringVar(&testCleanupPrefix, "prefix", "TEST", "Prefix to identify test schemas for cleanup")
	testCleanupCmd.Flags().StringVar(&testCleanupSchemaName, "schema-name", "", "Specific schema name to clean up")
	testCleanupCmd.Flags().BoolVar(&testCleanupDryRun, "dry-run", false, "Show what would be cleaned up without actually doing it")
}

func runTestCleanupCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	
	// List all schemas to find test schemas
	response, err := client.ListSchemas(ctx)
	if err != nil {
		return fmt.Errorf("failed to list schemas: %w", err)
	}
	
	if !response.Success {
		return fmt.Errorf("failed to list schemas: %v", response.Error)
	}
	
	schemasData := response.Data.(map[string]interface{})
	schemas := schemasData["schemas"]
	
	var testSchemas []map[string]interface{}
	
	// Find test schemas
	if schemaList, ok := schemas.([]interface{}); ok {
		for _, s := range schemaList {
			if schema, ok := s.(map[string]interface{}); ok {
				schemaName := schema["name"].(string)
				
				// Match by specific name or prefix
				if testCleanupSchemaName != "" {
					if schemaName == testCleanupSchemaName {
						testSchemas = append(testSchemas, schema)
					}
				} else if fmt.Sprintf("%s_", testCleanupPrefix) != "_" {
					if len(schemaName) > len(testCleanupPrefix) && schemaName[:len(testCleanupPrefix)] == testCleanupPrefix {
						testSchemas = append(testSchemas, schema)
					}
				}
			}
		}
	}
	
	result := map[string]interface{}{
		"action":              "cleanup_test_schemas",
		"prefix":              testCleanupPrefix,
		"schema_name":         testCleanupSchemaName,
		"dry_run":             testCleanupDryRun,
		"found_test_schemas":  len(testSchemas),
		"schemas_to_cleanup":  testSchemas,
	}
	
	if testCleanupDryRun {
		result["status"] = "dry_run_complete"
		result["message"] = fmt.Sprintf("Found %d test schemas that would be cleaned up", len(testSchemas))
	} else {
		result["status"] = "cleanup_not_implemented"
		result["message"] = "Schema deletion not yet implemented - use dry-run to see what would be cleaned"
	}
	
	return outputResult(NewSuccessResponse(result))
}

func init() {
	// Add subcommands to test command
	testCmd.AddCommand(testCreateSchemaCmd)
	testCmd.AddCommand(testCleanupCmd)
}