package graphql

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_UpsertPerson(t *testing.T) {
	t.Run("creates person with minimal required fields", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": {
					"createPerson": {
						"id": "person-123",
						"name": {
							"firstName": "John",
							"lastName": "Doe"
						},
						"emails": {
							"primaryEmail": "john@example.com"
						}
					}
				}
			}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", false)
		ctx := context.Background()

		input := &UpsertPersonInput{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
		}

		person, err := client.UpsertPerson(ctx, input)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if person == nil {
			t.Fatal("expected person, got nil")
		}
		if person.ID != "person-123" {
			t.Errorf("expected ID person-123, got %s", person.ID)
		}
		if person.Name.FirstName != "John" {
			t.Errorf("expected FirstName 'John', got %s", person.Name.FirstName)
		}
		if person.Name.LastName != "Doe" {
			t.Errorf("expected LastName 'Doe', got %s", person.Name.LastName)
		}
		if person.Email.PrimaryEmail != "john@example.com" {
			t.Errorf("expected Email 'john@example.com', got %s", person.Email.PrimaryEmail)
		}
	})

	t.Run("creates person with all fields", func(t *testing.T) {
		var receivedBody map[string]interface{}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &receivedBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": {
					"createPerson": {
						"id": "person-456",
						"name": {
							"firstName": "Jane",
							"lastName": "Smith"
						},
						"emails": {
							"primaryEmail": "jane@example.com"
						}
					}
				}
			}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", false)
		ctx := context.Background()

		input := &UpsertPersonInput{
			FirstName: "Jane",
			LastName:  "Smith",
			Email:     "jane@example.com",
			Phone:     "+1-555-123-4567",
			JobTitle:  "Software Engineer",
			CompanyID: "company-789",
		}

		person, err := client.UpsertPerson(ctx, input)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if person == nil {
			t.Fatal("expected person, got nil")
		}
		if person.ID != "person-456" {
			t.Errorf("expected ID person-456, got %s", person.ID)
		}

		// Verify the request body contained all fields
		if receivedBody["variables"] == nil {
			t.Fatal("expected variables in request body")
		}
		vars := receivedBody["variables"].(map[string]interface{})
		data := vars["data"].(map[string]interface{})

		// Check name was sent
		name := data["name"].(map[string]interface{})
		if name["firstName"] != "Jane" {
			t.Errorf("expected firstName 'Jane' in request, got %v", name["firstName"])
		}
		if name["lastName"] != "Smith" {
			t.Errorf("expected lastName 'Smith' in request, got %v", name["lastName"])
		}

		// Check emails was sent
		emails := data["emails"].(map[string]interface{})
		if emails["primaryEmail"] != "jane@example.com" {
			t.Errorf("expected email 'jane@example.com' in request, got %v", emails["primaryEmail"])
		}

		// Check phones was sent
		phones := data["phones"].(map[string]interface{})
		if phones["primaryPhoneNumber"] != "+1-555-123-4567" {
			t.Errorf("expected phone '+1-555-123-4567' in request, got %v", phones["primaryPhoneNumber"])
		}

		// Check jobTitle was sent
		if data["jobTitle"] != "Software Engineer" {
			t.Errorf("expected jobTitle 'Software Engineer' in request, got %v", data["jobTitle"])
		}

		// Check companyId was sent
		if data["companyId"] != "company-789" {
			t.Errorf("expected companyId 'company-789' in request, got %v", data["companyId"])
		}
	})

	t.Run("creates person without optional fields", func(t *testing.T) {
		var receivedBody map[string]interface{}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &receivedBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": {
					"createPerson": {
						"id": "person-minimal",
						"name": {"firstName": "Min", "lastName": "Imal"},
						"emails": {"primaryEmail": "minimal@example.com"}
					}
				}
			}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", false)
		ctx := context.Background()

		input := &UpsertPersonInput{
			FirstName: "Min",
			LastName:  "Imal",
			Email:     "minimal@example.com",
			// Phone, JobTitle, CompanyID are empty
		}

		_, err := client.UpsertPerson(ctx, input)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify optional fields are not in the request
		vars := receivedBody["variables"].(map[string]interface{})
		data := vars["data"].(map[string]interface{})

		// phones should not be present when empty
		if data["phones"] != nil {
			t.Error("expected phones to not be sent when empty")
		}

		// jobTitle should not be present when empty
		if data["jobTitle"] != nil {
			t.Error("expected jobTitle to not be sent when empty")
		}

		// companyId should not be present when empty
		if data["companyId"] != nil {
			t.Error("expected companyId to not be sent when empty")
		}
	})

	t.Run("handles GraphQL error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": null,
				"errors": [
					{
						"message": "Duplicate email address",
						"extensions": {"code": "DUPLICATE_ENTRY"}
					}
				]
			}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", false)
		ctx := context.Background()

		input := &UpsertPersonInput{
			FirstName: "Dup",
			LastName:  "User",
			Email:     "duplicate@example.com",
		}

		person, err := client.UpsertPerson(ctx, input)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if person != nil {
			t.Error("expected nil person on error")
		}
		if !strings.Contains(err.Error(), "Duplicate email") {
			t.Errorf("expected error about duplicate email, got: %v", err)
		}
	})

	t.Run("handles HTTP error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`Server Error`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", false)
		ctx := context.Background()

		input := &UpsertPersonInput{
			FirstName: "Error",
			LastName:  "User",
			Email:     "error@example.com",
		}

		person, err := client.UpsertPerson(ctx, input)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if person != nil {
			t.Error("expected nil person on error")
		}
	})

	t.Run("handles network error", func(t *testing.T) {
		client := NewClient("http://localhost:1", "token", false)
		ctx := context.Background()

		input := &UpsertPersonInput{
			FirstName: "Network",
			LastName:  "Error",
			Email:     "network@example.com",
		}

		person, err := client.UpsertPerson(ctx, input)

		if err == nil {
			t.Fatal("expected network error, got nil")
		}
		if person != nil {
			t.Error("expected nil person on error")
		}
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
				return
			}
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", false)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		input := &UpsertPersonInput{
			FirstName: "Cancel",
			LastName:  "User",
			Email:     "cancel@example.com",
		}

		person, err := client.UpsertPerson(ctx, input)

		if err == nil {
			t.Fatal("expected error due to context cancellation, got nil")
		}
		if person != nil {
			t.Error("expected nil person on cancellation")
		}
	})

	t.Run("uses upsert=true in mutation", func(t *testing.T) {
		var receivedBody map[string]interface{}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &receivedBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": {
					"createPerson": {
						"id": "upserted-123",
						"name": {"firstName": "Upsert", "lastName": "Test"},
						"emails": {"primaryEmail": "upsert@example.com"}
					}
				}
			}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", false)
		ctx := context.Background()

		input := &UpsertPersonInput{
			FirstName: "Upsert",
			LastName:  "Test",
			Email:     "upsert@example.com",
		}

		_, err := client.UpsertPerson(ctx, input)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify the query contains createPerson with upsert
		query := receivedBody["query"].(string)
		if !strings.Contains(query, "createPerson") {
			t.Error("expected query to contain 'createPerson'")
		}
		// The upsert parameter should be in the mutation
		if !strings.Contains(query, "upsert") {
			t.Error("expected query to contain 'upsert'")
		}
	})

	t.Run("handles empty first name", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": {
					"createPerson": {
						"id": "empty-first",
						"name": {"firstName": "", "lastName": "OnlyLast"},
						"emails": {"primaryEmail": "nofirst@example.com"}
					}
				}
			}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", false)
		ctx := context.Background()

		input := &UpsertPersonInput{
			FirstName: "",
			LastName:  "OnlyLast",
			Email:     "nofirst@example.com",
		}

		person, err := client.UpsertPerson(ctx, input)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if person.Name.FirstName != "" {
			t.Errorf("expected empty FirstName, got %s", person.Name.FirstName)
		}
		if person.Name.LastName != "OnlyLast" {
			t.Errorf("expected LastName 'OnlyLast', got %s", person.Name.LastName)
		}
	})

	t.Run("handles empty last name", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": {
					"createPerson": {
						"id": "empty-last",
						"name": {"firstName": "OnlyFirst", "lastName": ""},
						"emails": {"primaryEmail": "nolast@example.com"}
					}
				}
			}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", false)
		ctx := context.Background()

		input := &UpsertPersonInput{
			FirstName: "OnlyFirst",
			LastName:  "",
			Email:     "nolast@example.com",
		}

		person, err := client.UpsertPerson(ctx, input)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if person.Name.FirstName != "OnlyFirst" {
			t.Errorf("expected FirstName 'OnlyFirst', got %s", person.Name.FirstName)
		}
		if person.Name.LastName != "" {
			t.Errorf("expected empty LastName, got %s", person.Name.LastName)
		}
	})
}

