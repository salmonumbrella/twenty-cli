package tasks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestCreateCmd_Flags(t *testing.T) {
	flags := []string{"title", "due-at", "status", "assignee-id", "data"}
	for _, flag := range flags {
		if createCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestCreateCmd_Use(t *testing.T) {
	if createCmd.Use != "create" {
		t.Errorf("Use = %q, want %q", createCmd.Use, "create")
	}
}

func TestCreateCmd_Short(t *testing.T) {
	if createCmd.Short == "" {
		t.Error("Short description should not be empty")
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

func TestRunCreate_Success(t *testing.T) {
	createdAt := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/rest/tasks") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify request body
		var input map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		if input["title"] != "Test Task" {
			t.Errorf("title = %q, want %q", input["title"], "Test Task")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"createTask": map[string]interface{}{
					"id":        "task-123",
					"title":     "Test Task",
					"status":    "TODO",
					"createdAt": createdAt.Format(time.RFC3339),
					"updatedAt": createdAt.Format(time.RFC3339),
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Reset package-level variables
	createTitle = "Test Task"
	createDueAt = ""
	createStatus = "TODO"
	createAssigneeID = ""
	createData = ""

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runCreate(createCmd, []string{})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runCreate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "task-123") {
		t.Errorf("output missing task ID: %s", output)
	}
	if !strings.Contains(output, "Test Task") {
		t.Errorf("output missing task title: %s", output)
	}
}

func TestRunCreate_SuccessJSON(t *testing.T) {
	createdAt := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"createTask": map[string]interface{}{
					"id":        "task-456",
					"title":     "JSON Task",
					"createdAt": createdAt.Format(time.RFC3339),
					"updatedAt": createdAt.Format(time.RFC3339),
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	createTitle = ""
	createDueAt = ""
	createStatus = ""
	createAssigneeID = ""
	createData = `{"title": "JSON Task"}`

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runCreate(createCmd, []string{})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runCreate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "task-456") {
		t.Errorf("output missing task ID: %s", output)
	}
}

func TestRunCreate_InvalidJSONData(t *testing.T) {
	createTitle = ""
	createDueAt = ""
	createStatus = ""
	createAssigneeID = ""
	createData = `{invalid json}`

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	err := runCreate(createCmd, []string{})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid JSON data") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRunCreate_NoToken(t *testing.T) {
	createTitle = "Test Task"
	createDueAt = ""
	createStatus = ""
	createAssigneeID = ""
	createData = ""

	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	err := runCreate(createCmd, []string{})
	if err == nil {
		t.Fatal("expected error when no token is available")
	}
}

func TestRunCreate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "bad request"}`))
	}))
	defer server.Close()

	createTitle = "Test Task"
	createDueAt = ""
	createStatus = ""
	createAssigneeID = ""
	createData = ""

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	err := runCreate(createCmd, []string{})
	if err == nil {
		t.Fatal("expected error for API error response")
	}
	if !strings.Contains(err.Error(), "failed to create task") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRunCreate_WithAllFlags(t *testing.T) {
	createdAt := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input map[string]interface{}
		json.NewDecoder(r.Body).Decode(&input)

		// Verify all fields are sent
		if input["title"] != "Full Task" {
			t.Errorf("title = %q, want %q", input["title"], "Full Task")
		}
		if input["dueAt"] != "2024-12-31T23:59:59Z" {
			t.Errorf("dueAt = %q, want %q", input["dueAt"], "2024-12-31T23:59:59Z")
		}
		if input["status"] != "IN_PROGRESS" {
			t.Errorf("status = %q, want %q", input["status"], "IN_PROGRESS")
		}
		if input["assigneeId"] != "user-123" {
			t.Errorf("assigneeId = %q, want %q", input["assigneeId"], "user-123")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"createTask": map[string]interface{}{
					"id":         "task-789",
					"title":      "Full Task",
					"dueAt":      "2024-12-31T23:59:59Z",
					"status":     "IN_PROGRESS",
					"assigneeId": "user-123",
					"createdAt":  createdAt.Format(time.RFC3339),
					"updatedAt":  createdAt.Format(time.RFC3339),
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	createTitle = "Full Task"
	createDueAt = "2024-12-31T23:59:59Z"
	createStatus = "IN_PROGRESS"
	createAssigneeID = "user-123"
	createData = ""

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runCreate(createCmd, []string{})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runCreate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "task-789") {
		t.Errorf("output missing task ID: %s", output)
	}
}
