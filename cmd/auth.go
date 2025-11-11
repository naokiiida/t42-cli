package cmd

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/naokiiida/t42-cli/internal/api"
	"github.com/naokiiida/t42-cli/internal/config"
	"github.com/naokiiida/t42-cli/internal/oauth"
)

const (
	// OAuth2 endpoints for 42 API
	authorizeURL = "https://api.intra.42.fr/oauth/authorize"
	tokenURL     = "https://api.intra.42.fr/oauth/token"

	// Default redirect URL for local callback server
	defaultRedirectURL = "http://127.0.0.1:8080/callback"

	// OAuth2 scopes
	defaultScope = "public"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
	Long: `Manage authentication with the 42 API.

This command group allows you to log in to your 42 account,
check your authentication status, and log out.`,
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to your 42 account",
	Long: `Log in to your 42 account using OAuth2 Web Application Flow.

This will open your web browser to the 42 authentication page.
After you authorize the application, you will be redirected back
to the CLI and your credentials will be saved securely.`,
	RunE: runLogin,
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out of your 42 account",
	Long: `Log out of your 42 account by removing stored credentials.

This will delete your locally stored authentication token.
You will need to log in again to use authenticated features.`,
	RunE: runLogout,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	Long: `Check your current authentication status.

This will show information about your stored credentials,
including token scope, expiry time, and user information.`,
	RunE: runStatus,
}

// OAuth2 state for security
type oauthState struct {
	State     string `json:"state"`
	CreatedAt int64  `json:"created_at"`
}

func init() {
	// Add auth subcommands
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(logoutCmd)
	authCmd.AddCommand(statusCmd)

	// Add auth command to root
	rootCmd.AddCommand(authCmd)

	// Login command flags
	loginCmd.Flags().StringP("port", "p", "8080", "Port for local callback server")
	loginCmd.Flags().Bool("no-browser", false, "Don't automatically open browser")
}

// tryListen attempts to bind to the given address and port, returns net.Listener and error
func tryListen(addr string, port int) (net.Listener, error) {
	lnAddr := fmt.Sprintf("%s:%d", addr, port)
	return net.Listen("tcp", lnAddr)
}

// findFreePort tries to bind to a free port on the given address, returns net.Listener, port, error
func findFreePort(addr string) (net.Listener, int, error) {
	for p := 49152; p <= 65535; p++ { // Use ephemeral port range
		ln, err := tryListen(addr, p)
		if err == nil {
			return ln, p, nil
		}
	}
	return nil, 0, fmt.Errorf("no free port found on %s", addr)
}

