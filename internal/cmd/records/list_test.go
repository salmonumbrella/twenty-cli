package records

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/viper"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/auth"
	"github.com/salmonumbrella/twenty-cli/internal/secrets"
)

// setupTestEnv sets up a mock server and configures auth/viper for testing
func setupTestEnv(t *testing.T, handler http.Handler) func() {
	t.Helper()

	// Create mock server
	server := httptest.NewServer(handler)

	// Save original values
	originalEnv := os.Getenv("TWENTY_TOKEN")
	originalBaseURL := viper.GetString("base_url")
	originalDebug := viper.GetBool("debug")
	originalOutput := viper.GetString("output")
	originalProfile := viper.GetString("profile")
	originalQuery := viper.GetString("query")

	// Set up mock store with token
	mockStore := secrets.NewMockStore()
	_ = mockStore.SetToken("default", secrets.Token{
		Profile:      "default",
		RefreshToken: "test-token",
	})
	auth.SetStore(mockStore)

	// Clear environment variable to use mock store
	os.Unsetenv("TWENTY_TOKEN")

	// Configure viper
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("profile", "")
	viper.Set("query", "")

	// Return cleanup function
	return func() {
		server.Close()
		if originalEnv != "" {
			os.Setenv("TWENTY_TOKEN", originalEnv)
		} else {
			os.Unsetenv("TWENTY_TOKEN")
		}
		viper.Set("base_url", originalBaseURL)
		viper.Set("debug", originalDebug)
		viper.Set("output", originalOutput)
		viper.Set("profile", originalProfile)
		viper.Set("query", originalQuery)
	}
}

func TestListCmd_Flags(t *testing.T) {
	flags := []string{"limit", "cursor", "all", "filter", "filter-file", "sort", "order", "fields", "include", "param"}
	for _, flag := range flags {
		if listCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestListCmd_Use(t *testing.T) {
	if listCmd.Use != "list <object>" {
		t.Errorf("Use = %q, want %q", listCmd.Use, "list <object>")
	}
}

func TestListCmd_Short(t *testing.T) {
	if listCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestListCmd_Args(t *testing.T) {
	if listCmd.Args == nil {
		t.Error("Args validator not set")
	}

	// Test with zero args
	err := listCmd.Args(listCmd, []string{})
	if err == nil {
		t.Error("Expected error with zero args")
	}

	// Test with one arg
	err = listCmd.Args(listCmd, []string{"people"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}

	// Test with two args
	err = listCmd.Args(listCmd, []string{"arg1", "arg2"})
	if err == nil {
		t.Error("Expected error with two args")
	}
}

func TestListCmd_LimitFlagDefaults(t *testing.T) {
	flag := listCmd.Flags().Lookup("limit")
	if flag == nil {
		t.Fatal("limit flag not registered")
	}
	if flag.Shorthand != "l" {
		t.Errorf("limit flag shorthand = %q, want %q", flag.Shorthand, "l")
	}
	if flag.DefValue != "20" {
		t.Errorf("limit flag default = %q, want %q", flag.DefValue, "20")
	}
}

func TestRunList_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for object list call first
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"objects": []interface{}{},
				},
			})
			return
		}

		if r.URL.Path != "/rest/people" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"people":[{"id":"p1","name":"John"}]},"totalCount":1,"pageInfo":{"hasNextPage":false}}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Reset flags
	listLimit = 20
	listCursor = ""
	listAll = false
	listFilter = ""
	listFilterFile = ""
	listSort = ""
	listOrder = ""
	listFields = ""
	listInclude = ""
	listParams = nil
	noResolve = true

	err := runList(listCmd, []string{"people"})
	if err != nil {
		t.Fatalf("runList failed: %v", err)
	}
}

