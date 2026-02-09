package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestGetConfigDir(t *testing.T) {
	// Test development environment
	t.Run("development environment", func(t *testing.T) {
		if err := os.Setenv("T42_ENV", "development"); err != nil {
			t.Fatalf("Failed to set T42_ENV: %v", err)
		}
		defer func() {
			if err := os.Unsetenv("T42_ENV"); err != nil {
				t.Fatalf("Failed to unset T42_ENV: %v", err)
			}
		}()

		configDir, err := GetConfigDir()
		if err != nil {
			t.Fatalf("GetConfigDir() error = %v", err)
		}

		expected := SecretDirName
		if configDir != expected {
			t.Errorf("GetConfigDir() = %v, want %v", configDir, expected)
		}
	})

	// Test production environment
	t.Run("production environment", func(t *testing.T) {
		if err := os.Unsetenv("T42_ENV"); err != nil {
			t.Fatalf("Failed to unset T42_ENV: %v", err)
		}

		configDir, err := GetConfigDir()
		if err != nil {
			t.Fatalf("GetConfigDir() error = %v", err)
		}

		// Should contain the app name
		if !filepath.IsAbs(configDir) {
			t.Errorf("GetConfigDir() should return absolute path, got %v", configDir)
		}

		if filepath.Base(configDir) != AppName {
			t.Errorf("GetConfigDir() should end with %v, got %v", AppName, configDir)
		}
	})
}

func TestGetConfigFilePath(t *testing.T) {
	// Set development environment for predictable path
	if err := os.Setenv("T42_ENV", "development"); err != nil {
		t.Fatalf("Failed to set T42_ENV: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("T42_ENV"); err != nil {
			t.Fatalf("Failed to unset T42_ENV: %v", err)
		}
	}()

	path, err := GetConfigFilePath()
	if err != nil {
		t.Fatalf("GetConfigFilePath() error = %v", err)
	}

	expected := filepath.Join(SecretDirName, ConfigFileName)
	if path != expected {
		t.Errorf("GetConfigFilePath() = %v, want %v", path, expected)
	}
}

func TestGetCredentialsFilePath(t *testing.T) {
	// Set development environment for predictable path
	if err := os.Setenv("T42_ENV", "development"); err != nil {
		t.Fatalf("Failed to set T42_ENV: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("T42_ENV"); err != nil {
			t.Fatalf("Failed to unset T42_ENV: %v", err)
		}
	}()

	path, err := GetCredentialsFilePath()
	if err != nil {
		t.Fatalf("GetCredentialsFilePath() error = %v", err)
	}

	expected := filepath.Join(SecretDirName, CredentialsFileName)
	if path != expected {
		t.Errorf("GetCredentialsFilePath() = %v, want %v", path, expected)
	}
}

func TestGetDevelopmentEnvFilePath(t *testing.T) {
	path := GetDevelopmentEnvFilePath()
	expected := filepath.Join(SecretDirName, EnvFileName)
	if path != expected {
		t.Errorf("GetDevelopmentEnvFilePath() = %v, want %v", path, expected)
	}
}

