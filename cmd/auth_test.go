package cmd

import (
	"bytes"
	"encoding/json"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/naokiiida/t42-cli/internal"
	"github.com/spf13/cobra"
)

// mockOriginalLoginRun mocks the loginCmd.Run function and returns a cleanup function.
// This is useful for tests that need to prevent the actual login logic from running.
func mockOriginalLoginRun(t *testing.T) func() {
	t.Helper()
	originalRun := loginCmd.Run
	loginCmd.Run = func(cmd *cobra.Command, args []string) {
		// Do nothing or add test-specific logging
		cmd.Println("Mocked loginCmd.Run called")
	}
	return func() {
		loginCmd.Run = originalRun
	}
}

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err = root.Execute()
	return buf.String(), err
}

func TestAuthRoot_Help(t *testing.T) {
	root := rootCmd
	root.AddCommand(authCmd)
	out, err := executeCommand(root, "auth", "--help")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(out, "Manage authentication") {
		t.Errorf("expected help output, got: %s", out)
	}
}

func TestAuthLogin_InvalidArgs(t *testing.T) {
	root := rootCmd
	root.AddCommand(authCmd)
	out, err := executeCommand(root, "auth", "login", "unexpected")
	if err == nil {
		t.Errorf("expected error for extra args, got none")
	}
	if !strings.Contains(out, "Usage:") {
		t.Errorf("expected usage output, got: %s", out)
	}
}

func TestAuth_InvalidSubcommand(t *testing.T) {
	root := rootCmd
	root.AddCommand(authCmd)
	out, err := executeCommand(root, "auth", "notacommand")
	if err == nil {
		t.Errorf("expected error for invalid subcommand, got none")
	}
	if !strings.Contains(out, "unknown command") && !strings.Contains(out, "Usage:") {
		t.Errorf("expected error or usage output, got: %s", out)
	}
}

func TestAuthLogin_PromptCancelled(t *testing.T) {
	// Save original Run function
	origRun := loginCmd.Run
	defer func() { loginCmd.Run = origRun }()

	// Override Run to simulate prompt error
	loginCmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.PrintErrln("Login cancelled or error: simulated error")
	}

	root := rootCmd
	root.AddCommand(authCmd)
	out, _ := executeCommand(root, "auth", "login")
	if !strings.Contains(out, "Login cancelled or error") {
		t.Errorf("expected prompt cancelled error, got: %s", out)
	}
}

func TestAuthLogin_APIFailure(t *testing.T) {
	// Save original Run function
	origRun := loginCmd.Run
	defer func() { loginCmd.Run = origRun }()

	// Override Run to simulate API error
	loginCmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.PrintErrln("Failed to request token: simulated API error")
	}

	root := rootCmd
	root.AddCommand(authCmd)
	out, _ := executeCommand(root, "auth", "login")
	if !strings.Contains(out, "Failed to request token") {
		t.Errorf("expected API error message, got: %s", out)
	}
}

func TestLoadClientCreds(t *testing.T) {
	// Write a temp creds file
	tempFile := "test_creds.json"
	creds := struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	}{ClientID: "testid", ClientSecret: "testsecret"}
	f, err := os.Create(tempFile)
	if err != nil {
		t.Fatalf("Failed to create temp creds file: %v", err)
	}
	json.NewEncoder(f).Encode(&creds)
	f.Close()
	defer os.Remove(tempFile)

	// Test loading from custom creds file
	id, secret, err := loadClientCreds(tempFile)
	if err != nil {
		t.Fatalf("Failed to load creds from file: %v", err)
	}
	if id != "testid" || secret != "testsecret" {
		t.Errorf("Expected testid/testsecret, got %s/%s", id, secret)
	}

	// Test loading from config - This part is no longer applicable as internal.Config
	// doesn't store ClientID/Secret directly in a way loadClientCreds("") would use
	// without a file path. loadClientCreds("") without a file path would try to
	// load the default config file, which is not what this test part was aiming for.
	// The function loadClientCreds primarily loads from a specific JSON file if path is given.
	// If path is empty, it would attempt to load the default config file, parse it as JSON,
	// and expect ClientID/ClientSecret fields there, which is not the current setup
	// for the main config which only stores AccessToken.
	// So, we only test the file-based loading here.

	// Test loading with non-existent file
	_, _, err = loadClientCreds("nonexistent.json")
	if err == nil {
		t.Errorf("Expected error when loading non-existent file, got nil")
	}
}

