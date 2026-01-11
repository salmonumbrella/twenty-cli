package notes

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"

	"github.com/salmonumbrella/twenty-cli/internal/types"
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
	if getCmd.Args == nil {
		t.Error("Args validator not set")
	}

	// Test with zero args
	err := getCmd.Args(getCmd, []string{})
	if err == nil {
		t.Error("Expected error with zero args")
	}

	// Test with one arg
	err = getCmd.Args(getCmd, []string{"id-123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}

	// Test with two args
	err = getCmd.Args(getCmd, []string{"id-1", "id-2"})
	if err == nil {
		t.Error("Expected error with two args")
	}
}

func TestRunGet_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/rest/notes/note-123") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"note": {"id": "note-123", "title": "Test Note", "bodyV2": {"markdown": "Note content"}, "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}}}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGet(getCmd, []string{"note-123"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGet() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "note-123") {
		t.Errorf("output missing note ID: %s", output)
	}
	if !strings.Contains(output, "Test Note") {
		t.Errorf("output missing note title: %s", output)
	}
}

func TestRunGet_JSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"note": {"id": "note-json", "title": "JSON Note", "bodyV2": {"markdown": "Content"}, "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}}}`))
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

	err := runGet(getCmd, []string{"note-json"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGet() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should be valid JSON
	var result types.Note
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("output is not valid JSON: %v", err)
	}

	if result.ID != "note-json" {
		t.Errorf("parsed ID = %q, want %q", result.ID, "note-json")
	}
}

func TestRunGet_CSVOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"note": {"id": "note-csv", "title": "CSV Note", "bodyV2": {"markdown": "CSV Content"}, "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}}}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "csv")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGet(getCmd, []string{"note-csv"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGet() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should be valid CSV
	reader := csv.NewReader(strings.NewReader(output))
	records, err := reader.ReadAll()
	if err != nil {
		t.Errorf("output is not valid CSV: %v", err)
	}

	// Should have header + 1 data row
	if len(records) != 2 {
		t.Errorf("expected 2 rows, got %d", len(records))
	}
}

func TestRunGet_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "not found"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	err := runGet(getCmd, []string{"nonexistent-id"})
	if err == nil {
		t.Fatal("expected error for not found response")
	}

	if !strings.Contains(err.Error(), "failed to get note") {
		t.Errorf("error = %q, want error containing 'failed to get note'", err.Error())
	}
}

func TestRunGet_NoToken(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	err := runGet(getCmd, []string{"test-id"})
	if err == nil {
		t.Fatal("expected error when no token is available")
	}
}

func TestOutputNote_JSON(t *testing.T) {
	now := time.Now()
	n := &types.Note{
		ID:    "note-1",
		Title: "Test Note",
		BodyV2: &types.BodyV2{
			Markdown: "Note body content",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputNote(n, "json", "")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputNote() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	var parsed types.Note
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if parsed.ID != "note-1" {
		t.Errorf("ID = %q, want %q", parsed.ID, "note-1")
	}
	if parsed.Title != "Test Note" {
		t.Errorf("Title = %q, want %q", parsed.Title, "Test Note")
	}
}

func TestOutputNote_CSV(t *testing.T) {
	now := time.Now()
	n := &types.Note{
		ID:    "note-csv",
		Title: "CSV Test Note",
		BodyV2: &types.BodyV2{
			Markdown: "CSV body",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputNote(n, "csv", "")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputNote() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	reader := csv.NewReader(strings.NewReader(output))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("output is not valid CSV: %v", err)
	}

	// Should have header + 1 data row
	if len(records) != 2 {
		t.Errorf("expected 2 rows, got %d", len(records))
	}

	// Check header
	expectedHeaders := []string{"id", "title", "body", "createdAt", "updatedAt"}
	for i, h := range expectedHeaders {
		if records[0][i] != h {
			t.Errorf("header[%d] = %q, want %q", i, records[0][i], h)
		}
	}

	// Check data
	if records[1][0] != "note-csv" {
		t.Errorf("ID = %q, want %q", records[1][0], "note-csv")
	}
	if records[1][1] != "CSV Test Note" {
		t.Errorf("Title = %q, want %q", records[1][1], "CSV Test Note")
	}
}

func TestOutputNote_Text(t *testing.T) {
	now := time.Now()
	n := &types.Note{
		ID:    "note-text",
		Title: "Text Test Note",
		BodyV2: &types.BodyV2{
			Markdown: "Text body content",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputNote(n, "text", "")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputNote() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "note-text") {
		t.Errorf("output should contain ID")
	}
	if !strings.Contains(output, "Text Test Note") {
		t.Errorf("output should contain title")
	}
	if !strings.Contains(output, "Text body content") {
		t.Errorf("output should contain body")
	}
}

func TestOutputNote_DefaultFormat(t *testing.T) {
	now := time.Now()
	n := &types.Note{
		ID:        "note-default",
		Title:     "Default Format Note",
		CreatedAt: now,
		UpdatedAt: now,
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Empty format string should use default (text)
	err := outputNote(n, "", "")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputNote() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "note-default") {
		t.Errorf("output should contain ID")
	}
}

func TestOutputNote_NilBodyV2(t *testing.T) {
	now := time.Now()
	n := &types.Note{
		ID:        "note-nobody",
		Title:     "No Body Note",
		BodyV2:    nil,
		CreatedAt: now,
		UpdatedAt: now,
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputNote(n, "text", "")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputNote() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Body should be empty string
	if !strings.Contains(output, "Body:") {
		t.Errorf("output should contain Body field")
	}
}

func TestOutputNote_CSVWithNilBodyV2(t *testing.T) {
	now := time.Now()
	n := &types.Note{
		ID:        "note-csv-nobody",
		Title:     "CSV No Body Note",
		BodyV2:    nil,
		CreatedAt: now,
		UpdatedAt: now,
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputNote(n, "csv", "")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputNote() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	reader := csv.NewReader(strings.NewReader(output))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("output is not valid CSV: %v", err)
	}

	// Body column (index 2) should be empty
	if records[1][2] != "" {
		t.Errorf("body = %q, want empty string", records[1][2])
	}
}
