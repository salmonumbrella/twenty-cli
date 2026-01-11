package objects

import (
	"bytes"
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

func TestOutputObjectDetailTable(t *testing.T) {
	obj := &types.ObjectMetadata{
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
				ID:       "f1",
				Name:     "firstName",
				Label:    "First Name",
				Type:     "TEXT",
				IsCustom: false,
				IsActive: true,
			},
			{
				ID:       "f2",
				Name:     "lastName",
				Label:    "Last Name",
				Type:     "TEXT",
				IsCustom: false,
				IsActive: true,
			},
			{
				ID:       "f3",
				Name:     "email",
				Label:    "Email",
				Type:     "EMAIL",
				IsCustom: false,
				IsActive: true,
			},
		},
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 6, 20, 14, 45, 0, 0, time.UTC),
	}

	var buf bytes.Buffer
	err := outputObjectDetailTable(obj, &buf)
	if err != nil {
		t.Fatalf("outputObjectDetailTable failed: %v", err)
	}

	output := buf.String()

	// Check object info
	if !strings.Contains(output, "person") {
		t.Error("expected output to contain object name 'person'")
	}
	if !strings.Contains(output, "People") {
		t.Error("expected output to contain label 'People'")
	}
	if !strings.Contains(output, "A contact in the CRM") {
		t.Error("expected output to contain description")
	}
	if !strings.Contains(output, "false") {
		t.Error("expected output to contain 'false' for custom status")
	}
	if !strings.Contains(output, "true") {
		t.Error("expected output to contain 'true' for active status")
	}

	// Check fields header
	if !strings.Contains(output, "Fields:") {
		t.Error("expected output to contain 'Fields:' section")
	}
	if !strings.Contains(output, "NAME") {
		t.Error("expected output to contain field 'NAME' header")
	}
	if !strings.Contains(output, "LABEL") {
		t.Error("expected output to contain field 'LABEL' header")
	}
	if !strings.Contains(output, "TYPE") {
		t.Error("expected output to contain field 'TYPE' header")
	}
	if !strings.Contains(output, "CUSTOM") {
		t.Error("expected output to contain field 'CUSTOM' header")
	}

	// Check field data
	if !strings.Contains(output, "firstName") {
		t.Error("expected output to contain field 'firstName'")
	}
	if !strings.Contains(output, "First Name") {
		t.Error("expected output to contain field label 'First Name'")
	}
	if !strings.Contains(output, "TEXT") {
		t.Error("expected output to contain field type 'TEXT'")
	}
	if !strings.Contains(output, "EMAIL") {
		t.Error("expected output to contain field type 'EMAIL'")
	}
}

func TestOutputObjectDetailTableNoFields(t *testing.T) {
	obj := &types.ObjectMetadata{
		ID:            "obj-empty",
		NameSingular:  "emptyObject",
		NamePlural:    "emptyObjects",
		LabelSingular: "Empty Object",
		LabelPlural:   "Empty Objects",
		Description:   "",
		IsCustom:      true,
		IsActive:      false,
		Fields:        []types.FieldMetadata{},
	}

	var buf bytes.Buffer
	err := outputObjectDetailTable(obj, &buf)
	if err != nil {
		t.Fatalf("outputObjectDetailTable failed: %v", err)
	}

	output := buf.String()

	// Check object info still appears
	if !strings.Contains(output, "emptyObject") {
		t.Error("expected output to contain object name")
	}
	if !strings.Contains(output, "Empty Objects") {
		t.Error("expected output to contain label")
	}

	// Should still have Fields section header
	if !strings.Contains(output, "Fields:") {
		t.Error("expected output to contain 'Fields:' section even with no fields")
	}
}

func TestOutputObjectDetailJSON(t *testing.T) {
	obj := &types.ObjectMetadata{
		ID:            "obj-person",
		NameSingular:  "person",
		NamePlural:    "people",
		LabelSingular: "Person",
		LabelPlural:   "People",
		Description:   "A contact",
		IsCustom:      false,
		IsActive:      true,
		Fields: []types.FieldMetadata{
			{
				ID:       "f1",
				Name:     "firstName",
				Label:    "First Name",
				Type:     "TEXT",
				IsCustom: false,
				IsActive: true,
			},
		},
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 6, 20, 14, 45, 0, 0, time.UTC),
	}

	var buf bytes.Buffer
	err := outputObjectDetailJSON(obj, &buf, "")
	if err != nil {
		t.Fatalf("outputObjectDetailJSON failed: %v", err)
	}

	// Verify valid JSON
	var parsed types.ObjectMetadata
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if parsed.NameSingular != "person" {
		t.Errorf("expected NameSingular='person', got %q", parsed.NameSingular)
	}
	if len(parsed.Fields) != 1 {
		t.Errorf("expected 1 field, got %d", len(parsed.Fields))
	}
	if parsed.Fields[0].Name != "firstName" {
		t.Errorf("expected first field Name='firstName', got %q", parsed.Fields[0].Name)
	}
}

