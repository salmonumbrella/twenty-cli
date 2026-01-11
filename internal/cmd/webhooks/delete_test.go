package webhooks

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
	t.Cleanup(func() {
		forceDelete = false
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/rest/webhooks/webhook-to-delete") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
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

	err := runDelete(deleteCmd, []string{"webhook-to-delete"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runDelete() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Deleted webhook") {
		t.Errorf("output missing confirmation: %s", output)
	}
	if !strings.Contains(output, "webhook-to-delete") {
		t.Errorf("output missing webhook ID: %s", output)
	}
}

func TestRunDelete_NoForcePrompt(t *testing.T) {
	t.Cleanup(func() {
		forceDelete = false
	})

	// When force is false, should prompt for confirmation
	forceDelete = false

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runDelete(deleteCmd, []string{"webhook-123"})
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
	if !strings.Contains(output, "webhook-123") {
		t.Errorf("output missing webhook ID in prompt: %s", output)
	}
	if !strings.Contains(output, "--force") {
		t.Errorf("output missing --force hint: %s", output)
	}
}

func TestRunDelete_NoToken(t *testing.T) {
	t.Cleanup(func() {
		forceDelete = false
	})

	forceDelete = true

	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	err := runDelete(deleteCmd, []string{"webhook-123"})
	if err == nil {
		t.Fatal("expected error when no token is available")
	}
}

func TestRunDelete_APIError(t *testing.T) {
	t.Cleanup(func() {
		forceDelete = false
	})

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

	err := runDelete(deleteCmd, []string{"nonexistent-webhook"})
	if err == nil {
		t.Fatal("expected error for API error response")
	}
	if !strings.Contains(err.Error(), "failed to delete webhook") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRunDelete_ServerError(t *testing.T) {
	t.Cleanup(func() {
		forceDelete = false
	})

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

	err := runDelete(deleteCmd, []string{"webhook-123"})
	if err == nil {
		t.Fatal("expected error for server error response")
	}
}

func TestRunDelete_Unauthorized(t *testing.T) {
	t.Cleanup(func() {
		forceDelete = false
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer server.Close()

	forceDelete = true

	t.Setenv("TWENTY_TOKEN", "bad-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	err := runDelete(deleteCmd, []string{"webhook-123"})
	if err == nil {
		t.Fatal("expected error for unauthorized response")
	}
}

func TestRunDelete_ForbiddenError(t *testing.T) {
	t.Cleanup(func() {
		forceDelete = false
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "forbidden"}`))
	}))
	defer server.Close()

	forceDelete = true

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	err := runDelete(deleteCmd, []string{"protected-webhook"})
	if err == nil {
		t.Fatal("expected error for forbidden response")
	}
}

func TestRunDelete_VerifiesCorrectID(t *testing.T) {
	t.Cleanup(func() {
		forceDelete = false
	})

	var receivedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	forceDelete = true

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runDelete(deleteCmd, []string{"specific-webhook-id-12345"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runDelete() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)

	expectedPath := "/rest/webhooks/specific-webhook-id-12345"
	if receivedPath != expectedPath {
		t.Errorf("path = %q, want %q", receivedPath, expectedPath)
	}
}

func TestRunDelete_ConfirmationMessageFormat(t *testing.T) {
	t.Cleanup(func() {
		forceDelete = false
	})

	forceDelete = false

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runDelete(deleteCmd, []string{"test-webhook-456"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runDelete() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify the confirmation message contains the webhook ID
	if !strings.Contains(output, "test-webhook-456") {
		t.Errorf("confirmation message missing webhook ID: %s", output)
	}

	// Verify the confirmation message mentions how to confirm
	if !strings.Contains(output, "--force") {
		t.Errorf("confirmation message missing --force instruction: %s", output)
	}
}

func TestRunDelete_SuccessMessageFormat(t *testing.T) {
	t.Cleanup(func() {
		forceDelete = false
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	forceDelete = true

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runDelete(deleteCmd, []string{"deleted-webhook-789"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runDelete() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify the success message format
	if !strings.Contains(output, "Deleted webhook") {
		t.Errorf("success message missing 'Deleted webhook': %s", output)
	}
	if !strings.Contains(output, "deleted-webhook-789") {
		t.Errorf("success message missing webhook ID: %s", output)
	}
}

func TestRunDelete_NoForceDoesNotCallAPI(t *testing.T) {
	t.Cleanup(func() {
		forceDelete = false
	})

	apiCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	forceDelete = false

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runDelete(deleteCmd, []string{"webhook-no-force"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runDelete() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)

	if apiCalled {
		t.Error("API should not be called when --force is not specified")
	}
}

func TestDeleteCmd_ForceFlagDefaultValue(t *testing.T) {
	// The force flag should default to false
	flag := deleteCmd.Flags().Lookup("force")
	if flag == nil {
		t.Fatal("force flag not registered")
	}
	if flag.DefValue != "false" {
		t.Errorf("force flag default value = %q, want %q", flag.DefValue, "false")
	}
}
