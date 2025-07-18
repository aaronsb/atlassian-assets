package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the Assets CLI
type Config struct {
	Email         string
	Host          string
	APIToken      string
	WorkspaceID   string
	Profile       string
	CacheDir      string
	CacheTTLHours int
	AllowDelete   bool
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

	// Parse cache TTL hours
	cacheTTLHours := 24 // default
	if ttlStr := os.Getenv("ATLASSIAN_ASSETS_CACHE_TTL_HOURS"); ttlStr != "" {
		if ttl, err := strconv.Atoi(ttlStr); err == nil && ttl > 0 {
			cacheTTLHours = ttl
		}
	}

	// Parse allow delete flag
	allowDelete := false // default disabled for safety
	if deleteStr := os.Getenv("ATLASSIAN_ASSETS_ALLOW_DELETE"); deleteStr != "" {
		if delete, err := strconv.ParseBool(deleteStr); err == nil {
			allowDelete = delete
		}
	}

	config := &Config{
		Email:         os.Getenv("ATLASSIAN_EMAIL"),
		Host:          os.Getenv("ATLASSIAN_HOST"),
		APIToken:      os.Getenv("ATLASSIAN_API_TOKEN"),
		WorkspaceID:   os.Getenv("ATLASSIAN_ASSETS_WORKSPACE_ID"),
		Profile:       os.Getenv("ATLASSIAN_ASSETS_PROFILE"),
		CacheDir:      os.Getenv("ATLASSIAN_ASSETS_CACHE_DIR"),
		CacheTTLHours: cacheTTLHours,
		AllowDelete:   allowDelete,
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

	// Set default cache directory if not specified
	if config.CacheDir == "" {
		config.CacheDir = ".cache/atlassian-assets"
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

// GetCacheDir returns the cache directory path, creating it if necessary
func (c *Config) GetCacheDir() (string, error) {
	// Convert relative path to absolute if needed
	cacheDir := c.CacheDir
	if !filepath.IsAbs(cacheDir) {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
		cacheDir = filepath.Join(wd, cacheDir)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	return cacheDir, nil
}

// GetCacheTTL returns the cache TTL as a time.Duration
func (c *Config) GetCacheTTL() time.Duration {
	return time.Duration(c.CacheTTLHours) * time.Hour
}