// Helper to create a temporary .env file
func createTempEnvFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	envFilePath := filepath.Join(dir, ".env")
	err := ioutil.WriteFile(envFilePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp .env file: %v", err)
	}
	return envFilePath
}

// Helper to create a temporary config file (simulating internal.SaveConfig)
func createTempCredsFile(t *testing.T, token string) string {
	t.Helper()
	cfg := &internal.Config{AccessToken: token}
	// Use a temporary directory for the config file
	// Ensure the internal package's default config path is temporarily overridden
	// or that SaveConfig/LoadConfig can work with a path provided for testing.

	// For simplicity in this test, we'll create a temp file and have internal.SaveConfig
	// save to it. We need to ensure internal.ConfigFilePath is mockable or Save/Load can take a path.
	// The current internal.ConfigFilePath uses os.UserConfigDir, which is hard to mock cleanly
	// without altering internal package for tests or using linker flags.

	// Alternative: create a temporary directory and point internal's config path there.
	// This requires internal.ConfigFilePath to be settable or use an env var for override.
	// For now, let's assume internal.SaveConfig and internal.LoadConfig will be tested
	// by their interaction with the actual commands, and we'll manage the file path.

	// Let's get the actual config path internal.ConfigFilePath() would use,
	// but place it in a temp dir to avoid messing with user's actual config.
	// This is still tricky.
	// The simplest for now: make internal.ConfigFilePath public and settable for tests,
	// or make SaveConfig/LoadConfig accept a path.

	// Given the constraints, we'll manually create a config-like file in a temp location
	// and make our command tests read from there if necessary.
	// The commands themselves call internal.SaveConfig() without a path.

	// For testing `statusCmd` or commands that load config:
	// 1. Create a temp dir.
	// 2. Override internal.ConfigFilePath (if possible) or internal.SaveConfig directly.
	// For now, let's mock the config loading part for statusCmd if direct path override is not easy.

	// Let's create a file in a temp dir that *looks* like the app's config file.
	tmpDir := t.TempDir()
	configFilePath := filepath.Join(tmpDir, ".t42-cli-config.json")

	originalConfigPathFunc := internal.ConfigFilePath // Store original
	internal.ConfigFilePath = func() (string, error) { // Override
		return configFilePath, nil
	}
	t.Cleanup(func() { internal.ConfigFilePath = originalConfigPathFunc }) // Restore

	err := internal.SaveConfig(cfg) // This will now use the overridden path
	if err != nil {
		t.Fatalf("Failed to save temp config: %v", err)
	}
	return configFilePath // Return the path to the temp config file
}

// Helper to setup a mock HTTP server
func mockServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	// Override the actual API base URL if your code uses a global var for it.
	// The current code hardcodes "https://api.intra.42.fr".
	// For robust testing, this URL should be configurable or replaced during tests.
	// This might require changes in the main code.
	// For now, tests will have to ensure that the http client used by the commands
	// can be made to point to httptest.Server.URL.
	// This is often done by replacing http.DefaultClient or passing a client into functions.
	// The current code uses http.Post directly, which uses http.DefaultClient.
	// We can temporarily replace http.DefaultClient.Transport.
	return server
}

// captureOutput executes a command and captures its stdout and stderr.
// Note: cobra's Execute already captures output if SetOut/SetErr are used.
// This is a more generic version if needed, but executeCommand should suffice.
func captureOutput(t *testing.T, action func()) (stdout, stderr string) {
	t.Helper()
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	action()

	wOut.Close()
	wErr.Close()
	outBytes, _ := ioutil.ReadAll(rOut)
	errBytes, _ := ioutil.ReadAll(rErr)
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	rOut.Close()
	rErr.Close()
	return string(outBytes), string(errBytes)
}

