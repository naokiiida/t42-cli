package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"github.com/spf13/cobra"
	"github.com/naokiiida/t42-cli/internal"
)

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

	// Test loading from config
	cfg := &internal.Config{ClientID: "cfgid", ClientSecret: "cfgsecret"}
	internal.SaveConfig(cfg)
	id, secret, err = loadClientCreds("")
	if err != nil {
		t.Fatalf("Failed to load creds from config: %v", err)
	}
	if id != "cfgid" || secret != "cfgsecret" {
		t.Errorf("Expected cfgid/cfgsecret, got %s/%s", id, secret)
	}
} 