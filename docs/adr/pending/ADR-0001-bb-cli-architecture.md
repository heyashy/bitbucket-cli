# ADR-0001: bb CLI Architecture — Cobra + Viper + Bubble Tea

**Date:** 2026-03-11
**Status:** pending
**Pattern:** Adapter, Facade, Strategy (auth providers)

## Context

Our team uses Bitbucket Cloud for source control and CI/CD. There is no equivalent to GitHub's `gh` CLI for Bitbucket. AI agents and developers need a command-line tool to create PRs, manage reviews, trigger pipelines, and perform other Bitbucket-specific operations that git alone cannot do.

We need a CLI tool (`bb`) that:
1. Acts as a transparent git proxy — unrecognized commands pass through to git with zero overhead.
2. Adds Bitbucket Cloud API features: PRs, pipelines, auth, configuration.
3. Provides a rich interactive TUI for complex workflows (PR creation, browsing).
4. Supports both interactive (OAuth) and headless (app password) authentication.

## Decision

Build `bb` in **Go** as a single static binary with the following stack:

- **Cobra** — CLI command routing and help generation.
- **Viper** — Git-style hierarchical configuration (global `~/.config/bb/config.toml` + repo `.bb/config.toml`, env vars with `BB_` prefix).
- **Bubble Tea** — Interactive TUI for PR lists, creation forms, pipeline status.
- **Lip Gloss + Bubbles** — Styled output (tables, spinners, text input).
- **net/http** — Bitbucket Cloud REST API v2.0 client.

### Architecture

```
cmd/bb/main.go          — Entrypoint, root Cobra command, git passthrough fallback
internal/
  config/               — Viper-based hierarchical config
  git/                  — Git passthrough (syscall.Exec) + remote URL parser
  auth/                 — Auth provider interface, OAuth 2.0, app password, token storage
  bitbucket/            — API client interface + HTTP implementation, PR service
  pr/                   — Cobra PR subcommands
  ui/                   — Bubble Tea components, Lip Gloss styles
```

### Key Design Decisions

1. **Git passthrough via `syscall.Exec`** (Unix) / `os/exec` (Windows). If `bb` doesn't recognize a subcommand, it replaces itself with `git`. This means `bb commit`, `bb push`, etc. work identically to their git equivalents with no overhead.

2. **Auth provider interface (Strategy pattern)**. Commands depend on `auth.Provider` interface, never on concrete OAuth/app-password implementations. Two implementations:
   - `OAuthProvider` — Browser-based authorization code grant with localhost callback, token refresh.
   - `AppPasswordProvider` — HTTP Basic auth with static credentials.

3. **Bitbucket API client interface (Adapter pattern)**. All API calls go through `bitbucket.Client` interface, enabling test doubles without hitting the real API.

4. **TUI/plain-text auto-detection**. When stdout is a TTY, use Bubble Tea. When piped, output plain text/JSON. Commands support `--format json|table|plain`.

5. **Token storage**: `~/.config/bb/tokens.json` with 0600 permissions. Keychain integration deferred — can be added later behind the `auth.Provider` interface.

### Error Taxonomy

- `domain` — PR not found, merge conflict, not authorized
- `infra` — network timeout, token file unreadable
- `validation` — missing required flag, invalid PR ID

## Alternatives Considered

### Python CLI (Click/Typer)

Rejected. Go produces a single static binary with fast startup — critical for a git proxy where every command adds latency. Python requires a runtime, virtual environment, and has slower startup.

### Shell scripts wrapping curl

Rejected. Not maintainable, no interactive TUI, poor error handling, difficult to test.

### Using existing tools (e.g., bitbucket-cli npm package)

Rejected. Requires Node.js runtime, limited features, not actively maintained, no TUI.

## Consequences

- **Go expertise required** for maintenance — different from our Python-centric stack.
- **OAuth consumer setup required** — need to register a Bitbucket OAuth consumer for the client_id/client_secret.
- **Cross-platform build matrix** — need to compile for Linux, macOS, Windows (Go makes this trivial).
- **Tool is immediately useful** — even before Bitbucket features, the git passthrough makes `bb` a drop-in `git` replacement.

## Rollback

This is a standalone CLI tool, not infrastructure. Rollback is simply "stop using `bb` and use `git` directly". No infrastructure changes required.
