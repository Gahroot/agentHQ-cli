package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Gahroot/agentHQ-cli/internal/common/client"
	"github.com/Gahroot/agentHQ-cli/internal/common/config"
	"github.com/Gahroot/agentHQ-cli/pkg/output"
	"github.com/spf13/cobra"
)

// parseInviteArg accepts either a full invite URL or a bare token.
// Full URL: https://host.example.com/invite/AHQ-xxxxx-xxxx → hubURL + token
// Bare token: AHQ-xxxxx-xxxx → uses fallback hubURL
func parseInviteArg(arg string, fallbackURL string) (hubURL string, token string) {
	arg = strings.TrimSpace(arg)
	re := regexp.MustCompile(`^(https?://.+?)/invite/(AHQ-[A-Za-z0-9]+-[A-Za-z0-9]+)/?$`)
	if m := re.FindStringSubmatch(arg); m != nil {
		return m[1], m[2]
	}
	return fallbackURL, arg
}

func NewConnectCmd() *cobra.Command {
	var hubURL, name string

	cmd := &cobra.Command{
		Use:   "connect <invite-url-or-token>",
		Short: "Connect to a hub using an invite URL or token",
		Long: `Redeem an invite to register this machine as an agent and save credentials.

Accepts a full invite URL (recommended):
  agenthq connect https://hub.example.com/invite/AHQ-xxxxx-xxxx

Or a bare token with --hub-url:
  agenthq connect AHQ-xxxxx-xxxx --hub-url https://hub.example.com`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			parsedHub, token := parseInviteArg(args[0], hubURL)

			if parsedHub == "" {
				parsedHub = "http://localhost:3000"
			}

			if name == "" {
				hostname, err := os.Hostname()
				if err != nil {
					hostname = "agent"
				}
				name = fmt.Sprintf("Agent - %s", hostname)
			}

			// No auth token needed for redeem endpoint
			c := client.NewWithToken(parsedHub, "")
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
				HubURL:  parsedHub,
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

	cmd.Flags().StringVar(&hubURL, "hub-url", "", "Hub URL (only needed with bare tokens, not invite URLs)")
	cmd.Flags().StringVar(&name, "name", "", "Agent name (default: hostname-based)")

	return cmd
}
