package github

import (
	"context"
	"errors"
	"fmt"
	"time"

	gh "github.com/google/go-github/v72/github"

	"github.com/Milki7/gitpulse/internal/model"
)

// Client wraps the GitHub REST API.
type Client struct {
	gh *gh.Client
}

// NewClient creates a GitHub API client. If token is empty, the client
// is unauthenticated (60 requests/hour).
func NewClient(token string) *Client {
	c := gh.NewClient(nil)
	if token != "" {
		c = c.WithAuthToken(token)
	}
	return &Client{gh: c}
}

// ListRepos returns up to limit repositories for the given user,
// fetching multiple pages if necessary. Repos are returned in the
// API's default order (most recently pushed first); callers should
// sort as needed.
func (c *Client) ListRepos(ctx context.Context, user string, limit int) ([]model.Repo, error) {
	perPage := limit
	if perPage > 100 {
		perPage = 100
	}

	opts := &gh.RepositoryListByUserOptions{
		Sort:      "pushed",
		Direction: "desc",
		ListOptions: gh.ListOptions{
			PerPage: perPage,
		},
	}

	var repos []model.Repo
	for {
		ghRepos, resp, err := c.gh.Repositories.ListByUser(ctx, user, opts)
		if err != nil {
			return nil, wrapGitHubError("list repos", user, err)
		}

		for _, r := range ghRepos {
			if len(repos) >= limit {
				return repos, nil
			}
			repos = append(repos, repoFromGitHub(r))
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return repos, nil
}

func repoFromGitHub(r *gh.Repository) model.Repo {
	var lastPush time.Time
	if t := r.GetPushedAt(); !t.IsZero() {
		lastPush = t.GetTime().UTC()
	}

	return model.Repo{
		Name:     r.GetName(),
		Owner:    r.GetOwner().GetLogin(),
		Language: r.GetLanguage(),
		Stars:    r.GetStargazersCount(),
		Forks:    r.GetForksCount(),
		LastPush: lastPush,
		URL:      r.GetHTMLURL(),
	}
}

func wrapGitHubError(action, user string, err error) error {
	var rateLimitErr *gh.RateLimitError
	if errors.As(err, &rateLimitErr) {
		reset := rateLimitErr.Rate.Reset.GetTime().Format(time.RFC3339)
		return fmt.Errorf("%s for %q: rate limit exceeded, resets at %s: %w",
			action, user, reset, err)
	}
	return fmt.Errorf("%s for %q: %w", action, user, err)
}

