package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/naokiiida/t42-cli/internal/api"
	"github.com/naokiiida/t42-cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	// Version information (will be set during build)
	version = "dev"
	commit  = "unknown"
	date    = "unknown"

	// Global flags
	jsonOutput bool
	verbose    bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "t42",
	Short: "A CLI tool for interacting with the 42 API",
	Long: `t42-cli is a command-line interface for the 42 School API.
It allows you to manage your projects, view user information, 
and interact with the 42 ecosystem from your terminal.

Examples:
  t42 auth login              # Login to your 42 account
  t42 project list            # List your projects
  t42 project show libft      # Show details for a specific project
  t42 auth status             # Check your authentication status`,

	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// Version flag (for convenience)
	var versionFlag bool
	rootCmd.Flags().BoolVar(&versionFlag, "version", false, "Print version information")

	// Override the default run behavior to handle --version
	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		if versionFlag {
			versionCmd.Run(cmd, args)
			return
		}
		// If no subcommand is provided and no version flag, show help
		if err := cmd.Help(); err != nil {
			fmt.Fprintf(os.Stderr, "Error displaying help: %v\n", err)
		}
	}

	// Version command
	rootCmd.AddCommand(versionCmd)
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Print the version, commit hash, and build date of t42-cli.`,
	Run: func(cmd *cobra.Command, args []string) {
		if jsonOutput {
			fmt.Printf(`{"version":"%s","commit":"%s","date":"%s"}%s`, version, commit, date, "\n")
		} else {
			fmt.Printf("t42-cli version %s\n", version)
			fmt.Printf("Commit: %s\n", commit)
			fmt.Printf("Built: %s\n", date)
		}
	},
}

// GetJSONOutput returns the current state of the json flag
func GetJSONOutput() bool {
	return jsonOutput
}

// GetVerbose returns the current state of the verbose flag
func GetVerbose() bool {
	return verbose
}

// NewAPIClient creates a new API client with automatic token refresh
func NewAPIClient() (*api.Client, error) {
	// Load credentials
	credentials, err := config.LoadCredentials()
	if err != nil {
		return nil, fmt.Errorf("not authenticated - please run 't42 auth login' first: %w", err)
	}

	// Check if we need to refresh the token proactively
	if config.NeedsRefresh(credentials) {
		if err := RefreshTokenIfNeeded(); err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}
		// Reload credentials after refresh
		credentials, err = config.LoadCredentials()
		if err != nil {
			return nil, fmt.Errorf("failed to reload credentials after refresh: %w", err)
		}
	}

	// Create client with token refresher callback
	client := api.NewClient(
		credentials.AccessToken,
		api.WithTokenRefresher(func() (string, error) {
			// This callback will be called when the API returns 401
			if err := RefreshTokenIfNeeded(); err != nil {
				return "", err
			}

			// Load the new credentials
			newCreds, err := config.LoadCredentials()
			if err != nil {
				return "", err
			}

			return newCreds.AccessToken, nil
		}),
	)

	return client, nil
}

// RequireAuth ensures the user is authenticated and returns an API client
func RequireAuth(ctx context.Context) (*api.Client, error) {
	client, err := NewAPIClient()
	if err != nil {
		return nil, err
	}

	// Verify authentication works
	if !client.IsAuthenticated(ctx) {
		return nil, fmt.Errorf("authentication failed - please run 't42 auth login' again")
	}

	return client, nil
}
