package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCompletionCmd_Use(t *testing.T) {
	expected := "completion [bash|zsh|fish|powershell]"
	if completionCmd.Use != expected {
		t.Errorf("Use = %q, want %q", completionCmd.Use, expected)
	}
}

func TestCompletionCmd_Short(t *testing.T) {
	if completionCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
	if !strings.Contains(completionCmd.Short, "completion") {
		t.Error("Short description should mention completion")
	}
}

func TestCompletionCmd_Long(t *testing.T) {
	if completionCmd.Long == "" {
		t.Error("Long description should not be empty")
	}

	// Should contain instructions for each shell
	shells := []string{"Bash", "Zsh", "Fish", "PowerShell"}
	for _, shell := range shells {
		if !strings.Contains(completionCmd.Long, shell) {
			t.Errorf("Long description should contain instructions for %s", shell)
		}
	}
}

func TestCompletionCmd_ValidArgs(t *testing.T) {
	expected := []string{"bash", "zsh", "fish", "powershell"}

	if len(completionCmd.ValidArgs) != len(expected) {
		t.Errorf("ValidArgs count = %d, want %d", len(completionCmd.ValidArgs), len(expected))
	}

	for i, arg := range expected {
		if completionCmd.ValidArgs[i] != arg {
			t.Errorf("ValidArgs[%d] = %q, want %q", i, completionCmd.ValidArgs[i], arg)
		}
	}
}

func TestCompletionCmd_DisableFlagsInUseLine(t *testing.T) {
	if !completionCmd.DisableFlagsInUseLine {
		t.Error("DisableFlagsInUseLine should be true")
	}
}

func TestCompletionCmd_RequiresExactlyOneArg(t *testing.T) {
	// Test with no args
	cmd := &cobra.Command{Use: "test"}
	cmd.AddCommand(completionCmd)

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"completion"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error with no args")
	}
}

func TestCompletionCmd_RejectsInvalidArgs(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.AddCommand(completionCmd)

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"completion", "invalid"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error with invalid arg")
	}
}

func TestCompletionCmd_BashCompletion(t *testing.T) {
	// Test the completion command's RunE function directly
	buf := new(bytes.Buffer)

	// Call GenBashCompletion directly on rootCmd to verify it works
	err := rootCmd.GenBashCompletion(buf)
	if err != nil {
		t.Errorf("bash completion failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "bash") || !strings.Contains(output, "completion") {
		t.Error("bash completion output should contain bash completion script")
	}
}

func TestCompletionCmd_ZshCompletion(t *testing.T) {
	buf := new(bytes.Buffer)

	err := rootCmd.GenZshCompletion(buf)
	if err != nil {
		t.Errorf("zsh completion failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("zsh completion output should not be empty")
	}
}

func TestCompletionCmd_FishCompletion(t *testing.T) {
	buf := new(bytes.Buffer)

	err := rootCmd.GenFishCompletion(buf, true)
	if err != nil {
		t.Errorf("fish completion failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("fish completion output should not be empty")
	}
}

func TestCompletionCmd_PowerShellCompletion(t *testing.T) {
	buf := new(bytes.Buffer)

	err := rootCmd.GenPowerShellCompletionWithDesc(buf)
	if err != nil {
		t.Errorf("powershell completion failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("powershell completion output should not be empty")
	}
}

func TestCompletionCmd_IsRegistered(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"completion"})
	if err != nil {
		t.Errorf("failed to find completion subcommand: %v", err)
	}
	if cmd == nil {
		t.Error("completion subcommand not found")
	}
	if cmd.Name() != "completion" {
		t.Errorf("found command name = %q, want %q", cmd.Name(), "completion")
	}
}

func TestCompletionCmd_RunE_Bash(t *testing.T) {
	// Test the RunE function directly with bash argument
	err := completionCmd.RunE(completionCmd, []string{"bash"})
	if err != nil {
		t.Errorf("RunE with bash failed: %v", err)
	}
}

func TestCompletionCmd_RunE_Zsh(t *testing.T) {
	err := completionCmd.RunE(completionCmd, []string{"zsh"})
	if err != nil {
		t.Errorf("RunE with zsh failed: %v", err)
	}
}

func TestCompletionCmd_RunE_Fish(t *testing.T) {
	err := completionCmd.RunE(completionCmd, []string{"fish"})
	if err != nil {
		t.Errorf("RunE with fish failed: %v", err)
	}
}

func TestCompletionCmd_RunE_PowerShell(t *testing.T) {
	err := completionCmd.RunE(completionCmd, []string{"powershell"})
	if err != nil {
		t.Errorf("RunE with powershell failed: %v", err)
	}
}

func TestCompletionCmd_RunE_Unknown(t *testing.T) {
	// Test with an unknown shell (should return nil, no error)
	err := completionCmd.RunE(completionCmd, []string{"unknown"})
	if err != nil {
		t.Errorf("RunE with unknown shell should return nil, got: %v", err)
	}
}
