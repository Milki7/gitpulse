package cli

import (
	"errors"

	"github.com/spf13/cobra"
)

func newPRsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "prs [username]",
		Short: "List open pull requests authored by or awaiting review from the user",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("prs: not implemented")
		},
	}
}

