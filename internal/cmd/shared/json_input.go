package shared

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// ReadJSONInput reads JSON data from a string or file (use "-" for stdin).
func ReadJSONInput(data, file string) (interface{}, error) {
	if data == "" && file == "" {
		return nil, nil
	}

	var raw []byte
	if file != "" {
		var err error
		if file == "-" {
			raw, err = io.ReadAll(os.Stdin)
		} else {
			raw, err = os.ReadFile(file)
		}
		if err != nil {
			return nil, fmt.Errorf("read json file: %w", err)
		}
	} else {
		raw = []byte(data)
	}

	var out interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return out, nil
}

// ReadJSONMap reads JSON data and ensures it is an object.
func ReadJSONMap(data, file string) (map[string]interface{}, error) {
	raw, err := ReadJSONInput(data, file)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, nil
	}
	obj, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected JSON object")
	}
	return obj, nil
}

// ReadJSONMapRequired reads JSON object data and errors when missing.
func ReadJSONMapRequired(data, file string) (map[string]interface{}, error) {
	obj, err := ReadJSONMap(data, file)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, fmt.Errorf("either --data or --file is required")
	}
	return obj, nil
}
