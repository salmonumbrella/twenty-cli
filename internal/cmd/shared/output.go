package shared

import (
	"io"

	"github.com/salmonumbrella/twenty-cli/internal/outfmt"
)

// WriteJSONOutput writes JSON output for commands that only support JSON.
// Text and empty outputs are treated as JSON for compatibility.
func WriteJSONOutput(w io.Writer, format, query string, data interface{}) error {
	switch format {
	case "", "json", "text":
		return outfmt.WriteJSON(w, data, query)
	default:
		return outfmt.WriteJSON(w, data, query)
	}
}

// WriteRawOutput writes raw JSON data in the requested format.
func WriteRawOutput(w io.Writer, format, query string, data interface{}) error {
	switch format {
	case "json":
		return outfmt.WriteJSON(w, data, query)
	case "yaml":
		return outfmt.WriteYAMLFromJSON(w, data)
	case "csv":
		return outfmt.WriteCSVFromJSON(w, data)
	default:
		return outfmt.WriteTableFromJSON(w, data)
	}
}
