# gitpulse design

## Goal

gitpulse is a command-line tool that shows a GitHub user's repositories, pull requests, and recent activity in the terminal. It has plain commands for scripting and an interactive TUI for browsing.

I'm building it to practice Go concurrency, API clients, testing with fakes, and shipping a CLI people can install.

## Scope and milestones

The CLI comes first. The TUI is part of v1, but only after the CLI is finished and tested. If time runs short, the CLI ships as v1 and the TUI moves to v2.

| Milestone | Contents | Shippable? |
|-----------|----------|------------|
| 1. Scaffold | Module, Cobra skeleton, Makefile, builds cleanly | no |
| 2. Client and `repos` | GitHub client, pagination, `repos` command, table and `--json` output | no |
| 3. `prs` and `activity` | Search API for PRs, events for activity | no |
| 4. Concurrency | `activity` fetches in parallel with `errgroup` | no |
| 5. Cache | Simple JSON file cache with a 5-minute TTL | no |
| 6. CLI tests and CI | Fakes, table-driven tests, race detector, GitHub Actions | **yes: CLI v1** |
| 7. TUI | Bubble Tea app with three tabs | yes: full v1 |
| 8. Release | GoReleaser binaries, README, demo GIF | yes |

## Non-goals (v1)

- No GitHub App or OAuth flow (a personal access token is required)
- No write actions (no creating issues, PRs, or comments)
- No GitHub Enterprise support
- No database or background daemon
- No ETag or conditional-request caching (stretch goal, see below)

## Commands

| Command | What it does |
|---------|--------------|
| `gitpulse repos --user <name>` | List repos sorted by stars: name, language, stars, forks, last push |
| `gitpulse prs --user <name>` | List open PRs authored by the user and PRs awaiting their review |
| `gitpulse activity --user <name>` | Summarize recent events and a language breakdown |
| `gitpulse tui --user <name>` | Interactive view with tabs for repos, PRs, and activity (milestone 7) |

Shared flags:
- `--token` overrides the `GITHUB_TOKEN` environment variable
- `--json` prints machine-readable output (non-TUI commands)
- `--no-cache` skips the local cache
- `--limit <n>` caps the number of results (default 100)

## Package layout

```
cmd/gitpulse/     main.go: signal handling and wiring, nothing else
internal/model/   Plain structs only: Repo, PR, Event. No logic, no imports from other internal packages
internal/github/  Concrete Client that talks to the GitHub API and returns model types
internal/stats/   Pure functions over model types: sorting, aggregation, language breakdown
internal/cache/   Simple JSON file cache with TTL (milestone 5)
internal/cli/     Cobra commands; declare the interfaces they need, parse flags, print output
internal/tui/     Bubble Tea model, update, and view; declares the interfaces it needs (milestone 7)
```

Dependency rules:
- `model` imports nothing from this project
- `stats` imports only `model`, so it stays pure and trivial to test
- `github` imports `model` and the `go-github` library; no other package imports `go-github`
- `cli` and `tui` import `model` and `stats`, never `github` directly; `main.go` is the only place that connects the real client

## Interfaces

Interfaces are defined by the code that uses them, not by the code that implements them. `internal/github` exports a concrete `Client` type with methods like `ListRepos`. `internal/cli` declares only what it needs:

```go
// in internal/cli
type Client interface {
    ListRepos(ctx context.Context, user string, limit int) ([]model.Repo, error)
    ListPRs(ctx context.Context, user string, limit int) ([]model.PR, error)
    ListEvents(ctx context.Context, user string, limit int) ([]model.Event, error)
}
```

`internal/tui` declares its own, possibly smaller, interface the same way. Go's implicit interfaces mean `*github.Client` satisfies both with no extra code.

## Client construction

The token is a flag that Cobra parses after `main` starts, so `main` cannot build the client up front. Instead:

- `cli` defines `type ClientFactory func(token string) (Client, error)`
- `main.go` passes in a factory that builds the real `*github.Client`
- Each command calls the factory inside `RunE` once flags are parsed
- Tests pass a factory that returns a fake

## Data flow

1. `main` creates a context that is cancelled on Ctrl+C (`signal.NotifyContext`) and passes the factory to `cli`
2. A command parses flags, resolves the token, and calls the factory
3. The command calls the client with the context and a timeout
4. Results pass through `stats` for sorting and aggregation
5. `cli` prints a table or JSON; `tui` renders the same data in a view

`activity` fetches repos, PRs, and events concurrently using `errgroup` with `SetLimit` and a 10-second timeout.

## GitHub API notes

**Pagination.** Requests use `per_page=100`. The client follows pages until it reaches `--limit` or runs out of results. A token is required, since the unauthenticated rate limit is too low for this to be usable.

**Pull requests.** GitHub has no per-user PR listing endpoint. The client uses the Search API with queries such as `is:pr author:<user> is:open` and `is:pr review-requested:<user> is:open`. The Search API has a much stricter rate limit than the normal REST API, so `prs` should make as few search calls as possible and report rate-limit errors clearly. Check GitHub's current docs for the exact limits when implementing.

**Events.** `activity` uses the user events endpoint, which only returns a limited window of recent events. The summary says "recent activity", not full history, and the docs state the window.

## Cancellation and errors

- Ctrl+C cancels the root context, which stops in-flight requests and `errgroup` workers
- Missing or invalid token: print a clear message explaining how to set `GITHUB_TOKEN`
- Rate limit hit: show when the limit resets
- Network errors: wrap with context and exit non-zero
- In the TUI, errors appear in a status line instead of crashing

## Caching

v1 uses a simple JSON file cache under `os.UserCacheDir()/gitpulse`. Each entry stores the response and a timestamp; entries older than 5 minutes are ignored. Writes are atomic (write to a temp file, then rename). The cache sits behind a small interface so tests can fake it, and `--no-cache` bypasses it.

Stretch goal: ETags and conditional requests (`If-None-Match`) through a custom HTTP transport, so unchanged data doesn't count against the rate limit. This is deferred until the rest of v1 is done.

## TUI behavior (milestone 7)

- Three tabs (Repos, PRs, Activity), switched with `tab` or `1`-`3`
- A spinner shows while data loads
- Data loads in `tea.Cmd` functions, never inside `Update`
- Handles window resizing
- Keys: `/` filter, `s` change sort, `enter` open in browser, `?` help, `q` quit

## Testing plan

- Table-driven unit tests for everything in `stats`
- Fake `Client` for command tests (no network)
- Cache tests for TTL expiry and atomic writes
- A test that a cancelled context stops the concurrent `activity` fetch
- `go test -race ./...` locally and in CI

## Open questions

- Should `activity` include private repos when the token allows it?
- How many days of events should the summary cover by default?
- Is Homebrew distribution worth the effort for v1, or is `go install` enough?