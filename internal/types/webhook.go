package types

import "time"

type Webhook struct {
	ID          string    `json:"id"`
	TargetURL   string    `json:"targetUrl"`
	Operation   string    `json:"operation"` // e.g., "*.created", "person.updated"
	Description string    `json:"description,omitempty"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
