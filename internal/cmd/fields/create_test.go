package fields

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

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

func TestCreateCmd_Flags(t *testing.T) {
	flags := []string{"data", "file"}
	for _, flag := range flags {
		if createCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
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

func TestRunCreate_Success(t *testing.T) {
	// Set up mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/rest/metadata/fields" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{
			"id": "new-field-id",
			"objectMetadataId": "obj-123",
			"name": "customField",
			"label": "Custom Field",
			"type": "TEXT",
			"isCustom": true,
			"isActive": true,
			"isNullable": true
		}`))
	}))
	defer server.Close()

	// Set up environment and viper
	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Save and restore package-level variables
	oldData := createData
	oldDataFile := createDataFile
	defer func() {
		createData = oldData
		createDataFile = oldDataFile
	}()

	createData = `{"objectMetadataId": "obj-123", "name": "customField", "label": "Custom Field", "type": "TEXT"}`
	createDataFile = ""

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

	if !strings.Contains(output, "new-field-id") {
		t.Errorf("output missing 'new-field-id': %s", output)
	}
	if !strings.Contains(output, "customField") {
		t.Errorf("output missing 'customField': %s", output)
	}
}

func TestRunCreate_MissingData(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Save and restore package-level variables
	oldData := createData
	oldDataFile := createDataFile
	defer func() {
		createData = oldData
		createDataFile = oldDataFile
	}()

	createData = ""
	createDataFile = ""

	err := runCreate(createCmd, []string{})
	if err == nil {
		t.Fatal("expected error when data is missing")
	}
	if !strings.Contains(err.Error(), "missing JSON payload") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunCreate_InvalidJSON(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Save and restore package-level variables
	oldData := createData
	oldDataFile := createDataFile
	defer func() {
		createData = oldData
		createDataFile = oldDataFile
	}()

	createData = "{invalid json}"
	createDataFile = ""

	err := runCreate(createCmd, []string{})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestRunCreate_NonObjectJSON(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Save and restore package-level variables
	oldData := createData
	oldDataFile := createDataFile
	defer func() {
		createData = oldData
		createDataFile = oldDataFile
	}()

	// JSON array instead of object
	createData = `["item1", "item2"]`
	createDataFile = ""

	err := runCreate(createCmd, []string{})
	if err == nil {
		t.Fatal("expected error for non-object JSON")
	}
	if !strings.Contains(err.Error(), "must be a JSON object") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunCreate_NoToken(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	// Save and restore package-level variables
	oldData := createData
	oldDataFile := createDataFile
	defer func() {
		createData = oldData
		createDataFile = oldDataFile
	}()

	createData = `{"name": "test"}`
	createDataFile = ""

	err := runCreate(createCmd, []string{})
	if err == nil {
		t.Fatal("expected error when no token is available")
	}
}

func TestRunCreate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "validation error", "message": "name is required"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Save and restore package-level variables
	oldData := createData
	oldDataFile := createDataFile
	defer func() {
		createData = oldData
		createDataFile = oldDataFile
	}()

	createData = `{"type": "TEXT"}`
	createDataFile = ""

	err := runCreate(createCmd, []string{})
	if err == nil {
		t.Fatal("expected error for API error")
	}
}

func TestRunCreate_FromFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": "file-field-id", "name": "fileField"}`))
	}))
	defer server.Close()

	// Create temp file
	tmpfile, err := os.CreateTemp("", "field*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	content := `{"objectMetadataId": "obj-123", "name": "fileField", "label": "File Field", "type": "TEXT"}`
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpfile.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Save and restore package-level variables
	oldData := createData
	oldDataFile := createDataFile
	defer func() {
		createData = oldData
		createDataFile = oldDataFile
	}()

	createData = ""
	createDataFile = tmpfile.Name()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runCreate(createCmd, []string{})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runCreate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "file-field-id") {
		t.Errorf("output missing 'file-field-id': %s", output)
	}
}

func TestRunCreate_TextOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": "new-field-id", "name": "customField"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Save and restore package-level variables
	oldData := createData
	oldDataFile := createDataFile
	defer func() {
		createData = oldData
		createDataFile = oldDataFile
	}()

	createData = `{"name": "customField"}`
	createDataFile = ""

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

	// Should output JSON regardless of format (per implementation)
	if !strings.Contains(output, "new-field-id") {
		t.Errorf("output missing 'new-field-id': %s", output)
	}
}

func TestRunCreate_EmptyOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": "new-field-id"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Save and restore package-level variables
	oldData := createData
	oldDataFile := createDataFile
	defer func() {
		createData = oldData
		createDataFile = oldDataFile
	}()

	createData = `{"name": "test"}`
	createDataFile = ""

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

	if !strings.Contains(output, "new-field-id") {
		t.Errorf("output missing 'new-field-id': %s", output)
	}
}
