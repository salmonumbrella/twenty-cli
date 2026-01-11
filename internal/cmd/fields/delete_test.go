package fields

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestDeleteCmd_Use(t *testing.T) {
	if deleteCmd.Use != "delete <field-id>" {
		t.Errorf("Use = %q, want %q", deleteCmd.Use, "delete <field-id>")
	}
}

func TestDeleteCmd_Short(t *testing.T) {
	if deleteCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestDeleteCmd_Flags(t *testing.T) {
	if deleteCmd.Flags().Lookup("force") == nil {
		t.Error("force flag not registered")
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
	err = deleteCmd.Args(deleteCmd, []string{"field-id"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}

	// Test with two args
	err = deleteCmd.Args(deleteCmd, []string{"field-1", "field-2"})
	if err == nil {
		t.Error("Expected error with two args")
	}
}

func TestRunDelete_WithoutForce(t *testing.T) {
	// Save original value
	originalForce := deleteForce
	defer func() { deleteForce = originalForce }()

	deleteForce = false

	// Capture stdout to verify message
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runDelete(deleteCmd, []string{"field-123"})
	w.Close()
	os.Stdout = oldStdout

	// Should return nil but print a message
	if err != nil {
		t.Fatalf("runDelete() returned error = %v, want nil", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "About to delete field field-123") {
		t.Errorf("output missing confirmation prompt: %s", output)
	}
	if !strings.Contains(output, "--force") {
		t.Errorf("output missing '--force' instruction: %s", output)
	}
}

func TestRunDelete_WithForce_Success(t *testing.T) {
	// Set up mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/rest/metadata/fields/field-123") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Set up environment and viper
	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	// Save and restore package-level variable
	originalForce := deleteForce
	defer func() { deleteForce = originalForce }()
	deleteForce = true

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runDelete(deleteCmd, []string{"field-123"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runDelete() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Deleted field field-123") {
		t.Errorf("output missing confirmation: %s", output)
	}
}

func TestRunDelete_NoToken(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	// Save and restore package-level variable
	originalForce := deleteForce
	defer func() { deleteForce = originalForce }()
	deleteForce = true

	err := runDelete(deleteCmd, []string{"field-123"})
	if err == nil {
		t.Fatal("expected error when no token is available")
	}
}

func TestRunDelete_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "field not found"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	// Save and restore package-level variable
	originalForce := deleteForce
	defer func() { deleteForce = originalForce }()
	deleteForce = true

	err := runDelete(deleteCmd, []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for API error")
	}
}

func TestRunDelete_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	// Save and restore package-level variable
	originalForce := deleteForce
	defer func() { deleteForce = originalForce }()
	deleteForce = true

	err := runDelete(deleteCmd, []string{"field-123"})
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestRunDelete_ForbiddenError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "cannot delete system field"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	// Save and restore package-level variable
	originalForce := deleteForce
	defer func() { deleteForce = originalForce }()
	deleteForce = true

	err := runDelete(deleteCmd, []string{"system-field"})
	if err == nil {
		t.Fatal("expected error for forbidden error")
	}
}

func TestRunDelete_MultipleFields(t *testing.T) {
	// First delete
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	// Save and restore package-level variable
	originalForce := deleteForce
	defer func() { deleteForce = originalForce }()
	deleteForce = true

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Delete first field
	err := runDelete(deleteCmd, []string{"field-1"})
	if err != nil {
		w.Close()
		os.Stdout = oldStdout
		t.Fatalf("first delete error = %v", err)
	}

	// Delete second field
	err = runDelete(deleteCmd, []string{"field-2"})
	if err != nil {
		w.Close()
		os.Stdout = oldStdout
		t.Fatalf("second delete error = %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Deleted field field-1") {
		t.Errorf("output missing 'Deleted field field-1': %s", output)
	}
	if !strings.Contains(output, "Deleted field field-2") {
		t.Errorf("output missing 'Deleted field field-2': %s", output)
	}
}
