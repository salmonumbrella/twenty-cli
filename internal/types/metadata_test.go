package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestObjectMetadataJSONParsing(t *testing.T) {
	jsonData := `{
		"id": "obj-123",
		"nameSingular": "person",
		"namePlural": "people",
		"labelSingular": "Person",
		"labelPlural": "People",
		"description": "A contact in the CRM",
		"icon": "IconUser",
		"isCustom": false,
		"isActive": true,
		"createdAt": "2024-01-15T10:30:00Z",
		"updatedAt": "2024-06-20T14:45:00Z"
	}`

	var obj ObjectMetadata
	err := json.Unmarshal([]byte(jsonData), &obj)
	if err != nil {
		t.Fatalf("failed to unmarshal ObjectMetadata: %v", err)
	}

	if obj.ID != "obj-123" {
		t.Errorf("expected ID='obj-123', got %q", obj.ID)
	}
	if obj.NameSingular != "person" {
		t.Errorf("expected NameSingular='person', got %q", obj.NameSingular)
	}
	if obj.NamePlural != "people" {
		t.Errorf("expected NamePlural='people', got %q", obj.NamePlural)
	}
	if obj.LabelSingular != "Person" {
		t.Errorf("expected LabelSingular='Person', got %q", obj.LabelSingular)
	}
	if obj.LabelPlural != "People" {
		t.Errorf("expected LabelPlural='People', got %q", obj.LabelPlural)
	}
	if obj.Description != "A contact in the CRM" {
		t.Errorf("expected Description='A contact in the CRM', got %q", obj.Description)
	}
	if obj.Icon != "IconUser" {
		t.Errorf("expected Icon='IconUser', got %q", obj.Icon)
	}
	if obj.IsCustom != false {
		t.Errorf("expected IsCustom=false, got %v", obj.IsCustom)
	}
	if obj.IsActive != true {
		t.Errorf("expected IsActive=true, got %v", obj.IsActive)
	}
	if obj.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if obj.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestObjectMetadataWithFields(t *testing.T) {
	jsonData := `{
		"id": "obj-456",
		"nameSingular": "company",
		"namePlural": "companies",
		"labelSingular": "Company",
		"labelPlural": "Companies",
		"description": "A company entity",
		"icon": "IconBuilding",
		"isCustom": false,
		"isActive": true,
		"fields": [
			{
				"id": "field-1",
				"name": "name",
				"label": "Name",
				"type": "TEXT",
				"description": "Company name",
				"isCustom": false,
				"isActive": true
			},
			{
				"id": "field-2",
				"name": "website",
				"label": "Website",
				"type": "LINK",
				"description": "Company website URL",
				"isCustom": false,
				"isActive": true
			}
		],
		"createdAt": "2024-01-15T10:30:00Z",
		"updatedAt": "2024-06-20T14:45:00Z"
	}`

	var obj ObjectMetadata
	err := json.Unmarshal([]byte(jsonData), &obj)
	if err != nil {
		t.Fatalf("failed to unmarshal ObjectMetadata with fields: %v", err)
	}

	if len(obj.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(obj.Fields))
	}

	// Check first field
	if obj.Fields[0].ID != "field-1" {
		t.Errorf("expected field[0].ID='field-1', got %q", obj.Fields[0].ID)
	}
	if obj.Fields[0].Name != "name" {
		t.Errorf("expected field[0].Name='name', got %q", obj.Fields[0].Name)
	}
	if obj.Fields[0].Label != "Name" {
		t.Errorf("expected field[0].Label='Name', got %q", obj.Fields[0].Label)
	}
	if obj.Fields[0].Type != "TEXT" {
		t.Errorf("expected field[0].Type='TEXT', got %q", obj.Fields[0].Type)
	}

	// Check second field
	if obj.Fields[1].Type != "LINK" {
		t.Errorf("expected field[1].Type='LINK', got %q", obj.Fields[1].Type)
	}
}

func TestFieldMetadataJSONParsing(t *testing.T) {
	jsonData := `{
		"id": "field-abc",
		"name": "customField",
		"label": "Custom Field",
		"type": "NUMBER",
		"description": "A custom numeric field",
		"isCustom": true,
		"isActive": true
	}`

	var field FieldMetadata
	err := json.Unmarshal([]byte(jsonData), &field)
	if err != nil {
		t.Fatalf("failed to unmarshal FieldMetadata: %v", err)
	}

	if field.ID != "field-abc" {
		t.Errorf("expected ID='field-abc', got %q", field.ID)
	}
	if field.Name != "customField" {
		t.Errorf("expected Name='customField', got %q", field.Name)
	}
	if field.Label != "Custom Field" {
		t.Errorf("expected Label='Custom Field', got %q", field.Label)
	}
	if field.Type != "NUMBER" {
		t.Errorf("expected Type='NUMBER', got %q", field.Type)
	}
	if field.Description != "A custom numeric field" {
		t.Errorf("expected Description='A custom numeric field', got %q", field.Description)
	}
	if field.IsCustom != true {
		t.Errorf("expected IsCustom=true, got %v", field.IsCustom)
	}
	if field.IsActive != true {
		t.Errorf("expected IsActive=true, got %v", field.IsActive)
	}
}

func TestObjectMetadataJSONSerialize(t *testing.T) {
	obj := ObjectMetadata{
		ID:            "obj-789",
		NameSingular:  "task",
		NamePlural:    "tasks",
		LabelSingular: "Task",
		LabelPlural:   "Tasks",
		Description:   "A task to complete",
		Icon:          "IconCheckbox",
		IsCustom:      false,
		IsActive:      true,
		Fields: []FieldMetadata{
			{
				ID:          "field-title",
				Name:        "title",
				Label:       "Title",
				Type:        "TEXT",
				Description: "Task title",
				IsCustom:    false,
				IsActive:    true,
			},
		},
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 6, 20, 14, 45, 0, 0, time.UTC),
	}

	data, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("failed to marshal ObjectMetadata: %v", err)
	}

	// Verify we can unmarshal it back
	var parsed ObjectMetadata
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal serialized ObjectMetadata: %v", err)
	}

	if parsed.ID != obj.ID {
		t.Errorf("round-trip failed: expected ID=%q, got %q", obj.ID, parsed.ID)
	}
	if len(parsed.Fields) != 1 {
		t.Errorf("round-trip failed: expected 1 field, got %d", len(parsed.Fields))
	}
}

func TestObjectMetadataEmptyFields(t *testing.T) {
	jsonData := `{
		"id": "obj-empty",
		"nameSingular": "empty",
		"namePlural": "empties",
		"labelSingular": "Empty",
		"labelPlural": "Empties",
		"description": "",
		"icon": "",
		"isCustom": true,
		"isActive": false,
		"createdAt": "2024-01-15T10:30:00Z",
		"updatedAt": "2024-06-20T14:45:00Z"
	}`

	var obj ObjectMetadata
	err := json.Unmarshal([]byte(jsonData), &obj)
	if err != nil {
		t.Fatalf("failed to unmarshal ObjectMetadata: %v", err)
	}

	if obj.Fields != nil && len(obj.Fields) > 0 {
		t.Errorf("expected empty Fields slice, got %d fields", len(obj.Fields))
	}
	if obj.IsCustom != true {
		t.Errorf("expected IsCustom=true, got %v", obj.IsCustom)
	}
	if obj.IsActive != false {
		t.Errorf("expected IsActive=false, got %v", obj.IsActive)
	}
}
