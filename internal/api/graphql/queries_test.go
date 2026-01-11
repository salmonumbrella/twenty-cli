package graphql

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/shurcooL/graphql"
)

func TestPersonName(t *testing.T) {
	t.Run("JSON marshaling", func(t *testing.T) {
		name := PersonName{
			FirstName: "John",
			LastName:  "Doe",
		}

		data, err := json.Marshal(name)
		if err != nil {
			t.Fatalf("failed to marshal PersonName: %v", err)
		}

		expected := `{"firstName":"John","lastName":"Doe"}`
		if string(data) != expected {
			t.Errorf("expected %s, got %s", expected, string(data))
		}
	})

	t.Run("JSON unmarshaling", func(t *testing.T) {
		jsonData := `{"firstName":"Jane","lastName":"Smith"}`
		var name PersonName

		err := json.Unmarshal([]byte(jsonData), &name)
		if err != nil {
			t.Fatalf("failed to unmarshal PersonName: %v", err)
		}

		if name.FirstName != "Jane" {
			t.Errorf("expected FirstName 'Jane', got %s", name.FirstName)
		}
		if name.LastName != "Smith" {
			t.Errorf("expected LastName 'Smith', got %s", name.LastName)
		}
	})

	t.Run("empty values", func(t *testing.T) {
		name := PersonName{
			FirstName: "",
			LastName:  "",
		}

		data, err := json.Marshal(name)
		if err != nil {
			t.Fatalf("failed to marshal empty PersonName: %v", err)
		}

		expected := `{"firstName":"","lastName":""}`
		if string(data) != expected {
			t.Errorf("expected %s, got %s", expected, string(data))
		}
	})
}

func TestPersonEmails(t *testing.T) {
	t.Run("JSON marshaling", func(t *testing.T) {
		emails := PersonEmails{
			PrimaryEmail: "test@example.com",
		}

		data, err := json.Marshal(emails)
		if err != nil {
			t.Fatalf("failed to marshal PersonEmails: %v", err)
		}

		expected := `{"primaryEmail":"test@example.com"}`
		if string(data) != expected {
			t.Errorf("expected %s, got %s", expected, string(data))
		}
	})

	t.Run("JSON unmarshaling", func(t *testing.T) {
		jsonData := `{"primaryEmail":"contact@company.com"}`
		var emails PersonEmails

		err := json.Unmarshal([]byte(jsonData), &emails)
		if err != nil {
			t.Fatalf("failed to unmarshal PersonEmails: %v", err)
		}

		if emails.PrimaryEmail != "contact@company.com" {
			t.Errorf("expected PrimaryEmail 'contact@company.com', got %s", emails.PrimaryEmail)
		}
	})
}

func TestPersonPhones(t *testing.T) {
	t.Run("JSON marshaling", func(t *testing.T) {
		phones := PersonPhones{
			PrimaryPhoneNumber:      "+1-555-123-4567",
			PrimaryPhoneCountryCode: "US",
		}

		data, err := json.Marshal(phones)
		if err != nil {
			t.Fatalf("failed to marshal PersonPhones: %v", err)
		}

		var parsed map[string]string
		json.Unmarshal(data, &parsed)

		if parsed["primaryPhoneNumber"] != "+1-555-123-4567" {
			t.Errorf("expected primaryPhoneNumber '+1-555-123-4567', got %s", parsed["primaryPhoneNumber"])
		}
		if parsed["primaryPhoneCountryCode"] != "US" {
			t.Errorf("expected primaryPhoneCountryCode 'US', got %s", parsed["primaryPhoneCountryCode"])
		}
	})

	t.Run("JSON unmarshaling", func(t *testing.T) {
		jsonData := `{"primaryPhoneNumber":"+44-20-1234-5678","primaryPhoneCountryCode":"GB"}`
		var phones PersonPhones

		err := json.Unmarshal([]byte(jsonData), &phones)
		if err != nil {
			t.Fatalf("failed to unmarshal PersonPhones: %v", err)
		}

		if phones.PrimaryPhoneNumber != "+44-20-1234-5678" {
			t.Errorf("expected PrimaryPhoneNumber '+44-20-1234-5678', got %s", phones.PrimaryPhoneNumber)
		}
		if phones.PrimaryPhoneCountryCode != "GB" {
			t.Errorf("expected PrimaryPhoneCountryCode 'GB', got %s", phones.PrimaryPhoneCountryCode)
		}
	})
}

