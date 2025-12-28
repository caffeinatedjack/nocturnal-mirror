// Package cmd implements the CLI commands for nocturnal.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version is set at build time via ldflags
	Version   = "dev"
	BuildTime = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "nocturnal",
	Short: "Agent and docs utilities",
	Long: `Nocturnal - Agent and docs utilities.

Commands:
    agent   Read/write TODO.md in the current directory
    docs    Search/list components in ~/.docs

Examples:
    nocturnal agent todowrite < todos.json
    nocturnal agent todoread
    nocturnal docs list
    nocturnal docs search "component"`,
}

func init() {
	rootCmd.Version = fmt.Sprintf("%s (built %s)", Version, BuildTime)
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
