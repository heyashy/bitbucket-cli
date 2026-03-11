# bb

A drop-in replacement for `git` that adds Bitbucket Cloud superpowers. Use it exactly like `git` — any command `bb` doesn't recognise gets passed straight through. When you need Bitbucket features like pull requests, authentication, or pipelines, `bb` handles them natively.

```
bb commit -m "fix the thing"     # just git
bb push                          # just git
bb pr create -t "Fix the thing"  # bitbucket api
bb pr list                       # bitbucket api
bb pr merge 42                   # bitbucket api
```

## Install

### Homebrew (macOS / Linux)

```sh
brew install heyashy/tap/bb
```

### Go install

```sh
go install github.com/heyashy/bb/cmd/bb@latest
```

### Download binary

Grab the latest release for your platform from [Releases](https://github.com/heyashy/bitbucket-cli/releases) and put it on your `PATH`.

### Build from source

```sh
git clone git@github.com:heyashy/bitbucket-cli.git
cd bitbucket-cli
go build -o bb ./cmd/bb
sudo mv bb /usr/local/bin/
```

## Quick start

### 1. Authenticate

**OAuth (recommended)** — opens your browser:

```sh
# First, configure your OAuth consumer credentials (one-time)
bb config set --global auth.oauth.client_id YOUR_CLIENT_ID
bb config set --global auth.oauth.client_secret YOUR_CLIENT_SECRET

# Then login
bb auth login
```

**App password** — for CI or headless environments:

```sh
bb auth login --app-password
# Prompts for username and app password
```

> Create an app password at: Bitbucket > Personal settings > App passwords.
> Required scopes: `pullrequest`, `pullrequest:write`, `repository`, `repository:write`.

### 2. Use it

```sh
# Everything you do with git, just use bb instead
bb clone git@bitbucket.org:workspace/repo.git
bb checkout -b feature/cool-thing
bb add .
bb commit -m "add cool thing"
bb push -u origin feature/cool-thing

# Now the Bitbucket stuff
bb pr create --title "Add cool thing" --body "Does the cool thing"
bb pr list
bb pr view 1
bb pr approve 1
bb pr merge 1
```

## Commands

### Git passthrough

Any command that isn't a `bb` subcommand passes through to `git`:

```sh
bb status          # git status
bb log --oneline   # git log --oneline
bb stash pop       # git stash pop
bb rebase -i HEAD~3  # git rebase -i HEAD~3
```

### Auth

| Command | Description |
|---------|-------------|
| `bb auth login` | Authenticate via OAuth (browser flow) |
| `bb auth login --app-password` | Authenticate with an app password |
| `bb auth logout` | Clear stored credentials |
| `bb auth status` | Show current auth state |

### Pull requests

| Command | Description |
|---------|-------------|
| `bb pr create` | Create a PR from the current branch |
| `bb pr list` | List open pull requests |
| `bb pr view <id>` | View PR details |
| `bb pr merge <id>` | Merge a PR |
| `bb pr approve <id>` | Approve a PR |
| `bb pr decline <id>` | Decline a PR |
| `bb pr comment <id>` | Comment on a PR |
| `bb pr diff <id>` | Show PR diff |

#### PR create flags

```
-t, --title        PR title (required)
-b, --body         PR description
-s, --source       Source branch (default: current branch)
-d, --dest         Destination branch (default: main)
-r, --reviewer     Reviewer UUIDs (repeatable)
    --close-branch Close source branch after merge
```

#### PR list flags

```
--state   Filter: OPEN, MERGED, DECLINED, SUPERSEDED (default: OPEN)
```

#### PR merge flags

```
--strategy      merge_commit, squash, fast_forward
--close-branch  Close source branch after merge
-m, --message   Merge commit message
```

### Config

Git-style hierarchical configuration. Global config lives at `~/.config/bb/config.toml`, repo-level config at `.bb/config.toml`. Environment variables with `BB_` prefix override both.

| Command | Description |
|---------|-------------|
| `bb config get <key>` | Get a config value |
| `bb config set <key> <value>` | Set a config value |
| `bb config set --global <key> <value>` | Set a global config value |
| `bb config list` | List config |

#### Config keys

| Key | Description | Default |
|-----|-------------|---------|
| `workspace` | Bitbucket workspace (auto-detected from remote) | — |
| `repo_slug` | Repository slug (auto-detected from remote) | — |
| `auth.method` | `oauth` or `apppassword` | `oauth` |
| `auth.oauth.client_id` | OAuth consumer key | — |
| `auth.oauth.client_secret` | OAuth consumer secret | — |
| `pr.merge_strategy` | Default merge strategy | `merge_commit` |
| `pr.default_reviewers` | Default reviewer UUIDs | — |

## How it works

`bb` checks if the first argument is a known subcommand (`auth`, `pr`, `config`). If not, it calls `git` via `syscall.Exec`, replacing the process entirely — same stdin, stdout, stderr, signals, and exit code. There is zero overhead for git operations.

Bitbucket features use the [Bitbucket Cloud REST API v2.0](https://developer.atlassian.com/cloud/bitbucket/rest/intro/). The workspace and repository are auto-detected from your git remote URL, so there's nothing to configure for most repos.

## Setting up OAuth

To use `bb auth login` (browser-based OAuth), you need a Bitbucket OAuth consumer:

1. Go to **Bitbucket** > **Workspace settings** > **OAuth consumers** > **Add consumer**
2. Set the callback URL to `http://127.0.0.1` (any port — `bb` picks a random one)
3. Grant permissions: Account (read), Repositories (read/write), Pull requests (read/write)
4. Copy the **Key** (client ID) and **Secret** (client secret)
5. Configure `bb`:

```sh
bb config set --global auth.oauth.client_id YOUR_KEY
bb config set --global auth.oauth.client_secret YOUR_SECRET
```

## Releasing

Releases are automated with [GoReleaser](https://goreleaser.com/) and GitHub Actions.

To create a new release:

```sh
# Tag the release
git tag -a v0.1.0 -m "Initial release"

# Push the tag — GitHub Actions takes it from here
git push origin v0.1.0
```

This will:
1. Run tests
2. Cross-compile for Linux, macOS, and Windows (amd64 + arm64)
3. Create a GitHub Release with binaries and checksums
4. Update the Homebrew tap

Use [semver](https://semver.org/): `v0.x.y` for pre-1.0, `vMAJOR.MINOR.PATCH` after.

## Project structure

```
cmd/bb/              Main entrypoint
internal/
  auth/              OAuth 2.0, app passwords, token storage
  bitbucket/         API client, types, PR service
  cmd/               Cobra command definitions
    resolve/         Dependency wiring (auth + API client + repo detection)
  config/            Viper-based hierarchical config
  git/               Git passthrough + remote URL parser
  ui/                Bubble Tea components (coming soon)
docs/
  adr/               Architecture decision records
  log/               Change log
```

## License

[MIT](LICENSE)
