package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	ghclient "github.com/Milki7/gitpulse/internal/github"
	"github.com/Milki7/gitpulse/internal/model"
	"github.com/Milki7/gitpulse/internal/stats"
)

// RepoLister fetches repositories for a GitHub user.
type RepoLister interface {
	ListRepos(ctx context.Context, user string, limit int) ([]model.Repo, error)
}

func newReposCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "repos <username>",
		Short: "List repositories sorted by stars",
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
			return RunRepos(cmd.Context(), cmd.OutOrStdout(), client, args[0], limit, asJSON)
		},
	}
}

// RunRepos fetches repos from lister, sorts by stars, and writes the
// result to w as a table or JSON.
func RunRepos(ctx context.Context, w io.Writer, lister RepoLister, user string, limit int, asJSON bool) error {
	repos, err := lister.ListRepos(ctx, user, limit)
	if err != nil {
		return err
	}

	stats.SortByStars(repos)

	if asJSON {
		return writeJSON(w, repos)
	}
	return writeRepoTable(w, repos)
}

func writeRepoTable(w io.Writer, repos []model.Repo) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tLANGUAGE\tSTARS\tFORKS\tLAST PUSH")
	for _, r := range repos {
		lastPush := "—"
		if !r.LastPush.IsZero() {
			lastPush = r.LastPush.Format("2006-01-02")
		}
		lang := r.Language
		if lang == "" {
			lang = "—"
		}
		fmt.Fprintf(tw, "%s\t%s\t%d\t%d\t%s\n", r.Name, lang, r.Stars, r.Forks, lastPush)
	}
	return tw.Flush()
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
