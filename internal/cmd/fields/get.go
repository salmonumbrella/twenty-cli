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

var getCmd = &cobra.Command{
	Use:   "get <fieldId>",
	Short: "Get field metadata by ID",
	Long:  "Get detailed metadata for a specific field by its ID.",
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

func runGet(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	fieldID := args[0]

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	client := rt.RESTClient()

	field, err := client.GetField(ctx, fieldID)
	if err != nil {
		return fmt.Errorf("failed to get field: %w", err)
	}

	return outputFieldDetail(field, rt.Output, rt.Query)
}

func outputFieldDetail(field *types.FieldMetadata, format, query string) error {
	switch format {
	case "json":
		return outputFieldDetailJSON(field, os.Stdout, query)
	case "csv":
		return outputFieldDetailCSV(field, os.Stdout)
	default:
		return outputFieldDetailTable(field, os.Stdout)
	}
}

func outputFieldDetailTable(field *types.FieldMetadata, w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	// Field info
	fmt.Fprintf(tw, "Field:\t%s (%s)\n", field.Name, field.Label)
	fmt.Fprintf(tw, "Type:\t%s\n", field.Type)
	fmt.Fprintf(tw, "Object:\t%s\n", field.ObjectMetadataId)
	fmt.Fprintf(tw, "Custom:\t%v\n", field.IsCustom)
	fmt.Fprintf(tw, "Active:\t%v\n", field.IsActive)
	fmt.Fprintf(tw, "Nullable:\t%v\n", field.IsNullable)
	fmt.Fprintf(tw, "Description:\t%s\n", field.Description)

	return tw.Flush()
}

func outputFieldDetailJSON(field *types.FieldMetadata, w io.Writer, query string) error {
	return outfmt.WriteJSON(w, field, query)
}

func outputFieldDetailCSV(field *types.FieldMetadata, w io.Writer) error {
	headers := []string{"id", "name", "label", "type", "objectMetadataId", "isCustom", "isActive", "isNullable", "description"}
	rows := [][]string{
		{
			field.ID,
			field.Name,
			field.Label,
			field.Type,
			field.ObjectMetadataId,
			fmt.Sprintf("%v", field.IsCustom),
			fmt.Sprintf("%v", field.IsActive),
			fmt.Sprintf("%v", field.IsNullable),
			field.Description,
		},
	}
	return outfmt.WriteCSV(w, headers, rows)
}