func TestCredentialsOperations(t *testing.T) {
	// Set up test environment
	if err := os.Setenv("T42_ENV", "development"); err != nil {
		t.Fatalf("Failed to set T42_ENV: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("T42_ENV"); err != nil {
			t.Fatalf("Failed to unset T42_ENV: %v", err)
		}
	}()

	// Create test directory
	testDir := "test_secret"
	if err := os.Setenv("T42_ENV", "development"); err != nil {
		t.Fatalf("Failed to set T42_ENV: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(testDir); err != nil {
			t.Logf("Warning: failed to remove test dir: %v", err)
		}
		if err := os.Unsetenv("T42_ENV"); err != nil {
			t.Fatalf("Failed to unset T42_ENV: %v", err)
		}
	}()

	// Note: We're using the development environment which uses a predictable path

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "t42-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	// Test credentials that don't exist yet
	t.Run("load non-existent credentials", func(t *testing.T) {
		if err := DeleteCredentials(); err != nil {
			t.Logf("Warning: failed to delete credentials: %v", err)
		}
		_, err := LoadCredentials()
		if err == nil {
			t.Error("LoadCredentials() should error when file doesn't exist")
		}
	})

	// Test saving and loading credentials
	t.Run("save and load credentials", func(t *testing.T) {
		testCredentials := &Credentials{
			AccessToken:      "test_access_token",
			TokenType:        "bearer",
			ExpiresIn:        3600,
			RefreshToken:     "test_refresh_token",
			Scope:            "public",
			CreatedAt:        1640995200,
			SecretValidUntil: 1640998800,
		}

		// Save credentials
		err := SaveCredentials(testCredentials)
		if err != nil {
			t.Fatalf("SaveCredentials() error = %v", err)
		}

		// Load credentials
		loadedCredentials, err := LoadCredentials()
		if err != nil {
			t.Fatalf("LoadCredentials() error = %v", err)
		}

		// Compare credentials
		if loadedCredentials.AccessToken != testCredentials.AccessToken {
			t.Errorf("AccessToken mismatch: got %v, want %v", loadedCredentials.AccessToken, testCredentials.AccessToken)
		}
		if loadedCredentials.TokenType != testCredentials.TokenType {
			t.Errorf("TokenType mismatch: got %v, want %v", loadedCredentials.TokenType, testCredentials.TokenType)
		}
		if loadedCredentials.ExpiresIn != testCredentials.ExpiresIn {
			t.Errorf("ExpiresIn mismatch: got %v, want %v", loadedCredentials.ExpiresIn, testCredentials.ExpiresIn)
		}
	})

	// Test file permissions
	t.Run("credentials file permissions", func(t *testing.T) {
		testCredentials := &Credentials{
			AccessToken: "test_token",
			TokenType:   "bearer",
		}

		err := SaveCredentials(testCredentials)
		if err != nil {
			t.Fatalf("SaveCredentials() error = %v", err)
		}

		credentialsPath, err := GetCredentialsFilePath()
		if err != nil {
			t.Fatalf("GetCredentialsFilePath() error = %v", err)
		}

		fileInfo, err := os.Stat(credentialsPath)
		if err != nil {
			t.Fatalf("Failed to stat credentials file: %v", err)
		}

		// Check that permissions are 0600 (read/write for user only)
		expectedPerm := os.FileMode(0600)
		if fileInfo.Mode().Perm() != expectedPerm {
			t.Errorf("Credentials file permissions = %v, want %v", fileInfo.Mode().Perm(), expectedPerm)
		}
	})

	// Test deleting credentials
	t.Run("delete credentials", func(t *testing.T) {
		// First save some credentials
		testCredentials := &Credentials{AccessToken: "test"}
		err := SaveCredentials(testCredentials)
		if err != nil {
			t.Fatalf("SaveCredentials() error = %v", err)
		}

		// Delete credentials
		err = DeleteCredentials()
		if err != nil {
			t.Fatalf("DeleteCredentials() error = %v", err)
		}

		// Try to load credentials (should fail)
		_, err = LoadCredentials()
		if err == nil {
			t.Error("LoadCredentials() should error after deletion")
		}
	})

	// Test deleting non-existent credentials
	t.Run("delete non-existent credentials", func(t *testing.T) {
		err := DeleteCredentials()
		if err != nil {
			t.Errorf("DeleteCredentials() should not error when file doesn't exist, got: %v", err)
		}
	})
}

