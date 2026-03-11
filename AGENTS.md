# AGENTS.md

> Engineering standards for the `bb` CLI — a Go-based git proxy with Bitbucket Cloud superpowers.
> All contributors (human and AI) must follow these rules.

---

## Architecture

- A design pattern must be chosen before coding where applicable (state the pattern in the ADR).
- **ADR required for all non-trivial changes** (see [ADR Process](#adr-process) below).
  - Include: context, decision, alternatives, consequences.
- Dependency boundaries:
  - Keep domain logic isolated from infrastructure/framework code.
  - Bitbucket API calls wrapped behind interfaces (never called directly from commands).

---

## ADR Process

ADRs live in `docs/adr/` and follow a strict lifecycle:

```
docs/
  adr/
    pending/        # In-flight work — ADR proposed, change not yet implemented
    done/           # Implemented — ADR accepted and change shipped
    MANIFEST.md     # Master index of all ADRs (see format below)
```

### Lifecycle

| State     | Meaning                                              |
|-----------|------------------------------------------------------|
| `pending` | Change is in-flight; ADR written but not yet shipped |
| `done`    | Change implemented and ADR finalised                 |

> ADRs are never deleted. Superseded ADRs are moved to `done/` with a note referencing the new ADR.

### Naming

```
docs/adr/pending/ADR-0042-short-title.md
docs/adr/done/ADR-0042-short-title.md
```

### ADR Template

```markdown
# ADR-XXXX: Title

**Date:** YYYY-MM-DD
**Status:** pending | done
**Pattern:** (e.g. Adapter, Strategy, Facade)

## Context
What is the problem or opportunity?

## Decision
What was decided?

## Alternatives Considered
What else was evaluated and why was it rejected?

## Consequences
What are the trade-offs, risks, and follow-up actions?
```

### MANIFEST.md Format

```markdown
# ADR Manifest

| ADR    | Title                        | Status  | Date       |
|--------|------------------------------|---------|------------|
| ADR-0001 | bb CLI Architecture        | pending | 2026-03-11 |
```

---

## Change Log

**Every day that includes changes to the repository must have a corresponding log file.** Logs live in `docs/log/` with one file per calendar day:

```
docs/
  log/
    YYYY-MM-DD.md   # One file per day — mandatory if any changes are made that day
```

- A log entry must be created (or appended to) at the **end of every working session** where code or documentation was changed.
- If no log file exists for today, create one before finishing work.
- Log files are **append-only**. Do not edit historical log entries.

### Log File Format

```markdown
# Change Log — YYYY-MM-DD

## Summary
Brief description of what changed.

## Changes
- [ADR-XXXX] Short description of change
- [HOTFIX] Short description if no ADR required
- [CHORE] Dependency bumps, formatting, etc.

## ADRs Progressed
- ADR-XXXX moved from pending → done
```

---

## Go Coding Standards

- Module: `github.com/heyashy/bitbucket-cli`.
- Min Go version: 1.22+.
- TDD required:
  - Extract requirements into unit tests first.
  - Follow red → green → refactor.
- Readable code without comments; add comments only where the code is genuinely not self-explanatory.
- Code must be DRY and extensible.

### Stack

- **Cobra** — CLI command routing and git passthrough.
- **Viper** — git-style hierarchical config (global `~/.config/bb/config.toml` + repo `.bb/config.toml`, env vars prefixed `BB_`).
- **Bubble Tea** + **Lip Gloss** + **Bubbles** — interactive TUI components.
- **net/http** (stdlib) — HTTP client for Bitbucket API.
- **testify** — test assertions.
- **pkg/browser** — open URLs in the user's browser.

### Linting / Formatting

- `golangci-lint` with default rules + `govet`, `errcheck`, `staticcheck`.
- `gofmt` / `goimports` for formatting.

### Error Handling

- Errors must wrap context: `fmt.Errorf("description: %w", err)`.
- Error taxonomy: `domain` / `infra` / `validation`.
- Never discard errors silently.

### Configuration

- No `os.Getenv` inside core logic; config assembled at entrypoints via Viper.
- Fail fast on missing config.

### Testing

- Use `go test` with `testify` for assertions.
- Interfaces for all external dependencies (API client, auth provider) to enable test doubles.
- Use `httptest.NewServer` for HTTP integration tests.
- Table-driven tests where applicable.
- Naming: `Test<Behaviour>`.

### Build

- `go build -o bb ./cmd/bb`
- Version injected via `-ldflags "-X main.version=..."`.

---

## CI/CD

GitHub Actions with two workflows:

### CI (`.github/workflows/ci.yml`)

Runs on every push to `main` and on pull requests:
- `go build`
- `go test ./...`
- `go vet ./...`

### Release (`.github/workflows/release.yml`)

Triggered by pushing a semver tag (`v*`):
- Cross-compiles via GoReleaser for Linux, macOS, Windows (amd64 + arm64).
- Creates a GitHub Release with binaries and checksums.
- Optionally publishes to a Homebrew tap.

### Versioning

- Semver: `v0.x.y` for pre-1.0, `vMAJOR.MINOR.PATCH` after.
- To release: `git tag -a v0.x.y -m "description" && git push origin v0.x.y`.

---

## Project Structure

```
cmd/bb/              Main entrypoint
internal/
  auth/              OAuth 2.0, app passwords, token storage
  bitbucket/         API client interface, types, PR service
  cmd/               Cobra command definitions
    resolve/         Dependency wiring (auth + API client + repo detection)
  config/            Viper-based hierarchical config
  git/               Git passthrough + remote URL parser
  ui/                Bubble Tea components
docs/
  adr/               Architecture decision records
  log/               Daily change logs
```

---

## Makefile

| Target       | Description              |
|--------------|--------------------------|
| `make build` | Build the `bb` binary    |
| `make run`   | Build and run            |
| `make test`  | Run all tests            |
| `make lint`  | Run `golangci-lint`      |
| `make clean` | Remove build artifacts   |

---

## Definition of Done

A change is not done until:

- [ ] Tests added/updated.
- [ ] ADR written (if non-trivial) and added to `docs/adr/pending/`.
- [ ] `docs/adr/MANIFEST.md` updated.
- [ ] `docs/log/YYYY-MM-DD.md` entry added.
- [ ] `go vet` and `go test` pass.
- [ ] No secrets in code or logs.

---

## Code Review Rules

- 1–2 approvals required depending on risk level.
- Reviewer checklist:
  - [ ] Tests cover the behaviour.
  - [ ] Error handling follows taxonomy.
  - [ ] Interfaces used for external dependencies.
  - [ ] ADR present if non-trivial.
  - [ ] No secrets in code or logs.
  - [ ] Backward compatibility maintained.
