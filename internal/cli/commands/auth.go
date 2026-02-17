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

func NewAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication commands",
	}

	cmd.AddCommand(newLoginCmd())
	cmd.AddCommand(newLoginAgentCmd())
	cmd.AddCommand(newWhoamiCmd())
	cmd.AddCommand(newLogoutCmd())
	cmd.AddCommand(newExportCmd())

	return cmd
}

func newLoginCmd() *cobra.Command {
	var email, password, hubURL string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login as a human user",
		RunE: func(cmd *cobra.Command, args []string) error {
			if hubURL == "" {
				hubURL = "http://localhost:3000"
			}
			c := client.NewWithToken(hubURL, "")
			resp, err := c.Post("/api/v1/auth/login", map[string]string{
				"email":    email,
				"password": password,
			})
			if err != nil {
				output.PrintError(fmt.Sprintf("Login failed: %v", err))
				return nil
			}

			var data struct {
				User struct {
					ID    string `json:"id"`
					Email string `json:"email"`
					Name  string `json:"name"`
					OrgID string `json:"org_id"`
				} `json:"user"`
				AccessToken  string `json:"accessToken"`
				RefreshToken string `json:"refreshToken"`
			}
			json.Unmarshal(resp.Data, &data)

			cfg := &config.Config{
				HubURL:   hubURL,
				JWTToken: data.AccessToken,
				OrgID:    data.User.OrgID,
			}
			if err := config.Save(cfg); err != nil {
				output.PrintError(fmt.Sprintf("Failed to save config: %v", err))
				return nil
			}

			output.PrintSuccess(fmt.Sprintf("Logged in as %s (%s)", data.User.Name, data.User.Email))
			return nil
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "Email address")
	cmd.Flags().StringVar(&password, "password", "", "Password")
	cmd.Flags().StringVar(&hubURL, "hub-url", "", "Hub URL (default: http://localhost:3000)")
	cmd.MarkFlagRequired("email")
	cmd.MarkFlagRequired("password")

	return cmd
}

func newLoginAgentCmd() *cobra.Command {
	var name, description, hubURL, token string

	cmd := &cobra.Command{
		Use:   "login-agent",
		Short: "Register this machine as an agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			if hubURL == "" {
				hubURL = "http://localhost:3000"
			}
			c := client.NewWithToken(hubURL, token)
			resp, err := c.Post("/api/v1/auth/agents/register", map[string]string{
				"name":        name,
				"description": description,
			})
			if err != nil {
				output.PrintError(fmt.Sprintf("Agent registration failed: %v", err))
				return nil
			}

			var data struct {
				Agent struct {
					ID    string `json:"id"`
					OrgID string `json:"org_id"`
				} `json:"agent"`
				APIKey string `json:"apiKey"`
			}
			json.Unmarshal(resp.Data, &data)

			cfg := &config.Config{
				HubURL:  hubURL,
				APIKey:  data.APIKey,
				OrgID:   data.Agent.OrgID,
				AgentID: data.Agent.ID,
			}
			if err := config.Save(cfg); err != nil {
				output.PrintError(fmt.Sprintf("Failed to save config: %v", err))
				return nil
			}

			output.PrintSuccess(fmt.Sprintf("Agent registered: %s (ID: %s)", name, data.Agent.ID))
			fmt.Fprintf(os.Stderr, "API Key saved to config. Keep it safe!\n")
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Agent name")
	cmd.Flags().StringVar(&description, "description", "", "Agent description")
	cmd.Flags().StringVar(&hubURL, "hub-url", "", "Hub URL")
	cmd.Flags().StringVar(&token, "token", "", "JWT token for auth")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("token")

	return cmd
}

func newWhoamiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show current identity",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				output.PrintError("Not logged in")
				return nil
			}

			if cfg.APIKey != "" {
				output.PrintSuccess(fmt.Sprintf("Agent ID: %s, Org: %s, Hub: %s", cfg.AgentID, cfg.OrgID, cfg.HubURL))
			} else if cfg.JWTToken != "" {
				output.PrintSuccess(fmt.Sprintf("User, Org: %s, Hub: %s", cfg.OrgID, cfg.HubURL))
			} else {
				output.PrintError("Not logged in")
			}
			return nil
		},
	}
}

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Clear stored credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &config.Config{HubURL: "http://localhost:3000"}
			if err := config.Save(cfg); err != nil {
				output.PrintError(fmt.Sprintf("Failed to clear config: %v", err))
				return nil
			}
			output.PrintSuccess("Logged out")
			return nil
		},
	}
}

// newExportCmd outputs connection info in a format suitable for pocket-agent
func newExportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "export",
		Short: "Export connection info for pocket-agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				output.PrintError("Not logged in")
				return nil
			}

			if cfg.APIKey == "" {
				output.PrintError("No agent credentials found. Please run 'agenthq auth login-agent' first.")
				return nil
			}

			// Output in a clean format for pocket-agent to consume
			fmt.Println("AgentHQ Connection Info:")
			fmt.Printf("  HUB_URL=%s\n", cfg.HubURL)
			fmt.Printf("  AGENTHQ_API_KEY=%s\n", cfg.APIKey)
			fmt.Printf("  AGENTHQ_AGENT_ID=%s\n", cfg.AgentID)
			fmt.Printf("  AGENTHQ_ORG_ID=%s\n", cfg.OrgID)

			// Also output as JSON for programmatic use
			fmt.Println("\nJSON format:")
			exportData := map[string]string{
				"hur_url":  cfg.HubURL,
				"api_key":  cfg.APIKey,
				"agent_id": cfg.AgentID,
				"org_id":   cfg.OrgID,
			}
			jsonBytes, _ := json.MarshalIndent(exportData, "  ", "  ")
			fmt.Printf("  %s\n", string(jsonBytes))

			return nil
		},
	}
}
