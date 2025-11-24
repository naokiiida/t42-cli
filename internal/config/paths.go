package config

import (
	"os"
	"path/filepath"
)

const (
	// AppName is the name of the application used for config directories
	AppName = "t42"

	// ConfigFileName is the name of the user configuration file
	ConfigFileName = "config.yaml"

	// CredentialsFileName is the name of the credentials file
	CredentialsFileName = "credentials.json"

	// SecretsFileName is the name of the OAuth2 client secrets file
	SecretsFileName = "secrets.env"

	// SecretDirName is the name of the development secrets directory
	SecretDirName = "secret"

	// EnvFileName is the name of the environment file for development
	EnvFileName = ".env"
)

// GetConfigDir returns the OS-specific configuration directory for the application.
// If T42_ENV is set to "development", it returns the local secret directory.
func GetConfigDir() (string, error) {
	if os.Getenv("T42_ENV") == "development" {
		// For development, use the local secret directory
		return SecretDirName, nil
	}
	
	// Get the OS-specific user config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	
	// Return the app-specific subdirectory
	return filepath.Join(configDir, AppName), nil
}

// GetConfigFilePath returns the full path to the user configuration file
func GetConfigFilePath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, ConfigFileName), nil
}

// GetCredentialsFilePath returns the full path to the credentials file
func GetCredentialsFilePath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, CredentialsFileName), nil
}

// GetSecretsFilePath returns the full path to the OAuth2 client secrets file
// in the user's config directory (for deployed/production use)
func GetSecretsFilePath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, SecretsFileName), nil
}

// GetDevelopmentEnvFilePath returns the path to the development .env file
// (for local development use)
func GetDevelopmentEnvFilePath() string {
	return filepath.Join(SecretDirName, EnvFileName)
}

// EnsureConfigDir creates the configuration directory if it doesn't exist
func EnsureConfigDir() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	return os.MkdirAll(configDir, 0755)
}