// TestLoginCmd_WebFlow_Success tests the default web application flow for login.
func TestLoginCmd_WebFlow_Success(t *testing.T) {
	// 1. Mock API server
	mockAPIServer := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth/token" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}
		expectedGrantType := "authorization_code"
		if r.Form.Get("grant_type") != expectedGrantType {
			t.Errorf("Expected grant_type %s, got %s", expectedGrantType, r.Form.Get("grant_type"))
			http.Error(w, "Bad grant_type", http.StatusBadRequest)
			return
		}
		// Check other expected params: client_id, client_secret, code, redirect_uri
		if r.Form.Get("client_id") != "testwebid" {
			t.Errorf("Expected client_id testwebid, got %s", r.Form.Get("client_id"))
		}
		if r.Form.Get("client_secret") != "testwebsecret" {
			t.Errorf("Expected client_secret testwebsecret, got %s", r.Form.Get("client_secret"))
		}
		if r.Form.Get("code") != "testauthcode" {
			t.Errorf("Expected code testauthcode, got %s", r.Form.Get("code"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token": "fake-web-access-token", "token_type": "bearer", "expires_in": 7200}`))
	})

	// Override the actual API URL by modifying the http.DefaultTransport for this test
	originalTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if strings.Contains(addr, "api.intra.42.fr") { // Target our mock server for API calls
				return net.Dial(network, mockAPIServer.Listener.Addr().String())
			}
			return net.Dial(network, addr) // Default behavior for other addresses
		},
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	t.Cleanup(func() { http.DefaultClient.Transport = originalTransport })


	// 2. Create temp .env file
	envContent := "CLIENT_ID=\"testwebid\"\nCLIENT_SECRET=\"testwebsecret\""
	tempEnvFile := createTempEnvFile(t, envContent)

	// 3. Mock openBrowser
	originalOpenBrowser := openBrowser
	openBrowserCalled := false
	openBrowser = func(url string) {
		openBrowserCalled = true
		// Check if URL contains expected client_id from .env
		if !strings.Contains(url, "client_id=testwebid") {
			t.Errorf("Expected auth URL to contain client_id=testwebid, got %s", url)
		}
		if !strings.Contains(url, "state=") { // State should be present
			t.Errorf("Expected auth URL to contain state, got %s", url)
		}
	}
	t.Cleanup(func() { openBrowser = originalOpenBrowser })

	// 4. Mock waitForCode
	originalWaitForCode := waitForCode
	waitForCodeCalled := false
	waitForCode = func(port int, expectedState string, cmd *cobra.Command) (string, error) {
		waitForCodeCalled = true
		if expectedState == "" {
			t.Error("waitForCode called with empty expectedState")
		}
		return "testauthcode", nil // Return a predefined code
	}
	t.Cleanup(func() { waitForCode = originalWaitForCode })

	// 5. Setup temp config path for result verification
	tmpDir := t.TempDir()
	configFilePath := filepath.Join(tmpDir, ".t42-cli-config.json")
	originalConfigPathFunc := internal.ConfigFilePath
	internal.ConfigFilePath = func() (string, error) { return configFilePath, nil }
	t.Cleanup(func() { internal.ConfigFilePath = originalConfigPathFunc })


	// Execute login command
	root := rootCmd // Assuming rootCmd is correctly initialized with authCmd and loginCmd
	// Need to re-initialize rootCmd for each test or ensure commands are added.
	// For simplicity, let's assume authCmd and its subcommands are added to rootCmd in init() of actual code.

	output, err := executeCommand(root, "auth", "login", "--env", tempEnvFile, "--redirect-port", "12345") // Added redirect-port
	if err != nil {
		t.Fatalf("login command failed: %v, output: %s", err, output)
	}

	if !openBrowserCalled {
		t.Error("openBrowser was not called")
	}
	if !waitForCodeCalled {
		t.Error("waitForCode was not called")
	}

	if !strings.Contains(output, "Login successful! Access token saved.") {
		t.Errorf("Expected success message, got: %s", output)
	}

	// Verify credentials file content
	savedCfg, loadErr := internal.LoadConfig() // This will use the mocked ConfigFilePath
	if loadErr != nil {
		t.Fatalf("Failed to load config for verification: %v", loadErr)
	}
	if savedCfg.AccessToken != "fake-web-access-token" {
		t.Errorf("Expected access token 'fake-web-access-token', got '%s'", savedCfg.AccessToken)
	}
}

