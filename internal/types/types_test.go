package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPerson_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"id": "person-123",
		"name": {
			"firstName": "John",
			"lastName": "Doe"
		},
		"emails": {
			"primaryEmail": "john@example.com",
			"additionalEmails": ["john.doe@work.com", "johnd@personal.com"]
		},
		"phones": {
			"primaryPhoneNumber": "+1-555-123-4567",
			"additionalPhoneNumbers": ["+1-555-987-6543"]
		},
		"jobTitle": "Software Engineer",
		"city": "San Francisco",
		"companyId": "company-456",
		"createdAt": "2024-01-15T10:30:00Z",
		"updatedAt": "2024-06-20T14:45:00Z"
	}`

	var person Person
	err := json.Unmarshal([]byte(jsonData), &person)
	if err != nil {
		t.Fatalf("failed to unmarshal Person: %v", err)
	}

	if person.ID != "person-123" {
		t.Errorf("expected ID='person-123', got %q", person.ID)
	}
	if person.Name.FirstName != "John" {
		t.Errorf("expected FirstName='John', got %q", person.Name.FirstName)
	}
	if person.Name.LastName != "Doe" {
		t.Errorf("expected LastName='Doe', got %q", person.Name.LastName)
	}
	if person.Email.PrimaryEmail != "john@example.com" {
		t.Errorf("expected PrimaryEmail='john@example.com', got %q", person.Email.PrimaryEmail)
	}
	if len(person.Email.AdditionalEmails) != 2 {
		t.Errorf("expected 2 additional emails, got %d", len(person.Email.AdditionalEmails))
	}
	if person.Phone.PrimaryPhoneNumber != "+1-555-123-4567" {
		t.Errorf("expected PrimaryPhoneNumber='+1-555-123-4567', got %q", person.Phone.PrimaryPhoneNumber)
	}
	if len(person.Phone.AdditionalPhoneNumbers) != 1 {
		t.Errorf("expected 1 additional phone, got %d", len(person.Phone.AdditionalPhoneNumbers))
	}
	if person.JobTitle != "Software Engineer" {
		t.Errorf("expected JobTitle='Software Engineer', got %q", person.JobTitle)
	}
	if person.City != "San Francisco" {
		t.Errorf("expected City='San Francisco', got %q", person.City)
	}
	if person.CompanyID != "company-456" {
		t.Errorf("expected CompanyID='company-456', got %q", person.CompanyID)
	}
	if person.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if person.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestPerson_JSONMarshal(t *testing.T) {
	person := Person{
		ID: "person-789",
		Name: Name{
			FirstName: "Jane",
			LastName:  "Smith",
		},
		Email: Email{
			PrimaryEmail:     "jane@example.com",
			AdditionalEmails: []string{"jane.work@corp.com"},
		},
		Phone: Phone{
			PrimaryPhoneNumber:     "+1-555-000-1111",
			AdditionalPhoneNumbers: nil,
		},
		JobTitle:  "Product Manager",
		City:      "New York",
		CompanyID: "company-abc",
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 6, 20, 14, 45, 0, 0, time.UTC),
	}

	data, err := json.Marshal(person)
	if err != nil {
		t.Fatalf("failed to marshal Person: %v", err)
	}

	// Round-trip: unmarshal back
	var parsed Person
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal serialized Person: %v", err)
	}

	if parsed.ID != person.ID {
		t.Errorf("round-trip failed: expected ID=%q, got %q", person.ID, parsed.ID)
	}
	if parsed.Name.FirstName != person.Name.FirstName {
		t.Errorf("round-trip failed: expected FirstName=%q, got %q", person.Name.FirstName, parsed.Name.FirstName)
	}
	if parsed.Name.LastName != person.Name.LastName {
		t.Errorf("round-trip failed: expected LastName=%q, got %q", person.Name.LastName, parsed.Name.LastName)
	}
	if parsed.Email.PrimaryEmail != person.Email.PrimaryEmail {
		t.Errorf("round-trip failed: expected PrimaryEmail=%q, got %q", person.Email.PrimaryEmail, parsed.Email.PrimaryEmail)
	}
	if len(parsed.Email.AdditionalEmails) != 1 {
		t.Errorf("round-trip failed: expected 1 additional email, got %d", len(parsed.Email.AdditionalEmails))
	}
	if parsed.JobTitle != person.JobTitle {
		t.Errorf("round-trip failed: expected JobTitle=%q, got %q", person.JobTitle, parsed.JobTitle)
	}
}

func TestName_FullName(t *testing.T) {
	tests := []struct {
		name     string
		input    Name
		wantJSON string
	}{
		{
			name: "both names",
			input: Name{
				FirstName: "John",
				LastName:  "Doe",
			},
			wantJSON: `{"firstName":"John","lastName":"Doe"}`,
		},
		{
			name: "first name only",
			input: Name{
				FirstName: "Jane",
				LastName:  "",
			},
			wantJSON: `{"firstName":"Jane","lastName":""}`,
		},
		{
			name: "last name only",
			input: Name{
				FirstName: "",
				LastName:  "Smith",
			},
			wantJSON: `{"firstName":"","lastName":"Smith"}`,
		},
		{
			name: "empty",
			input: Name{
				FirstName: "",
				LastName:  "",
			},
			wantJSON: `{"firstName":"","lastName":""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("failed to marshal Name: %v", err)
			}
			if string(data) != tt.wantJSON {
				t.Errorf("expected JSON=%s, got %s", tt.wantJSON, string(data))
			}

			// Verify round-trip
			var parsed Name
			if err := json.Unmarshal(data, &parsed); err != nil {
				t.Fatalf("failed to unmarshal Name: %v", err)
			}
			if parsed.FirstName != tt.input.FirstName {
				t.Errorf("round-trip failed: FirstName mismatch")
			}
			if parsed.LastName != tt.input.LastName {
				t.Errorf("round-trip failed: LastName mismatch")
			}
		})
	}
}

