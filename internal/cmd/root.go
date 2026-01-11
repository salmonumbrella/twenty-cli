package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/salmonumbrella/twenty-cli/internal/cmd/attachments"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/auth"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/companies"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/config"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/favorites"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/fields"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/graphql"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/notes"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/objects"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/opportunities"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/people"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/records"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/rest"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/tasks"
	"github.com/salmonumbrella/twenty-cli/internal/cmd/webhooks"
)

var (
	cfgFile string
	output  string
	profile string
	query   string
	debug   bool
	noColor bool
	noRetry bool
	baseURL string
	version = "dev"
)

func SetVersion(v string) {
	version = v
	rootCmd.Version = v
}

var rootCmd = &cobra.Command{
	Use:   "twenty",
	Short: "CLI for Twenty CRM with 100% API coverage",
	Long:  `CLI for Twenty CRM with 100% API coverage`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	pf := rootCmd.PersistentFlags()
	pf.StringVar(&cfgFile, "config", "", "config file (default $HOME/.twenty.yaml)")
	pf.StringVarP(&output, "output", "o", "text", "output format: text, json, yaml, csv")
	pf.StringVarP(&profile, "profile", "p", "", "auth profile to use (defaults to active profile)")
	pf.StringVarP(&query, "query", "q", "", "jq-style query for JSON output")
	pf.BoolVar(&debug, "debug", false, "show API requests/responses")
	pf.BoolVar(&noColor, "no-color", false, "disable colored output")
	pf.BoolVar(&noRetry, "no-retry", false, "disable automatic retry on rate limiting (429, 502, 503, 504)")
	pf.StringVar(&baseURL, "base-url", "https://api.twenty.com", "API base URL")

	// Bind to viper (errors are programmer errors, safe to ignore)
	_ = viper.BindPFlag("output", pf.Lookup("output"))
	_ = viper.BindPFlag("profile", pf.Lookup("profile"))
	_ = viper.BindPFlag("query", pf.Lookup("query"))
	_ = viper.BindPFlag("debug", pf.Lookup("debug"))
	_ = viper.BindPFlag("no_color", pf.Lookup("no-color"))
	_ = viper.BindPFlag("no_retry", pf.Lookup("no-retry"))
	_ = viper.BindPFlag("base_url", pf.Lookup("base-url"))

	rootCmd.Version = version

	rootCmd.AddCommand(auth.Cmd)
	rootCmd.AddCommand(attachments.Cmd)
	rootCmd.AddCommand(companies.Cmd)
	rootCmd.AddCommand(config.Cmd)
	rootCmd.AddCommand(fields.Cmd)
	rootCmd.AddCommand(favorites.Cmd)
	rootCmd.AddCommand(graphql.Cmd)
	rootCmd.AddCommand(notes.Cmd)
	rootCmd.AddCommand(objects.Cmd)
	rootCmd.AddCommand(opportunities.Cmd)
	rootCmd.AddCommand(people.Cmd)
	rootCmd.AddCommand(records.Cmd)
	rootCmd.AddCommand(rest.Cmd)
	rootCmd.AddCommand(tasks.Cmd)
	rootCmd.AddCommand(webhooks.Cmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".twenty")
	}

	// Enable environment variables with TWENTY_ prefix
	viper.SetEnvPrefix("twenty")
	viper.AutomaticEnv()

	// Explicit env bindings for commonly used vars
	_ = viper.BindEnv("base_url", "TWENTY_BASE_URL")
	_ = viper.BindEnv("no_color", "TWENTY_NO_COLOR", "NO_COLOR")

	if err := viper.ReadInConfig(); err == nil {
		// Only show config file message for human-readable output
		out := viper.GetString("output")
		if out != "json" && out != "yaml" && out != "csv" {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}
