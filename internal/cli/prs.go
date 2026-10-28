package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	ghclient "github.com/Milki7/gitpulse/internal/github"
	"github.com/Milki7/gitpulse/internal/model"
)

// PRLister fetches pull requests for a GitHub user.
type PRLister interface {
	ListPRs(ctx context.Context, user string, limit int) ([]model.PR, error)
}

func newPRsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "prs <username>",
		Short: "List open pull requests authored by or awaiting review from the user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			token, _ := cmd.Flags().GetString("token")
			if token == "" {
				token = os.Getenv("GITHUB_TOKEN")
			}
			if token == "" {
				fmt.Fprintln(cmd.ErrOrStderr(), "Warning: no token set; requests are limited to 60/hour")
				fmt.Fprintln(cmd.ErrOrStderr(), "Set GITHUB_TOKEN or use --token to authenticate")
			}

			asJSON, _ := cmd.Flags().GetBool("json")
			limit, _ := cmd.Flags().GetInt("limit")

			client := ghclient.NewClient(token)
			return RunPRs(cmd.Context(), cmd.OutOrStdout(), client, args[0], limit, asJSON)
		},
	}
}

// RunPRs fetches PRs from lister and writes the result to w as a table or JSON.
func RunPRs(ctx context.Context, w io.Writer, lister PRLister, user string, limit int, asJSON bool) error {
	prs, err := lister.ListPRs(ctx, user, limit)
	if err != nil {
		return err
	}

	if asJSON {
		return writeJSON(w, prs)
	}
	return writePRTable(w, prs)
}

func writePRTable(w io.Writer, prs []model.PR) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "REPO\t#\tTITLE\tROLE\tCREATED")
	for _, p := range prs {
		created := "—"
		if !p.CreatedAt.IsZero() {
			created = p.CreatedAt.Format("2006-01-02")
		}
		repo := p.Repo
		if repo == "" {
			repo = "—"
		}
		fmt.Fprintf(tw, "%s\t#%d\t%s\t%s\t%s\n", repo, p.Number, p.Title, p.Role, created)
	}
	return tw.Flush()
}