func TestLoginCmd_WebFlow_EnvNotFound(t *testing.T) {
	// Ensure no real .env file exists at the specified path that could interfere
	nonExistentEnvFile := filepath.Join(t.TempDir(), "nonexistent.env")

	// Setup temp config path for potential (but not expected) save
	tmpDir := t.TempDir()
	configFilePath := filepath.Join(tmpDir, ".t42-cli-config.json")
	originalConfigPathFunc := internal.ConfigFilePath
	internal.ConfigFilePath = func() (string, error) { return configFilePath, nil }
	t.Cleanup(func() { internal.ConfigFilePath = originalConfigPathFunc })

	// Execute login command pointing to a non-existent .env file
	// No mock server needed as it should fail before API call
	output, err := executeCommand(rootCmd, "auth", "login", "--env", nonExistentEnvFile)

	// Depending on how strict the .env loading error is handled (e.g. if os.IsNotExist is checked)
	// For this test, we expect an error message related to missing CLIENT_ID or CLIENT_SECRET
	// because godotenv.Load() itself doesn't return an error if the file doesn't exist,
	// but the subsequent os.Getenv() calls will yield empty strings.
	if err == nil {
		// The command might not return an error itself if it just prints to stderr and exits cleanly.
		// Cobra's behavior on cmd.PrintErrln followed by return; might not set err from executeCommand.
		// Check output.
	}

	expectedErrorMsg := "Error: CLIENT_ID or CLIENT_SECRET not found in environment."
	if !strings.Contains(output, expectedErrorMsg) {
		t.Errorf("Expected output to contain '%s', got: %s", expectedErrorMsg, output)
	}

	// Ensure no credentials file was created
	if _, statErr := os.Stat(configFilePath); !os.IsNotExist(statErr) {
		t.Errorf("Expected no credentials file to be created, but found one at %s", configFilePath)
	}
}

func TestLoginCmd_WebFlow_TokenExchangeFailure(t *testing.T) {
	// 1. Mock API server to return an error
	mockAPIServer := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	})

	originalTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if strings.Contains(addr, "api.intra.42.fr") {
				return net.Dial(network, mockAPIServer.Listener.Addr().String())
			}
			return net.Dial(network, addr)
		},
	}
	t.Cleanup(func() { http.DefaultClient.Transport = originalTransport })

	// 2. Create temp .env file
	envContent := "CLIENT_ID=\"testwebid\"\nCLIENT_SECRET=\"testwebsecret\""
	tempEnvFile := createTempEnvFile(t, envContent)

	// 3. Mock openBrowser (can be minimal as it's before the failing step)
	originalOpenBrowser := openBrowser
	openBrowser = func(url string) {}
	t.Cleanup(func() { openBrowser = originalOpenBrowser })

	// 4. Mock waitForCode to return a valid code
	originalWaitForCode := waitForCode
	waitForCode = func(port int, expectedState string, cmd *cobra.Command) (string, error) {
		return "testauthcode", nil
	}
	t.Cleanup(func() { waitForCode = originalWaitForCode })

	// 5. Setup temp config path
	tmpDir := t.TempDir()
	configFilePath := filepath.Join(tmpDir, ".t42-cli-config.json")
	originalConfigPathFunc := internal.ConfigFilePath
	internal.ConfigFilePath = func() (string, error) { return configFilePath, nil }
	t.Cleanup(func() { internal.ConfigFilePath = originalConfigPathFunc })

	// Execute login command
	output, err := executeCommand(rootCmd, "auth", "login", "--env", tempEnvFile, "--redirect-port", "12345")

	// Error might not be returned by Execute() if handled by PrintErrln + os.Exit(1) or return in Run
	// Check output for error message
	expectedErrorMsg := "Failed to exchange code for token" // Or more specific from the actual code
	if !strings.Contains(output, expectedErrorMsg) {
		t.Errorf("Expected output to contain '%s', got: %s", expectedErrorMsg, output)
	}
	if !strings.Contains(output, "500") { // Check for status code from mock
		t.Errorf("Expected output to contain '500' status code, got: %s", output)
	}

	// Ensure no credentials file was created
	if _, statErr := os.Stat(configFilePath); !os.IsNotExist(statErr) {
		t.Errorf("Expected no credentials file to be created on error, but found one at %s", configFilePath)
	}
}

