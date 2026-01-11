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

func TestOutputFieldDetailTable(t *testing.T) {
	field := &types.FieldMetadata{
		ID:               "field-123",
		ObjectMetadataId: "obj-person",
		Name:             "firstName",
		Label:            "First Name",
		Type:             "TEXT",
		Description:      "The person's first name",
		IsCustom:         false,
		IsActive:         true,
		IsNullable:       false,
		CreatedAt:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2024, 6, 20, 14, 45, 0, 0, time.UTC),
	}

	var buf bytes.Buffer
	err := outputFieldDetailTable(field, &buf)
	if err != nil {
		t.Fatalf("outputFieldDetailTable failed: %v", err)
	}

	output := buf.String()

	// Check field info is displayed
	if !strings.Contains(output, "firstName") {
		t.Error("expected output to contain field name 'firstName'")
	}
	if !strings.Contains(output, "First Name") {
		t.Error("expected output to contain label 'First Name'")
	}
	if !strings.Contains(output, "TEXT") {
		t.Error("expected output to contain type 'TEXT'")
	}
	if !strings.Contains(output, "obj-person") {
		t.Error("expected output to contain object ID 'obj-person'")
	}
	if !strings.Contains(output, "false") {
		t.Error("expected output to contain 'false' for custom/nullable status")
	}
	if !strings.Contains(output, "true") {
		t.Error("expected output to contain 'true' for active status")
	}
	if !strings.Contains(output, "The person's first name") {
		t.Error("expected output to contain description")
	}
}

func TestOutputFieldDetailTableCustomField(t *testing.T) {
	field := &types.FieldMetadata{
		ID:               "field-custom",
		ObjectMetadataId: "obj-company",
		Name:             "annualRevenue",
		Label:            "Annual Revenue",
		Type:             "NUMBER",
		Description:      "Company annual revenue",
		IsCustom:         true,
		IsActive:         true,
		IsNullable:       true,
		CreatedAt:        time.Date(2024, 2, 10, 8, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2024, 7, 15, 12, 30, 0, 0, time.UTC),
	}

	var buf bytes.Buffer
	err := outputFieldDetailTable(field, &buf)
	if err != nil {
		t.Fatalf("outputFieldDetailTable failed: %v", err)
	}

	output := buf.String()

	// Check custom field specific info
	if !strings.Contains(output, "annualRevenue") {
		t.Error("expected output to contain field name 'annualRevenue'")
	}
	if !strings.Contains(output, "NUMBER") {
		t.Error("expected output to contain type 'NUMBER'")
	}
	if !strings.Contains(output, "obj-company") {
		t.Error("expected output to contain object ID 'obj-company'")
	}
	// IsCustom should be true
	if !strings.Contains(output, "Custom:") || !strings.Contains(output, "true") {
		t.Error("expected output to indicate custom field status")
	}
}

func TestOutputFieldDetailTableEmptyDescription(t *testing.T) {
	field := &types.FieldMetadata{
		ID:               "field-empty",
		ObjectMetadataId: "obj-task",
		Name:             "status",
		Label:            "Status",
		Type:             "SELECT",
		Description:      "",
		IsCustom:         false,
		IsActive:         true,
		IsNullable:       false,
		CreatedAt:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2024, 6, 20, 14, 45, 0, 0, time.UTC),
	}

	var buf bytes.Buffer
	err := outputFieldDetailTable(field, &buf)
	if err != nil {
		t.Fatalf("outputFieldDetailTable failed: %v", err)
	}

	output := buf.String()

	// Should still display field info
	if !strings.Contains(output, "status") {
		t.Error("expected output to contain field name 'status'")
	}
	if !strings.Contains(output, "SELECT") {
		t.Error("expected output to contain type 'SELECT'")
	}
}

func TestOutputFieldDetailJSON(t *testing.T) {
	field := &types.FieldMetadata{
		ID:               "field-json",
		ObjectMetadataId: "obj-person",
		Name:             "email",
		Label:            "Email",
		Type:             "EMAIL",
		Description:      "Email address",
		IsCustom:         false,
		IsActive:         true,
		IsNullable:       false,
		CreatedAt:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2024, 6, 20, 14, 45, 0, 0, time.UTC),
	}

	var buf bytes.Buffer
	err := outputFieldDetailJSON(field, &buf, "")
	if err != nil {
		t.Fatalf("outputFieldDetailJSON failed: %v", err)
	}

	// Verify valid JSON
	var parsed types.FieldMetadata
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if parsed.Name != "email" {
		t.Errorf("expected Name='email', got %q", parsed.Name)
	}
	if parsed.Type != "EMAIL" {
		t.Errorf("expected Type='EMAIL', got %q", parsed.Type)
	}
	if parsed.ObjectMetadataId != "obj-person" {
		t.Errorf("expected ObjectMetadataId='obj-person', got %q", parsed.ObjectMetadataId)
	}
}

