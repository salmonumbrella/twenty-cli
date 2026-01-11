package shared

import (
	"github.com/spf13/viper"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
)

// NewRESTClient creates a REST client with the global configuration options.
// It reads debug and no_retry flags from viper.
func NewRESTClient(baseURL, token string) *rest.Client {
	debug := viper.GetBool("debug")
	noRetry := viper.GetBool("no_retry")

	return NewRESTClientWithOptions(baseURL, token, debug, noRetry)
}

// NewRESTClientWithOptions creates a REST client using explicit options.
func NewRESTClientWithOptions(baseURL, token string, debug, noRetry bool) *rest.Client {
	var opts []rest.ClientOption
	if noRetry {
		opts = append(opts, rest.WithNoRetry())
	}

	return rest.NewClient(baseURL, token, debug, opts...)
}
