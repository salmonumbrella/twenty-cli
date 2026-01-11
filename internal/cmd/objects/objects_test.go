package objects

import (
	"testing"
)

func TestCmd_Use(t *testing.T) {
	if Cmd.Use != "objects" {
		t.Errorf("Use = %q, want %q", Cmd.Use, "objects")
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
	subcommands := Cmd.Commands()
	if len(subcommands) == 0 {
		t.Error("Cmd should have subcommands")
	}
}

func TestCmd_HasListSubcommand(t *testing.T) {
	found := false
	for _, cmd := range Cmd.Commands() {
		if cmd.Use == "list" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Cmd should have 'list' subcommand")
	}
}

func TestCmd_HasGetSubcommand(t *testing.T) {
	found := false
	for _, cmd := range Cmd.Commands() {
		if cmd.Use == "get <objectName>" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Cmd should have 'get' subcommand")
	}
}

func TestCmd_HasCreateSubcommand(t *testing.T) {
	found := false
	for _, cmd := range Cmd.Commands() {
		if cmd.Use == "create" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Cmd should have 'create' subcommand")
	}
}

func TestCmd_HasUpdateSubcommand(t *testing.T) {
	found := false
	for _, cmd := range Cmd.Commands() {
		if cmd.Use == "update <object-id>" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Cmd should have 'update' subcommand")
	}
}

func TestCmd_HasDeleteSubcommand(t *testing.T) {
	found := false
	for _, cmd := range Cmd.Commands() {
		if cmd.Use == "delete <object-id>" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Cmd should have 'delete' subcommand")
	}
}

func TestCmd_SubcommandCount(t *testing.T) {
	expected := 5 // list, get, create, update, delete
	actual := len(Cmd.Commands())
	if actual != expected {
		t.Errorf("expected %d subcommands, got %d", expected, actual)
	}
}
