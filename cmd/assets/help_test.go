package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestHelpCommands tests the help system without requiring full client setup
func TestHelpCommands(t *testing.T) {
	// Build the binary for testing
	binaryPath := filepath.Join(os.TempDir(), "assets-help-test")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build test binary: %v", err)
	}
	defer os.Remove(binaryPath)
	
	tests := []struct {
		name     string
		args     []string
		contains string
	}{
		{"general help", []string{"--help"}, "Available Commands"},
		{"schema help", []string{"schema", "--help"}, "Manage asset schemas"},
		{"list help", []string{"list", "--help"}, "List asset objects"},
		{"create help", []string{"create", "--help"}, "Create new object types"},
		{"delete help", []string{"delete", "--help"}, "Delete asset object types and instances"},
		{"remove help", []string{"remove", "--help"}, "Remove specific attributes, relationships, and properties"},
		{"workflows help", []string{"workflows", "--help"}, "Explore available workflows"},
		{"browse help", []string{"browse", "--help"}, "Composite commands"},
		{"catalog help", []string{"catalog", "--help"}, "Global catalog browsers"},
		{"search help", []string{"search", "--help"}, "Search for asset objects"},
		{"complete help", []string{"complete", "--help"}, "Intelligently complete"},
		{"validate help", []string{"validate", "--help"}, "Validate object properties"},
		{"config help", []string{"config", "--help"}, "Manage configuration"},
		{"resolve help", []string{"resolve", "--help"}, "Resolve between human-readable names"},
		{"trace help", []string{"trace", "--help"}, "Tools for discovering where references point"},
		{"extract help", []string{"extract", "--help"}, "Extract attributes"},
		{"apply help", []string{"apply", "--help"}, "Universal attribute application"},
		{"test help", []string{"test", "--help"}, "Tools for creating and managing test environments"},
		{"summary help", []string{"summary", "--help"}, "Composite commands that provide friendly summaries"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			output, err := cmd.CombinedOutput()
			
			// Help commands might exit with non-zero code, check output instead
			if err != nil && len(output) == 0 {
				t.Fatalf("No output from help command: %v", err)
			}
			
			if !strings.Contains(string(output), tt.contains) {
				t.Errorf("Output does not contain expected string %q\nOutput: %s", tt.contains, string(output))
			}
		})
	}
}

// TestCommandStructure tests that all expected commands are available
func TestCommandStructure(t *testing.T) {
	// Build the binary for testing
	binaryPath := filepath.Join(os.TempDir(), "assets-structure-test")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build test binary: %v", err)
	}
	defer os.Remove(binaryPath)
	
	// Get main help output
	cmd = exec.Command(binaryPath, "--help")
	output, err := cmd.CombinedOutput()
	if err != nil && len(output) == 0 {
		t.Fatalf("Failed to get help output: %v", err)
	}
	
	expectedCommands := []string{
		"apply", "attributes", "browse", "catalog", "complete", "completion",
		"config", "copy-attributes", "create", "delete", "extract", "get",
		"help", "list", "remove", "resolve", "schema", "search", "summary", "test",
		"trace", "update", "validate", "workflows",
	}
	
	helpOutput := string(output)
	
	for _, cmd := range expectedCommands {
		if !strings.Contains(helpOutput, cmd) {
			t.Errorf("Command '%s' not found in help output", cmd)
		}
	}
}

// TestSubcommandStructure tests that subcommands are properly structured
func TestSubcommandStructure(t *testing.T) {
	// Build the binary for testing
	binaryPath := filepath.Join(os.TempDir(), "assets-subcommand-test")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build test binary: %v", err)
	}
	defer os.Remove(binaryPath)
	
	subcommandTests := []struct {
		parent     string
		subcommand string
		contains   string
	}{
		{"schema", "list", "List all schemas"},
		{"schema", "get", "Get schema details"},
		{"schema", "types", "List object types"},
		{"create", "object-type", "Create object type"},
		{"create", "instance", "Create instance"},
		{"delete", "object-type", "Delete an object type"},
		{"delete", "instance", "Delete object instances"},
		{"remove", "attribute", "Remove an attribute from an object type"},
		{"remove", "relationship", "Remove a relationship from an object"},
		{"remove", "property", "Remove a property value from an object"},
		{"browse", "hierarchy", "Browse hierarchy"},
		{"browse", "children", "Browse children"},
		{"browse", "attrs", "Browse attributes"},
		{"workflows", "list", "List workflows"},
		{"workflows", "show", "Show workflow"},
		{"workflows", "simulate", "Simulate workflow"},
		{"catalog", "attributes", "Browse attributes"},
		{"extract", "attributes", "Extract attributes"},
		{"apply", "attributes", "Apply attributes"},
		{"trace", "reference", "Trace reference"},
		{"trace", "dependencies", "Trace dependencies"},
		{"test", "create-schema", "Create test schema"},
		{"test", "cleanup", "Clean up test schemas"},
		{"config", "show", "Show configuration"},
		{"config", "test", "Test connection"},
		{"resolve", "schema", "Resolve schema"},
		{"resolve", "type", "Resolve type"},
		{"resolve", "stats", "Get statistics"},
		{"summary", "completion", "Completion summary"},
		{"summary", "schema", "Schema summary"},
	}
	
	for _, tt := range subcommandTests {
		t.Run(tt.parent+" "+tt.subcommand, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.parent, tt.subcommand, "--help")
			output, err := cmd.CombinedOutput()
			
			if err != nil && len(output) == 0 {
				t.Fatalf("No output from subcommand help: %v", err)
			}
			
			if !strings.Contains(string(output), tt.contains) {
				t.Errorf("Subcommand help does not contain expected string %q\nOutput: %s", tt.contains, string(output))
			}
		})
	}
}

// TestVersionInfo tests version and build info
func TestVersionInfo(t *testing.T) {
	// Build the binary for testing
	binaryPath := filepath.Join(os.TempDir(), "assets-version-test")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build test binary: %v", err)
	}
	defer os.Remove(binaryPath)
	
	// Test general help contains version info
	cmd = exec.Command(binaryPath, "--help")
	output, err := cmd.CombinedOutput()
	if err != nil && len(output) == 0 {
		t.Fatalf("Failed to get help output: %v", err)
	}
	
	helpOutput := string(output)
	
	// Should contain project description
	if !strings.Contains(helpOutput, "command-line tool for managing Atlassian Assets") {
		t.Error("Help output should contain project description")
	}
	
	// Should contain MCP reference
	if !strings.Contains(helpOutput, "MCP") {
		t.Error("Help output should mention MCP (Model Context Protocol)")
	}
}