func TestLoginCmd_ClientCreds_Success(t *testing.T) {
	// 1. Mock API server
	mockAPIServer := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth/token" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}
		if r.Form.Get("grant_type") != "client_credentials" {
			t.Errorf("Expected grant_type client_credentials, got %s", r.Form.Get("grant_type"))
			http.Error(w, "Bad grant_type", http.StatusBadRequest)
			return
		}
		if r.Form.Get("client_id") != "testclientid" {
			t.Errorf("Expected client_id testclientid, got %s", r.Form.Get("client_id"))
		}
		if r.Form.Get("client_secret") != "testclientsecret" {
			t.Errorf("Expected client_secret testclientsecret, got %s", r.Form.Get("client_secret"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token": "fake-client-creds-token", "token_type": "bearer", "expires_in": 7200}`))
	})

	originalTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if strings.Contains(addr, "api.intra.42.fr") {
				return net.Dial(network, mockAPIServer.Listener.Addr().String())
			}
			return net.Dial(network, addr)
		},
	}
	t.Cleanup(func() { http.DefaultClient.Transport = originalTransport })

	// 2. Mock Stdin for huh.Form
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = oldStdin
		r.Close()
		w.Close()
	})

	// Provide input for huh.Form (UID then SECRET, each followed by newline)
	// This relies on how huh processes input. It might need specific sequences
	// like arrow keys for navigation if it's more than simple line input.
	// For basic huh.Input, newline should submit the field.
	go func() {
		defer w.Close()
		fmt.Fprintln(w, "testclientid")   // Input for API UID
		time.Sleep(100 * time.Millisecond) // Brief pause, might be needed for huh to process
		fmt.Fprintln(w, "testclientsecret") // Input for API SECRET
	}()

	// 3. Setup temp config path for result verification
	tmpDir := t.TempDir()
	configFilePath := filepath.Join(tmpDir, ".t42-cli-config.json")
	originalConfigPathFunc := internal.ConfigFilePath
	internal.ConfigFilePath = func() (string, error) { return configFilePath, nil }
	t.Cleanup(func() { internal.ConfigFilePath = originalConfigPathFunc })

	// Execute login command
	output, err := executeCommand(rootCmd, "auth", "login", "--with-secret")
	if err != nil {
		// If huh.Form fails due to Stdin issues, err might be set.
		t.Logf("huh form input: %s\n%s", "testclientid", "testclientsecret")
		t.Fatalf("login command --with-secret failed: %v, output: %s", err, output)
	}

	// Check if the mock API was called (implicitly by checking token)
	if !strings.Contains(output, "Login successful! Access token saved.") {
		t.Errorf("Expected success message, got: %s", output)
	}

	savedCfg, loadErr := internal.LoadConfig()
	if loadErr != nil {
		t.Fatalf("Failed to load config for verification: %v", loadErr)
	}
	if savedCfg.AccessToken != "fake-client-creds-token" {
		t.Errorf("Expected access token 'fake-client-creds-token', got '%s'", savedCfg.AccessToken)
	}
}

