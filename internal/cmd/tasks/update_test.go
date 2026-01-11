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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestUpdateCmd_Flags(t *testing.T) {
	flags := []string{"title", "due-at", "status", "assignee-id", "data"}
	for _, flag := range flags {
		if updateCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestUpdateCmd_Use(t *testing.T) {
	if updateCmd.Use != "update <id>" {
		t.Errorf("Use = %q, want %q", updateCmd.Use, "update <id>")
	}
}

func TestUpdateCmd_Short(t *testing.T) {
	if updateCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestUpdateCmd_Args(t *testing.T) {
	// Test that command requires exactly one argument
	err := updateCmd.Args(updateCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = updateCmd.Args(updateCmd, []string{"id1", "id2"})
	if err == nil {
		t.Error("expected error when too many args provided")
	}

	err = updateCmd.Args(updateCmd, []string{"id1"})
	if err != nil {
		t.Errorf("unexpected error for valid args: %v", err)
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

func TestRunUpdate_Success(t *testing.T) {
	createdAt := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/rest/tasks/task-to-update") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify request body
		var input map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		if input["title"] != "Updated Title" {
			t.Errorf("title = %q, want %q", input["title"], "Updated Title")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"updateTask": map[string]interface{}{
					"id":        "task-to-update",
					"title":     "Updated Title",
					"status":    "IN_PROGRESS",
					"createdAt": createdAt.Format(time.RFC3339),
					"updatedAt": time.Now().Format(time.RFC3339),
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create a fresh command for this test to properly track flag changes
	testCmd := &cobra.Command{Use: "update"}
	testCmd.Flags().StringVar(&updateTitle, "title", "", "task title")
	testCmd.Flags().StringVar(&updateDueAt, "due-at", "", "due date")
	testCmd.Flags().StringVar(&updateStatus, "status", "", "task status")
	testCmd.Flags().StringVar(&updateAssigneeID, "assignee-id", "", "assignee ID")
	testCmd.Flags().StringVarP(&updateData, "data", "d", "", "JSON data")

	// Reset and set variables
	updateTitle = "Updated Title"
	updateDueAt = ""
	updateStatus = ""
	updateAssigneeID = ""
	updateData = ""

	// Mark the title flag as changed
	testCmd.Flags().Set("title", "Updated Title")

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runUpdate(testCmd, []string{"task-to-update"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runUpdate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "task-to-update") {
		t.Errorf("output missing task ID: %s", output)
	}
	if !strings.Contains(output, "Updated Title") {
		t.Errorf("output missing updated title: %s", output)
	}
}

func TestRunUpdate_WithJSONData(t *testing.T) {
	createdAt := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}

		var input map[string]interface{}
		json.NewDecoder(r.Body).Decode(&input)

		if input["title"] != "JSON Updated Title" {
			t.Errorf("title = %q, want %q", input["title"], "JSON Updated Title")
		}
		if input["status"] != "DONE" {
			t.Errorf("status = %q, want %q", input["status"], "DONE")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"updateTask": map[string]interface{}{
					"id":        "task-json-update",
					"title":     "JSON Updated Title",
					"status":    "DONE",
					"createdAt": createdAt.Format(time.RFC3339),
					"updatedAt": time.Now().Format(time.RFC3339),
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	updateTitle = ""
	updateDueAt = ""
	updateStatus = ""
	updateAssigneeID = ""
	updateData = `{"title": "JSON Updated Title", "status": "DONE"}`

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runUpdate(updateCmd, []string{"task-json-update"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runUpdate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "task-json-update") {
		t.Errorf("output missing task ID: %s", output)
	}
}

func TestRunUpdate_InvalidJSONData(t *testing.T) {
	updateTitle = ""
	updateDueAt = ""
	updateStatus = ""
	updateAssigneeID = ""
	updateData = `{invalid json}`

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	err := runUpdate(updateCmd, []string{"task-123"})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid JSON data") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRunUpdate_NoToken(t *testing.T) {
	updateTitle = "Test"
	updateDueAt = ""
	updateStatus = ""
	updateAssigneeID = ""
	updateData = ""

	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	err := runUpdate(updateCmd, []string{"task-123"})
	if err == nil {
		t.Fatal("expected error when no token is available")
	}
}

func TestRunUpdate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"error": "validation error"}`))
	}))
	defer server.Close()

	updateTitle = ""
	updateDueAt = ""
	updateStatus = ""
	updateAssigneeID = ""
	updateData = `{"title": "Test"}`

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	err := runUpdate(updateCmd, []string{"task-123"})
	if err == nil {
		t.Fatal("expected error for API error response")
	}
	if !strings.Contains(err.Error(), "failed to update task") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRunUpdate_JSONOutput(t *testing.T) {
	createdAt := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"updateTask": map[string]interface{}{
					"id":        "task-json-output",
					"title":     "JSON Output Title",
					"createdAt": createdAt.Format(time.RFC3339),
					"updatedAt": time.Now().Format(time.RFC3339),
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	updateTitle = ""
	updateDueAt = ""
	updateStatus = ""
	updateAssigneeID = ""
	updateData = `{"title": "JSON Output Title"}`

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runUpdate(updateCmd, []string{"task-json-output"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runUpdate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify JSON output
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("output is not valid JSON: %v", err)
	}
	if result["id"] != "task-json-output" {
		t.Errorf("JSON output wrong 'id': got %v", result["id"])
	}
}

func TestRunUpdate_AllFlags(t *testing.T) {
	createdAt := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input map[string]interface{}
		json.NewDecoder(r.Body).Decode(&input)

		// Verify all fields are sent
		if input["title"] != "Full Update" {
			t.Errorf("title = %q, want %q", input["title"], "Full Update")
		}
		if input["dueAt"] != "2025-06-30T12:00:00Z" {
			t.Errorf("dueAt = %q, want %q", input["dueAt"], "2025-06-30T12:00:00Z")
		}
		if input["status"] != "IN_PROGRESS" {
			t.Errorf("status = %q, want %q", input["status"], "IN_PROGRESS")
		}
		if input["assigneeId"] != "new-user-id" {
			t.Errorf("assigneeId = %q, want %q", input["assigneeId"], "new-user-id")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"updateTask": map[string]interface{}{
					"id":         "task-full-update",
					"title":      "Full Update",
					"dueAt":      "2025-06-30T12:00:00Z",
					"status":     "IN_PROGRESS",
					"assigneeId": "new-user-id",
					"createdAt":  createdAt.Format(time.RFC3339),
					"updatedAt":  time.Now().Format(time.RFC3339),
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create a fresh command for this test
	testCmd := &cobra.Command{Use: "update"}
	testCmd.Flags().StringVar(&updateTitle, "title", "", "task title")
	testCmd.Flags().StringVar(&updateDueAt, "due-at", "", "due date")
	testCmd.Flags().StringVar(&updateStatus, "status", "", "task status")
	testCmd.Flags().StringVar(&updateAssigneeID, "assignee-id", "", "assignee ID")
	testCmd.Flags().StringVarP(&updateData, "data", "d", "", "JSON data")

	updateData = ""

	// Set all flags
	testCmd.Flags().Set("title", "Full Update")
	testCmd.Flags().Set("due-at", "2025-06-30T12:00:00Z")
	testCmd.Flags().Set("status", "IN_PROGRESS")
	testCmd.Flags().Set("assignee-id", "new-user-id")

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runUpdate(testCmd, []string{"task-full-update"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runUpdate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "task-full-update") {
		t.Errorf("output missing task ID: %s", output)
	}
}

func TestRunUpdate_PartialFlags(t *testing.T) {
	createdAt := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input map[string]interface{}
		json.NewDecoder(r.Body).Decode(&input)

		// Only status should be set
		if _, exists := input["title"]; exists {
			t.Error("title should not be in request")
		}
		if input["status"] != "DONE" {
			t.Errorf("status = %q, want %q", input["status"], "DONE")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"updateTask": map[string]interface{}{
					"id":        "task-partial",
					"title":     "Original Title",
					"status":    "DONE",
					"createdAt": createdAt.Format(time.RFC3339),
					"updatedAt": time.Now().Format(time.RFC3339),
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create a fresh command for this test
	testCmd := &cobra.Command{Use: "update"}
	testCmd.Flags().StringVar(&updateTitle, "title", "", "task title")
	testCmd.Flags().StringVar(&updateDueAt, "due-at", "", "due date")
	testCmd.Flags().StringVar(&updateStatus, "status", "", "task status")
	testCmd.Flags().StringVar(&updateAssigneeID, "assignee-id", "", "assignee ID")
	testCmd.Flags().StringVarP(&updateData, "data", "d", "", "JSON data")

	updateData = ""
	updateTitle = ""

	// Only set status flag
	testCmd.Flags().Set("status", "DONE")

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runUpdate(testCmd, []string{"task-partial"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runUpdate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "task-partial") {
		t.Errorf("output missing task ID: %s", output)
	}
}
