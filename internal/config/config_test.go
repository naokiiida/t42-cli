package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// TestGetConfigDir verifies that the config directory is created correctly.
func TestGetConfigDir(t *testing.T) {
	tempHome := t.TempDir()
	// Set environment variables that os.UserConfigDir() uses on different OSes.
	// This ensures our test is isolated from the actual user's home directory.
	t.Setenv("HOME", tempHome)             // For Linux/macOS
	t.Setenv("XDG_CONFIG_HOME", tempHome) // Overrides HOME on Linux
	t.Setenv("APPDATA", tempHome)         // For Windows

	dir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir() error = %v", err)
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("GetConfigDir() did not create directory: %s", dir)
	}

	// Check if the directory path ends with the application name.
	if filepath.Base(dir) != AppName {
		t.Errorf("GetConfigDir() path should end with %s, got %s", AppName, filepath.Base(dir))
	}
}

// TestPaths ensures that the generated paths for config and credentials are correct.
func TestPaths(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("XDG_CONFIG_HOME", tempHome)
	t.Setenv("APPDATA", tempHome)

	configDir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("pre-test setup failed to get config dir: %v", err)
	}

	credPath, err := GetCredentialsPath()
	if err != nil {
		t.Fatalf("GetCredentialsPath() error = %v", err)
	}
	expectedCredPath := filepath.Join(configDir, CredentialsFile)
	if credPath != expectedCredPath {
		t.Errorf("GetCredentialsPath() = %v, want %v", credPath, expectedCredPath)
	}

	confPath, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() error = %v", err)
	}
	expectedConfigPath := filepath.Join(configDir, ConfigFile)
	if confPath != expectedConfigPath {
		t.Errorf("GetConfigPath() = %v, want %v", confPath, expectedConfigPath)
	}
}

// TestCredentialsLifecycle tests the full Save/Load/Delete cycle for credentials.
func TestCredentialsLifecycle(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("XDG_CONFIG_HOME", tempHome)
	t.Setenv("APPDATA", tempHome)

	// 1. Test loading when no file exists
	t.Run("Load non-existent credentials", func(t *testing.T) {
		_, err := LoadCredentials()
		if err != ErrNotLoggedIn {
			t.Errorf("Expected ErrNotLoggedIn, got %v", err)
		}
	})

	// 2. Test saving credentials
	// Use Truncate to remove monotonic clock data, ensuring DeepEqual works reliably.
	creds := &Credentials{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(1 * time.Hour).UTC().Truncate(time.Second),
	}

	t.Run("Save and check permissions", func(t *testing.T) {
		err := SaveCredentials(creds)
		if err != nil {
			t.Fatalf("SaveCredentials() error = %v", err)
		}

		path, _ := GetCredentialsPath()
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("os.Stat() on credentials file error = %v", err)
		}

		// Check permissions - should be 0600
		if info.Mode().Perm() != 0600 {
			t.Errorf("Credentials file permissions are %v, want 0600", info.Mode().Perm())
		}
	})

	// 3. Test loading saved credentials
	t.Run("Load saved credentials", func(t *testing.T) {
		loadedCreds, err := LoadCredentials()
		if err != nil {
			t.Fatalf("LoadCredentials() error = %v", err)
		}
		if !reflect.DeepEqual(creds, loadedCreds) {
			t.Errorf("Loaded credentials do not match saved ones.\nGot: %+v\nWant:%+v", loadedCreds, creds)
		}
	})

	// 4. Test deleting credentials
	t.Run("Delete credentials", func(t *testing.T) {
		err := DeleteCredentials()
		if err != nil {
			t.Fatalf("DeleteCredentials() error = %v", err)
		}

		path, _ := GetCredentialsPath()
		_, err = os.Stat(path)
		if !os.IsNotExist(err) {
			t.Errorf("Credentials file should not exist after deletion, but it does.")
		}
	})
}

