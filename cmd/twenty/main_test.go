package main

import (
	"os"
	"testing"

	"github.com/salmonumbrella/twenty-cli/internal/cmd"
)

// Test coverage note for cmd/twenty/main.go:
//
// The main() function shows 0% coverage because it cannot be tested directly:
// 1. It calls os.Exit() which would terminate the test process
// 2. The version variable is set at build time via ldflags (-ldflags "-X main.version=...")
//
// This is a well-known limitation in Go testing for main packages.
// The actual logic exercised by main() is minimal and tested here indirectly:
// - cmd.SetVersion() - tested via TestSetVersionIntegration
// - cmd.Execute() - tested extensively in internal/cmd package
//
// The main() function is effectively just a 4-line glue function that
// wires together tested components.

// TestMainPackageCompiles verifies the main package compiles correctly
// and all imports resolve.
func TestMainPackageCompiles(t *testing.T) {
	// This test passes if the package compiles successfully.
	// It validates that:
	// - The package declaration is correct
	// - All imports are resolvable
	// - The version variable is defined
	// - The main function signature is correct
}

// TestVersionVariable verifies the version variable has a default value.
func TestVersionVariable(t *testing.T) {
	if version == "" {
		t.Error("version should have a default value")
	}
	if version != "dev" {
		t.Errorf("version = %q, want %q (default)", version, "dev")
	}
}

// TestVersionVariableType verifies version is a string type.
func TestVersionVariableType(t *testing.T) {
	// Compile-time check that version is a string
	var _ string = version
}

// TestSetVersionIntegration verifies cmd.SetVersion can be called
// without panicking. This mirrors what main() does before Execute().
func TestSetVersionIntegration(t *testing.T) {
	// Save original version
	originalVersion := version

	// Test that SetVersion doesn't panic with various inputs
	testCases := []string{
		"1.0.0",
		"v1.0.0",
		"dev",
		"1.0.0-beta.1",
		"1.0.0+build.123",
		"",
	}

	for _, tc := range testCases {
		cmd.SetVersion(tc)
	}

	// Restore for other tests
	cmd.SetVersion(originalVersion)
}

// TestExecuteReturnsError verifies cmd.Execute returns an error type.
// We don't actually execute because it would run the full CLI.
func TestExecuteReturnsError(t *testing.T) {
	// Verify Execute() returns error type (compile-time check)
	var executeFunc func() error = cmd.Execute
	_ = executeFunc
}

// TestOsExitAvailable verifies os.Exit is available for main().
// This is a compile-time check that the os package is properly imported.
func TestOsExitAvailable(t *testing.T) {
	// Verify os.Exit has correct signature (compile-time check)
	var exitFunc func(int) = os.Exit
	_ = exitFunc
}

// TestMainDependencies verifies all dependencies used by main() are available.
func TestMainDependencies(t *testing.T) {
	t.Run("cmd.SetVersion exists", func(t *testing.T) {
		// SetVersion should be callable
		cmd.SetVersion("test")
		cmd.SetVersion(version) // restore
	})

	t.Run("cmd.Execute exists", func(t *testing.T) {
		// Execute should exist and return error
		var _ func() error = cmd.Execute
	})

	t.Run("os.Exit exists", func(t *testing.T) {
		// os.Exit should be available
		var _ func(int) = os.Exit
	})
}
