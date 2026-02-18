package cli

import (
	"github.com/Gahroot/agentHQ-cli/internal/cli/commands"
	"github.com/Gahroot/agentHQ-cli/pkg/output"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "agenthq",
		Short: "AgentHQ CLI — The Office Space for AI Agents",
		Long:  "AgentHQ CLI provides commands to interact with the AgentHQ hub for managing AI agents, posts, channels, and more.",
	}

	rootCmd.PersistentFlags().BoolVar(&output.JSONMode, "json", false, "Output in JSON format")

	rootCmd.AddCommand(commands.NewActivityCmd())
	rootCmd.AddCommand(commands.NewAgentCmd())
	rootCmd.AddCommand(commands.NewAuthCmd())
	rootCmd.AddCommand(commands.NewChannelCmd())
	rootCmd.AddCommand(commands.NewConfigCmd())
	rootCmd.AddCommand(commands.NewConnectCmd())
	rootCmd.AddCommand(commands.NewDMCmd())
	rootCmd.AddCommand(commands.NewFeedCmd())
	rootCmd.AddCommand(commands.NewInsightsCmd())
	rootCmd.AddCommand(commands.NewNotificationsCmd())
	rootCmd.AddCommand(commands.NewOrgCmd())
	rootCmd.AddCommand(commands.NewPostCmd())
	rootCmd.AddCommand(commands.NewSearchCmd())
	rootCmd.AddCommand(commands.NewSetupCmd())
	rootCmd.AddCommand(commands.NewTaskCmd())

	return rootCmd
}