func runLogin(cmd *cobra.Command, args []string) error {
	var ln net.Listener

	// --- Loopback binding logic ---
	requestedPortStr, _ := cmd.Flags().GetString("port")
	requestedPort, err := strconv.Atoi(requestedPortStr)
	if err != nil {
		return fmt.Errorf("invalid port: %w", err)
	}
	bindAddr := "127.0.0.1"
	port := requestedPort
	ln, err = tryListen(bindAddr, port)
	if err != nil {
		// Try to find a free port
		ln, port, err = findFreePort(bindAddr)
		if err != nil {
			// Fallback to IPv6
			bindAddr = "::1"
			ln, port, err = findFreePort(bindAddr)
			if err != nil {
				return fmt.Errorf("failed to bind to any loopback address: %w", err)
			}
		}
	}
	redirectURL := fmt.Sprintf("http://%s:%d/callback", bindAddr, port)
	// --- End loopback binding logic ---

	// Check if already logged in
	if config.HasValidCredentials() {
		if !GetJSONOutput() {
			fmt.Println("You are already logged in!")

			// Ask if user wants to re-authenticate
			var reauth bool
			err := huh.NewConfirm().
				Title("Do you want to log in again?").
				Description("This will replace your current credentials.").
				Value(&reauth).
				Run()

			if err != nil {
				return fmt.Errorf("failed to get user confirmation: %w", err)
			}

			if !reauth {
				return nil
			}
		}
	}

	// Get OAuth2 configuration
	secrets, err := getOAuth2Config()
	if err != nil {
		return fmt.Errorf("failed to get OAuth2 configuration: %w", err)
	}

	// Get port from flag
	portStr, _ := cmd.Flags().GetString("port")
	port, err = strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid port: %w", err)
	}

	// Generate state for security
	state, err := generateState()
	if err != nil {
		return fmt.Errorf("failed to generate state: %w", err)
	}

	// Generate PKCE parameters
	pkce, err := oauth.GeneratePKCEParams()
	if err != nil {
		return fmt.Errorf("failed to generate PKCE parameters: %w", err)
	}

	if GetVerbose() {
		fmt.Printf("[DEBUG] PKCE generated:\n")
		fmt.Printf("  Code Verifier: %s...\n", pkce.CodeVerifier[:min(len(pkce.CodeVerifier), 20)])
		fmt.Printf("  Code Challenge: %s...\n", pkce.CodeChallenge[:min(len(pkce.CodeChallenge), 20)])
	}

	// Build authorization URL with PKCE
	authURL := buildAuthorizationURL(secrets.ClientID, redirectURL, state, defaultScope, pkce.CodeChallenge)

	// Start local callback server
	tokenChan := make(chan *config.Credentials, 1)
	errorChan := make(chan error, 1)

	// Update callback handler to pass PKCE verifier
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		handleCallback(w, r, secrets, redirectURL, state, pkce.CodeVerifier, tokenChan, errorChan)
	})

	// Start server in goroutine
	go func() {
		serveErr := http.Serve(ln, nil)
		if serveErr != nil {
			errorChan <- fmt.Errorf("callback server error: %w", serveErr)
		}
	}()

	// Wait a bit for server to start
	time.Sleep(100 * time.Millisecond)

	if !GetJSONOutput() {
		fmt.Printf("🔐 Starting OAuth2 flow...\n")
		fmt.Printf("📱 Opening browser to: %s\n", authURL)
		fmt.Printf("🌐 Waiting for callback on http://127.0.0.1:%d\n", port)
		fmt.Printf("⏰ This will timeout in 5 minutes...\n\n")
	}

	// Open browser unless disabled
	noBrowser, _ := cmd.Flags().GetBool("no-browser")
	if !noBrowser {
		if err := openBrowser(authURL); err != nil && !GetJSONOutput() {
			fmt.Printf("⚠️  Failed to open browser automatically: %v\n", err)
			fmt.Printf("Please manually open: %s\n", authURL)
		}
	} else {
		if !GetJSONOutput() {
			fmt.Printf("Please open the following URL in your browser:\n%s\n", authURL)
		}
	}

	// Wait for callback or timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var credentials *config.Credentials

	select {
	case creds := <-tokenChan:
		credentials = creds
	case err := <-errorChan:
		return err
	case <-ctx.Done():
		return fmt.Errorf("authentication timeout - no response received within 5 minutes")
	}

	// Shutdown server
	ln.Close()

	// Save credentials
	if err := config.SaveCredentials(credentials); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	// Get user info to confirm authentication
	client := api.NewClient(credentials.AccessToken)
	user, err := client.GetMe(context.Background())
	if err != nil {
		if !GetJSONOutput() {
			fmt.Printf("⚠️  Warning: Authentication succeeded but failed to get user info: %v\n", err)
		}
	}

	if GetJSONOutput() {
		result := map[string]interface{}{
			"success":    true,
			"scope":      credentials.Scope,
			"expires_in": credentials.ExpiresIn,
		}
		if user != nil {
			result["user"] = map[string]interface{}{
				"id":    user.ID,
				"login": user.Login,
				"email": user.Email,
			}
		}
		output, _ := json.Marshal(result)
		fmt.Println(string(output))
	} else {
		fmt.Printf("✅ Successfully logged in!\n")
		if user != nil {
			fmt.Printf("👋 Welcome, %s (%s)!\n", user.Login, user.Email)
		}
		fmt.Printf("🔑 Token scope: %s\n", credentials.Scope)
		fmt.Printf("⏰ Token expires in: %d seconds\n", credentials.ExpiresIn)
	}

	return nil
}

