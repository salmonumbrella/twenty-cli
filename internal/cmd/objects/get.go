package objects

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
	Use:   "get <objectName>",
	Short: "Get object metadata by name",
	Long:  "Get detailed metadata for a specific object type including its fields.",
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

func runGet(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	objectName := args[0]

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	client := rt.RESTClient()

	obj, err := client.GetObject(ctx, objectName)
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}

	return outputObjectDetail(obj, rt.Output, rt.Query)
}

func outputObjectDetail(obj *types.ObjectMetadata, format, query string) error {
	switch format {
	case "json":
		return outputObjectDetailJSON(obj, os.Stdout, query)
	case "csv":
		return outputObjectDetailCSV(obj, os.Stdout)
	default:
		return outputObjectDetailTable(obj, os.Stdout)
	}
}

func outputObjectDetailTable(obj *types.ObjectMetadata, w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	// Object info
	fmt.Fprintf(tw, "Object:\t%s (%s)\n", obj.NameSingular, obj.LabelPlural)
	fmt.Fprintf(tw, "Description:\t%s\n", obj.Description)
	fmt.Fprintf(tw, "Custom:\t%v\n", obj.IsCustom)
	fmt.Fprintf(tw, "Active:\t%v\n", obj.IsActive)
	fmt.Fprintln(tw)

	// Fields section
	fmt.Fprintln(tw, "Fields:")
	fmt.Fprintln(tw, "NAME\tLABEL\tTYPE\tCUSTOM")
	for _, field := range obj.Fields {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%v\n",
			field.Name,
			field.Label,
			field.Type,
			field.IsCustom,
		)
	}

	return tw.Flush()
}

func outputObjectDetailJSON(obj *types.ObjectMetadata, w io.Writer, query string) error {
	return outfmt.WriteJSON(w, obj, query)
}

func outputObjectDetailCSV(obj *types.ObjectMetadata, w io.Writer) error {
	// Output the object's fields as CSV rows
	headers := []string{"fieldName", "fieldLabel", "fieldType", "isCustom", "isActive"}
	rows := make([][]string, len(obj.Fields))
	for i, field := range obj.Fields {
		rows[i] = []string{
			field.Name,
			field.Label,
			field.Type,
			fmt.Sprintf("%v", field.IsCustom),
			fmt.Sprintf("%v", field.IsActive),
		}
	}
	return outfmt.WriteCSV(w, headers, rows)
}
