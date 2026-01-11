package rest

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func TestClient_ListCompanies(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	employees50 := 50
	employees200 := 200
	expectedCompanies := []types.Company{
		{
			ID:   "company-1",
			Name: "Acme Corp",
			DomainName: types.Link{
				PrimaryLinkLabel: "Website",
				PrimaryLinkUrl:   "https://acme.com",
			},
			Employees:     &employees50,
			IdealCustomer: true,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:   "company-2",
			Name: "Globex Inc",
			DomainName: types.Link{
				PrimaryLinkLabel: "Website",
				PrimaryLinkUrl:   "https://globex.com",
			},
			Employees:     &employees200,
			IdealCustomer: false,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/rest/companies" {
			t.Errorf("expected path /rest/companies, got %s", r.URL.Path)
		}

		resp := types.CompaniesListResponse{
			TotalCount: 2,
			PageInfo:   &types.PageInfo{HasNextPage: true, EndCursor: "company-cursor"},
		}
		resp.Data.Companies = expectedCompanies

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	result, err := client.ListCompanies(context.Background(), nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("expected TotalCount 2, got %d", result.TotalCount)
	}
	if len(result.Data) != 2 {
		t.Fatalf("expected 2 companies, got %d", len(result.Data))
	}
	if result.Data[0].ID != "company-1" {
		t.Errorf("expected first company ID 'company-1', got %s", result.Data[0].ID)
	}
	if result.Data[0].Name != "Acme Corp" {
		t.Errorf("expected first company name 'Acme Corp', got %s", result.Data[0].Name)
	}
	if result.Data[0].DomainName.PrimaryLinkUrl != "https://acme.com" {
		t.Errorf("expected domain URL 'https://acme.com', got %s", result.Data[0].DomainName.PrimaryLinkUrl)
	}
	if result.Data[0].Employees == nil || *result.Data[0].Employees != 50 {
		t.Errorf("expected 50 employees, got %v", result.Data[0].Employees)
	}
	if !result.Data[0].IdealCustomer {
		t.Error("expected IdealCustomer to be true for first company")
	}
	if result.PageInfo == nil {
		t.Fatal("expected PageInfo to be set")
	}
	if !result.PageInfo.HasNextPage {
		t.Error("expected HasNextPage to be true")
	}
	if result.PageInfo.EndCursor != "company-cursor" {
		t.Errorf("expected EndCursor 'company-cursor', got %s", result.PageInfo.EndCursor)
	}
}

func TestClient_GetCompany(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	employees100 := 100
	revenue1M := 1000000
	expectedCompany := types.Company{
		ID:   "company-123",
		Name: "Test Corp",
		DomainName: types.Link{
			PrimaryLinkLabel: "Website",
			PrimaryLinkUrl:   "https://testcorp.com",
		},
		Address: types.Address{
			AddressCity:     "San Francisco",
			AddressStreet1:  "123 Main St",
			AddressState:    "CA",
			AddressCountry:  "USA",
			AddressPostcode: "94102",
		},
		Employees:     &employees100,
		AnnualRevenue: &revenue1M,
		LinkedinLink: types.Link{
			PrimaryLinkLabel: "LinkedIn",
			PrimaryLinkUrl:   "https://linkedin.com/company/testcorp",
		},
		IdealCustomer: true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/rest/companies/company-123" {
			t.Errorf("expected path /rest/companies/company-123, got %s", r.URL.Path)
		}

		resp := types.CompanyResponse{}
		resp.Data.Company = expectedCompany
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	company, err := client.GetCompany(context.Background(), "company-123", nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if company.ID != "company-123" {
		t.Errorf("expected ID 'company-123', got %s", company.ID)
	}
	if company.Name != "Test Corp" {
		t.Errorf("expected name 'Test Corp', got %s", company.Name)
	}
	if company.DomainName.PrimaryLinkUrl != "https://testcorp.com" {
		t.Errorf("expected domain URL 'https://testcorp.com', got %s", company.DomainName.PrimaryLinkUrl)
	}
	if company.Address.AddressCity != "San Francisco" {
		t.Errorf("expected city 'San Francisco', got %s", company.Address.AddressCity)
	}
	if company.Employees == nil || *company.Employees != 100 {
		t.Errorf("expected 100 employees, got %v", company.Employees)
	}
	if company.AnnualRevenue == nil || *company.AnnualRevenue != 1000000 {
		t.Errorf("expected revenue 1000000, got %v", company.AnnualRevenue)
	}
	if !company.IdealCustomer {
		t.Error("expected IdealCustomer to be true")
	}
}

func TestClient_CreateCompany(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	employees25 := 25
	input := &CreateCompanyInput{
		Name: "New Startup",
		DomainName: &LinkInput{
			PrimaryLinkLabel: "Website",
			PrimaryLinkUrl:   "https://newstartup.io",
		},
		Address:       "456 Startup Ave, San Francisco, CA",
		Employees:     employees25,
		IdealCustomer: true,
		LinkedinLink: &LinkInput{
			PrimaryLinkLabel: "LinkedIn",
			PrimaryLinkUrl:   "https://linkedin.com/company/newstartup",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/rest/companies" {
			t.Errorf("expected path /rest/companies, got %s", r.URL.Path)
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var receivedInput CreateCompanyInput
		if err := json.Unmarshal(body, &receivedInput); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}
		if receivedInput.Name != "New Startup" {
			t.Errorf("expected name 'New Startup', got %s", receivedInput.Name)
		}
		if receivedInput.DomainName == nil || receivedInput.DomainName.PrimaryLinkUrl != "https://newstartup.io" {
			t.Errorf("expected domain URL 'https://newstartup.io', got %v", receivedInput.DomainName)
		}
		if receivedInput.Employees != 25 {
			t.Errorf("expected 25 employees, got %d", receivedInput.Employees)
		}
		if !receivedInput.IdealCustomer {
			t.Error("expected IdealCustomer to be true")
		}

		// Return created company
		resp := types.CreateCompanyResponse{}
		resp.Data.CreateCompany = types.Company{
			ID:   "new-company-id",
			Name: input.Name,
			DomainName: types.Link{
				PrimaryLinkLabel: input.DomainName.PrimaryLinkLabel,
				PrimaryLinkUrl:   input.DomainName.PrimaryLinkUrl,
			},
			Employees:     &employees25,
			IdealCustomer: input.IdealCustomer,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	company, err := client.CreateCompany(context.Background(), input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if company.ID != "new-company-id" {
		t.Errorf("expected ID 'new-company-id', got %s", company.ID)
	}
	if company.Name != "New Startup" {
		t.Errorf("expected name 'New Startup', got %s", company.Name)
	}
	if company.DomainName.PrimaryLinkUrl != "https://newstartup.io" {
		t.Errorf("expected domain URL 'https://newstartup.io', got %s", company.DomainName.PrimaryLinkUrl)
	}
	if company.Employees == nil || *company.Employees != 25 {
		t.Errorf("expected 25 employees, got %v", company.Employees)
	}
	if !company.IdealCustomer {
		t.Error("expected IdealCustomer to be true")
	}
}

func TestClient_UpdateCompany(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	updatedName := "Updated Startup"
	updatedEmployees := 50
	updatedIdealCustomer := false
	input := &UpdateCompanyInput{
		Name:      &updatedName,
		Employees: &updatedEmployees,
		DomainName: &LinkInput{
			PrimaryLinkLabel: "New Website",
			PrimaryLinkUrl:   "https://updatedstartup.io",
		},
		IdealCustomer: &updatedIdealCustomer,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/rest/companies/company-123" {
			t.Errorf("expected path /rest/companies/company-123, got %s", r.URL.Path)
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var receivedInput UpdateCompanyInput
		if err := json.Unmarshal(body, &receivedInput); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}
		if receivedInput.Name == nil || *receivedInput.Name != "Updated Startup" {
			t.Errorf("expected name 'Updated Startup', got %v", receivedInput.Name)
		}
		if receivedInput.Employees == nil || *receivedInput.Employees != 50 {
			t.Errorf("expected 50 employees, got %v", receivedInput.Employees)
		}

		// Return updated company
		resp := types.UpdateCompanyResponse{}
		resp.Data.UpdateCompany = types.Company{
			ID:   "company-123",
			Name: *input.Name,
			DomainName: types.Link{
				PrimaryLinkLabel: input.DomainName.PrimaryLinkLabel,
				PrimaryLinkUrl:   input.DomainName.PrimaryLinkUrl,
			},
			Employees:     &updatedEmployees,
			IdealCustomer: *input.IdealCustomer,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	company, err := client.UpdateCompany(context.Background(), "company-123", input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if company.ID != "company-123" {
		t.Errorf("expected ID 'company-123', got %s", company.ID)
	}
	if company.Name != "Updated Startup" {
		t.Errorf("expected name 'Updated Startup', got %s", company.Name)
	}
	if company.DomainName.PrimaryLinkUrl != "https://updatedstartup.io" {
		t.Errorf("expected domain URL 'https://updatedstartup.io', got %s", company.DomainName.PrimaryLinkUrl)
	}
	if company.Employees == nil || *company.Employees != 50 {
		t.Errorf("expected 50 employees, got %v", company.Employees)
	}
	if company.IdealCustomer {
		t.Error("expected IdealCustomer to be false")
	}
}

func TestClient_UpdateCompany_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":{"message":"Company not found"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	name := "Updated Name"
	_, err := client.UpdateCompany(context.Background(), "non-existent", &UpdateCompanyInput{
		Name: &name,
	})

	if err == nil {
		t.Fatal("expected error for non-existent company, got nil")
	}
}

func TestClient_DeleteCompany(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		var receivedMethod, receivedPath string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedMethod = r.Method
			receivedPath = r.URL.Path
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token", false, WithNoRetry())
		err := client.DeleteCompany(context.Background(), "company-to-delete")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if receivedMethod != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", receivedMethod)
		}
		if receivedPath != "/rest/companies/company-to-delete" {
			t.Errorf("expected path /rest/companies/company-to-delete, got %s", receivedPath)
		}
	})

	t.Run("delete non-existent returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":{"message":"Company not found"}}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token", false, WithNoRetry())
		err := client.DeleteCompany(context.Background(), "non-existent")

		if err == nil {
			t.Fatal("expected error for non-existent company, got nil")
		}
	})
}

func TestClient_ListCompanies_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":{"message":"Unauthorized"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	_, err := client.ListCompanies(context.Background(), nil)

	if err == nil {
		t.Fatal("expected error for unauthorized request, got nil")
	}
}

func TestClient_GetCompany_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":{"message":"Company not found"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	_, err := client.GetCompany(context.Background(), "non-existent", nil)

	if err == nil {
		t.Fatal("expected error for non-existent company, got nil")
	}
}

func TestClient_CreateCompany_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"code":"validation_error","message":"Name is required"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	_, err := client.CreateCompany(context.Background(), &CreateCompanyInput{})

	if err == nil {
		t.Fatal("expected error for invalid input, got nil")
	}
}