func TestRunList_WithPagination(t *testing.T) {
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		callCount++
		w.Header().Set("Content-Type", "application/json")

		if callCount == 1 {
			w.Write([]byte(`{"data":{"items":[{"id":"1"}]},"pageInfo":{"hasNextPage":true,"endCursor":"cursor1"}}`))
		} else {
			w.Write([]byte(`{"data":{"items":[{"id":"2"}]},"pageInfo":{"hasNextPage":false}}`))
		}
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Reset flags
	listLimit = 20
	listCursor = ""
	listAll = true
	listFilter = ""
	listFilterFile = ""
	listSort = ""
	listOrder = ""
	listFields = ""
	listInclude = ""
	listParams = nil
	noResolve = true

	err := runList(listCmd, []string{"items"})
	if err != nil {
		t.Fatalf("runList failed: %v", err)
	}

	if callCount != 2 {
		t.Errorf("expected 2 API calls for pagination, got %d", callCount)
	}
}

func TestRunList_WithFilter(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		filter := r.URL.Query().Get("filter")
		if filter == "" {
			t.Error("expected filter query param")
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"items":[]},"totalCount":0}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	// Reset flags
	listLimit = 20
	listCursor = ""
	listAll = false
	listFilter = `{"name":{"eq":"test"}}`
	listFilterFile = ""
	listSort = ""
	listOrder = ""
	listFields = ""
	listInclude = ""
	listParams = nil
	noResolve = true

	err := runList(listCmd, []string{"items"})
	if err != nil {
		t.Fatalf("runList failed: %v", err)
	}
}

func TestRunList_APIError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"server error"}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	noResolve = true
	listAll = false

	err := runList(listCmd, []string{"items"})
	if err == nil {
		t.Error("expected error for API error response")
	}
}

func TestRunList_InvalidFilter(t *testing.T) {
	cleanup := setupTestEnv(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer cleanup()

	listFilter = "{invalid"
	listFilterFile = ""
	noResolve = true
	listAll = false

	err := runList(listCmd, []string{"items"})
	if err == nil {
		t.Error("expected error for invalid filter JSON")
	}

	listFilter = ""
}

func TestRunList_TextOutput(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"items":[{"id":"1","name":"Item 1"}]},"totalCount":1}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	viper.Set("output", "text")
	noResolve = true
	listAll = false
	listFilter = ""

	err := runList(listCmd, []string{"items"})
	if err != nil {
		t.Fatalf("runList failed: %v", err)
	}
}

func TestRunList_CSVOutput(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"items":[{"id":"1","name":"Item 1"}]},"totalCount":1}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	viper.Set("output", "csv")
	noResolve = true
	listAll = false
	listFilter = ""

	err := runList(listCmd, []string{"items"})
	if err != nil {
		t.Fatalf("runList failed: %v", err)
	}
}

func TestRunList_YAMLOutput(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/meta/objects" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"objects": []interface{}{}},
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"items":[{"id":"1","name":"Item 1"}]},"totalCount":1}`))
	})

	cleanup := setupTestEnv(t, handler)
	defer cleanup()

	viper.Set("output", "yaml")
	noResolve = true
	listAll = false
	listFilter = ""

	err := runList(listCmd, []string{"items"})
	if err != nil {
		t.Fatalf("runList failed: %v", err)
	}
}

func TestOutputRecords_JSON(t *testing.T) {
	data := map[string]interface{}{"id": "1", "name": "test"}
	err := outputRecords(data, "json", "")
	if err != nil {
		t.Fatalf("outputRecords failed: %v", err)
	}
}

func TestOutputRecords_YAML(t *testing.T) {
	data := map[string]interface{}{"id": "1", "name": "test"}
	err := outputRecords(data, "yaml", "")
	if err != nil {
		t.Fatalf("outputRecords failed: %v", err)
	}
}

func TestOutputRecords_CSV(t *testing.T) {
	data := map[string]interface{}{
		"data": map[string]interface{}{
			"items": []interface{}{
				map[string]interface{}{"id": "1", "name": "test"},
			},
		},
	}
	err := outputRecords(data, "csv", "")
	if err != nil {
		t.Fatalf("outputRecords failed: %v", err)
	}
}

func TestOutputRecords_Default(t *testing.T) {
	data := map[string]interface{}{
		"data": map[string]interface{}{
			"items": []interface{}{
				map[string]interface{}{"id": "1", "name": "test"},
			},
		},
	}
	err := outputRecords(data, "", "")
	if err != nil {
		t.Fatalf("outputRecords failed: %v", err)
	}
}
