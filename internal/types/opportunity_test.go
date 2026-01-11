package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestOpportunity_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"id": "opp-123",
		"name": "Enterprise Deal",
		"amount": {
			"amountMicros": "5000000000",
			"currencyCode": "USD"
		},
		"closeDate": "2024-12-31",
		"stage": "NEGOTIATION",
		"probability": 75,
		"companyId": "company-456",
		"pointOfContactId": "person-789",
		"createdAt": "2024-01-15T10:30:00Z",
		"updatedAt": "2024-06-20T14:45:00Z"
	}`

	var opp Opportunity
	err := json.Unmarshal([]byte(jsonData), &opp)
	if err != nil {
		t.Fatalf("failed to unmarshal Opportunity: %v", err)
	}

	if opp.ID != "opp-123" {
		t.Errorf("expected ID='opp-123', got %q", opp.ID)
	}
	if opp.Name != "Enterprise Deal" {
		t.Errorf("expected Name='Enterprise Deal', got %q", opp.Name)
	}
	if opp.Amount == nil {
		t.Fatal("expected Amount to be set")
	}
	if opp.Amount.AmountMicros != "5000000000" {
		t.Errorf("expected AmountMicros='5000000000', got %q", opp.Amount.AmountMicros)
	}
	if opp.Amount.CurrencyCode != "USD" {
		t.Errorf("expected CurrencyCode='USD', got %q", opp.Amount.CurrencyCode)
	}
	if opp.CloseDate != "2024-12-31" {
		t.Errorf("expected CloseDate='2024-12-31', got %q", opp.CloseDate)
	}
	if opp.Stage != "NEGOTIATION" {
		t.Errorf("expected Stage='NEGOTIATION', got %q", opp.Stage)
	}
	if opp.Probability != 75 {
		t.Errorf("expected Probability=75, got %d", opp.Probability)
	}
	if opp.CompanyID != "company-456" {
		t.Errorf("expected CompanyID='company-456', got %q", opp.CompanyID)
	}
	if opp.PointOfContactID != "person-789" {
		t.Errorf("expected PointOfContactID='person-789', got %q", opp.PointOfContactID)
	}
	if opp.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if opp.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestOpportunity_JSONMarshal(t *testing.T) {
	opp := Opportunity{
		ID:   "opp-789",
		Name: "Small Business Deal",
		Amount: &Currency{
			AmountMicros: "100000000",
			CurrencyCode: "EUR",
		},
		CloseDate:        "2024-09-15",
		Stage:            "PROPOSAL",
		Probability:      50,
		CompanyID:        "company-123",
		PointOfContactID: "person-456",
		CreatedAt:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2024, 6, 20, 14, 45, 0, 0, time.UTC),
	}

	data, err := json.Marshal(opp)
	if err != nil {
		t.Fatalf("failed to marshal Opportunity: %v", err)
	}

	// Round-trip: unmarshal back
	var parsed Opportunity
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal serialized Opportunity: %v", err)
	}

	if parsed.ID != opp.ID {
		t.Errorf("round-trip failed: expected ID=%q, got %q", opp.ID, parsed.ID)
	}
	if parsed.Name != opp.Name {
		t.Errorf("round-trip failed: expected Name=%q, got %q", opp.Name, parsed.Name)
	}
	if parsed.Amount == nil {
		t.Fatal("round-trip failed: expected Amount to be set")
	}
	if parsed.Amount.AmountMicros != opp.Amount.AmountMicros {
		t.Errorf("round-trip failed: expected AmountMicros=%q, got %q", opp.Amount.AmountMicros, parsed.Amount.AmountMicros)
	}
	if parsed.Stage != opp.Stage {
		t.Errorf("round-trip failed: expected Stage=%q, got %q", opp.Stage, parsed.Stage)
	}
	if parsed.Probability != opp.Probability {
		t.Errorf("round-trip failed: expected Probability=%d, got %d", opp.Probability, parsed.Probability)
	}
}

func TestOpportunity_MinimalFields(t *testing.T) {
	jsonData := `{
		"id": "opp-minimal",
		"name": "Simple Opportunity",
		"createdAt": "2024-01-15T10:30:00Z",
		"updatedAt": "2024-06-20T14:45:00Z"
	}`

	var opp Opportunity
	err := json.Unmarshal([]byte(jsonData), &opp)
	if err != nil {
		t.Fatalf("failed to unmarshal Opportunity: %v", err)
	}

	if opp.ID != "opp-minimal" {
		t.Errorf("expected ID='opp-minimal', got %q", opp.ID)
	}
	if opp.Name != "Simple Opportunity" {
		t.Errorf("expected Name='Simple Opportunity', got %q", opp.Name)
	}
	if opp.Amount != nil {
		t.Errorf("expected Amount to be nil, got %v", opp.Amount)
	}
	if opp.CloseDate != "" {
		t.Errorf("expected CloseDate to be empty, got %q", opp.CloseDate)
	}
	if opp.Stage != "" {
		t.Errorf("expected Stage to be empty, got %q", opp.Stage)
	}
	if opp.Probability != 0 {
		t.Errorf("expected Probability=0, got %d", opp.Probability)
	}
}

func TestCurrency_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name         string
		json         string
		amountMicros string
		currencyCode string
	}{
		{
			name:         "USD amount",
			json:         `{"amountMicros": "1000000", "currencyCode": "USD"}`,
			amountMicros: "1000000",
			currencyCode: "USD",
		},
		{
			name:         "EUR amount",
			json:         `{"amountMicros": "5000000000", "currencyCode": "EUR"}`,
			amountMicros: "5000000000",
			currencyCode: "EUR",
		},
		{
			name:         "zero amount",
			json:         `{"amountMicros": "0", "currencyCode": "GBP"}`,
			amountMicros: "0",
			currencyCode: "GBP",
		},
		{
			name:         "empty fields",
			json:         `{"amountMicros": "", "currencyCode": ""}`,
			amountMicros: "",
			currencyCode: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var currency Currency
			if err := json.Unmarshal([]byte(tt.json), &currency); err != nil {
				t.Fatalf("failed to unmarshal Currency: %v", err)
			}
			if currency.AmountMicros != tt.amountMicros {
				t.Errorf("expected AmountMicros=%q, got %q", tt.amountMicros, currency.AmountMicros)
			}
			if currency.CurrencyCode != tt.currencyCode {
				t.Errorf("expected CurrencyCode=%q, got %q", tt.currencyCode, currency.CurrencyCode)
			}
		})
	}
}

func TestOpportunitiesListResponse_JSONParsing(t *testing.T) {
	jsonData := `{
		"data": {
			"opportunities": [
				{
					"id": "o1",
					"name": "Opportunity One",
					"stage": "QUALIFIED",
					"probability": 25,
					"createdAt": "2024-01-15T10:30:00Z",
					"updatedAt": "2024-06-20T14:45:00Z"
				},
				{
					"id": "o2",
					"name": "Opportunity Two",
					"stage": "CLOSED_WON",
					"probability": 100,
					"amount": {
						"amountMicros": "10000000000",
						"currencyCode": "USD"
					},
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

	var resp OpportunitiesListResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal OpportunitiesListResponse: %v", err)
	}

	if len(resp.Data.Opportunities) != 2 {
		t.Fatalf("expected 2 opportunities, got %d", len(resp.Data.Opportunities))
	}
	if resp.TotalCount != 2 {
		t.Errorf("expected TotalCount=2, got %d", resp.TotalCount)
	}
	if resp.Data.Opportunities[0].ID != "o1" {
		t.Errorf("expected first opportunity ID='o1', got %q", resp.Data.Opportunities[0].ID)
	}
	if resp.Data.Opportunities[0].Stage != "QUALIFIED" {
		t.Errorf("expected first opportunity Stage='QUALIFIED', got %q", resp.Data.Opportunities[0].Stage)
	}
	if resp.Data.Opportunities[1].Amount == nil {
		t.Error("expected second opportunity to have Amount set")
	}
	if resp.Data.Opportunities[1].Probability != 100 {
		t.Errorf("expected second opportunity Probability=100, got %d", resp.Data.Opportunities[1].Probability)
	}
	if resp.PageInfo == nil {
		t.Fatal("expected PageInfo to be set")
	}
	if !resp.PageInfo.HasNextPage {
		t.Error("expected HasNextPage=true")
	}
}

func TestOpportunityResponse_JSONParsing(t *testing.T) {
	jsonData := `{
		"data": {
			"opportunity": {
				"id": "opp-single",
				"name": "Single Opportunity",
				"stage": "DISCOVERY",
				"probability": 10,
				"createdAt": "2024-01-15T10:30:00Z",
				"updatedAt": "2024-06-20T14:45:00Z"
			}
		}
	}`

	var resp OpportunityResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal OpportunityResponse: %v", err)
	}

	if resp.Data.Opportunity.ID != "opp-single" {
		t.Errorf("expected ID='opp-single', got %q", resp.Data.Opportunity.ID)
	}
	if resp.Data.Opportunity.Name != "Single Opportunity" {
		t.Errorf("expected Name='Single Opportunity', got %q", resp.Data.Opportunity.Name)
	}
	if resp.Data.Opportunity.Stage != "DISCOVERY" {
		t.Errorf("expected Stage='DISCOVERY', got %q", resp.Data.Opportunity.Stage)
	}
}

func TestCreateOpportunityResponse_JSONParsing(t *testing.T) {
	jsonData := `{
		"data": {
			"createOpportunity": {
				"id": "new-opp",
				"name": "New Opportunity",
				"stage": "QUALIFIED",
				"createdAt": "2024-01-15T10:30:00Z",
				"updatedAt": "2024-06-20T14:45:00Z"
			}
		}
	}`

	var resp CreateOpportunityResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal CreateOpportunityResponse: %v", err)
	}

	if resp.Data.CreateOpportunity.ID != "new-opp" {
		t.Errorf("expected ID='new-opp', got %q", resp.Data.CreateOpportunity.ID)
	}
	if resp.Data.CreateOpportunity.Name != "New Opportunity" {
		t.Errorf("expected Name='New Opportunity', got %q", resp.Data.CreateOpportunity.Name)
	}
}

