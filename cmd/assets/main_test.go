package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aaronsb/atlassian-assets/internal/client"
	"github.com/aaronsb/atlassian-assets/internal/config"
)

// TestEnvironment holds the test environment state
type TestEnvironment struct {
	Client       *client.AssetsClient
	TestSchema   *TestSchema
	TestObjects  map[string]*TestObject
	BinaryPath   string
	WorkspaceID  string
}

// TestSchema represents a test schema
type TestSchema struct {
	ID          string
	Name        string
	Key         string
	ObjectTypes map[string]*TestObjectType
}

// TestObjectType represents a test object type
type TestObjectType struct {
	ID          string
	Name        string
	Description string
	SchemaID    string
	Objects     []string // Object IDs
}

// TestObject represents a test object instance
type TestObject struct {
	ID         string
	Name       string
	TypeID     string
	SchemaID   string
	Attributes map[string]interface{}
}

// Global test environment
var testEnv *TestEnvironment

// TestMain handles test setup and teardown
func TestMain(m *testing.M) {
	var err error
	
	// Build the binary for testing
	binaryPath := filepath.Join(os.TempDir(), "assets-test")
	if err := buildBinary(binaryPath); err != nil {
		fmt.Printf("Failed to build test binary: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(binaryPath)
	
	// Create test environment
	testEnv, err = setupTestEnvironment(binaryPath)
	if err != nil {
		fmt.Printf("Failed to setup test environment: %v\n", err)
		os.Exit(1)
	}
	
	// Run tests
	code := m.Run()
	
	// Cleanup
	if err := cleanupTestEnvironment(testEnv); err != nil {
		fmt.Printf("Failed to cleanup test environment: %v\n", err)
	}
	
	os.Exit(code)
}

// buildBinary builds the assets CLI binary for testing
func buildBinary(binaryPath string) error {
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "."
	return cmd.Run()
}

// setupTestEnvironment creates a test environment with schema and sample data
func setupTestEnvironment(binaryPath string) (*TestEnvironment, error) {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	
	// Create client
	assetsClient, err := client.NewAssetsClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	
	env := &TestEnvironment{
		Client:      assetsClient,
		TestObjects: make(map[string]*TestObject),
		BinaryPath:  binaryPath,
		WorkspaceID: cfg.WorkspaceID,
	}
	
	// Create test schema
	ctx := context.Background()
	
	// Note: We need to fix the schema creation API issue first
	// For now, we'll skip schema creation and use existing schemas for testing
	env.TestSchema = &TestSchema{
		ID:          "6", // Use Facilities schema for testing
		Name:        "Facilities",
		Key:         "FAC",
		ObjectTypes: make(map[string]*TestObjectType),
	}
	
	// Get object types in the test schema
	if err := env.loadObjectTypes(ctx); err != nil {
		return nil, fmt.Errorf("failed to load object types: %w", err)
	}
	
	return env, nil
}

// loadObjectTypes loads object types from the test schema
func (env *TestEnvironment) loadObjectTypes(ctx context.Context) error {
	response, err := env.Client.GetObjectTypes(ctx, env.TestSchema.ID)
	if err != nil {
		return fmt.Errorf("failed to get object types: %w", err)
	}
	
	if !response.Success {
		return fmt.Errorf("failed to get object types: %v", response.Error)
	}
	
	data := response.Data.(map[string]interface{})
	objectTypes := data["object_types"].([]interface{})
	
	for _, ot := range objectTypes {
		objType := ot.(map[string]interface{})
		typeID := objType["id"].(string)
		typeName := objType["name"].(string)
		
		// Only track a few object types for testing
		if typeName == "Offices" || typeName == "Rooms" || typeName == "HVAC systems" {
			env.TestSchema.ObjectTypes[typeName] = &TestObjectType{
				ID:          typeID,
				Name:        typeName,
				Description: fmt.Sprintf("Test object type: %s", typeName),
				SchemaID:    env.TestSchema.ID,
				Objects:     []string{},
			}
		}
	}
	
	return nil
}

// cleanupTestEnvironment cleans up the test environment
func cleanupTestEnvironment(env *TestEnvironment) error {
	if env == nil {
		return nil
	}
	
	// Close client
	if env.Client != nil {
		env.Client.Close()
	}
	
	// Note: In a real implementation, we would delete the test schema
	// For now, we're using existing schemas so no cleanup needed
	
	return nil
}

// execCommand executes a CLI command and returns the output
func (env *TestEnvironment) execCommand(args ...string) ([]byte, error) {
	cmd := exec.Command(env.BinaryPath, args...)
	return cmd.CombinedOutput()
}

// execCommandWithTimeout executes a CLI command with timeout
func (env *TestEnvironment) execCommandWithTimeout(timeout time.Duration, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, env.BinaryPath, args...)
	return cmd.CombinedOutput()
}

// parseJSONResponse parses JSON response from CLI command
func parseJSONResponse(output []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return result, nil
}

// assertSuccess checks if the command was successful
func assertSuccess(t *testing.T, output []byte, err error) map[string]interface{} {
	t.Helper()
	
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, string(output))
	}
	
	result, parseErr := parseJSONResponse(output)
	if parseErr != nil {
		t.Fatalf("Failed to parse JSON response: %v\nOutput: %s", parseErr, string(output))
	}
	
	success, ok := result["success"].(bool)
	if !ok || !success {
		t.Fatalf("Command was not successful: %v\nOutput: %s", result, string(output))
	}
	
	return result
}

