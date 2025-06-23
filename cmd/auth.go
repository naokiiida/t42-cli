package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
	"context"
	"crypto/rand"
	"encoding/base64"

	"github.com/charmbracelet/huh"
	"github.com/joho/godotenv"
	"github.com/naokiiida/t42-cli/internal"
	"github.com/spf13/cobra"
)

// authCmd represents the base 'auth' command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication (login, status, logout)",
}

var (
	withSecret     bool
	noLocalhost    bool
	credsFile      string
	redirectPort   int
	envFile        string
)

// loginCmd for 't42 auth login'
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with 42 API (browser-based OAuth2 or fallback)",
	Run: func(cmd *cobra.Command, args []string) {
		if withSecret {
			// Fallback: prompt for UID/SECRET and use Client Credentials Flow
			var uid, secret string
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Title("42 API UID (client_id)").Value(&uid),
					huh.NewInput().Title("42 API SECRET (client_secret)").Value(&secret),
				),
			)
			if err := form.Run(); err != nil {
				cmd.PrintErrln("Login cancelled or error:", err)
				return
			}
			// Request token
			data := []byte("grant_type=client_credentials&client_id=" + uid + "&client_secret=" + secret)
			resp, err := http.Post("https://api.intra.42.fr/oauth/token", "application/x-www-form-urlencoded", bytes.NewReader(data))
			if err != nil {
				cmd.PrintErrln("Failed to request token:", err)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				b, _ := io.ReadAll(resp.Body)
				cmd.PrintErrf("Token request failed (%d): %s\n", resp.StatusCode, string(b))
				return
			}
			var tokenResp struct {
				AccessToken  string `json:"access_token"`
				RefreshToken string `json:"refresh_token"`
				ExpiresIn    int    `json:"expires_in"`
				CreatedAt    int64  `json:"created_at"`
				TokenType    string `json:"token_type"`
				Scope        string `json:"scope"` // Scope is usually part of the response
			}
			if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
				cmd.PrintErrln("Failed to parse token response:", err)
				return
			}
			cfg := &internal.Config{
				AccessToken:  tokenResp.AccessToken,
				RefreshToken: tokenResp.RefreshToken,
				ExpiresIn:    tokenResp.ExpiresIn,
				CreatedAt:    tokenResp.CreatedAt,
				TokenType:    tokenResp.TokenType,
			}
			if err := internal.SaveConfig(cfg); err != nil {
				cmd.PrintErrln("Failed to save credentials:", err)
				return
			}
			cmd.Println("Login successful! Access token saved.")
			return
		}

		// Browser-based OAuth2 Web Application Flow
		err := godotenv.Load(envFile)
		if err != nil {
			// If the default file is not found, it's not a fatal error yet,
			// as the user might rely on already set env vars or provide a different file.
			// We will check for CLIENT_ID and CLIENT_SECRET availability later.
			if envFile != "secret/.env" || !os.IsNotExist(err) {
				cmd.PrintErrln("Error loading .env file:", err)
				// Potentially exit here if a custom .env file was specified but not found
				// For now, we'll let it proceed to check the actual env vars
			}
		}

		clientID := os.Getenv("CLIENT_ID")
		clientSecret := os.Getenv("CLIENT_SECRET")

		if clientID == "" || clientSecret == "" {
			cmd.PrintErrln("Error: CLIENT_ID or CLIENT_SECRET not found in environment.")
			cmd.PrintErrf("Please ensure they are set in the environment or in the .env file (default: %s, customize with --env <file>).\n", envFile)
			cmd.PrintErrln("Example .env file content:")
			cmd.PrintErrln("CLIENT_ID=\"your_client_id\"")
			cmd.PrintErrln("CLIENT_SECRET=\"your_client_secret\"")
			return
		}

		state := randomState()
		port := redirectPort
		if port == 0 {
			port = pickRandomPort()
		}
		redirectURI := fmt.Sprintf("http://localhost:%d/callback", port)
		scope := "public"
		authURL := fmt.Sprintf("https://api.intra.42.fr/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s",
			clientID, urlQueryEscape(redirectURI), scope, state)
		cmd.Println("If your browser does not open, visit this URL:")
		cmd.Println(authURL)
		openBrowser(authURL)
		var code string
		if noLocalhost {
			cmd.Println("Paste the code from the redirected URL:")
			fmt.Print("Code: ")
			fmt.Scanln(&code)
		} else {
			code, err = waitForCode(port, state, cmd)
			if err != nil {
				cmd.PrintErrln("Failed to receive code:", err)
				return
			}
		}
		// Exchange code for token
		data := fmt.Sprintf("grant_type=authorization_code&client_id=%s&client_secret=%s&code=%s&redirect_uri=%s",
			clientID, clientSecret, code, redirectURI)
		tokenResp, err := exchangeCodeForToken(data)
		if err != nil {
			cmd.PrintErrln("Failed to exchange code for token:", err)
			return
		}
		cfg := &internal.Config{
			AccessToken:  tokenResp.AccessToken,
			RefreshToken: tokenResp.RefreshToken,
			ExpiresIn:    tokenResp.ExpiresIn,
			CreatedAt:    tokenResp.CreatedAt,
			TokenType:    tokenResp.TokenType,
		}
		if err := internal.SaveConfig(cfg); err != nil {
			cmd.PrintErrln("Failed to save credentials:", err)
			return
		}
		cmd.Println("Login successful! Access token saved.")
	},
}