func TestUpdateOpportunityResponse_JSONParsing(t *testing.T) {
	jsonData := `{
		"data": {
			"updateOpportunity": {
				"id": "updated-opp",
				"name": "Updated Opportunity",
				"stage": "NEGOTIATION",
				"probability": 80,
				"createdAt": "2024-01-15T10:30:00Z",
				"updatedAt": "2024-06-20T14:45:00Z"
			}
		}
	}`

	var resp UpdateOpportunityResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal UpdateOpportunityResponse: %v", err)
	}

	if resp.Data.UpdateOpportunity.ID != "updated-opp" {
		t.Errorf("expected ID='updated-opp', got %q", resp.Data.UpdateOpportunity.ID)
	}
	if resp.Data.UpdateOpportunity.Stage != "NEGOTIATION" {
		t.Errorf("expected Stage='NEGOTIATION', got %q", resp.Data.UpdateOpportunity.Stage)
	}
	if resp.Data.UpdateOpportunity.Probability != 80 {
		t.Errorf("expected Probability=80, got %d", resp.Data.UpdateOpportunity.Probability)
	}
}

func TestDeleteOpportunityResponse_JSONParsing(t *testing.T) {
	jsonData := `{
		"data": {
			"deleteOpportunity": {
				"id": "deleted-opp",
				"name": "Deleted Opportunity",
				"createdAt": "2024-01-15T10:30:00Z",
				"updatedAt": "2024-06-20T14:45:00Z"
			}
		}
	}`

	var resp DeleteOpportunityResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal DeleteOpportunityResponse: %v", err)
	}

	if resp.Data.DeleteOpportunity.ID != "deleted-opp" {
		t.Errorf("expected ID='deleted-opp', got %q", resp.Data.DeleteOpportunity.ID)
	}
}