func runLogout(cmd *cobra.Command, args []string) error {
	// Check if logged in
	if !config.HasValidCredentials() {
		if GetJSONOutput() {
			fmt.Println(`{"success":true,"message":"Already logged out"}`)
		} else {
			fmt.Println("You are not currently logged in.")
		}
		return nil
	}

	// Confirm logout unless JSON output
	if !GetJSONOutput() {
		var confirm bool
		err := huh.NewConfirm().
			Title("Are you sure you want to log out?").
			Description("This will remove your stored credentials.").
			Value(&confirm).
			Run()

		if err != nil {
			return fmt.Errorf("failed to get user confirmation: %w", err)
		}

		if !confirm {
			fmt.Println("Logout cancelled.")
			return nil
		}
	}

	// Delete credentials
	if err := config.DeleteCredentials(); err != nil {
		return fmt.Errorf("failed to delete credentials: %w", err)
	}

	if GetJSONOutput() {
		fmt.Println(`{"success":true,"message":"Logged out successfully"}`)
	} else {
		fmt.Println("✅ Successfully logged out!")
	}

	return nil
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Check if logged in
	if !config.HasValidCredentials() {
		if GetJSONOutput() {
			fmt.Println(`{"authenticated":false,"message":"Not logged in"}`)
		} else {
			fmt.Println("❌ Not logged in")
			fmt.Println("Run 't42 auth login' to authenticate.")
		}
		return nil
	}

	// Load credentials
	credentials, err := config.LoadCredentials()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	// Create API client with automatic token refresh
	client, err := NewAPIClient()
	if err != nil {
		// If we can't create the client, still show credential info
		client = nil
	}

	// Get user info
	var user *api.User
	if client != nil {
		user, err = client.GetMe(context.Background())
		// Reload credentials in case they were refreshed
		credentials, _ = config.LoadCredentials()
	}

	// Calculate token expiry
	expiresAt := config.GetTokenExpiryTime(credentials)
	timeUntilExpiry := time.Until(expiresAt)
	isExpired := timeUntilExpiry < 0

	if GetJSONOutput() {
		result := map[string]interface{}{
			"authenticated": true,
			"scope":         credentials.Scope,
			"created_at":    credentials.CreatedAt,
			"expires_in":    credentials.ExpiresIn,
			"expires_at":    expiresAt.Unix(),
			"expired":       isExpired,
		}

		if !isExpired {
			result["time_until_expiry"] = int64(timeUntilExpiry.Seconds())
		}

		if err == nil && user != nil {
			result["user"] = map[string]interface{}{
				"id":    user.ID,
				"login": user.Login,
				"email": user.Email,
			}
		} else {
			result["user_error"] = err.Error()
		}

		output, _ := json.Marshal(result)
		fmt.Println(string(output))
	} else {
		fmt.Println("✅ Authenticated")

		if err == nil && user != nil {
			fmt.Printf("👤 User: %s (%s)\n", user.Login, user.Email)
			fmt.Printf("🆔 User ID: %d\n", user.ID)
		} else {
			fmt.Printf("⚠️  User info unavailable: %v\n", err)
		}

		fmt.Printf("🔑 Token scope: %s\n", credentials.Scope)
		fmt.Printf("📅 Token created: %s\n", time.Unix(credentials.CreatedAt, 0).Format(time.RFC3339))

		if isExpired {
			fmt.Printf("⏰ Token status: ❌ EXPIRED (%s ago)\n", (-timeUntilExpiry).Truncate(time.Second))
		} else {
			fmt.Printf("⏰ Token expires: %s (in %s)\n",
				expiresAt.Format(time.RFC3339),
				timeUntilExpiry.Truncate(time.Second))
		}
	}

	return nil
}

