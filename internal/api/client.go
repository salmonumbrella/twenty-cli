package api

import (
	"github.com/salmonumbrella/twenty-cli/internal/api/graphql"
	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
)

// Client provides unified access to both REST and GraphQL APIs
type Client struct {
	REST    *rest.Client
	GraphQL *graphql.Client
	debug   bool
}

// NewClient creates a new unified API client with both REST and GraphQL support
func NewClient(baseURL, token string, debug bool) *Client {
	return &Client{
		REST:    rest.NewClient(baseURL, token, debug),
		GraphQL: graphql.NewClient(baseURL, token, debug),
		debug:   debug,
	}
}
