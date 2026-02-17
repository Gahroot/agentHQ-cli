package commands

import (
	"encoding/json"
	"fmt"

	"github.com/Gahroot/agentHQ-cli/internal/common/client"
	"github.com/Gahroot/agentHQ-cli/internal/common/config"
	"github.com/Gahroot/agentHQ-cli/pkg/output"
	"github.com/spf13/cobra"
)

func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management",
	}

	cmd.AddCommand(newConfigSetCmd())
	cmd.AddCommand(newConfigGetCmd())

	return cmd
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a config value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to load config: %v", err))
				return nil
			}

			key, val := args[0], args[1]
			switch key {
			case "hub_url":
				cfg.HubURL = val
			case "api_key":
				cfg.APIKey = val
			case "org_id":
				cfg.OrgID = val
			case "agent_id":
				cfg.AgentID = val
			default:
				output.PrintError(fmt.Sprintf("Unknown config key: %s", key))
				return nil
			}

			if err := config.Save(cfg); err != nil {
				output.PrintError(fmt.Sprintf("Failed to save config: %v", err))
				return nil
			}

			output.PrintSuccess(fmt.Sprintf("Config %s set to %s", key, val))
			return nil
		},
	}
}

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to load config: %v", err))
				return nil
			}

			// Mask sensitive values
			display := map[string]string{
				"hub_url":  cfg.HubURL,
				"org_id":   cfg.OrgID,
				"agent_id": cfg.AgentID,
			}
			if cfg.APIKey != "" {
				display["api_key"] = cfg.APIKey[:12] + "..."
			}
			if cfg.JWTToken != "" {
				display["jwt_token"] = "***set***"
			}

			data, _ := json.MarshalIndent(display, "", "  ")
			fmt.Println(string(data))
			return nil
		},
	}
}

func NewSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup and connectivity commands",
	}
	cmd.AddCommand(newSetupTestCmd())
	return cmd
}

func newSetupTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Test hub connectivity",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to create client: %v", err))
				return nil
			}

			cfg, _ := config.Load()
			fmt.Printf("Testing connection to %s...\n", cfg.HubURL)

			_, err = c.Get("/health", nil)
			if err != nil {
				output.PrintError(fmt.Sprintf("Connection failed: %v", err))
				return nil
			}

			output.PrintSuccess("Hub is reachable")
			return nil
		},
	}
}
