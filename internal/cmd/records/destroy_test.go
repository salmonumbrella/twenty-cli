package records

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/spf13/viper"
)

func TestDestroyCmd_Flags(t *testing.T) {
	if destroyCmd.Flags().Lookup("force") == nil {
		t.Error("force flag not registered")
	}
}

func TestDestroyCmd_Use(t *testing.T) {
	if destroyCmd.Use != "destroy <object> <id>" {
		t.Errorf("Use = %q, want %q", destroyCmd.Use, "destroy <object> <id>")
	}
}

func TestDestroyCmd_Short(t *testing.T) {
	if destroyCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestDestroyCmd_Args(t *testing.T) {
	if destroyCmd.Args == nil {
		t.Error("Args validator not set")
	}

	// Test with zero args
	err := destroyCmd.Args(destroyCmd, []string{})
	if err == nil {
		t.Error("Expected error with zero args")
	}

	// Test with one arg
	err = destroyCmd.Args(destroyCmd, []string{"people"})
	if err == nil {
		t.Error("Expected error with one arg")
	}

	// Test with two args
	err = destroyCmd.Args(destroyCmd, []string{"people", "id-123"})
	if err != nil {
		t.Errorf("Expected no error with two args, got %v", err)
	}
}

func TestRunDestroy_WithoutForce(t *testing.T) {
	originalForce := destroyForce
	defer func() { destroyForce = originalForce }()

	destroyForce = false

	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be called without --force")
	}))
	defer cleanup()

	noResolve = true

	err := runDestroy(destroyCmd, []string{"people", "p1"})
	if err != nil {
		t.Errorf("Expected no error when --force is not set, got %v", err)
	}
}

func TestRunDestroy_WithForce_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		if r.URL.Path != "/rest/people/p1/destroy" {
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

	originalForce := destroyForce
	defer func() { destroyForce = originalForce }()

	destroyForce = true
	noResolve = true

	err := runDestroy(destroyCmd, []string{"people", "p1"})
	if err != nil {
		t.Fatalf("runDestroy failed: %v", err)
	}
}

func TestRunDestroy_EmptyResponse(t *testing.T) {
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
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	originalForce := destroyForce
	defer func() { destroyForce = originalForce }()

	destroyForce = true
	noResolve = true

	err := runDestroy(destroyCmd, []string{"people", "p1"})
	if err != nil {
		t.Fatalf("runDestroy failed: %v", err)
	}
}

func TestRunDestroy_APIError(t *testing.T) {
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

	originalForce := destroyForce
	defer func() { destroyForce = originalForce }()

	destroyForce = true
	noResolve = true

	err := runDestroy(destroyCmd, []string{"people", "nonexistent"})
	if err == nil {
		t.Error("expected error for API error")
	}
}

func TestRunDestroy_TextOutput(t *testing.T) {
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

	originalForce := destroyForce
	defer func() { destroyForce = originalForce }()

	destroyForce = true
	noResolve = true

	err := runDestroy(destroyCmd, []string{"people", "p1"})
	if err != nil {
		t.Fatalf("runDestroy failed: %v", err)
	}
}
