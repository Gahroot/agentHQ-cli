package commands

import (
	"encoding/json"
	"fmt"

	"github.com/agenthq/cli/internal/common/client"
	"github.com/agenthq/cli/pkg/output"
	"github.com/spf13/cobra"
)

func NewAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Agent management commands",
	}

	cmd.AddCommand(newAgentListCmd())
	cmd.AddCommand(newAgentStatusCmd())

	return cmd
}

func newAgentListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List agents in organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}
			resp, err := c.Get("/api/v1/agents", nil)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to list agents: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(json.RawMessage(resp.Data))
				return nil
			}

			var agents []struct {
				ID     string `json:"id"`
				Name   string `json:"name"`
				Status string `json:"status"`
			}
			json.Unmarshal(resp.Data, &agents)

			rows := make([][]string, len(agents))
			for i, a := range agents {
				rows[i] = []string{a.ID, a.Name, a.Status}
			}
			output.PrintTable([]string{"ID", "NAME", "STATUS"}, rows)
			return nil
		},
	}
}

func newAgentStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show agent online/offline status",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}
			resp, err := c.Get("/api/v1/agents", nil)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to get agent status: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(json.RawMessage(resp.Data))
				return nil
			}

			var agents []struct {
				Name          string `json:"name"`
				Status        string `json:"status"`
				LastHeartbeat string `json:"last_heartbeat"`
			}
			json.Unmarshal(resp.Data, &agents)

			rows := make([][]string, len(agents))
			for i, a := range agents {
				hb := a.LastHeartbeat
				if hb == "" {
					hb = "never"
				}
				rows[i] = []string{a.Name, a.Status, hb}
			}
			output.PrintTable([]string{"NAME", "STATUS", "LAST HEARTBEAT"}, rows)
			return nil
		},
	}
}
