package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"github.com/charmbracelet/huh"
	"github.com/naokiiida/t42-cli/internal"
)

// authCmd represents the base 'auth' command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication (login, status, logout)",
}

// loginCmd for 't42 auth login'
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with 42 API using OAuth2 Client Credentials Flow",
	Run: func(cmd *cobra.Command, args []string) {
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
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
			ExpiresIn   int    `json:"expires_in"`
			Scope       string `json:"scope"`
			CreatedAt   int64  `json:"created_at"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
			cmd.PrintErrln("Failed to parse token response:", err)
			return
		}
		cfg := &internal.Config{AccessToken: tokenResp.AccessToken}
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

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(statusCmd)
	authCmd.AddCommand(logoutCmd)
} 