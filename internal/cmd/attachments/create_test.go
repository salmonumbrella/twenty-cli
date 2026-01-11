package attachments

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
		if !strings.HasSuffix(r.URL.Path, "/rest/attachments") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify request body
		var input map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		if input["name"] != "document.pdf" {
			t.Errorf("name = %v, want %q", input["name"], "document.pdf")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": "attach-new", "name": "document.pdf", "type": "application/pdf", "fullPath": "/files/document.pdf", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	// Save and restore original flag values
	origData := createData
	origDataFile := createDataFile
	t.Cleanup(func() {
		createData = origData
		createDataFile = origDataFile
	})

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	createData = `{"name": "document.pdf", "type": "application/pdf", "fullPath": "/files/document.pdf"}`
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

	if !strings.Contains(output, "attach-new") {
		t.Errorf("output missing attachment ID: %s", output)
	}
}

func TestRunCreate_WithCompanyID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input map[string]interface{}
		json.NewDecoder(r.Body).Decode(&input)

		if input["companyId"] != "company-123" {
			t.Errorf("companyId = %v, want %q", input["companyId"], "company-123")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": "attach-company", "name": "contract.pdf", "type": "application/pdf", "fullPath": "/files/contract.pdf", "companyId": "company-123", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	origData := createData
	origDataFile := createDataFile
	t.Cleanup(func() {
		createData = origData
		createDataFile = origDataFile
	})

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	createData = `{"name": "contract.pdf", "companyId": "company-123", "type": "application/pdf", "fullPath": "/files/contract.pdf"}`
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

	if !strings.Contains(output, "attach-company") {
		t.Errorf("output missing attachment ID: %s", output)
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
		w.Write([]byte(`{"id": "attach-person", "name": "resume.pdf", "type": "application/pdf", "fullPath": "/files/resume.pdf", "personId": "person-456", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	origData := createData
	origDataFile := createDataFile
	t.Cleanup(func() {
		createData = origData
		createDataFile = origDataFile
	})

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	createData = `{"name": "resume.pdf", "personId": "person-456", "type": "application/pdf", "fullPath": "/files/resume.pdf"}`
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

	if !strings.Contains(output, "attach-person") {
		t.Errorf("output missing attachment ID: %s", output)
	}
}

func TestRunCreate_InvalidJSONData(t *testing.T) {
	origData := createData
	origDataFile := createDataFile
	t.Cleanup(func() {
		createData = origData
		createDataFile = origDataFile
	})

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
	t.Cleanup(func() {
		createData = origData
		createDataFile = origDataFile
	})

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
	t.Cleanup(func() {
		createData = origData
		createDataFile = origDataFile
	})

	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	createData = `{"name": "document.pdf"}`
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
	t.Cleanup(func() {
		createData = origData
		createDataFile = origDataFile
	})

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	createData = `{"name": "document.pdf"}`
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
	t.Cleanup(func() {
		createData = origData
		createDataFile = origDataFile
	})

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	createData = `{"name": "document.pdf"}`
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
		w.Write([]byte(`{"id": "attach-query", "name": "report.pdf", "type": "application/pdf", "fullPath": "/files/report.pdf", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	origData := createData
	origDataFile := createDataFile
	t.Cleanup(func() {
		createData = origData
		createDataFile = origDataFile
	})

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", ".id")
	t.Cleanup(viper.Reset)

	createData = `{"name": "report.pdf"}`
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

	if output != `"attach-query"` {
		t.Errorf("output = %q, want %q", output, `"attach-query"`)
	}
}

func TestRunCreate_FromFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": "attach-file", "name": "from-file.pdf", "type": "application/pdf", "fullPath": "/files/from-file.pdf", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	// Create a temp file with JSON content
	tmpFile, err := os.CreateTemp("", "attachment-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(`{"name": "from-file.pdf", "type": "application/pdf", "fullPath": "/files/from-file.pdf"}`); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	origData := createData
	origDataFile := createDataFile
	t.Cleanup(func() {
		createData = origData
		createDataFile = origDataFile
	})

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

	if !strings.Contains(output, "attach-file") {
		t.Errorf("output missing attachment ID: %s", output)
	}
}

func TestRunCreate_WithTaskID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input map[string]interface{}
		json.NewDecoder(r.Body).Decode(&input)

		if input["taskId"] != "task-789" {
			t.Errorf("taskId = %v, want %q", input["taskId"], "task-789")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": "attach-task", "name": "task-doc.pdf", "type": "application/pdf", "fullPath": "/files/task-doc.pdf", "taskId": "task-789", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	origData := createData
	origDataFile := createDataFile
	t.Cleanup(func() {
		createData = origData
		createDataFile = origDataFile
	})

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	createData = `{"name": "task-doc.pdf", "taskId": "task-789", "type": "application/pdf", "fullPath": "/files/task-doc.pdf"}`
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

	if !strings.Contains(output, "attach-task") {
		t.Errorf("output missing attachment ID: %s", output)
	}
}

func TestRunCreate_WithNoteID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input map[string]interface{}
		json.NewDecoder(r.Body).Decode(&input)

		if input["noteId"] != "note-abc" {
			t.Errorf("noteId = %v, want %q", input["noteId"], "note-abc")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": "attach-note", "name": "note-attach.pdf", "type": "application/pdf", "fullPath": "/files/note-attach.pdf", "noteId": "note-abc", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	origData := createData
	origDataFile := createDataFile
	t.Cleanup(func() {
		createData = origData
		createDataFile = origDataFile
	})

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	createData = `{"name": "note-attach.pdf", "noteId": "note-abc", "type": "application/pdf", "fullPath": "/files/note-attach.pdf"}`
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

	if !strings.Contains(output, "attach-note") {
		t.Errorf("output missing attachment ID: %s", output)
	}
}

func TestRunCreate_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer server.Close()

	origData := createData
	origDataFile := createDataFile
	t.Cleanup(func() {
		createData = origData
		createDataFile = origDataFile
	})

	t.Setenv("TWENTY_TOKEN", "invalid-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	createData = `{"name": "document.pdf"}`
	createDataFile = ""

	err := runCreate(createCmd, []string{})
	if err == nil {
		t.Fatal("expected error for unauthorized response")
	}
}

func TestRunCreate_DebugMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": "attach-debug", "name": "debug.pdf", "type": "application/pdf", "fullPath": "/files/debug.pdf", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	origData := createData
	origDataFile := createDataFile
	t.Cleanup(func() {
		createData = origData
		createDataFile = origDataFile
	})

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", true)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	createData = `{"name": "debug.pdf"}`
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

	if !strings.Contains(output, "attach-debug") {
		t.Errorf("output missing attachment ID: %s", output)
	}
}