func TestOpportunity_WithRelations(t *testing.T) {
	jsonData := `{
		"id": "opp-with-relations",
		"name": "Deal with Relations",
		"stage": "PROPOSAL",
		"company": {
			"id": "company-rel",
			"name": "Related Corp",
			"createdAt": "2024-01-15T10:30:00Z",
			"updatedAt": "2024-06-20T14:45:00Z"
		},
		"pointOfContact": {
			"id": "person-rel",
			"name": {
				"firstName": "Jane",
				"lastName": "Smith"
			},
			"createdAt": "2024-01-15T10:30:00Z",
			"updatedAt": "2024-06-20T14:45:00Z"
		},
		"createdAt": "2024-01-15T10:30:00Z",
		"updatedAt": "2024-06-20T14:45:00Z"
	}`

	var opp Opportunity
	err := json.Unmarshal([]byte(jsonData), &opp)
	if err != nil {
		t.Fatalf("failed to unmarshal Opportunity with relations: %v", err)
	}

	if opp.Company == nil {
		t.Fatal("expected Company relation to be set")
	}
	if opp.Company.ID != "company-rel" {
		t.Errorf("expected Company.ID='company-rel', got %q", opp.Company.ID)
	}
	if opp.Company.Name != "Related Corp" {
		t.Errorf("expected Company.Name='Related Corp', got %q", opp.Company.Name)
	}
	if opp.PointOfContact == nil {
		t.Fatal("expected PointOfContact relation to be set")
	}
	if opp.PointOfContact.ID != "person-rel" {
		t.Errorf("expected PointOfContact.ID='person-rel', got %q", opp.PointOfContact.ID)
	}
	if opp.PointOfContact.Name.FirstName != "Jane" {
		t.Errorf("expected PointOfContact.Name.FirstName='Jane', got %q", opp.PointOfContact.Name.FirstName)
	}
}

func TestOpportunity_StageValues(t *testing.T) {
	stages := []string{"QUALIFIED", "DISCOVERY", "PROPOSAL", "NEGOTIATION", "CLOSED_WON", "CLOSED_LOST"}

	for _, stage := range stages {
		t.Run(stage, func(t *testing.T) {
			jsonData := `{
				"id": "opp-stage-test",
				"name": "Test Opportunity",
				"stage": "` + stage + `",
				"createdAt": "2024-01-15T10:30:00Z",
				"updatedAt": "2024-06-20T14:45:00Z"
			}`

			var opp Opportunity
			err := json.Unmarshal([]byte(jsonData), &opp)
			if err != nil {
				t.Fatalf("failed to unmarshal Opportunity: %v", err)
			}

			if opp.Stage != stage {
				t.Errorf("expected Stage=%q, got %q", stage, opp.Stage)
			}
		})
	}
}
