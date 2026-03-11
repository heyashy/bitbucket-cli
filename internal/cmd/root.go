package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/heyashy/bitbucket-cli/internal/git"
)

var knownCommands = map[string]bool{
	"auth":   true,
	"pr":     true,
	"config": true,
	"help":   true,
}

func NewRootCmd(version string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "bb",
		Short:   "Bitbucket CLI — git proxy with Bitbucket superpowers",
		Version: version,
		// If no subcommand matches, pass through to git
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

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
