package fields

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/shared"
	"github.com/salmonumbrella/twenty-cli/internal/outfmt"
	"github.com/salmonumbrella/twenty-cli/internal/types"
)

var listCmd = &cobra.Command{
	Use:   "list [objectName]",
	Short: "List all fields",
	Long: `List field metadata. If an object name is provided, lists only fields for that object.
Otherwise, lists all fields across all objects in your Twenty workspace.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runList,
}

func runList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	client := rt.RESTClient()

	// If object name is provided, get fields from that object
	if len(args) > 0 {
		objectName := args[0]
		obj, err := client.GetObject(ctx, objectName)
		if err != nil {
			return fmt.Errorf("failed to get object: %w", err)
		}
		return outputFields(obj.Fields, rt.Output, rt.Query)
	}

	// Otherwise, list all fields
	fields, err := client.ListFields(ctx)
	if err != nil {
		return fmt.Errorf("failed to list fields: %w", err)
	}

	return outputFields(fields, rt.Output, rt.Query)
}

func outputFields(fields []types.FieldMetadata, format, query string) error {
	switch format {
	case "json":
		return outputFieldsJSON(fields, os.Stdout, query)
	case "yaml":
		return outfmt.WriteYAML(os.Stdout, fields, query)
	case "csv":
		return outputFieldsCSV(fields, os.Stdout)
	default:
		return outputFieldsTable(fields, os.Stdout)
	}
}

func outputFieldsTable(fields []types.FieldMetadata, w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tNAME\tLABEL\tTYPE\tOBJECT")
	for _, field := range fields {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			field.ID,
			field.Name,
			field.Label,
			field.Type,
			field.ObjectMetadataId,
		)
	}
	return tw.Flush()
}

func outputFieldsJSON(fields []types.FieldMetadata, w io.Writer, query string) error {
	return outfmt.WriteJSON(w, fields, query)
}

func outputFieldsCSV(fields []types.FieldMetadata, w io.Writer) error {
	headers := []string{"id", "name", "label", "type", "objectMetadataId", "isCustom", "isActive", "isNullable"}
	rows := make([][]string, len(fields))
	for i, field := range fields {
		rows[i] = []string{
			field.ID,
			field.Name,
			field.Label,
			field.Type,
			field.ObjectMetadataId,
			fmt.Sprintf("%v", field.IsCustom),
			fmt.Sprintf("%v", field.IsActive),
			fmt.Sprintf("%v", field.IsNullable),
		}
	}
	return outfmt.WriteCSV(w, headers, rows)
}
