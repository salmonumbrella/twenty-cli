package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCompany_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"id": "company-123",
		"name": "Acme Corporation",
		"domainName": {
			"primaryLinkLabel": "Website",
			"primaryLinkUrl": "https://acme.com",
			"secondaryLinks": [
				{"url": "https://blog.acme.com", "label": "Blog"}
			]
		},
		"address": {
			"addressCity": "San Francisco",
			"addressStreet1": "123 Main St",
			"addressStreet2": "Suite 100",
			"addressState": "CA",
			"addressCountry": "USA",
			"addressPostcode": "94105"
		},
		"employees": 250,
		"linkedinLink": {
			"primaryLinkLabel": "LinkedIn",
			"primaryLinkUrl": "https://linkedin.com/company/acme",
			"secondaryLinks": []
		},
		"xLink": {
			"primaryLinkLabel": "X",
			"primaryLinkUrl": "https://x.com/acme",
			"secondaryLinks": []
		},
		"annualRevenue": 10000000,
		"idealCustomerProfile": true,
		"createdAt": "2024-01-15T10:30:00Z",
		"updatedAt": "2024-06-20T14:45:00Z"
	}`

	var company Company
	err := json.Unmarshal([]byte(jsonData), &company)
	if err != nil {
		t.Fatalf("failed to unmarshal Company: %v", err)
	}

	if company.ID != "company-123" {
		t.Errorf("expected ID='company-123', got %q", company.ID)
	}
	if company.Name != "Acme Corporation" {
		t.Errorf("expected Name='Acme Corporation', got %q", company.Name)
	}
	if company.DomainName.PrimaryLinkUrl != "https://acme.com" {
		t.Errorf("expected DomainName.PrimaryLinkUrl='https://acme.com', got %q", company.DomainName.PrimaryLinkUrl)
	}
	if len(company.DomainName.SecondaryLinks) != 1 {
		t.Errorf("expected 1 secondary link, got %d", len(company.DomainName.SecondaryLinks))
	}
	if company.Address.AddressCity != "San Francisco" {
		t.Errorf("expected AddressCity='San Francisco', got %q", company.Address.AddressCity)
	}
	if company.Address.AddressStreet1 != "123 Main St" {
		t.Errorf("expected AddressStreet1='123 Main St', got %q", company.Address.AddressStreet1)
	}
	if company.Address.AddressState != "CA" {
		t.Errorf("expected AddressState='CA', got %q", company.Address.AddressState)
	}
	if company.Employees == nil || *company.Employees != 250 {
		t.Errorf("expected Employees=250, got %v", company.Employees)
	}
	if company.AnnualRevenue == nil || *company.AnnualRevenue != 10000000 {
		t.Errorf("expected AnnualRevenue=10000000, got %v", company.AnnualRevenue)
	}
	if !company.IdealCustomer {
		t.Error("expected IdealCustomer=true")
	}
	if company.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if company.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestCompany_JSONMarshal(t *testing.T) {
	employees := 100
	revenue := 5000000
	company := Company{
		ID:   "company-789",
		Name: "Test Corp",
		DomainName: Link{
			PrimaryLinkLabel: "Website",
			PrimaryLinkUrl:   "https://testcorp.com",
			SecondaryLinks:   nil,
		},
		Address: Address{
			AddressCity:    "New York",
			AddressStreet1: "456 Park Ave",
			AddressState:   "NY",
			AddressCountry: "USA",
		},
		Employees:     &employees,
		AnnualRevenue: &revenue,
		IdealCustomer: false,
		CreatedAt:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2024, 6, 20, 14, 45, 0, 0, time.UTC),
	}

	data, err := json.Marshal(company)
	if err != nil {
		t.Fatalf("failed to marshal Company: %v", err)
	}

	// Round-trip: unmarshal back
	var parsed Company
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal serialized Company: %v", err)
	}

	if parsed.ID != company.ID {
		t.Errorf("round-trip failed: expected ID=%q, got %q", company.ID, parsed.ID)
	}
	if parsed.Name != company.Name {
		t.Errorf("round-trip failed: expected Name=%q, got %q", company.Name, parsed.Name)
	}
	if parsed.Address.AddressCity != company.Address.AddressCity {
		t.Errorf("round-trip failed: expected AddressCity=%q, got %q", company.Address.AddressCity, parsed.Address.AddressCity)
	}
	if parsed.Employees == nil || *parsed.Employees != *company.Employees {
		t.Errorf("round-trip failed: expected Employees=%d", *company.Employees)
	}
}

func TestCompany_EmployeesNil(t *testing.T) {
	jsonData := `{
		"id": "company-no-employees",
		"name": "Small Startup",
		"createdAt": "2024-01-15T10:30:00Z",
		"updatedAt": "2024-06-20T14:45:00Z"
	}`

	var company Company
	err := json.Unmarshal([]byte(jsonData), &company)
	if err != nil {
		t.Fatalf("failed to unmarshal Company: %v", err)
	}

	if company.Employees != nil {
		t.Errorf("expected Employees to be nil, got %v", *company.Employees)
	}
	if company.AnnualRevenue != nil {
		t.Errorf("expected AnnualRevenue to be nil, got %v", *company.AnnualRevenue)
	}
}

func TestCompany_EmployeesZero(t *testing.T) {
	jsonData := `{
		"id": "company-zero-employees",
		"name": "Shell Company",
		"employees": 0,
		"createdAt": "2024-01-15T10:30:00Z",
		"updatedAt": "2024-06-20T14:45:00Z"
	}`

	var company Company
	err := json.Unmarshal([]byte(jsonData), &company)
	if err != nil {
		t.Fatalf("failed to unmarshal Company: %v", err)
	}

	if company.Employees == nil {
		t.Fatal("expected Employees to be set")
	}
	if *company.Employees != 0 {
		t.Errorf("expected Employees=0, got %d", *company.Employees)
	}
}

func TestDomainName_Fields(t *testing.T) {
	tests := []struct {
		name              string
		json              string
		primaryLabel      string
		primaryUrl        string
		secondaryLinksLen int
	}{
		{
			name:              "full domain",
			json:              `{"primaryLinkLabel": "Main Site", "primaryLinkUrl": "https://example.com", "secondaryLinks": [{"url": "https://docs.example.com", "label": "Docs"}]}`,
			primaryLabel:      "Main Site",
			primaryUrl:        "https://example.com",
			secondaryLinksLen: 1,
		},
		{
			name:              "no secondary links",
			json:              `{"primaryLinkLabel": "Website", "primaryLinkUrl": "https://simple.com", "secondaryLinks": []}`,
			primaryLabel:      "Website",
			primaryUrl:        "https://simple.com",
			secondaryLinksLen: 0,
		},
		{
			name:              "multiple secondary links",
			json:              `{"primaryLinkLabel": "Home", "primaryLinkUrl": "https://multi.com", "secondaryLinks": [{"url": "https://a.com", "label": "A"}, {"url": "https://b.com", "label": "B"}, {"url": "https://c.com", "label": "C"}]}`,
			primaryLabel:      "Home",
			primaryUrl:        "https://multi.com",
			secondaryLinksLen: 3,
		},
		{
			name:              "empty fields",
			json:              `{"primaryLinkLabel": "", "primaryLinkUrl": "", "secondaryLinks": null}`,
			primaryLabel:      "",
			primaryUrl:        "",
			secondaryLinksLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var link Link
			if err := json.Unmarshal([]byte(tt.json), &link); err != nil {
				t.Fatalf("failed to unmarshal Link: %v", err)
			}
			if link.PrimaryLinkLabel != tt.primaryLabel {
				t.Errorf("expected PrimaryLinkLabel=%q, got %q", tt.primaryLabel, link.PrimaryLinkLabel)
			}
			if link.PrimaryLinkUrl != tt.primaryUrl {
				t.Errorf("expected PrimaryLinkUrl=%q, got %q", tt.primaryUrl, link.PrimaryLinkUrl)
			}
			if len(link.SecondaryLinks) != tt.secondaryLinksLen {
				t.Errorf("expected %d secondary links, got %d", tt.secondaryLinksLen, len(link.SecondaryLinks))
			}
		})
	}
}

func TestSecondaryLink_Fields(t *testing.T) {
	jsonData := `{"url": "https://blog.example.com", "label": "Blog"}`

	var sl SecondaryLink
	err := json.Unmarshal([]byte(jsonData), &sl)
	if err != nil {
		t.Fatalf("failed to unmarshal SecondaryLink: %v", err)
	}

	if sl.URL != "https://blog.example.com" {
		t.Errorf("expected URL='https://blog.example.com', got %q", sl.URL)
	}
	if sl.Label != "Blog" {
		t.Errorf("expected Label='Blog', got %q", sl.Label)
	}
}

func TestAddress_Fields(t *testing.T) {
	jsonData := `{
		"addressCity": "Chicago",
		"addressStreet1": "789 Lake Shore Dr",
		"addressStreet2": "Floor 10",
		"addressState": "IL",
		"addressCountry": "USA",
		"addressPostcode": "60611"
	}`

	var addr Address
	err := json.Unmarshal([]byte(jsonData), &addr)
	if err != nil {
		t.Fatalf("failed to unmarshal Address: %v", err)
	}

	if addr.AddressCity != "Chicago" {
		t.Errorf("expected AddressCity='Chicago', got %q", addr.AddressCity)
	}
	if addr.AddressStreet1 != "789 Lake Shore Dr" {
		t.Errorf("expected AddressStreet1='789 Lake Shore Dr', got %q", addr.AddressStreet1)
	}
	if addr.AddressStreet2 != "Floor 10" {
		t.Errorf("expected AddressStreet2='Floor 10', got %q", addr.AddressStreet2)
	}
	if addr.AddressState != "IL" {
		t.Errorf("expected AddressState='IL', got %q", addr.AddressState)
	}
	if addr.AddressCountry != "USA" {
		t.Errorf("expected AddressCountry='USA', got %q", addr.AddressCountry)
	}
	if addr.AddressPostcode != "60611" {
		t.Errorf("expected AddressPostcode='60611', got %q", addr.AddressPostcode)
	}
}

func TestCompaniesListResponse_JSONParsing(t *testing.T) {
	jsonData := `{
		"data": {
			"companies": [
				{
					"id": "c1",
					"name": "Company One",
					"createdAt": "2024-01-15T10:30:00Z",
					"updatedAt": "2024-06-20T14:45:00Z"
				},
				{
					"id": "c2",
					"name": "Company Two",
					"employees": 50,
					"createdAt": "2024-02-10T08:00:00Z",
					"updatedAt": "2024-05-15T12:00:00Z"
				}
			]
		},
		"totalCount": 2,
		"pageInfo": {
			"hasNextPage": true,
			"endCursor": "next-cursor"
		}
	}`

	var resp CompaniesListResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal CompaniesListResponse: %v", err)
	}

	if len(resp.Data.Companies) != 2 {
		t.Fatalf("expected 2 companies, got %d", len(resp.Data.Companies))
	}
	if resp.TotalCount != 2 {
		t.Errorf("expected TotalCount=2, got %d", resp.TotalCount)
	}
	if resp.Data.Companies[0].ID != "c1" {
		t.Errorf("expected first company ID='c1', got %q", resp.Data.Companies[0].ID)
	}
	if resp.Data.Companies[1].Employees == nil || *resp.Data.Companies[1].Employees != 50 {
		t.Error("expected second company to have 50 employees")
	}
	if resp.PageInfo == nil {
		t.Fatal("expected PageInfo to be set")
	}
	if !resp.PageInfo.HasNextPage {
		t.Error("expected HasNextPage=true")
	}
}

func TestCompanyResponse_JSONParsing(t *testing.T) {
	jsonData := `{
		"data": {
			"company": {
				"id": "company-single",
				"name": "Single Company",
				"employees": 75,
				"createdAt": "2024-01-15T10:30:00Z",
				"updatedAt": "2024-06-20T14:45:00Z"
			}
		}
	}`

	var resp CompanyResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal CompanyResponse: %v", err)
	}

	if resp.Data.Company.ID != "company-single" {
		t.Errorf("expected ID='company-single', got %q", resp.Data.Company.ID)
	}
	if resp.Data.Company.Name != "Single Company" {
		t.Errorf("expected Name='Single Company', got %q", resp.Data.Company.Name)
	}
	if resp.Data.Company.Employees == nil || *resp.Data.Company.Employees != 75 {
		t.Error("expected Employees=75")
	}
}

func TestCompany_WithRelations(t *testing.T) {
	jsonData := `{
		"id": "company-with-relations",
		"name": "Big Corp",
		"people": [
			{
				"id": "p1",
				"name": {"firstName": "John", "lastName": "Doe"},
				"emails": {"primaryEmail": "john@bigcorp.com"},
				"createdAt": "2024-01-15T10:30:00Z",
				"updatedAt": "2024-06-20T14:45:00Z"
			}
		],
		"opportunities": [
			{
				"id": "opp1",
				"name": "Big Deal",
				"createdAt": "2024-01-15T10:30:00Z",
				"updatedAt": "2024-06-20T14:45:00Z"
			}
		],
		"createdAt": "2024-01-15T10:30:00Z",
		"updatedAt": "2024-06-20T14:45:00Z"
	}`

	var company Company
	err := json.Unmarshal([]byte(jsonData), &company)
	if err != nil {
		t.Fatalf("failed to unmarshal Company with relations: %v", err)
	}

	if len(company.People) != 1 {
		t.Fatalf("expected 1 person relation, got %d", len(company.People))
	}
	if company.People[0].ID != "p1" {
		t.Errorf("expected person ID='p1', got %q", company.People[0].ID)
	}
	if len(company.Opportunities) != 1 {
		t.Fatalf("expected 1 opportunity relation, got %d", len(company.Opportunities))
	}
	if company.Opportunities[0].ID != "opp1" {
		t.Errorf("expected opportunity ID='opp1', got %q", company.Opportunities[0].ID)
	}
}

func TestCreateCompanyResponse_JSONParsing(t *testing.T) {
	jsonData := `{
		"data": {
			"createCompany": {
				"id": "new-company",
				"name": "New Corp",
				"createdAt": "2024-01-15T10:30:00Z",
				"updatedAt": "2024-06-20T14:45:00Z"
			}
		}
	}`

	var resp CreateCompanyResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal CreateCompanyResponse: %v", err)
	}

	if resp.Data.CreateCompany.ID != "new-company" {
		t.Errorf("expected ID='new-company', got %q", resp.Data.CreateCompany.ID)
	}
	if resp.Data.CreateCompany.Name != "New Corp" {
		t.Errorf("expected Name='New Corp', got %q", resp.Data.CreateCompany.Name)
	}
}

func TestUpdateCompanyResponse_JSONParsing(t *testing.T) {
	jsonData := `{
		"data": {
			"updateCompany": {
				"id": "updated-company",
				"name": "Updated Corp",
				"employees": 200,
				"createdAt": "2024-01-15T10:30:00Z",
				"updatedAt": "2024-06-20T14:45:00Z"
			}
		}
	}`

	var resp UpdateCompanyResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal UpdateCompanyResponse: %v", err)
	}

	if resp.Data.UpdateCompany.ID != "updated-company" {
		t.Errorf("expected ID='updated-company', got %q", resp.Data.UpdateCompany.ID)
	}
	if resp.Data.UpdateCompany.Name != "Updated Corp" {
		t.Errorf("expected Name='Updated Corp', got %q", resp.Data.UpdateCompany.Name)
	}
	if resp.Data.UpdateCompany.Employees == nil || *resp.Data.UpdateCompany.Employees != 200 {
		t.Error("expected Employees=200")
	}
}
