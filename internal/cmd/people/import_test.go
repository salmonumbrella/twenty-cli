package people

import (
	"strings"
	"testing"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
)

func TestParseJSONInput(t *testing.T) {
	jsonInput := `[
		{
			"name": {"firstName": "John", "lastName": "Doe"},
			"emails": {"primaryEmail": "john@example.com"},
			"phones": {"primaryPhoneNumber": "+1234567890"},
			"jobTitle": "Engineer",
			"city": "New York"
		},
		{
			"name": {"firstName": "Jane", "lastName": "Smith"},
			"emails": {"primaryEmail": "jane@example.com"},
			"jobTitle": "Manager"
		}
	]`

	reader := strings.NewReader(jsonInput)
	inputs, err := parseJSONInput(reader)
	if err != nil {
		t.Fatalf("parseJSONInput failed: %v", err)
	}

	if len(inputs) != 2 {
		t.Errorf("expected 2 inputs, got %d", len(inputs))
	}

	if inputs[0].Name.FirstName != "John" {
		t.Errorf("input[0].Name.FirstName: expected 'John', got %q", inputs[0].Name.FirstName)
	}
	if inputs[0].Email.PrimaryEmail != "john@example.com" {
		t.Errorf("input[0].Email.PrimaryEmail: expected 'john@example.com', got %q", inputs[0].Email.PrimaryEmail)
	}
	if inputs[0].Phone.PrimaryPhoneNumber != "+1234567890" {
		t.Errorf("input[0].Phone.PrimaryPhoneNumber: expected '+1234567890', got %q", inputs[0].Phone.PrimaryPhoneNumber)
	}

	if inputs[1].Name.FirstName != "Jane" {
		t.Errorf("input[1].Name.FirstName: expected 'Jane', got %q", inputs[1].Name.FirstName)
	}
}

func TestParseCSVInput(t *testing.T) {
	csvInput := `FirstName,LastName,Email,JobTitle,City,Phone
John,Doe,john@example.com,Engineer,New York,+1234567890
Jane,Smith,jane@example.com,Manager,Los Angeles,+0987654321`

	reader := strings.NewReader(csvInput)
	inputs, err := parseCSVInput(reader)
	if err != nil {
		t.Fatalf("parseCSVInput failed: %v", err)
	}

	if len(inputs) != 2 {
		t.Errorf("expected 2 inputs, got %d", len(inputs))
	}

	// Check first record
	if inputs[0].Name.FirstName != "John" {
		t.Errorf("input[0].Name.FirstName: expected 'John', got %q", inputs[0].Name.FirstName)
	}
	if inputs[0].Name.LastName != "Doe" {
		t.Errorf("input[0].Name.LastName: expected 'Doe', got %q", inputs[0].Name.LastName)
	}
	if inputs[0].Email.PrimaryEmail != "john@example.com" {
		t.Errorf("input[0].Email.PrimaryEmail: expected 'john@example.com', got %q", inputs[0].Email.PrimaryEmail)
	}
	if inputs[0].Phone.PrimaryPhoneNumber != "+1234567890" {
		t.Errorf("input[0].Phone.PrimaryPhoneNumber: expected '+1234567890', got %q", inputs[0].Phone.PrimaryPhoneNumber)
	}

	// Check second record
	if inputs[1].Name.FirstName != "Jane" {
		t.Errorf("input[1].Name.FirstName: expected 'Jane', got %q", inputs[1].Name.FirstName)
	}
	if inputs[1].City != "Los Angeles" {
		t.Errorf("input[1].City: expected 'Los Angeles', got %q", inputs[1].City)
	}
}

func TestParseCSVInputWithAlternateHeaders(t *testing.T) {
	// Test case-insensitive headers
	csvInput := `firstname,lastname,email,jobtitle,city,phone
John,Doe,john@example.com,Engineer,New York,+1234567890`

	reader := strings.NewReader(csvInput)
	inputs, err := parseCSVInput(reader)
	if err != nil {
		t.Fatalf("parseCSVInput failed: %v", err)
	}

	if len(inputs) != 1 {
		t.Errorf("expected 1 input, got %d", len(inputs))
	}

	if inputs[0].Name.FirstName != "John" {
		t.Errorf("input[0].Name.FirstName: expected 'John', got %q", inputs[0].Name.FirstName)
	}
}

