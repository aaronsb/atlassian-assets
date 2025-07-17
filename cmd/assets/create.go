package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/aaronsb/atlassian-assets/internal/client"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new asset object",
	Long: `Create a new asset object in the specified schema.
	
This command creates a new asset object with the provided data.
The data should be provided as a JSON string containing the object attributes.`,
	Example: `  # Create a new laptop asset
  assets create --schema computers --type laptop --data '{"name":"MacBook Pro","owner":"john.doe"}'`,
	RunE: runCreateCmd,
}

var (
	createSchema string
	createType   string
	createData   string
)

func init() {
	createCmd.Flags().StringVar(&createSchema, "schema", "", "Schema ID or name (required)")
	createCmd.Flags().StringVar(&createType, "type", "", "Object type (required)")
	createCmd.Flags().StringVar(&createData, "data", "", "Object data as JSON string (required)")
	
	createCmd.MarkFlagRequired("schema")
	createCmd.MarkFlagRequired("type")
	createCmd.MarkFlagRequired("data")
}

func runCreateCmd(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	return createAsset(ctx, client, createSchema, createType, createData)
}

// createAsset creates a new asset object
func createAsset(ctx context.Context, client *client.AssetsClient, schema, objectType, data string) error {
	// 1. Parse the JSON data
	var attributes map[string]interface{}
	if err := json.Unmarshal([]byte(data), &attributes); err != nil {
		return fmt.Errorf("failed to parse JSON data: %w", err)
	}

	// 2. Create the object using our client
	response, err := client.CreateObject(ctx, objectType, attributes)
	if err != nil {
		return fmt.Errorf("failed to create object: %w", err)
	}

	// 3. Output the result
	return outputResult(response)
}