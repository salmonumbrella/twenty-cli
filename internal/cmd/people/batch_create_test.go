package people

import (
	"encoding/json"
	"testing"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func TestBatchCreateCmd_Flags(t *testing.T) {
	if batchCreateCmd.Flags().Lookup("file") == nil {
		t.Error("file flag not registered")
	}
}

func TestBatchCreateCmd_FileFlagShorthand(t *testing.T) {
	flag := batchCreateCmd.Flags().Lookup("file")
	if flag == nil {
		t.Fatal("file flag not registered")
	}
	if flag.Shorthand != "f" {
		t.Errorf("file flag shorthand = %q, want %q", flag.Shorthand, "f")
	}
}

func TestBatchCreateCmd_Use(t *testing.T) {
	if batchCreateCmd.Use != "batch-create" {
		t.Errorf("Use = %q, want %q", batchCreateCmd.Use, "batch-create")
	}
}

func TestBatchCreateCmd_Short(t *testing.T) {
	if batchCreateCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestBatchCreateCmd_Long(t *testing.T) {
	if batchCreateCmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestBatchCreateInput_ParseJSON(t *testing.T) {
	jsonData := `[
		{
			"name": {"firstName": "John", "lastName": "Doe"},
			"emails": {"primaryEmail": "john@example.com"},
			"jobTitle": "Engineer"
		},
		{
			"name": {"firstName": "Jane", "lastName": "Smith"},
			"emails": {"primaryEmail": "jane@example.com"},
			"jobTitle": "Manager"
		}
	]`

	var inputs []rest.CreatePersonInput
	err := json.Unmarshal([]byte(jsonData), &inputs)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(inputs) != 2 {
		t.Errorf("Expected 2 inputs, got %d", len(inputs))
	}

	if inputs[0].Name.FirstName != "John" {
		t.Errorf("FirstName = %q, want %q", inputs[0].Name.FirstName, "John")
	}
	if inputs[1].Name.FirstName != "Jane" {
		t.Errorf("FirstName = %q, want %q", inputs[1].Name.FirstName, "Jane")
	}
}

func TestBatchCreateInput_ParseJSONEmpty(t *testing.T) {
	jsonData := `[]`

	var inputs []rest.CreatePersonInput
	err := json.Unmarshal([]byte(jsonData), &inputs)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(inputs) != 0 {
		t.Errorf("Expected 0 inputs, got %d", len(inputs))
	}
}

func TestBatchCreateInput_InvalidJSON(t *testing.T) {
	invalidJSON := `{not an array}`

	var inputs []rest.CreatePersonInput
	err := json.Unmarshal([]byte(invalidJSON), &inputs)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestBatchCreateInput_WithAllFields(t *testing.T) {
	jsonData := `[
		{
			"name": {"firstName": "John", "lastName": "Doe"},
			"emails": {"primaryEmail": "john@example.com"},
			"phones": {"primaryPhoneNumber": "+1234567890"},
			"jobTitle": "Engineer",
			"city": "New York"
		}
	]`

	var inputs []rest.CreatePersonInput
	err := json.Unmarshal([]byte(jsonData), &inputs)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(inputs) != 1 {
		t.Fatalf("Expected 1 input, got %d", len(inputs))
	}

	input := inputs[0]
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

func TestBatchCreateResult_Structure(t *testing.T) {
	// Test the result structure
	result := map[string]interface{}{
		"created": []string{"id-1", "id-2"},
		"errors":  []string{},
	}

	created, ok := result["created"].([]string)
	if !ok {
		t.Fatal("created should be []string")
	}
	if len(created) != 2 {
		t.Errorf("Expected 2 created IDs, got %d", len(created))
	}

	errors, ok := result["errors"].([]string)
	if !ok {
		t.Fatal("errors should be []string")
	}
	if len(errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(errors))
	}
}

func TestBatchCreateResult_WithErrors(t *testing.T) {
	result := map[string]interface{}{
		"created": []string{"id-1"},
		"errors":  []string{"record 1: failed to create"},
	}

	created, ok := result["created"].([]string)
	if !ok {
		t.Fatal("created should be []string")
	}
	if len(created) != 1 {
		t.Errorf("Expected 1 created ID, got %d", len(created))
	}

	errors, ok := result["errors"].([]string)
	if !ok {
		t.Fatal("errors should be []string")
	}
	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors))
	}
}

func TestBatchCreateInput_MinimalData(t *testing.T) {
	input := rest.CreatePersonInput{
		Name: types.Name{
			FirstName: "Test",
		},
	}

	if input.Name.FirstName != "Test" {
		t.Errorf("FirstName = %q, want %q", input.Name.FirstName, "Test")
	}
	if input.Name.LastName != "" {
		t.Errorf("LastName should be empty, got %q", input.Name.LastName)
	}
}
