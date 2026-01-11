package fields

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"

	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func TestOutputFieldsTable(t *testing.T) {
	fields := []types.FieldMetadata{
		{
			ID:               "550e8400-e29b-41d4-a716-446655440000",
			ObjectMetadataId: "obj-person",
			Name:             "firstName",
			Label:            "First Name",
			Type:             "TEXT",
			Description:      "Person's first name",
			IsCustom:         false,
			IsActive:         true,
			IsNullable:       false,
			CreatedAt:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt:        time.Date(2024, 6, 20, 14, 45, 0, 0, time.UTC),
		},
		{
			ID:               "550e8400-e29b-41d4-a716-446655440001",
			ObjectMetadataId: "obj-person",
			Name:             "lastName",
			Label:            "Last Name",
			Type:             "TEXT",
			Description:      "Person's last name",
			IsCustom:         false,
			IsActive:         true,
			IsNullable:       false,
			CreatedAt:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt:        time.Date(2024, 6, 20, 14, 45, 0, 0, time.UTC),
		},
		{
			ID:               "550e8400-e29b-41d4-a716-446655440002",
			ObjectMetadataId: "obj-company",
			Name:             "name",
			Label:            "Name",
			Type:             "TEXT",
			Description:      "Company name",
			IsCustom:         false,
			IsActive:         true,
			IsNullable:       false,
			CreatedAt:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt:        time.Date(2024, 6, 20, 14, 45, 0, 0, time.UTC),
		},
	}

	var buf bytes.Buffer
	err := outputFieldsTable(fields, &buf)
	if err != nil {
		t.Fatalf("outputFieldsTable failed: %v", err)
	}

	output := buf.String()

	// Check header
	if !strings.Contains(output, "ID") {
		t.Error("expected output to contain 'ID' header")
	}
	if !strings.Contains(output, "NAME") {
		t.Error("expected output to contain 'NAME' header")
	}
	if !strings.Contains(output, "LABEL") {
		t.Error("expected output to contain 'LABEL' header")
	}
	if !strings.Contains(output, "TYPE") {
		t.Error("expected output to contain 'TYPE' header")
	}
	if !strings.Contains(output, "OBJECT") {
		t.Error("expected output to contain 'OBJECT' header")
	}

	// Check data rows
	if !strings.Contains(output, "firstName") {
		t.Error("expected output to contain 'firstName'")
	}
	if !strings.Contains(output, "First Name") {
		t.Error("expected output to contain 'First Name'")
	}
	if !strings.Contains(output, "lastName") {
		t.Error("expected output to contain 'lastName'")
	}
	if !strings.Contains(output, "obj-person") {
		t.Error("expected output to contain 'obj-person'")
	}
	if !strings.Contains(output, "obj-company") {
		t.Error("expected output to contain 'obj-company'")
	}
	if !strings.Contains(output, "TEXT") {
		t.Error("expected output to contain 'TEXT'")
	}
}

func TestOutputFieldsTableEmpty(t *testing.T) {
	var fields []types.FieldMetadata

	var buf bytes.Buffer
	err := outputFieldsTable(fields, &buf)
	if err != nil {
		t.Fatalf("outputFieldsTable failed: %v", err)
	}

	output := buf.String()

	// Should still have header
	if !strings.Contains(output, "NAME") {
		t.Error("expected empty output to still contain header")
	}
}

func TestOutputFieldsJSON(t *testing.T) {
	fields := []types.FieldMetadata{
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
			CreatedAt:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt:        time.Date(2024, 6, 20, 14, 45, 0, 0, time.UTC),
		},
	}

	var buf bytes.Buffer
	err := outputFieldsJSON(fields, &buf, "")
	if err != nil {
		t.Fatalf("outputFieldsJSON failed: %v", err)
	}

	// Verify valid JSON
	var parsed []types.FieldMetadata
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if len(parsed) != 1 {
		t.Errorf("expected 1 field, got %d", len(parsed))
	}
	if parsed[0].Name != "firstName" {
		t.Errorf("expected Name='firstName', got %q", parsed[0].Name)
	}
}

