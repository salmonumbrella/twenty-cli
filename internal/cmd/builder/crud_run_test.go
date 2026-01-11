package builder

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestReadJSONInput_DataFlag(t *testing.T) {
	data, err := readJSONPayload(`{"name": "John"}`, "", payloadOptions{})
	if err != nil {
		t.Fatalf("readJSONPayload() error = %v", err)
	}

	if data == nil {
		t.Fatal("expected non-nil data")
	}

	obj, ok := data.(map[string]interface{})
	if !ok {
		t.Fatal("expected JSON object")
	}

	name, ok := obj["name"]
	if !ok {
		t.Fatal("expected 'name' key in data")
	}

	if name != "John" {
		t.Errorf("name = %q, want %q", name, "John")
	}
}

func TestReadJSONInput_FileFlag(t *testing.T) {
	// Create a temp file with JSON data
	tmpfile, err := os.CreateTemp("", "test*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	content := `{"email": "test@example.com"}`
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpfile.Close()

	data, err := readJSONPayload("", tmpfile.Name(), payloadOptions{})
	if err != nil {
		t.Fatalf("readJSONPayload() error = %v", err)
	}

	if data == nil {
		t.Fatal("expected non-nil data")
	}

	obj, ok := data.(map[string]interface{})
	if !ok {
		t.Fatal("expected JSON object")
	}

	email, ok := obj["email"]
	if !ok {
		t.Fatal("expected 'email' key in data")
	}

	if email != "test@example.com" {
		t.Errorf("email = %q, want %q", email, "test@example.com")
	}
}

func TestReadJSONInput_BothEmpty(t *testing.T) {
	_, err := readJSONPayload("", "", payloadOptions{})
	if err == nil {
		t.Fatal("expected error when both data and file are empty")
	}

	if !strings.Contains(err.Error(), "missing JSON payload") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestReadJSONInput_InvalidJSON(t *testing.T) {
	_, err := readJSONPayload(`{invalid json}`, "", payloadOptions{})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestReadJSONInput_NonObjectJSON(t *testing.T) {
	_, err := readJSONPayload(`["item1", "item2"]`, "", payloadOptions{})
	if err == nil {
		t.Fatal("expected error for non-object JSON")
	}
}

func TestOutputRaw_JSON(t *testing.T) {
	data := json.RawMessage(`{"id": "123", "name": "Test"}`)

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := writeOutput(data, "json", "", false)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputRaw() error = %v", err)
	}

	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, `"id"`) {
		t.Errorf("output missing 'id' field: %s", output)
	}
	if !strings.Contains(output, `"123"`) {
		t.Errorf("output missing '123' value: %s", output)
	}
}

func TestOutputRaw_YAML(t *testing.T) {
	data := json.RawMessage(`{"id": "456", "name": "Test"}`)

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := writeOutput(data, "yaml", "", false)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputRaw() error = %v", err)
	}

	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "id:") {
		t.Errorf("output missing 'id:' field in YAML: %s", output)
	}
}

func TestOutputRaw_CSV(t *testing.T) {
	// CSV requires an array of records
	data := json.RawMessage(`[{"id": "789", "name": "Test"}]`)

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := writeOutput(data, "csv", "", false)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputRaw() error = %v", err)
	}

	buf.ReadFrom(r)
	output := buf.String()

	// CSV should have headers and data
	if !strings.Contains(output, "id") {
		t.Errorf("output missing 'id' header in CSV: %s", output)
	}
}

func TestOutputRaw_Table(t *testing.T) {
	// Table requires an array of records
	data := json.RawMessage(`[{"id": "999", "name": "TableTest"}]`)

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := writeOutput(data, "text", "", false)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputRaw() error = %v", err)
	}

	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "999") {
		t.Errorf("output missing '999' in table: %s", output)
	}
}

func TestRunCreate_Success(t *testing.T) {
	// Set up mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/rest/people") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": "new-id", "name": "Created"}`))
	}))
	defer server.Close()

	// Set up environment and viper
	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runCreate(CreateConfig{Resource: "people"}, `{"name": "Test"}`, "")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runCreate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "new-id") {
		t.Errorf("output missing 'new-id': %s", output)
	}
}

