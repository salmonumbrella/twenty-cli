package notes

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
	origForce := forceDelete
	defer func() { forceDelete = origForce }()

	forceDelete = false

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runDelete(deleteCmd, []string{"test-id"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runDelete() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Are you sure") {
		t.Errorf("output should contain confirmation prompt: %s", output)
	}
	if !strings.Contains(output, "--force") {
		t.Errorf("output should mention --force flag: %s", output)
	}
}

func TestRunDelete_WithForce_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/rest/notes/delete-id") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"deleteNote": {"id": "delete-id"}}}`))
	}))
	defer server.Close()

	origForce := forceDelete
	defer func() { forceDelete = origForce }()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	forceDelete = true

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runDelete(deleteCmd, []string{"delete-id"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runDelete() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Deleted note") {
		t.Errorf("output missing confirmation: %s", output)
	}
	if !strings.Contains(output, "delete-id") {
		t.Errorf("output missing note ID: %s", output)
	}
}

func TestRunDelete_NoToken(t *testing.T) {
	origForce := forceDelete
	defer func() { forceDelete = origForce }()

	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	forceDelete = true

	err := runDelete(deleteCmd, []string{"test-id"})
	if err == nil {
		t.Fatal("expected error when no token is available")
	}
}

func TestRunDelete_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "not found"}`))
	}))
	defer server.Close()

	origForce := forceDelete
	defer func() { forceDelete = origForce }()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	forceDelete = true

	err := runDelete(deleteCmd, []string{"nonexistent-id"})
	if err == nil {
		t.Fatal("expected error for API error response")
	}

	if !strings.Contains(err.Error(), "failed to delete note") {
		t.Errorf("error = %q, want error containing 'failed to delete note'", err.Error())
	}
}

func TestRunDelete_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	origForce := forceDelete
	defer func() { forceDelete = origForce }()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	forceDelete = true

	err := runDelete(deleteCmd, []string{"test-id"})
	if err == nil {
		t.Fatal("expected error for server error response")
	}
}
