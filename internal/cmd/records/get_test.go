package records

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/spf13/viper"
)

func TestGetCmd_Flags(t *testing.T) {
	flags := []string{"fields", "include", "param"}
	for _, flag := range flags {
		if getCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestGetCmd_Use(t *testing.T) {
	if getCmd.Use != "get <object> <id>" {
		t.Errorf("Use = %q, want %q", getCmd.Use, "get <object> <id>")
	}
}

func TestGetCmd_Short(t *testing.T) {
	if getCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestGetCmd_Args(t *testing.T) {
	if getCmd.Args == nil {
		t.Error("Args validator not set")
	}

	// Test with zero args
	err := getCmd.Args(getCmd, []string{})
	if err == nil {
		t.Error("Expected error with zero args")
	}

	// Test with one arg
	err = getCmd.Args(getCmd, []string{"people"})
	if err == nil {
		t.Error("Expected error with one arg")
	}

	// Test with two args
	err = getCmd.Args(getCmd, []string{"people", "id-123"})
	if err != nil {
		t.Errorf("Expected no error with two args, got %v", err)
	}

	// Test with three args
	err = getCmd.Args(getCmd, []string{"people", "id-123", "extra"})
	if err == nil {
		t.Error("Expected error with three args")
	}
}

func TestRunGet_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		if r.URL.Path != "/rest/people/person-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"person":{"id":"person-123","name":"John Doe"}}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Reset flags
	getFields = ""
	getInclude = ""
	getParams = nil
	noResolve = true

	err := runGet(getCmd, []string{"people", "person-123"})
	if err != nil {
		t.Fatalf("runGet failed: %v", err)
	}
}

func TestRunGet_WithFields(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		fields := r.URL.Query().Get("fields")
		if fields != "id,name,email" {
			t.Errorf("fields = %q, want %q", fields, "id,name,email")
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"person":{"id":"p1"}}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	getFields = "id,name,email"
	getInclude = ""
	getParams = nil
	noResolve = true

	err := runGet(getCmd, []string{"people", "p1"})
	if err != nil {
		t.Fatalf("runGet failed: %v", err)
	}

	getFields = ""
}

func TestRunGet_WithInclude(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		depth := r.URL.Query().Get("depth")
		if depth != "1" {
			t.Errorf("depth = %q, want %q", depth, "1")
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"person":{"id":"p1","company":{"id":"c1"}}}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	getFields = ""
	getInclude = "company"
	getParams = nil
	noResolve = true

	err := runGet(getCmd, []string{"people", "p1"})
	if err != nil {
		t.Fatalf("runGet failed: %v", err)
	}

	getInclude = ""
}

func TestRunGet_NotFound(t *testing.T) {
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
	getFields = ""
	getInclude = ""
	getParams = nil

	err := runGet(getCmd, []string{"people", "nonexistent"})
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestRunGet_TextOutput(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"person":{"id":"p1","name":"Test"}}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	viper.Set("output", "text")
	noResolve = true
	getFields = ""
	getInclude = ""
	getParams = nil

	err := runGet(getCmd, []string{"people", "p1"})
	if err != nil {
		t.Fatalf("runGet failed: %v", err)
	}
}

func TestRunGet_WithCustomParams(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		custom := r.URL.Query().Get("custom")
		if custom != "value" {
			t.Errorf("custom = %q, want %q", custom, "value")
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"item":{"id":"1"}}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	getFields = ""
	getInclude = ""
	getParams = []string{"custom=value"}
	noResolve = true

	err := runGet(getCmd, []string{"items", "1"})
	if err != nil {
		t.Fatalf("runGet failed: %v", err)
	}

	getParams = nil
}