func TestUpsertPersonInput(t *testing.T) {
	t.Run("struct fields are accessible", func(t *testing.T) {
		input := UpsertPersonInput{
			FirstName: "Test",
			LastName:  "User",
			Email:     "test@example.com",
			Phone:     "+1-555-000-0000",
			JobTitle:  "Developer",
			CompanyID: "comp-123",
		}

		if input.FirstName != "Test" {
			t.Errorf("expected FirstName 'Test', got %s", input.FirstName)
		}
		if input.LastName != "User" {
			t.Errorf("expected LastName 'User', got %s", input.LastName)
		}
		if input.Email != "test@example.com" {
			t.Errorf("expected Email 'test@example.com', got %s", input.Email)
		}
		if input.Phone != "+1-555-000-0000" {
			t.Errorf("expected Phone '+1-555-000-0000', got %s", input.Phone)
		}
		if input.JobTitle != "Developer" {
			t.Errorf("expected JobTitle 'Developer', got %s", input.JobTitle)
		}
		if input.CompanyID != "comp-123" {
			t.Errorf("expected CompanyID 'comp-123', got %s", input.CompanyID)
		}
	})

	t.Run("zero value initialization", func(t *testing.T) {
		var input UpsertPersonInput

		if input.FirstName != "" {
			t.Error("expected empty FirstName")
		}
		if input.LastName != "" {
			t.Error("expected empty LastName")
		}
		if input.Email != "" {
			t.Error("expected empty Email")
		}
		if input.Phone != "" {
			t.Error("expected empty Phone")
		}
		if input.JobTitle != "" {
			t.Error("expected empty JobTitle")
		}
		if input.CompanyID != "" {
			t.Error("expected empty CompanyID")
		}
	})
}