func TestParseCSVInputEmpty(t *testing.T) {
	csvInput := `FirstName,LastName,Email,JobTitle,City,Phone`

	reader := strings.NewReader(csvInput)
	inputs, err := parseCSVInput(reader)
	if err != nil {
		t.Fatalf("parseCSVInput failed: %v", err)
	}

	if len(inputs) != 0 {
		t.Errorf("expected 0 inputs, got %d", len(inputs))
	}
}

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		filename string
		explicit string
		expected string
	}{
		{"people.json", "", "json"},
		{"people.csv", "", "csv"},
		{"people.JSON", "", "json"},
		{"people.CSV", "", "csv"},
		{"people.txt", "json", "json"},
		{"people.txt", "csv", "csv"},
		{"people", "json", "json"},
	}

	for _, tt := range tests {
		result := detectFormat(tt.filename, tt.explicit)
		if result != tt.expected {
			t.Errorf("detectFormat(%q, %q): expected %q, got %q", tt.filename, tt.explicit, tt.expected, result)
		}
	}
}

func TestBuildImportSummary(t *testing.T) {
	inputs := []rest.CreatePersonInput{
		{
			Name:     rest.CreatePersonInput{}.Name,
			JobTitle: "Engineer",
		},
		{
			Name:     rest.CreatePersonInput{}.Name,
			JobTitle: "Manager",
		},
	}
	inputs[0].Name.FirstName = "John"
	inputs[0].Name.LastName = "Doe"
	inputs[1].Name.FirstName = "Jane"
	inputs[1].Name.LastName = "Smith"

	summary := buildImportSummary(inputs)

	if !strings.Contains(summary, "2 people") {
		t.Errorf("summary should contain '2 people': %s", summary)
	}
	if !strings.Contains(summary, "John Doe") {
		t.Errorf("summary should contain 'John Doe': %s", summary)
	}
	if !strings.Contains(summary, "Jane Smith") {
		t.Errorf("summary should contain 'Jane Smith': %s", summary)
	}
}

func TestBuildImportSummary_NoName(t *testing.T) {
	inputs := []rest.CreatePersonInput{
		{
			JobTitle: "Engineer",
		},
	}

	summary := buildImportSummary(inputs)

	if !strings.Contains(summary, "(no name)") {
		t.Errorf("summary should contain '(no name)': %s", summary)
	}
}

func TestBuildImportSummary_NoEmail(t *testing.T) {
	inputs := []rest.CreatePersonInput{
		{
			Name: rest.CreatePersonInput{}.Name,
		},
	}
	inputs[0].Name.FirstName = "John"
	inputs[0].Name.LastName = "Doe"

	summary := buildImportSummary(inputs)

	if !strings.Contains(summary, "(no email)") {
		t.Errorf("summary should contain '(no email)': %s", summary)
	}
}

func TestBuildImportSummary_WithEmail(t *testing.T) {
	inputs := []rest.CreatePersonInput{
		{
			Name:  rest.CreatePersonInput{}.Name,
			Email: rest.CreatePersonInput{}.Email,
		},
	}
	inputs[0].Name.FirstName = "John"
	inputs[0].Name.LastName = "Doe"
	inputs[0].Email.PrimaryEmail = "john@example.com"

	summary := buildImportSummary(inputs)

	if !strings.Contains(summary, "john@example.com") {
		t.Errorf("summary should contain email: %s", summary)
	}
	if strings.Contains(summary, "(no email)") {
		t.Errorf("summary should not contain '(no email)': %s", summary)
	}
}

func TestBuildImportSummary_Empty(t *testing.T) {
	inputs := []rest.CreatePersonInput{}

	summary := buildImportSummary(inputs)

	if !strings.Contains(summary, "0 people") {
		t.Errorf("summary should contain '0 people': %s", summary)
	}
}

func TestDetectFormat_UnknownExtension(t *testing.T) {
	result := detectFormat("people.txt", "")
	if result != "" {
		t.Errorf("detectFormat should return empty for unknown extension, got %q", result)
	}
}

func TestDetectFormat_ExplicitOverridesExtension(t *testing.T) {
	result := detectFormat("people.csv", "json")
	if result != "json" {
		t.Errorf("detectFormat should return explicit format, got %q", result)
	}
}

