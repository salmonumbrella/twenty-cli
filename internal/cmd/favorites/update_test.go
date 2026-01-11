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

func TestUpdateCmd_Flags(t *testing.T) {
	flags := []string{"data", "file"}
	for _, flag := range flags {
		if updateCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestUpdateCmd_Use(t *testing.T) {
	if updateCmd.Use != "update <id>" {
		t.Errorf("Use = %q, want %q", updateCmd.Use, "update <id>")
	}
}

func TestUpdateCmd_Short(t *testing.T) {
	if updateCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestUpdateCmd_Args(t *testing.T) {
	// Test that command requires exactly one argument
	err := updateCmd.Args(updateCmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = updateCmd.Args(updateCmd, []string{"id1", "id2"})
	if err == nil {
		t.Error("expected error when too many args provided")
	}

	err = updateCmd.Args(updateCmd, []string{"id1"})
	if err != nil {
		t.Errorf("unexpected error for valid args: %v", err)
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

func TestRunUpdate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/rest/favorites/fav-to-update") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify request body
		var input map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		if input["position"] != float64(5) {
			t.Errorf("position = %v, want %v", input["position"], 5)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "fav-to-update", "position": 5, "workspaceMemberId": "member-1", "companyId": "company-123", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-02T00:00:00Z"}`))
	}))
	defer server.Close()

	// Save and restore original flag values
	origData := updateData
	origDataFile := updateDataFile
	defer func() {
		updateData = origData
		updateDataFile = origDataFile
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	updateData = `{"position": 5}`
	updateDataFile = ""

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runUpdate(updateCmd, []string{"fav-to-update"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runUpdate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "fav-to-update") {
		t.Errorf("output missing favorite ID: %s", output)
	}
}

func TestRunUpdate_WithJSONData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input map[string]interface{}
		json.NewDecoder(r.Body).Decode(&input)

		if input["position"] != float64(10) {
			t.Errorf("position = %v, want %v", input["position"], 10)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "fav-json-update", "position": 10, "workspaceMemberId": "member-1", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-02T00:00:00Z"}`))
	}))
	defer server.Close()

	origData := updateData
	origDataFile := updateDataFile
	defer func() {
		updateData = origData
		updateDataFile = origDataFile
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	updateData = `{"position": 10}`
	updateDataFile = ""

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runUpdate(updateCmd, []string{"fav-json-update"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runUpdate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "fav-json-update") {
		t.Errorf("output missing favorite ID: %s", output)
	}
}

func TestRunUpdate_InvalidJSONData(t *testing.T) {
	origData := updateData
	origDataFile := updateDataFile
	defer func() {
		updateData = origData
		updateDataFile = origDataFile
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	updateData = `{invalid json}`
	updateDataFile = ""

	err := runUpdate(updateCmd, []string{"fav-123"})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRunUpdate_MissingPayload(t *testing.T) {
	origData := updateData
	origDataFile := updateDataFile
	defer func() {
		updateData = origData
		updateDataFile = origDataFile
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	updateData = ""
	updateDataFile = ""

	err := runUpdate(updateCmd, []string{"fav-123"})
	if err == nil {
		t.Fatal("expected error when no payload provided")
	}

	if !strings.Contains(err.Error(), "missing JSON payload") {
		t.Errorf("error = %q, want error containing 'missing JSON payload'", err.Error())
	}
}

func TestRunUpdate_NoToken(t *testing.T) {
	origData := updateData
	origDataFile := updateDataFile
	defer func() {
		updateData = origData
		updateDataFile = origDataFile
	}()

	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	updateData = `{"position": 5}`
	updateDataFile = ""

	err := runUpdate(updateCmd, []string{"fav-123"})
	if err == nil {
		t.Fatal("expected error when no token is available")
	}
}

func TestRunUpdate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"error": "validation error"}`))
	}))
	defer server.Close()

	origData := updateData
	origDataFile := updateDataFile
	defer func() {
		updateData = origData
		updateDataFile = origDataFile
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	updateData = `{"position": 5}`
	updateDataFile = ""

	err := runUpdate(updateCmd, []string{"fav-123"})
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestRunUpdate_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	origData := updateData
	origDataFile := updateDataFile
	defer func() {
		updateData = origData
		updateDataFile = origDataFile
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	updateData = `{"position": 5}`
	updateDataFile = ""

	err := runUpdate(updateCmd, []string{"fav-123"})
	if err == nil {
		t.Fatal("expected error for server error response")
	}
}

func TestRunUpdate_WithQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "fav-query", "position": 7, "workspaceMemberId": "member-1", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-02T00:00:00Z"}`))
	}))
	defer server.Close()

	origData := updateData
	origDataFile := updateDataFile
	defer func() {
		updateData = origData
		updateDataFile = origDataFile
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", ".position")
	t.Cleanup(viper.Reset)

	updateData = `{"position": 7}`
	updateDataFile = ""

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runUpdate(updateCmd, []string{"fav-query"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runUpdate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	if output != "7" {
		t.Errorf("output = %q, want %q", output, "7")
	}
}

func TestRunUpdate_FromFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "fav-file", "position": 15, "workspaceMemberId": "member-1", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-02T00:00:00Z"}`))
	}))
	defer server.Close()

	// Create a temp file with JSON content
	tmpFile, err := os.CreateTemp("", "favorite-update-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(`{"position": 15}`); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	origData := updateData
	origDataFile := updateDataFile
	defer func() {
		updateData = origData
		updateDataFile = origDataFile
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	updateData = ""
	updateDataFile = tmpFile.Name()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runUpdate(updateCmd, []string{"fav-file"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runUpdate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "fav-file") {
		t.Errorf("output missing favorite ID: %s", output)
	}
}

func TestRunUpdate_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "favorite not found"}`))
	}))
	defer server.Close()

	origData := updateData
	origDataFile := updateDataFile
	defer func() {
		updateData = origData
		updateDataFile = origDataFile
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	updateData = `{"position": 5}`
	updateDataFile = ""

	err := runUpdate(updateCmd, []string{"nonexistent-fav"})
	if err == nil {
		t.Fatal("expected error for not found response")
	}
}