func TestListResponse_PageInfo(t *testing.T) {
	jsonData := `{
		"totalCount": 150,
		"pageInfo": {
			"hasNextPage": true,
			"endCursor": "cursor-abc123"
		}
	}`

	var resp ListResponse[Person]
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal ListResponse: %v", err)
	}

	if resp.TotalCount != 150 {
		t.Errorf("expected TotalCount=150, got %d", resp.TotalCount)
	}
	if resp.PageInfo == nil {
		t.Fatal("expected PageInfo to be set")
	}
	if !resp.PageInfo.HasNextPage {
		t.Error("expected HasNextPage=true")
	}
	if resp.PageInfo.EndCursor != "cursor-abc123" {
		t.Errorf("expected EndCursor='cursor-abc123', got %q", resp.PageInfo.EndCursor)
	}
}

func TestListResponse_NoPageInfo(t *testing.T) {
	jsonData := `{
		"totalCount": 5
	}`

	var resp ListResponse[Person]
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal ListResponse: %v", err)
	}

	if resp.TotalCount != 5 {
		t.Errorf("expected TotalCount=5, got %d", resp.TotalCount)
	}
	if resp.PageInfo != nil {
		t.Error("expected PageInfo to be nil")
	}
}

func TestPageInfo_JSONParsing(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		hasNextPage bool
		endCursor   string
	}{
		{
			name:        "has next page",
			json:        `{"hasNextPage": true, "endCursor": "next-cursor"}`,
			hasNextPage: true,
			endCursor:   "next-cursor",
		},
		{
			name:        "no next page",
			json:        `{"hasNextPage": false, "endCursor": ""}`,
			hasNextPage: false,
			endCursor:   "",
		},
		{
			name:        "last page with cursor",
			json:        `{"hasNextPage": false, "endCursor": "final-cursor"}`,
			hasNextPage: false,
			endCursor:   "final-cursor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pi PageInfo
			if err := json.Unmarshal([]byte(tt.json), &pi); err != nil {
				t.Fatalf("failed to unmarshal PageInfo: %v", err)
			}
			if pi.HasNextPage != tt.hasNextPage {
				t.Errorf("expected HasNextPage=%v, got %v", tt.hasNextPage, pi.HasNextPage)
			}
			if pi.EndCursor != tt.endCursor {
				t.Errorf("expected EndCursor=%q, got %q", tt.endCursor, pi.EndCursor)
			}
		})
	}
}

