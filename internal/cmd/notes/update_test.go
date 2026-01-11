package notes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
)

func TestUpdateCmd_Flags(t *testing.T) {
	flags := []string{"title", "body", "data"}
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
	if updateCmd.Args == nil {
		t.Error("Args validator not set")
	}

	// Test with zero args
	err := updateCmd.Args(updateCmd, []string{})
	if err == nil {
		t.Error("Expected error with zero args")
	}

	// Test with one arg
	err = updateCmd.Args(updateCmd, []string{"id-123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}

	// Test with two args
	err = updateCmd.Args(updateCmd, []string{"id-1", "id-2"})
	if err == nil {
		t.Error("Expected error with two args")
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

func TestUpdateNoteInput_FromJSON(t *testing.T) {
	jsonData := `{
		"title": "Updated Title",
		"bodyV2": {"markdown": "Updated content"}
	}`

	var input rest.UpdateNoteInput
	err := json.Unmarshal([]byte(jsonData), &input)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if input.Title == nil {
		t.Fatal("Title should not be nil")
	}
	if *input.Title != "Updated Title" {
		t.Errorf("Title = %q, want %q", *input.Title, "Updated Title")
	}
	if input.BodyV2 == nil {
		t.Fatal("BodyV2 should not be nil")
	}
	if input.BodyV2.Markdown != "Updated content" {
		t.Errorf("Markdown = %q, want %q", input.BodyV2.Markdown, "Updated content")
	}
}

func TestUpdateNoteInput_InvalidJSON(t *testing.T) {
	invalidJSON := `{invalid json}`

	var input rest.UpdateNoteInput
	err := json.Unmarshal([]byte(invalidJSON), &input)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestUpdateNoteInput_EmptyInput(t *testing.T) {
	input := rest.UpdateNoteInput{}

	if input.Title != nil {
		t.Error("Title should be nil for empty input")
	}
	if input.BodyV2 != nil {
		t.Error("BodyV2 should be nil for empty input")
	}
}

func TestUpdateNoteInput_PartialUpdate(t *testing.T) {
	title := "Only Title"
	input := rest.UpdateNoteInput{
		Title: &title,
	}

	if input.Title == nil {
		t.Fatal("Title should not be nil")
	}
	if *input.Title != "Only Title" {
		t.Errorf("Title = %q, want %q", *input.Title, "Only Title")
	}
	if input.BodyV2 != nil {
		t.Error("BodyV2 should be nil when not updated")
	}
}

func TestUpdateNoteInput_BodyOnly(t *testing.T) {
	input := rest.UpdateNoteInput{
		BodyV2: &rest.BodyV2Input{
			Markdown: "Only body content",
		},
	}

	if input.BodyV2 == nil {
		t.Fatal("BodyV2 should not be nil")
	}
	if input.BodyV2.Markdown != "Only body content" {
		t.Errorf("Markdown = %q, want %q", input.BodyV2.Markdown, "Only body content")
	}
	if input.Title != nil {
		t.Error("Title should be nil when not updated")
	}
}

func TestRunUpdate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/rest/notes/update-id") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"updateNote": {"id": "update-id", "title": "Updated Note", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}}}`))
	}))
	defer server.Close()

	origTitle := updateTitle
	origBody := updateBody
	origData := updateData
	defer func() {
		updateTitle = origTitle
		updateBody = origBody
		updateData = origData
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	updateTitle = ""
	updateBody = ""
	updateData = `{"title": "Updated Note"}`

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runUpdate(updateCmd, []string{"update-id"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runUpdate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "update-id") {
		t.Errorf("output missing note ID: %s", output)
	}
}

func TestRunUpdate_WithTitleFlag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input rest.UpdateNoteInput
		json.NewDecoder(r.Body).Decode(&input)

		if input.Title == nil || *input.Title != "Flag Title" {
			t.Errorf("expected title 'Flag Title', got %v", input.Title)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"updateNote": {"id": "flag-id", "title": "Flag Title", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}}}`))
	}))
	defer server.Close()

	origTitle := updateTitle
	origBody := updateBody
	origData := updateData
	defer func() {
		updateTitle = origTitle
		updateBody = origBody
		updateData = origData
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	updateTitle = "Flag Title"
	updateBody = ""
	updateData = ""

	// Create a new command to test flag.Changed behavior
	cmd := &cobra.Command{}
	cmd.Flags().StringVar(&updateTitle, "title", "", "")
	cmd.Flags().Set("title", "Flag Title")

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runUpdate(cmd, []string{"flag-id"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runUpdate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "flag-id") {
		t.Errorf("output missing note ID: %s", output)
	}
}

func TestRunUpdate_WithBodyFlag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input rest.UpdateNoteInput
		json.NewDecoder(r.Body).Decode(&input)

		if input.BodyV2 == nil || input.BodyV2.Markdown != "Flag Body Content" {
			t.Errorf("expected body 'Flag Body Content', got %v", input.BodyV2)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"updateNote": {"id": "body-id", "title": "Body Note", "bodyV2": {"markdown": "Flag Body Content"}, "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}}}`))
	}))
	defer server.Close()

	origTitle := updateTitle
	origBody := updateBody
	origData := updateData
	defer func() {
		updateTitle = origTitle
		updateBody = origBody
		updateData = origData
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	updateTitle = ""
	updateBody = "Flag Body Content"
	updateData = ""

	// Create a new command to test flag.Changed behavior
	cmd := &cobra.Command{}
	cmd.Flags().StringVar(&updateBody, "body", "", "")
	cmd.Flags().Set("body", "Flag Body Content")

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runUpdate(cmd, []string{"body-id"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runUpdate() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "body-id") {
		t.Errorf("output missing note ID: %s", output)
	}
}

func TestRunUpdate_JSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"updateNote": {"id": "json-id", "title": "JSON Output Note", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}}}`))
	}))
	defer server.Close()

	origTitle := updateTitle
	origBody := updateBody
	origData := updateData
	defer func() {
		updateTitle = origTitle
		updateBody = origBody
		updateData = origData
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	updateTitle = ""
	updateBody = ""
	updateData = `{"title": "JSON Output Note"}`

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runUpdate(updateCmd, []string{"json-id"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runUpdate() error = %v", err)
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

func TestRunUpdate_InvalidJSONData(t *testing.T) {
	origTitle := updateTitle
	origBody := updateBody
	origData := updateData
	defer func() {
		updateTitle = origTitle
		updateBody = origBody
		updateData = origData
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", "http://localhost")
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	updateTitle = ""
	updateBody = ""
	updateData = `{invalid json}`

	err := runUpdate(updateCmd, []string{"test-id"})
	if err == nil {
		t.Fatal("expected error for invalid JSON data")
	}

	if !strings.Contains(err.Error(), "invalid JSON data") {
		t.Errorf("error = %q, want error containing 'invalid JSON data'", err.Error())
	}
}

func TestRunUpdate_NoToken(t *testing.T) {
	origTitle := updateTitle
	origBody := updateBody
	origData := updateData
	defer func() {
		updateTitle = origTitle
		updateBody = origBody
		updateData = origData
	}()

	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	updateTitle = ""
	updateBody = ""
	updateData = `{"title": "Test"}`

	err := runUpdate(updateCmd, []string{"test-id"})
	if err == nil {
		t.Fatal("expected error when no token is available")
	}
}

func TestRunUpdate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "bad request"}`))
	}))
	defer server.Close()

	origTitle := updateTitle
	origBody := updateBody
	origData := updateData
	defer func() {
		updateTitle = origTitle
		updateBody = origBody
		updateData = origData
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	updateTitle = ""
	updateBody = ""
	updateData = `{"title": "Test"}`

	err := runUpdate(updateCmd, []string{"test-id"})
	if err == nil {
		t.Fatal("expected error for API error response")
	}

	if !strings.Contains(err.Error(), "failed to update note") {
		t.Errorf("error = %q, want error containing 'failed to update note'", err.Error())
	}
}

func TestRunUpdate_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "not found"}`))
	}))
	defer server.Close()

	origTitle := updateTitle
	origBody := updateBody
	origData := updateData
	defer func() {
		updateTitle = origTitle
		updateBody = origBody
		updateData = origData
	}()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	updateTitle = ""
	updateBody = ""
	updateData = `{"title": "Test"}`

	err := runUpdate(updateCmd, []string{"nonexistent-id"})
	if err == nil {
		t.Fatal("expected error for not found response")
	}
}
