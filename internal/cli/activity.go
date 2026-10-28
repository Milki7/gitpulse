package cli

import (
	"errors"

	"github.com/spf13/cobra"
)

func newActivityCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "activity [username]",
		Short: "Summarize recent events and language breakdown",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("activity: not implemented")
		},
	}
}
