package graphql

import (
	"context"

	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
)

// Client wraps the shurcooL/graphql client with authentication
type Client struct {
	client *graphql.Client
	debug  bool
}

// NewClient creates a new GraphQL client for the Twenty CRM API
func NewClient(baseURL, token string, debug bool) *Client {
	// Create HTTP client with OAuth2 bearer token
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	// Create GraphQL client pointing to /graphql endpoint
	gqlClient := graphql.NewClient(baseURL+"/graphql", httpClient)

	return &Client{
		client: gqlClient,
		debug:  debug,
	}
}

// Query executes a GraphQL query
func (c *Client) Query(ctx context.Context, q interface{}, variables map[string]interface{}) error {
	return c.client.Query(ctx, q, variables)
}

// Mutate executes a GraphQL mutation
func (c *Client) Mutate(ctx context.Context, m interface{}, variables map[string]interface{}) error {
	return c.client.Mutate(ctx, m, variables)
}
