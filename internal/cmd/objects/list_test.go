package objects

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func TestOutputObjectsTable(t *testing.T) {
	objects := []types.ObjectMetadata{
		{
			ID:            "obj-1",
			NameSingular:  "person",
			NamePlural:    "people",
			LabelSingular: "Person",
			LabelPlural:   "People",
			IsCustom:      false,
			IsActive:      true,
			Fields: []types.FieldMetadata{
				{ID: "f1", Name: "firstName"},
				{ID: "f2", Name: "lastName"},
				{ID: "f3", Name: "email"},
			},
		},
		{
			ID:            "obj-2",
			NameSingular:  "company",
			NamePlural:    "companies",
			LabelSingular: "Company",
			LabelPlural:   "Companies",
			IsCustom:      false,
			IsActive:      true,
			Fields: []types.FieldMetadata{
				{ID: "f4", Name: "name"},
				{ID: "f5", Name: "website"},
			},
		},
		{
			ID:            "obj-3",
			NameSingular:  "customObject",
			NamePlural:    "customObjects",
			LabelSingular: "Custom Object",
			LabelPlural:   "Custom Objects",
			IsCustom:      true,
			IsActive:      true,
			Fields:        []types.FieldMetadata{},
		},
	}

	var buf bytes.Buffer
	err := outputObjectsTable(objects, &buf)
	if err != nil {
		t.Fatalf("outputObjectsTable failed: %v", err)
	}

	output := buf.String()

	// Check header
	if !strings.Contains(output, "NAME") {
		t.Error("expected output to contain 'NAME' header")
	}
	if !strings.Contains(output, "LABEL") {
		t.Error("expected output to contain 'LABEL' header")
	}
	if !strings.Contains(output, "CUSTOM") {
		t.Error("expected output to contain 'CUSTOM' header")
	}
	if !strings.Contains(output, "FIELDS") {
		t.Error("expected output to contain 'FIELDS' header")
	}

	// Check data rows
	if !strings.Contains(output, "person") {
		t.Error("expected output to contain 'person'")
	}
	if !strings.Contains(output, "People") {
		t.Error("expected output to contain 'People'")
	}
	if !strings.Contains(output, "company") {
		t.Error("expected output to contain 'company'")
	}
	if !strings.Contains(output, "customObject") {
		t.Error("expected output to contain 'customObject'")
	}

	// Check field counts
	if !strings.Contains(output, "3") {
		t.Error("expected output to contain field count '3' for person")
	}
	if !strings.Contains(output, "2") {
		t.Error("expected output to contain field count '2' for company")
	}
	if !strings.Contains(output, "0") {
		t.Error("expected output to contain field count '0' for customObject")
	}
}

func TestOutputObjectsTableEmpty(t *testing.T) {
	var objects []types.ObjectMetadata

	var buf bytes.Buffer
	err := outputObjectsTable(objects, &buf)
	if err != nil {
		t.Fatalf("outputObjectsTable failed: %v", err)
	}

	output := buf.String()

	// Should still have header
	if !strings.Contains(output, "NAME") {
		t.Error("expected empty output to still contain header")
	}
}

func TestOutputObjectsJSON(t *testing.T) {
	objects := []types.ObjectMetadata{
		{
			ID:            "obj-1",
			NameSingular:  "person",
			NamePlural:    "people",
			LabelSingular: "Person",
			LabelPlural:   "People",
			IsCustom:      false,
			IsActive:      true,
		},
	}

	var buf bytes.Buffer
	err := outputObjectsJSON(objects, &buf, "")
	if err != nil {
		t.Fatalf("outputObjectsJSON failed: %v", err)
	}

	// Verify valid JSON
	var parsed []types.ObjectMetadata
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if len(parsed) != 1 {
		t.Errorf("expected 1 object, got %d", len(parsed))
	}
	if parsed[0].NameSingular != "person" {
		t.Errorf("expected NameSingular='person', got %q", parsed[0].NameSingular)
	}
}

func TestOutputObjectsCSV(t *testing.T) {
	objects := []types.ObjectMetadata{
		{
			ID:            "obj-1",
			NameSingular:  "person",
			NamePlural:    "people",
			LabelSingular: "Person",
			LabelPlural:   "People",
			IsCustom:      false,
			IsActive:      true,
			Fields: []types.FieldMetadata{
				{ID: "f1", Name: "firstName"},
				{ID: "f2", Name: "lastName"},
			},
		},
		{
			ID:            "obj-2",
			NameSingular:  "company",
			NamePlural:    "companies",
			LabelSingular: "Company",
			LabelPlural:   "Companies",
			IsCustom:      true,
			IsActive:      false,
			Fields:        []types.FieldMetadata{},
		},
	}

	var buf bytes.Buffer
	err := outputObjectsCSV(objects, &buf)
	if err != nil {
		t.Fatalf("outputObjectsCSV failed: %v", err)
	}

	output := buf.String()

	// Check header row
	if !strings.Contains(output, "name,namePlural,label,labelPlural,isCustom,isActive,fieldCount") {
		t.Error("expected CSV header row")
	}

	// Check data rows
	if !strings.Contains(output, "person,people,Person,People,false,true,2") {
		t.Errorf("expected first CSV row with person data, got: %s", output)
	}
	if !strings.Contains(output, "company,companies,Company,Companies,true,false,0") {
		t.Errorf("expected second CSV row with company data, got: %s", output)
	}
}

