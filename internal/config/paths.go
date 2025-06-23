package config

import (
	"os"
	"path/filepath"
)

const (
	// AppName is the name of the application, used to construct configuration paths.
	AppName = "t42"
	// CredentialsFile is the name of the file storing session credentials.
	CredentialsFile = "credentials.json"
	// ConfigFile is the name of the file storing user preferences.
	ConfigFile = "config.yaml"
)

// GetConfigDir returns the OS-specific configuration directory for the application.
// It creates the directory if it doesn't exist.
//   - macOS: ~/Library/Application Support/t42
//   - Linux: ~/.config/t42
//   - Windows: %APPDATA%\t42
func GetConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appConfigDir := filepath.Join(configDir, AppName)
	// Create the directory if it does not exist.
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		return "", err
	}
	return appConfigDir, nil
}

// GetCredentialsPath returns the full path to the credentials file.
func GetCredentialsPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, CredentialsFile), nil
}

// GetConfigPath returns the full path to the user preferences file.
func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, ConfigFile), nil
}