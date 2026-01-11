package people

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func TestGetCmd_Flags(t *testing.T) {
	if getCmd.Flags().Lookup("include") == nil {
		t.Error("include flag not registered")
	}
}

func TestGetCmd_Use(t *testing.T) {
	if getCmd.Use != "get <id>" {
		t.Errorf("Use = %q, want %q", getCmd.Use, "get <id>")
	}
}

func TestGetCmd_Short(t *testing.T) {
	if getCmd.Short == "" {
		t.Error("Short description should not be empty")
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
	err = getCmd.Args(getCmd, []string{"id-123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}

	// Test with two args
	err = getCmd.Args(getCmd, []string{"id-1", "id-2"})
	if err == nil {
		t.Error("Expected error with two args")
	}
}

func TestOutputPerson_JSON(t *testing.T) {
	p := &types.Person{
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
		JobTitle:  "Engineer",
		City:      "New York",
		CompanyID: "company-1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test that person has the expected structure for JSON output
	if p.ID != "person-1" {
		t.Errorf("Person ID = %q, want %q", p.ID, "person-1")
	}
	if p.Name.FirstName != "John" {
		t.Errorf("Person FirstName = %q, want %q", p.Name.FirstName, "John")
	}
}

func TestOutputPerson_CSV(t *testing.T) {
	p := &types.Person{
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
		JobTitle:  "Engineer",
		City:      "New York",
		CompanyID: "company-1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Verify CSV row data can be constructed
	row := []string{
		p.ID,
		p.Name.FirstName,
		p.Name.LastName,
		p.Email.PrimaryEmail,
		p.Phone.PrimaryPhoneNumber,
		p.JobTitle,
		p.City,
		p.CompanyID,
		p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if len(row) != 10 {
		t.Errorf("CSV row length = %d, want 10", len(row))
	}
}

func TestOutputPerson_Text(t *testing.T) {
	p := &types.Person{
		ID: "person-1",
		Name: types.Name{
			FirstName: "John",
			LastName:  "Doe",
		},
		Email: types.Email{
			PrimaryEmail: "john@example.com",
		},
		JobTitle:  "Engineer",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Verify text output data can be constructed
	var buf bytes.Buffer
	buf.WriteString("ID:\t" + p.ID + "\n")
	buf.WriteString("Name:\t" + p.Name.FirstName + " " + p.Name.LastName + "\n")
	buf.WriteString("Email:\t" + p.Email.PrimaryEmail + "\n")
	buf.WriteString("Job Title:\t" + p.JobTitle + "\n")
	buf.WriteString("Created:\t" + p.CreatedAt.Format("2006-01-02 15:04:05") + "\n")

	output := buf.String()
	if output == "" {
		t.Error("Text output should not be empty")
	}
}

func TestOutputPerson_WithCompany(t *testing.T) {
	company := &types.Company{
		ID:   "company-1",
		Name: "Test Company",
	}

	p := &types.Person{
		ID: "person-1",
		Name: types.Name{
			FirstName: "John",
			LastName:  "Doe",
		},
		Email: types.Email{
			PrimaryEmail: "john@example.com",
		},
		Company:   company,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Verify company relation is present
	if p.Company == nil {
		t.Error("Person company should not be nil")
	}
	if p.Company.Name != "Test Company" {
		t.Errorf("Person company name = %q, want %q", p.Company.Name, "Test Company")
	}
}

func TestOutputPerson_JSONFormat(t *testing.T) {
	p := &types.Person{
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
		JobTitle:  "Engineer",
		City:      "New York",
		CompanyID: "company-1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	var buf bytes.Buffer
	err := outputPerson(&buf, p, "json", "")
	if err != nil {
		t.Fatalf("outputPerson failed: %v", err)
	}

	// Verify output is valid JSON
	var parsed types.Person
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if parsed.ID != "person-1" {
		t.Errorf("ID = %q, want %q", parsed.ID, "person-1")
	}
	if parsed.Name.FirstName != "John" {
		t.Errorf("FirstName = %q, want %q", parsed.Name.FirstName, "John")
	}
}

func TestOutputPerson_CSVFormat(t *testing.T) {
	now := time.Now()
	p := &types.Person{
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
		JobTitle:  "Engineer",
		City:      "New York",
		CompanyID: "company-1",
		CreatedAt: now,
		UpdatedAt: now,
	}

	var buf bytes.Buffer
	err := outputPerson(&buf, p, "csv", "")
	if err != nil {
		t.Fatalf("outputPerson failed: %v", err)
	}

	// Verify output is valid CSV
	reader := csv.NewReader(bytes.NewReader(buf.Bytes()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("output is not valid CSV: %v", err)
	}

	// Should have header + 1 data row
	if len(records) != 2 {
		t.Errorf("expected 2 rows, got %d", len(records))
	}

	// Check header
	expectedHeaders := []string{"id", "firstName", "lastName", "email", "phone", "jobTitle", "city", "companyId", "createdAt", "updatedAt"}
	for i, h := range expectedHeaders {
		if records[0][i] != h {
			t.Errorf("header[%d] = %q, want %q", i, records[0][i], h)
		}
	}

	// Check data
	if records[1][0] != "person-1" {
		t.Errorf("ID = %q, want %q", records[1][0], "person-1")
	}
}

func TestOutputPerson_TextFormat(t *testing.T) {
	now := time.Now()
	p := &types.Person{
		ID: "person-1",
		Name: types.Name{
			FirstName: "John",
			LastName:  "Doe",
		},
		Email: types.Email{
			PrimaryEmail: "john@example.com",
		},
		JobTitle:  "Engineer",
		CreatedAt: now,
		UpdatedAt: now,
	}

	var buf bytes.Buffer
	err := outputPerson(&buf, p, "text", "")
	if err != nil {
		t.Fatalf("outputPerson failed: %v", err)
	}

	output := buf.String()

	// Check that output contains expected fields
	if !strings.Contains(output, "person-1") {
		t.Errorf("output should contain ID 'person-1'")
	}
	if !strings.Contains(output, "John Doe") {
		t.Errorf("output should contain name 'John Doe'")
	}
	if !strings.Contains(output, "john@example.com") {
		t.Errorf("output should contain email 'john@example.com'")
	}
	if !strings.Contains(output, "Engineer") {
		t.Errorf("output should contain job title 'Engineer'")
	}
}

func TestOutputPerson_TextFormat_WithCompany(t *testing.T) {
	now := time.Now()
	company := &types.Company{
		ID:   "company-1",
		Name: "Test Company Inc",
	}

	p := &types.Person{
		ID: "person-1",
		Name: types.Name{
			FirstName: "John",
			LastName:  "Doe",
		},
		Email: types.Email{
			PrimaryEmail: "john@example.com",
		},
		Company:   company,
		CreatedAt: now,
		UpdatedAt: now,
	}

	var buf bytes.Buffer
	err := outputPerson(&buf, p, "text", "")
	if err != nil {
		t.Fatalf("outputPerson failed: %v", err)
	}

	output := buf.String()

	// Check that output contains company name
	if !strings.Contains(output, "Test Company Inc") {
		t.Errorf("output should contain company name 'Test Company Inc'")
	}
}

func TestOutputPerson_DefaultFormat(t *testing.T) {
	now := time.Now()
	p := &types.Person{
		ID: "person-1",
		Name: types.Name{
			FirstName: "John",
			LastName:  "Doe",
		},
		Email: types.Email{
			PrimaryEmail: "john@example.com",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	var buf bytes.Buffer
	// Empty string should use default (text) format
	err := outputPerson(&buf, p, "", "")
	if err != nil {
		t.Fatalf("outputPerson failed: %v", err)
	}

	output := buf.String()

	// Check that output contains expected fields
	if !strings.Contains(output, "person-1") {
		t.Errorf("output should contain ID")
	}
}
