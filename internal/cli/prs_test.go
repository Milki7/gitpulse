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

type fakePRLister struct {
	prs []model.PR
	err error
}

func (f *fakePRLister) ListPRs(_ context.Context, _ string, _ int) ([]model.PR, error) {
	return f.prs, f.err
}

func TestRunPRs_Table(t *testing.T) {
	lister := &fakePRLister{
		prs: []model.PR{
			{
				Number:    42,
				Title:     "Add cache support",
				Repo:      "user/gitpulse",
				Author:    "alice",
				Role:      "Author",
				CreatedAt: time.Date(2025, 4, 10, 0, 0, 0, 0, time.UTC),
			},
			{
				Number:    101,
				Title:     "Fix race condition in runner",
				Repo:      "org/engine",
				Author:    "bob",
				Role:      "Reviewer",
				CreatedAt: time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	var buf bytes.Buffer
	err := cli.RunPRs(context.Background(), &buf, lister, "alice", 30, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()

	if !strings.Contains(out, "REPO") || !strings.Contains(out, "TITLE") || !strings.Contains(out, "ROLE") {
		t.Errorf("missing expected headers in output:\n%s", out)
	}

	if !strings.Contains(out, "user/gitpulse") || !strings.Contains(out, "#42") || !strings.Contains(out, "Author") {
		t.Errorf("missing PR 42 details in output:\n%s", out)
	}

	if !strings.Contains(out, "org/engine") || !strings.Contains(out, "#101") || !strings.Contains(out, "Reviewer") {
		t.Errorf("missing PR 101 details in output:\n%s", out)
	}
}

func TestRunPRs_JSON(t *testing.T) {
	lister := &fakePRLister{
		prs: []model.PR{
			{Number: 15, Title: "Refactor CLI", Repo: "user/tool", Role: "Author"},
		},
	}

	var buf bytes.Buffer
	err := cli.RunPRs(context.Background(), &buf, lister, "user", 30, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"Title": "Refactor CLI"`) || !strings.Contains(out, `"Number": 15`) {
		t.Errorf("expected JSON with PR fields, got:\n%s", out)
	}
}

func TestRunPRs_Empty(t *testing.T) {
	lister := &fakePRLister{prs: []model.PR{}}

	var buf bytes.Buffer
	err := cli.RunPRs(context.Background(), &buf, lister, "user", 30, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Errorf("expected only header line for empty result, got %d lines:\n%s", len(lines), buf.String())
	}
}

func TestRunPRs_ListerError(t *testing.T) {
	lister := &fakePRLister{err: errors.New("search rate limit reached")}

	var buf bytes.Buffer
	err := cli.RunPRs(context.Background(), &buf, lister, "user", 30, false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "search rate limit reached") {
		t.Errorf("expected error to propagate, got: %v", err)
	}
}

func TestRunPRs_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	lister := &fakePRLister{err: context.Canceled}

	var buf bytes.Buffer
	err := cli.RunPRs(ctx, &buf, lister, "user", 30, false)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}
