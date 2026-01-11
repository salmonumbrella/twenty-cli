package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNote_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"id": "note-123",
		"title": "Meeting Notes",
		"bodyV2": {
			"blocknote": "[{\"type\":\"paragraph\",\"content\":[{\"type\":\"text\",\"text\":\"Discussion points\"}]}]",
			"markdown": "Discussion points"
		},
		"createdAt": "2024-01-15T10:30:00Z",
		"updatedAt": "2024-06-20T14:45:00Z"
	}`

	var note Note
	err := json.Unmarshal([]byte(jsonData), &note)
	if err != nil {
		t.Fatalf("failed to unmarshal Note: %v", err)
	}

	if note.ID != "note-123" {
		t.Errorf("expected ID='note-123', got %q", note.ID)
	}
	if note.Title != "Meeting Notes" {
		t.Errorf("expected Title='Meeting Notes', got %q", note.Title)
	}
	if note.BodyV2 == nil {
		t.Fatal("expected BodyV2 to be set")
	}
	if note.BodyV2.Markdown != "Discussion points" {
		t.Errorf("expected BodyV2.Markdown='Discussion points', got %q", note.BodyV2.Markdown)
	}
	if note.BodyV2.Blocknote == "" {
		t.Error("expected BodyV2.Blocknote to be set")
	}
	if note.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if note.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestNote_JSONMarshal(t *testing.T) {
	note := Note{
		ID:    "note-789",
		Title: "Project Notes",
		BodyV2: &BodyV2{
			Blocknote: "[{\"type\":\"paragraph\"}]",
			Markdown:  "Project details",
		},
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 6, 20, 14, 45, 0, 0, time.UTC),
	}

	data, err := json.Marshal(note)
	if err != nil {
		t.Fatalf("failed to marshal Note: %v", err)
	}

	// Round-trip: unmarshal back
	var parsed Note
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal serialized Note: %v", err)
	}

	if parsed.ID != note.ID {
		t.Errorf("round-trip failed: expected ID=%q, got %q", note.ID, parsed.ID)
	}
	if parsed.Title != note.Title {
		t.Errorf("round-trip failed: expected Title=%q, got %q", note.Title, parsed.Title)
	}
	if parsed.BodyV2 == nil {
		t.Fatal("round-trip failed: expected BodyV2 to be set")
	}
	if parsed.BodyV2.Markdown != note.BodyV2.Markdown {
		t.Errorf("round-trip failed: expected Markdown=%q, got %q", note.BodyV2.Markdown, parsed.BodyV2.Markdown)
	}
}

func TestNote_Body(t *testing.T) {
	tests := []struct {
		name     string
		note     Note
		expected string
	}{
		{
			name: "with bodyV2",
			note: Note{
				ID:    "note-1",
				Title: "Test",
				BodyV2: &BodyV2{
					Markdown: "Hello world",
				},
			},
			expected: "Hello world",
		},
		{
			name: "nil bodyV2",
			note: Note{
				ID:     "note-2",
				Title:  "Test",
				BodyV2: nil,
			},
			expected: "",
		},
		{
			name: "empty markdown",
			note: Note{
				ID:    "note-3",
				Title: "Test",
				BodyV2: &BodyV2{
					Markdown: "",
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.note.Body()
			if result != tt.expected {
				t.Errorf("expected Body()=%q, got %q", tt.expected, result)
			}
		})
	}
}

func TestNote_MinimalFields(t *testing.T) {
	jsonData := `{
		"id": "note-minimal",
		"createdAt": "2024-01-15T10:30:00Z",
		"updatedAt": "2024-06-20T14:45:00Z"
	}`

	var note Note
	err := json.Unmarshal([]byte(jsonData), &note)
	if err != nil {
		t.Fatalf("failed to unmarshal Note: %v", err)
	}

	if note.ID != "note-minimal" {
		t.Errorf("expected ID='note-minimal', got %q", note.ID)
	}
	if note.Title != "" {
		t.Errorf("expected Title to be empty, got %q", note.Title)
	}
	if note.BodyV2 != nil {
		t.Errorf("expected BodyV2 to be nil, got %v", note.BodyV2)
	}
}

func TestBodyV2_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"blocknote": "[{\"type\":\"heading\",\"level\":1}]",
		"markdown": "# Heading"
	}`

	var body BodyV2
	err := json.Unmarshal([]byte(jsonData), &body)
	if err != nil {
		t.Fatalf("failed to unmarshal BodyV2: %v", err)
	}

	if body.Blocknote != "[{\"type\":\"heading\",\"level\":1}]" {
		t.Errorf("expected Blocknote to match, got %q", body.Blocknote)
	}
	if body.Markdown != "# Heading" {
		t.Errorf("expected Markdown='# Heading', got %q", body.Markdown)
	}
}

