package config

import (
	"testing"
)

func TestCmd_Use(t *testing.T) {
	if Cmd.Use != "config" {
		t.Errorf("Use = %q, want %q", Cmd.Use, "config")
	}
}

func TestCmd_Short(t *testing.T) {
	if Cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestCmd_HasSubcommands(t *testing.T) {
	subcommands := []string{"show", "set"}
	for _, sub := range subcommands {
		found := false
		for _, cmd := range Cmd.Commands() {
			if cmd.Name() == sub {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("subcommand %q not found", sub)
		}
	}
}

func TestCmd_SubcommandCount(t *testing.T) {
	expected := 2 // show, set
	if len(Cmd.Commands()) != expected {
		t.Errorf("expected %d subcommands, got %d", expected, len(Cmd.Commands()))
	}
}

func TestCmd_ShowSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "show" {
			if cmd.Short == "" {
				t.Error("show subcommand should have short description")
			}
			return
		}
	}
	t.Error("show subcommand not found")
}

func TestCmd_SetSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "set" {
			if cmd.Short == "" {
				t.Error("set subcommand should have short description")
			}
			return
		}
	}
	t.Error("set subcommand not found")
}
