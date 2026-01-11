package companies

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
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
	err = deleteCmd.Args(deleteCmd, []string{"company-123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}

	// Test with two args
	err = deleteCmd.Args(deleteCmd, []string{"id-1", "id-2"})
	if err == nil {
		t.Error("Expected error with two args")
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
	origForce := forceDelete
	defer func() { forceDelete = origForce }()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	// Set force to false
	forceDelete = false

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runDelete(deleteCmd, []string{"test-id"})
	w.Close()
	os.Stdout = oldStdout

	// When force is false, it should print a message and return nil
	if err != nil {
		t.Fatalf("runDelete() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Are you sure") {
		t.Errorf("output missing confirmation prompt: %s", output)
	}
	if !strings.Contains(output, "--force") {
		t.Errorf("output missing '--force': %s", output)
	}
}

func TestRunDelete_WithForce_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/rest/companies/company-123" {
			t.Errorf("expected path /rest/companies/company-123, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	// Save original value
	origForce := forceDelete
	defer func() { forceDelete = origForce }()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	forceDelete = true

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runDelete(deleteCmd, []string{"company-123"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runDelete() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Deleted company") {
		t.Errorf("output missing 'Deleted company': %s", output)
	}
	if !strings.Contains(output, "company-123") {
		t.Errorf("output missing 'company-123': %s", output)
	}
}

func TestRunDelete_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":{"message":"Company not found"}}`))
	}))
	defer server.Close()

	// Save original value
	origForce := forceDelete
	defer func() { forceDelete = origForce }()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	forceDelete = true

	err := runDelete(deleteCmd, []string{"non-existent"})
	if err == nil {
		t.Fatal("expected error for non-existent company")
	}

	if !strings.Contains(err.Error(), "failed to delete company") {
		t.Errorf("error message should contain 'failed to delete company', got: %v", err)
	}
}

func TestRunDelete_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":{"message":"Internal server error"}}`))
	}))
	defer server.Close()

	// Save original value
	origForce := forceDelete
	defer func() { forceDelete = origForce }()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	forceDelete = true

	err := runDelete(deleteCmd, []string{"company-123"})
	if err == nil {
		t.Fatal("expected error for server error response")
	}
}

func TestRunDelete_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":{"message":"Unauthorized"}}`))
	}))
	defer server.Close()

	// Save original value
	origForce := forceDelete
	defer func() { forceDelete = origForce }()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	forceDelete = true

	err := runDelete(deleteCmd, []string{"company-123"})
	if err == nil {
		t.Fatal("expected error for unauthorized response")
	}
}

func TestRunDelete_NoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Save original value
	origForce := forceDelete
	defer func() { forceDelete = origForce }()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	forceDelete = true

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runDelete(deleteCmd, []string{"company-204"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runDelete() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Deleted company") {
		t.Errorf("output missing 'Deleted company': %s", output)
	}
}

func TestRunDelete_ConfirmationMessage(t *testing.T) {
	// Save original value
	origForce := forceDelete
	defer func() { forceDelete = origForce }()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	forceDelete = false

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runDelete(deleteCmd, []string{"specific-company-id"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runDelete() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should include the company ID in the confirmation message
	if !strings.Contains(output, "specific-company-id") {
		t.Errorf("output missing company ID: %s", output)
	}
}
