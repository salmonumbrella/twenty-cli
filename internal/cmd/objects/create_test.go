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

func TestCreateCmd_HasRunE(t *testing.T) {
	if createCmd.RunE == nil {
		t.Error("createCmd should have RunE set")
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

func TestCreateCmd_FlagShorthands(t *testing.T) {
	dataFlag := createCmd.Flags().Lookup("data")
	if dataFlag == nil {
		t.Fatal("data flag not registered")
	}
	if dataFlag.Shorthand != "d" {
		t.Errorf("data flag shorthand = %q, want %q", dataFlag.Shorthand, "d")
	}

	fileFlag := createCmd.Flags().Lookup("file")
	if fileFlag == nil {
		t.Fatal("file flag not registered")
	}
	if fileFlag.Shorthand != "f" {
		t.Errorf("file flag shorthand = %q, want %q", fileFlag.Shorthand, "f")
	}
}

func TestRunCreate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/rest/metadata/objects" {
			t.Errorf("expected path /rest/metadata/objects, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"new-obj-id","nameSingular":"customObject"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Save original values and restore after test
	oldCreateData := createData
	oldCreateDataFile := createDataFile
	defer func() {
		createData = oldCreateData
		createDataFile = oldCreateDataFile
	}()

	createData = `{"nameSingular":"customObject","namePlural":"customObjects"}`
	createDataFile = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runCreate(nil, nil)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runCreate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "new-obj-id") {
		t.Errorf("output missing 'new-obj-id': %s", output)
	}
}

func TestRunCreate_MissingData(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Save original values and restore after test
	oldCreateData := createData
	oldCreateDataFile := createDataFile
	defer func() {
		createData = oldCreateData
		createDataFile = oldCreateDataFile
	}()

	createData = ""
	createDataFile = ""

	err := runCreate(nil, nil)
	if err == nil {
		t.Fatal("expected error when data is missing")
	}
	if !strings.Contains(err.Error(), "missing JSON payload") {
		t.Errorf("expected 'missing JSON payload' error, got: %v", err)
	}
}

func TestRunCreate_InvalidJSON(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Save original values and restore after test
	oldCreateData := createData
	oldCreateDataFile := createDataFile
	defer func() {
		createData = oldCreateData
		createDataFile = oldCreateDataFile
	}()

	createData = `{invalid json}`
	createDataFile = ""

	err := runCreate(nil, nil)
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

	// Save original values and restore after test
	oldCreateData := createData
	oldCreateDataFile := createDataFile
	defer func() {
		createData = oldCreateData
		createDataFile = oldCreateDataFile
	}()

	createData = `["item1", "item2"]`
	createDataFile = ""

	err := runCreate(nil, nil)
	if err == nil {
		t.Fatal("expected error for non-object JSON")
	}
	if !strings.Contains(err.Error(), "object payload must be a JSON object") {
		t.Errorf("expected 'object payload must be a JSON object' error, got: %v", err)
	}
}

func TestRunCreate_NoToken(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	// Save original values and restore after test
	oldCreateData := createData
	oldCreateDataFile := createDataFile
	defer func() {
		createData = oldCreateData
		createDataFile = oldCreateDataFile
	}()

	createData = `{"nameSingular":"test"}`
	createDataFile = ""

	err := runCreate(nil, nil)
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

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Save original values and restore after test
	oldCreateData := createData
	oldCreateDataFile := createDataFile
	defer func() {
		createData = oldCreateData
		createDataFile = oldCreateDataFile
	}()

	createData = `{"nameSingular":"test"}`
	createDataFile = ""

	err := runCreate(nil, nil)
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestRunCreate_FromFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"file-created-id","nameSingular":"fromFile"}`))
	}))
	defer server.Close()

	// Create temp file with JSON data
	tmpfile, err := os.CreateTemp("", "test*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	content := `{"nameSingular":"fromFile","namePlural":"fromFiles"}`
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
	oldCreateData := createData
	oldCreateDataFile := createDataFile
	defer func() {
		createData = oldCreateData
		createDataFile = oldCreateDataFile
	}()

	createData = ""
	createDataFile = tmpfile.Name()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runCreate(nil, nil)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runCreate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "file-created-id") {
		t.Errorf("output missing 'file-created-id': %s", output)
	}
}

func TestRunCreate_TextOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"new-id","nameSingular":"textTest"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Save original values and restore after test
	oldCreateData := createData
	oldCreateDataFile := createDataFile
	defer func() {
		createData = oldCreateData
		createDataFile = oldCreateDataFile
	}()

	createData = `{"nameSingular":"textTest"}`
	createDataFile = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runCreate(nil, nil)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runCreate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Text output still uses JSON for raw message
	if !strings.Contains(output, "new-id") {
		t.Errorf("output missing 'new-id': %s", output)
	}
}
