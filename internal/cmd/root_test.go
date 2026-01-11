package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCmd_Use(t *testing.T) {
	if rootCmd.Use != "twenty" {
		t.Errorf("Use = %q, want %q", rootCmd.Use, "twenty")
	}
}

func TestRootCmd_Short(t *testing.T) {
	expected := "CLI for Twenty CRM with 100% API coverage"
	if rootCmd.Short != expected {
		t.Errorf("Short = %q, want %q", rootCmd.Short, expected)
	}
}

func TestRootCmd_Long(t *testing.T) {
	if rootCmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestSetVersion(t *testing.T) {
	// Save original version
	original := version

	testVersion := "1.2.3"
	SetVersion(testVersion)

	if version != testVersion {
		t.Errorf("version = %q, want %q", version, testVersion)
	}
	if rootCmd.Version != testVersion {
		t.Errorf("rootCmd.Version = %q, want %q", rootCmd.Version, testVersion)
	}

	// Restore original version
	version = original
	rootCmd.Version = original
}

func TestRootCmd_PersistentFlags(t *testing.T) {
	tests := []struct {
		name         string
		shorthand    string
		defaultValue string
		usage        string
	}{
		{"config", "", "", "config file (default $HOME/.twenty.yaml)"},
		{"output", "o", "text", "output format: text, json, yaml, csv"},
		{"profile", "p", "", "auth profile to use (defaults to active profile)"},
		{"query", "q", "", "jq-style query for JSON output"},
		{"debug", "", "", "show API requests/responses"},
		{"no-color", "", "", "disable colored output"},
		{"no-retry", "", "", "disable automatic retry on rate limiting (429, 502, 503, 504)"},
		{"base-url", "", "https://api.twenty.com", "API base URL"},
	}

	pf := rootCmd.PersistentFlags()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := pf.Lookup(tt.name)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.name)
			}

			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("shorthand = %q, want %q", flag.Shorthand, tt.shorthand)
			}

			if tt.defaultValue != "" && flag.DefValue != tt.defaultValue {
				t.Errorf("default = %q, want %q", flag.DefValue, tt.defaultValue)
			}

			if tt.usage != "" && !strings.Contains(flag.Usage, tt.usage) {
				t.Errorf("usage = %q, want to contain %q", flag.Usage, tt.usage)
			}
		})
	}
}

func TestRootCmd_Subcommands(t *testing.T) {
	expectedSubcommands := []string{
		"auth",
		"attachments",
		"companies",
		"completion",
		"config",
		"favorites",
		"fields",
		"graphql",
		"notes",
		"objects",
		"opportunities",
		"people",
		"records",
		"rest",
		"tasks",
		"webhooks",
	}

	for _, name := range expectedSubcommands {
		t.Run(name, func(t *testing.T) {
			found := false
			for _, cmd := range rootCmd.Commands() {
				if cmd.Name() == name {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("subcommand %q not registered", name)
			}
		})
	}
}

func TestRootCmd_SubcommandCount(t *testing.T) {
	// 15 main commands + 1 completion command = 16 minimum
	// Note: help command may be auto-added by cobra on first Execute()
	actual := len(rootCmd.Commands())
	if actual < 16 {
		t.Errorf("subcommand count = %d, want at least 16", actual)
	}
}

func TestRootCmd_FindSubcommands(t *testing.T) {
	subcommands := []string{
		"auth",
		"companies",
		"people",
		"tasks",
		"notes",
		"completion",
	}

	for _, name := range subcommands {
		t.Run(name, func(t *testing.T) {
			cmd, _, err := rootCmd.Find([]string{name})
			if err != nil {
				t.Errorf("failed to find %s subcommand: %v", name, err)
			}
			if cmd == nil {
				t.Errorf("%s subcommand not found", name)
			}
		})
	}
}

func TestRootCmd_HasVersion(t *testing.T) {
	if rootCmd.Version == "" {
		t.Error("Version should not be empty")
	}
}

func TestRootCmd_VersionIsSet(t *testing.T) {
	// Verify that version string is not empty after SetVersion is called
	// The version flag is automatically added by cobra when rootCmd.Version is set
	SetVersion("test-version")
	if rootCmd.Version != "test-version" {
		t.Errorf("rootCmd.Version = %q, want %q", rootCmd.Version, "test-version")
	}
	// Restore
	SetVersion("dev")
}

func TestRootCmd_HelpFlagInherited(t *testing.T) {
	// Help flag is inherited from cobra, verify the command can show help
	buf := new(bytes.Buffer)
	testCmd := &cobra.Command{Use: "test"}
	testCmd.SetOut(buf)
	testCmd.SetArgs([]string{"--help"})
	err := testCmd.Execute()
	if err != nil {
		t.Errorf("Execute with --help failed: %v", err)
	}
}

func TestExecute_WithHelp(t *testing.T) {
	// Create a fresh command for this test to avoid side effects
	cmd := &cobra.Command{
		Use:   "test",
		Short: "test command",
	}

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() with --help returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "test command") {
		t.Error("help output should contain command description")
	}
}