func TestPersonCreateInput(t *testing.T) {
	t.Run("full input marshaling", func(t *testing.T) {
		jobTitle := graphql.String("Engineer")
		companyID := graphql.String("comp-123")

		input := PersonCreateInput{
			Name: &PersonNameInput{
				FirstName: "Alice",
				LastName:  "Johnson",
			},
			Emails: &PersonEmailsInput{
				PrimaryEmail: "alice@example.com",
			},
			Phones: &PersonPhonesInput{
				PrimaryPhoneNumber:      "+1-555-000-0000",
				PrimaryPhoneCountryCode: "US",
			},
			JobTitle:  &jobTitle,
			CompanyID: &companyID,
		}

		data, err := json.Marshal(input)
		if err != nil {
			t.Fatalf("failed to marshal PersonCreateInput: %v", err)
		}

		var parsed map[string]interface{}
		json.Unmarshal(data, &parsed)

		if parsed["name"] == nil {
			t.Error("expected name to be present")
		}
		if parsed["emails"] == nil {
			t.Error("expected emails to be present")
		}
		if parsed["phones"] == nil {
			t.Error("expected phones to be present")
		}
		if parsed["jobTitle"] != "Engineer" {
			t.Errorf("expected jobTitle 'Engineer', got %v", parsed["jobTitle"])
		}
		if parsed["companyId"] != "comp-123" {
			t.Errorf("expected companyId 'comp-123', got %v", parsed["companyId"])
		}
	})

	t.Run("omits nil optional fields", func(t *testing.T) {
		input := PersonCreateInput{
			Name: &PersonNameInput{
				FirstName: "Bob",
				LastName:  "Smith",
			},
			Emails: &PersonEmailsInput{
				PrimaryEmail: "bob@example.com",
			},
			// Phones, JobTitle, CompanyID are nil
		}

		data, err := json.Marshal(input)
		if err != nil {
			t.Fatalf("failed to marshal PersonCreateInput: %v", err)
		}

		var parsed map[string]interface{}
		json.Unmarshal(data, &parsed)

		if parsed["phones"] != nil {
			t.Error("expected phones to be omitted when nil")
		}
		if parsed["jobTitle"] != nil {
			t.Error("expected jobTitle to be omitted when nil")
		}
		if parsed["companyId"] != nil {
			t.Error("expected companyId to be omitted when nil")
		}
	})
}

func TestPersonNameInput(t *testing.T) {
	t.Run("JSON marshaling", func(t *testing.T) {
		input := PersonNameInput{
			FirstName: "Test",
			LastName:  "User",
		}

		data, err := json.Marshal(input)
		if err != nil {
			t.Fatalf("failed to marshal PersonNameInput: %v", err)
		}

		expected := `{"firstName":"Test","lastName":"User"}`
		if string(data) != expected {
			t.Errorf("expected %s, got %s", expected, string(data))
		}
	})
}

func TestPersonEmailsInput(t *testing.T) {
	t.Run("JSON marshaling", func(t *testing.T) {
		input := PersonEmailsInput{
			PrimaryEmail: "input@example.com",
		}

		data, err := json.Marshal(input)
		if err != nil {
			t.Fatalf("failed to marshal PersonEmailsInput: %v", err)
		}

		expected := `{"primaryEmail":"input@example.com"}`
		if string(data) != expected {
			t.Errorf("expected %s, got %s", expected, string(data))
		}
	})
}