func TestPeopleListResponse_JSONParsing(t *testing.T) {
	jsonData := `{
		"data": {
			"people": [
				{
					"id": "p1",
					"name": {"firstName": "Alice", "lastName": "Johnson"},
					"emails": {"primaryEmail": "alice@example.com"},
					"createdAt": "2024-01-15T10:30:00Z",
					"updatedAt": "2024-06-20T14:45:00Z"
				},
				{
					"id": "p2",
					"name": {"firstName": "Bob", "lastName": "Williams"},
					"emails": {"primaryEmail": "bob@example.com"},
					"createdAt": "2024-02-10T08:00:00Z",
					"updatedAt": "2024-05-15T12:00:00Z"
				}
			]
		},
		"totalCount": 2,
		"pageInfo": {
			"hasNextPage": false,
			"endCursor": ""
		}
	}`

	var resp PeopleListResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal PeopleListResponse: %v", err)
	}

	if len(resp.Data.People) != 2 {
		t.Fatalf("expected 2 people, got %d", len(resp.Data.People))
	}
	if resp.TotalCount != 2 {
		t.Errorf("expected TotalCount=2, got %d", resp.TotalCount)
	}
	if resp.Data.People[0].ID != "p1" {
		t.Errorf("expected first person ID='p1', got %q", resp.Data.People[0].ID)
	}
	if resp.Data.People[0].Name.FirstName != "Alice" {
		t.Errorf("expected first person FirstName='Alice', got %q", resp.Data.People[0].Name.FirstName)
	}
	if resp.Data.People[1].ID != "p2" {
		t.Errorf("expected second person ID='p2', got %q", resp.Data.People[1].ID)
	}
}

func TestPersonResponse_JSONParsing(t *testing.T) {
	jsonData := `{
		"data": {
			"person": {
				"id": "person-single",
				"name": {"firstName": "Charlie", "lastName": "Brown"},
				"emails": {"primaryEmail": "charlie@example.com"},
				"jobTitle": "Designer",
				"createdAt": "2024-01-15T10:30:00Z",
				"updatedAt": "2024-06-20T14:45:00Z"
			}
		}
	}`

	var resp PersonResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal PersonResponse: %v", err)
	}

	if resp.Data.Person.ID != "person-single" {
		t.Errorf("expected ID='person-single', got %q", resp.Data.Person.ID)
	}
	if resp.Data.Person.Name.FirstName != "Charlie" {
		t.Errorf("expected FirstName='Charlie', got %q", resp.Data.Person.Name.FirstName)
	}
	if resp.Data.Person.JobTitle != "Designer" {
		t.Errorf("expected JobTitle='Designer', got %q", resp.Data.Person.JobTitle)
	}
}

