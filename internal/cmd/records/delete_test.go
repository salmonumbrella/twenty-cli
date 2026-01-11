package records

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/spf13/viper"
)

func TestDeleteCmd_Flags(t *testing.T) {
	if deleteCmd.Flags().Lookup("force") == nil {
		t.Error("force flag not registered")
	}
}

func TestDeleteCmd_Use(t *testing.T) {
	if deleteCmd.Use != "delete <object> <id>" {
		t.Errorf("Use = %q, want %q", deleteCmd.Use, "delete <object> <id>")
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
	err = deleteCmd.Args(deleteCmd, []string{"people"})
	if err == nil {
		t.Error("Expected error with one arg")
	}

	// Test with two args
	err = deleteCmd.Args(deleteCmd, []string{"people", "id-123"})
	if err != nil {
		t.Errorf("Expected no error with two args, got %v", err)
	}

	// Test with three args
	err = deleteCmd.Args(deleteCmd, []string{"people", "id-123", "extra"})
	if err == nil {
		t.Error("Expected error with three args")
	}
}

func TestRunDelete_WithoutForce(t *testing.T) {
	// Save original value
	originalForce := deleteForce
	defer func() { deleteForce = originalForce }()

	deleteForce = false

	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be called without --force")
	}))
	defer cleanup()

	noResolve = true

	err := runDelete(deleteCmd, []string{"people", "p1"})
	if err != nil {
		t.Errorf("Expected no error when --force is not set, got %v", err)
	}
}

func TestRunDelete_WithForce_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		if r.URL.Path != "/rest/people/p1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"person":{"id":"p1"}}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Save and restore deleteForce flag
	originalForce := deleteForce
	defer func() { deleteForce = originalForce }()

	deleteForce = true
	noResolve = true

	err := runDelete(deleteCmd, []string{"people", "p1"})
	if err != nil {
		t.Fatalf("runDelete failed: %v", err)
	}
}

func TestRunDelete_EmptyResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		// Empty response body
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	originalForce := deleteForce
	defer func() { deleteForce = originalForce }()

	deleteForce = true
	noResolve = true

	err := runDelete(deleteCmd, []string{"people", "p1"})
	if err != nil {
		t.Fatalf("runDelete failed: %v", err)
	}
}

func TestRunDelete_APIError(t *testing.T) {
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

	originalForce := deleteForce
	defer func() { deleteForce = originalForce }()

	deleteForce = true
	noResolve = true

	err := runDelete(deleteCmd, []string{"people", "nonexistent"})
	if err == nil {
		t.Error("expected error for API error")
	}
}

func TestRunDelete_TextOutput(t *testing.T) {
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

	originalForce := deleteForce
	defer func() { deleteForce = originalForce }()

	deleteForce = true
	noResolve = true

	err := runDelete(deleteCmd, []string{"people", "p1"})
	if err != nil {
		t.Fatalf("runDelete failed: %v", err)
	}
}
