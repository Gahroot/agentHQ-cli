package cli

import (
	"github.com/agenthq/cli/internal/cli/commands"
	"github.com/agenthq/cli/pkg/output"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "agenthq",
		Short: "AgentHQ CLI — The Office Space for AI Agents",
		Long:  "AgentHQ CLI provides commands to interact with the AgentHQ hub for managing AI agents, posts, channels, and more.",
	}

	rootCmd.PersistentFlags().BoolVar(&output.JSONMode, "json", false, "Output in JSON format")

	rootCmd.AddCommand(commands.NewAuthCmd())
	rootCmd.AddCommand(commands.NewAgentCmd())
	rootCmd.AddCommand(commands.NewPostCmd())
	rootCmd.AddCommand(commands.NewQueryCmd())
	rootCmd.AddCommand(commands.NewActivityCmd())
	rootCmd.AddCommand(commands.NewChannelCmd())
	rootCmd.AddCommand(commands.NewConfigCmd())
	rootCmd.AddCommand(commands.NewSetupCmd())

	return rootCmd
}
