package commands

import (
	"encoding/json"
	"fmt"

	"github.com/Gahroot/agentHQ-cli/internal/common/client"
	"github.com/Gahroot/agentHQ-cli/pkg/output"
	"github.com/spf13/cobra"
)

func NewActivityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activity",
		Short: "Activity log commands",
	}

	cmd.AddCommand(newActivityLogCmd())
	cmd.AddCommand(newActivityListCmd())

	return cmd
}

func newActivityLogCmd() *cobra.Command {
	var action, resourceType, resourceID string

	cmd := &cobra.Command{
		Use:   "log",
		Short: "Log an activity",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			body := map[string]string{
				"action": action,
			}
			if resourceType != "" {
				body["resource_type"] = resourceType
			}
			if resourceID != "" {
				body["resource_id"] = resourceID
			}

			_, err = c.Post("/api/v1/activity", body)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to log activity: %v", err))
				return nil
			}

			output.PrintSuccess("Activity logged")
			return nil
		},
	}

	cmd.Flags().StringVar(&action, "action", "", "Action name (e.g., 'listing.viewed')")
	cmd.Flags().StringVar(&resourceType, "resource-type", "", "Resource type")
	cmd.Flags().StringVar(&resourceID, "resource-id", "", "Resource ID")
	cmd.MarkFlagRequired("action")

	return cmd
}

func newActivityListCmd() *cobra.Command {
	var actorID, action string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List activity log entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			query := map[string]string{}
			if actorID != "" {
				query["actor_id"] = actorID
			}
			if action != "" {
				query["action"] = action
			}

			resp, err := c.Get("/api/v1/activity", query)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to list activity: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(json.RawMessage(resp.Data))
				return nil
			}

			var entries []struct {
				ID        string `json:"id"`
				ActorID   string `json:"actor_id"`
				Action    string `json:"action"`
				CreatedAt string `json:"created_at"`
			}
			if err := json.Unmarshal(resp.Data, &entries); err != nil {
				output.PrintError(fmt.Sprintf("Failed to parse response: %v", err))
				return nil
			}

			rows := make([][]string, len(entries))
			for i, e := range entries {
				rows[i] = []string{e.ID, e.ActorID, e.Action, e.CreatedAt}
			}
			output.PrintTable([]string{"ID", "ACTOR", "ACTION", "TIME"}, rows)
			return nil
		},
	}

	cmd.Flags().StringVar(&actorID, "actor", "", "Filter by actor ID")
	cmd.Flags().StringVar(&action, "action", "", "Filter by action")

	return cmd
}
