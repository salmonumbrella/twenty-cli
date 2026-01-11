package types

import "time"

// ObjectMetadata represents metadata about an object type in Twenty CRM
type ObjectMetadata struct {
	ID            string          `json:"id"`
	NameSingular  string          `json:"nameSingular"`
	NamePlural    string          `json:"namePlural"`
	LabelSingular string          `json:"labelSingular"`
	LabelPlural   string          `json:"labelPlural"`
	Description   string          `json:"description"`
	Icon          string          `json:"icon"`
	IsCustom      bool            `json:"isCustom"`
	IsActive      bool            `json:"isActive"`
	Fields        []FieldMetadata `json:"fields,omitempty"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

// FieldMetadata represents metadata about a field within an object type
type FieldMetadata struct {
	ID               string    `json:"id"`
	ObjectMetadataId string    `json:"objectMetadataId,omitempty"`
	Name             string    `json:"name"`
	Label            string    `json:"label"`
	Type             string    `json:"type"`
	Description      string    `json:"description"`
	IsCustom         bool      `json:"isCustom"`
	IsActive         bool      `json:"isActive"`
	IsNullable       bool      `json:"isNullable"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// ObjectsListResponse represents the API response for listing objects
type ObjectsListResponse struct {
	Data struct {
		Objects []ObjectMetadata `json:"objects"`
	} `json:"data"`
	TotalCount int       `json:"totalCount"`
	PageInfo   *PageInfo `json:"pageInfo,omitempty"`
}

// ObjectResponse represents the API response for a single object
type ObjectResponse struct {
	Data struct {
		Object ObjectMetadata `json:"object"`
	} `json:"data"`
}

// FieldsListResponse represents the API response for listing fields
type FieldsListResponse struct {
	Data struct {
		Fields []FieldMetadata `json:"fields"`
	} `json:"data"`
	TotalCount int       `json:"totalCount"`
	PageInfo   *PageInfo `json:"pageInfo,omitempty"`
}

// FieldResponse represents the API response for a single field
type FieldResponse struct {
	Data struct {
		Field FieldMetadata `json:"field"`
	} `json:"data"`
}