func TestOutputObjectDetailCSV(t *testing.T) {
	obj := &types.ObjectMetadata{
		ID:            "obj-person",
		NameSingular:  "person",
		NamePlural:    "people",
		LabelSingular: "Person",
		LabelPlural:   "People",
		Description:   "A contact",
		IsCustom:      false,
		IsActive:      true,
		Fields: []types.FieldMetadata{
			{
				ID:       "f1",
				Name:     "firstName",
				Label:    "First Name",
				Type:     "TEXT",
				IsCustom: false,
				IsActive: true,
			},
			{
				ID:       "f2",
				Name:     "email",
				Label:    "Email",
				Type:     "EMAIL",
				IsCustom: true,
				IsActive: false,
			},
		},
	}

	var buf bytes.Buffer
	err := outputObjectDetailCSV(obj, &buf)
	if err != nil {
		t.Fatalf("outputObjectDetailCSV failed: %v", err)
	}

	output := buf.String()

	// Check header row
	if !strings.Contains(output, "fieldName,fieldLabel,fieldType,isCustom,isActive") {
		t.Error("expected CSV header row")
	}

	// Check data rows
	if !strings.Contains(output, "firstName,First Name,TEXT,false,true") {
		t.Errorf("expected first CSV row with firstName data, got: %s", output)
	}
	if !strings.Contains(output, "email,Email,EMAIL,true,false") {
		t.Errorf("expected second CSV row with email data, got: %s", output)
	}
}

func TestOutputObjectDetailCSVEmpty(t *testing.T) {
	obj := &types.ObjectMetadata{
		ID:           "obj-empty",
		NameSingular: "emptyObject",
		NamePlural:   "emptyObjects",
		Fields:       []types.FieldMetadata{},
	}

	var buf bytes.Buffer
	err := outputObjectDetailCSV(obj, &buf)
	if err != nil {
		t.Fatalf("outputObjectDetailCSV failed: %v", err)
	}

	output := buf.String()

	// Should still have header
	if !strings.Contains(output, "fieldName,fieldLabel,fieldType,isCustom,isActive") {
		t.Error("expected CSV header row even with no fields")
	}
}

