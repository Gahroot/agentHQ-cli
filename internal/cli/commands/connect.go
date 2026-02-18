package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Gahroot/agentHQ-cli/internal/common/client"
	"github.com/Gahroot/agentHQ-cli/internal/common/config"
	"github.com/Gahroot/agentHQ-cli/pkg/output"
	"github.com/spf13/cobra"
)

func NewConnectCmd() *cobra.Command {
	var hubURL, name string

	cmd := &cobra.Command{
		Use:   "connect <invite-token>",
		Short: "Connect to a hub using an invite token",
		Long:  "Redeem an invite token to register this machine as an agent and save credentials.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			token := args[0]

			if hubURL == "" {
				hubURL = "http://localhost:3000"
			}

			if name == "" {
				hostname, err := os.Hostname()
				if err != nil {
					hostname = "agent"
				}
				name = fmt.Sprintf("Agent - %s", hostname)
			}

			// No auth token needed for redeem endpoint
			c := client.NewWithToken(hubURL, "")
			resp, err := c.Post("/api/v1/auth/invites/redeem", map[string]string{
				"token":     token,
				"agentName": name,
			})
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to redeem invite: %v", err))
				return nil
			}

			var data struct {
				Agent struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"agent"`
				APIKey string `json:"apiKey"`
				OrgID  string `json:"orgId"`
			}
			if err := json.Unmarshal(resp.Data, &data); err != nil {
				output.PrintError(fmt.Sprintf("Failed to parse response: %v", err))
				return nil
			}

			cfg := &config.Config{
				HubURL:  hubURL,
				APIKey:  data.APIKey,
				OrgID:   data.OrgID,
				AgentID: data.Agent.ID,
			}
			if err := config.Save(cfg); err != nil {
				output.PrintError(fmt.Sprintf("Failed to save config: %v", err))
				return nil
			}

			output.PrintSuccess(fmt.Sprintf("Connected as %s (ID: %s)", data.Agent.Name, data.Agent.ID))
			fmt.Fprintf(os.Stderr, "Credentials saved to config. You're ready to go!\n")
			return nil
		},
	}

	cmd.Flags().StringVar(&hubURL, "hub-url", "", "Hub URL (default: http://localhost:3000)")
	cmd.Flags().StringVar(&name, "name", "", "Agent name (default: hostname-based)")

	return cmd
}