func TestNotesListResponse_JSONParsing(t *testing.T) {
	jsonData := `{
		"data": {
			"notes": [
				{
					"id": "n1",
					"title": "Note One",
					"bodyV2": {
						"markdown": "Content one"
					},
					"createdAt": "2024-01-15T10:30:00Z",
					"updatedAt": "2024-06-20T14:45:00Z"
				},
				{
					"id": "n2",
					"title": "Note Two",
					"createdAt": "2024-02-10T08:00:00Z",
					"updatedAt": "2024-05-15T12:00:00Z"
				}
			]
		},
		"totalCount": 2,
		"pageInfo": {
			"hasNextPage": false,
			"endCursor": ""
		}
	}`

	var resp NotesListResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal NotesListResponse: %v", err)
	}

	if len(resp.Data.Notes) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(resp.Data.Notes))
	}
	if resp.TotalCount != 2 {
		t.Errorf("expected TotalCount=2, got %d", resp.TotalCount)
	}
	if resp.Data.Notes[0].ID != "n1" {
		t.Errorf("expected first note ID='n1', got %q", resp.Data.Notes[0].ID)
	}
	if resp.Data.Notes[0].Title != "Note One" {
		t.Errorf("expected first note Title='Note One', got %q", resp.Data.Notes[0].Title)
	}
	if resp.Data.Notes[0].BodyV2 == nil || resp.Data.Notes[0].BodyV2.Markdown != "Content one" {
		t.Error("expected first note to have BodyV2 with Markdown='Content one'")
	}
	if resp.Data.Notes[1].BodyV2 != nil {
		t.Error("expected second note BodyV2 to be nil")
	}
	if resp.PageInfo == nil {
		t.Fatal("expected PageInfo to be set")
	}
	if resp.PageInfo.HasNextPage {
		t.Error("expected HasNextPage=false")
	}
}

func TestNoteResponse_JSONParsing(t *testing.T) {
	jsonData := `{
		"data": {
			"note": {
				"id": "note-single",
				"title": "Single Note",
				"bodyV2": {
					"markdown": "Note content"
				},
				"createdAt": "2024-01-15T10:30:00Z",
				"updatedAt": "2024-06-20T14:45:00Z"
			}
		}
	}`

	var resp NoteResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal NoteResponse: %v", err)
	}

	if resp.Data.Note.ID != "note-single" {
		t.Errorf("expected ID='note-single', got %q", resp.Data.Note.ID)
	}
	if resp.Data.Note.Title != "Single Note" {
		t.Errorf("expected Title='Single Note', got %q", resp.Data.Note.Title)
	}
}

func TestCreateNoteResponse_JSONParsing(t *testing.T) {
	jsonData := `{
		"data": {
			"createNote": {
				"id": "new-note",
				"title": "New Note",
				"createdAt": "2024-01-15T10:30:00Z",
				"updatedAt": "2024-06-20T14:45:00Z"
			}
		}
	}`

	var resp CreateNoteResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal CreateNoteResponse: %v", err)
	}

	if resp.Data.CreateNote.ID != "new-note" {
		t.Errorf("expected ID='new-note', got %q", resp.Data.CreateNote.ID)
	}
	if resp.Data.CreateNote.Title != "New Note" {
		t.Errorf("expected Title='New Note', got %q", resp.Data.CreateNote.Title)
	}
}

func TestUpdateNoteResponse_JSONParsing(t *testing.T) {
	jsonData := `{
		"data": {
			"updateNote": {
				"id": "updated-note",
				"title": "Updated Note",
				"bodyV2": {
					"markdown": "Updated content"
				},
				"createdAt": "2024-01-15T10:30:00Z",
				"updatedAt": "2024-06-20T14:45:00Z"
			}
		}
	}`

	var resp UpdateNoteResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal UpdateNoteResponse: %v", err)
	}

	if resp.Data.UpdateNote.ID != "updated-note" {
		t.Errorf("expected ID='updated-note', got %q", resp.Data.UpdateNote.ID)
	}
	if resp.Data.UpdateNote.BodyV2 == nil || resp.Data.UpdateNote.BodyV2.Markdown != "Updated content" {
		t.Error("expected BodyV2.Markdown='Updated content'")
	}
}

func TestDeleteNoteResponse_JSONParsing(t *testing.T) {
	jsonData := `{
		"data": {
			"deleteNote": {
				"id": "deleted-note",
				"title": "Deleted Note",
				"createdAt": "2024-01-15T10:30:00Z",
				"updatedAt": "2024-06-20T14:45:00Z"
			}
		}
	}`

	var resp DeleteNoteResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal DeleteNoteResponse: %v", err)
	}

	if resp.Data.DeleteNote.ID != "deleted-note" {
		t.Errorf("expected ID='deleted-note', got %q", resp.Data.DeleteNote.ID)
	}
}
