package types

import "time"

// BodyV2 represents the rich text content structure used by Twenty API
type BodyV2 struct {
	Blocknote string `json:"blocknote,omitempty"`
	Markdown  string `json:"markdown,omitempty"`
}

type Note struct {
	ID        string    `json:"id"`
	Title     string    `json:"title,omitempty"`
	BodyV2    *BodyV2   `json:"bodyV2,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Body returns the markdown content from BodyV2 for convenience
func (n *Note) Body() string {
	if n.BodyV2 == nil {
		return ""
	}
	return n.BodyV2.Markdown
}

// NotesListResponse represents the API response for listing notes
type NotesListResponse struct {
	Data struct {
		Notes []Note `json:"notes"`
	} `json:"data"`
	TotalCount int       `json:"totalCount"`
	PageInfo   *PageInfo `json:"pageInfo,omitempty"`
}

// NoteResponse represents the API response for a single note
type NoteResponse struct {
	Data struct {
		Note Note `json:"note"`
	} `json:"data"`
}

// CreateNoteResponse represents the API response for creating a note
type CreateNoteResponse struct {
	Data struct {
		CreateNote Note `json:"createNote"`
	} `json:"data"`
}

// UpdateNoteResponse represents the API response for updating a note
type UpdateNoteResponse struct {
	Data struct {
		UpdateNote Note `json:"updateNote"`
	} `json:"data"`
}

// DeleteNoteResponse represents the API response for deleting a note
type DeleteNoteResponse struct {
	Data struct {
		DeleteNote Note `json:"deleteNote"`
	} `json:"data"`
}
