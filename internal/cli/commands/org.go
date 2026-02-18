package commands

import (
	"encoding/json"
	"fmt"

	"github.com/Gahroot/agentHQ-cli/internal/common/client"
	"github.com/Gahroot/agentHQ-cli/pkg/output"
	"github.com/spf13/cobra"
)

func NewOrgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "org",
		Short: "Organization management commands",
	}

	cmd.AddCommand(newOrgGetCmd())
	cmd.AddCommand(newOrgUpdateCmd())

	return cmd
}

func newOrgGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Get organization details",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			resp, err := c.Get("/api/v1/org", nil)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to get organization: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(json.RawMessage(resp.Data))
				return nil
			}

			var org struct {
				ID       string                 `json:"id"`
				Name     string                 `json:"name"`
				Settings map[string]interface{} `json:"settings"`
			}
			if err := json.Unmarshal(resp.Data, &org); err != nil {
				output.PrintError(fmt.Sprintf("Failed to parse response: %v", err))
				return nil
			}

			fmt.Printf("ID:\t%s\n", org.ID)
			fmt.Printf("Name:\t%s\n", org.Name)
			if len(org.Settings) > 0 {
				settingsJSON, _ := json.Marshal(org.Settings)
				fmt.Printf("Settings:\t%s\n", string(settingsJSON))
			} else {
				fmt.Printf("Settings:\t(empty)\n")
			}
			return nil
		},
	}
}

func newOrgUpdateCmd() *cobra.Command {
	var name, settingsStr string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			body := make(map[string]interface{})
			if name != "" {
				body["name"] = name
			}
			if settingsStr != "" {
				var settings map[string]interface{}
				if err := json.Unmarshal([]byte(settingsStr), &settings); err != nil {
					output.PrintError(fmt.Sprintf("Invalid settings JSON: %v", err))
					return nil
				}
				body["settings"] = settings
			}

			if len(body) == 0 {
				output.PrintError("At least one of --name or --settings must be provided")
				return nil
			}

			resp, err := c.Patch("/api/v1/org", body)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to update organization: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(json.RawMessage(resp.Data))
				return nil
			}

			var org struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}
			if err := json.Unmarshal(resp.Data, &org); err != nil {
				output.PrintError(fmt.Sprintf("Failed to parse response: %v", err))
				return nil
			}

			parts := []string{}
			if name != "" {
				parts = append(parts, fmt.Sprintf("name to %s", name))
			}
			if settingsStr != "" {
				parts = append(parts, "settings")
			}
			output.PrintSuccess(fmt.Sprintf("Organization updated: %s (%s)", org.Name, org.ID))
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Organization name")
	cmd.Flags().StringVar(&settingsStr, "settings", "", "Organization settings as JSON string")

	return cmd
}
