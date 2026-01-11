package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func TestListObjects(t *testing.T) {
	// Mock API response - Twenty API returns {"data":{"objects":[...]}} format
	response := types.ObjectsListResponse{
		TotalCount: 2,
	}
	response.Data.Objects = []types.ObjectMetadata{
		{
			ID:            "obj-1",
			NameSingular:  "person",
			NamePlural:    "people",
			LabelSingular: "Person",
			LabelPlural:   "People",
			Description:   "A contact",
			Icon:          "IconUser",
			IsCustom:      false,
			IsActive:      true,
		},
		{
			ID:            "obj-2",
			NameSingular:  "company",
			NamePlural:    "companies",
			LabelSingular: "Company",
			LabelPlural:   "Companies",
			Description:   "An organization",
			Icon:          "IconBuilding",
			IsCustom:      false,
			IsActive:      true,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify correct endpoint is called
		if r.URL.Path != "/rest/metadata/objects" {
			t.Errorf("expected path /rest/metadata/objects, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		// Verify auth header
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-token" {
			t.Errorf("expected Bearer test-token, got %s", auth)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	result, err := client.ListObjects(context.Background())
	if err != nil {
		t.Fatalf("ListObjects failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 objects, got %d", len(result))
	}

	if result[0].NameSingular != "person" {
		t.Errorf("expected first object NameSingular='person', got %q", result[0].NameSingular)
	}
	if result[1].NameSingular != "company" {
		t.Errorf("expected second object NameSingular='company', got %q", result[1].NameSingular)
	}
}

func TestListObjectsEmpty(t *testing.T) {
	response := types.ObjectsListResponse{
		TotalCount: 0,
	}
	response.Data.Objects = []types.ObjectMetadata{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	result, err := client.ListObjects(context.Background())
	if err != nil {
		t.Fatalf("ListObjects failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 objects, got %d", len(result))
	}
}

func TestListObjectsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "bad-token", false)
	_, err := client.ListObjects(context.Background())
	if err == nil {
		t.Fatal("expected error for unauthorized request")
	}
}

func TestGetObject(t *testing.T) {
	// GetObject with a name (not UUID) first lists all objects, then fetches by ID
	listResponse := types.ObjectsListResponse{}
	listResponse.Data.Objects = []types.ObjectMetadata{
		{
			ID:           "obj-person",
			NameSingular: "person",
			NamePlural:   "people",
		},
	}

	// Full object response when fetching by ID
	objectResponse := types.ObjectResponse{}
	objectResponse.Data.Object = types.ObjectMetadata{
		ID:            "obj-person",
		NameSingular:  "person",
		NamePlural:    "people",
		LabelSingular: "Person",
		LabelPlural:   "People",
		Description:   "A contact in the CRM",
		Icon:          "IconUser",
		IsCustom:      false,
		IsActive:      true,
		Fields: []types.FieldMetadata{
			{
				ID:          "field-1",
				Name:        "firstName",
				Label:       "First Name",
				Type:        "TEXT",
				Description: "Person's first name",
				IsCustom:    false,
				IsActive:    true,
			},
			{
				ID:          "field-2",
				Name:        "lastName",
				Label:       "Last Name",
				Type:        "TEXT",
				Description: "Person's last name",
				IsCustom:    false,
				IsActive:    true,
			},
			{
				ID:          "field-3",
				Name:        "email",
				Label:       "Email",
				Type:        "EMAIL",
				Description: "Email address",
				IsCustom:    false,
				IsActive:    true,
			},
		},
	}

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		// Verify auth header
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-token" {
			t.Errorf("expected Bearer test-token, got %s", auth)
		}

		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/rest/metadata/objects" {
			// First call: list all objects
			json.NewEncoder(w).Encode(listResponse)
		} else if r.URL.Path == "/rest/metadata/objects/obj-person" {
			// Second call: get object by ID
			json.NewEncoder(w).Encode(objectResponse)
		} else {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	result, err := client.GetObject(context.Background(), "person")
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}

	if requestCount != 2 {
		t.Errorf("expected 2 requests (list + get), got %d", requestCount)
	}
	if result.NameSingular != "person" {
		t.Errorf("expected NameSingular='person', got %q", result.NameSingular)
	}
	if len(result.Fields) != 3 {
		t.Errorf("expected 3 fields, got %d", len(result.Fields))
	}
	if result.Fields[0].Name != "firstName" {
		t.Errorf("expected first field Name='firstName', got %q", result.Fields[0].Name)
	}
	if result.Fields[2].Type != "EMAIL" {
		t.Errorf("expected third field Type='EMAIL', got %q", result.Fields[2].Type)
	}
}

func TestGetObjectNotFound(t *testing.T) {
	// When searching by name, GetObject lists all objects first
	// If the object is not in the list, it returns "object not found"
	listResponse := types.ObjectsListResponse{}
	listResponse.Data.Objects = []types.ObjectMetadata{
		{ID: "obj-1", NameSingular: "person", NamePlural: "people"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/rest/metadata/objects" {
			json.NewEncoder(w).Encode(listResponse)
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "object not found"}`))
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	_, err := client.GetObject(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found object")
	}
}

func TestGetObjectWithSpecialChars(t *testing.T) {
	// GetObject with a custom object name first lists all objects, then fetches by ID
	listResponse := types.ObjectsListResponse{}
	listResponse.Data.Objects = []types.ObjectMetadata{
		{
			ID:           "obj-custom",
			NameSingular: "myCustomObject",
			NamePlural:   "myCustomObjects",
		},
	}

	objectResponse := types.ObjectResponse{}
	objectResponse.Data.Object = types.ObjectMetadata{
		ID:           "obj-custom",
		NameSingular: "myCustomObject",
		NamePlural:   "myCustomObjects",
		IsCustom:     true,
		IsActive:     true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/rest/metadata/objects" {
			json.NewEncoder(w).Encode(listResponse)
		} else if r.URL.Path == "/rest/metadata/objects/obj-custom" {
			json.NewEncoder(w).Encode(objectResponse)
		} else {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	result, err := client.GetObject(context.Background(), "myCustomObject")
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}

	if result.NameSingular != "myCustomObject" {
		t.Errorf("expected NameSingular='myCustomObject', got %q", result.NameSingular)
	}
}

func TestListFields(t *testing.T) {
	// Mock API response - Twenty API returns {"data":{"fields":[...]}} format
	response := types.FieldsListResponse{
		TotalCount: 3,
	}
	response.Data.Fields = []types.FieldMetadata{
		{
			ID:               "field-1",
			ObjectMetadataId: "obj-person",
			Name:             "firstName",
			Label:            "First Name",
			Type:             "TEXT",
			Description:      "Person's first name",
			IsCustom:         false,
			IsActive:         true,
			IsNullable:       false,
		},
		{
			ID:               "field-2",
			ObjectMetadataId: "obj-person",
			Name:             "lastName",
			Label:            "Last Name",
			Type:             "TEXT",
			Description:      "Person's last name",
			IsCustom:         false,
			IsActive:         true,
			IsNullable:       false,
		},
		{
			ID:               "field-3",
			ObjectMetadataId: "obj-company",
			Name:             "name",
			Label:            "Name",
			Type:             "TEXT",
			Description:      "Company name",
			IsCustom:         false,
			IsActive:         true,
			IsNullable:       false,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify correct endpoint is called
		if r.URL.Path != "/rest/metadata/fields" {
			t.Errorf("expected path /rest/metadata/fields, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		// Verify auth header
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-token" {
			t.Errorf("expected Bearer test-token, got %s", auth)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	result, err := client.ListFields(context.Background())
	if err != nil {
		t.Fatalf("ListFields failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 fields, got %d", len(result))
	}

	if result[0].Name != "firstName" {
		t.Errorf("expected first field Name='firstName', got %q", result[0].Name)
	}
	if result[0].ObjectMetadataId != "obj-person" {
		t.Errorf("expected first field ObjectMetadataId='obj-person', got %q", result[0].ObjectMetadataId)
	}
	if result[2].ObjectMetadataId != "obj-company" {
		t.Errorf("expected third field ObjectMetadataId='obj-company', got %q", result[2].ObjectMetadataId)
	}
}

func TestListFieldsEmpty(t *testing.T) {
	response := types.FieldsListResponse{
		TotalCount: 0,
	}
	response.Data.Fields = []types.FieldMetadata{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	result, err := client.ListFields(context.Background())
	if err != nil {
		t.Fatalf("ListFields failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 fields, got %d", len(result))
	}
}

func TestListFieldsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "bad-token", false)
	_, err := client.ListFields(context.Background())
	if err == nil {
		t.Fatal("expected error for unauthorized request")
	}
}

func TestGetField(t *testing.T) {
	// Twenty API returns {"data":{"field":{...}}} format
	response := types.FieldResponse{}
	response.Data.Field = types.FieldMetadata{
		ID:               "field-123",
		ObjectMetadataId: "obj-person",
		Name:             "firstName",
		Label:            "First Name",
		Type:             "TEXT",
		Description:      "Person's first name",
		IsCustom:         false,
		IsActive:         true,
		IsNullable:       false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify correct endpoint is called
		if r.URL.Path != "/rest/metadata/fields/field-123" {
			t.Errorf("expected path /rest/metadata/fields/field-123, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		// Verify auth header
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-token" {
			t.Errorf("expected Bearer test-token, got %s", auth)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	result, err := client.GetField(context.Background(), "field-123")
	if err != nil {
		t.Fatalf("GetField failed: %v", err)
	}

	if result.ID != "field-123" {
		t.Errorf("expected ID='field-123', got %q", result.ID)
	}
	if result.Name != "firstName" {
		t.Errorf("expected Name='firstName', got %q", result.Name)
	}
	if result.ObjectMetadataId != "obj-person" {
		t.Errorf("expected ObjectMetadataId='obj-person', got %q", result.ObjectMetadataId)
	}
	if result.Type != "TEXT" {
		t.Errorf("expected Type='TEXT', got %q", result.Type)
	}
}

func TestGetFieldNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "field not found"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	_, err := client.GetField(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found field")
	}
}

func TestGetObjectByUUID(t *testing.T) {
	// Test direct UUID lookup (36-char format with dashes)
	objectResponse := types.ObjectResponse{}
	objectResponse.Data.Object = types.ObjectMetadata{
		ID:            "12345678-1234-1234-1234-123456789012",
		NameSingular:  "customObject",
		NamePlural:    "customObjects",
		LabelSingular: "Custom Object",
		LabelPlural:   "Custom Objects",
		IsCustom:      true,
		IsActive:      true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should call the direct endpoint with UUID, not list first
		if r.URL.Path != "/rest/metadata/objects/12345678-1234-1234-1234-123456789012" {
			t.Errorf("expected direct path with UUID, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(objectResponse)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	result, err := client.GetObject(context.Background(), "12345678-1234-1234-1234-123456789012")
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}

	if result.ID != "12345678-1234-1234-1234-123456789012" {
		t.Errorf("expected ID='12345678-1234-1234-1234-123456789012', got %q", result.ID)
	}
	if result.NameSingular != "customObject" {
		t.Errorf("expected NameSingular='customObject', got %q", result.NameSingular)
	}
}

func TestCreateObject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/rest/metadata/objects" {
			t.Errorf("expected path /rest/metadata/objects, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"new-obj-id","nameSingular":"newObject"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	result, err := client.CreateObject(context.Background(), map[string]interface{}{
		"nameSingular": "newObject",
		"namePlural":   "newObjects",
	})
	if err != nil {
		t.Fatalf("CreateObject failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

func TestUpdateObject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/rest/metadata/objects/obj-123" {
			t.Errorf("expected path /rest/metadata/objects/obj-123, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"obj-123","nameSingular":"updatedObject"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	result, err := client.UpdateObject(context.Background(), "obj-123", map[string]interface{}{
		"description": "Updated description",
	})
	if err != nil {
		t.Fatalf("UpdateObject failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

func TestDeleteObject(t *testing.T) {
	var receivedMethod, receivedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	err := client.DeleteObject(context.Background(), "obj-to-delete")
	if err != nil {
		t.Fatalf("DeleteObject failed: %v", err)
	}
	if receivedMethod != http.MethodDelete {
		t.Errorf("expected DELETE, got %s", receivedMethod)
	}
	if receivedPath != "/rest/metadata/objects/obj-to-delete" {
		t.Errorf("expected path /rest/metadata/objects/obj-to-delete, got %s", receivedPath)
	}
}

func TestCreateField(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/rest/metadata/fields" {
			t.Errorf("expected path /rest/metadata/fields, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"new-field-id","name":"newField"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	result, err := client.CreateField(context.Background(), map[string]interface{}{
		"name":             "newField",
		"label":            "New Field",
		"type":             "TEXT",
		"objectMetadataId": "obj-123",
	})
	if err != nil {
		t.Fatalf("CreateField failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

func TestUpdateField(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/rest/metadata/fields/field-123" {
			t.Errorf("expected path /rest/metadata/fields/field-123, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"field-123","name":"updatedField"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	result, err := client.UpdateField(context.Background(), "field-123", map[string]interface{}{
		"label": "Updated Label",
	})
	if err != nil {
		t.Fatalf("UpdateField failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

func TestDeleteField(t *testing.T) {
	var receivedMethod, receivedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	err := client.DeleteField(context.Background(), "field-to-delete")
	if err != nil {
		t.Fatalf("DeleteField failed: %v", err)
	}
	if receivedMethod != http.MethodDelete {
		t.Errorf("expected DELETE, got %s", receivedMethod)
	}
	if receivedPath != "/rest/metadata/fields/field-to-delete" {
		t.Errorf("expected path /rest/metadata/fields/field-to-delete, got %s", receivedPath)
	}
}