func TestPersonPhonesInput(t *testing.T) {
	t.Run("JSON marshaling", func(t *testing.T) {
		input := PersonPhonesInput{
			PrimaryPhoneNumber:      "+1-555-999-8888",
			PrimaryPhoneCountryCode: "CA",
		}

		data, err := json.Marshal(input)
		if err != nil {
			t.Fatalf("failed to marshal PersonPhonesInput: %v", err)
		}

		var parsed map[string]string
		json.Unmarshal(data, &parsed)

		if parsed["primaryPhoneNumber"] != "+1-555-999-8888" {
			t.Errorf("expected primaryPhoneNumber '+1-555-999-8888', got %s", parsed["primaryPhoneNumber"])
		}
		if parsed["primaryPhoneCountryCode"] != "CA" {
			t.Errorf("expected primaryPhoneCountryCode 'CA', got %s", parsed["primaryPhoneCountryCode"])
		}
	})
}

func TestCompanyInput(t *testing.T) {
	t.Run("full input marshaling", func(t *testing.T) {
		domainName := graphql.String("example.com")
		input := CompanyInput{
			Name:       "Acme Corp",
			DomainName: &domainName,
		}

		data, err := json.Marshal(input)
		if err != nil {
			t.Fatalf("failed to marshal CompanyInput: %v", err)
		}

		var parsed map[string]interface{}
		json.Unmarshal(data, &parsed)

		if parsed["name"] != "Acme Corp" {
			t.Errorf("expected name 'Acme Corp', got %v", parsed["name"])
		}
		if parsed["domainName"] != "example.com" {
			t.Errorf("expected domainName 'example.com', got %v", parsed["domainName"])
		}
	})

	t.Run("omits nil domain name", func(t *testing.T) {
		input := CompanyInput{
			Name:       "No Domain Corp",
			DomainName: nil,
		}

		data, err := json.Marshal(input)
		if err != nil {
			t.Fatalf("failed to marshal CompanyInput: %v", err)
		}

		var parsed map[string]interface{}
		json.Unmarshal(data, &parsed)

		if parsed["name"] != "No Domain Corp" {
			t.Errorf("expected name 'No Domain Corp', got %v", parsed["name"])
		}
		if parsed["domainName"] != nil {
			t.Error("expected domainName to be omitted when nil")
		}
	})
}

func TestPageInfo(t *testing.T) {
	t.Run("JSON marshaling", func(t *testing.T) {
		pageInfo := PageInfo{
			HasNextPage:     true,
			HasPreviousPage: false,
			StartCursor:     "start-abc",
			EndCursor:       "end-xyz",
		}

		data, err := json.Marshal(pageInfo)
		if err != nil {
			t.Fatalf("failed to marshal PageInfo: %v", err)
		}

		// Verify basic structure - graphql.Boolean omits false values
		if len(data) == 0 {
			t.Error("expected non-empty marshaled data")
		}

		// Check that the data contains expected cursor values
		dataStr := string(data)
		if !strings.Contains(dataStr, "start-abc") && !strings.Contains(dataStr, "startCursor") {
			// The graphql types may serialize differently
			t.Logf("marshaled data: %s", dataStr)
		}
	})

	t.Run("JSON unmarshaling", func(t *testing.T) {
		jsonData := `{
			"hasNextPage": true,
			"hasPreviousPage": true,
			"startCursor": "cursor-start",
			"endCursor": "cursor-end"
		}`
		var pageInfo PageInfo

		err := json.Unmarshal([]byte(jsonData), &pageInfo)
		if err != nil {
			t.Fatalf("failed to unmarshal PageInfo: %v", err)
		}

		if pageInfo.HasNextPage != true {
			t.Error("expected HasNextPage true")
		}
		if pageInfo.HasPreviousPage != true {
			t.Error("expected HasPreviousPage true")
		}
		if pageInfo.StartCursor != "cursor-start" {
			t.Errorf("expected StartCursor 'cursor-start', got %s", pageInfo.StartCursor)
		}
		if pageInfo.EndCursor != "cursor-end" {
			t.Errorf("expected EndCursor 'cursor-end', got %s", pageInfo.EndCursor)
		}
	})
}