func getOAuth2Config() (*config.DevelopmentSecrets, error) {
	// Try to load from development secrets first
	if secrets, err := config.LoadDevelopmentSecrets(); err == nil {
		return secrets, nil
	}

	// If no development secrets, check environment variables
	clientID := os.Getenv("FT_UID")
	clientSecret := os.Getenv("FT_SECRET")

	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("OAuth2 configuration not found. Please create secret/.env with FT_UID and FT_SECRET, or set environment variables")
	}

	return &config.DevelopmentSecrets{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  defaultRedirectURL,
	}, nil
}

func generateState() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func buildAuthorizationURL(clientID, redirectURL, state, scope string, pkceChallenge string) string {
	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURL)
	params.Set("response_type", "code")
	params.Set("scope", scope)
	params.Set("state", state)

	// Add PKCE parameters
	if pkceChallenge != "" {
		params.Set("code_challenge", pkceChallenge)
		params.Set("code_challenge_method", "S256")
	}

	return authorizeURL + "?" + params.Encode()
}

func handleCallback(w http.ResponseWriter, r *http.Request, secrets *config.DevelopmentSecrets, redirectURL, expectedState, pkceVerifier string, tokenChan chan<- *config.Credentials, errorChan chan<- error) {
	// Parse query parameters
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	if GetVerbose() {
		fmt.Printf("[DEBUG] Callback received:\n")
		fmt.Printf("  Code: %s...\n", code[:min(len(code), 20)])
		fmt.Printf("  State: %s...\n", state[:min(len(state), 20)])
		fmt.Printf("  Redirect URI: %s\n", redirectURL)
	}

	// Check for OAuth2 errors
	if errorParam != "" {
		errorDesc := r.URL.Query().Get("error_description")
		msg := fmt.Sprintf("OAuth2 error: %s", errorParam)
		if errorDesc != "" {
			msg += fmt.Sprintf(" (%s)", errorDesc)
		}

		http.Error(w, msg, http.StatusBadRequest)
		errorChan <- fmt.Errorf("%s", msg)
		return
	}

	// Validate state parameter
	if state != expectedState {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		errorChan <- fmt.Errorf("invalid state parameter - possible CSRF attack")
		return
	}

	// Validate authorization code
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		errorChan <- fmt.Errorf("missing authorization code")
		return
	}

	// Exchange code for token (with PKCE verifier)
	if GetVerbose() {
		fmt.Printf("[DEBUG] Exchanging authorization code for token with PKCE...\n")
	}
	credentials, err := exchangeCodeForToken(code, redirectURL, secrets, pkceVerifier)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to exchange code for token: %v", err)
		http.Error(w, errorMsg, http.StatusInternalServerError)
		errorChan <- err
		return
	}

	// Send success response
	successHTML := `
<!DOCTYPE html>
<html>
<head>
    <title>42 CLI - Authentication Successful</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; text-align: center; padding: 50px; background: #f5f5f5; }
        .container { background: white; border-radius: 10px; padding: 40px; max-width: 500px; margin: 0 auto; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .success { color: #28a745; font-size: 48px; margin-bottom: 20px; }
        h1 { color: #333; margin-bottom: 10px; }
        p { color: #666; line-height: 1.5; }
        .close-btn { background: #007bff; color: white; border: none; padding: 10px 20px; border-radius: 5px; cursor: pointer; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="success">✅</div>
        <h1>Authentication Successful!</h1>
        <p>You have successfully logged in to your 42 account.</p>
        <p>You can now close this window and return to your terminal.</p>
        <button class="close-btn" onclick="window.close()">Close Window</button>
    </div>
    <script>
        // Auto-close after 3 seconds
        setTimeout(() => window.close(), 3000);
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(successHTML))

	// Send credentials to main goroutine
	tokenChan <- credentials
}

func exchangeCodeForToken(code, redirectURL string, secrets *config.DevelopmentSecrets, pkceVerifier string) (*config.Credentials, error) {
	// Prepare token request
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", secrets.ClientID)
	data.Set("client_secret", secrets.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURL)

	// Add PKCE code verifier
	if pkceVerifier != "" {
		data.Set("code_verifier", pkceVerifier)
	}

	if GetVerbose() {
		fmt.Printf("[DEBUG] Token exchange request:\n")
		fmt.Printf("  URL: %s\n", tokenURL)
		fmt.Printf("  Grant Type: %s\n", data.Get("grant_type"))
		fmt.Printf("  Client ID: %s\n", secrets.ClientID)
		fmt.Printf("  Redirect URI: %s\n", redirectURL)
		fmt.Printf("  Code: %s...\n", code[:min(len(code), 20)])
		if pkceVerifier != "" {
			fmt.Printf("  PKCE: enabled (verifier: %s...)\n", pkceVerifier[:min(len(pkceVerifier), 20)])
		}
	}

	// Make token request
	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to make token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if GetVerbose() {
		fmt.Printf("[DEBUG] Token response:\n")
		fmt.Printf("  Status: %d\n", resp.StatusCode)
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("  Body: %s\n", string(body))
		}
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var errorResp api.ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return nil, fmt.Errorf("token request failed (status %d): %s - %s", resp.StatusCode, errorResp.Error, errorResp.ErrorDescription)
		}
		return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse token response
	var tokenResp api.Token
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Convert to our credentials format
	credentials := &config.Credentials{
		AccessToken:      tokenResp.AccessToken,
		TokenType:        tokenResp.TokenType,
		ExpiresIn:        tokenResp.ExpiresIn,
		RefreshToken:     tokenResp.RefreshToken,
		Scope:            tokenResp.Scope,
		CreatedAt:        tokenResp.CreatedAt,
		SecretValidUntil: tokenResp.SecretValidUntil,
	}

	return credentials, nil
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}

	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

// refreshAccessToken refreshes the access token using the refresh token
func refreshAccessToken(refreshToken string) (*config.Credentials, error) {
	secrets, err := getOAuth2Config()
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth2 configuration: %w", err)
	}

	// Prepare token refresh request
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", secrets.ClientID)
	data.Set("client_secret", secrets.ClientSecret)
	data.Set("refresh_token", refreshToken)

	// Make token request
	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to make token refresh request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var errorResp api.ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return nil, fmt.Errorf("token refresh failed: %s", errorResp.ErrorDescription)
		}
		return nil, fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse token response
	var tokenResp api.Token
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Convert to our credentials format
	credentials := &config.Credentials{
		AccessToken:      tokenResp.AccessToken,
		TokenType:        tokenResp.TokenType,
		ExpiresIn:        tokenResp.ExpiresIn,
		RefreshToken:     tokenResp.RefreshToken,
		Scope:            tokenResp.Scope,
		CreatedAt:        tokenResp.CreatedAt,
		SecretValidUntil: tokenResp.SecretValidUntil,
	}

	return credentials, nil
}

// RefreshTokenIfNeeded checks if the token is expired or about to expire and refreshes it
func RefreshTokenIfNeeded() error {
	credentials, err := config.LoadCredentials()
	if err != nil {
		return err // No credentials to refresh
	}

	// Check if token is expired or will expire in the next 5 minutes
	expiresAt := time.Unix(credentials.CreatedAt, 0).Add(time.Duration(credentials.ExpiresIn) * time.Second)
	timeUntilExpiry := time.Until(expiresAt)

	// If token is valid for more than 5 minutes, no need to refresh
	if timeUntilExpiry > 5*time.Minute {
		return nil
	}

	// Check if we have a refresh token
	if credentials.RefreshToken == "" {
		return fmt.Errorf("access token expired and no refresh token available - please log in again")
	}

	// Refresh the token
	newCredentials, err := refreshAccessToken(credentials.RefreshToken)
	if err != nil {
		return fmt.Errorf("failed to refresh access token: %w", err)
	}

	// Save the new credentials
	if err := config.SaveCredentials(newCredentials); err != nil {
		return fmt.Errorf("failed to save refreshed credentials: %w", err)
	}

	return nil
}

// Helper function to return minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
