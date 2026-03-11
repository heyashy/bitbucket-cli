# AGENTS.md

> This file defines the engineering standards, conventions, and rules for AI agents and human
> contributors working in this repository. All contributors (human and AI) must follow these rules.

---

## Architecture

- A design pattern must be chosen before coding where applicable (state the pattern in the ADR).
- **ADR required for all non-trivial changes** (see [ADR Process](#adr-process) below).
  - Include: context, decision, alternatives, consequences.
  - Any infra change ADR must include an explicit **Rollback** section.
- Explicit non-functional requirements (baseline per service):
  - Latency, availability, cost constraints, scaling expectations, RTO/RPO, data retention.
- Dependency boundaries:
  - Keep domain logic isolated from infrastructure/framework code.
  - No AWS SDK calls in core domain modules; wrap behind interfaces/adapters.
- Event-driven rules (if applicable):
  - Idempotency keys required.
  - Deduplication strategy documented.
  - Ordering expectations documented.
  - Retry policy and DLQ policy documented.
  - Exactly-once vs at-least-once assumptions documented.

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
**Pattern:** (e.g. Adapter, Event-Driven, CQRS)

## Context
What is the problem or opportunity?

## Decision
What was decided?

## Alternatives Considered
What else was evaluated and why was it rejected?

## Consequences
What are the trade-offs, risks, and follow-up actions?

## Rollback
*(Required for all infrastructure changes)*  
How do we undo this if it goes wrong?
```

### MANIFEST.md Format

`docs/adr/MANIFEST.md` is a markdown table updated whenever an ADR is created or its status changes:

```markdown
# ADR Manifest

| ADR    | Title                        | Status  | Date       |
|--------|------------------------------|---------|------------|
| ADR-0001 | Chose SQLModel as ORM      | done    | 2024-01-10 |
| ADR-0042 | Migrate to event-driven    | pending | 2024-06-01 |
```

---

## Change Log

**Every day that includes changes to the repository must have a corresponding log file.** Logs live in `docs/log/` with one file per calendar day:

```
docs/
  log/
    YYYY-MM-DD.md   # One file per day — mandatory if any changes are made that day
```

- A log entry must be created (or appended to) at the **end of every working session** where code, infrastructure, or documentation was changed.
- If no log file exists for today, create one before finishing work.
- Log files are **append-only**. Do not edit historical log entries.

### Log File Format

```markdown
# Change Log — YYYY-MM-DD

## Summary
Brief description of what changed this day/sprint.

## Changes
- [ADR-XXXX] Short description of change
- [HOTFIX] Short description if no ADR required
- [CHORE] Dependency bumps, formatting, etc.

## ADRs Progressed
- ADR-XXXX moved from pending → done
```

---

## Coding

- TDD required:
  - Extract requirements into unit tests first.
  - Follow red → green → refactor.
- Readable code without comments; add comments only where the code is genuinely not self-explanatory.
- Code must be DRY and extensible.

### Python

- Ruff formatted (Python).
- Use `uv` for Python package management.
- Logging and monitoring throughout, with CloudWatch enabled.
  - Sentry can be added for larger or more critical applications.
- Preferred libraries:
  - `SQLModel` when an ORM is required.
  - `Pydantic` for validation — for Lambdas remain lightweight and use dataclasses.
  - `Tenacity` for retry logic.
  - `pytest` — all tests must be tagged: `unit`, `integration`, or `smoke`.
  - Prefer `httpx` over `requests`.
- Error handling standards:
  - No naked `except:`.
  - Errors must either be handled or re-raised with context.
  - Define a consistent error taxonomy: `domain` / `infra` / `validation`.
- Configuration rules:
  - No `os.getenv` inside core logic; config assembled at entrypoints only.
  - Fail fast on missing config.
- Use modern Python type hints throughout.

### Linting / Formatting (Python)

- `ruff` line length: `100`.
- `ruff` import sorting: enabled (`I`).
- `ruff` ruleset:
  - select: `E`, `F`, `I`, `B`, `UP`, `N`, `S`, `C90`, `SIM`, `PL`, `RUF`.
  - ignore: `S101` — **only for test files**, enforced via `pyproject.toml`:

```toml
[tool.ruff.lint.per-file-ignores]
"tests/**" = ["S101"]
```

> `S101` must NOT be suppressed globally. Assert usage outside `tests/` is a lint failure.

### Go (`bb` CLI)

- Module: `github.com/heyashy/bb`.
- Min Go version: 1.22+.
- CLI framework: **Cobra** for command routing.
- Configuration: **Viper** — git-style hierarchical config (global `~/.config/bb/config.toml` + repo `.bb/config.toml`, env vars prefixed `BB_`).
- TUI: **Bubble Tea** + **Lip Gloss** + **Bubbles** for interactive components.
- HTTP client: `net/http` (stdlib).
- Preferred libraries:
  - `github.com/spf13/cobra` — CLI commands.
  - `github.com/spf13/viper` — config management.
  - `github.com/charmbracelet/bubbletea` — interactive TUI.
  - `github.com/charmbracelet/lipgloss` — styled terminal output.
  - `github.com/charmbracelet/bubbles` — reusable TUI components.
  - `github.com/stretchr/testify` — test assertions.
  - `github.com/pkg/browser` — open URLs in the user's browser.
- Linting: `golangci-lint` with default rules + `govet`, `errcheck`, `staticcheck`.
- Formatting: `gofmt` / `goimports`.
- Error handling:
  - Errors must wrap context: `fmt.Errorf("description: %w", err)`.
  - Error taxonomy: `domain` / `infra` / `validation` (same as Python).
  - Never discard errors silently.
- Configuration rules:
  - No `os.Getenv` inside core logic; config assembled at entrypoints via Viper.
  - Fail fast on missing config.
- Testing:
  - Use `go test` with `testify` for assertions.
  - Interfaces for all external dependencies (API client, auth provider) to enable test doubles.
  - Use `httptest.NewServer` for HTTP integration tests.
  - Table-driven tests where applicable.
- Build:
  - `go build -o bb ./cmd/bb`
  - Version injected via `-ldflags "-X main.version=..."`.

---

## Testing

- Testing pyramid:
  - Unit tests required.
  - Contract tests where services integrate.
  - Integration tests for AWS interactions (LocalStack or real AWS sandbox).
  - Smoke tests post-deploy.
  - Unit tests must **never** touch live services.
- Test conventions:
  - Naming: `test_<behavior>`.
  - Use AAA (Arrange / Act / Assert) or Given/When/Then.
  - Avoid mocking implementation details; mock boundaries only.

---

## CI/CD

- Bitbucket is the CI/CD tool.
- Packaging of Lambdas and Docker image builds must be triggered by **Terraform**, not the Bitbucket pipeline.

### Branches → Environments

| Branch | Environment | Stages              |
|--------|-------------|---------------------|
| `qa`   | QA / dev    | `qa-plan`, `qa-apply` |
| `main` | Live / prod | `live-plan`, `live-apply` |

- Prior to `qa-plan`: run unit tests, `ruff`, and `terraform fmt`.

### Bitbucket Project Variables

| Variable                  | Environment |
|---------------------------|-------------|
| `AWS_ACCESS_KEY_ID_QA`          | QA          |
| `AWS_SECRET_ACCESS_KEY_QA`      | QA          |
| `DATA_STATE_BUCKET_QA`          | QA          |
| `DATA_ENV_QA`                   | QA          |
| `AWS_TERRAFORM_ACCESS_KEY_LIVE` | Live        |
| `AWS_TERRAFORM_SECRET_KEY_LIVE` | Live        |
| `DATA_STATE_BUCKET_LIVE`        | Live        |
| `DATA_ENV_LIVE`           | Live        |

### Quality Gates (must pass before apply)

- Unit tests.
- `ruff`.
- Type check (if adopted).
- `terraform fmt` and `terraform validate`.

### Plan / Apply Safety

- Manual approval gates for all `*-apply` steps, especially live.
- Require a clean plan before apply — no drift surprises.
- Store plan artifacts and link them to PRs/builds.

### Versioning / Release

- Artifacts versioned as `semver + git sha`.
- Live deployments must reference a tag.

### Rollback Strategy

- **Lambda:** roll back via previous version/alias.
- **ECS/ECR:** roll back via previous image digest.
- **Terraform:** state rollback is hard — mitigation must be documented in the ADR's Rollback section for every infra change.

---

## Infra / Terraform

- Artifact naming: `{stack-name}-{env}-{hash}`.
  - Hash must be content-based; avoid time-based hashes.
  - Add timestamp/build number only if content-hash collisions are possible.
  - Several AWS resources have a 64-character name limit (IAM roles, Lambda functions, EventBridge rules). Always verify names fit under 64 chars using the **live** environment (4 chars), not qa (2 chars).
- Builds triggered by Terraform, not the pipeline.
  - Build scripts must be POSIX compliant (`#!/bin/sh`), not bash.
  - **Exception:** `terraform/terraform_qa.sh` is a developer convenience script, not a build artifact, and is exempt from the POSIX build rule. It may use bash if needed.
- Always-trigger via timestamp is allowed only when reproducibility isn't possible.

### State & Locking

- S3 backend only — **no DynamoDB lock table** and **no `encrypt` flag**.
- Backend config requires only `bucket`, `key`, and `region`:

```hcl
backend "s3" {
  key    = "{stack-name}/terraform.tfstate"
  region = "eu-west-1"
}
```

- The `bucket` is supplied at init time via `-backend-config="bucket=..."`.
- Use `terraform workspace select -or-create {env}` — never `workspace select || workspace new`.

### Secrets

- Pull from AWS Secrets Manager.
- Secret names come from `tfvars`.
- `tfvars` are the source of truth for app variables.
- Never output secrets in Terraform outputs.
- Avoid secrets in state where possible; call out unavoidable cases in the ADR.

### tfvars

Both `live.tfvars` and `qa.tfvars` must exist and contain infra truths:

```hcl
vpc_id = "vpc-007d0afc723dbf977"
private_subnet_ids = [
  "subnet-07639910743c7e5f6",
  "subnet-070ce1c50cc106ee0",
  "subnet-01f44334156e92379"
]
```

### Tags

All resources must be tagged via `locals` in `main.tf`:

```hcl
locals {
  project_name = "CHANGE_ME"
  stack_name   = replace(lower(local.project_name), " ", "-")
  environment  = lower(terraform.workspace)
  random_id    = random_id.stack.hex
  region       = "eu-west-1"

  tags = {
    Name        = "CHANGE_ME"
    Project     = "CHANGE_ME"
    Service     = "Data"
    Environment = upper(terraform.workspace)
    ManagedBy   = "Terraform"
    Repo        = "CHANGE_ME"  # Populate with the upstream origin URL
  }
}
```

> `Name`, `Project`, and `Repo` must be updated per project before first deploy. Leaving them as
> `CHANGE_ME` is a hard review failure.

### Module Standards

- Inputs and outputs documented.
- No mega-modules — group logically and confirm structure with the team.
- Modules versioned (tagged) or pinned to a commit.

### EventBridge Schedulers

- Schedulers must be **disabled in QA** and **enabled in Live** by default.
- Use a `enable_scheduler` variable (boolean, default `false`) to control the EventBridge rule state.
- QA jobs should be triggered manually; only Live runs on a schedule.

### CloudWatch

- Log retention must be set explicitly per environment.
- No "never expire" defaults.

### IAM

- Least privilege required.
- Wildcard permissions must be justified in a code review or ADR.

### Drift Detection

- Scheduled plan-only runs (no apply) to detect drift between environments.
- Drift must be resolved before any apply is permitted.

### QA Deployment Script

`terraform/terraform_qa.sh` must exist as a **local developer convenience script** for deploying to QA from a workstation. It is not used in CI — Bitbucket Pipelines uses its own project variables.

Responsibilities:

- Pull credentials from the local AWS CLI `qa` profile using `aws configure get`:

```bash
export AWS_ACCESS_KEY_ID=$(aws configure get aws_access_key_id --profile qa)
export AWS_SECRET_ACCESS_KEY=$(aws configure get aws_secret_access_key --profile qa)
export AWS_REGION=$(aws configure get region --profile qa)
```

- Run `terraform init -backend-config="bucket=..." -reconfigure`.
- Run `terraform workspace select -or-create qa`.
- Run `terraform apply -var-file=vars/qa.tfvars` (shows plan and prompts for confirmation).

> **No auto-apply.** The script must require manual confirmation before applying. Do not use `-auto-approve`, and do not use saved plan files (`-out` / `terraform apply plan.tfplan`) as these skip the interactive prompt.

> Do **not** use `aws sts get-caller-identity` or similar commands that print credentials or sensitive output to the terminal.

---

## Observability

- Structured logging (JSON).
- Required log fields: `service`, `env`, `version`, `request_id` / `correlation_id`, `tenant` (if relevant).
- Redaction rules must be applied — no secrets or PII in logs.

### Metrics Minimums

- Invocation count.
- Duration (p50 / p95 / p99).
- Error count and error rate.
- Throttles.
- DLQ depth (if event-driven).

### Tracing

- Use X-Ray.
- Propagate correlation IDs across async boundaries.

### Alarms (per service)

- Error rate spike.
- Latency regression.
- DLQ depth > threshold.
- No events processed for X minutes (dead consumer).

### Notification Routing

Every service must define its notification routing explicitly. Minimum standard:

- One SNS topic per environment: `{stack-name}-{env}-alerts`.
- SNS topic routes to a Slack channel following the convention: `#alerts-{stack}-{env}`.
- Example: `#alerts-my-service-live`, `#alerts-my-service-qa`.
- On-call escalation path must be documented in `RUNBOOK.md`.

### Log Retention

| Environment | Retention |
|-------------|-----------|
| QA          | 7 days    |
| Live        | 90 days   |

---

## Security & Compliance

- Encryption at rest and in transit required by default.
- Public access must be explicitly justified and documented.
- Dependency management:
  - Vulnerability scanning via `pip-audit`, Dependabot, or Snyk.
  - Pin dependencies with hashes where supported.
- Secrets hygiene:
  - No secrets in the repo, CI vars, or logs.
  - Redaction rules applied to structured logs.
- Threat modeling required for new externally-facing endpoints or sensitive data flows.

---

## Repo & Developer Experience

### Project Structure

```
src/
  lib/          # Shared modules / domain logic
tests/
  unit/
  integration/
  smoke/
terraform/
  modules/
  terraform_qa.sh
scripts/
docs/
  adr/
    pending/
    done/
    MANIFEST.md
  log/
```

One obvious entrypoint per Lambda/service.

### Makefile

A `Makefile` is required with the following targets:

| Target               | Description                                          |
|----------------------|------------------------------------------------------|
| `make run`           | Run the application locally                          |
| `make local`         | Spin up the full local stack (LocalStack/docker-compose) |
| `make test`          | Run unit tests                                       |
| `make integration-test` | Run integration tests                             |
| `make smoke`         | Run smoke tests                                      |
| `make lint`          | Run `ruff`                                           |
| `make terraform-lint`| Run `terraform fmt`                                  |
| `make terraform-validate` | Run `terraform validate`                       |

> `make run` starts the app process only. `make local` brings up the full local infrastructure
> stack (LocalStack, docker-compose, etc.) needed to support it.

### Local Dev

- Document how to run locally in `README.md` (docker-compose / SAM / LocalStack).
- Provide `.env.example` — no secrets, no real values.

### Documentation

- `RUNBOOK.md` per service must include:
  - Alarm descriptions and thresholds.
  - Dashboard links.
  - Common failure modes and remediation.
  - Rollback steps.
  - Redeploy steps.
  - On-call escalation path.

---

## Definition of Done

A change is not done until all of the following are true:

- [ ] Tests added/updated (unit + integration where applicable).
- [ ] Logging and metrics added.
- [ ] Alarms configured (live environment).
- [ ] `RUNBOOK.md` updated.
- [ ] ADR written and added to `docs/adr/pending/` (if non-trivial).
- [ ] `docs/adr/MANIFEST.md` updated.
- [ ] `docs/log/YYYY-MM-DD.md` entry added.
- [ ] Terraform plan reviewed — no unexpected drift.
- [ ] Smoke tests passing in QA.

---

## Code Review Rules

- 1–2 approvals required depending on risk level.
- Reviewer checklist:
  - [ ] Security and IAM least privilege.
  - [ ] Logging and metrics present.
  - [ ] Tests cover the behaviour.
  - [ ] Backward compatibility maintained.
  - [ ] No `CHANGE_ME` tags left in Terraform.
  - [ ] ADR present if non-trivial.
  - [ ] No secrets in code, logs, or outputs.

---

## Backward Compatibility

- Migration strategy required for data/schema changes.
- Event and API versioning rules must be documented — don't break consumers.
