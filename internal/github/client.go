package github

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
	perPage := min(limit, 100)

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

// ListPRs returns open pull requests authored by user or awaiting user review, up to limit.
func (c *Client) ListPRs(ctx context.Context, user string, limit int) ([]model.PR, error) {
	seen := make(map[string]bool)
	var allPRs []model.PR

	authored, err := c.searchPRs(ctx, fmt.Sprintf("is:pr is:open author:%s", user), "Author", limit, seen)
	if err != nil {
		return nil, wrapGitHubError("list authored PRs", user, err)
	}
	allPRs = append(allPRs, authored...)

	if remaining := limit - len(allPRs); remaining > 0 {
		reviewReq, err := c.searchPRs(ctx, fmt.Sprintf("is:pr is:open review-requested:%s", user), "Reviewer", remaining, seen)
		if err != nil {
			return nil, wrapGitHubError("list review-requested PRs", user, err)
		}
		allPRs = append(allPRs, reviewReq...)
	}

	return allPRs, nil
}

func (c *Client) searchPRs(ctx context.Context, query string, role string, limit int, seen map[string]bool) ([]model.PR, error) {
	perPage := min(limit, 100)
	opts := &gh.SearchOptions{
		ListOptions: gh.ListOptions{
			PerPage: perPage,
		},
	}

	var prs []model.PR
	for len(prs) < limit {
		result, resp, err := c.gh.Search.Issues(ctx, query, opts)
		if err != nil {
			return nil, err
		}

		for _, issue := range result.Issues {
			url := issue.GetHTMLURL()
			if seen[url] {
				continue
			}
			seen[url] = true

			var createdAt time.Time
			if t := issue.GetCreatedAt(); !t.IsZero() {
				createdAt = t.GetTime().UTC()
			}

			author := ""
			if u := issue.GetUser(); u != nil {
				author = u.GetLogin()
			}

			prs = append(prs, model.PR{
				Number:    issue.GetNumber(),
				Title:     issue.GetTitle(),
				Repo:      repoFromIssue(issue),
				Author:    author,
				Role:      role,
				CreatedAt: createdAt,
				URL:       url,
			})

			if len(prs) >= limit {
				return prs, nil
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return prs, nil
}

func repoFromIssue(issue *gh.Issue) string {
	if repoURL := issue.GetRepositoryURL(); repoURL != "" {
		const prefix = "https://api.github.com/repos/"
		if strings.HasPrefix(repoURL, prefix) {
			return strings.TrimPrefix(repoURL, prefix)
		}
	}
	parts := strings.Split(issue.GetHTMLURL(), "/")
	if len(parts) >= 5 {
		return parts[3] + "/" + parts[4]
	}
	return ""
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
