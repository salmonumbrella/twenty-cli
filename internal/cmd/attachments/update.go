package attachments

import (
	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/builder"
)

var updateCmd = newUpdateCmd()

var (
	updateData     string
	updateDataFile string
)

func newUpdateCmd() *cobra.Command {
	return builder.NewUpdateCommand(builder.UpdateConfig{
		Use:                   "update",
		Short:                 "Update an attachment",
		Resource:              "attachments",
		AllowNonObject:        true,
		MissingPayloadMessage: "missing JSON payload; use --data or --file",
		JSONOutput:            true,
		SkipErrorWrap:         true,
	})
}

func runUpdate(cmd *cobra.Command, args []string) error {
	_ = cmd.Flags().Set("data", updateData)
	_ = cmd.Flags().Set("file", updateDataFile)
	return updateCmd.RunE(cmd, args)
}
