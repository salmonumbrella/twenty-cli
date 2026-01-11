package objects

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
	if updateCmd.Use != "update <object-id>" {
		t.Errorf("Use = %q, want %q", updateCmd.Use, "update <object-id>")
	}
}

func TestUpdateCmd_Short(t *testing.T) {
	if updateCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestUpdateCmd_HasRunE(t *testing.T) {
	if updateCmd.RunE == nil {
		t.Error("updateCmd should have RunE set")
	}
}

func TestUpdateCmd_Args(t *testing.T) {
	if updateCmd.Args == nil {
		t.Error("updateCmd should have Args validation set")
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

func TestUpdateCmd_FlagShorthands(t *testing.T) {
	dataFlag := updateCmd.Flags().Lookup("data")
	if dataFlag == nil {
		t.Fatal("data flag not registered")
	}
	if dataFlag.Shorthand != "d" {
		t.Errorf("data flag shorthand = %q, want %q", dataFlag.Shorthand, "d")
	}

	fileFlag := updateCmd.Flags().Lookup("file")
	if fileFlag == nil {
		t.Fatal("file flag not registered")
	}
	if fileFlag.Shorthand != "f" {
		t.Errorf("file flag shorthand = %q, want %q", fileFlag.Shorthand, "f")
	}
}

func TestRunUpdate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/rest/metadata/objects/obj-123" {
			t.Errorf("expected path /rest/metadata/objects/obj-123, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"obj-123","nameSingular":"updatedObject","description":"Updated"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Save original values and restore after test
	oldUpdateData := updateData
	oldUpdateDataFile := updateDataFile
	defer func() {
		updateData = oldUpdateData
		updateDataFile = oldUpdateDataFile
	}()

	updateData = `{"description":"Updated"}`
	updateDataFile = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runUpdate(nil, []string{"obj-123"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runUpdate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "obj-123") {
		t.Errorf("output missing 'obj-123': %s", output)
	}
	if !strings.Contains(output, "Updated") {
		t.Errorf("output missing 'Updated': %s", output)
	}
}

func TestRunUpdate_MissingData(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Save original values and restore after test
	oldUpdateData := updateData
	oldUpdateDataFile := updateDataFile
	defer func() {
		updateData = oldUpdateData
		updateDataFile = oldUpdateDataFile
	}()

	updateData = ""
	updateDataFile = ""

	err := runUpdate(nil, []string{"obj-123"})
	if err == nil {
		t.Fatal("expected error when data is missing")
	}
	if !strings.Contains(err.Error(), "missing JSON payload") {
		t.Errorf("expected 'missing JSON payload' error, got: %v", err)
	}
}

func TestRunUpdate_InvalidJSON(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Save original values and restore after test
	oldUpdateData := updateData
	oldUpdateDataFile := updateDataFile
	defer func() {
		updateData = oldUpdateData
		updateDataFile = oldUpdateDataFile
	}()

	updateData = `{invalid json}`
	updateDataFile = ""

	err := runUpdate(nil, []string{"obj-123"})
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

	// Save original values and restore after test
	oldUpdateData := updateData
	oldUpdateDataFile := updateDataFile
	defer func() {
		updateData = oldUpdateData
		updateDataFile = oldUpdateDataFile
	}()

	updateData = `["item1", "item2"]`
	updateDataFile = ""

	err := runUpdate(nil, []string{"obj-123"})
	if err == nil {
		t.Fatal("expected error for non-object JSON")
	}
	if !strings.Contains(err.Error(), "object payload must be a JSON object") {
		t.Errorf("expected 'object payload must be a JSON object' error, got: %v", err)
	}
}

func TestRunUpdate_NoToken(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	// Save original values and restore after test
	oldUpdateData := updateData
	oldUpdateDataFile := updateDataFile
	defer func() {
		updateData = oldUpdateData
		updateDataFile = oldUpdateDataFile
	}()

	updateData = `{"description":"test"}`
	updateDataFile = ""

	err := runUpdate(nil, []string{"obj-123"})
	if err == nil {
		t.Fatal("expected error when no token is available")
	}
}

func TestRunUpdate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "object not found"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Save original values and restore after test
	oldUpdateData := updateData
	oldUpdateDataFile := updateDataFile
	defer func() {
		updateData = oldUpdateData
		updateDataFile = oldUpdateDataFile
	}()

	updateData = `{"description":"test"}`
	updateDataFile = ""

	err := runUpdate(nil, []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestRunUpdate_FromFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"obj-file","description":"From File"}`))
	}))
	defer server.Close()

	// Create temp file with JSON data
	tmpfile, err := os.CreateTemp("", "test*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	content := `{"description":"From File"}`
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

	// Save original values and restore after test
	oldUpdateData := updateData
	oldUpdateDataFile := updateDataFile
	defer func() {
		updateData = oldUpdateData
		updateDataFile = oldUpdateDataFile
	}()

	updateData = ""
	updateDataFile = tmpfile.Name()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runUpdate(nil, []string{"obj-file"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runUpdate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "obj-file") {
		t.Errorf("output missing 'obj-file': %s", output)
	}
}

func TestRunUpdate_TextOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"obj-text","nameSingular":"textTest"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Save original values and restore after test
	oldUpdateData := updateData
	oldUpdateDataFile := updateDataFile
	defer func() {
		updateData = oldUpdateData
		updateDataFile = oldUpdateDataFile
	}()

	updateData = `{"nameSingular":"textTest"}`
	updateDataFile = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runUpdate(nil, []string{"obj-text"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runUpdate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Text output still uses JSON for raw message
	if !strings.Contains(output, "obj-text") {
		t.Errorf("output missing 'obj-text': %s", output)
	}
}
