package tasks

import (
	"testing"
)

func TestCmd_Use(t *testing.T) {
	if Cmd.Use != "tasks" {
		t.Errorf("Use = %q, want %q", Cmd.Use, "tasks")
	}
}

func TestCmd_Short(t *testing.T) {
	if Cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestCmd_Long(t *testing.T) {
	if Cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestCmd_HasSubcommands(t *testing.T) {
	subcommands := []string{"list", "get", "create", "update", "delete"}
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
	expected := 5 // list, get, create, update, delete
	if len(Cmd.Commands()) != expected {
		t.Errorf("expected %d subcommands, got %d", expected, len(Cmd.Commands()))
	}
}

func TestCmd_ListSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "list" {
			if cmd.Short == "" {
				t.Error("list subcommand should have short description")
			}
			return
		}
	}
	t.Error("list subcommand not found")
}

func TestCmd_GetSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "get" {
			if cmd.Short == "" {
				t.Error("get subcommand should have short description")
			}
			return
		}
	}
	t.Error("get subcommand not found")
}

func TestCmd_CreateSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "create" {
			if cmd.Short == "" {
				t.Error("create subcommand should have short description")
			}
			return
		}
	}
	t.Error("create subcommand not found")
}

func TestCmd_UpdateSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "update" {
			if cmd.Short == "" {
				t.Error("update subcommand should have short description")
			}
			return
		}
	}
	t.Error("update subcommand not found")
}

func TestCmd_DeleteSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "delete" {
			if cmd.Short == "" {
				t.Error("delete subcommand should have short description")
			}
			return
		}
	}
	t.Error("delete subcommand not found")
}
