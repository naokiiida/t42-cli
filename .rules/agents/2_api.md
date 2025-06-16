# Agent 2: API Client Core – Implementation Documentation

## How to Run the API Client Test

1. **Create a credentials file:**
   - Use your 42 API UID and SECRET to get an access token:
     ```sh
     curl -X POST --data "grant_type=client_credentials&client_id=$MY_AWESOME_UID&client_secret=$MY_AWESOME_SECRET" https://api.intra.42.fr/oauth/token
     ```
   - Save the JSON response as your credentials file:
     - **On macOS:** `$HOME/Library/Application Support/t42/credentials.json`
     - **On Linux:** `$HOME/.config/t42/credentials.json`
   - The exact path is determined by Go's `os.UserConfigDir()`. You can run the test `TestConfigFilePath` to see the path your system uses.

2. **Run the test from the project root:**
   ```sh
   go test ./internal
   ```
   - The test will pass if the token is valid and the API is reachable.

3. **To see the API response, run with verbose output:**
   ```sh
   go test ./internal -v
   ```
   - The test will always print the first 4096 bytes of the API response for inspection.

4. **Troubleshooting:**
   - If you see `No credentials found; skipping integration test` or `Credentials file does not exist at: ...`, make sure your credentials file is at the path shown in the test output (see `TestConfigFilePath`).

## Summary
Implements a minimal, robust API client for the 42 API, as required by the t42 CLI. Handles HTTP requests, error handling, config/token storage, pagination, retries, rate limits, and low-level passthrough for unsupported endpoints.

## Features
- Loads and saves OAuth2 access tokens securely (0600 permissions, user-only, cross-platform)
- Adds `Authorization: Bearer <token>` header to all requests
- Handles HTTP errors per 42 API conventions (see [42 API error codes](https://api.intra.42.fr/apidoc/guides/specification))
- Supports pagination (`page[number]`, `page[size]`), retries, and rate limiting (2 req/sec)
- Provides a low-level passthrough for arbitrary API calls (`t42 api ...`)

## Code Structure
- **internal/api.go**: Implements the API client, config loading/saving, and all HTTP logic
- **internal/api_test.go**: Usage/integration test for the API client

## Usage Example (Test)
```go
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
```

## References
- [42 API Getting Started](https://api.intra.42.fr/apidoc/guides/getting_started)
- [42 API Error Codes](https://api.intra.42.fr/apidoc/guides/specification)
- `.rules/42api.llms.md`

## Notes
- Extend as needed for richer error types, response parsing, or advanced rate limit handling.
- All API interactions for future features should use this client. 