func TestRunCreate_MissingData(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	t.Cleanup(viper.Reset)

	err := runCreate(CreateConfig{Resource: "people"}, "", "")
	if err == nil {
		t.Fatal("expected error when data is missing")
	}
}

func TestRunUpdate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/rest/people/update-id") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "update-id", "name": "Updated"}`))
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

	err := runUpdate(UpdateConfig{Resource: "people"}, "update-id", `{"name": "Updated"}`, "")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runUpdate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "update-id") {
		t.Errorf("output missing 'update-id': %s", output)
	}
}

func TestRunUpdate_MissingData(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	t.Cleanup(viper.Reset)

	err := runUpdate(UpdateConfig{Resource: "people"}, "id-123", "", "")
	if err == nil {
		t.Fatal("expected error when data is missing")
	}
}

func TestRunDelete_Force(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/rest/people/delete-id") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runDelete("people", "delete-id", true)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runDelete() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Deleted people delete-id") {
		t.Errorf("output missing confirmation: %s", output)
	}
}

func TestRunDelete_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "not found"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	err := runDelete("people", "nonexistent", true)
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestRunCreate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "bad request"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	err := runCreate(CreateConfig{Resource: "people"}, `{"name": "Test"}`, "")
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestRunUpdate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"error": "validation error"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	err := runUpdate(UpdateConfig{Resource: "people"}, "id-123", `{"name": "Test"}`, "")
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestOutputRaw_JSONWithQuery(t *testing.T) {
	data := json.RawMessage(`{"id": "123", "name": "Test"}`)

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := writeOutput(data, "json", ".id", true)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputRaw() with query error = %v", err)
	}

	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	if output != `"123"` {
		t.Errorf("output = %q, want %q", output, `"123"`)
	}
}

func TestReadJSONInput_FileNotFound(t *testing.T) {
	_, err := readJSONPayload("", "/nonexistent/path/file.json", payloadOptions{})
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestRunUpdate_NoToken(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	err := runUpdate(UpdateConfig{Resource: "people"}, "id-123", `{"name": "Test"}`, "")
	if err == nil {
		t.Fatal("expected error when no token is available")
	}
}

func TestRunDelete_Confirmed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	// Save original stdin and restore later
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create a pipe and write "y\n" to simulate user confirmation
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		w.Write([]byte("y\n"))
		w.Close()
	}()

	// Capture stdout
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	err := runDelete("people", "confirm-id", false)
	wOut.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runDelete() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(rOut)
	output := buf.String()

	if !strings.Contains(output, "Deleted people confirm-id") {
		t.Errorf("output missing confirmation: %s", output)
	}
}

func TestRunDelete_Canceled(t *testing.T) {
	// Save original stdin and restore later
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create a pipe and write "n\n" to simulate user declining
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		w.Write([]byte("n\n"))
		w.Close()
	}()

	// Capture stdout
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	// No server needed - should cancel before making API call
	err := runDelete("people", "cancel-id", false)
	wOut.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runDelete() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(rOut)
	output := buf.String()

	if !strings.Contains(output, "Canceled") {
		t.Errorf("output missing 'Canceled': %s", output)
	}
}

func TestRunDelete_YesConfirmation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	t.Cleanup(viper.Reset)

	// Save original stdin and restore later
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create a pipe and write "yes\n" to simulate user confirmation
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		w.Write([]byte("yes\n"))
		w.Close()
	}()

	// Capture stdout
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	err := runDelete("people", "yes-id", false)
	wOut.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runDelete() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(rOut)
	output := buf.String()

	if !strings.Contains(output, "Deleted people yes-id") {
		t.Errorf("output missing confirmation: %s", output)
	}
}

func TestRunDelete_NoToken(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	err := runDelete("people", "test-id", true)
	if err == nil {
		t.Fatal("expected error when no token is available")
	}
}

func TestRunCreate_NoToken(t *testing.T) {
	// Unset the token environment variable and use nonexistent profile
	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	// This should fail because there's no token
	err := runCreate(CreateConfig{Resource: "people"}, `{"name": "Test"}`, "")
	if err == nil {
		t.Fatal("expected error when no token is available")
	}
}
