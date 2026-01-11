package graphql

import (
	"testing"
)

func TestCmd_Use(t *testing.T) {
	if Cmd.Use != "graphql" {
		t.Errorf("Use = %q, want %q", Cmd.Use, "graphql")
	}
}

func TestCmd_Short(t *testing.T) {
	if Cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestCmd_HasSubcommands(t *testing.T) {
	subcommands := []string{"query", "mutate", "schema"}
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
	expected := 3 // query, mutate, schema
	if len(Cmd.Commands()) != expected {
		t.Errorf("expected %d subcommands, got %d", expected, len(Cmd.Commands()))
	}
}

func TestCmd_QuerySubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "query" {
			if cmd.Short == "" {
				t.Error("query subcommand should have short description")
			}
			return
		}
	}
	t.Error("query subcommand not found")
}

func TestCmd_MutateSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "mutate" {
			if cmd.Short == "" {
				t.Error("mutate subcommand should have short description")
			}
			return
		}
	}
	t.Error("mutate subcommand not found")
}

func TestCmd_SchemaSubcommand(t *testing.T) {
	for _, cmd := range Cmd.Commands() {
		if cmd.Name() == "schema" {
			if cmd.Short == "" {
				t.Error("schema subcommand should have short description")
			}
			return
		}
	}
	t.Error("schema subcommand not found")
}

func TestCmd_FindQuerySubcommand(t *testing.T) {
	cmd, _, err := Cmd.Find([]string{"query"})
	if err != nil {
		t.Errorf("failed to find query subcommand: %v", err)
	}
	if cmd == nil {
		t.Error("query subcommand not found")
	}
}

func TestCmd_FindMutateSubcommand(t *testing.T) {
	cmd, _, err := Cmd.Find([]string{"mutate"})
	if err != nil {
		t.Errorf("failed to find mutate subcommand: %v", err)
	}
	if cmd == nil {
		t.Error("mutate subcommand not found")
	}
}

func TestCmd_FindSchemaSubcommand(t *testing.T) {
	cmd, _, err := Cmd.Find([]string{"schema"})
	if err != nil {
		t.Errorf("failed to find schema subcommand: %v", err)
	}
	if cmd == nil {
		t.Error("schema subcommand not found")
	}
}