func TestOutputFieldsCSV(t *testing.T) {
	fields := []types.FieldMetadata{
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
			CreatedAt:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt:        time.Date(2024, 6, 20, 14, 45, 0, 0, time.UTC),
		},
		{
			ID:               "field-2",
			ObjectMetadataId: "obj-company",
			Name:             "companyName",
			Label:            "Company Name",
			Type:             "TEXT",
			Description:      "Company name field",
			IsCustom:         true,
			IsActive:         true,
			IsNullable:       true,
			CreatedAt:        time.Date(2024, 2, 20, 8, 0, 0, 0, time.UTC),
			UpdatedAt:        time.Date(2024, 7, 10, 12, 0, 0, 0, time.UTC),
		},
	}

	var buf bytes.Buffer
	err := outputFieldsCSV(fields, &buf)
	if err != nil {
		t.Fatalf("outputFieldsCSV failed: %v", err)
	}

	// Verify valid CSV
	reader := csv.NewReader(bytes.NewReader(buf.Bytes()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("output is not valid CSV: %v", err)
	}

	// Should have header + 2 data rows
	if len(records) != 3 {
		t.Errorf("expected 3 rows (header + 2 data), got %d", len(records))
	}

	// Check header
	expectedHeaders := []string{"id", "name", "label", "type", "objectMetadataId", "isCustom", "isActive", "isNullable"}
	for i, h := range expectedHeaders {
		if records[0][i] != h {
			t.Errorf("header[%d] = %q, want %q", i, records[0][i], h)
		}
	}

	// Check first data row
	if records[1][0] != "field-1" {
		t.Errorf("row[1][0] = %q, want %q", records[1][0], "field-1")
	}
	if records[1][1] != "firstName" {
		t.Errorf("row[1][1] = %q, want %q", records[1][1], "firstName")
	}
	if records[1][5] != "false" {
		t.Errorf("row[1][5] (isCustom) = %q, want %q", records[1][5], "false")
	}

	// Check second data row
	if records[2][0] != "field-2" {
		t.Errorf("row[2][0] = %q, want %q", records[2][0], "field-2")
	}
	if records[2][5] != "true" {
		t.Errorf("row[2][5] (isCustom) = %q, want %q", records[2][5], "true")
	}
}

