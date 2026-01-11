package rest

import (
	"fmt"
	"net/url"
)

// ListOptions configures list endpoints.
type ListOptions struct {
	Limit   int
	Cursor  string
	Filter  map[string]interface{}
	Sort    string
	Order   string
	Fields  string
	Include []string
	Depth   int // 0 = no relations, 1 = include relations
	Params  url.Values
}

func applyListParams(path string, opts *ListOptions) (string, error) {
	if opts == nil {
		return path, nil
	}

	params := url.Values{}
	if opts.Limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", opts.Limit))
	}
	if opts.Cursor != "" {
		params.Set("starting_after", opts.Cursor)
	}
	if opts.Sort != "" {
		params.Set("order_by", opts.Sort)
	}
	if opts.Order != "" {
		params.Set("order_by_direction", opts.Order)
	}
	if opts.Fields != "" {
		params.Set("fields", opts.Fields)
	}
	// Twenty API uses 'depth' to fetch relations (0=none, 1=include)
	// When --include is specified, set depth=1 to fetch all relations
	if len(opts.Include) > 0 {
		params.Set("depth", "1")
	} else if opts.Depth > 0 {
		params.Set("depth", fmt.Sprintf("%d", opts.Depth))
	}
	// Twenty REST API uses filter[field][operator]:value format
	// e.g., filter[emails][primaryEmail][ilike]:%test%
	if len(opts.Filter) > 0 {
		addFilterParams(params, "filter", opts.Filter)
	}
	for key, vals := range opts.Params {
		for _, val := range vals {
			params.Add(key, val)
		}
	}

	if len(params) > 0 {
		return path + "?" + params.Encode(), nil
	}
	return path, nil
}

// addFilterParams recursively adds filter parameters in the format:
// filter[field][operator]:value or filter[field][subfield][operator]:value
func addFilterParams(params url.Values, prefix string, filter map[string]interface{}) {
	for key, value := range filter {
		newPrefix := fmt.Sprintf("%s[%s]", prefix, key)
		switch v := value.(type) {
		case map[string]interface{}:
			addFilterParams(params, newPrefix, v)
		case map[string]string:
			for op, val := range v {
				params.Set(fmt.Sprintf("%s[%s]", newPrefix, op), val)
			}
		case string:
			params.Set(newPrefix, v)
		default:
			params.Set(newPrefix, fmt.Sprintf("%v", v))
		}
	}
}
