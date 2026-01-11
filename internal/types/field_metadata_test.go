package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestFieldMetadataExtendedJSONParsing(t *testing.T) {
	// This tests the extended FieldMetadata with ObjectMetadataId, IsNullable, CreatedAt, UpdatedAt
	jsonData := `{
		"id": "field-123",
		"objectMetadataId": "obj-person-456",
		"name": "firstName",
		"label": "First Name",
		"type": "TEXT",
		"description": "The person's first name",
		"isCustom": false,
		"isActive": true,
		"isNullable": true,
		"createdAt": "2024-01-15T10:30:00Z",
		"updatedAt": "2024-06-20T14:45:00Z"
	}`

	var field FieldMetadata
	err := json.Unmarshal([]byte(jsonData), &field)
	if err != nil {
		t.Fatalf("failed to unmarshal FieldMetadata: %v", err)
	}

	if field.ID != "field-123" {
		t.Errorf("expected ID='field-123', got %q", field.ID)
	}
	if field.ObjectMetadataId != "obj-person-456" {
		t.Errorf("expected ObjectMetadataId='obj-person-456', got %q", field.ObjectMetadataId)
	}
	if field.Name != "firstName" {
		t.Errorf("expected Name='firstName', got %q", field.Name)
	}
	if field.Label != "First Name" {
		t.Errorf("expected Label='First Name', got %q", field.Label)
	}
	if field.Type != "TEXT" {
		t.Errorf("expected Type='TEXT', got %q", field.Type)
	}
	if field.Description != "The person's first name" {
		t.Errorf("expected Description='The person's first name', got %q", field.Description)
	}
	if field.IsCustom != false {
		t.Errorf("expected IsCustom=false, got %v", field.IsCustom)
	}
	if field.IsActive != true {
		t.Errorf("expected IsActive=true, got %v", field.IsActive)
	}
	if field.IsNullable != true {
		t.Errorf("expected IsNullable=true, got %v", field.IsNullable)
	}
	if field.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	expectedCreatedAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	if !field.CreatedAt.Equal(expectedCreatedAt) {
		t.Errorf("expected CreatedAt=%v, got %v", expectedCreatedAt, field.CreatedAt)
	}
	if field.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
	expectedUpdatedAt := time.Date(2024, 6, 20, 14, 45, 0, 0, time.UTC)
	if !field.UpdatedAt.Equal(expectedUpdatedAt) {
		t.Errorf("expected UpdatedAt=%v, got %v", expectedUpdatedAt, field.UpdatedAt)
	}
}

func TestFieldMetadataExtendedSerialize(t *testing.T) {
	field := FieldMetadata{
		ID:               "field-789",
		ObjectMetadataId: "obj-company-123",
		Name:             "revenue",
		Label:            "Revenue",
		Type:             "NUMBER",
		Description:      "Annual revenue",
		IsCustom:         true,
		IsActive:         true,
		IsNullable:       false,
		CreatedAt:        time.Date(2024, 2, 10, 8, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2024, 7, 15, 12, 30, 0, 0, time.UTC),
	}

	data, err := json.Marshal(field)
	if err != nil {
		t.Fatalf("failed to marshal FieldMetadata: %v", err)
	}

	// Verify we can unmarshal it back
	var parsed FieldMetadata
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal serialized FieldMetadata: %v", err)
	}

	if parsed.ID != field.ID {
		t.Errorf("round-trip failed: expected ID=%q, got %q", field.ID, parsed.ID)
	}
	if parsed.ObjectMetadataId != field.ObjectMetadataId {
		t.Errorf("round-trip failed: expected ObjectMetadataId=%q, got %q", field.ObjectMetadataId, parsed.ObjectMetadataId)
	}
	if parsed.IsNullable != field.IsNullable {
		t.Errorf("round-trip failed: expected IsNullable=%v, got %v", field.IsNullable, parsed.IsNullable)
	}
	if !parsed.CreatedAt.Equal(field.CreatedAt) {
		t.Errorf("round-trip failed: expected CreatedAt=%v, got %v", field.CreatedAt, parsed.CreatedAt)
	}
	if !parsed.UpdatedAt.Equal(field.UpdatedAt) {
		t.Errorf("round-trip failed: expected UpdatedAt=%v, got %v", field.UpdatedAt, parsed.UpdatedAt)
	}
}
