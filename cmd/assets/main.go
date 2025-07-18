package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/aaronsb/atlassian-assets/internal/config"
	"github.com/aaronsb/atlassian-assets/internal/client"
)

var (
	cfgFile     string
	profile     string
	workspaceID string
	output      string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "assets",
	Short: "Atlassian Assets CLI tool",
	Long: `A command-line tool for managing Atlassian Assets.
	
This tool provides CRUD operations for Atlassian Assets and is designed
to serve as a prototype for an MCP (Model Context Protocol) interface
for AI agents.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.config/atlassian-assets/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&profile, "profile", "", "configuration profile to use")
	rootCmd.PersistentFlags().StringVar(&workspaceID, "workspace-id", "", "Atlassian Assets workspace ID")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "json", "output format (json, yaml, table)")

	// Add subcommands
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(attributesCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(completeCmd)
	rootCmd.AddCommand(copyAttributesCmd)
	rootCmd.AddCommand(browseCmd)
	rootCmd.AddCommand(summaryCmd)
	rootCmd.AddCommand(schemaCmd)
	rootCmd.AddCommand(traceCmd)
	rootCmd.AddCommand(extractCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(catalogCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(resolveCmd)
	rootCmd.AddCommand(workflowsCmd)
	rootCmd.AddCommand(testCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// This will be called before any command runs
	// We can initialize configuration here
}

// getClient creates and returns a configured Assets client
func getClient() (*client.AssetsClient, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override with command-line flags if provided
	if workspaceID != "" {
		cfg.WorkspaceID = workspaceID
	}
	if profile != "" {
		cfg.Profile = profile
	}

	client, err := client.NewAssetsClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return client, nil
}

// Response wrapper for consistent output
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// NewSuccessResponse creates a successful response
func NewSuccessResponse(data interface{}) *Response {
	return &Response{
		Success: true,
		Data:    data,
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(err error) *Response {
	return &Response{
		Success: false,
		Error:   err.Error(),
	}
}

// outputResult formats and outputs the result based on the output format
func outputResult(data interface{}) error {
	switch output {
	case "json":
		return outputJSON(data)
	case "yaml":
		return outputYAML(data)
	case "table":
		return outputTable(data)
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}

func outputJSON(data interface{}) error {
	// Pretty-print JSON output
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

func outputYAML(data interface{}) error {
	// YAML output implementation
	fmt.Printf("%+v\n", data)
	return nil
}

func outputTable(data interface{}) error {
	// Table output implementation
	fmt.Printf("%+v\n", data)
	return nil
}

func main() {
	Execute()
}