package rest

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/salmonumbrella/twenty-cli/internal/types"
)

func TestClient_ListPeople(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	expectedPeople := []types.Person{
		{
			ID:        "person-1",
			Name:      types.Name{FirstName: "John", LastName: "Doe"},
			Email:     types.Email{PrimaryEmail: "john@example.com"},
			JobTitle:  "Engineer",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "person-2",
			Name:      types.Name{FirstName: "Jane", LastName: "Smith"},
			Email:     types.Email{PrimaryEmail: "jane@example.com"},
			JobTitle:  "Manager",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/rest/people" {
			t.Errorf("expected path /rest/people, got %s", r.URL.Path)
		}

		resp := types.PeopleListResponse{
			TotalCount: 2,
			PageInfo:   &types.PageInfo{HasNextPage: false, EndCursor: "cursor-2"},
		}
		resp.Data.People = expectedPeople

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	result, err := client.ListPeople(context.Background(), nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("expected TotalCount 2, got %d", result.TotalCount)
	}
	if len(result.Data) != 2 {
		t.Fatalf("expected 2 people, got %d", len(result.Data))
	}
	if result.Data[0].ID != "person-1" {
		t.Errorf("expected first person ID 'person-1', got %s", result.Data[0].ID)
	}
	if result.Data[0].Name.FirstName != "John" {
		t.Errorf("expected first name 'John', got %s", result.Data[0].Name.FirstName)
	}
	if result.Data[1].Email.PrimaryEmail != "jane@example.com" {
		t.Errorf("expected second email 'jane@example.com', got %s", result.Data[1].Email.PrimaryEmail)
	}
	if result.PageInfo == nil {
		t.Fatal("expected PageInfo to be set")
	}
	if result.PageInfo.EndCursor != "cursor-2" {
		t.Errorf("expected EndCursor 'cursor-2', got %s", result.PageInfo.EndCursor)
	}
}

func TestClient_ListPeople_WithOptions(t *testing.T) {
	tests := []struct {
		name           string
		opts           *ListOptions
		expectedParams map[string]string
	}{
		{
			name: "with limit",
			opts: &ListOptions{Limit: 10},
			expectedParams: map[string]string{
				"limit": "10",
			},
		},
		{
			name: "with cursor",
			opts: &ListOptions{Cursor: "abc123"},
			expectedParams: map[string]string{
				"starting_after": "abc123",
			},
		},
		{
			name: "with sort and order",
			opts: &ListOptions{Sort: "createdAt", Order: "desc"},
			expectedParams: map[string]string{
				"order_by":           "createdAt",
				"order_by_direction": "desc",
			},
		},
		{
			name: "with include (sets depth)",
			opts: &ListOptions{Include: []string{"company"}},
			expectedParams: map[string]string{
				"depth": "1",
			},
		},
		{
			name: "combined options",
			opts: &ListOptions{
				Limit:   25,
				Cursor:  "xyz789",
				Sort:    "name",
				Order:   "asc",
				Include: []string{"company"},
			},
			expectedParams: map[string]string{
				"limit":              "25",
				"starting_after":     "xyz789",
				"order_by":           "name",
				"order_by_direction": "asc",
				"depth":              "1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedQuery string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedQuery = r.URL.RawQuery
				resp := types.PeopleListResponse{TotalCount: 0}
				resp.Data.People = []types.Person{}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-token", false, WithNoRetry())
			_, err := client.ListPeople(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for param, expected := range tt.expectedParams {
				if !strings.Contains(receivedQuery, param+"="+expected) {
					t.Errorf("expected query to contain %s=%s, got query: %s", param, expected, receivedQuery)
				}
			}
		})
	}
}

func TestClient_GetPerson(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	expectedPerson := types.Person{
		ID:        "person-123",
		Name:      types.Name{FirstName: "Alice", LastName: "Wonder"},
		Email:     types.Email{PrimaryEmail: "alice@example.com"},
		Phone:     types.Phone{PrimaryPhoneNumber: "+1234567890"},
		JobTitle:  "CTO",
		City:      "San Francisco",
		CompanyID: "company-456",
		CreatedAt: now,
		UpdatedAt: now,
	}

	t.Run("basic get", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/rest/people/person-123" {
				t.Errorf("expected path /rest/people/person-123, got %s", r.URL.Path)
			}

			resp := types.PersonResponse{}
			resp.Data.Person = expectedPerson
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token", false, WithNoRetry())
		person, err := client.GetPerson(context.Background(), "person-123", nil)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if person.ID != "person-123" {
			t.Errorf("expected ID 'person-123', got %s", person.ID)
		}
		if person.Name.FirstName != "Alice" {
			t.Errorf("expected first name 'Alice', got %s", person.Name.FirstName)
		}
		if person.Email.PrimaryEmail != "alice@example.com" {
			t.Errorf("expected email 'alice@example.com', got %s", person.Email.PrimaryEmail)
		}
		if person.Phone.PrimaryPhoneNumber != "+1234567890" {
			t.Errorf("expected phone '+1234567890', got %s", person.Phone.PrimaryPhoneNumber)
		}
	})

	t.Run("with include option adds depth param", func(t *testing.T) {
		var receivedPath string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedPath = r.URL.RequestURI()
			resp := types.PersonResponse{}
			resp.Data.Person = expectedPerson
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token", false, WithNoRetry())
		_, err := client.GetPerson(context.Background(), "person-123", &GetPersonOptions{
			Include: []string{"company"},
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(receivedPath, "depth=1") {
			t.Errorf("expected path to contain depth=1, got %s", receivedPath)
		}
	})
}

func TestClient_CreatePerson(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	input := &CreatePersonInput{
		Name:     types.Name{FirstName: "Bob", LastName: "Builder"},
		Email:    types.Email{PrimaryEmail: "bob@example.com"},
		Phone:    types.Phone{PrimaryPhoneNumber: "+9876543210"},
		JobTitle: "Architect",
		City:     "New York",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/rest/people" {
			t.Errorf("expected path /rest/people, got %s", r.URL.Path)
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var receivedInput CreatePersonInput
		if err := json.Unmarshal(body, &receivedInput); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}
		if receivedInput.Name.FirstName != "Bob" {
			t.Errorf("expected firstName 'Bob', got %s", receivedInput.Name.FirstName)
		}
		if receivedInput.Email.PrimaryEmail != "bob@example.com" {
			t.Errorf("expected email 'bob@example.com', got %s", receivedInput.Email.PrimaryEmail)
		}

		// Return created person
		resp := types.CreatePersonResponse{}
		resp.Data.CreatePerson = types.Person{
			ID:        "new-person-id",
			Name:      input.Name,
			Email:     input.Email,
			Phone:     input.Phone,
			JobTitle:  input.JobTitle,
			City:      input.City,
			CreatedAt: now,
			UpdatedAt: now,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	person, err := client.CreatePerson(context.Background(), input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if person.ID != "new-person-id" {
		t.Errorf("expected ID 'new-person-id', got %s", person.ID)
	}
	if person.Name.FirstName != "Bob" {
		t.Errorf("expected first name 'Bob', got %s", person.Name.FirstName)
	}
	if person.Name.LastName != "Builder" {
		t.Errorf("expected last name 'Builder', got %s", person.Name.LastName)
	}
	if person.JobTitle != "Architect" {
		t.Errorf("expected job title 'Architect', got %s", person.JobTitle)
	}
}

func TestClient_DeletePerson(t *testing.T) {
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
		err := client.DeletePerson(context.Background(), "person-to-delete")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if receivedMethod != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", receivedMethod)
		}
		if receivedPath != "/rest/people/person-to-delete" {
			t.Errorf("expected path /rest/people/person-to-delete, got %s", receivedPath)
		}
	})

	t.Run("delete non-existent returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":{"message":"Person not found"}}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-token", false, WithNoRetry())
		err := client.DeletePerson(context.Background(), "non-existent")

		if err == nil {
			t.Fatal("expected error for non-existent person, got nil")
		}
	})
}

