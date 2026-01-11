package rest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// Test command flags
func TestRequestCmd_Flags(t *testing.T) {
	flags := []string{"data", "data-file"}
	for _, flag := range flags {
		if requestCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered on requestCmd", flag)
		}
	}
}

func TestPostCmd_Flags(t *testing.T) {
	flags := []string{"data", "data-file"}
	for _, flag := range flags {
		if postCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered on postCmd", flag)
		}
	}
}

func TestPatchCmd_Flags(t *testing.T) {
	flags := []string{"data", "data-file"}
	for _, flag := range flags {
		if patchCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered on patchCmd", flag)
		}
	}
}

func TestDeleteCmd_Flags(t *testing.T) {
	flags := []string{"data", "data-file"}
	for _, flag := range flags {
		if deleteCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered on deleteCmd", flag)
		}
	}
}

func TestDataFlag_Shorthand(t *testing.T) {
	flag := requestCmd.Flags().Lookup("data")
	if flag == nil {
		t.Fatal("data flag not registered")
	}
	if flag.Shorthand != "d" {
		t.Errorf("data flag shorthand = %q, want %q", flag.Shorthand, "d")
	}
}

func TestDataFileFlag_Shorthand(t *testing.T) {
	flag := requestCmd.Flags().Lookup("data-file")
	if flag == nil {
		t.Fatal("data-file flag not registered")
	}
	if flag.Shorthand != "f" {
		t.Errorf("data-file flag shorthand = %q, want %q", flag.Shorthand, "f")
	}
}