// assertContains checks if output contains expected string
func assertContains(t *testing.T, output []byte, expected string) {
	t.Helper()
	
	if !strings.Contains(string(output), expected) {
		t.Fatalf("Output does not contain expected string %q\nOutput: %s", expected, string(output))
	}
}

// assertNotContains checks if output does not contain unexpected string
func assertNotContains(t *testing.T, output []byte, unexpected string) {
	t.Helper()
	
	if strings.Contains(string(output), unexpected) {
		t.Fatalf("Output contains unexpected string %q\nOutput: %s", unexpected, string(output))
	}
}

// Rate limiting helper
func rateLimit() {
	time.Sleep(500 * time.Millisecond)
}

// TestHelp tests the help system
func TestHelp(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains string
	}{
		{"general help", []string{"--help"}, "Available Commands"},
		{"schema help", []string{"schema", "--help"}, "Manage asset schemas"},
		{"list help", []string{"list", "--help"}, "List asset objects"},
		{"create help", []string{"create", "--help"}, "Create assets"},
		{"delete help", []string{"delete", "--help"}, "Delete assets"},
		{"remove help", []string{"remove", "--help"}, "Remove attributes"},
		{"workflows help", []string{"workflows", "--help"}, "Explore available workflows"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := testEnv.execCommand(tt.args...)
			if err != nil {
				// Help commands might exit with non-zero code, check output instead
				if len(output) == 0 {
					t.Fatalf("No output from help command: %v", err)
				}
			}
			
			assertContains(t, output, tt.contains)
			rateLimit()
		})
	}
}

// TestSchemaOperations tests schema-related operations
func TestSchemaOperations(t *testing.T) {
	t.Run("list schemas", func(t *testing.T) {
		output, err := testEnv.execCommand("schema", "list")
		result := assertSuccess(t, output, err)
		
		// Check that we have schemas
		data := result["data"].(map[string]interface{})
		schemas := data["schemas"].([]interface{})
		
		if len(schemas) == 0 {
			t.Error("Expected at least one schema")
		}
		
		// Check that our test schema exists
		found := false
		for _, s := range schemas {
			schema := s.(map[string]interface{})
			if schema["id"].(string) == testEnv.TestSchema.ID {
				found = true
				break
			}
		}
		
		if !found {
			t.Errorf("Test schema %s not found in schema list", testEnv.TestSchema.ID)
		}
		
		rateLimit()
	})
	
	t.Run("get schema details", func(t *testing.T) {
		output, err := testEnv.execCommand("schema", "get", "--id", testEnv.TestSchema.ID)
		result := assertSuccess(t, output, err)
		
		// Check schema details
		data := result["data"].(map[string]interface{})
		if data["id"].(string) != testEnv.TestSchema.ID {
			t.Errorf("Expected schema ID %s, got %s", testEnv.TestSchema.ID, data["id"].(string))
		}
		
		rateLimit()
	})
	
	t.Run("list object types", func(t *testing.T) {
		output, err := testEnv.execCommand("schema", "types", "--schema", testEnv.TestSchema.ID)
		result := assertSuccess(t, output, err)
		
		// Check that we have object types
		data := result["data"].(map[string]interface{})
		objectTypes := data["object_types"].([]interface{})
		
		if len(objectTypes) == 0 {
			t.Error("Expected at least one object type")
		}
		
		rateLimit()
	})
}

