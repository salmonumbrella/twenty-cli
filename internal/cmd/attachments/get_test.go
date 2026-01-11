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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/rest/attachments/attach-123") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "attach-123", "name": "document.pdf", "type": "application/pdf", "fullPath": "/files/document.pdf", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGet(getCmd, []string{"attach-123"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGet() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "attach-123") {
		t.Errorf("output missing attachment ID: %s", output)
	}
}

func TestRunGet_JSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "attach-json", "name": "report.pdf", "type": "application/pdf", "fullPath": "/files/report.pdf", "companyId": "company-456", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGet(getCmd, []string{"attach-json"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGet() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify output is valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("output is not valid JSON: %v", err)
	}
	if result["id"] != "attach-json" {
		t.Errorf("JSON output wrong 'id': got %v, want %q", result["id"], "attach-json")
	}
	if result["name"] != "report.pdf" {
		t.Errorf("JSON output wrong 'name': got %v, want %q", result["name"], "report.pdf")
	}
}

func TestRunGet_WithQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "attach-query", "name": "data.csv", "type": "text/csv", "fullPath": "/files/data.csv", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", ".name")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGet(getCmd, []string{"attach-query"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGet() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	if output != `"data.csv"` {
		t.Errorf("output = %q, want %q", output, `"data.csv"`)
	}
}

func TestRunGet_NoToken(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	err := runGet(getCmd, []string{"attach-123"})
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
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	err := runGet(getCmd, []string{"nonexistent-attach"})
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestRunGet_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	err := runGet(getCmd, []string{"attach-123"})
	if err == nil {
		t.Fatal("expected error for server error response")
	}
}

func TestRunGet_WithFullPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "attach-path", "name": "image.png", "type": "image/png", "fullPath": "/uploads/2024/01/image.png", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", ".fullPath")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGet(getCmd, []string{"attach-path"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGet() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	if output != `"/uploads/2024/01/image.png"` {
		t.Errorf("output = %q, want %q", output, `"/uploads/2024/01/image.png"`)
	}
}

func TestRunGet_DebugMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "attach-debug", "name": "debug.pdf", "type": "application/pdf", "fullPath": "/files/debug.pdf", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", true) // Enable debug mode
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGet(getCmd, []string{"attach-debug"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGet() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "attach-debug") {
		t.Errorf("output missing attachment ID: %s", output)
	}
}

func TestRunGet_WithCompanyID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "attach-company", "name": "contract.pdf", "type": "application/pdf", "fullPath": "/files/contract.pdf", "companyId": "company-789", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", ".companyId")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGet(getCmd, []string{"attach-company"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGet() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	if output != `"company-789"` {
		t.Errorf("output = %q, want %q", output, `"company-789"`)
	}
}

func TestRunGet_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "invalid-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	err := runGet(getCmd, []string{"attach-123"})
	if err == nil {
		t.Fatal("expected error for unauthorized response")
	}
}

func TestRunGet_WithType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "attach-type", "name": "spreadsheet.xlsx", "type": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "fullPath": "/files/spreadsheet.xlsx", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", ".type")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGet(getCmd, []string{"attach-type"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGet() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	if output != `"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"` {
		t.Errorf("output = %q, want xlsx mime type", output)
	}
}
