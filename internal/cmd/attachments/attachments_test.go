package attachments

import (
	"testing"
)

func TestCmd_Use(t *testing.T) {
	if Cmd.Use != "attachments" {
		t.Errorf("Use = %q, want %q", Cmd.Use, "attachments")
	}
}

func TestCmd_Short(t *testing.T) {
	if Cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestCmd_Subcommands(t *testing.T) {
	subcommands := []string{"list", "get", "create", "update", "delete"}

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

func TestCmd_HasGetSubcommand(t *testing.T) {
	cmd, _, err := Cmd.Find([]string{"get"})
	if err != nil {
		t.Errorf("failed to find get subcommand: %v", err)
	}
	if cmd == nil {
		t.Error("get subcommand not found")
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

func TestCmd_HasUpdateSubcommand(t *testing.T) {
	cmd, _, err := Cmd.Find([]string{"update"})
	if err != nil {
		t.Errorf("failed to find update subcommand: %v", err)
	}
	if cmd == nil {
		t.Error("update subcommand not found")
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
	expected := 5
	actual := len(Cmd.Commands())
	if actual != expected {
		t.Errorf("subcommand count = %d, want %d", actual, expected)
	}
}