// TestObjectOperations tests object-related operations
func TestObjectOperations(t *testing.T) {
	t.Run("list objects", func(t *testing.T) {
		output, err := testEnv.execCommand("list", "--schema", testEnv.TestSchema.ID)
		result := assertSuccess(t, output, err)
		
		// Check the structure (even if empty)
		data := result["data"].(map[string]interface{})
		if _, ok := data["objects"]; !ok {
			t.Error("Expected 'objects' field in response")
		}
		
		if _, ok := data["total"]; !ok {
			t.Error("Expected 'total' field in response")
		}
		
		rateLimit()
	})
	
	t.Run("search objects", func(t *testing.T) {
		query := fmt.Sprintf("objectSchemaId = %s", testEnv.TestSchema.ID)
		output, err := testEnv.execCommand("search", "--query", query)
		result := assertSuccess(t, output, err)
		
		// Check search results structure
		data := result["data"].(map[string]interface{})
		if _, ok := data["objects"]; !ok {
			t.Error("Expected 'objects' field in search response")
		}
		
		if _, ok := data["total"]; !ok {
			t.Error("Expected 'total' field in search response")
		}
		
		rateLimit()
	})
}

// TestAttributeOperations tests attribute-related operations
func TestAttributeOperations(t *testing.T) {
	// Get a test object type
	var testTypeID string
	for _, objType := range testEnv.TestSchema.ObjectTypes {
		testTypeID = objType.ID
		break
	}
	
	if testTypeID == "" {
		t.Skip("No object types available for testing")
	}
	
	t.Run("get object type attributes", func(t *testing.T) {
		output, err := testEnv.execCommand("attributes", "--type", testTypeID)
		result := assertSuccess(t, output, err)
		
		// Check attributes structure
		data := result["data"].(map[string]interface{})
		if _, ok := data["attributes"]; !ok {
			t.Error("Expected 'attributes' field in response")
		}
		
		rateLimit()
	})
	
	t.Run("catalog attributes", func(t *testing.T) {
		output, err := testEnv.execCommand("catalog", "attributes", "--per-page", "5")
		result := assertSuccess(t, output, err)
		
		// Check catalog structure
		data := result["data"].(map[string]interface{})
		if _, ok := data["attributes"]; !ok {
			t.Error("Expected 'attributes' field in catalog response")
		}
		
		rateLimit()
	})
}

// TestBrowseOperations tests browse operations
func TestBrowseOperations(t *testing.T) {
	t.Run("browse hierarchy", func(t *testing.T) {
		output, err := testEnv.execCommand("browse", "hierarchy", "--schema", testEnv.TestSchema.ID)
		result := assertSuccess(t, output, err)
		
		// Check hierarchy structure
		data := result["data"].(map[string]interface{})
		if _, ok := data["hierarchy"]; !ok {
			t.Error("Expected 'hierarchy' field in browse response")
		}
		
		rateLimit()
	})
}

// TestWorkflowOperations tests workflow and intelligence features
func TestWorkflowOperations(t *testing.T) {
	t.Run("list workflows", func(t *testing.T) {
		output, err := testEnv.execCommand("workflows", "list")
		result := assertSuccess(t, output, err)
		
		// Check workflows structure
		data := result["data"].(map[string]interface{})
		if _, ok := data["workflows"]; !ok {
			t.Error("Expected 'workflows' field in response")
		}
		
		workflows := data["workflows"].(map[string]interface{})
		expectedWorkflows := []string{"object_type_creation", "attribute_marketplace", "discovery_and_analysis", "instance_management"}
		
		for _, expected := range expectedWorkflows {
			if _, ok := workflows[expected]; !ok {
				t.Errorf("Expected workflow '%s' not found", expected)
			}
		}
		
		rateLimit()
	})
	
	t.Run("simulate workflow context", func(t *testing.T) {
		variables := `{"success":true,"object_type_name":"Test","schema_id":"6"}`
		output, err := testEnv.execCommand("workflows", "simulate", "--context", "create_object_type", "--variables", variables)
		result := assertSuccess(t, output, err)
		
		// Check simulation results
		data := result["data"].(map[string]interface{})
		if _, ok := data["hints"]; !ok {
			t.Error("Expected 'hints' field in simulation response")
		}
		
		rateLimit()
	})
}

