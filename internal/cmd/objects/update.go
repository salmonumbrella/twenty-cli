package objects

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
		Short:                 "Update a custom object",
		Resource:              "metadata/objects",
		IDArg:                 "object-id",
		InvalidObjectMessage:  "object payload must be a JSON object",
		MissingPayloadMessage: "missing JSON payload; use --data or --file",
		JSONOutput:            true,
		SkipErrorWrap:         true,
	})
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if cmd == nil {
		cmd = updateCmd
	}
	_ = cmd.Flags().Set("data", updateData)
	_ = cmd.Flags().Set("file", updateDataFile)
	return updateCmd.RunE(cmd, args)
}
