---
name: gitpulse-go-standards
description: Coding standards and anti-slop rules for the gitpulse Go CLI/TUI project (Cobra, Bubble Tea, go-github). Use this skill whenever writing, editing, reviewing, refactoring, or testing any Go code in this repo, adding dependencies, writing commit messages, or drafting README and docs, even if the user doesn't mention "standards" or "style". Also use it before declaring any task done.
---

# gitpulse Go standards

This is a portfolio project. The owner must be able to explain every line in an interview. Write code a careful intermediate Go developer would write and defend, not code that merely looks impressive.

## Working agreement

1. **Understand before editing.** Read the surrounding package and existing tests first. Match existing naming, error style, and structure.
2. **One goal per task.** Do not touch unrelated files. No drive-by refactors, renames, or reformatting.
3. **Plan before multi-file changes.** State which files change and why, then wait for approval.
4. **Ask instead of guessing.** If a requirement is ambiguous, ask one specific question.
5. **Explain choices briefly.** After non-trivial changes, give 2-4 sentences on the key design decision and the main alternative rejected.

## Code rules

### Simplicity
- Prefer the standard library. Do not add a dependency without asking, and say what it replaces.
- No premature abstraction. Create an interface only when there are two real implementations or a test needs a fake. Define interfaces in the package that uses them, and keep them small (1-3 methods).
- No `utils`, `helpers`, `common`, `manager`, or `base` packages or types. Name things for what they do.
- No speculative config, flags, options, or "future-proofing" that nothing uses yet.
- No unused parameters, dead code, or commented-out code.

### Errors
- Never swallow errors. No `_ = err` and no empty `if err != nil {}`.
- Wrap with context and `%w`: `fmt.Errorf("list repos for %q: %w", user, err)`. Context says what was being done, not "failed to".
- Don't log and return the same error. Handle it once, at the boundary (the Cobra command).
- Do not use `panic` outside `main` or truly impossible states.
- Fix root causes. Do not add nil checks or `recover` to silence a bug you don't understand; find out why the value is nil.

### Concurrency
- Every goroutine needs a clear owner, a way to stop (context), and a way to report errors. Prefer `errgroup` with `SetLimit`.
- Pass `context.Context` as the first parameter. Never store it in a struct.
- Never block in Bubble Tea's `Update`. Do I/O in a `tea.Cmd` and return the result as a message.
- Don't add concurrency unless there's independent I/O to overlap. Say why it is safe (what is shared, what protects it).
- Run `go test -race ./...` on anything concurrent.

### Structure and naming
- Layout: `cmd/gitpulse`, `internal/cli`, `internal/github`, `internal/stats`, `internal/tui`. `internal/stats` stays pure: no I/O, no GitHub types leaking in.
- Short, idiomatic names (`ctx`, `err`, `repo`). Exported names get doc comments starting with the name; unexported ones usually don't need any.
- Keep functions short enough to read in one screen. Return early; avoid deep nesting.
- Accept interfaces, return concrete types.

### Comments
- Comment the **why**, never the **what**. Delete any comment that restates the code, e.g. `// increment counter`, `// create a new client`.
- No section banners, no "This function does X" boilerplate, no emoji, and no TODOs without an owner and a reason.

## Tests

- Table-driven tests with descriptive case names. Test behavior, not implementation.
- Use fakes of the small interfaces, not mocks that assert call order. No real network calls in unit tests.
- Every bug fix gets a regression test that fails before the fix.
- Do not write tests that only assert the code does what it does (e.g. testing a getter). Cover edge cases: empty input, pagination boundaries, rate limits, cancelled context, cache expiry.
- Before finishing, run and report real output of:
  `gofmt -l .`, `go vet ./...`, `go build ./...`, `go test -race ./...`
  Never claim tests pass without having run them. If you couldn't run something, say so.

## Commits, docs, and prose

- Commits: imperative, specific, under 72 chars for the subject (`Add ETag caching to repo listing`). One logical change per commit. No "various improvements" or "update code".
- README and docs must describe only what the code actually does. Flag anything claimed but not implemented.
- Write plainly. Avoid filler and AI-tell phrasing: "robust", "seamless", "leverage", "comprehensive", "delve", "it's worth noting", "in today's fast-paced world". No inflated adjectives, no closing summaries that repeat what was just said, and no emoji in code, commits, or docs.
- Be specific and give numbers. Prefer "cuts startup from 800ms to 120ms (cached)" over "significantly faster".

## Before saying "done"

Run this checklist and report any item you skipped:

- [ ] Only the intended files changed (`git diff --stat`)
- [ ] No new dependencies, or each one was approved
- [ ] No ignored errors, dead code, or restating comments
- [ ] New logic has tests, and `go test -race ./...` output is shown
- [ ] Docs match behavior
- [ ] A short note on the main design decision and any known limitations

If something is incomplete, uncertain, or hacky, say so plainly. An honest "this part is untested" is better than a confident claim.
