package favorites

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

func TestCreateCmd_Flags(t *testing.T) {
	flags := []string{"data", "file"}
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/rest/favorites") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify request body
		var input map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		if input["companyId"] != "company-123" {
			t.Errorf("companyId = %v, want %q", input["companyId"], "company-123")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": "fav-new", "position": 1, "workspaceMemberId": "member-1", "companyId": "company-123", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	// Save and restore original flag values
	origData := createData
	origDataFile := createDataFile
	defer func() {
		createData = origData
		createDataFile = origDataFile
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	createData = `{"companyId": "company-123", "position": 1}`
	createDataFile = ""

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

	if !strings.Contains(output, "fav-new") {
		t.Errorf("output missing favorite ID: %s", output)
	}
}

func TestRunCreate_WithPersonID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input map[string]interface{}
		json.NewDecoder(r.Body).Decode(&input)

		if input["personId"] != "person-456" {
			t.Errorf("personId = %v, want %q", input["personId"], "person-456")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": "fav-person", "position": 2, "workspaceMemberId": "member-1", "personId": "person-456", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	origData := createData
	origDataFile := createDataFile
	defer func() {
		createData = origData
		createDataFile = origDataFile
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	createData = `{"personId": "person-456", "position": 2}`
	createDataFile = ""

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

	if !strings.Contains(output, "fav-person") {
		t.Errorf("output missing favorite ID: %s", output)
	}
}

func TestRunCreate_InvalidJSONData(t *testing.T) {
	origData := createData
	origDataFile := createDataFile
	defer func() {
		createData = origData
		createDataFile = origDataFile
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	createData = `{invalid json}`
	createDataFile = ""

	err := runCreate(createCmd, []string{})
	if err == nil {
		t.Fatal("expected error for invalid JSON data")
	}

	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("error = %q, want error containing 'invalid'", err.Error())
	}
}

func TestRunCreate_MissingPayload(t *testing.T) {
	origData := createData
	origDataFile := createDataFile
	defer func() {
		createData = origData
		createDataFile = origDataFile
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	createData = ""
	createDataFile = ""

	err := runCreate(createCmd, []string{})
	if err == nil {
		t.Fatal("expected error when no payload provided")
	}

	if !strings.Contains(err.Error(), "missing JSON payload") {
		t.Errorf("error = %q, want error containing 'missing JSON payload'", err.Error())
	}
}

func TestRunCreate_NoToken(t *testing.T) {
	origData := createData
	origDataFile := createDataFile
	defer func() {
		createData = origData
		createDataFile = origDataFile
	}()

	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	createData = `{"companyId": "company-123"}`
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
		w.Write([]byte(`{"error": "bad request"}`))
	}))
	defer server.Close()

	origData := createData
	origDataFile := createDataFile
	defer func() {
		createData = origData
		createDataFile = origDataFile
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	createData = `{"companyId": "company-123"}`
	createDataFile = ""

	err := runCreate(createCmd, []string{})
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestRunCreate_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	origData := createData
	origDataFile := createDataFile
	defer func() {
		createData = origData
		createDataFile = origDataFile
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	createData = `{"companyId": "company-123"}`
	createDataFile = ""

	err := runCreate(createCmd, []string{})
	if err == nil {
		t.Fatal("expected error for server error response")
	}
}

func TestRunCreate_WithQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": "fav-query", "position": 1, "workspaceMemberId": "member-1", "companyId": "company-123", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	origData := createData
	origDataFile := createDataFile
	defer func() {
		createData = origData
		createDataFile = origDataFile
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", ".id")
	t.Cleanup(viper.Reset)

	createData = `{"companyId": "company-123"}`
	createDataFile = ""

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
	output := strings.TrimSpace(buf.String())

	if output != `"fav-query"` {
		t.Errorf("output = %q, want %q", output, `"fav-query"`)
	}
}

func TestRunCreate_FromFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": "fav-file", "position": 1, "workspaceMemberId": "member-1", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	// Create a temp file with JSON content
	tmpFile, err := os.CreateTemp("", "favorite-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(`{"taskId": "task-123", "position": 1}`); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	origData := createData
	origDataFile := createDataFile
	defer func() {
		createData = origData
		createDataFile = origDataFile
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	createData = ""
	createDataFile = tmpFile.Name()

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

	if !strings.Contains(output, "fav-file") {
		t.Errorf("output missing favorite ID: %s", output)
	}
}
