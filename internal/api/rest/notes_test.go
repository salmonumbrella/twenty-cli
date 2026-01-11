package rest

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func TestClient_ListNotes(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	expectedNotes := []types.Note{
		{
			ID:    "note-1",
			Title: "Meeting Notes",
			BodyV2: &types.BodyV2{
				Markdown: "Discussed project timeline",
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:    "note-2",
			Title: "Follow-up",
			BodyV2: &types.BodyV2{
				Markdown: "Schedule next meeting",
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/rest/notes" {
			t.Errorf("expected path /rest/notes, got %s", r.URL.Path)
		}

		resp := types.NotesListResponse{
			TotalCount: 2,
			PageInfo:   &types.PageInfo{HasNextPage: false, EndCursor: "cursor-2"},
		}
		resp.Data.Notes = expectedNotes

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	result, err := client.ListNotes(context.Background(), nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("expected TotalCount 2, got %d", result.TotalCount)
	}
	if len(result.Data) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(result.Data))
	}
	if result.Data[0].ID != "note-1" {
		t.Errorf("expected first note ID 'note-1', got %s", result.Data[0].ID)
	}
	if result.Data[0].Title != "Meeting Notes" {
		t.Errorf("expected title 'Meeting Notes', got %s", result.Data[0].Title)
	}
	if result.Data[0].Body() != "Discussed project timeline" {
		t.Errorf("expected body 'Discussed project timeline', got %s", result.Data[0].Body())
	}
	if result.Data[1].Title != "Follow-up" {
		t.Errorf("expected second title 'Follow-up', got %s", result.Data[1].Title)
	}
	if result.PageInfo == nil {
		t.Fatal("expected PageInfo to be set")
	}
	if result.PageInfo.EndCursor != "cursor-2" {
		t.Errorf("expected EndCursor 'cursor-2', got %s", result.PageInfo.EndCursor)
	}
}

func TestClient_ListNotes_WithOptions(t *testing.T) {
	tests := []struct {
		name           string
		opts           *ListOptions
		expectedParams map[string]string
	}{
		{
			name: "with limit",
			opts: &ListOptions{Limit: 10},
			expectedParams: map[string]string{
				"limit": "10",
			},
		},
		{
			name: "with cursor",
			opts: &ListOptions{Cursor: "abc123"},
			expectedParams: map[string]string{
				"starting_after": "abc123",
			},
		},
		{
			name: "with sort and order",
			opts: &ListOptions{Sort: "createdAt", Order: "desc"},
			expectedParams: map[string]string{
				"order_by":           "createdAt",
				"order_by_direction": "desc",
			},
		},
		{
			name: "combined options",
			opts: &ListOptions{
				Limit:  25,
				Cursor: "xyz789",
				Sort:   "title",
				Order:  "asc",
			},
			expectedParams: map[string]string{
				"limit":              "25",
				"starting_after":     "xyz789",
				"order_by":           "title",
				"order_by_direction": "asc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedQuery string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedQuery = r.URL.RawQuery
				resp := types.NotesListResponse{TotalCount: 0}
				resp.Data.Notes = []types.Note{}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-token", false, WithNoRetry())
			_, err := client.ListNotes(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for param, expected := range tt.expectedParams {
				if !strings.Contains(receivedQuery, param+"="+expected) {
					t.Errorf("expected query to contain %s=%s, got query: %s", param, expected, receivedQuery)
				}
			}
		})
	}
}

func TestClient_GetNote(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	expectedNote := types.Note{
		ID:    "note-123",
		Title: "Important Note",
		BodyV2: &types.BodyV2{
			Markdown:  "This is the note content in markdown",
			Blocknote: `{"type":"doc","content":[]}`,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	t.Run("basic get", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/rest/notes/note-123" {
				t.Errorf("expected path /rest/notes/note-123, got %s", r.URL.Path)
			}

			resp := types.NoteResponse{}
			resp.Data.Note = expectedNote
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token", false, WithNoRetry())
		note, err := client.GetNote(context.Background(), "note-123")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if note.ID != "note-123" {
			t.Errorf("expected ID 'note-123', got %s", note.ID)
		}
		if note.Title != "Important Note" {
			t.Errorf("expected title 'Important Note', got %s", note.Title)
		}
		if note.Body() != "This is the note content in markdown" {
			t.Errorf("expected body content, got %s", note.Body())
		}
		if note.BodyV2.Blocknote != `{"type":"doc","content":[]}` {
			t.Errorf("expected blocknote content, got %s", note.BodyV2.Blocknote)
		}
	})

	t.Run("note not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":{"message":"Note not found"}}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token", false, WithNoRetry())
		_, err := client.GetNote(context.Background(), "non-existent")

		if err == nil {
			t.Fatal("expected error for non-existent note, got nil")
		}
	})

	t.Run("note with nil body", func(t *testing.T) {
		noteWithoutBody := types.Note{
			ID:        "note-no-body",
			Title:     "Empty Note",
			BodyV2:    nil,
			CreatedAt: now,
			UpdatedAt: now,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := types.NoteResponse{}
			resp.Data.Note = noteWithoutBody
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token", false, WithNoRetry())
		note, err := client.GetNote(context.Background(), "note-no-body")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if note.Body() != "" {
			t.Errorf("expected empty body for nil BodyV2, got %s", note.Body())
		}
	})
}

func TestClient_CreateNote(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	input := &CreateNoteInput{
		Title: "New Note",
		BodyV2: &BodyV2Input{
			Markdown: "This is the content of the new note",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/rest/notes" {
			t.Errorf("expected path /rest/notes, got %s", r.URL.Path)
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var receivedInput CreateNoteInput
		if err := json.Unmarshal(body, &receivedInput); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}
		if receivedInput.Title != "New Note" {
			t.Errorf("expected title 'New Note', got %s", receivedInput.Title)
		}
		if receivedInput.BodyV2 == nil || receivedInput.BodyV2.Markdown != "This is the content of the new note" {
			t.Errorf("expected body content, got %v", receivedInput.BodyV2)
		}

		// Return created note
		resp := types.CreateNoteResponse{}
		resp.Data.CreateNote = types.Note{
			ID:    "new-note-id",
			Title: input.Title,
			BodyV2: &types.BodyV2{
				Markdown: input.BodyV2.Markdown,
			},
			CreatedAt: now,
			UpdatedAt: now,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	note, err := client.CreateNote(context.Background(), input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if note.ID != "new-note-id" {
		t.Errorf("expected ID 'new-note-id', got %s", note.ID)
	}
	if note.Title != "New Note" {
		t.Errorf("expected title 'New Note', got %s", note.Title)
	}
	if note.Body() != "This is the content of the new note" {
		t.Errorf("expected body content, got %s", note.Body())
	}
}

func TestClient_CreateNote_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"code":"validation_error","message":"Invalid input"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	_, err := client.CreateNote(context.Background(), &CreateNoteInput{})

	if err == nil {
		t.Fatal("expected error for invalid input, got nil")
	}
}

func TestClient_UpdateNote(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	updatedTitle := "Updated Note Title"
	input := &UpdateNoteInput{
		Title: &updatedTitle,
		BodyV2: &BodyV2Input{
			Markdown: "Updated content",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/rest/notes/note-123" {
			t.Errorf("expected path /rest/notes/note-123, got %s", r.URL.Path)
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var receivedInput UpdateNoteInput
		if err := json.Unmarshal(body, &receivedInput); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}
		if receivedInput.Title == nil || *receivedInput.Title != "Updated Note Title" {
			t.Errorf("expected title 'Updated Note Title', got %v", receivedInput.Title)
		}

		// Return updated note
		resp := types.UpdateNoteResponse{}
		resp.Data.UpdateNote = types.Note{
			ID:    "note-123",
			Title: *input.Title,
			BodyV2: &types.BodyV2{
				Markdown: input.BodyV2.Markdown,
			},
			CreatedAt: now,
			UpdatedAt: now,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	note, err := client.UpdateNote(context.Background(), "note-123", input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if note.ID != "note-123" {
		t.Errorf("expected ID 'note-123', got %s", note.ID)
	}
	if note.Title != "Updated Note Title" {
		t.Errorf("expected title 'Updated Note Title', got %s", note.Title)
	}
	if note.Body() != "Updated content" {
		t.Errorf("expected body 'Updated content', got %s", note.Body())
	}
}

func TestClient_UpdateNote_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":{"message":"Note not found"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	title := "Updated"
	_, err := client.UpdateNote(context.Background(), "non-existent", &UpdateNoteInput{
		Title: &title,
	})

	if err == nil {
		t.Fatal("expected error for non-existent note, got nil")
	}
}

func TestClient_DeleteNote(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		var receivedMethod, receivedPath string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedMethod = r.Method
			receivedPath = r.URL.Path
			resp := types.DeleteNoteResponse{}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token", false, WithNoRetry())
		err := client.DeleteNote(context.Background(), "note-to-delete")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if receivedMethod != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", receivedMethod)
		}
		if receivedPath != "/rest/notes/note-to-delete" {
			t.Errorf("expected path /rest/notes/note-to-delete, got %s", receivedPath)
		}
	})

	t.Run("delete non-existent returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":{"message":"Note not found"}}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token", false, WithNoRetry())
		err := client.DeleteNote(context.Background(), "non-existent")

		if err == nil {
			t.Fatal("expected error for non-existent note, got nil")
		}
	})
}

func TestClient_ListNotes_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":{"message":"Unauthorized"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	_, err := client.ListNotes(context.Background(), nil)

	if err == nil {
		t.Fatal("expected error for unauthorized request, got nil")
	}
}
