package builder

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestCreateCommand(t *testing.T) {
	cfg := CreateConfig{
		Use:      "create",
		Short:    "Create a person",
		Resource: "person",
	}

	cmd := NewCreateCommand(cfg)

	if cmd.Use != "create" {
		t.Errorf("Use = %q, want %q", cmd.Use, "create")
	}

	// Verify required flags
	if cmd.Flags().Lookup("data") == nil {
		t.Error("data flag not registered")
	}
}

func TestCreateCommand_Short(t *testing.T) {
	cfg := CreateConfig{
		Use:      "create",
		Short:    "Create a new person",
		Resource: "person",
	}

	cmd := NewCreateCommand(cfg)

	if cmd.Short != "Create a new person" {
		t.Errorf("Short = %q, want %q", cmd.Short, "Create a new person")
	}
}

func TestCreateCommand_FileFlag(t *testing.T) {
	cfg := CreateConfig{
		Use:      "create",
		Short:    "Create a person",
		Resource: "person",
	}

	cmd := NewCreateCommand(cfg)

	if cmd.Flags().Lookup("file") == nil {
		t.Error("file flag not registered")
	}
}

func TestCreateCommand_ExtraFlags(t *testing.T) {
	cfg := CreateConfig{
		Use:      "create",
		Short:    "Create a person",
		Resource: "person",
		ExtraFlags: func(cmd *cobra.Command) {
			cmd.Flags().String("custom", "", "custom flag")
		},
	}

	cmd := NewCreateCommand(cfg)

	if cmd.Flags().Lookup("custom") == nil {
		t.Error("custom flag not registered")
	}
}

func TestUpdateCommand(t *testing.T) {
	cfg := UpdateConfig{
		Use:      "update",
		Short:    "Update a person",
		Resource: "person",
	}

	cmd := NewUpdateCommand(cfg)

	if cmd.Use != "update <id>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "update <id>")
	}
}

func TestUpdateCommand_Short(t *testing.T) {
	cfg := UpdateConfig{
		Use:      "update",
		Short:    "Update an existing person",
		Resource: "person",
	}

	cmd := NewUpdateCommand(cfg)

	if cmd.Short != "Update an existing person" {
		t.Errorf("Short = %q, want %q", cmd.Short, "Update an existing person")
	}
}

func TestUpdateCommand_DataFlag(t *testing.T) {
	cfg := UpdateConfig{
		Use:      "update",
		Short:    "Update a person",
		Resource: "person",
	}

	cmd := NewUpdateCommand(cfg)

	if cmd.Flags().Lookup("data") == nil {
		t.Error("data flag not registered")
	}
}

func TestUpdateCommand_FileFlag(t *testing.T) {
	cfg := UpdateConfig{
		Use:      "update",
		Short:    "Update a person",
		Resource: "person",
	}

	cmd := NewUpdateCommand(cfg)

	if cmd.Flags().Lookup("file") == nil {
		t.Error("file flag not registered")
	}
}

func TestUpdateCommand_Args(t *testing.T) {
	cfg := UpdateConfig{
		Use:      "update",
		Short:    "Update a person",
		Resource: "person",
	}

	cmd := NewUpdateCommand(cfg)

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

func TestUpdateCommand_ExtraFlags(t *testing.T) {
	cfg := UpdateConfig{
		Use:      "update",
		Short:    "Update a person",
		Resource: "person",
		ExtraFlags: func(cmd *cobra.Command) {
			cmd.Flags().Bool("merge", false, "merge update")
		},
	}

	cmd := NewUpdateCommand(cfg)

	if cmd.Flags().Lookup("merge") == nil {
		t.Error("merge flag not registered")
	}
}

func TestDeleteCommand(t *testing.T) {
	cfg := DeleteConfig{
		Use:      "delete",
		Short:    "Delete a person",
		Resource: "person",
	}

	cmd := NewDeleteCommand(cfg)

	if cmd.Use != "delete <id>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "delete <id>")
	}

	// Verify force flag
	if cmd.Flags().Lookup("force") == nil {
		t.Error("force flag not registered")
	}
}

func TestDeleteCommand_Short(t *testing.T) {
	cfg := DeleteConfig{
		Use:      "delete",
		Short:    "Delete an existing person",
		Resource: "person",
	}

	cmd := NewDeleteCommand(cfg)

	if cmd.Short != "Delete an existing person" {
		t.Errorf("Short = %q, want %q", cmd.Short, "Delete an existing person")
	}
}

func TestDeleteCommand_Args(t *testing.T) {
	cfg := DeleteConfig{
		Use:      "delete",
		Short:    "Delete a person",
		Resource: "person",
	}

	cmd := NewDeleteCommand(cfg)

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

func TestNewCreateCommand_PanicsOnEmptyResource(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when Resource is empty")
		} else if r != "builder: CreateConfig.Resource is required" {
			t.Errorf("unexpected panic message: %v", r)
		}
	}()

	cfg := CreateConfig{
		Use:   "create",
		Short: "Create a person",
		// Resource is empty
	}

	NewCreateCommand(cfg)
}

func TestNewUpdateCommand_PanicsOnEmptyResource(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when Resource is empty")
		} else if r != "builder: UpdateConfig.Resource is required" {
			t.Errorf("unexpected panic message: %v", r)
		}
	}()

	cfg := UpdateConfig{
		Use:   "update",
		Short: "Update a person",
		// Resource is empty
	}

	NewUpdateCommand(cfg)
}

func TestNewDeleteCommand_PanicsOnEmptyResource(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when Resource is empty")
		} else if r != "builder: DeleteConfig.Resource is required" {
			t.Errorf("unexpected panic message: %v", r)
		}
	}()

	cfg := DeleteConfig{
		Use:   "delete",
		Short: "Delete a person",
		// Resource is empty
	}

	NewDeleteCommand(cfg)
}
