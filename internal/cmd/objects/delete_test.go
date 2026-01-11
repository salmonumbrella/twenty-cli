package objects

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
	if deleteCmd.Use != "delete <object-id>" {
		t.Errorf("Use = %q, want %q", deleteCmd.Use, "delete <object-id>")
	}
}

func TestDeleteCmd_Short(t *testing.T) {
	if deleteCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestDeleteCmd_HasRunE(t *testing.T) {
	if deleteCmd.RunE == nil {
		t.Error("deleteCmd should have RunE set")
	}
}

func TestDeleteCmd_Args(t *testing.T) {
	if deleteCmd.Args == nil {
		t.Error("deleteCmd should have Args validation set")
	}
}

func TestDeleteCmd_ForceFlag(t *testing.T) {
	flag := deleteCmd.Flags().Lookup("force")
	if flag == nil {
		t.Fatal("force flag not registered")
	}
}

func TestRunDelete_WithoutForce(t *testing.T) {
	// Save original value and restore after test
	oldDeleteForce := deleteForce
	defer func() {
		deleteForce = oldDeleteForce
	}()

	deleteForce = false

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runDelete(nil, []string{"obj-123"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runDelete() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should show warning message about using --force
	if !strings.Contains(output, "About to delete object obj-123") {
		t.Errorf("output missing deletion warning: %s", output)
	}
	if !strings.Contains(output, "--force") {
		t.Errorf("output missing --force instruction: %s", output)
	}
}

func TestRunDelete_WithForce_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/rest/metadata/objects/obj-to-delete" {
			t.Errorf("expected path /rest/metadata/objects/obj-to-delete, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	// Save original value and restore after test
	oldDeleteForce := deleteForce
	defer func() {
		deleteForce = oldDeleteForce
	}()

	deleteForce = true

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runDelete(nil, []string{"obj-to-delete"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runDelete() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Deleted object obj-to-delete") {
		t.Errorf("output missing confirmation: %s", output)
	}
}

func TestRunDelete_WithForce_NoToken(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	// Save original value and restore after test
	oldDeleteForce := deleteForce
	defer func() {
		deleteForce = oldDeleteForce
	}()

	deleteForce = true

	err := runDelete(nil, []string{"obj-123"})
	if err == nil {
		t.Fatal("expected error when no token is available")
	}
}

func TestRunDelete_WithForce_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "object not found"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	// Save original value and restore after test
	oldDeleteForce := deleteForce
	defer func() {
		deleteForce = oldDeleteForce
	}()

	deleteForce = true

	err := runDelete(nil, []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestRunDelete_WithForce_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	// Save original value and restore after test
	oldDeleteForce := deleteForce
	defer func() {
		deleteForce = oldDeleteForce
	}()

	deleteForce = true

	err := runDelete(nil, []string{"obj-123"})
	if err == nil {
		t.Fatal("expected error for unauthorized response")
	}
}

func TestRunDelete_DifferentObjectIDs(t *testing.T) {
	tests := []struct {
		objectID string
	}{
		{"simple-id"},
		{"12345678-1234-1234-1234-123456789012"},
		{"custom_object"},
		{"obj-with-dashes-123"},
	}

	for _, tt := range tests {
		t.Run(tt.objectID, func(t *testing.T) {
			var receivedPath string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedPath = r.URL.Path
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			t.Setenv("TWENTY_TOKEN", "test-token")
			viper.Set("base_url", server.URL)
			viper.Set("debug", false)
			t.Cleanup(viper.Reset)

			// Save original value and restore after test
			oldDeleteForce := deleteForce
			defer func() {
				deleteForce = oldDeleteForce
			}()

			deleteForce = true

			// Capture stdout
			oldStdout := os.Stdout
			_, w, _ := os.Pipe()
			os.Stdout = w

			err := runDelete(nil, []string{tt.objectID})
			w.Close()
			os.Stdout = oldStdout

			if err != nil {
				t.Fatalf("runDelete() error = %v", err)
			}

			expectedPath := "/rest/metadata/objects/" + tt.objectID
			if receivedPath != expectedPath {
				t.Errorf("expected path %s, got %s", expectedPath, receivedPath)
			}
		})
	}
}
