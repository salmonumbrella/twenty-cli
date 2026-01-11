package favorites

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
		if !strings.HasSuffix(r.URL.Path, "/rest/favorites/fav-to-delete") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "fav-to-delete"}`))
	}))
	defer server.Close()

	// Save and restore original flag value
	origForce := deleteForce
	defer func() {
		deleteForce = origForce
	}()

	// Set force flag
	deleteForce = true

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runDelete(deleteCmd, []string{"fav-to-delete"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runDelete() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Deleted favorite") {
		t.Errorf("output missing confirmation: %s", output)
	}
	if !strings.Contains(output, "fav-to-delete") {
		t.Errorf("output missing favorite ID: %s", output)
	}
}

func TestRunDelete_NoForcePrompt(t *testing.T) {
	// Save and restore original flag value
	origForce := deleteForce
	defer func() {
		deleteForce = origForce
	}()

	// When force is false, should prompt for confirmation
	deleteForce = false

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runDelete(deleteCmd, []string{"fav-123"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runDelete() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "About to delete") {
		t.Errorf("output missing confirmation prompt: %s", output)
	}
	if !strings.Contains(output, "fav-123") {
		t.Errorf("output missing favorite ID in prompt: %s", output)
	}
	if !strings.Contains(output, "--force") {
		t.Errorf("output missing --force hint: %s", output)
	}
}

func TestRunDelete_NoToken(t *testing.T) {
	// Save and restore original flag value
	origForce := deleteForce
	defer func() {
		deleteForce = origForce
	}()

	deleteForce = true

	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	err := runDelete(deleteCmd, []string{"fav-123"})
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

	// Save and restore original flag value
	origForce := deleteForce
	defer func() {
		deleteForce = origForce
	}()

	deleteForce = true

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	err := runDelete(deleteCmd, []string{"nonexistent-fav"})
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestRunDelete_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	// Save and restore original flag value
	origForce := deleteForce
	defer func() {
		deleteForce = origForce
	}()

	deleteForce = true

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	err := runDelete(deleteCmd, []string{"fav-123"})
	if err == nil {
		t.Fatal("expected error for server error response")
	}
}

func TestRunDelete_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer server.Close()

	// Save and restore original flag value
	origForce := deleteForce
	defer func() {
		deleteForce = origForce
	}()

	deleteForce = true

	t.Setenv("TWENTY_TOKEN", "invalid-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	err := runDelete(deleteCmd, []string{"fav-123"})
	if err == nil {
		t.Fatal("expected error for unauthorized response")
	}
}

func TestRunDelete_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "forbidden"}`))
	}))
	defer server.Close()

	// Save and restore original flag value
	origForce := deleteForce
	defer func() {
		deleteForce = origForce
	}()

	deleteForce = true

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	err := runDelete(deleteCmd, []string{"fav-123"})
	if err == nil {
		t.Fatal("expected error for forbidden response")
	}
}

func TestRunDelete_DebugMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "fav-debug"}`))
	}))
	defer server.Close()

	// Save and restore original flag value
	origForce := deleteForce
	defer func() {
		deleteForce = origForce
	}()

	deleteForce = true

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", true) // Enable debug mode
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runDelete(deleteCmd, []string{"fav-debug"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runDelete() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Deleted favorite") {
		t.Errorf("output missing confirmation: %s", output)
	}
}

func TestRunDelete_EmptyID(t *testing.T) {
	// Save and restore original flag value
	origForce := deleteForce
	defer func() {
		deleteForce = origForce
	}()

	deleteForce = false

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Even with empty ID, command should handle gracefully
	err := runDelete(deleteCmd, []string{""})
	w.Close()
	os.Stdout = oldStdout

	// Should still show prompt since force is false
	if err != nil {
		t.Fatalf("runDelete() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "About to delete") {
		t.Errorf("output missing confirmation prompt: %s", output)
	}
}

func TestRunDelete_LongID(t *testing.T) {
	longID := "very-long-favorite-id-12345678901234567890"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/rest/favorites/"+longID) {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "` + longID + `"}`))
	}))
	defer server.Close()

	// Save and restore original flag value
	origForce := deleteForce
	defer func() {
		deleteForce = origForce
	}()

	deleteForce = true

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runDelete(deleteCmd, []string{longID})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runDelete() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, longID) {
		t.Errorf("output missing favorite ID: %s", output)
	}
}
