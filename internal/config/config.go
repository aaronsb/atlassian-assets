package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the Assets CLI
type Config struct {
	Email       string
	Host        string
	APIToken    string
	WorkspaceID string
	Profile     string
}

// LoadConfig loads configuration from environment variables and .env file
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// Not an error if .env doesn't exist
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
	}

	config := &Config{
		Email:       os.Getenv("ATLASSIAN_EMAIL"),
		Host:        os.Getenv("ATLASSIAN_HOST"),
		APIToken:    os.Getenv("ATLASSIAN_API_TOKEN"),
		WorkspaceID: os.Getenv("ATLASSIAN_ASSETS_WORKSPACE_ID"),
		Profile:     os.Getenv("ATLASSIAN_ASSETS_PROFILE"),
	}

	// Validate required fields
	if config.Email == "" {
		return nil, fmt.Errorf("ATLASSIAN_EMAIL is required")
	}
	if config.Host == "" {
		return nil, fmt.Errorf("ATLASSIAN_HOST is required")
	}
	if config.APIToken == "" {
		return nil, fmt.Errorf("ATLASSIAN_API_TOKEN is required")
	}

	// Set default profile if not specified
	if config.Profile == "" {
		config.Profile = "default"
	}

	return config, nil
}

// GetConfigDir returns the configuration directory path
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "atlassian-assets")
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return configDir, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Email == "" {
		return fmt.Errorf("email is required")
	}
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.APIToken == "" {
		return fmt.Errorf("API token is required")
	}
	return nil
}

// GetBaseURL returns the base URL for API calls
func (c *Config) GetBaseURL() string {
	return c.Host
}

// GetUsername returns the username for basic auth (email for Atlassian)
func (c *Config) GetUsername() string {
	return c.Email
}

// GetPassword returns the password for basic auth (API token for Atlassian)
func (c *Config) GetPassword() string {
	return c.APIToken
}