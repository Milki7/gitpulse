package cli

import (
	"errors"

	"github.com/spf13/cobra"
)

func newReposCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "repos [username]",
		Short: "List repositories sorted by stars",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("repos: not implemented")
		},
	}
}