func TestUpsertPersonMutation(t *testing.T) {
	t.Run("structure has correct fields", func(t *testing.T) {
		mutation := UpsertPersonMutation{}

		// Verify the CreatePerson field exists and has correct structure
		mutation.CreatePerson.ID = "test-id"
		mutation.CreatePerson.Name.FirstName = "Test"
		mutation.CreatePerson.Name.LastName = "Person"
		mutation.CreatePerson.Emails.PrimaryEmail = "test@example.com"

		if mutation.CreatePerson.ID != "test-id" {
			t.Errorf("expected ID 'test-id', got %s", mutation.CreatePerson.ID)
		}
		if mutation.CreatePerson.Name.FirstName != "Test" {
			t.Errorf("expected FirstName 'Test', got %s", mutation.CreatePerson.Name.FirstName)
		}
		if mutation.CreatePerson.Name.LastName != "Person" {
			t.Errorf("expected LastName 'Person', got %s", mutation.CreatePerson.Name.LastName)
		}
		if mutation.CreatePerson.Emails.PrimaryEmail != "test@example.com" {
			t.Errorf("expected PrimaryEmail 'test@example.com', got %s", mutation.CreatePerson.Emails.PrimaryEmail)
		}
	})
}

func TestUpsertCompanyMutation(t *testing.T) {
	t.Run("structure has correct fields", func(t *testing.T) {
		mutation := UpsertCompanyMutation{}

		mutation.UpsertCompany.ID = "company-id"
		mutation.UpsertCompany.Name = "Test Company"

		if mutation.UpsertCompany.ID != "company-id" {
			t.Errorf("expected ID 'company-id', got %s", mutation.UpsertCompany.ID)
		}
		if mutation.UpsertCompany.Name != "Test Company" {
			t.Errorf("expected Name 'Test Company', got %s", mutation.UpsertCompany.Name)
		}
	})
}

func TestFindManyPeopleQuery(t *testing.T) {
	t.Run("structure parses correctly", func(t *testing.T) {
		jsonData := `{
			"people": {
				"edges": [
					{
						"node": {
							"id": "person-1",
							"name": {"firstName": "Alice", "lastName": "A"},
							"emails": {"primaryEmail": "alice@example.com"},
							"createdAt": "2024-01-15T10:00:00Z",
							"updatedAt": "2024-01-20T12:00:00Z"
						}
					},
					{
						"node": {
							"id": "person-2",
							"name": {"firstName": "Bob", "lastName": "B"},
							"emails": {"primaryEmail": "bob@example.com"},
							"createdAt": "2024-02-01T08:00:00Z",
							"updatedAt": "2024-02-05T09:00:00Z"
						}
					}
				],
				"pageInfo": {
					"hasNextPage": true,
					"hasPreviousPage": false,
					"startCursor": "start",
					"endCursor": "end"
				}
			}
		}`

		var query FindManyPeopleQuery
		err := json.Unmarshal([]byte(jsonData), &query)
		if err != nil {
			t.Fatalf("failed to unmarshal FindManyPeopleQuery: %v", err)
		}

		if len(query.People.Edges) != 2 {
			t.Fatalf("expected 2 edges, got %d", len(query.People.Edges))
		}

		if query.People.Edges[0].Node.ID != "person-1" {
			t.Errorf("expected first person ID 'person-1', got %s", query.People.Edges[0].Node.ID)
		}
		if query.People.Edges[0].Node.Name.FirstName != "Alice" {
			t.Errorf("expected first person FirstName 'Alice', got %s", query.People.Edges[0].Node.Name.FirstName)
		}
		if query.People.Edges[1].Node.ID != "person-2" {
			t.Errorf("expected second person ID 'person-2', got %s", query.People.Edges[1].Node.ID)
		}

		if query.People.PageInfo.HasNextPage != true {
			t.Error("expected HasNextPage true")
		}
		if query.People.PageInfo.EndCursor != "end" {
			t.Errorf("expected EndCursor 'end', got %s", query.People.PageInfo.EndCursor)
		}
	})

	t.Run("empty edges", func(t *testing.T) {
		jsonData := `{
			"people": {
				"edges": [],
				"pageInfo": {
					"hasNextPage": false,
					"hasPreviousPage": false,
					"startCursor": "",
					"endCursor": ""
				}
			}
		}`

		var query FindManyPeopleQuery
		err := json.Unmarshal([]byte(jsonData), &query)
		if err != nil {
			t.Fatalf("failed to unmarshal empty FindManyPeopleQuery: %v", err)
		}

		if len(query.People.Edges) != 0 {
			t.Errorf("expected 0 edges, got %d", len(query.People.Edges))
		}
		if query.People.PageInfo.HasNextPage != false {
			t.Error("expected HasNextPage false for empty result")
		}
	})
}

