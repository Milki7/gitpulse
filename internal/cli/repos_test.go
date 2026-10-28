package cli_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Milki7/gitpulse/internal/cli"
	"github.com/Milki7/gitpulse/internal/model"
)

// fakeRepoLister is a test fake that returns canned data.
type fakeRepoLister struct {
	repos []model.Repo
	err   error
}

func (f *fakeRepoLister) ListRepos(_ context.Context, _ string, _ int) ([]model.Repo, error) {
	return f.repos, f.err
}

func TestRunRepos_Table(t *testing.T) {
	lister := &fakeRepoLister{
		repos: []model.Repo{
			{Name: "beta", Language: "Go", Stars: 10, Forks: 2, LastPush: time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)},
			{Name: "alpha", Language: "Rust", Stars: 50, Forks: 5, LastPush: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)},
		},
	}

	var buf bytes.Buffer
	err := cli.RunRepos(context.Background(), &buf, lister, "someone", 30, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()

	// Header must be present
	if !strings.Contains(out, "NAME") {
		t.Error("missing table header")
	}

	// alpha has more stars, so it must appear before beta
	alphaIdx := strings.Index(out, "alpha")
	betaIdx := strings.Index(out, "beta")
	if alphaIdx < 0 || betaIdx < 0 {
		t.Fatalf("missing repo names in output:\n%s", out)
	}
	if alphaIdx > betaIdx {
		t.Errorf("expected alpha before beta (sorted by stars desc), got:\n%s", out)
	}
}

func TestRunRepos_JSON(t *testing.T) {
	lister := &fakeRepoLister{
		repos: []model.Repo{
			{Name: "myrepo", Stars: 7},
		},
	}

	var buf bytes.Buffer
	err := cli.RunRepos(context.Background(), &buf, lister, "someone", 30, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"Name": "myrepo"`) {
		t.Errorf("expected JSON with repo name, got:\n%s", out)
	}
}

func TestRunRepos_Empty(t *testing.T) {
	lister := &fakeRepoLister{repos: []model.Repo{}}

	var buf bytes.Buffer
	err := cli.RunRepos(context.Background(), &buf, lister, "someone", 30, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Errorf("expected only header line for empty result, got %d lines:\n%s", len(lines), buf.String())
	}
}

func TestRunRepos_ListerError(t *testing.T) {
	lister := &fakeRepoLister{err: errors.New("API is down")}

	var buf bytes.Buffer
	err := cli.RunRepos(context.Background(), &buf, lister, "someone", 30, false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "API is down") {
		t.Errorf("error should propagate lister error, got: %v", err)
	}
}

func TestRunRepos_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	lister := &fakeRepoLister{err: context.Canceled}

	var buf bytes.Buffer
	err := cli.RunRepos(ctx, &buf, lister, "someone", 30, false)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}

