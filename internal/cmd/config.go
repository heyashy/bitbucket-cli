package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/heyashy/bitbucket-cli/internal/config"
	"github.com/heyashy/bitbucket-cli/internal/ui"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage bb configuration",
		Long: `Read and write bb configuration values.

Configuration is hierarchical, loaded in order (later overrides earlier):
  1. Global:  ~/.config/bb/config.toml
  2. Repo:    .bb/config.toml  (in your repository root)
  3. Env:     BB_* environment variables (e.g. BB_WORKSPACE)

AVAILABLE KEYS
  workspace                  Bitbucket workspace slug (auto-detected from git remote)
  repo_slug                  Repository slug (auto-detected from git remote)
  auth.method                Authentication method: apitoken, oauth, apppassword
  auth.oauth.client_id       OAuth consumer client ID
  auth.oauth.client_secret   OAuth consumer client secret
  pr.merge_strategy          Default merge strategy: merge_commit, squash, fast_forward
  pr.default_reviewers       Default reviewer UUIDs (comma-separated)

EXAMPLES
  bb config get workspace
  bb config set --global auth.method apitoken
  bb config set workspace my-workspace
  bb config list`,
	}

	cmd.AddCommand(newConfigGetCmd())
	cmd.AddCommand(newConfigSetCmd())
	cmd.AddCommand(newConfigListCmd())

	return cmd
}

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a config value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			val := cfg.GetString(args[0])
			if val == "" {
				return fmt.Errorf("key not found: %s", args[0])
			}

			fmt.Println(val)
			return nil
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	var global bool

	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a config value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			cfg.Set(args[0], args[1])

			if global {
				if err := cfg.WriteGlobal(); err != nil {
					return err
				}
				fmt.Printf("  %s Set %s = %s (global)\n", ui.CheckMark(), ui.Key.Render(args[0]), ui.Value.Render(args[1]))
				return nil
			}

			// TODO: write to repo-local config
			if err := cfg.WriteGlobal(); err != nil {
				return err
			}
			fmt.Printf("  %s Set %s = %s\n", ui.CheckMark(), ui.Key.Render(args[0]), ui.Value.Render(args[1]))
			return nil
		},
	}

	cmd.Flags().BoolVar(&global, "global", false, "Write to global config (~/.config/bb/config.toml)")

	return cmd
}

func newConfigListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all config values",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			settings := cfg.AllSettings()
			if len(settings) == 0 {
				fmt.Println(ui.Faint.Render("  No configuration values set."))
				return nil
			}

			fmt.Println()
			fmt.Println(ui.Title.Render("Configuration"))
			fmt.Println(ui.Divider.String())

			flat := flattenMap("", settings)
			keys := make([]string, 0, len(flat))
			for k := range flat {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				fmt.Printf("  %s %s\n",
					ui.Key.Render(k+":"),
					ui.Value.Render(fmt.Sprintf("%v", flat[k])),
				)
			}
			fmt.Println()
			return nil
		},
	}
}

func flattenMap(prefix string, m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		if sub, ok := v.(map[string]interface{}); ok {
			for sk, sv := range flattenMap(key, sub) {
				result[sk] = sv
			}
		} else {
			result[key] = v
		}
	}
	return result
}
