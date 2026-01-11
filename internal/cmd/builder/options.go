// Package builder provides generic command builders for Twenty CLI commands.
//
// This package uses Go generics to eliminate boilerplate code when creating
// list, get, create, update, and delete commands for API resources. Instead
// of writing repetitive command logic for each resource type, you configure
// a builder with the resource-specific details and it generates a fully
// functional Cobra command.
//
// # List Commands
//
// Use [NewListCommand] to create commands that list resources with pagination,
// filtering, sorting, and multiple output formats:
//
//	var listCmd = builder.NewListCommand(builder.ListConfig[types.Person]{
//	    Use:   "list",
//	    Short: "List people",
//	    ListFunc: func(ctx context.Context, client *rest.Client, opts *rest.ListOptions) (*types.ListResponse[types.Person], error) {
//	        return client.ListPeople(ctx, opts)
//	    },
//	    TableHeaders: []string{"ID", "NAME", "EMAIL"},
//	    TableRow: func(p types.Person) []string {
//	        return []string{p.ID, p.Name.FirstName, p.Email}
//	    },
//	})
//
// # Get Commands
//
// Use [NewGetCommand] to create commands that retrieve a single resource by ID:
//
//	var getCmd = builder.NewGetCommand(builder.GetConfig[types.Person]{
//	    Use:   "get",
//	    Short: "Get a person",
//	    GetFunc: func(ctx context.Context, client *rest.Client, id string) (*types.Person, error) {
//	        return client.GetPerson(ctx, id)
//	    },
//	    TableHeaders: []string{"FIELD", "VALUE"},
//	    TableRows: func(p types.Person) [][]string {
//	        return [][]string{
//	            {"ID", p.ID},
//	            {"Name", p.Name.FirstName},
//	        }
//	    },
//	})
//
// # CRUD Commands
//
// Use [NewCreateCommand], [NewUpdateCommand], and [NewDeleteCommand] for
// standard create, update, and delete operations. These accept JSON data
// via flags or stdin.
package builder

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/api/rest"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
)

// ListOptions holds the common flags shared by all list commands.
// These options control pagination, filtering, sorting, field selection,
// and relation inclusion for API list operations.
type ListOptions struct {
	// Limit is the maximum number of records to return per request.
	// Default is 20.
	Limit int

	// Cursor is the pagination cursor for fetching the next page of results.
	// This is typically obtained from the previous response's PageInfo.
	Cursor string

	// All enables fetching all records by automatically following pagination.
	// When true, the command will make multiple API requests until all
	// records are retrieved.
	All bool

	// Filter is a JSON string containing filter conditions.
	// Example: '{"email":{"like":"%@example.com"}}'
	Filter string

	// FilterFile is the path to a JSON file containing filter conditions.
	// Use "-" to read from stdin.
	FilterFile string

	// Params contains additional query parameters as key=value pairs.
	// These are passed directly to the API request.
	Params []string

	// Sort is the field name to sort results by.
	Sort string

	// Order is the sort direction: "asc" or "desc".
	Order string

	// Fields is a comma-separated list of fields to include in the response.
	// When empty, all fields are returned.
	Fields string

	// Include is a comma-separated list of relations to include.
	// Example: "company,notes"
	Include string
}

// RegisterFlags adds the standard list command flags to a Cobra command.
// This includes flags for pagination (--limit, --cursor, --all), filtering
// (--filter, --filter-file, --param), sorting (--sort, --order), and
// response customization (--fields, --include).
func (o *ListOptions) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().IntVarP(&o.Limit, "limit", "l", 20, "max records to return")
	cmd.Flags().StringVar(&o.Cursor, "cursor", "", "pagination cursor")
	cmd.Flags().BoolVar(&o.All, "all", false, "fetch all records")
	cmd.Flags().StringVar(&o.Filter, "filter", "", "JSON filter")
	cmd.Flags().StringVar(&o.FilterFile, "filter-file", "", "JSON filter file (use - for stdin)")
	cmd.Flags().StringArrayVar(&o.Params, "param", nil, "additional query params (key=value)")
	cmd.Flags().StringVar(&o.Sort, "sort", "", "sort field")
	cmd.Flags().StringVar(&o.Order, "order", "", "sort order (asc or desc)")
	cmd.Flags().StringVar(&o.Fields, "fields", "", "fields to select (comma-separated)")
	cmd.Flags().StringVar(&o.Include, "include", "", "relations to include (comma-separated)")
}

// ToRESTOptions converts the CLI flags to a [rest.ListOptions] struct
// suitable for passing to the REST API client. It parses the filter JSON,
// builds query parameters, and splits comma-separated includes.
//
// Returns an error if the filter JSON is invalid or if query parameters
// cannot be parsed.
func (o *ListOptions) ToRESTOptions() (*rest.ListOptions, error) {
	filter, err := shared.ReadJSONMap(o.Filter, o.FilterFile)
	if err != nil {
		return nil, err
	}

	params, err := shared.BuildQueryParams(0, "", "", "", o.Params)
	if err != nil {
		return nil, err
	}

	var include []string
	if o.Include != "" {
		include = strings.Split(o.Include, ",")
	}

	return &rest.ListOptions{
		Limit:   o.Limit,
		Cursor:  o.Cursor,
		Filter:  filter,
		Sort:    o.Sort,
		Order:   o.Order,
		Fields:  o.Fields,
		Include: include,
		Params:  params,
	}, nil
}