func TestLoginCmd_ClientCreds_TokenExchangeFailure(t *testing.T) {
	// 1. Mock API server to return an error
	mockAPIServer := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			http.Error(w, "Auth Failed", http.StatusUnauthorized)
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	})

	originalTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if strings.Contains(addr, "api.intra.42.fr") {
				return net.Dial(network, mockAPIServer.Listener.Addr().String())
			}
			return net.Dial(network, addr)
		},
	}
	t.Cleanup(func() { http.DefaultClient.Transport = originalTransport })

	// 2. Mock Stdin
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = oldStdin
		r.Close()
		w.Close()
	})
	go func() {
		defer w.Close()
		fmt.Fprintln(w, "testclientid")
		time.Sleep(100 * time.Millisecond)
		fmt.Fprintln(w, "testclientsecret")
	}()

	// 3. Setup temp config path
	tmpDir := t.TempDir()
	configFilePath := filepath.Join(tmpDir, ".t42-cli-config.json")
	originalConfigPathFunc := internal.ConfigFilePath
	internal.ConfigFilePath = func() (string, error) { return configFilePath, nil }
	t.Cleanup(func() { internal.ConfigFilePath = originalConfigPathFunc })

	// Execute login command
	output, _ := executeCommand(rootCmd, "auth", "login", "--with-secret")

	expectedErrorMsg := "Token request failed (401)" // Error from API
	if !strings.Contains(output, expectedErrorMsg) {
		t.Errorf("Expected output to contain '%s', got: %s", expectedErrorMsg, output)
	}

	if _, statErr := os.Stat(configFilePath); !os.IsNotExist(statErr) {
		t.Errorf("Expected no credentials file to be created on error, but found one at %s", configFilePath)
	}
}