// TestConfigOperations tests configuration operations
func TestConfigOperations(t *testing.T) {
	t.Run("show config", func(t *testing.T) {
		output, err := testEnv.execCommand("config", "show")
		result := assertSuccess(t, output, err)
		
		// Check config structure
		data := result["data"].(map[string]interface{})
		if _, ok := data["workspace_id"]; !ok {
			t.Error("Expected 'workspace_id' field in config response")
		}
		
		rateLimit()
	})
	
	t.Run("test connection", func(t *testing.T) {
		output, err := testEnv.execCommand("config", "test")
		result := assertSuccess(t, output, err)
		
		// Check connection test
		data := result["data"].(map[string]interface{})
		if _, ok := data["connection_status"]; !ok {
			t.Error("Expected 'connection_status' field in test response")
		}
		
		rateLimit()
	})
}

// TestValidationOperations tests validation operations
func TestValidationOperations(t *testing.T) {
	// Get a test object type
	var testTypeID string
	for _, objType := range testEnv.TestSchema.ObjectTypes {
		testTypeID = objType.ID
		break
	}
	
	if testTypeID == "" {
		t.Skip("No object types available for testing")
	}
	
	t.Run("validate object data", func(t *testing.T) {
		testData := `{"name":"Test Object"}`
		output, err := testEnv.execCommand("validate", "--type", testTypeID, "--data", testData)
		result := assertSuccess(t, output, err)
		
		// Check validation results
		data := result["data"].(map[string]interface{})
		if _, ok := data["validation_result"]; !ok {
			t.Error("Expected 'validation_result' field in validation response")
		}
		
		rateLimit()
	})
}

// TestCompletionOperations tests completion operations
func TestCompletionOperations(t *testing.T) {
	// Get a test object type
	var testTypeID string
	for _, objType := range testEnv.TestSchema.ObjectTypes {
		testTypeID = objType.ID
		break
	}
	
	if testTypeID == "" {
		t.Skip("No object types available for testing")
	}
	
	t.Run("complete object data", func(t *testing.T) {
		testData := `{"name":"Test Completion"}`
		output, err := testEnv.execCommand("complete", "--type", testTypeID, "--data", testData)
		result := assertSuccess(t, output, err)
		
		// Check completion results
		data := result["data"].(map[string]interface{})
		if _, ok := data["completion_result"]; !ok {
			t.Error("Expected 'completion_result' field in completion response")
		}
		
		rateLimit()
	})
}

// TestResolverOperations tests resolver operations
func TestResolverOperations(t *testing.T) {
	t.Run("resolve schema by name", func(t *testing.T) {
		output, err := testEnv.execCommand("resolve", "schema", "--name", testEnv.TestSchema.Name)
		result := assertSuccess(t, output, err)
		
		// Check resolver results
		data := result["data"].(map[string]interface{})
		if _, ok := data["resolved_id"]; !ok {
			t.Error("Expected 'resolved_id' field in resolver response")
		}
		
		rateLimit()
	})
	
	t.Run("get resolver stats", func(t *testing.T) {
		output, err := testEnv.execCommand("resolve", "stats")
		result := assertSuccess(t, output, err)
		
		// Check stats structure
		data := result["data"].(map[string]interface{})
		if _, ok := data["cache_stats"]; !ok {
			t.Error("Expected 'cache_stats' field in stats response")
		}
		
		rateLimit()
	})
}

// TestErrorHandling tests error handling
func TestErrorHandling(t *testing.T) {
	t.Run("invalid command", func(t *testing.T) {
		output, err := testEnv.execCommand("invalid-command")
		if err == nil {
			t.Error("Expected error for invalid command")
		}
		
		// Should show help or error message
		if len(output) == 0 {
			t.Error("Expected output for invalid command")
		}
		
		rateLimit()
	})
	
	t.Run("missing required flag", func(t *testing.T) {
		output, err := testEnv.execCommand("list")
		if err == nil {
			t.Error("Expected error for missing required flag")
		}
		
		// Should show usage or error message
		assertContains(t, output, "schema")
		
		rateLimit()
	})
	
	t.Run("invalid schema ID", func(t *testing.T) {
		output, err := testEnv.execCommand("schema", "get", "--id", "invalid-schema-id")
		if err == nil {
			// Command might not fail, but should show error in response
			result, parseErr := parseJSONResponse(output)
			if parseErr == nil {
				if success, ok := result["success"].(bool); ok && success {
					t.Error("Expected failure for invalid schema ID")
				}
			}
		}
		
		rateLimit()
	})
}

