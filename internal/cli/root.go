package cli

import "github.com/spf13/cobra"

// NewRootCmd builds the top-level gitpulse command with persistent flags
// and all subcommands attached.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gitpulse",
		Short: "GitHub stats in your terminal",

		// SilenceUsage: prevents Cobra from printing usage on every RunE
		// error. Application errors like "not implemented" are not usage
		// mistakes and should not trigger the help text.
		SilenceUsage: true,

		// SilenceErrors: prevents Cobra from printing the error itself.
		// main.go handles all error output to stderr exactly once.
		SilenceErrors: true,
	}

	cmd.PersistentFlags().String("token", "", "GitHub personal access token (overrides GITHUB_TOKEN)")
	cmd.PersistentFlags().Bool("json", false, "print machine-readable JSON output")
	cmd.PersistentFlags().Bool("no-cache", false, "skip the local cache")
	cmd.PersistentFlags().Int("limit", 30, "maximum number of results")

	cmd.AddCommand(newReposCmd(), newPRsCmd(), newActivityCmd())
	return cmd
}

