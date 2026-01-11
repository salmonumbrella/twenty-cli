package records

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/spf13/viper"
)

func TestCreateCmd_Flags(t *testing.T) {
	flags := []string{"data", "file", "set"}
	for _, flag := range flags {
		if createCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestCreateCmd_Use(t *testing.T) {
	if createCmd.Use != "create <object>" {
		t.Errorf("Use = %q, want %q", createCmd.Use, "create <object>")
	}
}

func TestCreateCmd_Short(t *testing.T) {
	if createCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestCreateCmd_Args(t *testing.T) {
	if createCmd.Args == nil {
		t.Error("Args validator not set")
	}

	// Test with zero args
	err := createCmd.Args(createCmd, []string{})
	if err == nil {
		t.Error("Expected error with zero args")
	}

	// Test with one arg
	err = createCmd.Args(createCmd, []string{"people"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}

	// Test with two args
	err = createCmd.Args(createCmd, []string{"people", "extra"})
	if err == nil {
		t.Error("Expected error with two args")
	}
}

func TestCreateCmd_DataFlagShorthand(t *testing.T) {
	flag := createCmd.Flags().Lookup("data")
	if flag == nil {
		t.Fatal("data flag not registered")
	}
	if flag.Shorthand != "d" {
		t.Errorf("data flag shorthand = %q, want %q", flag.Shorthand, "d")
	}
}

func TestCreateCmd_FileFlagShorthand(t *testing.T) {
	flag := createCmd.Flags().Lookup("file")
	if flag == nil {
		t.Fatal("file flag not registered")
	}
	if flag.Shorthand != "f" {
		t.Errorf("file flag shorthand = %q, want %q", flag.Shorthand, "f")
	}
}

func TestRunCreate_WithData(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		if r.URL.Path != "/rest/people" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		json.Unmarshal(body, &payload)

		if payload["name"] != "John Doe" {
			t.Errorf("name = %v, want %q", payload["name"], "John Doe")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"data":{"person":{"id":"new-id","name":"John Doe"}}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	createData = `{"name":"John Doe"}`
	createDataFile = ""
	createSet = nil
	noResolve = true

	err := runCreate(createCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runCreate failed: %v", err)
	}

	createData = ""
}

func TestRunCreate_WithSet(t *testing.T) {
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

		if payload["name"] != "Jane" {
			t.Errorf("name = %v, want %q", payload["name"], "Jane")
		}
		if payload["active"] != true {
			t.Errorf("active = %v, want %v", payload["active"], true)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"data":{"person":{"id":"new-id"}}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	createData = ""
	createDataFile = ""
	createSet = []string{"name=Jane", "active=true"}
	noResolve = true

	err := runCreate(createCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runCreate failed: %v", err)
	}

	createSet = nil
}

func TestRunCreate_WithNestedSet(t *testing.T) {
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

		person, ok := payload["person"].(map[string]interface{})
		if !ok {
			t.Error("person should be an object")
			return
		}
		if person["firstName"] != "John" {
			t.Errorf("person.firstName = %v, want %q", person["firstName"], "John")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"data":{"person":{"id":"new-id"}}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	createData = ""
	createDataFile = ""
	createSet = []string{"person.firstName=John"}
	noResolve = true

	err := runCreate(createCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runCreate failed: %v", err)
	}

	createSet = nil
}

func TestRunCreate_NoPayload(t *testing.T) {
	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer cleanup()

	createData = ""
	createDataFile = ""
	createSet = nil
	noResolve = true

	err := runCreate(createCmd, []string{"people"})
	if err == nil {
		t.Error("expected error when no payload provided")
	}
}

func TestRunCreate_InvalidJSON(t *testing.T) {
	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer cleanup()

	createData = "{invalid"
	createDataFile = ""
	createSet = nil
	noResolve = true

	err := runCreate(createCmd, []string{"people"})
	if err == nil {
		t.Error("expected error for invalid JSON")
	}

	createData = ""
}

func TestRunCreate_APIError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"message":"Invalid input"}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	createData = `{"name":"test"}`
	createDataFile = ""
	createSet = nil
	noResolve = true

	err := runCreate(createCmd, []string{"people"})
	if err == nil {
		t.Error("expected error for API error")
	}

	createData = ""
}

func TestRunCreate_TextOutput(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"data":{"person":{"id":"new-id","name":"Test"}}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	viper.Set("output", "text")
	createData = `{"name":"Test"}`
	createDataFile = ""
	createSet = nil
	noResolve = true

	err := runCreate(createCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runCreate failed: %v", err)
	}

	createData = ""
}

func TestRunCreate_DataWithSetOverride(t *testing.T) {
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
		if payload["name"] != "Override" {
			t.Errorf("name = %v, want %q (set should override data)", payload["name"], "Override")
		}
		if payload["extra"] != "value" {
			t.Errorf("extra = %v, want %q", payload["extra"], "value")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"data":{"person":{"id":"new-id"}}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	createData = `{"name":"Original","extra":"value"}`
	createDataFile = ""
	createSet = []string{"name=Override"}
	noResolve = true

	err := runCreate(createCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runCreate failed: %v", err)
	}

	createData = ""
	createSet = nil
}