// TestContextualHints tests contextual hints functionality
func TestContextualHints(t *testing.T) {
	t.Run("hints in successful operations", func(t *testing.T) {
		output, err := testEnv.execCommand("schema", "list")
		result := assertSuccess(t, output, err)
		
		// Check for next_steps hints
		if nextSteps, ok := result["next_steps"]; ok {
			hints := nextSteps.([]interface{})
			if len(hints) == 0 {
				t.Error("Expected contextual hints in successful operation")
			}
		}
		
		rateLimit()
	})
	
	t.Run("hints in workflow operations", func(t *testing.T) {
		output, err := testEnv.execCommand("workflows", "list")
		result := assertSuccess(t, output, err)
		
		// Workflow operations should include hints
		if nextSteps, ok := result["next_steps"]; ok {
			hints := nextSteps.([]interface{})
			if len(hints) == 0 {
				t.Error("Expected contextual hints in workflow operation")
			}
		}
		
		rateLimit()
	})
}

// BenchmarkOperations benchmarks key operations
func BenchmarkOperations(b *testing.B) {
	if testEnv == nil {
		b.Skip("Test environment not available")
	}
	
	benchmarks := []struct {
		name string
		args []string
	}{
		{"schema list", []string{"schema", "list"}},
		{"list objects", []string{"list", "--schema", testEnv.TestSchema.ID}},
		{"search objects", []string{"search", "--query", fmt.Sprintf("objectSchemaId = %s", testEnv.TestSchema.ID)}},
		{"workflows list", []string{"workflows", "list"}},
	}
	
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				output, err := testEnv.execCommand(bm.args...)
				if err != nil {
					b.Fatalf("Benchmark failed: %v\nOutput: %s", err, string(output))
				}
				
				// Rate limiting for benchmarks
				if i < b.N-1 {
					time.Sleep(100 * time.Millisecond)
				}
			}
		})
	}
}

// TestDeleteOperations tests delete operations with safety checks
func TestDeleteOperations(t *testing.T) {
	t.Run("delete without permission", func(t *testing.T) {
		// Should fail when ATLASSIAN_ASSETS_ALLOW_DELETE is not set
		output, err := testEnv.execCommand("delete", "object-type", "--id", "123", "--confirm")
		if err == nil {
			t.Error("Expected error for delete without permission")
		}
		
		// Should contain message about environment variable
		assertContains(t, output, "ATLASSIAN_ASSETS_ALLOW_DELETE")
		
		rateLimit()
	})
	
	t.Run("delete help structure", func(t *testing.T) {
		output, err := testEnv.execCommand("delete", "--help")
		if err != nil && len(output) == 0 {
			t.Fatalf("No output from delete help: %v", err)
		}
		
		// Should contain subcommands
		assertContains(t, output, "object-type")
		assertContains(t, output, "instance")
		assertContains(t, output, "permanent deletion")
		
		rateLimit()
	})
	
	t.Run("delete object-type help", func(t *testing.T) {
		output, err := testEnv.execCommand("delete", "object-type", "--help")
		if err != nil && len(output) == 0 {
			t.Fatalf("No output from delete object-type help: %v", err)
		}
		
		// Should contain safety warnings
		assertContains(t, output, "WARNING")
		assertContains(t, output, "cannot be undone")
		
		rateLimit()
	})
	
	t.Run("delete instance help", func(t *testing.T) {
		output, err := testEnv.execCommand("delete", "instance", "--help")
		if err != nil && len(output) == 0 {
			t.Fatalf("No output from delete instance help: %v", err)
		}
		
		// Should contain multiple deletion options
		assertContains(t, output, "comma-separated")
		assertContains(t, output, "AQL query")
		
		rateLimit()
	})
}

