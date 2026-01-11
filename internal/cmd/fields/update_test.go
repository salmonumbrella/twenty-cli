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

func TestUpdateCmd_Use(t *testing.T) {
	if updateCmd.Use != "update <field-id>" {
		t.Errorf("Use = %q, want %q", updateCmd.Use, "update <field-id>")
	}
}

func TestUpdateCmd_Short(t *testing.T) {
	if updateCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestUpdateCmd_Flags(t *testing.T) {
	flags := []string{"data", "file"}
	for _, flag := range flags {
		if updateCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
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

func TestUpdateCmd_FileFlagShorthand(t *testing.T) {
	flag := updateCmd.Flags().Lookup("file")
	if flag == nil {
		t.Fatal("file flag not registered")
	}
	if flag.Shorthand != "f" {
		t.Errorf("file flag shorthand = %q, want %q", flag.Shorthand, "f")
	}
}

func TestUpdateCmd_Args(t *testing.T) {
	// Command should require exactly 1 argument
	if updateCmd.Args == nil {
		t.Error("Args validator not set")
	}

	// Test with zero args
	err := updateCmd.Args(updateCmd, []string{})
	if err == nil {
		t.Error("Expected error with zero args")
	}

	// Test with one arg
	err = updateCmd.Args(updateCmd, []string{"field-id"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}

	// Test with two args
	err = updateCmd.Args(updateCmd, []string{"field-1", "field-2"})
	if err == nil {
		t.Error("Expected error with two args")
	}
}

func TestRunUpdate_Success(t *testing.T) {
	// Set up mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/rest/metadata/fields/field-123") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "field-123",
			"objectMetadataId": "obj-123",
			"name": "updatedField",
			"label": "Updated Field",
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
	oldData := updateData
	oldDataFile := updateDataFile
	defer func() {
		updateData = oldData
		updateDataFile = oldDataFile
	}()

	updateData = `{"label": "Updated Field"}`
	updateDataFile = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runUpdate(updateCmd, []string{"field-123"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runUpdate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "field-123") {
		t.Errorf("output missing 'field-123': %s", output)
	}
	if !strings.Contains(output, "updatedField") {
		t.Errorf("output missing 'updatedField': %s", output)
	}
}

func TestRunUpdate_MissingData(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Save and restore package-level variables
	oldData := updateData
	oldDataFile := updateDataFile
	defer func() {
		updateData = oldData
		updateDataFile = oldDataFile
	}()

	updateData = ""
	updateDataFile = ""

	err := runUpdate(updateCmd, []string{"field-123"})
	if err == nil {
		t.Fatal("expected error when data is missing")
	}
	if !strings.Contains(err.Error(), "missing JSON payload") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunUpdate_InvalidJSON(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Save and restore package-level variables
	oldData := updateData
	oldDataFile := updateDataFile
	defer func() {
		updateData = oldData
		updateDataFile = oldDataFile
	}()

	updateData = "{invalid json}"
	updateDataFile = ""

	err := runUpdate(updateCmd, []string{"field-123"})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestRunUpdate_NonObjectJSON(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Save and restore package-level variables
	oldData := updateData
	oldDataFile := updateDataFile
	defer func() {
		updateData = oldData
		updateDataFile = oldDataFile
	}()

	// JSON array instead of object
	updateData = `["item1", "item2"]`
	updateDataFile = ""

	err := runUpdate(updateCmd, []string{"field-123"})
	if err == nil {
		t.Fatal("expected error for non-object JSON")
	}
	if !strings.Contains(err.Error(), "must be a JSON object") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunUpdate_NoToken(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	// Save and restore package-level variables
	oldData := updateData
	oldDataFile := updateDataFile
	defer func() {
		updateData = oldData
		updateDataFile = oldDataFile
	}()

	updateData = `{"label": "test"}`
	updateDataFile = ""

	err := runUpdate(updateCmd, []string{"field-123"})
	if err == nil {
		t.Fatal("expected error when no token is available")
	}
}

func TestRunUpdate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "field not found"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Save and restore package-level variables
	oldData := updateData
	oldDataFile := updateDataFile
	defer func() {
		updateData = oldData
		updateDataFile = oldDataFile
	}()

	updateData = `{"label": "test"}`
	updateDataFile = ""

	err := runUpdate(updateCmd, []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for API error")
	}
}

func TestRunUpdate_FromFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "field-123", "label": "File Updated"}`))
	}))
	defer server.Close()

	// Create temp file
	tmpfile, err := os.CreateTemp("", "field*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	content := `{"label": "File Updated"}`
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
	oldData := updateData
	oldDataFile := updateDataFile
	defer func() {
		updateData = oldData
		updateDataFile = oldDataFile
	}()

	updateData = ""
	updateDataFile = tmpfile.Name()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runUpdate(updateCmd, []string{"field-123"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runUpdate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "File Updated") {
		t.Errorf("output missing 'File Updated': %s", output)
	}
}

func TestRunUpdate_TextOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "field-123", "name": "updatedField"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Save and restore package-level variables
	oldData := updateData
	oldDataFile := updateDataFile
	defer func() {
		updateData = oldData
		updateDataFile = oldDataFile
	}()

	updateData = `{"name": "updatedField"}`
	updateDataFile = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runUpdate(updateCmd, []string{"field-123"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runUpdate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should output JSON regardless of format (per implementation)
	if !strings.Contains(output, "field-123") {
		t.Errorf("output missing 'field-123': %s", output)
	}
}

func TestRunUpdate_EmptyOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "field-123"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Save and restore package-level variables
	oldData := updateData
	oldDataFile := updateDataFile
	defer func() {
		updateData = oldData
		updateDataFile = oldDataFile
	}()

	updateData = `{"name": "test"}`
	updateDataFile = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runUpdate(updateCmd, []string{"field-123"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runUpdate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "field-123") {
		t.Errorf("output missing 'field-123': %s", output)
	}
}

func TestRunUpdate_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"error": "validation error", "message": "invalid field type"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Save and restore package-level variables
	oldData := updateData
	oldDataFile := updateDataFile
	defer func() {
		updateData = oldData
		updateDataFile = oldDataFile
	}()

	updateData = `{"type": "INVALID"}`
	updateDataFile = ""

	err := runUpdate(updateCmd, []string{"field-123"})
	if err == nil {
		t.Fatal("expected error for validation error")
	}
}