// TestPreferencesLifecycle tests the Save/Load cycle for user preferences.
func TestPreferencesLifecycle(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("XDG_CONFIG_HOME", tempHome)
	t.Setenv("APPDATA", tempHome)

	// 1. Test loading when no file exists (should return default)
	t.Run("Load non-existent preferences", func(t *testing.T) {
		prefs, err := LoadPreferences()
		if err != nil {
			t.Fatalf("LoadPreferences() error = %v", err)
		}
		if !reflect.DeepEqual(prefs, &Preferences{}) {
			t.Errorf("Expected empty preferences, got %+v", prefs)
		}
	})

	// 2. Test saving preferences
	prefs := &Preferences{
		// Add fields here for testing when they are defined in the struct
	}

	t.Run("Save and load preferences", func(t *testing.T) {
		err := SavePreferences(prefs)
		if err != nil {
			t.Fatalf("SavePreferences() error = %v", err)
		}

		path, _ := GetConfigPath()
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatalf("SavePreferences() did not create the file at %s", path)
		}

		loadedPrefs, err := LoadPreferences()
		if err != nil {
			t.Fatalf("LoadPreferences() after saving error = %v", err)
		}

		if !reflect.DeepEqual(prefs, loadedPrefs) {
			t.Errorf("Loaded preferences do not match saved ones.\nGot: %+v\nWant:%+v", loadedPrefs, prefs)
		}
	})
}

// TestLoadDotEnv verifies that .env file loading works for development.
func TestLoadDotEnv(t *testing.T) {
	// Create a temporary directory structure mimicking the project layout
	tempProjectDir := t.TempDir()
	secretDir := filepath.Join(tempProjectDir, "secret")
	if err := os.Mkdir(secretDir, 0755); err != nil {
		t.Fatalf("Failed to create temp secret dir: %v", err)
	}

	envFilePath := filepath.Join(secretDir, ".env")
	envContent := "TEST_KEY=TEST_VALUE\nANOTHER_KEY=123"
	if err := os.WriteFile(envFilePath, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to write temp .env file: %v", err)
	}

	// Change working directory to the temp project dir so LoadDotEnv can find "secret/.env"
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	if err := os.Chdir(tempProjectDir); err != nil {
		t.Fatalf("Failed to change working directory: %v", err)
	}
	defer os.Chdir(originalWD) // Ensure we change back to the original directory

	// Load the .env file
	LoadDotEnv()

	// Check if the environment variables are set
	if val := os.Getenv("TEST_KEY"); val != "TEST_VALUE" {
		t.Errorf("Expected TEST_KEY to be 'TEST_VALUE', got '%s'", val)
	}
	if val := os.Getenv("ANOTHER_KEY"); val != "123" {
		t.Errorf("Expected ANOTHER_KEY to be '123', got '%s'", val)
	}
}

// TestIsAccessTokenExpired checks the token expiry logic.
func TestIsAccessTokenExpired(t *testing.T) {
	t.Run("Token is not expired", func(t *testing.T) {
		creds := &Credentials{ExpiresAt: time.Now().Add(10 * time.Minute)}
		if creds.IsAccessTokenExpired() {
			t.Error("Token should not be considered expired")
		}
	})

	t.Run("Token is expired", func(t *testing.T) {
		creds := &Credentials{ExpiresAt: time.Now().Add(-10 * time.Minute)}
		if !creds.IsAccessTokenExpired() {
			t.Error("Token should be considered expired")
		}
	})

	t.Run("Token expires within the buffer", func(t *testing.T) {
		creds := &Credentials{ExpiresAt: time.Now().Add(30 * time.Second)}
		if !creds.IsAccessTokenExpired() {
			t.Error("Token expiring in 30 seconds should be considered expired due to buffer")
		}
	})
}

// TestCredentialsSerialization ensures that the JSON tags are correct and time is handled properly.
func TestCredentialsSerialization(t *testing.T) {
	// Use a fixed time to make the test deterministic
	fixedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	creds := &Credentials{
		AccessToken:  "access",
		RefreshToken: "refresh",
		TokenType:    "Bearer",
		ExpiresAt:    fixedTime,
	}

	data, err := json.Marshal(creds)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaledCreds Credentials
	err = json.Unmarshal(data, &unmarshaledCreds)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if !reflect.DeepEqual(creds, &unmarshaledCreds) {
		t.Errorf("Unmarshaled credentials do not match original.\nGot: %+v\nWant:%+v", unmarshaledCreds, creds)
	}
}