func TestClient_UpdatePerson(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	updatedName := types.Name{FirstName: "Updated", LastName: "Person"}
	updatedJobTitle := "Senior Engineer"
	input := &UpdatePersonInput{
		Name:     &updatedName,
		JobTitle: &updatedJobTitle,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/rest/people/person-123" {
			t.Errorf("expected path /rest/people/person-123, got %s", r.URL.Path)
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var receivedInput UpdatePersonInput
		if err := json.Unmarshal(body, &receivedInput); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}
		if receivedInput.Name == nil || receivedInput.Name.FirstName != "Updated" {
			t.Errorf("expected firstName 'Updated', got %v", receivedInput.Name)
		}
		if receivedInput.JobTitle == nil || *receivedInput.JobTitle != "Senior Engineer" {
			t.Errorf("expected jobTitle 'Senior Engineer', got %v", receivedInput.JobTitle)
		}

		// Return updated person
		resp := types.UpdatePersonResponse{}
		resp.Data.UpdatePerson = types.Person{
			ID:        "person-123",
			Name:      *input.Name,
			JobTitle:  *input.JobTitle,
			CreatedAt: now,
			UpdatedAt: now,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	person, err := client.UpdatePerson(context.Background(), "person-123", input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if person.ID != "person-123" {
		t.Errorf("expected ID 'person-123', got %s", person.ID)
	}
	if person.Name.FirstName != "Updated" {
		t.Errorf("expected firstName 'Updated', got %s", person.Name.FirstName)
	}
	if person.JobTitle != "Senior Engineer" {
		t.Errorf("expected jobTitle 'Senior Engineer', got %s", person.JobTitle)
	}
}

func TestClient_UpdatePerson_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":{"message":"Person not found"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	name := types.Name{FirstName: "Test", LastName: "Person"}
	_, err := client.UpdatePerson(context.Background(), "non-existent", &UpdatePersonInput{
		Name: &name,
	})

	if err == nil {
		t.Fatal("expected error for non-existent person, got nil")
	}
}

func TestClient_ListPeople_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":{"message":"Unauthorized"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	_, err := client.ListPeople(context.Background(), nil)

	if err == nil {
		t.Fatal("expected error for unauthorized request, got nil")
	}
}

func TestClient_GetPerson_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":{"message":"Person not found"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	_, err := client.GetPerson(context.Background(), "non-existent", nil)

	if err == nil {
		t.Fatal("expected error for non-existent person, got nil")
	}
}

func TestClient_CreatePerson_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"code":"validation_error","message":"Name is required"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false, WithNoRetry())
	_, err := client.CreatePerson(context.Background(), &CreatePersonInput{})

	if err == nil {
		t.Fatal("expected error for invalid input, got nil")
	}
}
