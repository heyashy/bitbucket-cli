package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/heyashy/bitbucket-cli/internal/git"
)

var knownCommands = map[string]bool{
	"api":    true,
	"auth":   true,
	"pr":     true,
	"config": true,
	"help":   true,
}

func NewRootCmd(version string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "bb",
		Short: "Bitbucket CLI — a drop-in git replacement with Bitbucket Cloud superpowers",
		Long: `bb is a drop-in replacement for git that adds Bitbucket Cloud features.

Any command bb does not recognise is passed directly to git, so you can use
bb everywhere you currently use git — commit, push, pull, rebase, etc.

Bitbucket-specific commands (pull requests, auth, config) are built in.

GIT PASSTHROUGH
  bb commit -m "fix"      Runs: git commit -m "fix"
  bb push                 Runs: git push
  bb log --oneline        Runs: git log --oneline
  Any unrecognised command is forwarded to git with zero overhead.

AUTHENTICATION
  bb auth login            Log in with a Bitbucket API token (interactive)
  bb auth login --oauth    Log in via OAuth browser flow
  bb auth logout           Clear stored credentials
  bb auth status           Verify credentials against the Bitbucket API

  API tokens are created at: https://id.atlassian.com/manage-profile/security/api-tokens
  Select "Bitbucket" as the product and grant the scopes you need.

PULL REQUESTS
  bb pr create             Create a PR from the current branch
  bb pr list               List open PRs in this repository
  bb pr view [id]          View PR details (defaults to current branch PR)
  bb pr merge [id]         Merge a PR
  bb pr approve [id]       Approve a PR
  bb pr decline [id]       Decline a PR
  bb pr comment [id]       Comment on a PR
  bb pr diff [id]          Show the PR diff

  When [id] is omitted, bb finds the open PR for your current git branch.

RAW API ACCESS
  bb api <endpoint>        Make authenticated requests to any Bitbucket API endpoint
  bb api /user             Get current user
  bb api /repositories/{workspace}/{repo}/pipelines
  bb api -X POST /repositories/{workspace}/{repo}/pullrequests/42/approve

  {workspace} and {repo} are auto-replaced from your git remote.
  See 'bb api --help' for the full list of common endpoints.

CONFIGURATION
  bb config get <key>      Read a config value
  bb config set <key> <v>  Write a config value (use --global for global)
  bb config list           Show loaded config

  Config is hierarchical (like git):
    Global:  ~/.config/bb/config.toml
    Repo:    .bb/config.toml
    Env:     BB_* environment variables override both

AUTO-DETECTION
  bb reads your git remote URL to detect the Bitbucket workspace and repository.
  No manual config needed for most repos. Override with:
    bb config set workspace <name>
    bb config set repo_slug <name>

EXAMPLES
  bb push -u origin feature/my-branch
  bb pr create --title "Add feature" --dest main
  bb pr list --state MERGED
  bb pr view
  bb pr merge --strategy squash --close-branch`,
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.AddCommand(newAPICmd())
	rootCmd.AddCommand(newAuthCmd())
	rootCmd.AddCommand(newPRCmd())
	rootCmd.AddCommand(newConfigCmd())

	return rootCmd
}

func init() {
	// If the first arg is not a known bb command, passthrough to git.
	// This runs before cobra parses anything.
	if len(os.Args) > 1 {
		firstArg := os.Args[1]

		// Let cobra handle flags like --help, --version
		if firstArg[0] == '-' {
			return
		}

		if !knownCommands[firstArg] {
			git.ExecGit(os.Args[1:])
			// ExecGit replaces the process on Unix, so we only reach here on error
			os.Exit(1)
		}
	}
}
