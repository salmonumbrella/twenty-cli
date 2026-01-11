package outfmt

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

// Format represents output format
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
	FormatCSV   Format = "csv"
	FormatText  Format = "text"
)

// Printer handles formatted output
type Printer struct {
	Format Format
	Writer io.Writer
}

// NewPrinter creates a new printer with the given format
func NewPrinter(format string) *Printer {
	return &Printer{
		Format: Format(format),
		Writer: os.Stdout,
	}
}

// Print outputs data in the configured format
func (p *Printer) Print(data interface{}) error {
	switch p.Format {
	case FormatJSON:
		return p.printJSON(data)
	case FormatYAML:
		return p.printYAML(data)
	default:
		return fmt.Errorf("table format requires PrintTable method")
	}
}

func (p *Printer) printJSON(data interface{}) error {
	encoder := json.NewEncoder(p.Writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (p *Printer) printYAML(data interface{}) error {
	encoder := yaml.NewEncoder(p.Writer)
	encoder.SetIndent(2)
	return encoder.Encode(data)
}

// PrintTable prints data in table format
func (p *Printer) PrintTable(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(p.Writer, 0, 0, 2, ' ', 0)

	// Print headers
	for i, h := range headers {
		if i > 0 {
			fmt.Fprint(w, "\t")
		}
		fmt.Fprint(w, h)
	}
	fmt.Fprintln(w)

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				fmt.Fprint(w, "\t")
			}
			fmt.Fprint(w, cell)
		}
		fmt.Fprintln(w)
	}

	w.Flush()
}

// WriteCSV writes data as CSV output.
// headers is a slice of column names; rows is a slice of records where each record
// is a slice of string values corresponding to the headers.
func WriteCSV(w io.Writer, headers []string, rows [][]string) error {
	writer := csv.NewWriter(w)
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("write csv row: %w", err)
		}
	}
	writer.Flush()
	return writer.Error()
}

// WriteCSVFromStruct writes a slice of structs as CSV output.
// It uses struct field names (or json tags) as headers and flattens the struct values.
func WriteCSVFromStruct(w io.Writer, data interface{}) error {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice {
		// For single item, wrap in a slice
		slice := reflect.MakeSlice(reflect.SliceOf(v.Type()), 1, 1)
		slice.Index(0).Set(v)
		v = slice
	}

	if v.Len() == 0 {
		return nil
	}

	// Get headers from the first element
	elem := v.Index(0)
	if elem.Kind() == reflect.Ptr {
		elem = elem.Elem()
	}
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("csv output requires struct type, got %s", elem.Kind())
	}

	headers := getStructHeaders(elem.Type())
	writer := csv.NewWriter(w)
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}

	// Write rows
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		row := getStructValues(elem)
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("write csv row: %w", err)
		}
	}

	writer.Flush()
	return writer.Error()
}

func getStructHeaders(t reflect.Type) []string {
	var headers []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" { // unexported
			continue
		}
		name := field.Tag.Get("json")
		if name == "" || name == "-" {
			name = field.Name
		} else {
			// Remove omitempty and other options
			if idx := len(name); idx > 0 {
				for j, c := range name {
					if c == ',' {
						name = name[:j]
						break
					}
				}
			}
		}
		headers = append(headers, name)
	}
	return headers
}

func getStructValues(v reflect.Value) []string {
	var values []string
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" { // unexported
			continue
		}
		fv := v.Field(i)
		values = append(values, formatValue(fv))
	}
	return values
}

func formatValue(v reflect.Value) string {
	if !v.IsValid() {
		return ""
	}
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return ""
		}
		return formatValue(v.Elem())
	case reflect.Struct:
		// For nested structs, try to get a reasonable string representation
		// Check for common patterns like Name field or String() method
		if v.Type().String() == "time.Time" {
			return fmt.Sprintf("%v", v.Interface())
		}
		// Try to convert to JSON for nested structs
		if b, err := json.Marshal(v.Interface()); err == nil {
			return string(b)
		}
		return fmt.Sprintf("%v", v.Interface())
	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return ""
		}
		if b, err := json.Marshal(v.Interface()); err == nil {
			return string(b)
		}
		return fmt.Sprintf("%v", v.Interface())
	case reflect.Map:
		if v.Len() == 0 {
			return ""
		}
		if b, err := json.Marshal(v.Interface()); err == nil {
			return string(b)
		}
		return fmt.Sprintf("%v", v.Interface())
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}