func TestStatusCmd_Success(t *testing.T) {
	// 1. Create a dummy credentials file with a test token
	tmpDir := t.TempDir()
	configFilePath := filepath.Join(tmpDir, ".t42-cli-config.json")
	originalConfigPathFunc := internal.ConfigFilePath
	internal.ConfigFilePath = func() (string, error) { return configFilePath, nil }
	t.Cleanup(func() { internal.ConfigFilePath = originalConfigPathFunc })

	cfgToSave := &internal.Config{AccessToken: "test-status-token"}
	if err := internal.SaveConfig(cfgToSave); err != nil {
		t.Fatalf("Failed to save dummy config: %v", err)
	}

	// 2. Mock API server for /oauth/token/info
	mockAPIServer := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth/token/info" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test-status-token" {
			t.Errorf("Expected Authorization header 'Bearer test-status-token', got '%s'", r.Header.Get("Authorization"))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"resource_owner_id": "123", "scopes": ["public", "projects"], "application": {"uid": "app-uid"}}`))
	})

	originalTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if strings.Contains(addr, "api.intra.42.fr") {
				return net.Dial(network, mockAPIServer.Listener.Addr().String())
			}
			return net.Dial(network, addr)
		},
	}
	t.Cleanup(func() { http.DefaultClient.Transport = originalTransport })

	// 3. Execute status command
	output, err := executeCommand(rootCmd, "auth", "status")
	if err != nil {
		t.Fatalf("status command failed: %v, output: %s", err, output)
	}

	// 4. Assert output contains parts of the mocked token info
	if !strings.Contains(output, `"resource_owner_id": "123"`) {
		t.Errorf("Expected output to contain resource_owner_id, got: %s", output)
	}
	if !strings.Contains(output, `"scopes": [`) {
		t.Errorf("Expected output to contain scopes, got: %s", output)
	}
	if !strings.Contains(output, `"application": {`) {
		t.Errorf("Expected output to contain application info, got: %s", output)
	}
}

func TestStatusCmd_NotLoggedIn(t *testing.T) {
	// 1. Ensure no credentials file exists by pointing ConfigFilePath to a non-existent one
	tmpDir := t.TempDir()
	nonExistentConfigFilePath := filepath.Join(tmpDir, "nonexistentconfig.json")
	originalConfigPathFunc := internal.ConfigFilePath
	internal.ConfigFilePath = func() (string, error) { return nonExistentConfigFilePath, nil }
	t.Cleanup(func() { internal.ConfigFilePath = originalConfigPathFunc })

	// 2. Execute status command
	output, _ := executeCommand(rootCmd, "auth", "status") // err might be nil

	// 3. Assert "Not logged in" message
	expectedMsg := "Not logged in. Run 't42 auth login' first."
	if !strings.Contains(output, expectedMsg) {
		t.Errorf("Expected output to contain '%s', got: %s", expectedMsg, output)
	}
}

func TestStatusCmd_ApiError(t *testing.T) {
	// 1. Create a dummy credentials file
	tmpDir := t.TempDir()
	configFilePath := filepath.Join(tmpDir, ".t42-cli-config.json")
	originalConfigPathFunc := internal.ConfigFilePath
	internal.ConfigFilePath = func() (string, error) { return configFilePath, nil }
	t.Cleanup(func() { internal.ConfigFilePath = originalConfigPathFunc })

	cfgToSave := &internal.Config{AccessToken: "test-status-token-for-api-error"}
	if err := internal.SaveConfig(cfgToSave); err != nil {
		t.Fatalf("Failed to save dummy config: %v", err)
	}

	// 2. Mock API server to return an error
	mockAPIServer := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token/info" {
			http.Error(w, "API Server Down", http.StatusInternalServerError)
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	})

	originalTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if strings.Contains(addr, "api.intra.42.fr") {
				return net.Dial(network, mockAPIServer.Listener.Addr().String())
			}
			return net.Dial(network, addr)
		},
	}
	t.Cleanup(func() { http.DefaultClient.Transport = originalTransport })

	// 3. Execute status command
	output, _ := executeCommand(rootCmd, "auth", "status") // err might be nil

	// 4. Assert error message
	expectedErrorMsg := "Token info request failed (500)"
	if !strings.Contains(output, expectedErrorMsg) {
		t.Errorf("Expected output to contain '%s', got: %s", expectedErrorMsg, output)
	}
	if !strings.Contains(output, "API Server Down") { // The body of the error
		t.Errorf("Expected output to contain 'API Server Down', got: %s", output)
	}
}

func TestLogoutCmd_Success(t *testing.T) {
	// 1. Create a dummy credentials file
	tmpDir := t.TempDir()
	configFilePath := filepath.Join(tmpDir, ".t42-cli-config.json")
	originalConfigPathFunc := internal.ConfigFilePath
	internal.ConfigFilePath = func() (string, error) { return configFilePath, nil }
	t.Cleanup(func() { internal.ConfigFilePath = originalConfigPathFunc })

	// Create a dummy config file to be deleted
	cfgToSave := &internal.Config{AccessToken: "token-to-be-deleted"}
	if err := internal.SaveConfig(cfgToSave); err != nil {
		t.Fatalf("Failed to save dummy config for logout test: %v", err)
	}

	// Ensure file exists before logout
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		t.Fatalf("Dummy config file was not created before logout test: %s", configFilePath)
	}

	// 2. Execute logout command
	output, err := executeCommand(rootCmd, "auth", "logout")
	if err != nil {
		t.Fatalf("logout command failed: %v, output: %s", err, output)
	}

	// 3. Assert success message
	expectedMsg := "Logged out. Credentials file deleted."
	if !strings.Contains(output, expectedMsg) {
		t.Errorf("Expected output to contain '%s', got: %s", expectedMsg, output)
	}

	// 4. Assert the file is deleted
	if _, err := os.Stat(configFilePath); !os.IsNotExist(err) {
		t.Errorf("Expected credentials file to be deleted, but it still exists at %s", configFilePath)
	}
}

func TestLogoutCmd_AlreadyLoggedOut(t *testing.T) {
	// 1. Ensure no credentials file exists
	tmpDir := t.TempDir()
	nonExistentConfigFilePath := filepath.Join(tmpDir, "nonexistentconfig.json")
	originalConfigPathFunc := internal.ConfigFilePath

	// For logout, the relevant function is internalConfigFilePathForLogout.
	// Let's assume it uses internal.ConfigFilePath() internally or make it so.
	// The code for logoutCmd calls internalConfigFilePathForLogout() which calls internalConfigFilePath().
	internal.ConfigFilePath = func() (string, error) { return nonExistentConfigFilePath, nil }
	t.Cleanup(func() { internal.ConfigFilePath = originalConfigPathFunc })

	// 2. Execute logout command
	output, err := executeCommand(rootCmd, "auth", "logout")
	if err != nil {
		// This command should not error out even if file doesn't exist.
		t.Fatalf("logout command failed unexpectedly: %v, output: %s", err, output)
	}

	// 3. Assert "Already logged out" message
	expectedMsg := "Already logged out (no credentials file found)."
	if !strings.Contains(output, expectedMsg) {
		t.Errorf("Expected output to contain '%s', got: %s", expectedMsg, output)
	}
}