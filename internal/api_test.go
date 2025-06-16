package internal

import (
	"encoding/json"
	"os"
	"testing"
)

func TestAPIClient_PublicEndpoint(t *testing.T) {
	cfg, err := LoadConfig()
	if err != nil {
		t.Skip("No credentials found; skipping integration test")
	}
	client := NewAPIClient(cfg)
	resp, err := client.DoRequestWithRetry("GET", "/v2/cursus", nil, 3)
	if err != nil {
		t.Fatalf("API request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
	// Always print response for inspection
	b := make([]byte, 4096)
	n, _ := resp.Body.Read(b)
	t.Logf("Response: %s", string(b[:n]))
}

func TestConfigFilePath(t *testing.T) {
	path, err := configFilePath()
	if err != nil {
		t.Fatalf("Failed to get config file path: %v", err)
	}
	t.Logf("Config file path: %s", path)
	if _, err := os.Stat(path); err != nil {
		t.Errorf("Credentials file does not exist at: %s", path)
	} else {
		t.Logf("Credentials file exists at: %s", path)
	}
}

func TestAuthLoginWithSecretFallback(t *testing.T) {
	// This is a placeholder: in real tests, mock the HTTP call or use test credentials
	// For now, just check that the fallback code path can be called (manual test)
	t.Log("Manual test: run 't42 auth login --with-secret' and check that credentials are saved and status works.")
} 