package config

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// ErrNotLoggedIn is returned by LoadCredentials when the credentials file does not exist,
// indicating that the user needs to authenticate.
var ErrNotLoggedIn = errors.New("not logged in, please run 't42 auth login'")

// Credentials represents the OAuth2 token data stored securely on disk.
// It is designed to be long-lived and includes the refresh token.
type Credentials struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	// ExpiresAt is the absolute time when the access token expires.
	// This is calculated as time.Now() + expires_in when the token is received.
	ExpiresAt time.Time `json:"expires_at"`
}

// IsAccessTokenExpired checks if the access token has expired or will expire within the next minute.
// This gives a small buffer to avoid using a token right before it expires.
func (c *Credentials) IsAccessTokenExpired() bool {
	return time.Now().After(c.ExpiresAt.Add(-1 * time.Minute))
}

// Preferences represents user-configurable settings stored in config.yaml.
type Preferences struct {
	// This struct is reserved for future user-specific preferences.
	// Example:
	// DefaultCampus string `yaml:"default_campus"`
}

// SaveCredentials marshals the credentials to JSON and writes them to the
// credentials file with secure 0600 permissions.
func SaveCredentials(creds *Credentials) error {
	path, err := GetCredentialsPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}

	// Write with 0600 permissions to ensure only the user can read/write it.
	return os.WriteFile(path, data, 0600)
}

// LoadCredentials reads and unmarshals the credentials from the credentials file.
// It returns ErrNotLoggedIn if the file does not exist.
func LoadCredentials() (*Credentials, error) {
	path, err := GetCredentialsPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotLoggedIn
		}
		return nil, err
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}

	return &creds, nil
}

// DeleteCredentials removes the credentials file from the disk.
// It does not return an error if the file is already absent.
func DeleteCredentials() error {
	path, err := GetCredentialsPath()
	if err != nil {
		return err
	}

	// If the file doesn't exist, that's a success case for a logout operation.
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// LoadPreferences reads and unmarshals user preferences from the config file.
// If the file doesn't exist, it returns a default (empty) Preferences struct.
func LoadPreferences() (*Preferences, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	prefs := &Preferences{}

	data, err := os.ReadFile(path)
	if err != nil {
		// If the config file doesn't exist, it's not an error; we just use the defaults.
		if os.IsNotExist(err) {
			return prefs, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, prefs); err != nil {
		return nil, err
	}

	return prefs, nil
}

// SavePreferences marshals the preferences to YAML and writes them to the config file.
func SavePreferences(prefs *Preferences) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(prefs)
	if err != nil {
		return err
	}

	// Write with standard 0644 permissions.
	return os.WriteFile(path, data, 0644)
}

// LoadDotEnv loads environment variables from `secret/.env` if it exists.
// This is intended for development use and does not return an error if the file is missing.
func LoadDotEnv() {
	// godotenv.Load searches for a .env file. We specify the path to ensure
	// it only loads the one from the project's `secret` directory.
	// The error is ignored because the file is optional and only used for development.
	_ = godotenv.Load("secret/.env")
}