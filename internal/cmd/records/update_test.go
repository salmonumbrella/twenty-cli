package records

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/spf13/viper"
)

func TestUpdateCmd_Flags(t *testing.T) {
	flags := []string{"data", "file", "set"}
	for _, flag := range flags {
		if updateCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestUpdateCmd_Use(t *testing.T) {
	if updateCmd.Use != "update <object> <id>" {
		t.Errorf("Use = %q, want %q", updateCmd.Use, "update <object> <id>")
	}
}

func TestUpdateCmd_Short(t *testing.T) {
	if updateCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestUpdateCmd_Args(t *testing.T) {
	if updateCmd.Args == nil {
		t.Error("Args validator not set")
	}

	// Test with zero args
	err := updateCmd.Args(updateCmd, []string{})
	if err == nil {
		t.Error("Expected error with zero args")
	}

	// Test with one arg
	err = updateCmd.Args(updateCmd, []string{"people"})
	if err == nil {
		t.Error("Expected error with one arg")
	}

	// Test with two args
	err = updateCmd.Args(updateCmd, []string{"people", "id-123"})
	if err != nil {
		t.Errorf("Expected no error with two args, got %v", err)
	}

	// Test with three args
	err = updateCmd.Args(updateCmd, []string{"people", "id-123", "extra"})
	if err == nil {
		t.Error("Expected error with three args")
	}
}

func TestUpdateCmd_DataFlagShorthand(t *testing.T) {
	flag := updateCmd.Flags().Lookup("data")
	if flag == nil {
		t.Fatal("data flag not registered")
	}
	if flag.Shorthand != "d" {
		t.Errorf("data flag shorthand = %q, want %q", flag.Shorthand, "d")
	}
}

func TestRunUpdate_WithData(t *testing.T) {
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

		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		json.Unmarshal(body, &payload)

		if payload["name"] != "Updated Name" {
			t.Errorf("name = %v, want %q", payload["name"], "Updated Name")
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"person":{"id":"person-123","name":"Updated Name"}}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	updateData = `{"name":"Updated Name"}`
	updateDataFile = ""
	updateSet = nil
	noResolve = true

	err := runUpdate(updateCmd, []string{"people", "person-123"})
	if err != nil {
		t.Fatalf("runUpdate failed: %v", err)
	}

	updateData = ""
}

func TestRunUpdate_WithSet(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		json.Unmarshal(body, &payload)

		if payload["status"] != "active" {
			t.Errorf("status = %v, want %q", payload["status"], "active")
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"person":{"id":"p1","status":"active"}}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	updateData = ""
	updateDataFile = ""
	updateSet = []string{"status=active"}
	noResolve = true

	err := runUpdate(updateCmd, []string{"people", "p1"})
	if err != nil {
		t.Fatalf("runUpdate failed: %v", err)
	}

	updateSet = nil
}

func TestRunUpdate_NoPayload(t *testing.T) {
	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer cleanup()

	updateData = ""
	updateDataFile = ""
	updateSet = nil
	noResolve = true

	err := runUpdate(updateCmd, []string{"people", "p1"})
	if err == nil {
		t.Error("expected error when no payload provided")
	}
}

func TestRunUpdate_InvalidJSON(t *testing.T) {
	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer cleanup()

	updateData = "{invalid"
	updateDataFile = ""
	updateSet = nil
	noResolve = true

	err := runUpdate(updateCmd, []string{"people", "p1"})
	if err == nil {
		t.Error("expected error for invalid JSON")
	}

	updateData = ""
}

func TestRunUpdate_APIError(t *testing.T) {
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

	updateData = `{"name":"test"}`
	updateDataFile = ""
	updateSet = nil
	noResolve = true

	err := runUpdate(updateCmd, []string{"people", "nonexistent"})
	if err == nil {
		t.Error("expected error for API error")
	}

	updateData = ""
}

func TestRunUpdate_TextOutput(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"person":{"id":"p1","name":"Updated"}}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	viper.Set("output", "text")
	updateData = `{"name":"Updated"}`
	updateDataFile = ""
	updateSet = nil
	noResolve = true

	err := runUpdate(updateCmd, []string{"people", "p1"})
	if err != nil {
		t.Fatalf("runUpdate failed: %v", err)
	}

	updateData = ""
}

func TestRunUpdate_DataWithSetOverride(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		json.Unmarshal(body, &payload)

		// Set should override data
		if payload["name"] != "Final" {
			t.Errorf("name = %v, want %q (set should override data)", payload["name"], "Final")
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"person":{"id":"p1"}}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	updateData = `{"name":"Original"}`
	updateDataFile = ""
	updateSet = []string{"name=Final"}
	noResolve = true

	err := runUpdate(updateCmd, []string{"people", "p1"})
	if err != nil {
		t.Fatalf("runUpdate failed: %v", err)
	}

	updateData = ""
	updateSet = nil
}
