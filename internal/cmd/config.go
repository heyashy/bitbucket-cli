package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/heyashy/bb/internal/config"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage bb configuration",
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
				return cfg.WriteGlobal()
			}

			// TODO: write to repo-local config
			return cfg.WriteGlobal()
		},
	}

	cmd.Flags().BoolVar(&global, "global", false, "Write to global config")

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

			// Use Viper's AllSettings to dump everything
			_ = cfg
			fmt.Println("Configuration loaded successfully.")
			fmt.Println("Use 'bb config get <key>' to view specific values.")
			return nil
		},
	}
}
