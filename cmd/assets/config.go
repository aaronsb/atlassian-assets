package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// CONFIG command with subcommands
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long: `Manage configuration profiles and settings.
	
Configuration can be set via environment variables or configuration files.`,
}

// CONFIG SHOW subcommand
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long: `Display the current configuration settings.
	
This shows the active configuration including profile, host, and other settings.
Sensitive information like tokens are masked.`,
	Example: `  # Show current config
  assets config show`,
	RunE: runConfigShowCmd,
}

func runConfigShowCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	cfg := client.GetConfig()
	
	// Mask sensitive information
	maskedToken := cfg.APIToken
	if len(maskedToken) > 8 {
		maskedToken = maskedToken[:4] + "..." + maskedToken[len(maskedToken)-4:]
	}

	response := NewSuccessResponse(map[string]interface{}{
		"email":        cfg.Email,
		"host":         cfg.Host,
		"api_token":    maskedToken,
		"workspace_id": cfg.WorkspaceID,
		"profile":      cfg.Profile,
	})

	return outputResult(response)
}

// CONFIG TEST subcommand
var configTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test configuration and connection",
	Long: `Test the current configuration by attempting to connect to the Atlassian instance.
	
This verifies that the credentials are valid and the service is reachable.`,
	Example: `  # Test connection
  assets config test`,
	RunE: runConfigTestCmd,
}

func runConfigTestCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	// Test connection
	// TODO: Implement actual connection test
	response := NewSuccessResponse(map[string]interface{}{
		"action": "config_test",
		"status": "not_implemented",
		"message": "Connection test will be implemented using go-atlassian SDK",
	})

	return outputResult(response)
}

// CONFIG DISCOVER-WORKSPACE subcommand
var configDiscoverWorkspaceCmd = &cobra.Command{
	Use:   "discover-workspace",
	Short: "Discover workspace ID",
	Long: `Attempt to discover the Assets workspace ID from the configured Atlassian instance.
	
This addresses the abstraction between site name and workspace UID.`,
	Example: `  # Discover workspace
  assets config discover-workspace`,
	RunE: runConfigDiscoverWorkspaceCmd,
}

func runConfigDiscoverWorkspaceCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	// TODO: Implement workspace discovery
	response := NewSuccessResponse(map[string]interface{}{
		"action": "discover_workspace",
		"status": "not_implemented",
		"message": "Workspace discovery will be implemented using go-atlassian SDK",
	})

	return outputResult(response)
}

func init() {
	// Add subcommands to config command
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configTestCmd)
	configCmd.AddCommand(configDiscoverWorkspaceCmd)
}