func TestOutputObjectDetail_Text(t *testing.T) {
	obj := &types.ObjectMetadata{
		ID:            "obj-1",
		NameSingular:  "person",
		NamePlural:    "people",
		LabelSingular: "Person",
		LabelPlural:   "People",
		IsCustom:      false,
		IsActive:      true,
		Fields: []types.FieldMetadata{
			{ID: "f1", Name: "firstName", Label: "First Name", Type: "TEXT"},
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputObjectDetail(obj, "text", "")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputObjectDetail (text) failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "person") {
		t.Error("expected output to contain 'person'")
	}
	if !strings.Contains(output, "Fields:") {
		t.Error("expected output to contain 'Fields:' section")
	}
}

func TestOutputObjectDetail_JSON(t *testing.T) {
	obj := &types.ObjectMetadata{
		ID:            "obj-1",
		NameSingular:  "person",
		NamePlural:    "people",
		LabelSingular: "Person",
		LabelPlural:   "People",
		IsCustom:      false,
		IsActive:      true,
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputObjectDetail(obj, "json", "")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputObjectDetail (json) failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should be valid JSON
	var parsed types.ObjectMetadata
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if parsed.NameSingular != "person" {
		t.Errorf("expected NameSingular='person', got %q", parsed.NameSingular)
	}
}

func TestOutputObjectDetail_CSV(t *testing.T) {
	obj := &types.ObjectMetadata{
		ID:            "obj-1",
		NameSingular:  "person",
		NamePlural:    "people",
		LabelSingular: "Person",
		LabelPlural:   "People",
		IsCustom:      false,
		IsActive:      true,
		Fields: []types.FieldMetadata{
			{ID: "f1", Name: "firstName", Label: "First Name", Type: "TEXT", IsCustom: false, IsActive: true},
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputObjectDetail(obj, "csv", "")
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputObjectDetail (csv) failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "fieldName,fieldLabel") {
		t.Error("expected CSV header row")
	}
	if !strings.Contains(output, "firstName") {
		t.Error("expected output to contain 'firstName'")
	}
}

func TestRunGet_Success(t *testing.T) {
	// Set up mock server
	listResponse := types.ObjectsListResponse{}
	listResponse.Data.Objects = []types.ObjectMetadata{
		{
			ID:           "obj-person",
			NameSingular: "person",
			NamePlural:   "people",
		},
	}

	objectResponse := types.ObjectResponse{}
	objectResponse.Data.Object = types.ObjectMetadata{
		ID:            "obj-person",
		NameSingular:  "person",
		NamePlural:    "people",
		LabelSingular: "Person",
		LabelPlural:   "People",
		Description:   "A contact",
		IsCustom:      false,
		IsActive:      true,
		Fields: []types.FieldMetadata{
			{ID: "f1", Name: "firstName", Label: "First Name", Type: "TEXT"},
			{ID: "f2", Name: "lastName", Label: "Last Name", Type: "TEXT"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/rest/metadata/objects" {
			json.NewEncoder(w).Encode(listResponse)
		} else if r.URL.Path == "/rest/metadata/objects/obj-person" {
			json.NewEncoder(w).Encode(objectResponse)
		} else {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

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

	err := runGet(nil, []string{"person"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGet() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "person") {
		t.Errorf("output missing 'person': %s", output)
	}
	if !strings.Contains(output, "firstName") {
		t.Errorf("output missing 'firstName': %s", output)
	}
}

func TestRunGet_ByUUID(t *testing.T) {
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
		// Should call direct endpoint with UUID
		if r.URL.Path != "/rest/metadata/objects/12345678-1234-1234-1234-123456789012" {
			t.Errorf("expected direct path with UUID, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(objectResponse)
	}))
	defer server.Close()

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

	err := runGet(nil, []string{"12345678-1234-1234-1234-123456789012"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGet() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "customObject") {
		t.Errorf("output missing 'customObject': %s", output)
	}
}

func TestRunGet_NotFound(t *testing.T) {
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
			w.Write([]byte(`{"error": "not found"}`))
		}
	}))
	defer server.Close()

	t.Setenv("TWENTY_TOKEN", "test-token")
	viper.Set("base_url", server.URL)
	viper.Set("debug", false)
	viper.Set("output", "json")
	viper.Set("query", "")
	t.Cleanup(viper.Reset)

	err := runGet(nil, []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for not found object")
	}
	if !strings.Contains(err.Error(), "failed to get object") {
		t.Errorf("expected 'failed to get object' error, got: %v", err)
	}
}

func TestRunGet_APIError(t *testing.T) {
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

	err := runGet(nil, []string{"person"})
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestRunGet_NoToken(t *testing.T) {
	t.Setenv("TWENTY_TOKEN", "")
	viper.Set("base_url", "http://localhost")
	viper.Set("profile", "nonexistent-profile-for-testing")
	t.Cleanup(viper.Reset)

	err := runGet(nil, []string{"person"})
	if err == nil {
		t.Fatal("expected error when no token is available")
	}
}

func TestRunGet_TextOutput(t *testing.T) {
	listResponse := types.ObjectsListResponse{}
	listResponse.Data.Objects = []types.ObjectMetadata{
		{ID: "obj-person", NameSingular: "person", NamePlural: "people"},
	}

	objectResponse := types.ObjectResponse{}
	objectResponse.Data.Object = types.ObjectMetadata{
		ID:            "obj-person",
		NameSingular:  "person",
		NamePlural:    "people",
		LabelSingular: "Person",
		LabelPlural:   "People",
		Description:   "A contact",
		IsCustom:      false,
		IsActive:      true,
		Fields: []types.FieldMetadata{
			{ID: "f1", Name: "firstName", Label: "First Name", Type: "TEXT"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/rest/metadata/objects" {
			json.NewEncoder(w).Encode(listResponse)
		} else {
			json.NewEncoder(w).Encode(objectResponse)
		}
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

	err := runGet(nil, []string{"person"})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runGet() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify table output contains object info and fields
	if !strings.Contains(output, "Object:") {
		t.Errorf("output missing 'Object:' header: %s", output)
	}
	if !strings.Contains(output, "person") {
		t.Errorf("output missing 'person': %s", output)
	}
	if !strings.Contains(output, "Fields:") {
		t.Errorf("output missing 'Fields:' section: %s", output)
	}
}

func TestGetCmd_Use(t *testing.T) {
	if getCmd.Use != "get <objectName>" {
		t.Errorf("Use = %q, want %q", getCmd.Use, "get <objectName>")
	}
}

func TestGetCmd_Short(t *testing.T) {
	if getCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestGetCmd_HasRunE(t *testing.T) {
	if getCmd.RunE == nil {
		t.Error("getCmd should have RunE set")
	}
}

func TestGetCmd_Args(t *testing.T) {
	// getCmd requires exactly 1 argument
	if getCmd.Args == nil {
		t.Error("getCmd should have Args validation set")
	}
}
