package people

import (
	"encoding/json"
	"testing"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func TestCreateCmd_Flags(t *testing.T) {
	flags := []string{"first-name", "last-name", "email", "phone", "job-title", "city", "data"}
	for _, flag := range flags {
		if createCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestCreateCmd_Use(t *testing.T) {
	if createCmd.Use != "create" {
		t.Errorf("Use = %q, want %q", createCmd.Use, "create")
	}
}

func TestCreateCmd_Short(t *testing.T) {
	if createCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestCreateCmd_DataFlagShorthand(t *testing.T) {
	flag := createCmd.Flags().Lookup("data")
	if flag == nil {
		t.Fatal("data flag not registered")
	}
	if flag.Shorthand != "d" {
		t.Errorf("data flag shorthand = %q, want %q", flag.Shorthand, "d")
	}
}

func TestCreatePersonInput_FromFlags(t *testing.T) {
	input := rest.CreatePersonInput{
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
	}

	if input.Name.FirstName != "John" {
		t.Errorf("FirstName = %q, want %q", input.Name.FirstName, "John")
	}
	if input.Name.LastName != "Doe" {
		t.Errorf("LastName = %q, want %q", input.Name.LastName, "Doe")
	}
	if input.Email.PrimaryEmail != "john@example.com" {
		t.Errorf("Email = %q, want %q", input.Email.PrimaryEmail, "john@example.com")
	}
	if input.Phone.PrimaryPhoneNumber != "+1234567890" {
		t.Errorf("Phone = %q, want %q", input.Phone.PrimaryPhoneNumber, "+1234567890")
	}
	if input.JobTitle != "Engineer" {
		t.Errorf("JobTitle = %q, want %q", input.JobTitle, "Engineer")
	}
	if input.City != "New York" {
		t.Errorf("City = %q, want %q", input.City, "New York")
	}
}

func TestCreatePersonInput_FromJSON(t *testing.T) {
	jsonData := `{
		"name": {"firstName": "Jane", "lastName": "Smith"},
		"emails": {"primaryEmail": "jane@example.com"},
		"phones": {"primaryPhoneNumber": "+0987654321"},
		"jobTitle": "Manager",
		"city": "Los Angeles"
	}`

	var input rest.CreatePersonInput
	err := json.Unmarshal([]byte(jsonData), &input)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if input.Name.FirstName != "Jane" {
		t.Errorf("FirstName = %q, want %q", input.Name.FirstName, "Jane")
	}
	if input.Name.LastName != "Smith" {
		t.Errorf("LastName = %q, want %q", input.Name.LastName, "Smith")
	}
	if input.Email.PrimaryEmail != "jane@example.com" {
		t.Errorf("Email = %q, want %q", input.Email.PrimaryEmail, "jane@example.com")
	}
}

func TestCreatePersonInput_InvalidJSON(t *testing.T) {
	invalidJSON := `{invalid json}`

	var input rest.CreatePersonInput
	err := json.Unmarshal([]byte(invalidJSON), &input)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestCreatePersonInput_EmptyFields(t *testing.T) {
	input := rest.CreatePersonInput{}

	if input.Name.FirstName != "" {
		t.Errorf("FirstName should be empty, got %q", input.Name.FirstName)
	}
	if input.Email.PrimaryEmail != "" {
		t.Errorf("Email should be empty, got %q", input.Email.PrimaryEmail)
	}
}

func TestCreatePersonInput_PartialData(t *testing.T) {
	jsonData := `{
		"name": {"firstName": "John"}
	}`

	var input rest.CreatePersonInput
	err := json.Unmarshal([]byte(jsonData), &input)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if input.Name.FirstName != "John" {
		t.Errorf("FirstName = %q, want %q", input.Name.FirstName, "John")
	}
	if input.Name.LastName != "" {
		t.Errorf("LastName should be empty, got %q", input.Name.LastName)
	}
}
