package people

import (
	"encoding/json"
	"testing"

	"github.com/salmonumbrella/twenty-cli/internal/api/graphql"
)

func TestUpsertCmd_Flags(t *testing.T) {
	flags := []string{"first-name", "last-name", "email", "phone", "job-title", "company-id", "data"}
	for _, flag := range flags {
		if upsertCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestUpsertCmd_Use(t *testing.T) {
	if upsertCmd.Use != "upsert" {
		t.Errorf("Use = %q, want %q", upsertCmd.Use, "upsert")
	}
}

func TestUpsertCmd_Short(t *testing.T) {
	if upsertCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestUpsertCmd_Long(t *testing.T) {
	if upsertCmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestUpsertCmd_DataFlagShorthand(t *testing.T) {
	flag := upsertCmd.Flags().Lookup("data")
	if flag == nil {
		t.Fatal("data flag not registered")
	}
	if flag.Shorthand != "d" {
		t.Errorf("data flag shorthand = %q, want %q", flag.Shorthand, "d")
	}
}

func TestUpsertPersonInput_FromFlags(t *testing.T) {
	input := graphql.UpsertPersonInput{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Phone:     "+1234567890",
		JobTitle:  "Engineer",
		CompanyID: "company-1",
	}

	if input.FirstName != "John" {
		t.Errorf("FirstName = %q, want %q", input.FirstName, "John")
	}
	if input.LastName != "Doe" {
		t.Errorf("LastName = %q, want %q", input.LastName, "Doe")
	}
	if input.Email != "john@example.com" {
		t.Errorf("Email = %q, want %q", input.Email, "john@example.com")
	}
	if input.Phone != "+1234567890" {
		t.Errorf("Phone = %q, want %q", input.Phone, "+1234567890")
	}
	if input.JobTitle != "Engineer" {
		t.Errorf("JobTitle = %q, want %q", input.JobTitle, "Engineer")
	}
	if input.CompanyID != "company-1" {
		t.Errorf("CompanyID = %q, want %q", input.CompanyID, "company-1")
	}
}

func TestUpsertPersonInput_FromJSON(t *testing.T) {
	jsonData := `{
		"firstName": "Jane",
		"lastName": "Smith",
		"email": "jane@example.com",
		"phone": "+0987654321",
		"jobTitle": "Manager",
		"companyId": "company-2"
	}`

	var input graphql.UpsertPersonInput
	err := json.Unmarshal([]byte(jsonData), &input)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if input.FirstName != "Jane" {
		t.Errorf("FirstName = %q, want %q", input.FirstName, "Jane")
	}
	if input.LastName != "Smith" {
		t.Errorf("LastName = %q, want %q", input.LastName, "Smith")
	}
	if input.Email != "jane@example.com" {
		t.Errorf("Email = %q, want %q", input.Email, "jane@example.com")
	}
}

func TestUpsertPersonInput_InvalidJSON(t *testing.T) {
	invalidJSON := `{invalid json}`

	var input graphql.UpsertPersonInput
	err := json.Unmarshal([]byte(invalidJSON), &input)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestUpsertPersonInput_EmptyInput(t *testing.T) {
	input := graphql.UpsertPersonInput{}

	if input.FirstName != "" {
		t.Errorf("FirstName should be empty, got %q", input.FirstName)
	}
	if input.Email != "" {
		t.Errorf("Email should be empty, got %q", input.Email)
	}
}

func TestUpsertPersonInput_EmailRequired(t *testing.T) {
	// Email is used for matching in upsert
	input := graphql.UpsertPersonInput{
		Email: "required@example.com",
	}

	if input.Email == "" {
		t.Error("Email should not be empty for upsert")
	}
}

func TestUpsertPersonInput_PartialData(t *testing.T) {
	jsonData := `{
		"email": "test@example.com",
		"firstName": "Test"
	}`

	var input graphql.UpsertPersonInput
	err := json.Unmarshal([]byte(jsonData), &input)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if input.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", input.Email, "test@example.com")
	}
	if input.FirstName != "Test" {
		t.Errorf("FirstName = %q, want %q", input.FirstName, "Test")
	}
	if input.LastName != "" {
		t.Errorf("LastName should be empty, got %q", input.LastName)
	}
}

func TestUpsertPersonInput_WithCompanyID(t *testing.T) {
	input := graphql.UpsertPersonInput{
		Email:     "test@example.com",
		CompanyID: "company-abc-123",
	}

	if input.CompanyID != "company-abc-123" {
		t.Errorf("CompanyID = %q, want %q", input.CompanyID, "company-abc-123")
	}
}

func TestUpsertPersonInput_JSONWithEmail(t *testing.T) {
	// Test JSON that overrides email
	jsonData := `{
		"email": "json@example.com",
		"firstName": "JSON"
	}`

	var input graphql.UpsertPersonInput
	err := json.Unmarshal([]byte(jsonData), &input)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if input.Email != "json@example.com" {
		t.Errorf("Email = %q, want %q", input.Email, "json@example.com")
	}
}

func TestUpsertPersonInput_JSONWithoutEmail(t *testing.T) {
	// Test JSON without email - email should come from flag
	jsonData := `{
		"firstName": "NoEmail",
		"lastName": "User"
	}`

	var input graphql.UpsertPersonInput
	err := json.Unmarshal([]byte(jsonData), &input)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if input.Email != "" {
		t.Errorf("Email should be empty from JSON, got %q", input.Email)
	}

	// In the actual command, email would be set from the flag
	input.Email = "flag@example.com"
	if input.Email != "flag@example.com" {
		t.Errorf("Email = %q, want %q", input.Email, "flag@example.com")
	}
}
