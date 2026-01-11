package objects

import (
	"github.com/spf13/cobra"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/builder"
)

var createCmd = newCreateCmd()

var (
	createData     string
	createDataFile string
)

func newCreateCmd() *cobra.Command {
	return builder.NewCreateCommand(builder.CreateConfig{
		Use:                   "create",
		Short:                 "Create a custom object",
		Resource:              "metadata/objects",
		InvalidObjectMessage:  "object payload must be a JSON object",
		MissingPayloadMessage: "missing JSON payload; use --data or --file",
		JSONOutput:            true,
		SkipErrorWrap:         true,
	})
}

func runCreate(cmd *cobra.Command, args []string) error {
	if cmd == nil {
		cmd = createCmd
	}
	_ = cmd.Flags().Set("data", createData)
	_ = cmd.Flags().Set("file", createDataFile)
	return createCmd.RunE(cmd, args)
}
