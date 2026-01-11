package people

import (
	"encoding/json"
	"testing"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func TestUpdateCmd_Flags(t *testing.T) {
	flags := []string{"first-name", "last-name", "email", "phone", "job-title", "data"}
	for _, flag := range flags {
		if updateCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestUpdateCmd_Use(t *testing.T) {
	if updateCmd.Use != "update <id>" {
		t.Errorf("Use = %q, want %q", updateCmd.Use, "update <id>")
	}
}

func TestUpdateCmd_Short(t *testing.T) {
	if updateCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestUpdateCmd_Args(t *testing.T) {
	// Command should require exactly 1 argument
	if updateCmd.Args == nil {
		t.Error("Args validator not set")
	}

	// Test with zero args
	err := updateCmd.Args(updateCmd, []string{})
	if err == nil {
		t.Error("Expected error with zero args")
	}

	// Test with one arg
	err = updateCmd.Args(updateCmd, []string{"id-123"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got %v", err)
	}

	// Test with two args
	err = updateCmd.Args(updateCmd, []string{"id-1", "id-2"})
	if err == nil {
		t.Error("Expected error with two args")
	}
}

func TestUpdateCmd_DataFlagShorthand(t *testing.T) {
	flag := updateCmd.Flags().Lookup("data")
	if flag == nil {
		t.Fatal("data flag not registered")
	}
	if flag.Shorthand != "d" {
		t.Errorf("data flag shorthand = %q, want %q", flag.Shorthand, "d")
	}
}

func TestUpdatePersonInput_FromJSON(t *testing.T) {
	jsonData := `{
		"name": {"firstName": "Jane", "lastName": "Smith"},
		"emails": {"primaryEmail": "jane@example.com"},
		"jobTitle": "Director"
	}`

	var input rest.UpdatePersonInput
	err := json.Unmarshal([]byte(jsonData), &input)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if input.Name == nil {
		t.Fatal("Name should not be nil")
	}
	if input.Name.FirstName != "Jane" {
		t.Errorf("FirstName = %q, want %q", input.Name.FirstName, "Jane")
	}
}

func TestUpdatePersonInput_InvalidJSON(t *testing.T) {
	invalidJSON := `{invalid json}`

	var input rest.UpdatePersonInput
	err := json.Unmarshal([]byte(invalidJSON), &input)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestUpdatePersonInput_EmptyInput(t *testing.T) {
	input := rest.UpdatePersonInput{}

	if input.Name != nil {
		t.Error("Name should be nil for empty input")
	}
	if input.Email != nil {
		t.Error("Email should be nil for empty input")
	}
	if input.Phone != nil {
		t.Error("Phone should be nil for empty input")
	}
	if input.JobTitle != nil {
		t.Error("JobTitle should be nil for empty input")
	}
}

func TestUpdatePersonInput_PartialUpdate(t *testing.T) {
	// Test updating only name
	input := rest.UpdatePersonInput{
		Name: &types.Name{
			FirstName: "NewFirst",
			LastName:  "NewLast",
		},
	}

	if input.Name == nil {
		t.Fatal("Name should not be nil")
	}
	if input.Name.FirstName != "NewFirst" {
		t.Errorf("FirstName = %q, want %q", input.Name.FirstName, "NewFirst")
	}
	if input.Email != nil {
		t.Error("Email should be nil when not updated")
	}
}

func TestUpdatePersonInput_EmailOnly(t *testing.T) {
	input := rest.UpdatePersonInput{
		Email: &types.Email{
			PrimaryEmail: "new@example.com",
		},
	}

	if input.Email == nil {
		t.Fatal("Email should not be nil")
	}
	if input.Email.PrimaryEmail != "new@example.com" {
		t.Errorf("Email = %q, want %q", input.Email.PrimaryEmail, "new@example.com")
	}
	if input.Name != nil {
		t.Error("Name should be nil when not updated")
	}
}

func TestUpdatePersonInput_PhoneOnly(t *testing.T) {
	input := rest.UpdatePersonInput{
		Phone: &types.Phone{
			PrimaryPhoneNumber: "+9999999999",
		},
	}

	if input.Phone == nil {
		t.Fatal("Phone should not be nil")
	}
	if input.Phone.PrimaryPhoneNumber != "+9999999999" {
		t.Errorf("Phone = %q, want %q", input.Phone.PrimaryPhoneNumber, "+9999999999")
	}
}

func TestUpdatePersonInput_JobTitleOnly(t *testing.T) {
	jobTitle := "Senior Engineer"
	input := rest.UpdatePersonInput{
		JobTitle: &jobTitle,
	}

	if input.JobTitle == nil {
		t.Fatal("JobTitle should not be nil")
	}
	if *input.JobTitle != "Senior Engineer" {
		t.Errorf("JobTitle = %q, want %q", *input.JobTitle, "Senior Engineer")
	}
}
