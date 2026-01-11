package api

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	t.Run("creates client with REST and GraphQL sub-clients", func(t *testing.T) {
		baseURL := "https://api.twenty.com"
		token := "test-token-abc123"
		debug := false

		client := NewClient(baseURL, token, debug)

		if client == nil {
			t.Fatal("expected client to be non-nil")
		}
		if client.REST == nil {
			t.Error("expected REST client to be non-nil")
		}
		if client.GraphQL == nil {
			t.Error("expected GraphQL client to be non-nil")
		}
		if client.debug != debug {
			t.Errorf("expected debug to be %v, got %v", debug, client.debug)
		}
	})

	t.Run("creates client with debug mode enabled", func(t *testing.T) {
		baseURL := "https://api.twenty.com"
		token := "test-token"
		debug := true

		client := NewClient(baseURL, token, debug)

		if client == nil {
			t.Fatal("expected client to be non-nil")
		}
		if client.debug != true {
			t.Error("expected debug to be true")
		}
	})

	t.Run("creates client with empty token", func(t *testing.T) {
		baseURL := "https://api.twenty.com"
		token := ""
		debug := false

		client := NewClient(baseURL, token, debug)

		// Client should still be created even with empty token
		// (authentication errors will occur during API calls)
		if client == nil {
			t.Fatal("expected client to be non-nil")
		}
		if client.REST == nil {
			t.Error("expected REST client to be non-nil")
		}
		if client.GraphQL == nil {
			t.Error("expected GraphQL client to be non-nil")
		}
	})

	t.Run("creates client with different base URLs", func(t *testing.T) {
		testCases := []struct {
			name    string
			baseURL string
		}{
			{name: "standard HTTPS", baseURL: "https://api.twenty.com"},
			{name: "with trailing slash", baseURL: "https://api.twenty.com/"},
			{name: "localhost", baseURL: "http://localhost:3000"},
			{name: "custom domain", baseURL: "https://twenty.example.org"},
			{name: "with port", baseURL: "https://api.twenty.com:8443"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				client := NewClient(tc.baseURL, "token", false)

				if client == nil {
					t.Fatal("expected client to be non-nil")
				}
				if client.REST == nil {
					t.Error("expected REST client to be non-nil")
				}
				if client.GraphQL == nil {
					t.Error("expected GraphQL client to be non-nil")
				}
			})
		}
	})
}

func TestClient_StructureIntegrity(t *testing.T) {
	t.Run("Client struct fields are accessible", func(t *testing.T) {
		client := NewClient("https://api.twenty.com", "token", true)

		// Verify that the public fields are accessible and properly typed
		// This ensures the Client struct maintains its contract
		_ = client.REST    // Should be *rest.Client
		_ = client.GraphQL // Should be *graphql.Client
	})
}

func TestClient_MultipleInstances(t *testing.T) {
	t.Run("multiple clients are independent", func(t *testing.T) {
		client1 := NewClient("https://api1.twenty.com", "token1", false)
		client2 := NewClient("https://api2.twenty.com", "token2", true)

		// Clients should be independent instances
		if client1 == client2 {
			t.Error("expected different client instances")
		}
		if client1.REST == client2.REST {
			t.Error("expected different REST client instances")
		}
		if client1.GraphQL == client2.GraphQL {
			t.Error("expected different GraphQL client instances")
		}
		if client1.debug == client2.debug {
			t.Error("expected different debug values")
		}
	})
}
