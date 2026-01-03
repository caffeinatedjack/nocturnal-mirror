package cmd

import (
	"github.com/spf13/cobra"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Agent commands for proposals and documentation",
}

func init() {
	agentCmd.Long = helpText("agent")

	rootCmd.AddCommand(agentCmd)
}
