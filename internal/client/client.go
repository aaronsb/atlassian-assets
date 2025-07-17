package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ctreminiom/go-atlassian/v2/assets"
	"github.com/aaronsb/atlassian-assets/internal/config"
)

// AssetsClient wraps the go-atlassian client with our configuration
type AssetsClient struct {
	httpClient  *http.Client
	assetsAPI   *assets.Client
	config      *config.Config
	workspaceID string
}

// NewAssetsClient creates a new Assets client with the given configuration
func NewAssetsClient(cfg *config.Config) (*AssetsClient, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create HTTP client for basic auth
	httpClient := &http.Client{}

	// Create the assets API client - use default Assets API URL
	assetsAPI, err := assets.New(httpClient, "https://api.atlassian.com/")
	if err != nil {
		return nil, fmt.Errorf("failed to create assets client: %w", err)
	}

	// Set basic authentication
	assetsAPI.Auth.SetBasicAuth(cfg.GetUsername(), cfg.GetPassword())

	ac := &AssetsClient{
		httpClient: httpClient,
		assetsAPI:  assetsAPI,
		config:     cfg,
	}

	// Discover workspace ID if not provided
	if cfg.WorkspaceID == "" {
		workspaceID, err := ac.discoverWorkspaceID(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to discover workspace ID: %w", err)
		}
		ac.workspaceID = workspaceID
	} else {
		ac.workspaceID = cfg.WorkspaceID
	}

	return ac, nil
}

// discoverWorkspaceID attempts to discover the workspace ID from the site
// This addresses the abstraction issue you mentioned about site name -> workspace UID
func (ac *AssetsClient) discoverWorkspaceID(ctx context.Context) (string, error) {
	// Use the JSM Service Desk API to get workspace information
	// This is the key to bridging site name -> workspace UID
	
	url := fmt.Sprintf("%s/rest/servicedeskapi/insight/workspace", ac.config.GetBaseURL())
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create workspace discovery request: %w", err)
	}

	// Set basic auth
	req.SetBasicAuth(ac.config.GetUsername(), ac.config.GetPassword())
	req.Header.Set("Accept", "application/json")

	resp, err := ac.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to discover workspace: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("workspace discovery failed: HTTP %d", resp.StatusCode)
	}

	var workspaceResp struct {
		Values []struct {
			WorkspaceID string `json:"workspaceId"`
		} `json:"values"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&workspaceResp); err != nil {
		return "", fmt.Errorf("failed to decode workspace response: %w", err)
	}

	if len(workspaceResp.Values) == 0 {
		return "", fmt.Errorf("no workspaces found")
	}

	// Return the first workspace ID
	return workspaceResp.Values[0].WorkspaceID, nil
}

// GetWorkspaceID returns the current workspace ID
func (ac *AssetsClient) GetWorkspaceID() string {
	return ac.workspaceID
}

// TestConnection tests the connection to the Atlassian instance
func (ac *AssetsClient) TestConnection(ctx context.Context) error {
	// Try to make a simple API call to verify connectivity
	// This could be a call to get current user or workspace info
	return fmt.Errorf("connection test not yet implemented")
}

// GetAssetsAPI returns the assets API client
func (ac *AssetsClient) GetAssetsAPI() *assets.Client {
	return ac.assetsAPI
}

// GetConfig returns the current configuration
func (ac *AssetsClient) GetConfig() *config.Config {
	return ac.config
}

// Close closes the client connection
func (ac *AssetsClient) Close() error {
	// Clean up any resources if needed
	return nil
}

// WithWorkspaceID sets the workspace ID for this client
func (ac *AssetsClient) WithWorkspaceID(workspaceID string) *AssetsClient {
	ac.workspaceID = workspaceID
	return ac
}

// Common response wrapper for consistent error handling
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

// ListSchemas lists all object schemas in the workspace
func (ac *AssetsClient) ListSchemas(ctx context.Context) (*Response, error) {
	if ac.workspaceID == "" {
		return NewErrorResponse(fmt.Errorf("workspace ID not set")), nil
	}

	schemas, response, err := ac.assetsAPI.ObjectSchema.List(ctx, ac.workspaceID)
	if err != nil {
		return NewErrorResponse(fmt.Errorf("failed to list schemas: %w", err)), nil
	}

	if response.Code != 200 {
		return NewErrorResponse(fmt.Errorf("API error: %d - %s", response.Code, response.Bytes.String())), nil
	}

	return NewSuccessResponse(map[string]interface{}{
		"schemas": schemas.Values,
		"total":   schemas.Total,
	}), nil
}

// GetSchema gets details of a specific schema
func (ac *AssetsClient) GetSchema(ctx context.Context, schemaID string) (*Response, error) {
	if ac.workspaceID == "" {
		return NewErrorResponse(fmt.Errorf("workspace ID not set")), nil
	}

	schema, response, err := ac.assetsAPI.ObjectSchema.Get(ctx, ac.workspaceID, schemaID)
	if err != nil {
		return NewErrorResponse(fmt.Errorf("failed to get schema: %w", err)), nil
	}

	if response.Code != 200 {
		return NewErrorResponse(fmt.Errorf("API error: %d - %s", response.Code, response.Bytes.String())), nil
	}

	return NewSuccessResponse(schema), nil
}