func TestEmail_JSONParsing(t *testing.T) {
	tests := []struct {
		name            string
		json            string
		primaryEmail    string
		additionalCount int
	}{
		{
			name:            "primary only",
			json:            `{"primaryEmail": "test@example.com"}`,
			primaryEmail:    "test@example.com",
			additionalCount: 0,
		},
		{
			name:            "with additional",
			json:            `{"primaryEmail": "main@example.com", "additionalEmails": ["alt1@example.com", "alt2@example.com"]}`,
			primaryEmail:    "main@example.com",
			additionalCount: 2,
		},
		{
			name:            "empty additional",
			json:            `{"primaryEmail": "solo@example.com", "additionalEmails": []}`,
			primaryEmail:    "solo@example.com",
			additionalCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var email Email
			if err := json.Unmarshal([]byte(tt.json), &email); err != nil {
				t.Fatalf("failed to unmarshal Email: %v", err)
			}
			if email.PrimaryEmail != tt.primaryEmail {
				t.Errorf("expected PrimaryEmail=%q, got %q", tt.primaryEmail, email.PrimaryEmail)
			}
			if len(email.AdditionalEmails) != tt.additionalCount {
				t.Errorf("expected %d additional emails, got %d", tt.additionalCount, len(email.AdditionalEmails))
			}
		})
	}
}

func TestPhone_JSONParsing(t *testing.T) {
	tests := []struct {
		name            string
		json            string
		primaryPhone    string
		additionalCount int
	}{
		{
			name:            "primary only",
			json:            `{"primaryPhoneNumber": "+1-555-123-4567"}`,
			primaryPhone:    "+1-555-123-4567",
			additionalCount: 0,
		},
		{
			name:            "with additional",
			json:            `{"primaryPhoneNumber": "+1-555-000-0000", "additionalPhoneNumbers": ["+1-555-111-1111"]}`,
			primaryPhone:    "+1-555-000-0000",
			additionalCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var phone Phone
			if err := json.Unmarshal([]byte(tt.json), &phone); err != nil {
				t.Fatalf("failed to unmarshal Phone: %v", err)
			}
			if phone.PrimaryPhoneNumber != tt.primaryPhone {
				t.Errorf("expected PrimaryPhoneNumber=%q, got %q", tt.primaryPhone, phone.PrimaryPhoneNumber)
			}
			if len(phone.AdditionalPhoneNumbers) != tt.additionalCount {
				t.Errorf("expected %d additional phones, got %d", tt.additionalCount, len(phone.AdditionalPhoneNumbers))
			}
		})
	}
}

func TestPerson_WithCompanyRelation(t *testing.T) {
	jsonData := `{
		"id": "person-with-company",
		"name": {"firstName": "David", "lastName": "Lee"},
		"emails": {"primaryEmail": "david@example.com"},
		"companyId": "company-rel",
		"company": {
			"id": "company-rel",
			"name": "Acme Corp",
			"createdAt": "2024-01-01T00:00:00Z",
			"updatedAt": "2024-01-01T00:00:00Z"
		},
		"createdAt": "2024-01-15T10:30:00Z",
		"updatedAt": "2024-06-20T14:45:00Z"
	}`

	var person Person
	err := json.Unmarshal([]byte(jsonData), &person)
	if err != nil {
		t.Fatalf("failed to unmarshal Person with company: %v", err)
	}

	if person.Company == nil {
		t.Fatal("expected Company relation to be set")
	}
	if person.Company.ID != "company-rel" {
		t.Errorf("expected Company.ID='company-rel', got %q", person.Company.ID)
	}
	if person.Company.Name != "Acme Corp" {
		t.Errorf("expected Company.Name='Acme Corp', got %q", person.Company.Name)
	}
}

func TestPerson_WithoutCompanyRelation(t *testing.T) {
	jsonData := `{
		"id": "person-no-company",
		"name": {"firstName": "Eve", "lastName": "Wilson"},
		"emails": {"primaryEmail": "eve@example.com"},
		"createdAt": "2024-01-15T10:30:00Z",
		"updatedAt": "2024-06-20T14:45:00Z"
	}`

	var person Person
	err := json.Unmarshal([]byte(jsonData), &person)
	if err != nil {
		t.Fatalf("failed to unmarshal Person without company: %v", err)
	}

	if person.Company != nil {
		t.Error("expected Company to be nil")
	}
	if person.CompanyID != "" {
		t.Errorf("expected CompanyID to be empty, got %q", person.CompanyID)
	}
}
