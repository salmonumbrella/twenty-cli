package people

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"testing"

	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func TestFormatPeopleAsJSON(t *testing.T) {
	people := []types.Person{
		{
			ID: "person-1",
			Name: types.Name{
				FirstName: "John",
				LastName:  "Doe",
			},
			Email: types.Email{
				PrimaryEmail: "john@example.com",
			},
			Phone: types.Phone{
				PrimaryPhoneNumber: "+1234567890",
			},
			JobTitle: "Engineer",
			City:     "New York",
		},
		{
			ID: "person-2",
			Name: types.Name{
				FirstName: "Jane",
				LastName:  "Smith",
			},
			Email: types.Email{
				PrimaryEmail: "jane@example.com",
			},
			Phone: types.Phone{
				PrimaryPhoneNumber: "+0987654321",
			},
			JobTitle: "Manager",
			City:     "Los Angeles",
		},
	}

	var buf bytes.Buffer
	err := formatPeopleAsJSON(people, &buf, "")
	if err != nil {
		t.Fatalf("formatPeopleAsJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var parsed []types.Person
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if len(parsed) != 2 {
		t.Errorf("expected 2 people, got %d", len(parsed))
	}

	if parsed[0].Name.FirstName != "John" {
		t.Errorf("expected first person FirstName='John', got %q", parsed[0].Name.FirstName)
	}
}

func TestFormatPeopleAsCSV(t *testing.T) {
	people := []types.Person{
		{
			ID: "person-1",
			Name: types.Name{
				FirstName: "John",
				LastName:  "Doe",
			},
			Email: types.Email{
				PrimaryEmail: "john@example.com",
			},
			Phone: types.Phone{
				PrimaryPhoneNumber: "+1234567890",
			},
			JobTitle: "Engineer",
			City:     "New York",
		},
		{
			ID: "person-2",
			Name: types.Name{
				FirstName: "Jane",
				LastName:  "Smith",
			},
			Email: types.Email{
				PrimaryEmail: "jane@example.com",
			},
			Phone: types.Phone{
				PrimaryPhoneNumber: "+0987654321",
			},
			JobTitle: "Manager",
			City:     "Los Angeles",
		},
	}

	var buf bytes.Buffer
	err := formatPeopleAsCSV(people, &buf)
	if err != nil {
		t.Fatalf("formatPeopleAsCSV failed: %v", err)
	}

	// Parse CSV output
	reader := csv.NewReader(bytes.NewReader(buf.Bytes()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("output is not valid CSV: %v", err)
	}

	// Check header
	expectedHeader := []string{"ID", "FirstName", "LastName", "Email", "JobTitle", "City", "Phone"}
	if len(records) < 1 {
		t.Fatal("CSV has no header row")
	}
	for i, h := range expectedHeader {
		if records[0][i] != h {
			t.Errorf("header[%d]: expected %q, got %q", i, h, records[0][i])
		}
	}

	// Check data rows (header + 2 data rows = 3 total)
	if len(records) != 3 {
		t.Errorf("expected 3 rows (header + 2 data), got %d", len(records))
	}

	// Check first data row
	if records[1][0] != "person-1" {
		t.Errorf("row 1 ID: expected 'person-1', got %q", records[1][0])
	}
	if records[1][1] != "John" {
		t.Errorf("row 1 FirstName: expected 'John', got %q", records[1][1])
	}
	if records[1][3] != "john@example.com" {
		t.Errorf("row 1 Email: expected 'john@example.com', got %q", records[1][3])
	}
}

func TestFormatPeopleAsCSVEmpty(t *testing.T) {
	var people []types.Person

	var buf bytes.Buffer
	err := formatPeopleAsCSV(people, &buf)
	if err != nil {
		t.Fatalf("formatPeopleAsCSV failed: %v", err)
	}

	// Parse CSV output - should still have header
	reader := csv.NewReader(bytes.NewReader(buf.Bytes()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("output is not valid CSV: %v", err)
	}

	if len(records) != 1 {
		t.Errorf("expected 1 row (header only), got %d", len(records))
	}
}

func TestFormatPeopleAsJSON_Empty(t *testing.T) {
	var people []types.Person

	var buf bytes.Buffer
	err := formatPeopleAsJSON(people, &buf, "")
	if err != nil {
		t.Fatalf("formatPeopleAsJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var parsed []types.Person
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if len(parsed) != 0 {
		t.Errorf("expected 0 people, got %d", len(parsed))
	}
}

func TestFormatPeopleAsJSON_SinglePerson(t *testing.T) {
	people := []types.Person{
		{
			ID: "person-1",
			Name: types.Name{
				FirstName: "John",
				LastName:  "Doe",
			},
		},
	}

	var buf bytes.Buffer
	err := formatPeopleAsJSON(people, &buf, "")
	if err != nil {
		t.Fatalf("formatPeopleAsJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var parsed []types.Person
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if len(parsed) != 1 {
		t.Errorf("expected 1 person, got %d", len(parsed))
	}
}

func TestFormatPeopleAsCSV_SinglePerson(t *testing.T) {
	people := []types.Person{
		{
			ID: "person-1",
			Name: types.Name{
				FirstName: "John",
				LastName:  "Doe",
			},
			Email: types.Email{
				PrimaryEmail: "john@example.com",
			},
		},
	}

	var buf bytes.Buffer
	err := formatPeopleAsCSV(people, &buf)
	if err != nil {
		t.Fatalf("formatPeopleAsCSV failed: %v", err)
	}

	// Parse CSV output
	reader := csv.NewReader(bytes.NewReader(buf.Bytes()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("output is not valid CSV: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("expected 2 rows (header + 1 data), got %d", len(records))
	}
}

func TestFormatPeopleAsCSV_AllFieldsFilled(t *testing.T) {
	people := []types.Person{
		{
			ID: "person-1",
			Name: types.Name{
				FirstName: "John",
				LastName:  "Doe",
			},
			Email: types.Email{
				PrimaryEmail: "john@example.com",
			},
			Phone: types.Phone{
				PrimaryPhoneNumber: "+1234567890",
			},
			JobTitle: "Engineer",
			City:     "New York",
		},
	}

	var buf bytes.Buffer
	err := formatPeopleAsCSV(people, &buf)
	if err != nil {
		t.Fatalf("formatPeopleAsCSV failed: %v", err)
	}

	reader := csv.NewReader(bytes.NewReader(buf.Bytes()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("output is not valid CSV: %v", err)
	}

	// Check all columns in data row
	row := records[1]
	expectedValues := []string{"person-1", "John", "Doe", "john@example.com", "Engineer", "New York", "+1234567890"}
	for i, expected := range expectedValues {
		if row[i] != expected {
			t.Errorf("row[%d]: expected %q, got %q", i, expected, row[i])
		}
	}
}

func TestExportCmd_Flags(t *testing.T) {
	flags := []string{"format", "output", "all"}
	for _, flag := range flags {
		if exportCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestExportCmd_Use(t *testing.T) {
	if exportCmd.Use != "export" {
		t.Errorf("Use = %q, want %q", exportCmd.Use, "export")
	}
}

func TestExportCmd_Short(t *testing.T) {
	if exportCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestExportCmd_Long(t *testing.T) {
	if exportCmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestExportCmd_OutputFlagShorthand(t *testing.T) {
	flag := exportCmd.Flags().Lookup("output")
	if flag == nil {
		t.Fatal("output flag not registered")
	}
	if flag.Shorthand != "o" {
		t.Errorf("output flag shorthand = %q, want %q", flag.Shorthand, "o")
	}
}

func TestExportCmd_FormatDefaultValue(t *testing.T) {
	flag := exportCmd.Flags().Lookup("format")
	if flag == nil {
		t.Fatal("format flag not registered")
	}
	if flag.DefValue != "json" {
		t.Errorf("format flag default = %q, want %q", flag.DefValue, "json")
	}
}
