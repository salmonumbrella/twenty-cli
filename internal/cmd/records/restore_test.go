package records

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/spf13/viper"
)

func TestRestoreCmd_Use(t *testing.T) {
	if restoreCmd.Use != "restore <object> <id>" {
		t.Errorf("Use = %q, want %q", restoreCmd.Use, "restore <object> <id>")
	}
}

func TestRestoreCmd_Short(t *testing.T) {
	if restoreCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestRestoreCmd_Args(t *testing.T) {
	if restoreCmd.Args == nil {
		t.Error("Args validator not set")
	}

	// Test with zero args
	err := restoreCmd.Args(restoreCmd, []string{})
	if err == nil {
		t.Error("Expected error with zero args")
	}

	// Test with one arg
	err = restoreCmd.Args(restoreCmd, []string{"people"})
	if err == nil {
		t.Error("Expected error with one arg")
	}

	// Test with two args
	err = restoreCmd.Args(restoreCmd, []string{"people", "id-123"})
	if err != nil {
		t.Errorf("Expected no error with two args, got %v", err)
	}
}

func TestRunRestore_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		if r.URL.Path != "/rest/people/p1/restore" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"person":{"id":"p1","deletedAt":null}}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	noResolve = true

	err := runRestore(restoreCmd, []string{"people", "p1"})
	if err != nil {
		t.Fatalf("runRestore failed: %v", err)
	}
}

func TestRunRestore_APIError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":{"message":"Not found"}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	noResolve = true

	err := runRestore(restoreCmd, []string{"people", "nonexistent"})
	if err == nil {
		t.Error("expected error for API error")
	}
}

func TestRunRestore_TextOutput(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"person":{"id":"p1"}}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	viper.Set("output", "text")
	noResolve = true

	err := runRestore(restoreCmd, []string{"people", "p1"})
	if err != nil {
		t.Fatalf("runRestore failed: %v", err)
	}
}