// Test command args validation
func TestRequestCmd_Args(t *testing.T) {
	// Should require exactly 2 args
	err := requestCmd.Args(requestCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = requestCmd.Args(requestCmd, []string{"GET"})
	if err == nil {
		t.Error("expected error when only 1 arg provided")
	}

	err = requestCmd.Args(requestCmd, []string{"GET", "/path", "extra"})
	if err == nil {
		t.Error("expected error when too many args provided")
	}

	err = requestCmd.Args(requestCmd, []string{"GET", "/path"})
	if err != nil {
		t.Errorf("unexpected error for valid args: %v", err)
	}
}

func TestGetCmd_Args(t *testing.T) {
	err := getCmd.Args(getCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = getCmd.Args(getCmd, []string{"/path", "extra"})
	if err == nil {
		t.Error("expected error when too many args provided")
	}

	err = getCmd.Args(getCmd, []string{"/path"})
	if err != nil {
		t.Errorf("unexpected error for valid args: %v", err)
	}
}

func TestPostCmd_Args(t *testing.T) {
	err := postCmd.Args(postCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = postCmd.Args(postCmd, []string{"/path"})
	if err != nil {
		t.Errorf("unexpected error for valid args: %v", err)
	}
}

func TestPatchCmd_Args(t *testing.T) {
	err := patchCmd.Args(patchCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = patchCmd.Args(patchCmd, []string{"/path"})
	if err != nil {
		t.Errorf("unexpected error for valid args: %v", err)
	}
}

func TestDeleteCmd_Args(t *testing.T) {
	err := deleteCmd.Args(deleteCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = deleteCmd.Args(deleteCmd, []string{"/path"})
	if err != nil {
		t.Errorf("unexpected error for valid args: %v", err)
	}
}

// Test normalizePath function
func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty path",
			input:    "",
			expected: "/rest",
		},
		{
			name:     "path without leading slash",
			input:    "companies",
			expected: "/rest/companies",
		},
		{
			name:     "path with leading slash but no prefix",
			input:    "/companies",
			expected: "/rest/companies",
		},
		{
			name:     "path already has /rest prefix",
			input:    "/rest/companies",
			expected: "/rest/companies",
		},
		{
			name:     "path with /graphql prefix",
			input:    "/graphql",
			expected: "/graphql",
		},
		{
			name:     "path with /metadata prefix",
			input:    "/metadata/objects",
			expected: "/metadata/objects",
		},
		{
			name:     "path without slash and /rest prefix needed",
			input:    "people/123",
			expected: "/rest/people/123",
		},
		{
			name:     "path starting with rest without slash",
			input:    "rest/companies",
			expected: "/rest/companies",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePath(tt.input)
			if result != tt.expected {
				t.Errorf("normalizePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Test parseBody function
func TestParseBody_NoBody(t *testing.T) {
	// Save and restore package-level vars
	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()

	bodyData = ""
	bodyDataFile = ""

	result, err := parseBody()
	if err != nil {
		t.Fatalf("parseBody() error = %v", err)
	}
	if result != nil {
		t.Errorf("parseBody() = %v, want nil", result)
	}
}

func TestParseBody_WithData(t *testing.T) {
	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()

	bodyData = `{"name": "test", "value": 123}`
	bodyDataFile = ""

	result, err := parseBody()
	if err != nil {
		t.Fatalf("parseBody() error = %v", err)
	}

	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("parseBody() result is not a map")
	}

	if m["name"] != "test" {
		t.Errorf("parseBody() name = %v, want %q", m["name"], "test")
	}
	if m["value"] != float64(123) {
		t.Errorf("parseBody() value = %v, want %v", m["value"], 123)
	}
}

func TestParseBody_InvalidJSON(t *testing.T) {
	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()

	bodyData = `{invalid json}`
	bodyDataFile = ""

	_, err := parseBody()
	if err == nil {
		t.Fatal("parseBody() should return error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid JSON body") {
		t.Errorf("parseBody() error = %v, want error containing 'invalid JSON body'", err)
	}
}

func TestParseBody_FromFile(t *testing.T) {
	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()

	// Create a temp file with JSON content
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "body.json")
	content := `{"file": "data", "num": 42}`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	bodyData = ""
	bodyDataFile = tmpFile

	result, err := parseBody()
	if err != nil {
		t.Fatalf("parseBody() error = %v", err)
	}

	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("parseBody() result is not a map")
	}

	if m["file"] != "data" {
		t.Errorf("parseBody() file = %v, want %q", m["file"], "data")
	}
	if m["num"] != float64(42) {
		t.Errorf("parseBody() num = %v, want %v", m["num"], 42)
	}
}

func TestParseBody_FromNonexistentFile(t *testing.T) {
	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()

	bodyData = ""
	bodyDataFile = "/nonexistent/path/to/file.json"

	_, err := parseBody()
	if err == nil {
		t.Fatal("parseBody() should return error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "read body file") {
		t.Errorf("parseBody() error = %v, want error containing 'read body file'", err)
	}
}

func TestParseBody_FromFileWithInvalidJSON(t *testing.T) {
	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()

	// Create a temp file with invalid JSON
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(tmpFile, []byte(`{not valid json}`), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	bodyData = ""
	bodyDataFile = tmpFile

	_, err := parseBody()
	if err == nil {
		t.Fatal("parseBody() should return error for invalid JSON in file")
	}
	if !strings.Contains(err.Error(), "invalid JSON body") {
		t.Errorf("parseBody() error = %v, want error containing 'invalid JSON body'", err)
	}
}

func TestParseBody_DataFileTakesPrecedence(t *testing.T) {
	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()

	// Create a temp file with JSON content
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "body.json")
	content := `{"source": "file"}`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	bodyData = `{"source": "flag"}`
	bodyDataFile = tmpFile

	result, err := parseBody()
	if err != nil {
		t.Fatalf("parseBody() error = %v", err)
	}

	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("parseBody() result is not a map")
	}

	// File should take precedence over flag
	if m["source"] != "file" {
		t.Errorf("parseBody() source = %v, want %q (file should take precedence)", m["source"], "file")
	}
}

func TestParseBody_JSONArray(t *testing.T) {
	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()

	bodyData = `[{"id": 1}, {"id": 2}]`
	bodyDataFile = ""

	result, err := parseBody()
	if err != nil {
		t.Fatalf("parseBody() error = %v", err)
	}

	arr, ok := result.([]interface{})
	if !ok {
		t.Fatalf("parseBody() result is not an array")
	}

	if len(arr) != 2 {
		t.Errorf("parseBody() array length = %d, want 2", len(arr))
	}
}

// Test runRequest with mock server
func TestRunRequest_GETSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/rest/companies") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "1", "name": "Company 1"},
				{"id": "2", "name": "Company 2"},
			},
		})
	}))
	defer server.Close()

	// Reset package vars
	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()
	bodyData = ""
	bodyDataFile = ""

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

	err := runRequest("GET", "/companies")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runRequest() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Company 1") {
		t.Errorf("output missing Company 1: %s", output)
	}
}

func TestRunRequest_POSTSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		// Verify request body
		var input map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		if input["name"] != "New Company" {
			t.Errorf("name = %q, want %q", input["name"], "New Company")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "new-123",
				"name": "New Company",
			},
		})
	}))
	defer server.Close()

	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()
	bodyData = `{"name": "New Company"}`
	bodyDataFile = ""

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runRequest("POST", "/companies")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runRequest() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "new-123") {
		t.Errorf("output missing new-123: %s", output)
	}
}

func TestRunRequest_PATCHSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "123",
				"name": "Updated Company",
			},
		})
	}))
	defer server.Close()

	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()
	bodyData = `{"name": "Updated Company"}`
	bodyDataFile = ""

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runRequest("PATCH", "/companies/123")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runRequest() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Updated Company") {
		t.Errorf("output missing Updated Company: %s", output)
	}
}

func TestRunRequest_DELETESuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"deletedAt": "2024-01-15T10:30:00Z",
			},
		})
	}))
	defer server.Close()

	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()
	bodyData = ""
	bodyDataFile = ""

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runRequest("DELETE", "/companies/123")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runRequest() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "deletedAt") {
		t.Errorf("output missing deletedAt: %s", output)
	}
}

