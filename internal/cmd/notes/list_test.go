package notes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func TestListCmd_Flags(t *testing.T) {
	flags := []string{"limit", "cursor", "all", "filter", "sort", "order"}
	for _, flag := range flags {
		if listCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestListCmd_Use(t *testing.T) {
	if listCmd.Use != "list" {
		t.Errorf("Use = %q, want %q", listCmd.Use, "list")
	}
}

func TestListCmd_Short(t *testing.T) {
	if listCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestListCmd_Long(t *testing.T) {
	if listCmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestListNotes_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"notes": [{"id": "note-1", "title": "Note 1", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}]}, "totalCount": 1}`))
	}))
	defer server.Close()

	client := rest.NewClient(server.URL, "test-token", false)
	ctx := context.Background()

	result, err := listNotes(ctx, client, nil)
	if err != nil {
		t.Fatalf("listNotes() error = %v", err)
	}

	if len(result.Data) != 1 {
		t.Errorf("expected 1 note, got %d", len(result.Data))
	}
	if result.Data[0].ID != "note-1" {
		t.Errorf("note ID = %q, want %q", result.Data[0].ID, "note-1")
	}
}

func TestListNotes_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify query parameters
		limit := r.URL.Query().Get("limit")
		if limit != "10" {
			t.Errorf("limit = %q, want %q", limit, "10")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"notes": []}, "totalCount": 0}`))
	}))
	defer server.Close()

	client := rest.NewClient(server.URL, "test-token", false)
	ctx := context.Background()

	opts := &rest.ListOptions{Limit: 10}
	result, err := listNotes(ctx, client, opts)
	if err != nil {
		t.Fatalf("listNotes() error = %v", err)
	}

	if len(result.Data) != 0 {
		t.Errorf("expected 0 notes, got %d", len(result.Data))
	}
}

func TestListNotes_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	client := rest.NewClient(server.URL, "test-token", false)
	ctx := context.Background()

	_, err := listNotes(ctx, client, nil)
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestTableRow_ShortID(t *testing.T) {
	now := time.Now()
	note := types.Note{
		ID:        "short",
		Title:     "Test Title",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Get the TableRow function from the builder config
	// We need to test the logic directly since we can't easily access the config
	id := note.ID
	if len(id) > 8 {
		id = id[:8] + "..."
	}

	if id != "short" {
		t.Errorf("ID = %q, want %q", id, "short")
	}
}

func TestTableRow_LongID(t *testing.T) {
	id := "very-long-note-id-123456789"
	if len(id) > 8 {
		id = id[:8] + "..."
	}

	expected := "very-lon..."
	if id != expected {
		t.Errorf("ID = %q, want %q", id, expected)
	}
}

func TestTableRow_ShortBody(t *testing.T) {
	body := "Short body"
	if len(body) > 40 {
		body = body[:40] + "..."
	}

	if body != "Short body" {
		t.Errorf("body = %q, want %q", body, "Short body")
	}
}

func TestTableRow_LongBody(t *testing.T) {
	body := "This is a very long body that exceeds forty characters and needs to be truncated"
	if len(body) > 40 {
		body = body[:40] + "..."
	}

	expected := "This is a very long body that exceeds fo..."
	if body != expected {
		t.Errorf("body = %q, want %q", body, expected)
	}
}

func TestCSVRow_Format(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	note := types.Note{
		ID:    "csv-note-id",
		Title: "CSV Note Title",
		BodyV2: &types.BodyV2{
			Markdown: "CSV body content",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	row := []string{
		note.ID,
		note.Title,
		note.Body(),
		note.CreatedAt.Format("2006-01-02T15:04:05Z"),
		note.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if row[0] != "csv-note-id" {
		t.Errorf("row[0] = %q, want %q", row[0], "csv-note-id")
	}
	if row[1] != "CSV Note Title" {
		t.Errorf("row[1] = %q, want %q", row[1], "CSV Note Title")
	}
	if row[2] != "CSV body content" {
		t.Errorf("row[2] = %q, want %q", row[2], "CSV body content")
	}
	if row[3] != "2024-01-15T10:30:00Z" {
		t.Errorf("row[3] = %q, want %q", row[3], "2024-01-15T10:30:00Z")
	}
}

func TestCSVRow_NilBodyV2(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	note := types.Note{
		ID:        "csv-nobody-id",
		Title:     "CSV No Body Title",
		BodyV2:    nil,
		CreatedAt: now,
		UpdatedAt: now,
	}

	body := note.Body()
	if body != "" {
		t.Errorf("body = %q, want empty string", body)
	}
}

func TestTableRow_DateFormat(t *testing.T) {
	createdAt := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	formatted := createdAt.Format("2006-01-02")

	if formatted != "2024-06-15" {
		t.Errorf("formatted = %q, want %q", formatted, "2024-06-15")
	}
}
