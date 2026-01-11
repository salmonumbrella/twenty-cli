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

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all object types",
	Long:  "List all object metadata (types) in your Twenty workspace.",
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	rt, err := shared.RequireAuthRuntime()
	if err != nil {
		return err
	}

	client := rt.RESTClient()

	objects, err := client.ListObjects(ctx)
	if err != nil {
		return fmt.Errorf("failed to list objects: %w", err)
	}

	return outputObjects(objects, rt.Output, rt.Query)
}

func outputObjects(objects []types.ObjectMetadata, format, query string) error {
	switch format {
	case "json":
		return outputObjectsJSON(objects, os.Stdout, query)
	case "yaml":
		return outfmt.WriteYAML(os.Stdout, objects, query)
	case "csv":
		return outputObjectsCSV(objects, os.Stdout)
	default:
		return outputObjectsTable(objects, os.Stdout)
	}
}

func outputObjectsTable(objects []types.ObjectMetadata, w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tLABEL\tCUSTOM\tFIELDS")
	for _, obj := range objects {
		fmt.Fprintf(tw, "%s\t%s\t%v\t%d\n",
			obj.NameSingular,
			obj.LabelPlural,
			obj.IsCustom,
			len(obj.Fields),
		)
	}
	return tw.Flush()
}

func outputObjectsJSON(objects []types.ObjectMetadata, w io.Writer, query string) error {
	return outfmt.WriteJSON(w, objects, query)
}

func outputObjectsCSV(objects []types.ObjectMetadata, w io.Writer) error {
	headers := []string{"name", "namePlural", "label", "labelPlural", "isCustom", "isActive", "fieldCount"}
	rows := make([][]string, len(objects))
	for i, obj := range objects {
		rows[i] = []string{
			obj.NameSingular,
			obj.NamePlural,
			obj.LabelSingular,
			obj.LabelPlural,
			fmt.Sprintf("%v", obj.IsCustom),
			fmt.Sprintf("%v", obj.IsActive),
			fmt.Sprintf("%d", len(obj.Fields)),
		}
	}
	return outfmt.WriteCSV(w, headers, rows)
}
