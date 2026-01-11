package shared

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/salmonumbrella/twenty-cli/internal/api/graphql"
	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/auth"
	"github.com/salmonumbrella/twenty-cli/internal/secrets"
)

// Hook for testing
var ensureKeychainAccessFunc = secrets.EnsureKeychainAccess

// Runtime captures resolved global CLI settings.
type Runtime struct {
	BaseURL string
	Output  string
	Query   string
	Debug   bool
	NoRetry bool
}

// ResolveRuntime reads runtime settings from viper.
func ResolveRuntime() Runtime {
	return Runtime{
		BaseURL: viper.GetString("base_url"),
		Output:  viper.GetString("output"),
		Query:   viper.GetString("query"),
		Debug:   viper.GetBool("debug"),
		NoRetry: viper.GetBool("no_retry"),
	}
}

// AuthRuntime includes resolved auth data and lazily built clients.
type AuthRuntime struct {
	Runtime
	Profile string
	Token   string

	restClient    *rest.Client
	graphqlClient *graphql.Client
}

// RequireAuthRuntime resolves runtime settings and requires auth.
func RequireAuthRuntime() (*AuthRuntime, error) {
	// Pre-flight: ensure keychain is accessible before attempting auth
	if err := ensureKeychainAccessIfNeeded(); err != nil {
		return nil, fmt.Errorf("keychain access: %w", err)
	}

	profile, token, err := auth.RequireToken()
	if err != nil {
		return nil, err
	}

	rt := ResolveRuntime()
	return &AuthRuntime{
		Runtime: rt,
		Profile: profile,
		Token:   token,
	}, nil
}

// ensureKeychainAccessIfNeeded checks keychain access for non-file backends.
func ensureKeychainAccessIfNeeded() error {
	backendInfo, err := secrets.ResolveKeyringBackendInfo()
	if err != nil {
		return fmt.Errorf("resolve keyring backend: %w", err)
	}
	// Skip check for file backend (no system keychain involved)
	if backendInfo.Value == "file" {
		return nil
	}
	return ensureKeychainAccessFunc()
}

// RESTClient returns a cached REST client.
func (r *AuthRuntime) RESTClient() *rest.Client {
	if r.restClient == nil {
		r.restClient = NewRESTClientWithOptions(r.BaseURL, r.Token, r.Debug, r.NoRetry)
	}
	return r.restClient
}

// GraphQLClient returns a cached GraphQL client.
func (r *AuthRuntime) GraphQLClient() *graphql.Client {
	if r.graphqlClient == nil {
		r.graphqlClient = graphql.NewClient(r.BaseURL, r.Token, r.Debug)
	}
	return r.graphqlClient
}
