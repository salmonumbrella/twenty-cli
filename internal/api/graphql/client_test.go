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

func TestNewClient(t *testing.T) {
	t.Run("creates client with correct configuration", func(t *testing.T) {
		client := NewClient("https://api.example.com", "test-token", false)

		if client == nil {
			t.Fatal("expected client to be created, got nil")
		}
		if client.client == nil {
			t.Error("expected internal graphql client to be set")
		}
		if client.debug != false {
			t.Error("expected debug to be false")
		}
	})

	t.Run("creates client with debug mode enabled", func(t *testing.T) {
		client := NewClient("https://api.example.com", "test-token", true)

		if client == nil {
			t.Fatal("expected client to be created, got nil")
		}
		if client.debug != true {
			t.Error("expected debug to be true")
		}
	})

	t.Run("sends bearer token in authorization header", func(t *testing.T) {
		expectedToken := "my-secret-token-12345"
		var receivedAuthHeader string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedAuthHeader = r.Header.Get("Authorization")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":{}}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, expectedToken, false)
		ctx := context.Background()

		var query struct{}
		_ = client.Query(ctx, &query, nil)

		expectedAuthHeader := "Bearer " + expectedToken
		if receivedAuthHeader != expectedAuthHeader {
			t.Errorf("expected Authorization header %q, got %q", expectedAuthHeader, receivedAuthHeader)
		}
	})

	t.Run("makes requests to /graphql endpoint", func(t *testing.T) {
		var receivedPath string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedPath = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":{}}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", false)
		ctx := context.Background()

		var query struct{}
		_ = client.Query(ctx, &query, nil)

		if receivedPath != "/graphql" {
			t.Errorf("expected path /graphql, got %s", receivedPath)
		}
	})
}

func TestClient_Query(t *testing.T) {
	t.Run("executes query and parses response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": {
					"user": {
						"id": "user-123",
						"name": "John Doe"
					}
				}
			}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", false)
		ctx := context.Background()

		var query struct {
			User struct {
				ID   string `graphql:"id"`
				Name string `graphql:"name"`
			} `graphql:"user"`
		}

		err := client.Query(ctx, &query, nil)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if query.User.ID != "user-123" {
			t.Errorf("expected ID user-123, got %s", query.User.ID)
		}
		if query.User.Name != "John Doe" {
			t.Errorf("expected Name 'John Doe', got %s", query.User.Name)
		}
	})

	t.Run("passes variables correctly", func(t *testing.T) {
		var receivedBody map[string]interface{}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &receivedBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":{}}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", false)
		ctx := context.Background()

		var query struct{}
		variables := map[string]interface{}{
			"id":    "123",
			"limit": 10,
		}

		_ = client.Query(ctx, &query, variables)

		if receivedBody["variables"] == nil {
			t.Error("expected variables to be sent")
		}
		vars, ok := receivedBody["variables"].(map[string]interface{})
		if !ok {
			t.Fatal("expected variables to be a map")
		}
		if vars["id"] != "123" {
			t.Errorf("expected variable id=123, got %v", vars["id"])
		}
		if vars["limit"] != float64(10) {
			t.Errorf("expected variable limit=10, got %v", vars["limit"])
		}
	})

	t.Run("handles GraphQL errors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": null,
				"errors": [
					{
						"message": "Field 'unknownField' not found",
						"locations": [{"line": 1, "column": 10}]
					}
				]
			}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", false)
		ctx := context.Background()

		var query struct{}
		err := client.Query(ctx, &query, nil)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "unknownField") {
			t.Errorf("expected error to contain 'unknownField', got: %v", err)
		}
	})

	t.Run("handles HTTP errors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`Internal Server Error`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", false)
		ctx := context.Background()

		var query struct{}
		err := client.Query(ctx, &query, nil)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Slow response
			select {
			case <-r.Context().Done():
				return
			}
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", false)
		ctx, cancel := context.WithCancel(context.Background())

		// Cancel immediately
		cancel()

		var query struct{}
		err := client.Query(ctx, &query, nil)

		if err == nil {
			t.Fatal("expected error due to context cancellation, got nil")
		}
	})

	t.Run("handles network errors", func(t *testing.T) {
		// Use invalid URL to cause network error
		client := NewClient("http://localhost:1", "token", false)
		ctx := context.Background()

		var query struct{}
		err := client.Query(ctx, &query, nil)

		if err == nil {
			t.Fatal("expected network error, got nil")
		}
	})
}

