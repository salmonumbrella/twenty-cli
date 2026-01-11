package builder

import (
	"context"
	"testing"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func mockGetPerson(ctx context.Context, client *rest.Client, id string) (*types.Person, error) {
	return &types.Person{
		ID:   id,
		Name: types.Name{FirstName: "John", LastName: "Doe"},
	}, nil
}

func TestGetCommand(t *testing.T) {
	cfg := GetConfig[types.Person]{
		Use:          "get",
		Short:        "Get a person",
		GetFunc:      mockGetPerson,
		TableHeaders: []string{"FIELD", "VALUE"},
		TableRows: func(p types.Person) [][]string {
			return [][]string{
				{"ID", p.ID},
				{"Name", p.Name.FirstName + " " + p.Name.LastName},
			}
		},
	}

	cmd := NewGetCommand(cfg)

	if cmd.Use != "get <id>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "get <id>")
	}
}

func TestGetCommand_Short(t *testing.T) {
	cfg := GetConfig[types.Person]{
		Use:          "get",
		Short:        "Get a person by ID",
		GetFunc:      mockGetPerson,
		TableHeaders: []string{"FIELD", "VALUE"},
		TableRows: func(p types.Person) [][]string {
			return [][]string{
				{"ID", p.ID},
			}
		},
	}

	cmd := NewGetCommand(cfg)

	if cmd.Short != "Get a person by ID" {
		t.Errorf("Short = %q, want %q", cmd.Short, "Get a person by ID")
	}
}

func TestGetCommand_Args(t *testing.T) {
	cfg := GetConfig[types.Person]{
		Use:          "get",
		Short:        "Get a person",
		GetFunc:      mockGetPerson,
		TableHeaders: []string{"FIELD", "VALUE"},
		TableRows: func(p types.Person) [][]string {
			return [][]string{
				{"ID", p.ID},
			}
		},
	}

	cmd := NewGetCommand(cfg)

	// Command should require exactly 1 argument
	if cmd.Args == nil {
		t.Error("Args validator not set")
	}

	// Test with zero args
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("Expected error with zero args")
	}

	// Test with one arg
	err = cmd.Args(cmd, []string{"id-123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}

	// Test with two args
	err = cmd.Args(cmd, []string{"id-1", "id-2"})
	if err == nil {
		t.Error("Expected error with two args")
	}
}

func TestNewGetCommand_PanicsOnNilGetFunc(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when GetFunc is nil")
		} else if r != "builder: GetConfig.GetFunc is required" {
			t.Errorf("unexpected panic message: %v", r)
		}
	}()

	cfg := GetConfig[types.Person]{
		Use:   "get",
		Short: "Get a person",
		// GetFunc is nil
		TableHeaders: []string{"FIELD", "VALUE"},
		TableRows: func(p types.Person) [][]string {
			return [][]string{{"ID", p.ID}}
		},
	}

	NewGetCommand(cfg)
}

func TestNewGetCommand_PanicsOnNilTableRows(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when TableRows is nil")
		} else if r != "builder: GetConfig.TableRows is required" {
			t.Errorf("unexpected panic message: %v", r)
		}
	}()

	cfg := GetConfig[types.Person]{
		Use:          "get",
		Short:        "Get a person",
		GetFunc:      mockGetPerson,
		TableHeaders: []string{"FIELD", "VALUE"},
		// TableRows is nil
	}

	NewGetCommand(cfg)
}
