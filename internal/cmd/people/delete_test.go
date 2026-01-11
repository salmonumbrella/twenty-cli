package people

import (
	"testing"
)

func TestDeleteCmd_Flags(t *testing.T) {
	if deleteCmd.Flags().Lookup("force") == nil {
		t.Error("force flag not registered")
	}
}

func TestDeleteCmd_ForceFlagShorthand(t *testing.T) {
	flag := deleteCmd.Flags().Lookup("force")
	if flag == nil {
		t.Fatal("force flag not registered")
	}
	if flag.Shorthand != "f" {
		t.Errorf("force flag shorthand = %q, want %q", flag.Shorthand, "f")
	}
}

func TestDeleteCmd_Use(t *testing.T) {
	if deleteCmd.Use != "delete <id>" {
		t.Errorf("Use = %q, want %q", deleteCmd.Use, "delete <id>")
	}
}

func TestDeleteCmd_Short(t *testing.T) {
	if deleteCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestDeleteCmd_Long(t *testing.T) {
	if deleteCmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestDeleteCmd_Args(t *testing.T) {
	// Command should require exactly 1 argument
	if deleteCmd.Args == nil {
		t.Error("Args validator not set")
	}

	// Test with zero args
	err := deleteCmd.Args(deleteCmd, []string{})
	if err == nil {
		t.Error("Expected error with zero args")
	}

	// Test with one arg
	err = deleteCmd.Args(deleteCmd, []string{"id-123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}

	// Test with two args
	err = deleteCmd.Args(deleteCmd, []string{"id-1", "id-2"})
	if err == nil {
		t.Error("Expected error with two args")
	}
}

func TestDeleteCmd_SilenceUsage(t *testing.T) {
	if !deleteCmd.SilenceUsage {
		t.Error("SilenceUsage should be true")
	}
}

func TestDeleteCmd_ForceDefaultValue(t *testing.T) {
	flag := deleteCmd.Flags().Lookup("force")
	if flag == nil {
		t.Fatal("force flag not registered")
	}
	if flag.DefValue != "false" {
		t.Errorf("force flag default = %q, want %q", flag.DefValue, "false")
	}
}

func TestRunDelete_WithoutForce(t *testing.T) {
	// Save original value
	originalForce := forceDelete
	defer func() { forceDelete = originalForce }()

	// Set force to false
	forceDelete = false

	// runDelete should return error when force is not set
	err := runDelete(deleteCmd, []string{"test-id"})
	if err == nil {
		t.Error("Expected error when --force is not set")
	}

	errStr := err.Error()
	if errStr != "delete aborted: use --force to confirm deletion of test-id" {
		t.Errorf("Error message = %q, want contains 'use --force'", errStr)
	}
}
