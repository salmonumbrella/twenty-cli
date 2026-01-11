package companies

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func TestListCmd_Flags(t *testing.T) {
	flags := []string{"limit", "cursor", "all", "filter", "sort", "order"}
	for _, flag := range flags {
		if listCmd.Flags().Lookup(flag) == nil {
			t.Errorf("%s flag not registered", flag)
		}
	}
}

func TestListCmd_Use(t *testing.T) {
	if listCmd.Use != "list" {
		t.Errorf("Use = %q, want %q", listCmd.Use, "list")
	}
}

func TestListCmd_Short(t *testing.T) {
	if listCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestListCompanies_Success(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	employees50 := 50
	employees100 := 100

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/rest/companies" {
			t.Errorf("expected path /rest/companies, got %s", r.URL.Path)
		}

		resp := types.CompaniesListResponse{
			TotalCount: 2,
			PageInfo:   &types.PageInfo{HasNextPage: false},
		}
		resp.Data.Companies = []types.Company{
			{
				ID:        "company-1",
				Name:      "First Corp",
				Employees: &employees50,
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				ID:        "company-2",
				Name:      "Second Corp",
				Employees: &employees100,
				CreatedAt: now,
				UpdatedAt: now,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := rest.NewClient(server.URL, "test-token", false, rest.WithNoRetry())
	result, err := listCompanies(context.Background(), client, nil)

	if err != nil {
		t.Fatalf("listCompanies() error = %v", err)
	}

	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", result.TotalCount)
	}

	if len(result.Data) != 2 {
		t.Fatalf("expected 2 companies, got %d", len(result.Data))
	}

	if result.Data[0].ID != "company-1" {
		t.Errorf("first company ID = %q, want %q", result.Data[0].ID, "company-1")
	}

	if result.Data[1].Name != "Second Corp" {
		t.Errorf("second company Name = %q, want %q", result.Data[1].Name, "Second Corp")
	}
}

func TestListCompanies_WithOptions(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify query parameters
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("expected limit=10, got %q", r.URL.Query().Get("limit"))
		}

		resp := types.CompaniesListResponse{
			TotalCount: 1,
		}
		resp.Data.Companies = []types.Company{
			{
				ID:        "limited-company",
				Name:      "Limited Corp",
				CreatedAt: now,
				UpdatedAt: now,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := rest.NewClient(server.URL, "test-token", false, rest.WithNoRetry())
	opts := &rest.ListOptions{
		Limit: 10,
	}
	result, err := listCompanies(context.Background(), client, opts)

	if err != nil {
		t.Fatalf("listCompanies() error = %v", err)
	}

	if len(result.Data) != 1 {
		t.Errorf("expected 1 company, got %d", len(result.Data))
	}
}

func TestListCompanies_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":{"message":"Unauthorized"}}`))
	}))
	defer server.Close()

	client := rest.NewClient(server.URL, "test-token", false, rest.WithNoRetry())
	_, err := listCompanies(context.Background(), client, nil)

	if err == nil {
		t.Fatal("expected error for unauthorized request")
	}
}

func TestListCompanies_EmptyResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.CompaniesListResponse{
			TotalCount: 0,
		}
		resp.Data.Companies = []types.Company{}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := rest.NewClient(server.URL, "test-token", false, rest.WithNoRetry())
	result, err := listCompanies(context.Background(), client, nil)

	if err != nil {
		t.Fatalf("listCompanies() error = %v", err)
	}

	if result.TotalCount != 0 {
		t.Errorf("TotalCount = %d, want 0", result.TotalCount)
	}

	if len(result.Data) != 0 {
		t.Errorf("expected 0 companies, got %d", len(result.Data))
	}
}
