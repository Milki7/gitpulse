# gitpulse design

## Goal

gitpulse is a command-line tool that shows a GitHub user's repositories, pull requests, and recent activity in the terminal. It has plain commands for scripting and an interactive TUI for browsing.

I'm building it to practice Go concurrency, API clients, testing with fakes, and shipping a CLI people can install.

## Non-goals (v1)

- No GitHub App or OAuth flow (a personal access token is enough)
- No write actions (no creating issues, PRs, or comments)
- No GitHub Enterprise support
- No database or background daemon

## Commands

| Command | What it does |
|---------|--------------|
| `gitpulse repos --user <name>` | List repos sorted by stars: name, language, stars, forks, last push |
| `gitpulse prs --user <name>` | List open PRs authored by the user and PRs awaiting their review |
| `gitpulse activity --user <name>` | Summarize recent events and a language breakdown |
| `gitpulse tui --user <name>` | Interactive view with tabs for repos, PRs, and activity |

Shared flags:
- `--token` overrides the `GITHUB_TOKEN` environment variable
- `--json` prints machine-readable output (non-TUI commands)
- `--no-cache` skips the local cache
- `--limit <n>` caps the number of results

## Package layout

```
cmd/gitpulse/     main.go: wires everything together, nothing else
internal/cli/     Cobra commands; parse flags, call services, print output
internal/github/  GitHub API wrapper and the interfaces the rest of the app uses
internal/stats/   Pure functions: sorting, aggregation, language breakdown
internal/tui/     Bubble Tea model, update, and view
internal/cache/   Disk cache with TTL and ETag support (added later)
```

Dependency rules:
- `stats` imports nothing from the other internal packages
- `cli` and `tui` depend on interfaces from `github`, never on `go-github` types directly
- `main.go` is the only place that builds the real client

## Key interfaces

```go
type RepoLister interface {
    ListRepos(ctx context.Context, user string) ([]Repo, error)
}

type PRLister interface {
    ListPRs(ctx context.Context, user string) ([]PR, error)
}
```

`Repo` and `PR` are my own small structs, so the rest of the app doesn't depend on the `go-github` library. Tests use fake implementations of these interfaces.

## Data flow

1. `cli` parses flags and builds a client from the token
2. The command calls a lister interface with a context and timeout
3. Results pass through `stats` for sorting and aggregation
4. `cli` prints a table or JSON; `tui` renders the same data in a view

The `activity` summary fetches repos, PRs, and events concurrently using `errgroup` with a limit and a 10-second timeout.

## TUI behavior

- Three tabs (Repos, PRs, Activity), switched with `tab` or `1`-`3`
- A spinner shows while data loads; errors appear in a status line instead of crashing
- Data loads in `tea.Cmd` functions, never inside `Update`
- Keys: `/` filter, `s` change sort, `enter` open in browser, `?` help, `q` quit

## Error handling

- Missing or invalid token: print a clear message explaining how to set `GITHUB_TOKEN`
- Rate limit hit: show when the limit resets
- Network errors: wrap with context and exit non-zero; the TUI shows them in the status line

## Caching

API responses are cached under `~/.cache/gitpulse` for 5 minutes. ETags and conditional requests avoid spending rate limit on unchanged data. The cache sits behind an interface so tests can use a fake.

## Testing plan

- Table-driven unit tests for everything in `stats`
- Fake `RepoLister` and `PRLister` for command tests (no network)
- Cache tests for TTL expiry and ETag handling
- `go test -race ./...` in CI

## Milestones

1. Scaffold and CLI skeleton builds
2. GitHub client and `repos` command
3. `prs` and `activity` commands
4. Concurrent fetching
5. Cache
6. TUI
7. Tests, CI, and release binaries
8. README with demo GIF

## Open questions

- Should `activity` include private repos when the token allows it?
- How many days of events should the summary cover by default?
- Is Homebrew distribution worth the effort for v1, or is `go install` enough?