func TestClient_Mutate(t *testing.T) {
	t.Run("executes mutation and parses response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": {
					"createUser": {
						"id": "new-user-456",
						"name": "Jane Doe"
					}
				}
			}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", false)
		ctx := context.Background()

		var mutation struct {
			CreateUser struct {
				ID   string `graphql:"id"`
				Name string `graphql:"name"`
			} `graphql:"createUser"`
		}

		err := client.Mutate(ctx, &mutation, nil)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mutation.CreateUser.ID != "new-user-456" {
			t.Errorf("expected ID new-user-456, got %s", mutation.CreateUser.ID)
		}
		if mutation.CreateUser.Name != "Jane Doe" {
			t.Errorf("expected Name 'Jane Doe', got %s", mutation.CreateUser.Name)
		}
	})

	t.Run("passes mutation variables correctly", func(t *testing.T) {
		var receivedBody map[string]interface{}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &receivedBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":{"createUser":{"id":"123"}}}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", false)
		ctx := context.Background()

		var mutation struct {
			CreateUser struct {
				ID string `graphql:"id"`
			} `graphql:"createUser(input: $input)"`
		}

		variables := map[string]interface{}{
			"input": map[string]string{
				"name":  "Test User",
				"email": "test@example.com",
			},
		}

		_ = client.Mutate(ctx, &mutation, variables)

		if receivedBody["variables"] == nil {
			t.Error("expected variables to be sent")
		}
	})

	t.Run("handles mutation errors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": null,
				"errors": [
					{
						"message": "Validation error: email is required",
						"path": ["createUser"]
					}
				]
			}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", false)
		ctx := context.Background()

		var mutation struct{}
		err := client.Mutate(ctx, &mutation, nil)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "email is required") {
			t.Errorf("expected error to contain 'email is required', got: %v", err)
		}
	})

	t.Run("handles HTTP errors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(`Bad Gateway`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", false)
		ctx := context.Background()

		var mutation struct{}
		err := client.Mutate(ctx, &mutation, nil)

		if err == nil {
			t.Fatal("expected error, got nil")
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

		var mutation struct{}
		err := client.Mutate(ctx, &mutation, nil)

		if err == nil {
			t.Fatal("expected error due to context cancellation, got nil")
		}
	})

	t.Run("handles network errors", func(t *testing.T) {
		client := NewClient("http://localhost:1", "token", false)
		ctx := context.Background()

		var mutation struct{}
		err := client.Mutate(ctx, &mutation, nil)

		if err == nil {
			t.Fatal("expected network error, got nil")
		}
	})
}

func TestClient_RequestContentType(t *testing.T) {
	t.Run("sends correct content type headers", func(t *testing.T) {
		var receivedContentType string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedContentType = r.Header.Get("Content-Type")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":{}}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", false)
		ctx := context.Background()

		var query struct{}
		_ = client.Query(ctx, &query, nil)

		if receivedContentType != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", receivedContentType)
		}
	})

	t.Run("uses POST method for all requests", func(t *testing.T) {
		var receivedMethod string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":{}}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", false)
		ctx := context.Background()

		var query struct{}
		_ = client.Query(ctx, &query, nil)

		if receivedMethod != http.MethodPost {
			t.Errorf("expected POST method, got %s", receivedMethod)
		}
	})
}

func TestClient_MultipleGraphQLErrors(t *testing.T) {
	t.Run("handles multiple GraphQL errors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": null,
				"errors": [
					{
						"message": "First error message",
						"path": ["field1"]
					},
					{
						"message": "Second error message",
						"path": ["field2"]
					}
				]
			}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", false)
		ctx := context.Background()

		var query struct{}
		err := client.Query(ctx, &query, nil)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		// At minimum, the first error should be present
		if !strings.Contains(err.Error(), "First error") {
			t.Errorf("expected error to contain first error message, got: %v", err)
		}
	})
}

func TestClient_EmptyResponse(t *testing.T) {
	t.Run("handles empty data response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":{}}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", false)
		ctx := context.Background()

		var query struct{}
		err := client.Query(ctx, &query, nil)

		if err != nil {
			t.Fatalf("unexpected error for empty data: %v", err)
		}
	})

	t.Run("handles null data response without errors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":null}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", false)
		ctx := context.Background()

		var query struct{}
		err := client.Query(ctx, &query, nil)

		// null data without errors should not cause an error
		if err != nil {
			t.Fatalf("unexpected error for null data: %v", err)
		}
	})
}

func TestClient_NestedData(t *testing.T) {
	t.Run("parses deeply nested response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": {
					"company": {
						"id": "comp-1",
						"name": "Acme Corp",
						"employees": {
							"edges": [
								{
									"node": {
										"id": "emp-1",
										"name": "Alice"
									}
								},
								{
									"node": {
										"id": "emp-2",
										"name": "Bob"
									}
								}
							]
						}
					}
				}
			}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", false)
		ctx := context.Background()

		var query struct {
			Company struct {
				ID        string `graphql:"id"`
				Name      string `graphql:"name"`
				Employees struct {
					Edges []struct {
						Node struct {
							ID   string `graphql:"id"`
							Name string `graphql:"name"`
						} `graphql:"node"`
					} `graphql:"edges"`
				} `graphql:"employees"`
			} `graphql:"company"`
		}

		err := client.Query(ctx, &query, nil)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if query.Company.ID != "comp-1" {
			t.Errorf("expected company ID comp-1, got %s", query.Company.ID)
		}
		if query.Company.Name != "Acme Corp" {
			t.Errorf("expected company name 'Acme Corp', got %s", query.Company.Name)
		}
		if len(query.Company.Employees.Edges) != 2 {
			t.Errorf("expected 2 employees, got %d", len(query.Company.Employees.Edges))
		}
		if query.Company.Employees.Edges[0].Node.Name != "Alice" {
			t.Errorf("expected first employee 'Alice', got %s", query.Company.Employees.Edges[0].Node.Name)
		}
	})
}

func TestClient_NilVariables(t *testing.T) {
	t.Run("handles nil variables", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":{"test":"value"}}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", false)
		ctx := context.Background()

		var query struct {
			Test string `graphql:"test"`
		}

		err := client.Query(ctx, &query, nil)

		if err != nil {
			t.Fatalf("unexpected error with nil variables: %v", err)
		}
		if query.Test != "value" {
			t.Errorf("expected 'value', got %s", query.Test)
		}
	})

	t.Run("handles empty variables map", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":{"test":"value"}}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", false)
		ctx := context.Background()

		var query struct {
			Test string `graphql:"test"`
		}

		err := client.Query(ctx, &query, map[string]interface{}{})

		if err != nil {
			t.Fatalf("unexpected error with empty variables: %v", err)
		}
	})
}
