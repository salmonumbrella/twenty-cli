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

func TestClient_ListOpportunities(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	expectedOpportunities := []types.Opportunity{
		{
			ID:   "opp-1",
			Name: "Enterprise Deal",
			Amount: &types.Currency{
				AmountMicros: "50000000000",
				CurrencyCode: "USD",
			},
			CloseDate:   "2025-06-30",
			Stage:       "NEGOTIATION",
			Probability: 75,
			CompanyID:   "company-1",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:   "opp-2",
			Name: "Startup Contract",
			Amount: &types.Currency{
				AmountMicros: "10000000000",
				CurrencyCode: "USD",
			},
			CloseDate:        "2025-07-15",
			Stage:            "PROPOSAL",
			Probability:      50,
			CompanyID:        "company-2",
			PointOfContactID: "person-1",
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/rest/opportunities" {
			t.Errorf("expected path /rest/opportunities, got %s", r.URL.Path)
		}

		resp := types.OpportunitiesListResponse{
			TotalCount: 2,
			PageInfo:   &types.PageInfo{HasNextPage: true, EndCursor: "opp-cursor"},
		}
		resp.Data.Opportunities = expectedOpportunities

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	result, err := client.ListOpportunities(context.Background(), nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("expected TotalCount 2, got %d", result.TotalCount)
	}
	if len(result.Data) != 2 {
		t.Fatalf("expected 2 opportunities, got %d", len(result.Data))
	}
	if result.Data[0].ID != "opp-1" {
		t.Errorf("expected first opportunity ID 'opp-1', got %s", result.Data[0].ID)
	}
	if result.Data[0].Name != "Enterprise Deal" {
		t.Errorf("expected first opportunity name 'Enterprise Deal', got %s", result.Data[0].Name)
	}
	if result.Data[0].Amount == nil {
		t.Fatal("expected Amount to be set")
	}
	if result.Data[0].Amount.AmountMicros != "50000000000" {
		t.Errorf("expected AmountMicros '50000000000', got %s", result.Data[0].Amount.AmountMicros)
	}
	if result.Data[0].Amount.CurrencyCode != "USD" {
		t.Errorf("expected CurrencyCode 'USD', got %s", result.Data[0].Amount.CurrencyCode)
	}
	if result.Data[0].Stage != "NEGOTIATION" {
		t.Errorf("expected Stage 'NEGOTIATION', got %s", result.Data[0].Stage)
	}
	if result.Data[0].Probability != 75 {
		t.Errorf("expected Probability 75, got %d", result.Data[0].Probability)
	}
	if result.Data[0].CompanyID != "company-1" {
		t.Errorf("expected CompanyID 'company-1', got %s", result.Data[0].CompanyID)
	}
	if result.PageInfo == nil {
		t.Fatal("expected PageInfo to be set")
	}
	if !result.PageInfo.HasNextPage {
		t.Error("expected HasNextPage to be true")
	}
	if result.PageInfo.EndCursor != "opp-cursor" {
		t.Errorf("expected EndCursor 'opp-cursor', got %s", result.PageInfo.EndCursor)
	}
}

func TestClient_GetOpportunity(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	expectedOpportunity := types.Opportunity{
		ID:   "opp-123",
		Name: "Big Enterprise Deal",
		Amount: &types.Currency{
			AmountMicros: "100000000000",
			CurrencyCode: "EUR",
		},
		CloseDate:        "2025-09-30",
		Stage:            "CLOSED_WON",
		Probability:      100,
		CompanyID:        "company-abc",
		PointOfContactID: "person-xyz",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/rest/opportunities/opp-123" {
			t.Errorf("expected path /rest/opportunities/opp-123, got %s", r.URL.Path)
		}

		resp := types.OpportunityResponse{}
		resp.Data.Opportunity = expectedOpportunity
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	opportunity, err := client.GetOpportunity(context.Background(), "opp-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opportunity.ID != "opp-123" {
		t.Errorf("expected ID 'opp-123', got %s", opportunity.ID)
	}
	if opportunity.Name != "Big Enterprise Deal" {
		t.Errorf("expected name 'Big Enterprise Deal', got %s", opportunity.Name)
	}
	if opportunity.Amount == nil {
		t.Fatal("expected Amount to be set")
	}
	if opportunity.Amount.AmountMicros != "100000000000" {
		t.Errorf("expected AmountMicros '100000000000', got %s", opportunity.Amount.AmountMicros)
	}
	if opportunity.Amount.CurrencyCode != "EUR" {
		t.Errorf("expected CurrencyCode 'EUR', got %s", opportunity.Amount.CurrencyCode)
	}
	if opportunity.CloseDate != "2025-09-30" {
		t.Errorf("expected CloseDate '2025-09-30', got %s", opportunity.CloseDate)
	}
	if opportunity.Stage != "CLOSED_WON" {
		t.Errorf("expected Stage 'CLOSED_WON', got %s", opportunity.Stage)
	}
	if opportunity.Probability != 100 {
		t.Errorf("expected Probability 100, got %d", opportunity.Probability)
	}
	if opportunity.CompanyID != "company-abc" {
		t.Errorf("expected CompanyID 'company-abc', got %s", opportunity.CompanyID)
	}
	if opportunity.PointOfContactID != "person-xyz" {
		t.Errorf("expected PointOfContactID 'person-xyz', got %s", opportunity.PointOfContactID)
	}
}

func TestClient_GetOpportunity_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":{"message":"Opportunity not found"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	_, err := client.GetOpportunity(context.Background(), "non-existent")

	if err == nil {
		t.Fatal("expected error for non-existent opportunity, got nil")
	}
}

