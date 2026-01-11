package webhooks

import (
	"testing"
)

func TestCmd_Use(t *testing.T) {
	if Cmd.Use != "webhooks" {
		t.Errorf("Use = %q, want %q", Cmd.Use, "webhooks")
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

func TestCmd_Subcommands(t *testing.T) {
	subcommands := []string{"list", "create", "delete"}

	for _, name := range subcommands {
		found := false
		for _, cmd := range Cmd.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("subcommand %q not registered", name)
		}
	}
}

func TestCmd_HasListSubcommand(t *testing.T) {
	cmd, _, err := Cmd.Find([]string{"list"})
	if err != nil {
		t.Errorf("failed to find list subcommand: %v", err)
	}
	if cmd == nil {
		t.Error("list subcommand not found")
	}
}

func TestCmd_HasCreateSubcommand(t *testing.T) {
	cmd, _, err := Cmd.Find([]string{"create"})
	if err != nil {
		t.Errorf("failed to find create subcommand: %v", err)
	}
	if cmd == nil {
		t.Error("create subcommand not found")
	}
}

func TestCmd_HasDeleteSubcommand(t *testing.T) {
	cmd, _, err := Cmd.Find([]string{"delete"})
	if err != nil {
		t.Errorf("failed to find delete subcommand: %v", err)
	}
	if cmd == nil {
		t.Error("delete subcommand not found")
	}
}

func TestCmd_SubcommandCount(t *testing.T) {
	expected := 3 // list, create, delete
	actual := len(Cmd.Commands())
	if actual != expected {
		t.Errorf("subcommand count = %d, want %d", actual, expected)
	}
}

func TestCmd_ListSubcommandShortDescription(t *testing.T) {
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

func TestCmd_CreateSubcommandShortDescription(t *testing.T) {
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

func TestCmd_DeleteSubcommandShortDescription(t *testing.T) {
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
