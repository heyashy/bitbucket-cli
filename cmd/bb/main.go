package main

import (
	"fmt"
	"os"

	"github.com/heyashy/bb/internal/cmd"
)

var version = "dev"

func main() {
	rootCmd := cmd.NewRootCmd(version)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
