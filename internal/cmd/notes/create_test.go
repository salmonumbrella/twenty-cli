package notes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
)

func TestCreateCmd_Flags(t *testing.T) {
	flags := []string{"title", "body", "data"}
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

func TestCreateNoteInput_FromFlags(t *testing.T) {
	input := rest.CreateNoteInput{
		Title: "Test Note",
		BodyV2: &rest.BodyV2Input{
			Markdown: "Note body content",
		},
	}

	if input.Title != "Test Note" {
		t.Errorf("Title = %q, want %q", input.Title, "Test Note")
	}
	if input.BodyV2 == nil {
		t.Fatal("BodyV2 should not be nil")
	}
	if input.BodyV2.Markdown != "Note body content" {
		t.Errorf("Markdown = %q, want %q", input.BodyV2.Markdown, "Note body content")
	}
}

func TestCreateNoteInput_FromJSON(t *testing.T) {
	jsonData := `{
		"title": "JSON Note",
		"bodyV2": {"markdown": "JSON body content"}
	}`

	var input rest.CreateNoteInput
	err := json.Unmarshal([]byte(jsonData), &input)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if input.Title != "JSON Note" {
		t.Errorf("Title = %q, want %q", input.Title, "JSON Note")
	}
	if input.BodyV2 == nil {
		t.Fatal("BodyV2 should not be nil")
	}
	if input.BodyV2.Markdown != "JSON body content" {
		t.Errorf("Markdown = %q, want %q", input.BodyV2.Markdown, "JSON body content")
	}
}

func TestCreateNoteInput_InvalidJSON(t *testing.T) {
	invalidJSON := `{invalid json}`

	var input rest.CreateNoteInput
	err := json.Unmarshal([]byte(invalidJSON), &input)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestCreateNoteInput_EmptyFields(t *testing.T) {
	input := rest.CreateNoteInput{}

	if input.Title != "" {
		t.Errorf("Title should be empty, got %q", input.Title)
	}
	if input.BodyV2 != nil {
		t.Error("BodyV2 should be nil for empty input")
	}
}

func TestCreateNoteInput_TitleOnly(t *testing.T) {
	input := rest.CreateNoteInput{
		Title: "Title Only Note",
	}

	if input.Title != "Title Only Note" {
		t.Errorf("Title = %q, want %q", input.Title, "Title Only Note")
	}
	if input.BodyV2 != nil {
		t.Error("BodyV2 should be nil when not set")
	}
}

func TestRunCreate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/rest/notes") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"data": {"createNote": {"id": "note-123", "title": "Created Note", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}}}`))
	}))
	defer server.Close()

	// Save and restore original flag values
	origTitle := createTitle
	origBody := createBody
	origData := createData
	defer func() {
		createTitle = origTitle
		createBody = origBody
		createData = origData
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	createTitle = "Created Note"
	createBody = ""
	createData = ""

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

	if !strings.Contains(output, "note-123") {
		t.Errorf("output missing note ID: %s", output)
	}
}

func TestRunCreate_WithBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input rest.CreateNoteInput
		json.NewDecoder(r.Body).Decode(&input)

		if input.BodyV2 == nil {
			t.Error("expected BodyV2 to be set")
		} else if input.BodyV2.Markdown != "Test body content" {
			t.Errorf("BodyV2.Markdown = %q, want %q", input.BodyV2.Markdown, "Test body content")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"data": {"createNote": {"id": "note-456", "title": "Note with Body", "bodyV2": {"markdown": "Test body content"}, "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}}}`))
	}))
	defer server.Close()

	origTitle := createTitle
	origBody := createBody
	origData := createData
	defer func() {
		createTitle = origTitle
		createBody = origBody
		createData = origData
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	createTitle = "Note with Body"
	createBody = "Test body content"
	createData = ""

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

	if !strings.Contains(output, "note-456") {
		t.Errorf("output missing note ID: %s", output)
	}
}

func TestRunCreate_WithJSONData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input rest.CreateNoteInput
		json.NewDecoder(r.Body).Decode(&input)

		if input.Title != "JSON Data Note" {
			t.Errorf("Title = %q, want %q", input.Title, "JSON Data Note")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"data": {"createNote": {"id": "note-789", "title": "JSON Data Note", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}}}`))
	}))
	defer server.Close()

	origTitle := createTitle
	origBody := createBody
	origData := createData
	defer func() {
		createTitle = origTitle
		createBody = origBody
		createData = origData
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	createTitle = ""
	createBody = ""
	createData = `{"title": "JSON Data Note"}`

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

	if !strings.Contains(output, "note-789") {
		t.Errorf("output missing note ID: %s", output)
	}
}

func TestRunCreate_JSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"data": {"createNote": {"id": "note-json", "title": "JSON Output Note", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}}}`))
	}))
	defer server.Close()

	origTitle := createTitle
	origBody := createBody
	origData := createData
	defer func() {
		createTitle = origTitle
		createBody = origBody
		createData = origData
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	createTitle = "JSON Output Note"
	createBody = ""
	createData = ""

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

	// Should be valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("output is not valid JSON: %v", err)
	}
}

func TestRunCreate_InvalidJSONData(t *testing.T) {
	origTitle := createTitle
	origBody := createBody
	origData := createData
	defer func() {
		createTitle = origTitle
		createBody = origBody
		createData = origData
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	createTitle = ""
	createBody = ""
	createData = `{invalid json}`

	err := runCreate(createCmd, []string{})
	if err == nil {
		t.Fatal("expected error for invalid JSON data")
	}

	if !strings.Contains(err.Error(), "invalid JSON data") {
		t.Errorf("error = %q, want error containing 'invalid JSON data'", err.Error())
	}
}

func TestRunCreate_NoToken(t *testing.T) {
	origTitle := createTitle
	origBody := createBody
	origData := createData
	defer func() {
		createTitle = origTitle
		createBody = origBody
		createData = origData
	}()

	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	createTitle = "Test"
	createBody = ""
	createData = ""

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

	origTitle := createTitle
	origBody := createBody
	origData := createData
	defer func() {
		createTitle = origTitle
		createBody = origBody
		createData = origData
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	createTitle = "Test Note"
	createBody = ""
	createData = ""

	err := runCreate(createCmd, []string{})
	if err == nil {
		t.Fatal("expected error for API error response")
	}

	if !strings.Contains(err.Error(), "failed to create note") {
		t.Errorf("error = %q, want error containing 'failed to create note'", err.Error())
	}
}
