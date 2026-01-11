package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate completion script for your shell.

To load completions:

Bash:
  $ source <(twenty completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ twenty completion bash > /etc/bash_completion.d/twenty
  # macOS:
  $ twenty completion bash > $(brew --prefix)/etc/bash_completion.d/twenty

Zsh:
  $ source <(twenty completion zsh)
  # To load completions for each session, execute once:
  $ twenty completion zsh > "${fpath[1]}/_twenty"

Fish:
  $ twenty completion fish | source
  # To load completions for each session, execute once:
  $ twenty completion fish > ~/.config/fish/completions/twenty.fish

PowerShell:
  PS> twenty completion powershell | Out-String | Invoke-Expression
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			return cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			return cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