func TestConfigOperations(t *testing.T) {
	// Set up test environment
	if err := os.Setenv("T42_ENV", "development"); err != nil {
		t.Fatalf("Failed to set T42_ENV: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("T42_ENV"); err != nil {
			t.Fatalf("Failed to unset T42_ENV: %v", err)
		}
	}()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "t42-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	// Test default config
	t.Run("default config", func(t *testing.T) {
		config := DefaultConfig()
		if config.DefaultFormat != "table" {
			t.Errorf("Default format should be 'table', got %v", config.DefaultFormat)
		}
		if config.Interactive != true {
			t.Errorf("Interactive should be true by default, got %v", config.Interactive)
		}
		if config.APIBaseURL != "https://api.intra.42.fr" {
			t.Errorf("API base URL should be 'https://api.intra.42.fr', got %v", config.APIBaseURL)
		}
	})

	// Test loading non-existent config (should return defaults)
	t.Run("load non-existent config", func(t *testing.T) {
		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig() should not error when file doesn't exist, got: %v", err)
		}

		// Should return default values
		defaultConfig := DefaultConfig()
		if config.DefaultFormat != defaultConfig.DefaultFormat {
			t.Errorf("Should return default format when file doesn't exist")
		}
	})

	// Test saving and loading config
	t.Run("save and load config", func(t *testing.T) {
		testConfig := &Config{
			DefaultFormat: "json",
			Interactive:   false,
			APIBaseURL:    "https://api.example.com",
		}

		// Save config
		err := SaveConfig(testConfig)
		if err != nil {
			t.Fatalf("SaveConfig() error = %v", err)
		}

		// Load config
		loadedConfig, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}

		// Compare configs
		if loadedConfig.DefaultFormat != testConfig.DefaultFormat {
			t.Errorf("DefaultFormat mismatch: got %v, want %v", loadedConfig.DefaultFormat, testConfig.DefaultFormat)
		}
		if loadedConfig.Interactive != testConfig.Interactive {
			t.Errorf("Interactive mismatch: got %v, want %v", loadedConfig.Interactive, testConfig.Interactive)
		}
		if loadedConfig.APIBaseURL != testConfig.APIBaseURL {
			t.Errorf("APIBaseURL mismatch: got %v, want %v", loadedConfig.APIBaseURL, testConfig.APIBaseURL)
		}
	})

	// Test config with missing fields (should fill defaults)
	t.Run("config with missing fields", func(t *testing.T) {
		// Create a config file with only some fields
		partialConfig := map[string]interface{}{
			"default_format": "json",
			// Missing interactive and api_base_url
		}

		configPath, err := GetConfigFilePath()
		if err != nil {
			t.Fatalf("GetConfigFilePath() error = %v", err)
		}

		// Ensure directory exists
		if err := EnsureConfigDir(); err != nil {
			t.Fatalf("EnsureConfigDir() error = %v", err)
		}

		// Write partial config
		data, err := yaml.Marshal(partialConfig)
		if err != nil {
			t.Fatalf("Failed to marshal partial config: %v", err)
		}

		if err := os.WriteFile(configPath, data, 0644); err != nil {
			t.Fatalf("Failed to write partial config: %v", err)
		}

		// Load config
		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}

		// Check that defaults are filled in
		if config.DefaultFormat != "json" {
			t.Errorf("DefaultFormat should be preserved: got %v", config.DefaultFormat)
		}
		if config.APIBaseURL != "https://api.intra.42.fr" {
			t.Errorf("APIBaseURL should be filled with default: got %v", config.APIBaseURL)
		}
	})
}

func TestHasValidCredentials(t *testing.T) {
	// Set up test environment
	if err := os.Setenv("T42_ENV", "development"); err != nil {
		t.Fatalf("Failed to set T42_ENV: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("T42_ENV"); err != nil {
			t.Fatalf("Failed to unset T42_ENV: %v", err)
		}
	}()

	// Test with no credentials
	t.Run("no credentials", func(t *testing.T) {
		// Ensure no credentials exist
		if err := DeleteCredentials(); err != nil {
			t.Logf("Warning: failed to delete credentials: %v", err)
		}

		if HasValidCredentials() {
			t.Error("HasValidCredentials() should return false when no credentials exist")
		}
	})

	// Test with valid credentials
	t.Run("valid credentials", func(t *testing.T) {
		testCredentials := &Credentials{
			AccessToken: "valid_token",
			TokenType:   "bearer",
			CreatedAt:   time.Now().Unix(),
			ExpiresIn:   7200, // 2 hours
		}

		err := SaveCredentials(testCredentials)
		if err != nil {
			t.Fatalf("SaveCredentials() error = %v", err)
		}

		if !HasValidCredentials() {
			t.Error("HasValidCredentials() should return true when valid credentials exist")
		}
	})

	// Test with empty access token
	t.Run("empty access token", func(t *testing.T) {
		testCredentials := &Credentials{
			AccessToken: "",
			TokenType:   "bearer",
		}

		err := SaveCredentials(testCredentials)
		if err != nil {
			t.Fatalf("SaveCredentials() error = %v", err)
		}

		if HasValidCredentials() {
			t.Error("HasValidCredentials() should return false when access token is empty")
		}
	})
}

func TestEnsureConfigDir(t *testing.T) {
	// Set up test environment
	if err := os.Setenv("T42_ENV", "development"); err != nil {
		t.Fatalf("Failed to set T42_ENV: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("T42_ENV"); err != nil {
			t.Fatalf("Failed to unset T42_ENV: %v", err)
		}
	}()

	// Remove the directory if it exists
	configDir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir() error = %v", err)
	}
	if err := os.RemoveAll(configDir); err != nil {
		t.Logf("Warning: failed to remove config dir: %v", err)
	}

	// Ensure the directory
	err = EnsureConfigDir()
	if err != nil {
		t.Fatalf("EnsureConfigDir() error = %v", err)
	}

	// Check that directory exists
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Config directory should exist after EnsureConfigDir()")
	}
}

