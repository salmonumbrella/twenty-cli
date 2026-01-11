package shared

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// BuildQueryParams builds query parameters with limit, cursor, filter, and custom params.
func BuildQueryParams(limit int, cursor, filter, filterFile string, params []string) (url.Values, error) {
	values := url.Values{}
	if limit > 0 {
		values.Set("limit", fmt.Sprintf("%d", limit))
	}
	if cursor != "" {
		values.Set("starting_after", cursor)
	}

	if filter != "" || filterFile != "" {
		raw, err := ReadJSONInput(filter, filterFile)
		if err != nil {
			return nil, err
		}
		if raw != nil {
			encoded, err := json.Marshal(raw)
			if err != nil {
				return nil, fmt.Errorf("encode filter: %w", err)
			}
			values.Set("filter", string(encoded))
		}
	}

	for _, p := range params {
		key, val, ok := strings.Cut(p, "=")
		if !ok {
			return nil, fmt.Errorf("invalid param %q (expected key=value)", p)
		}
		values.Add(key, val)
	}

	return values, nil
}
