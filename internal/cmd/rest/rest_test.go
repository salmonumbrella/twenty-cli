package rest

import (
	"testing"
)

func TestCmd_Use(t *testing.T) {
	if Cmd.Use != "rest" {
		t.Errorf("Use = %q, want %q", Cmd.Use, "rest")
	}
}

func TestCmd_Short(t *testing.T) {
	if Cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestCmd_HasSubcommands(t *testing.T) {
	subcommands := []string{"request", "get", "post", "patch", "delete"}
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
	expected := 5 // request, get, post, patch, delete
	if len(Cmd.Commands()) != expected {
		t.Errorf("expected %d subcommands, got %d", expected, len(Cmd.Commands()))
	}
}

func TestCmd_RequestSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "request" {
			if cmd.Short == "" {
				t.Error("request subcommand should have short description")
			}
			if cmd.Use != "request <method> <path>" {
				t.Errorf("request Use = %q, want %q", cmd.Use, "request <method> <path>")
			}
			return
		}
	}
	t.Error("request subcommand not found")
}

func TestCmd_GetSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "get" {
			if cmd.Short == "" {
				t.Error("get subcommand should have short description")
			}
			if cmd.Use != "get <path>" {
				t.Errorf("get Use = %q, want %q", cmd.Use, "get <path>")
			}
			return
		}
	}
	t.Error("get subcommand not found")
}

func TestCmd_PostSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "post" {
			if cmd.Short == "" {
				t.Error("post subcommand should have short description")
			}
			if cmd.Use != "post <path>" {
				t.Errorf("post Use = %q, want %q", cmd.Use, "post <path>")
			}
			return
		}
	}
	t.Error("post subcommand not found")
}

func TestCmd_PatchSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "patch" {
			if cmd.Short == "" {
				t.Error("patch subcommand should have short description")
			}
			if cmd.Use != "patch <path>" {
				t.Errorf("patch Use = %q, want %q", cmd.Use, "patch <path>")
			}
			return
		}
	}
	t.Error("patch subcommand not found")
}

func TestCmd_DeleteSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "delete" {
			if cmd.Short == "" {
				t.Error("delete subcommand should have short description")
			}
			if cmd.Use != "delete <path>" {
				t.Errorf("delete Use = %q, want %q", cmd.Use, "delete <path>")
			}
			return
		}
	}
	t.Error("delete subcommand not found")
}
