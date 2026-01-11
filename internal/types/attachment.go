package types

import "time"

// Attachment represents a file attachment in Twenty CRM
type Attachment struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	FullPath   string    `json:"fullPath"`
	Type       string    `json:"type"`
	CompanyID  *string   `json:"companyId,omitempty"`
	PersonID   *string   `json:"personId,omitempty"`
	ActivityID *string   `json:"activityId,omitempty"`
	TaskID     *string   `json:"taskId,omitempty"`
	NoteID     *string   `json:"noteId,omitempty"`
	AuthorID   *string   `json:"authorId,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	DeletedAt  *string   `json:"deletedAt,omitempty"`
}

// AttachmentsListResponse is the Twenty API response for listing attachments
type AttachmentsListResponse struct {
	Data struct {
		Attachments []Attachment `json:"attachments"`
	} `json:"data"`
	TotalCount int       `json:"totalCount"`
	PageInfo   *PageInfo `json:"pageInfo,omitempty"`
}

// AttachmentResponse is the Twenty API response for a single attachment
type AttachmentResponse struct {
	Data struct {
		Attachment Attachment `json:"attachment"`
	} `json:"data"`
}
