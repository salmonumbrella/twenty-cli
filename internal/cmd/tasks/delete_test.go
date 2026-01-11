package tasks

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
	// Test that command requires exactly one argument
	err := deleteCmd.Args(deleteCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = deleteCmd.Args(deleteCmd, []string{"id1", "id2"})
	if err == nil {
		t.Error("expected error when too many args provided")
	}

	err = deleteCmd.Args(deleteCmd, []string{"id1"})
	if err != nil {
		t.Errorf("unexpected error for valid args: %v", err)
	}
}

func TestRunDelete_Force(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/rest/tasks/task-to-delete") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"deleteTask": {"id": "task-to-delete"}}}`))
	}))
	defer server.Close()

	// Set force flag
	forceDelete = true

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runDelete(deleteCmd, []string{"task-to-delete"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runDelete() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Deleted task") {
		t.Errorf("output missing confirmation: %s", output)
	}
	if !strings.Contains(output, "task-to-delete") {
		t.Errorf("output missing task ID: %s", output)
	}
}

func TestRunDelete_NoForcePrompt(t *testing.T) {
	// When force is false, should prompt for confirmation
	forceDelete = false

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runDelete(deleteCmd, []string{"task-123"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runDelete() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Are you sure") {
		t.Errorf("output missing confirmation prompt: %s", output)
	}
	if !strings.Contains(output, "task-123") {
		t.Errorf("output missing task ID in prompt: %s", output)
	}
	if !strings.Contains(output, "--force") {
		t.Errorf("output missing --force hint: %s", output)
	}
}

func TestRunDelete_NoToken(t *testing.T) {
	forceDelete = true

	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	err := runDelete(deleteCmd, []string{"task-123"})
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

	forceDelete = true

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	err := runDelete(deleteCmd, []string{"nonexistent-task"})
	if err == nil {
		t.Fatal("expected error for API error response")
	}
	if !strings.Contains(err.Error(), "failed to delete task") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRunDelete_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	forceDelete = true

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	err := runDelete(deleteCmd, []string{"task-123"})
	if err == nil {
		t.Fatal("expected error for server error response")
	}
}
