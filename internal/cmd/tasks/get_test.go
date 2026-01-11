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

	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func TestGetCmd_Use(t *testing.T) {
	if getCmd.Use != "get <id>" {
		t.Errorf("Use = %q, want %q", getCmd.Use, "get <id>")
	}
}

func TestGetCmd_Short(t *testing.T) {
	if getCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestGetCmd_Args(t *testing.T) {
	// Test that command requires exactly one argument
	err := getCmd.Args(getCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = getCmd.Args(getCmd, []string{"id1", "id2"})
	if err == nil {
		t.Error("expected error when too many args provided")
	}

	err = getCmd.Args(getCmd, []string{"id1"})
	if err != nil {
		t.Errorf("unexpected error for valid args: %v", err)
	}
}

func TestRunGet_Success(t *testing.T) {
	createdAt := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/rest/tasks/task-123") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"task": map[string]interface{}{
					"id":         "task-123",
					"title":      "Test Task",
					"dueAt":      "2024-12-31T23:59:59Z",
					"status":     "TODO",
					"assigneeId": "user-456",
					"createdAt":  createdAt.Format(time.RFC3339),
					"updatedAt":  createdAt.Format(time.RFC3339),
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGet(getCmd, []string{"task-123"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGet() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Check for expected fields in text output
	if !strings.Contains(output, "task-123") {
		t.Errorf("output missing task ID: %s", output)
	}
	if !strings.Contains(output, "Test Task") {
		t.Errorf("output missing title: %s", output)
	}
	if !strings.Contains(output, "TODO") {
		t.Errorf("output missing status: %s", output)
	}
}

func TestRunGet_JSONOutput(t *testing.T) {
	createdAt := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"task": map[string]interface{}{
					"id":        "task-json",
					"title":     "JSON Task",
					"status":    "IN_PROGRESS",
					"createdAt": createdAt.Format(time.RFC3339),
					"updatedAt": createdAt.Format(time.RFC3339),
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGet(getCmd, []string{"task-json"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGet() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify JSON output
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("output is not valid JSON: %v", err)
	}
	if result["id"] != "task-json" {
		t.Errorf("JSON output missing 'id': %s", output)
	}
}

func TestRunGet_CSVOutput(t *testing.T) {
	createdAt := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"task": map[string]interface{}{
					"id":        "task-csv",
					"title":     "CSV Task",
					"status":    "DONE",
					"createdAt": createdAt.Format(time.RFC3339),
					"updatedAt": createdAt.Format(time.RFC3339),
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "csv")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGet(getCmd, []string{"task-csv"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGet() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify CSV output has headers
	if !strings.Contains(output, "id,title,status") {
		t.Errorf("CSV output missing headers: %s", output)
	}
	if !strings.Contains(output, "task-csv") {
		t.Errorf("CSV output missing task ID: %s", output)
	}
}

func TestRunGet_NoToken(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	err := runGet(getCmd, []string{"task-123"})
	if err == nil {
		t.Fatal("expected error when no token is available")
	}
}

func TestRunGet_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "not found"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	err := runGet(getCmd, []string{"nonexistent-task"})
	if err == nil {
		t.Fatal("expected error for API error response")
	}
	if !strings.Contains(err.Error(), "failed to get task") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestOutputTask_TextFormat(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	task := &types.Task{
		ID:         "task-output-test",
		Title:      "Output Test Task",
		DueAt:      "2024-12-31T23:59:59Z",
		Status:     "TODO",
		AssigneeID: "user-789",
		CreatedAt:  createdAt,
		UpdatedAt:  createdAt,
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputTask(task, "text", "")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputTask() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	expectedFields := []string{
		"task-output-test",
		"Output Test Task",
		"2024-12-31T23:59:59Z",
		"TODO",
		"user-789",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("output missing field %q: %s", field, output)
		}
	}
}

func TestOutputTask_JSONFormat(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	task := &types.Task{
		ID:        "task-json-test",
		Title:     "JSON Test",
		Status:    "DONE",
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputTask(task, "json", "")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputTask() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("output is not valid JSON: %v", err)
	}
	if result["id"] != "task-json-test" {
		t.Errorf("JSON output wrong 'id': got %v, want %q", result["id"], "task-json-test")
	}
}

func TestOutputTask_CSVFormat(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	task := &types.Task{
		ID:         "task-csv-test",
		Title:      "CSV Test",
		Status:     "IN_PROGRESS",
		DueAt:      "2024-06-30",
		AssigneeID: "user-csv",
		CreatedAt:  createdAt,
		UpdatedAt:  createdAt,
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputTask(task, "csv", "")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputTask() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// CSV should have headers and data row
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		t.Errorf("CSV output should have at least 2 lines (header + data): %s", output)
	}

	if !strings.Contains(lines[0], "id") {
		t.Errorf("CSV header missing 'id': %s", lines[0])
	}
	if !strings.Contains(output, "task-csv-test") {
		t.Errorf("CSV data missing task ID: %s", output)
	}
}

func TestOutputTask_JSONWithQuery(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	task := &types.Task{
		ID:        "query-task",
		Title:     "Query Test",
		Status:    "TODO",
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputTask(task, "json", ".id")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputTask() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	if output != `"query-task"` {
		t.Errorf("output = %q, want %q", output, `"query-task"`)
	}
}