func TestClient_CreateOpportunity(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	input := &CreateOpportunityInput{
		Name: "New Deal",
		Amount: &types.Currency{
			AmountMicros: "25000000000",
			CurrencyCode: "USD",
		},
		CloseDate:        "2025-06-30",
		Stage:            "PROPOSAL",
		Probability:      50,
		CompanyID:        "company-123",
		PointOfContactID: "person-456",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/rest/opportunities" {
			t.Errorf("expected path /rest/opportunities, got %s", r.URL.Path)
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var receivedInput CreateOpportunityInput
		if err := json.Unmarshal(body, &receivedInput); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}
		if receivedInput.Name != "New Deal" {
			t.Errorf("expected name 'New Deal', got %s", receivedInput.Name)
		}
		if receivedInput.Amount == nil || receivedInput.Amount.AmountMicros != "25000000000" {
			t.Errorf("expected amount content, got %v", receivedInput.Amount)
		}
		if receivedInput.Stage != "PROPOSAL" {
			t.Errorf("expected stage 'PROPOSAL', got %s", receivedInput.Stage)
		}
		if receivedInput.Probability != 50 {
			t.Errorf("expected probability 50, got %d", receivedInput.Probability)
		}

		// Return created opportunity
		resp := types.CreateOpportunityResponse{}
		resp.Data.CreateOpportunity = types.Opportunity{
			ID:               "new-opp-id",
			Name:             input.Name,
			Amount:           input.Amount,
			CloseDate:        input.CloseDate,
			Stage:            input.Stage,
			Probability:      input.Probability,
			CompanyID:        input.CompanyID,
			PointOfContactID: input.PointOfContactID,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	opp, err := client.CreateOpportunity(context.Background(), input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opp.ID != "new-opp-id" {
		t.Errorf("expected ID 'new-opp-id', got %s", opp.ID)
	}
	if opp.Name != "New Deal" {
		t.Errorf("expected name 'New Deal', got %s", opp.Name)
	}
	if opp.Amount == nil || opp.Amount.AmountMicros != "25000000000" {
		t.Errorf("expected amount content, got %v", opp.Amount)
	}
	if opp.Stage != "PROPOSAL" {
		t.Errorf("expected stage 'PROPOSAL', got %s", opp.Stage)
	}
	if opp.Probability != 50 {
		t.Errorf("expected probability 50, got %d", opp.Probability)
	}
}

func TestClient_CreateOpportunity_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"code":"validation_error","message":"Name is required"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	_, err := client.CreateOpportunity(context.Background(), &CreateOpportunityInput{})

	if err == nil {
		t.Fatal("expected error for invalid input, got nil")
	}
}

func TestClient_UpdateOpportunity(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	updatedName := "Updated Deal"
	updatedStage := "NEGOTIATION"
	updatedProbability := 75
	input := &UpdateOpportunityInput{
		Name:        &updatedName,
		Stage:       &updatedStage,
		Probability: &updatedProbability,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/rest/opportunities/opp-123" {
			t.Errorf("expected path /rest/opportunities/opp-123, got %s", r.URL.Path)
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var receivedInput UpdateOpportunityInput
		if err := json.Unmarshal(body, &receivedInput); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}
		if receivedInput.Name == nil || *receivedInput.Name != "Updated Deal" {
			t.Errorf("expected name 'Updated Deal', got %v", receivedInput.Name)
		}
		if receivedInput.Stage == nil || *receivedInput.Stage != "NEGOTIATION" {
			t.Errorf("expected stage 'NEGOTIATION', got %v", receivedInput.Stage)
		}
		if receivedInput.Probability == nil || *receivedInput.Probability != 75 {
			t.Errorf("expected probability 75, got %v", receivedInput.Probability)
		}

		// Return updated opportunity
		resp := types.UpdateOpportunityResponse{}
		resp.Data.UpdateOpportunity = types.Opportunity{
			ID:          "opp-123",
			Name:        *input.Name,
			Stage:       *input.Stage,
			Probability: *input.Probability,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	opp, err := client.UpdateOpportunity(context.Background(), "opp-123", input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opp.ID != "opp-123" {
		t.Errorf("expected ID 'opp-123', got %s", opp.ID)
	}
	if opp.Name != "Updated Deal" {
		t.Errorf("expected name 'Updated Deal', got %s", opp.Name)
	}
	if opp.Stage != "NEGOTIATION" {
		t.Errorf("expected stage 'NEGOTIATION', got %s", opp.Stage)
	}
	if opp.Probability != 75 {
		t.Errorf("expected probability 75, got %d", opp.Probability)
	}
}

func TestClient_UpdateOpportunity_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":{"message":"Opportunity not found"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	name := "Updated"
	_, err := client.UpdateOpportunity(context.Background(), "non-existent", &UpdateOpportunityInput{
		Name: &name,
	})

	if err == nil {
		t.Fatal("expected error for non-existent opportunity, got nil")
	}
}

func TestClient_DeleteOpportunity(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		var receivedMethod, receivedPath string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedMethod = r.Method
			receivedPath = r.URL.Path
			resp := types.DeleteOpportunityResponse{}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token", false, WithNoRetry())
		err := client.DeleteOpportunity(context.Background(), "opp-to-delete")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if receivedMethod != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", receivedMethod)
		}
		if receivedPath != "/rest/opportunities/opp-to-delete" {
			t.Errorf("expected path /rest/opportunities/opp-to-delete, got %s", receivedPath)
		}
	})

	t.Run("delete non-existent returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":{"message":"Opportunity not found"}}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token", false, WithNoRetry())
		err := client.DeleteOpportunity(context.Background(), "non-existent")

		if err == nil {
			t.Fatal("expected error for non-existent opportunity, got nil")
		}
	})
}

func TestClient_ListOpportunities_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":{"message":"Unauthorized"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	_, err := client.ListOpportunities(context.Background(), nil)

	if err == nil {
		t.Fatal("expected error for unauthorized request, got nil")
	}
}
