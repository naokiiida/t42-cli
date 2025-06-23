package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Credentials represents the OAuth2 token response from 42 API
type Credentials struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope"`
	CreatedAt    int64  `json:"created_at"`
	SecretValidUntil int64 `json:"secret_valid_until,omitempty"`
}

// Config represents user preferences and settings
type Config struct {
	DefaultFormat string `yaml:"default_format,omitempty"` // "table" or "json"
	Interactive   bool   `yaml:"interactive"`              // Enable interactive prompts
	APIBaseURL    string `yaml:"api_base_url,omitempty"`   // Custom API base URL
}

// DevelopmentSecrets represents the development environment variables
type DevelopmentSecrets struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		DefaultFormat: "table",
		Interactive:   true,
		APIBaseURL:    "https://api.intra.42.fr",
	}
}

// LoadCredentials loads the OAuth2 credentials from the credentials file
func LoadCredentials() (*Credentials, error) {
	credentialsPath, err := GetCredentialsFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials file path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(credentialsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("credentials file not found at %s", credentialsPath)
	}

	// Read the file
	data, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	// Parse JSON
	var credentials Credentials
	if err := json.Unmarshal(data, &credentials); err != nil {
		return nil, fmt.Errorf("failed to parse credentials JSON: %w", err)
	}

	return &credentials, nil
}

// SaveCredentials saves the OAuth2 credentials to the credentials file with secure permissions
func SaveCredentials(credentials *Credentials) error {
	// Ensure config directory exists
	if err := EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	credentialsPath, err := GetCredentialsFilePath()
	if err != nil {
		return fmt.Errorf("failed to get credentials file path: %w", err)
	}

	// Marshal to JSON with proper indentation
	data, err := json.MarshalIndent(credentials, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials to JSON: %w", err)
	}

	// Write file with secure permissions (0600 = read/write for user only)
	if err := os.WriteFile(credentialsPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	return nil
}

// DeleteCredentials removes the credentials file
func DeleteCredentials() error {
	credentialsPath, err := GetCredentialsFilePath()
	if err != nil {
		return fmt.Errorf("failed to get credentials file path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(credentialsPath); os.IsNotExist(err) {
		// File doesn't exist, nothing to delete
		return nil
	}

	// Remove the file
	if err := os.Remove(credentialsPath); err != nil {
		return fmt.Errorf("failed to delete credentials file: %w", err)
	}

	return nil
}

// LoadConfig loads the user configuration from the config file
func LoadConfig() (*Config, error) {
	configPath, err := GetConfigFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config file path: %w", err)
	}

	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	// Read the file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	// Fill in any missing defaults
	defaultCfg := DefaultConfig()
	if config.DefaultFormat == "" {
		config.DefaultFormat = defaultCfg.DefaultFormat
	}
	if config.APIBaseURL == "" {
		config.APIBaseURL = defaultCfg.APIBaseURL
	}

	return &config, nil
}

// SaveConfig saves the user configuration to the config file
func SaveConfig(config *Config) error {
	// Ensure config directory exists
	if err := EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath, err := GetConfigFilePath()
	if err != nil {
		return fmt.Errorf("failed to get config file path: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	// Write file with standard permissions (0644 = read/write for user, read for group/others)
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadDevelopmentSecrets loads development secrets from the .env file
func LoadDevelopmentSecrets() (*DevelopmentSecrets, error) {
	envPath := GetDevelopmentEnvFilePath()

	// Check if file exists
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("development .env file not found at %s", envPath)
	}

	// Load environment variables from file
	if err := godotenv.Load(envPath); err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	secrets := &DevelopmentSecrets{
		ClientID:     os.Getenv("FT_UID"),
		ClientSecret: os.Getenv("FT_SECRET"),
		RedirectURL:  os.Getenv("REDIRECT_URL"),
	}

	// Validate required fields
	if secrets.ClientID == "" {
		return nil, fmt.Errorf("FT_UID not found in %s", envPath)
	}
	if secrets.ClientSecret == "" {
		return nil, fmt.Errorf("FT_SECRET not found in %s", envPath)
	}

	// RedirectURL is optional for some flows
	if secrets.RedirectURL == "" {
		secrets.RedirectURL = "http://localhost:8080/callback"
	}

	return secrets, nil
}

// HasValidCredentials checks if valid credentials exist
func HasValidCredentials() bool {
	credentials, err := LoadCredentials()
	if err != nil {
		return false
	}
	
	// Basic validation - check if access token exists
	return credentials.AccessToken != ""
}