func TestDetectFormat_ExplicitUppercase(t *testing.T) {
	result := detectFormat("people.txt", "JSON")
	if result != "json" {
		t.Errorf("detectFormat should lowercase explicit format, got %q", result)
	}
}

func TestParseJSONInput_InvalidJSON(t *testing.T) {
	invalidJSON := `{not valid json`

	reader := strings.NewReader(invalidJSON)
	_, err := parseJSONInput(reader)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseJSONInput_EmptyArray(t *testing.T) {
	jsonInput := `[]`

	reader := strings.NewReader(jsonInput)
	inputs, err := parseJSONInput(reader)
	if err != nil {
		t.Fatalf("parseJSONInput failed: %v", err)
	}

	if len(inputs) != 0 {
		t.Errorf("expected 0 inputs, got %d", len(inputs))
	}
}

func TestParseCSVInput_EmptyFile(t *testing.T) {
	csvInput := ``

	reader := strings.NewReader(csvInput)
	_, err := parseCSVInput(reader)
	if err == nil {
		t.Error("expected error for empty CSV")
	}
}

func TestParseCSVInput_WhitespaceHeaders(t *testing.T) {
	csvInput := ` FirstName , LastName , Email
John,Doe,john@example.com`

	reader := strings.NewReader(csvInput)
	inputs, err := parseCSVInput(reader)
	if err != nil {
		t.Fatalf("parseCSVInput failed: %v", err)
	}

	if len(inputs) != 1 {
		t.Errorf("expected 1 input, got %d", len(inputs))
	}

	if inputs[0].Name.FirstName != "John" {
		t.Errorf("expected FirstName 'John', got %q", inputs[0].Name.FirstName)
	}
}

func TestParseCSVInput_ExtraColumns(t *testing.T) {
	csvInput := `FirstName,LastName,Email,ExtraColumn,AnotherExtra
John,Doe,john@example.com,extra1,extra2`

	reader := strings.NewReader(csvInput)
	inputs, err := parseCSVInput(reader)
	if err != nil {
		t.Fatalf("parseCSVInput failed: %v", err)
	}

	if len(inputs) != 1 {
		t.Errorf("expected 1 input, got %d", len(inputs))
	}

	// Extra columns should be ignored
	if inputs[0].Name.FirstName != "John" {
		t.Errorf("expected FirstName 'John', got %q", inputs[0].Name.FirstName)
	}
}

func TestParseCSVInput_MissingColumns(t *testing.T) {
	csvInput := `FirstName,Email
John,john@example.com`

	reader := strings.NewReader(csvInput)
	inputs, err := parseCSVInput(reader)
	if err != nil {
		t.Fatalf("parseCSVInput failed: %v", err)
	}

	if len(inputs) != 1 {
		t.Errorf("expected 1 input, got %d", len(inputs))
	}

	if inputs[0].Name.FirstName != "John" {
		t.Errorf("expected FirstName 'John', got %q", inputs[0].Name.FirstName)
	}
	// LastName should be empty since column is missing
	if inputs[0].Name.LastName != "" {
		t.Errorf("expected LastName empty, got %q", inputs[0].Name.LastName)
	}
}

func TestImportCmd_Flags(t *testing.T) {
	flags := []string{"format", "dry-run"}
	for _, flag := range flags {
		if importCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestImportCmd_Use(t *testing.T) {
	if importCmd.Use != "import <file>" {
		t.Errorf("Use = %q, want %q", importCmd.Use, "import <file>")
	}
}

func TestImportCmd_Short(t *testing.T) {
	if importCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestImportCmd_Long(t *testing.T) {
	if importCmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestImportCmd_Args(t *testing.T) {
	// Command should require exactly 1 argument
	if importCmd.Args == nil {
		t.Error("Args validator not set")
	}

	// Test with zero args
	err := importCmd.Args(importCmd, []string{})
	if err == nil {
		t.Error("Expected error with zero args")
	}

	// Test with one arg
	err = importCmd.Args(importCmd, []string{"file.json"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}

	// Test with two args
	err = importCmd.Args(importCmd, []string{"file1.json", "file2.json"})
	if err == nil {
		t.Error("Expected error with two args")
	}
}