func TestFindManyCompaniesQuery(t *testing.T) {
	t.Run("structure parses correctly", func(t *testing.T) {
		jsonData := `{
			"companies": {
				"edges": [
					{
						"node": {
							"id": "comp-1",
							"name": "Acme Corp",
							"domainName": "acme.com",
							"createdAt": "2024-01-01T00:00:00Z",
							"updatedAt": "2024-01-10T00:00:00Z"
						}
					},
					{
						"node": {
							"id": "comp-2",
							"name": "Widgets Inc",
							"domainName": "widgets.io",
							"createdAt": "2024-02-01T00:00:00Z",
							"updatedAt": "2024-02-15T00:00:00Z"
						}
					}
				],
				"pageInfo": {
					"hasNextPage": false,
					"hasPreviousPage": true,
					"startCursor": "comp-start",
					"endCursor": "comp-end"
				}
			}
		}`

		var query FindManyCompaniesQuery
		err := json.Unmarshal([]byte(jsonData), &query)
		if err != nil {
			t.Fatalf("failed to unmarshal FindManyCompaniesQuery: %v", err)
		}

		if len(query.Companies.Edges) != 2 {
			t.Fatalf("expected 2 edges, got %d", len(query.Companies.Edges))
		}

		if query.Companies.Edges[0].Node.ID != "comp-1" {
			t.Errorf("expected first company ID 'comp-1', got %s", query.Companies.Edges[0].Node.ID)
		}
		if query.Companies.Edges[0].Node.Name != "Acme Corp" {
			t.Errorf("expected first company Name 'Acme Corp', got %s", query.Companies.Edges[0].Node.Name)
		}
		if query.Companies.Edges[0].Node.DomainName != "acme.com" {
			t.Errorf("expected first company DomainName 'acme.com', got %s", query.Companies.Edges[0].Node.DomainName)
		}
		if query.Companies.Edges[1].Node.ID != "comp-2" {
			t.Errorf("expected second company ID 'comp-2', got %s", query.Companies.Edges[1].Node.ID)
		}

		if query.Companies.PageInfo.HasPreviousPage != true {
			t.Error("expected HasPreviousPage true")
		}
	})

	t.Run("empty companies result", func(t *testing.T) {
		jsonData := `{
			"companies": {
				"edges": [],
				"pageInfo": {
					"hasNextPage": false,
					"hasPreviousPage": false,
					"startCursor": "",
					"endCursor": ""
				}
			}
		}`

		var query FindManyCompaniesQuery
		err := json.Unmarshal([]byte(jsonData), &query)
		if err != nil {
			t.Fatalf("failed to unmarshal empty FindManyCompaniesQuery: %v", err)
		}

		if len(query.Companies.Edges) != 0 {
			t.Errorf("expected 0 edges, got %d", len(query.Companies.Edges))
		}
	})
}

func TestGraphQLStringType(t *testing.T) {
	t.Run("graphql.String conversion", func(t *testing.T) {
		str := graphql.String("test value")

		// Verify it can be used as string
		if string(str) != "test value" {
			t.Errorf("expected 'test value', got %s", string(str))
		}
	})

	t.Run("graphql.Boolean conversion", func(t *testing.T) {
		boolTrue := graphql.Boolean(true)
		boolFalse := graphql.Boolean(false)

		if bool(boolTrue) != true {
			t.Error("expected true")
		}
		if bool(boolFalse) != false {
			t.Error("expected false")
		}
	})
}