func TestOutputFieldsCSVEmpty(t *testing.T) {
	var fields []types.FieldMetadata

	var buf bytes.Buffer
	err := outputFieldsCSV(fields, &buf)
	if err != nil {
		t.Fatalf("outputFieldsCSV failed: %v", err)
	}

	// Should still have header
	reader := csv.NewReader(bytes.NewReader(buf.Bytes()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("output is not valid CSV: %v", err)
	}

	if len(records) != 1 {
		t.Errorf("expected 1 row (header only), got %d", len(records))
	}
}

func TestListCmd_Use(t *testing.T) {
	if listCmd.Use != "list [objectName]" {
		t.Errorf("Use = %q, want %q", listCmd.Use, "list [objectName]")
	}
}

func TestListCmd_Short(t *testing.T) {
	if listCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestListCmd_Long(t *testing.T) {
	if listCmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestListCmd_Args(t *testing.T) {
	// Command should accept 0 or 1 argument
	if listCmd.Args == nil {
		t.Error("Args validator not set")
	}

	// Test with zero args
	err := listCmd.Args(listCmd, []string{})
	if err != nil {
		t.Errorf("Expected no error with zero args, got %v", err)
	}

	// Test with one arg
	err = listCmd.Args(listCmd, []string{"person"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}

	// Test with two args
	err = listCmd.Args(listCmd, []string{"person", "extra"})
	if err == nil {
		t.Error("Expected error with two args")
	}
}

func TestRunList_AllFields(t *testing.T) {
	// Set up mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/rest/metadata/fields" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"fields": [
					{
						"id": "field-1",
						"objectMetadataId": "obj-1",
						"name": "firstName",
						"label": "First Name",
						"type": "TEXT",
						"description": "First name field",
						"isCustom": false,
						"isActive": true,
						"isNullable": false
					},
					{
						"id": "field-2",
						"objectMetadataId": "obj-1",
						"name": "lastName",
						"label": "Last Name",
						"type": "TEXT",
						"description": "Last name field",
						"isCustom": false,
						"isActive": true,
						"isNullable": false
					}
				]
			}
		}`))
	}))
	defer server.Close()

	// Set up environment and viper
	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runList(listCmd, []string{})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runList() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "firstName") {
		t.Errorf("output missing 'firstName': %s", output)
	}
	if !strings.Contains(output, "lastName") {
		t.Errorf("output missing 'lastName': %s", output)
	}
}

func TestRunList_ForObject(t *testing.T) {
	// Set up mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		// First call lists objects to find by name
		if r.URL.Path == "/rest/metadata/objects" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": {
					"objects": [
						{
							"id": "obj-123",
							"nameSingular": "person",
							"namePlural": "people",
							"labelSingular": "Person",
							"labelPlural": "People"
						}
					]
				}
			}`))
			return
		}

		// Second call gets the object details with fields
		if strings.HasPrefix(r.URL.Path, "/rest/metadata/objects/obj-123") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": {
					"object": {
						"id": "obj-123",
						"nameSingular": "person",
						"namePlural": "people",
						"labelSingular": "Person",
						"labelPlural": "People",
						"fields": [
							{
								"id": "field-1",
								"objectMetadataId": "obj-123",
								"name": "firstName",
								"label": "First Name",
								"type": "TEXT",
								"isCustom": false,
								"isActive": true,
								"isNullable": false
							}
						]
					}
				}
			}`))
			return
		}

		t.Errorf("unexpected path: %s", r.URL.Path)
	}))
	defer server.Close()

	// Set up environment and viper
	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runList(listCmd, []string{"person"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runList() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "firstName") {
		t.Errorf("output missing 'firstName': %s", output)
	}
}

func TestRunList_ObjectNotFound(t *testing.T) {
	// Set up mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/metadata/objects" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": {
					"objects": []
				}
			}`))
			return
		}
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	err := runList(listCmd, []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent object")
	}
	if !strings.Contains(err.Error(), "failed to get object") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunList_APIError(t *testing.T) {
	// Set up mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	err := runList(listCmd, []string{})
	if err == nil {
		t.Fatal("expected error for API error")
	}
	if !strings.Contains(err.Error(), "failed to list fields") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunList_NoToken(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	err := runList(listCmd, []string{})
	if err == nil {
		t.Fatal("expected error when no token is available")
	}
}

func TestRunList_TextOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"fields": [
					{
						"id": "field-1",
						"objectMetadataId": "obj-1",
						"name": "firstName",
						"label": "First Name",
						"type": "TEXT",
						"isCustom": false,
						"isActive": true,
						"isNullable": false
					}
				]
			}
		}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runList(listCmd, []string{})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runList() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should have table headers
	if !strings.Contains(output, "ID") {
		t.Errorf("output missing 'ID' header: %s", output)
	}
	if !strings.Contains(output, "NAME") {
		t.Errorf("output missing 'NAME' header: %s", output)
	}
}

func TestRunList_CSVOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"fields": [
					{
						"id": "field-1",
						"objectMetadataId": "obj-1",
						"name": "firstName",
						"label": "First Name",
						"type": "TEXT",
						"isCustom": false,
						"isActive": true,
						"isNullable": false
					}
				]
			}
		}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "csv")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runList(listCmd, []string{})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runList() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify valid CSV
	reader := csv.NewReader(strings.NewReader(output))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("output is not valid CSV: %v", err)
	}

	if len(records) < 2 {
		t.Errorf("expected at least 2 rows (header + data), got %d", len(records))
	}
}

func TestRunList_YAMLOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"fields": [
					{
						"id": "field-1",
						"objectMetadataId": "obj-1",
						"name": "firstName",
						"label": "First Name",
						"type": "TEXT",
						"isCustom": false,
						"isActive": true,
						"isNullable": false
					}
				]
			}
		}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "yaml")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runList(listCmd, []string{})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runList() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// YAML output should have field names
	if !strings.Contains(output, "firstName") {
		t.Errorf("output missing 'firstName': %s", output)
	}
}