func TestLoadDevelopmentSecrets(t *testing.T) {
	// Create a temporary .env file for testing
	tempDir, err := os.MkdirTemp("", "t42-secrets-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	// Test with missing .env file
	t.Run("missing env file", func(t *testing.T) {
		_, err := LoadDevelopmentSecrets()
		if err == nil {
			t.Error("LoadDevelopmentSecrets() should error when .env file doesn't exist")
		}
	})

	// Test with valid .env file
	t.Run("valid env file", func(t *testing.T) {
		// Clear environment variables from previous tests
		if err := os.Unsetenv("FT_UID"); err != nil {
			t.Logf("Warning: failed to unset FT_UID: %v", err)
		}
		if err := os.Unsetenv("FT_SECRET"); err != nil {
			t.Logf("Warning: failed to unset FT_SECRET: %v", err)
		}
		if err := os.Unsetenv("REDIRECT_URL"); err != nil {
			t.Logf("Warning: failed to unset REDIRECT_URL: %v", err)
		}

		// Create test .env file
		envContent := `FT_UID=test_client_id
FT_SECRET=test_client_secret
REDIRECT_URL=http://localhost:8080/callback
`
		envPath := GetDevelopmentEnvFilePath()

		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(envPath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		err := os.WriteFile(envPath, []byte(envContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test .env file: %v", err)
		}
		defer func() {
			if err := os.Remove(envPath); err != nil {
				t.Logf("Warning: failed to remove env file: %v", err)
			}
		}()

		secrets, err := LoadDevelopmentSecrets()
		if err != nil {
			t.Fatalf("LoadDevelopmentSecrets() error = %v", err)
		}

		if secrets.ClientID != "test_client_id" {
			t.Errorf("ClientID = %v, want %v", secrets.ClientID, "test_client_id")
		}
		if secrets.ClientSecret != "test_client_secret" {
			t.Errorf("ClientSecret = %v, want %v", secrets.ClientSecret, "test_client_secret")
		}
		if secrets.RedirectURL != "http://localhost:8080/callback" {
			t.Errorf("RedirectURL = %v, want %v", secrets.RedirectURL, "http://localhost:8080/callback")
		}
	})

	// Test with missing required fields
	t.Run("missing client id", func(t *testing.T) {
		// Clear environment variables from previous tests
		if err := os.Unsetenv("FT_UID"); err != nil {
			t.Logf("Warning: failed to unset FT_UID: %v", err)
		}
		if err := os.Unsetenv("FT_SECRET"); err != nil {
			t.Logf("Warning: failed to unset FT_SECRET: %v", err)
		}
		if err := os.Unsetenv("REDIRECT_URL"); err != nil {
			t.Logf("Warning: failed to unset REDIRECT_URL: %v", err)
		}

		envContent := `FT_SECRET=test_client_secret`
		envPath := GetDevelopmentEnvFilePath()

		if err := os.MkdirAll(filepath.Dir(envPath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		err := os.WriteFile(envPath, []byte(envContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test .env file: %v", err)
		}
		defer func() {
			if err := os.Remove(envPath); err != nil {
				t.Logf("Warning: failed to remove env file: %v", err)
			}
		}()

		_, err = LoadDevelopmentSecrets()
		if err == nil {
			t.Error("LoadDevelopmentSecrets() should error when FT_UID is missing")
		}
	})

	// Test with default redirect URL
	t.Run("default redirect url", func(t *testing.T) {
		// Clear environment variables from previous tests
		if err := os.Unsetenv("FT_UID"); err != nil {
			t.Logf("Warning: failed to unset FT_UID: %v", err)
		}
		if err := os.Unsetenv("FT_SECRET"); err != nil {
			t.Logf("Warning: failed to unset FT_SECRET: %v", err)
		}
		if err := os.Unsetenv("REDIRECT_URL"); err != nil {
			t.Logf("Warning: failed to unset REDIRECT_URL: %v", err)
		}

		envContent := `FT_UID=test_client_id
FT_SECRET=test_client_secret`
		envPath := GetDevelopmentEnvFilePath()

		if err := os.MkdirAll(filepath.Dir(envPath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		err := os.WriteFile(envPath, []byte(envContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test .env file: %v", err)
		}
		defer func() {
			if err := os.Remove(envPath); err != nil {
				t.Logf("Warning: failed to remove env file: %v", err)
			}
		}()

		secrets, err := LoadDevelopmentSecrets()
		if err != nil {
			t.Fatalf("LoadDevelopmentSecrets() error = %v", err)
		}

		expectedDefault := "http://127.0.0.1:8080/callback"
		if secrets.RedirectURL != expectedDefault {
			t.Errorf("RedirectURL should default to %v, got %v", expectedDefault, secrets.RedirectURL)
		}
	})
}
