package fields

import (
	"testing"
)

func TestCmd_Use(t *testing.T) {
	if Cmd.Use != "fields" {
		t.Errorf("Cmd.Use = %q, want %q", Cmd.Use, "fields")
	}
}

func TestCmd_Short(t *testing.T) {
	if Cmd.Short == "" {
		t.Error("Cmd.Short description should not be empty")
	}
}

func TestCmd_Long(t *testing.T) {
	if Cmd.Long == "" {
		t.Error("Cmd.Long description should not be empty")
	}
}

func TestCmd_HasSubcommands(t *testing.T) {
	subcommands := Cmd.Commands()
	if len(subcommands) == 0 {
		t.Error("Cmd should have subcommands")
	}

	// Check that all expected subcommands are registered
	expectedCmds := map[string]bool{
		"list":   false,
		"get":    false,
		"create": false,
		"update": false,
		"delete": false,
	}

	for _, cmd := range subcommands {
		if _, ok := expectedCmds[cmd.Name()]; ok {
			expectedCmds[cmd.Name()] = true
		}
	}

	for name, found := range expectedCmds {
		if !found {
			t.Errorf("expected subcommand %q not found", name)
		}
	}
}

func TestCmd_ListSubcommand(t *testing.T) {
	cmd, _, err := Cmd.Find([]string{"list"})
	if err != nil {
		t.Fatalf("failed to find 'list' subcommand: %v", err)
	}
	if cmd.Name() != "list" {
		t.Errorf("found wrong command: %s", cmd.Name())
	}
}

func TestCmd_GetSubcommand(t *testing.T) {
	cmd, _, err := Cmd.Find([]string{"get"})
	if err != nil {
		t.Fatalf("failed to find 'get' subcommand: %v", err)
	}
	if cmd.Name() != "get" {
		t.Errorf("found wrong command: %s", cmd.Name())
	}
}

func TestCmd_CreateSubcommand(t *testing.T) {
	cmd, _, err := Cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("failed to find 'create' subcommand: %v", err)
	}
	if cmd.Name() != "create" {
		t.Errorf("found wrong command: %s", cmd.Name())
	}
}

func TestCmd_UpdateSubcommand(t *testing.T) {
	cmd, _, err := Cmd.Find([]string{"update"})
	if err != nil {
		t.Fatalf("failed to find 'update' subcommand: %v", err)
	}
	if cmd.Name() != "update" {
		t.Errorf("found wrong command: %s", cmd.Name())
	}
}

func TestCmd_DeleteSubcommand(t *testing.T) {
	cmd, _, err := Cmd.Find([]string{"delete"})
	if err != nil {
		t.Fatalf("failed to find 'delete' subcommand: %v", err)
	}
	if cmd.Name() != "delete" {
		t.Errorf("found wrong command: %s", cmd.Name())
	}
}
