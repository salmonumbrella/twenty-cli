package builder

import (
	"context"
	"testing"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

// Mock list function
func mockListPeople(ctx context.Context, client *rest.Client, opts *rest.ListOptions) (*types.ListResponse[types.Person], error) {
	return &types.ListResponse[types.Person]{
		Data: []types.Person{
			{ID: "1", Name: types.Name{FirstName: "John", LastName: "Doe"}},
			{ID: "2", Name: types.Name{FirstName: "Jane", LastName: "Smith"}},
		},
		TotalCount: 2,
		PageInfo:   &types.PageInfo{HasNextPage: false},
	}, nil
}

func TestListCommand_TableOutput(t *testing.T) {
	cfg := ListConfig[types.Person]{
		Use:          "list",
		Short:        "List people",
		ListFunc:     mockListPeople,
		TableHeaders: []string{"ID", "NAME", "EMAIL"},
		TableRow: func(p types.Person) []string {
			return []string{p.ID, p.Name.FirstName + " " + p.Name.LastName, p.Email.PrimaryEmail}
		},
	}

	cmd := NewListCommand(cfg)

	if cmd.Use != "list" {
		t.Errorf("Use = %q, want %q", cmd.Use, "list")
	}
	if cmd.Short != "List people" {
		t.Errorf("Short = %q, want %q", cmd.Short, "List people")
	}

	// Verify flags are registered
	if cmd.Flags().Lookup("limit") == nil {
		t.Error("limit flag not registered")
	}
	if cmd.Flags().Lookup("cursor") == nil {
		t.Error("cursor flag not registered")
	}
	if cmd.Flags().Lookup("all") == nil {
		t.Error("all flag not registered")
	}
	if cmd.Flags().Lookup("filter") == nil {
		t.Error("filter flag not registered")
	}
	if cmd.Flags().Lookup("sort") == nil {
		t.Error("sort flag not registered")
	}
	if cmd.Flags().Lookup("order") == nil {
		t.Error("order flag not registered")
	}
}

func TestListConfig_ExtraFlags(t *testing.T) {
	var customFlag string

	cfg := ListConfig[types.Person]{
		Use:          "list",
		Short:        "List people",
		ListFunc:     mockListPeople,
		TableHeaders: []string{"ID", "NAME"},
		TableRow: func(p types.Person) []string {
			return []string{p.ID, p.Name.FirstName}
		},
		ExtraFlags: func(cmd *cobra.Command) {
			cmd.Flags().StringVar(&customFlag, "custom", "", "custom flag")
		},
	}

	cmd := NewListCommand(cfg)

	// Verify custom flag is registered
	if cmd.Flags().Lookup("custom") == nil {
		t.Error("custom flag not registered")
	}
}

func TestNewListCommand_PanicsOnNilListFunc(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when ListFunc is nil")
		} else if r != "builder: ListConfig.ListFunc is required" {
			t.Errorf("unexpected panic message: %v", r)
		}
	}()

	cfg := ListConfig[types.Person]{
		Use:   "list",
		Short: "List people",
		// ListFunc is nil
		TableHeaders: []string{"ID", "NAME"},
		TableRow: func(p types.Person) []string {
			return []string{p.ID, p.Name.FirstName}
		},
	}

	NewListCommand(cfg)
}

func TestNewListCommand_PanicsOnNilTableRow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when TableRow is nil")
		} else if r != "builder: ListConfig.TableRow is required" {
			t.Errorf("unexpected panic message: %v", r)
		}
	}()

	cfg := ListConfig[types.Person]{
		Use:          "list",
		Short:        "List people",
		ListFunc:     mockListPeople,
		TableHeaders: []string{"ID", "NAME"},
		// TableRow is nil
	}

	NewListCommand(cfg)
}
