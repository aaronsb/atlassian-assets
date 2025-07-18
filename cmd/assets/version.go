package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/aaronsb/atlassian-assets/internal/version"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long: `Display version information including build details, commit hash, and build date.

The version command shows detailed information about the current build including:
- Version number (semantic version or development build)
- Git commit hash
- Build date and time
- Go version used for compilation
- Target platform (OS/architecture)
- Author information`,
	Example: `  # Show version information
  assets version
  
  # Show version in JSON format
  assets version --output json`,
	RunE: runVersionCmd,
}

func init() {
	// Add the version command to the root command
	rootCmd.AddCommand(versionCmd)
}

func runVersionCmd(cmd *cobra.Command, args []string) error {
	info := version.GetInfo()
	
	// Check if output format is JSON
	if output == "json" {
		jsonBytes, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal version info to JSON: %w", err)
		}
		fmt.Println(string(jsonBytes))
	} else {
		// Default human-readable format
		fmt.Println(info.String())
	}
	
	return nil
}