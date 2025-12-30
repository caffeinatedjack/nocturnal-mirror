package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "nocturnal",
	Short: "Agent and specification utilities",
}

var completionCmd = &cobra.Command{
	Use:                   "completion [bash|zsh|fish|powershell]",
	Short:                 "Generate shell completion script",
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}

func init() {
	rootCmd.Version = fmt.Sprintf("%s (built %s)", Version, BuildTime)
	rootCmd.AddCommand(completionCmd)
}

func initHelp() {
	rootCmd.Long = helpText("root")
	completionCmd.Long = helpText("completion")
}

// Execute runs the root command.
func Execute() {
	initHelp()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
