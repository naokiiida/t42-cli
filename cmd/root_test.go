package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRoot_Help(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})
	_ = rootCmd.Execute()
	out := buf.String()
	if !strings.Contains(out, "A brief description") {
		t.Errorf("expected help output, got: %s", out)
	}
}

func TestRoot_UnknownCommand(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"notacommand"})
	err := rootCmd.Execute()
	if err == nil {
		t.Errorf("expected error for unknown command, got none")
	}
	out := buf.String()
	if !strings.Contains(out, "unknown command") && !strings.Contains(out, "Usage:") {
		t.Errorf("expected error or usage output, got: %s", out)
	}
} 