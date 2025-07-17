package client

import (
	"context"
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

	// Create the assets API client
	assetsAPI, err := assets.New(httpClient, cfg.GetBaseURL())
	if err != nil {
		return nil, fmt.Errorf("failed to create assets client: %w", err)
	}

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
	// Try to get workspaces/objects to discover the workspace ID
	// This is a common pattern when the workspace ID isn't known
	
	// Note: The exact method depends on the go-atlassian API structure
	// We may need to call a general endpoint that returns workspace info
	
	// For now, return an error asking the user to provide the workspace ID
	// We can implement auto-discovery once we test with a real instance
	return "", fmt.Errorf("workspace ID discovery not yet implemented - please set ATLASSIAN_ASSETS_WORKSPACE_ID environment variable")
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