func TestOutputObjectsCSVEmpty(t *testing.T) {
	var objects []types.ObjectMetadata

	var buf bytes.Buffer
	err := outputObjectsCSV(objects, &buf)
	if err != nil {
		t.Fatalf("outputObjectsCSV failed: %v", err)
	}

	output := buf.String()

	// Should still have header
	if !strings.Contains(output, "name,namePlural,label,labelPlural,isCustom,isActive,fieldCount") {
		t.Error("expected CSV header row even with empty data")
	}
}

func TestOutputObjects_Text(t *testing.T) {
	objects := []types.ObjectMetadata{
		{
			ID:            "obj-1",
			NameSingular:  "person",
			NamePlural:    "people",
			LabelSingular: "Person",
			LabelPlural:   "People",
			IsCustom:      false,
			IsActive:      true,
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputObjects(objects, "text", "")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputObjects (text) failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "person") {
		t.Error("expected output to contain 'person'")
	}
	if !strings.Contains(output, "NAME") {
		t.Error("expected output to contain table header 'NAME'")
	}
}

func TestOutputObjects_JSON(t *testing.T) {
	objects := []types.ObjectMetadata{
		{
			ID:            "obj-1",
			NameSingular:  "person",
			NamePlural:    "people",
			LabelSingular: "Person",
			LabelPlural:   "People",
			IsCustom:      false,
			IsActive:      true,
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputObjects(objects, "json", "")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputObjects (json) failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should be valid JSON
	var parsed []types.ObjectMetadata
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
}

func TestOutputObjects_YAML(t *testing.T) {
	objects := []types.ObjectMetadata{
		{
			ID:            "obj-1",
			NameSingular:  "person",
			NamePlural:    "people",
			LabelSingular: "Person",
			LabelPlural:   "People",
			IsCustom:      false,
			IsActive:      true,
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputObjects(objects, "yaml", "")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputObjects (yaml) failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// YAML should contain nameSingular field
	if !strings.Contains(output, "namesingular: person") && !strings.Contains(output, "nameSingular: person") {
		t.Errorf("expected YAML output to contain nameSingular, got: %s", output)
	}
}

func TestOutputObjects_CSV(t *testing.T) {
	objects := []types.ObjectMetadata{
		{
			ID:            "obj-1",
			NameSingular:  "person",
			NamePlural:    "people",
			LabelSingular: "Person",
			LabelPlural:   "People",
			IsCustom:      false,
			IsActive:      true,
			Fields:        []types.FieldMetadata{{ID: "f1"}},
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputObjects(objects, "csv", "")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputObjects (csv) failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "person") {
		t.Error("expected output to contain 'person'")
	}
	if !strings.Contains(output, "name,namePlural") {
		t.Error("expected CSV header row")
	}
}

func TestRunList_Success(t *testing.T) {
	// Set up mock server
	response := types.ObjectsListResponse{}
	response.Data.Objects = []types.ObjectMetadata{
		{
			ID:            "obj-1",
			NameSingular:  "person",
			NamePlural:    "people",
			LabelSingular: "Person",
			LabelPlural:   "People",
			IsCustom:      false,
			IsActive:      true,
		},
		{
			ID:            "obj-2",
			NameSingular:  "company",
			NamePlural:    "companies",
			LabelSingular: "Company",
			LabelPlural:   "Companies",
			IsCustom:      false,
			IsActive:      true,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/metadata/objects" {
			t.Errorf("expected path /rest/metadata/objects, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
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

	err := runList(nil, nil)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runList() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify JSON output contains expected objects
	if !strings.Contains(output, "person") {
		t.Errorf("output missing 'person': %s", output)
	}
	if !strings.Contains(output, "company") {
		t.Errorf("output missing 'company': %s", output)
	}
}

func TestRunList_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	err := runList(nil, nil)
	if err == nil {
		t.Fatal("expected error for API error response")
	}
	if !strings.Contains(err.Error(), "failed to list objects") {
		t.Errorf("expected 'failed to list objects' error, got: %v", err)
	}
}

func TestRunList_NoToken(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	err := runList(nil, nil)
	if err == nil {
		t.Fatal("expected error when no token is available")
	}
}

func TestRunList_TextOutput(t *testing.T) {
	response := types.ObjectsListResponse{}
	response.Data.Objects = []types.ObjectMetadata{
		{
			ID:            "obj-1",
			NameSingular:  "person",
			NamePlural:    "people",
			LabelSingular: "Person",
			LabelPlural:   "People",
			IsCustom:      false,
			IsActive:      true,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "text")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runList(nil, nil)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runList() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify table output contains headers and data
	if !strings.Contains(output, "NAME") {
		t.Errorf("output missing 'NAME' header: %s", output)
	}
	if !strings.Contains(output, "person") {
		t.Errorf("output missing 'person': %s", output)
	}
}

func TestListCmd_Use(t *testing.T) {
	if listCmd.Use != "list" {
		t.Errorf("Use = %q, want %q", listCmd.Use, "list")
	}
}

func TestListCmd_Short(t *testing.T) {
	if listCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestListCmd_HasRunE(t *testing.T) {
	if listCmd.RunE == nil {
		t.Error("listCmd should have RunE set")
	}
}
