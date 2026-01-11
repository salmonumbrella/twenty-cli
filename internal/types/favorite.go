package types

import "time"

// Favorite represents a favorite item in Twenty CRM
type Favorite struct {
	ID                string    `json:"id"`
	Position          float64   `json:"position"`
	WorkspaceMemberID string    `json:"workspaceMemberId"`
	CompanyID         *string   `json:"companyId,omitempty"`
	PersonID          *string   `json:"personId,omitempty"`
	OpportunityID     *string   `json:"opportunityId,omitempty"`
	TaskID            *string   `json:"taskId,omitempty"`
	NoteID            *string   `json:"noteId,omitempty"`
	ViewID            *string   `json:"viewId,omitempty"`
	WorkflowID        *string   `json:"workflowId,omitempty"`
	RocketID          *string   `json:"rocketId,omitempty"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
	DeletedAt         *string   `json:"deletedAt,omitempty"`
}

// FavoritesListResponse is the Twenty API response for listing favorites
type FavoritesListResponse struct {
	Data struct {
		Favorites []Favorite `json:"favorites"`
	} `json:"data"`
	TotalCount int       `json:"totalCount"`
	PageInfo   *PageInfo `json:"pageInfo,omitempty"`
}

// FavoriteResponse is the Twenty API response for a single favorite
type FavoriteResponse struct {
	Data struct {
		Favorite Favorite `json:"favorite"`
	} `json:"data"`
}