func TestRunRequest_NoToken(t *testing.T) {
	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()
	bodyData = ""
	bodyDataFile = ""

	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	err := runRequest("GET", "/companies")
	if err == nil {
		t.Fatal("expected error when no token is available")
	}
}

func TestRunRequest_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "bad request"}`))
	}))
	defer server.Close()

	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()
	bodyData = ""
	bodyDataFile = ""

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	err := runRequest("GET", "/companies")
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestRunRequest_TextOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "text-test",
				"name": "Text Output Test",
			},
		})
	}))
	defer server.Close()

	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()
	bodyData = ""
	bodyDataFile = ""

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runRequest("GET", "/companies/text-test")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runRequest() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "text-test") {
		t.Errorf("output missing text-test: %s", output)
	}
}

func TestRunRequest_EmptyOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": "test",
		})
	}))
	defer server.Close()

	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()
	bodyData = ""
	bodyDataFile = ""

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "") // empty output should default to JSON
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runRequest("GET", "/companies")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runRequest() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if output == "" {
		t.Error("expected some output")
	}
}

func TestRunRequest_UnsupportedOutputFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": "test",
		})
	}))
	defer server.Close()

	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()
	bodyData = ""
	bodyDataFile = ""

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "csv") // unsupported, should default to JSON
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runRequest("GET", "/companies")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runRequest() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if output == "" {
		t.Error("expected some output")
	}
}

func TestRunRequest_WithQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "query-test",
				"name": "Query Test",
			},
		})
	}))
	defer server.Close()

	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()
	bodyData = ""
	bodyDataFile = ""

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", ".data.id")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runRequest("GET", "/companies/query-test")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runRequest() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	if output != `"query-test"` {
		t.Errorf("output = %q, want %q", output, `"query-test"`)
	}
}

func TestRunRequest_MethodCaseInsensitive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": "ok"})
	}))
	defer server.Close()

	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()
	bodyData = ""
	bodyDataFile = ""

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	// Test lowercase method - should be converted to uppercase
	err := runRequest("get", "/companies")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runRequest() error = %v", err)
	}
}

func TestRunRequest_InvalidBody(t *testing.T) {
	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()
	bodyData = `{invalid json}`
	bodyDataFile = ""

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	err := runRequest("POST", "/companies")
	if err == nil {
		t.Fatal("expected error for invalid JSON body")
	}
	if !strings.Contains(err.Error(), "invalid JSON body") {
		t.Errorf("error = %v, want error containing 'invalid JSON body'", err)
	}
}

// Test subcommand RunE functions
func TestGetCmd_RunE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": "test"})
	}))
	defer server.Close()

	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()
	bodyData = ""
	bodyDataFile = ""

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := getCmd.RunE(getCmd, []string{"/companies"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("getCmd.RunE() error = %v", err)
	}
}

func TestPostCmd_RunE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": "created"})
	}))
	defer server.Close()

	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()
	bodyData = `{"name": "test"}`
	bodyDataFile = ""

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := postCmd.RunE(postCmd, []string{"/companies"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("postCmd.RunE() error = %v", err)
	}
}

func TestPatchCmd_RunE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": "updated"})
	}))
	defer server.Close()

	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()
	bodyData = `{"name": "updated"}`
	bodyDataFile = ""

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := patchCmd.RunE(patchCmd, []string{"/companies/123"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("patchCmd.RunE() error = %v", err)
	}
}

func TestDeleteCmd_RunE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": "deleted"})
	}))
	defer server.Close()

	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()
	bodyData = ""
	bodyDataFile = ""

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := deleteCmd.RunE(deleteCmd, []string{"/companies/123"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("deleteCmd.RunE() error = %v", err)
	}
}

func TestRequestCmd_RunE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": "put-response"})
	}))
	defer server.Close()

	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()
	bodyData = `{"name": "put-test"}`
	bodyDataFile = ""

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := requestCmd.RunE(requestCmd, []string{"PUT", "/companies/123"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("requestCmd.RunE() error = %v", err)
	}
}

func TestRunRequest_WithDebug(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": "debug-test"})
	}))
	defer server.Close()

	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()
	bodyData = ""
	bodyDataFile = ""

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", true) // Enable debug mode
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := runRequest("GET", "/companies")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runRequest() error = %v", err)
	}
}

func TestRunRequest_PathWithGraphQL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/graphql") {
			t.Errorf("path should start with /graphql, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": "graphql-response"})
	}))
	defer server.Close()

	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()
	bodyData = ""
	bodyDataFile = ""

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := runRequest("POST", "/graphql")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runRequest() error = %v", err)
	}
}

func TestRunRequest_PathWithMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/metadata") {
			t.Errorf("path should start with /metadata, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": "metadata-response"})
	}))
	defer server.Close()

	origData := bodyData
	origDataFile := bodyDataFile
	defer func() {
		bodyData = origData
		bodyDataFile = origDataFile
	}()
	bodyData = ""
	bodyDataFile = ""

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := runRequest("GET", "/metadata/objects")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runRequest() error = %v", err)
	}
}