// statusCmd for 't42 auth status'
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status (token info, expiry, scopes, app roles)",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := internal.LoadConfig()
		if err != nil || cfg.AccessToken == "" {
			cmd.PrintErrln("Not logged in. Run 't42 auth login' first.")
			return
		}
		// Call /oauth/token/info
		req, _ := http.NewRequest("GET", "https://api.intra.42.fr/oauth/token/info", nil)
		req.Header.Set("Authorization", "Bearer "+cfg.AccessToken)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			cmd.PrintErrln("Failed to request token info:", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			b, _ := io.ReadAll(resp.Body)
			cmd.PrintErrf("Token info request failed (%d): %s\n", resp.StatusCode, string(b))
			return
		}
		var info map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
			cmd.PrintErrln("Failed to parse token info:", err)
			return
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		enc.Encode(info)
	},
}

// logoutCmd for 't42 auth logout'
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear stored credentials",
	Run: func(cmd *cobra.Command, args []string) {
		path, err := internalConfigFilePathForLogout()
		if err != nil {
			cmd.PrintErrln("Failed to determine credentials file path:", err)
			return
		}
		err = os.Remove(path)
		if err != nil {
			if os.IsNotExist(err) {
				cmd.Println("Already logged out (no credentials file found).")
			} else {
				cmd.PrintErrln("Failed to delete credentials file:", err)
			}
			return
		}
		cmd.Println("Logged out. Credentials file deleted.")
	},
}

// internalConfigFilePathForLogout is a helper to get the config file path for logout
func internalConfigFilePathForLogout() (string, error) {
	return internalConfigFilePath()
}

// internalConfigFilePath is a helper to get the config file path from internal package
func internalConfigFilePath() (string, error) {
	return internal.ConfigFilePath()
}

// Helper functions for browser-based OAuth2
func loadClientCreds(path string) (string, string, error) {
	// If a custom creds file is provided, load from there
	if path != "" {
		f, err := os.Open(path)
		if err != nil {
			return "", "", err
		}
		defer f.Close()
		var creds struct {
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
		}
		if err := json.NewDecoder(f).Decode(&creds); err != nil {
			return "", "", err
		}
		return creds.ClientID, creds.ClientSecret, nil
	}
	// Otherwise, use the default config loader (internal.LoadConfig)
	cfg, err := internal.LoadConfig()
	if err != nil {
		return "", "", err
	}
	// Assume config has ClientID and ClientSecret fields (extend Config struct if needed)
	type clientCreds interface {
		GetClientID() string
		GetClientSecret() string
	}
	if c, ok := any(cfg).(clientCreds); ok {
		return c.GetClientID(), c.GetClientSecret(), nil
	}
	return "", "", fmt.Errorf("client_id and client_secret not found in config")
}
func randomState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
func pickRandomPort() int {
	l, _ := net.Listen("tcp", ":0")
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}
func urlQueryEscape(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, ":", "%3A"), "/", "%2F")
}
func openBrowser(url string) {
	switch runtime.GOOS {
	case "darwin":
		exec.Command("open", url).Start()
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		exec.Command("xdg-open", url).Start()
	}
}
func waitForCode(port int, expectedState string, cmd *cobra.Command) (string, error) {
	codeCh := make(chan string)
	server := &http.Server{Addr: fmt.Sprintf(":%d", port)}
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != expectedState {
			w.WriteHeader(400)
			w.Write([]byte("Invalid state"))
			return
		}
		code := r.URL.Query().Get("code")
		w.Write([]byte("You may now close this window and return to the CLI."))
		go func() { codeCh <- code }()
	})
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("Server error:", err)
		}
	}()
	select {
	case code := <-codeCh:
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		server.Shutdown(ctx)
		return code, nil
	case <-time.After(120 * time.Second):
		server.Shutdown(context.Background())
		return "", fmt.Errorf("timeout waiting for code")
	}
}

// FullTokenResponse defines the structure for all fields from the token endpoint.
type FullTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	CreatedAt    int64  `json:"created_at"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
}

func exchangeCodeForToken(data string) (FullTokenResponse, error) {
	var tokenResp FullTokenResponse
	resp, err := http.Post("https://api.intra.42.fr/oauth/token", "application/x-www-form-urlencoded", strings.NewReader(data))
	if err != nil {
		return tokenResp, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return tokenResp, fmt.Errorf("Token request failed (%d): %s", resp.StatusCode, string(b))
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return tokenResp, err
	}
	return tokenResp, nil
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(statusCmd)
	authCmd.AddCommand(logoutCmd)

	loginCmd.Flags().BoolVar(&withSecret, "with-secret", false, "Use OAuth2 Client Credentials Flow (manual UID/SECRET input)")
	loginCmd.Flags().BoolVar(&noLocalhost, "no-localhost", false, "Do not run a local server, manually enter code instead")
	loginCmd.Flags().StringVar(&credsFile, "creds", "", "Relative path to OAuth client secret file (default: OS config dir) [DEPRECATED for web flow, use --env]")
	loginCmd.Flags().IntVar(&redirectPort, "redirect-port", 0, "Specify a custom port for the redirect URL")
	loginCmd.Flags().StringVar(&envFile, "env", "secret/.env", "Path to .env file for CLIENT_ID and CLIENT_SECRET")
}