func TestOutputFieldDetailCSV(t *testing.T) {
	field := &types.FieldMetadata{
		ID:               "field-csv",
		ObjectMetadataId: "obj-person",
		Name:             "phone",
		Label:            "Phone Number",
		Type:             "PHONE",
		Description:      "Contact phone number",
		IsCustom:         true,
		IsActive:         true,
		IsNullable:       true,
		CreatedAt:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2024, 6, 20, 14, 45, 0, 0, time.UTC),
	}

	var buf bytes.Buffer
	err := outputFieldDetailCSV(field, &buf)
	if err != nil {
		t.Fatalf("outputFieldDetailCSV failed: %v", err)
	}

	// Verify valid CSV
	reader := csv.NewReader(bytes.NewReader(buf.Bytes()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("output is not valid CSV: %v", err)
	}

	// Should have header + 1 data row
	if len(records) != 2 {
		t.Errorf("expected 2 rows (header + data), got %d", len(records))
	}

	// Check header
	expectedHeaders := []string{"id", "name", "label", "type", "objectMetadataId", "isCustom", "isActive", "isNullable", "description"}
	for i, h := range expectedHeaders {
		if records[0][i] != h {
			t.Errorf("header[%d] = %q, want %q", i, records[0][i], h)
		}
	}

	// Check data
	if records[1][0] != "field-csv" {
		t.Errorf("ID = %q, want %q", records[1][0], "field-csv")
	}
	if records[1][1] != "phone" {
		t.Errorf("name = %q, want %q", records[1][1], "phone")
	}
	if records[1][5] != "true" {
		t.Errorf("isCustom = %q, want %q", records[1][5], "true")
	}
	if records[1][8] != "Contact phone number" {
		t.Errorf("description = %q, want %q", records[1][8], "Contact phone number")
	}
}

func TestGetCmd_Use(t *testing.T) {
	if getCmd.Use != "get <fieldId>" {
		t.Errorf("Use = %q, want %q", getCmd.Use, "get <fieldId>")
	}
}

func TestGetCmd_Short(t *testing.T) {
	if getCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestGetCmd_Long(t *testing.T) {
	if getCmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestGetCmd_Args(t *testing.T) {
	// Command should require exactly 1 argument
	if getCmd.Args == nil {
		t.Error("Args validator not set")
	}

	// Test with zero args
	err := getCmd.Args(getCmd, []string{})
	if err == nil {
		t.Error("Expected error with zero args")
	}

	// Test with one arg
	err = getCmd.Args(getCmd, []string{"field-id"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}

	// Test with two args
	err = getCmd.Args(getCmd, []string{"field-1", "field-2"})
	if err == nil {
		t.Error("Expected error with two args")
	}
}

func TestRunGet_Success(t *testing.T) {
	// Set up mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/rest/metadata/fields/field-123") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"field": {
					"id": "field-123",
					"objectMetadataId": "obj-person",
					"name": "firstName",
					"label": "First Name",
					"type": "TEXT",
					"description": "First name field",
					"isCustom": false,
					"isActive": true,
					"isNullable": false
				}
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

	err := runGet(getCmd, []string{"field-123"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGet() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "firstName") {
		t.Errorf("output missing 'firstName': %s", output)
	}
	if !strings.Contains(output, "field-123") {
		t.Errorf("output missing 'field-123': %s", output)
	}
}

func TestRunGet_NotFound(t *testing.T) {
	// Set up mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "field not found"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	err := runGet(getCmd, []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent field")
	}
	if !strings.Contains(err.Error(), "failed to get field") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunGet_NoToken(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	err := runGet(getCmd, []string{"field-123"})
	if err == nil {
		t.Fatal("expected error when no token is available")
	}
}

func TestRunGet_TextOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"field": {
					"id": "field-123",
					"objectMetadataId": "obj-person",
					"name": "firstName",
					"label": "First Name",
					"type": "TEXT",
					"description": "First name field",
					"isCustom": false,
					"isActive": true,
					"isNullable": false
				}
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

	err := runGet(getCmd, []string{"field-123"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGet() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should have table-formatted output with labels
	if !strings.Contains(output, "Field:") {
		t.Errorf("output missing 'Field:' label: %s", output)
	}
	if !strings.Contains(output, "Type:") {
		t.Errorf("output missing 'Type:' label: %s", output)
	}
}

func TestRunGet_CSVOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"field": {
					"id": "field-123",
					"objectMetadataId": "obj-person",
					"name": "firstName",
					"label": "First Name",
					"type": "TEXT",
					"description": "First name field",
					"isCustom": false,
					"isActive": true,
					"isNullable": false
				}
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

	err := runGet(getCmd, []string{"field-123"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGet() error = %v", err)
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

	if len(records) != 2 {
		t.Errorf("expected 2 rows (header + data), got %d", len(records))
	}
}

func TestRunGet_APIError(t *testing.T) {
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

	err := runGet(getCmd, []string{"field-123"})
	if err == nil {
		t.Fatal("expected error for API error")
	}
}

func TestOutputFieldDetail_AllFormats(t *testing.T) {
	field := &types.FieldMetadata{
		ID:               "field-test",
		ObjectMetadataId: "obj-test",
		Name:             "testField",
		Label:            "Test Field",
		Type:             "TEXT",
		Description:      "Test description",
		IsCustom:         false,
		IsActive:         true,
		IsNullable:       false,
	}

	tests := []struct {
		name   string
		format string
		check  func(string) bool
	}{
		{
			name:   "json format",
			format: "json",
			check: func(output string) bool {
				return strings.Contains(output, `"name"`) && strings.Contains(output, "testField")
			},
		},
		{
			name:   "csv format",
			format: "csv",
			check: func(output string) bool {
				return strings.Contains(output, "id,name,label,type")
			},
		},
		{
			name:   "text format",
			format: "text",
			check: func(output string) bool {
				return strings.Contains(output, "Field:") && strings.Contains(output, "testField")
			},
		},
		{
			name:   "default format",
			format: "",
			check: func(output string) bool {
				return strings.Contains(output, "Field:") && strings.Contains(output, "testField")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := outputFieldDetail(field, tt.format, "")
			w.Close()
			os.Stdout = oldStdout

			if err != nil {
				t.Fatalf("outputFieldDetail() error = %v", err)
			}

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			if !tt.check(output) {
				t.Errorf("format %q output check failed: %s", tt.format, output)
			}
		})
	}
}