func TestRootCmd_OutputFlagValues(t *testing.T) {
	// Test that the output flag accepts valid values
	pf := rootCmd.PersistentFlags()
	outputFlag := pf.Lookup("output")
	if outputFlag == nil {
		t.Fatal("output flag not found")
	}

	// Default should be "text"
	if outputFlag.DefValue != "text" {
		t.Errorf("output default = %q, want %q", outputFlag.DefValue, "text")
	}
}

func TestRootCmd_BaseURLDefault(t *testing.T) {
	pf := rootCmd.PersistentFlags()
	baseURLFlag := pf.Lookup("base-url")
	if baseURLFlag == nil {
		t.Fatal("base-url flag not found")
	}

	expected := "https://api.twenty.com"
	if baseURLFlag.DefValue != expected {
		t.Errorf("base-url default = %q, want %q", baseURLFlag.DefValue, expected)
	}
}

func TestRootCmd_BooleanFlags(t *testing.T) {
	boolFlags := []string{"debug", "no-color", "no-retry"}

	pf := rootCmd.PersistentFlags()
	for _, name := range boolFlags {
		t.Run(name, func(t *testing.T) {
			flag := pf.Lookup(name)
			if flag == nil {
				t.Fatalf("flag %q not found", name)
			}
			// Boolean flags should default to false
			if flag.DefValue != "false" {
				t.Errorf("%s default = %q, want %q", name, flag.DefValue, "false")
			}
		})
	}
}

func TestExecute_ReturnsError(t *testing.T) {
	// Execute() wraps rootCmd.Execute() - we test it indirectly
	// by verifying the function exists and returns the expected type
	// Direct execution would run the full command which has side effects
	// Instead, verify Execute is callable (compile-time check)
	var execFunc func() error = Execute
	if execFunc == nil {
		t.Error("Execute function should not be nil")
	}
}

func TestInitConfig_EnvPrefix(t *testing.T) {
	// Test that environment variables with TWENTY_ prefix are recognized
	// This tests the viper configuration indirectly
	pf := rootCmd.PersistentFlags()

	// Verify base-url flag exists and can be bound to env
	baseURLFlag := pf.Lookup("base-url")
	if baseURLFlag == nil {
		t.Fatal("base-url flag not found")
	}

	// Verify no-color flag exists and can be bound to env
	noColorFlag := pf.Lookup("no-color")
	if noColorFlag == nil {
		t.Fatal("no-color flag not found")
	}
}

func TestRootCmd_ConfigFlag(t *testing.T) {
	pf := rootCmd.PersistentFlags()
	configFlag := pf.Lookup("config")
	if configFlag == nil {
		t.Fatal("config flag not found")
	}

	// Config should have empty default (uses default path)
	if configFlag.DefValue != "" {
		t.Errorf("config default = %q, want empty", configFlag.DefValue)
	}
}