// TestRemoveOperations tests remove operations
func TestRemoveOperations(t *testing.T) {
	t.Run("remove help structure", func(t *testing.T) {
		output, err := testEnv.execCommand("remove", "--help")
		if err != nil && len(output) == 0 {
			t.Fatalf("No output from remove help: %v", err)
		}
		
		// Should contain subcommands
		assertContains(t, output, "attribute")
		assertContains(t, output, "relationship")
		assertContains(t, output, "property")
		assertContains(t, output, "modifies existing entities")
		
		rateLimit()
	})
	
	t.Run("remove attribute help", func(t *testing.T) {
		output, err := testEnv.execCommand("remove", "attribute", "--help")
		if err != nil && len(output) == 0 {
			t.Fatalf("No output from remove attribute help: %v", err)
		}
		
		// Should contain required flags
		assertContains(t, output, "type-id")
		assertContains(t, output, "attribute-id")
		assertContains(t, output, "attribute-name")
		
		rateLimit()
	})
	
	t.Run("remove relationship help", func(t *testing.T) {
		output, err := testEnv.execCommand("remove", "relationship", "--help")
		if err != nil && len(output) == 0 {
			t.Fatalf("No output from remove relationship help: %v", err)
		}
		
		// Should contain relationship options
		assertContains(t, output, "object-id")
		assertContains(t, output, "relationship-id")
		assertContains(t, output, "relationship-type")
		
		rateLimit()
	})
	
	t.Run("remove property help", func(t *testing.T) {
		output, err := testEnv.execCommand("remove", "property", "--help")
		if err != nil && len(output) == 0 {
			t.Fatalf("No output from remove property help: %v", err)
		}
		
		// Should contain property options
		assertContains(t, output, "property-name")
		assertContains(t, output, "property-id")
		assertContains(t, output, "comma-separated")
		
		rateLimit()
	})
	
	t.Run("remove missing required flags", func(t *testing.T) {
		// Should fail when required flags are missing
		output, err := testEnv.execCommand("remove", "attribute", "--confirm")
		if err == nil {
			t.Error("Expected error for missing required flags")
		}
		
		// Should show usage information
		assertContains(t, output, "type-id")
		
		rateLimit()
	})
}

// TestDeleteRemoveContextualHints tests contextual hints for delete/remove operations
func TestDeleteRemoveContextualHints(t *testing.T) {
	t.Run("delete hints in workflows", func(t *testing.T) {
		output, err := testEnv.execCommand("workflows", "simulate", "--context", "delete_object_type", "--variables", `{"success":true,"object_type_id":"123","force":false}`)
		result := assertSuccess(t, output, err)
		
		// Check for contextual hints related to deletion
		data := result["data"].(map[string]interface{})
		if hints, ok := data["hints"]; ok {
			hintsArray := hints.([]interface{})
			if len(hintsArray) == 0 {
				t.Error("Expected contextual hints for delete operation")
			}
			
			// Should contain warning about cascading deletion
			hintsStr := fmt.Sprintf("%v", hintsArray)
			if !strings.Contains(hintsStr, "Warning") {
				t.Log("Expected warning hints for delete operation")
			}
		}
		
		rateLimit()
	})
	
	t.Run("remove hints in workflows", func(t *testing.T) {
		output, err := testEnv.execCommand("workflows", "simulate", "--context", "remove_attribute", "--variables", `{"success":true,"type_id":"123","attribute_id":"456"}`)
		result := assertSuccess(t, output, err)
		
		// Check for contextual hints related to removal
		data := result["data"].(map[string]interface{})
		if hints, ok := data["hints"]; ok {
			hintsArray := hints.([]interface{})
			if len(hintsArray) == 0 {
				t.Error("Expected contextual hints for remove operation")
			}
			
			// Should contain guidance about validation
			hintsStr := fmt.Sprintf("%v", hintsArray)
			if !strings.Contains(hintsStr, "Validate") {
				t.Log("Expected validation hints for remove operation")
			}
		}
		
		rateLimit()
	})
}

// TestDeleteRemoveErrorHandling tests error handling for delete/remove operations
func TestDeleteRemoveErrorHandling(t *testing.T) {
	t.Run("delete missing confirmation", func(t *testing.T) {
		// Should fail without confirmation flags
		output, err := testEnv.execCommand("delete", "object-type", "--id", "123")
		if err == nil {
			t.Error("Expected error for missing confirmation")
		}
		
		// Should mention confirmation requirement
		assertContains(t, output, "confirm")
		
		rateLimit()
	})
	
	t.Run("remove missing confirmation", func(t *testing.T) {
		// Should fail without confirmation flags  
		output, err := testEnv.execCommand("remove", "attribute", "--type-id", "123", "--attribute-id", "456")
		if err == nil {
			t.Error("Expected error for missing confirmation")
		}
		
		// Should mention confirmation requirement
		assertContains(t, output, "confirm")
		
		rateLimit()
	})
	
	t.Run("invalid delete parameters", func(t *testing.T) {
		// Should fail with invalid parameters
		output, err := testEnv.execCommand("delete", "object-type", "--confirm")
		if err == nil {
			t.Error("Expected error for missing object type ID")
		}
		
		// Should mention required parameters
		assertContains(t, output, "id")
		
		rateLimit()
	})
}