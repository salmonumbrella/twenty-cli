package outfmt

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/itchyny/gojq"
	"gopkg.in/yaml.v3"
)

// WriteJSON writes JSON output, applying an optional jq-style query.
func WriteJSON(w io.Writer, data interface{}, query string) error {
	payload, err := normalizeJSON(data)
	if err != nil {
		return err
	}

	if query != "" {
		payload, err = applyQuery(payload, query)
		if err != nil {
			return err
		}
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(payload)
}

// WriteYAML writes YAML output, applying an optional jq-style query.
func WriteYAML(w io.Writer, data interface{}, query string) error {
	payload, err := normalizeJSON(data)
	if err != nil {
		return err
	}

	if query != "" {
		payload, err = applyQuery(payload, query)
		if err != nil {
			return err
		}
	}

	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	defer enc.Close()
	return enc.Encode(payload)
}

func normalizeJSON(data interface{}) (interface{}, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal json: %w", err)
	}

	var out interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("unmarshal json: %w", err)
	}
	return out, nil
}

func applyQuery(data interface{}, query string) (interface{}, error) {
	q, err := gojq.Parse(query)
	if err != nil {
		return nil, fmt.Errorf("parse query: %w", err)
	}

	iter := q.Run(data)
	var results []interface{}
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return nil, err
		}
		results = append(results, v)
	}

	if len(results) == 1 {
		return results[0], nil
	}
	return results, nil
}