// WriteCSVFromJSON writes CSV output from json.RawMessage or []byte JSON data.
// It automatically detects the array structure within typical Twenty API responses
// (e.g., {"data":{"favorites":[...]}} or direct array).
func WriteCSVFromJSON(w io.Writer, data interface{}) error {
	var raw []byte
	switch v := data.(type) {
	case json.RawMessage:
		raw = v
	case []byte:
		raw = v
	default:
		// Try to marshal it
		var err error
		raw, err = json.Marshal(data)
		if err != nil {
			return fmt.Errorf("marshal data for csv: %w", err)
		}
	}

	// Parse the JSON
	var parsed interface{}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return fmt.Errorf("parse json for csv: %w", err)
	}

	// Find the array of records
	records := findRecordArray(parsed)
	if len(records) == 0 {
		return nil // No records to write
	}

	// Extract headers from the first record
	headers := extractHeaders(records[0])
	if len(headers) == 0 {
		return fmt.Errorf("no fields found in records")
	}

	// Build rows
	rows := make([][]string, 0, len(records))
	for _, record := range records {
		row := extractRow(record, headers)
		rows = append(rows, row)
	}

	return WriteCSV(w, headers, rows)
}

// WriteYAMLFromJSON writes YAML output from json.RawMessage or []byte JSON data.
func WriteYAMLFromJSON(w io.Writer, data interface{}) error {
	var raw []byte
	switch v := data.(type) {
	case json.RawMessage:
		raw = v
	case []byte:
		raw = v
	default:
		var err error
		raw, err = json.Marshal(data)
		if err != nil {
			return fmt.Errorf("marshal data for yaml: %w", err)
		}
	}

	var parsed interface{}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return fmt.Errorf("parse json for yaml: %w", err)
	}

	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	defer enc.Close()
	return enc.Encode(parsed)
}

// WriteTableFromJSON writes tabular output from json.RawMessage or []byte JSON data.
// It detects the array structure in typical Twenty API responses and renders a table.
func WriteTableFromJSON(w io.Writer, data interface{}) error {
	var raw []byte
	switch v := data.(type) {
	case json.RawMessage:
		raw = v
	case []byte:
		raw = v
	default:
		var err error
		raw, err = json.Marshal(data)
		if err != nil {
			return fmt.Errorf("marshal data for table: %w", err)
		}
	}

	var parsed interface{}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return fmt.Errorf("parse json for table: %w", err)
	}

	records := findRecordArray(parsed)
	if len(records) == 0 {
		return nil
	}

	headers := extractHeaders(records[0])
	if len(headers) == 0 {
		return fmt.Errorf("no fields found in records")
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for i, header := range headers {
		if i > 0 {
			fmt.Fprint(tw, "\t")
		}
		fmt.Fprint(tw, header)
	}
	fmt.Fprintln(tw)

	for _, record := range records {
		row := extractRow(record, headers)
		for i, cell := range row {
			if i > 0 {
				fmt.Fprint(tw, "\t")
			}
			fmt.Fprint(tw, cell)
		}
		fmt.Fprintln(tw)
	}

	return tw.Flush()
}

// findRecordArray finds the array of records in a typical API response.
// Handles structures like {"data":{"resourceName":[...]}} or direct arrays.
func findRecordArray(data interface{}) []map[string]interface{} {
	switch v := data.(type) {
	case []interface{}:
		// Direct array
		return toMapSlice(v)
	case map[string]interface{}:
		// Check for "data" wrapper
		if dataField, ok := v["data"]; ok {
			return findRecordArray(dataField)
		}
		// Check for any array field (e.g., "favorites", "attachments")
		for _, value := range v {
			if arr, ok := value.([]interface{}); ok {
				return toMapSlice(arr)
			}
		}
	}
	return nil
}

func toMapSlice(arr []interface{}) []map[string]interface{} {
	var result []map[string]interface{}
	for _, item := range arr {
		if m, ok := item.(map[string]interface{}); ok {
			result = append(result, m)
		}
	}
	return result
}

func extractHeaders(record map[string]interface{}) []string {
	headers := make([]string, 0, len(record))
	for key := range record {
		headers = append(headers, key)
	}
	// Sort for consistent output
	sortStrings(headers)
	return headers
}

func sortStrings(s []string) {
	for i := 0; i < len(s); i++ {
		for j := i + 1; j < len(s); j++ {
			if s[i] > s[j] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

func extractRow(record map[string]interface{}, headers []string) []string {
	row := make([]string, len(headers))
	for i, header := range headers {
		if val, ok := record[header]; ok {
			row[i] = formatJSONValue(val)
		}
	}
	return row
}

func formatJSONValue(val interface{}) string {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		return v
	case float64:
		// Check if it's a whole number
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%v", v)
	case bool:
		return fmt.Sprintf("%v", v)
	case map[string]interface{}, []interface{}:
		// Nested objects/arrays: serialize to JSON
		if b, err := json.Marshal(v); err == nil {
			return string(b)
		}
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
