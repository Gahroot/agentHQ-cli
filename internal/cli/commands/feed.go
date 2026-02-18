package commands

import (
	"encoding/json"
	"fmt"

	"github.com/Gahroot/agentHQ-cli/internal/common/client"
	"github.com/Gahroot/agentHQ-cli/pkg/output"
	"github.com/spf13/cobra"
)

func NewFeedCmd() *cobra.Command {
	var since, types, actorID string

	cmd := &cobra.Command{
		Use:   "feed",
		Short: "View unified timeline of recent hub activity",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			query := map[string]string{}
			if since != "" {
				query["since"] = since
			}
			if types != "" {
				query["types"] = types
			}
			if actorID != "" {
				query["actor_id"] = actorID
			}

			resp, err := c.Get("/api/v1/feed", query)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to get feed: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(json.RawMessage(resp.Data))
				return nil
			}

			var items []struct {
				ResourceType string `json:"resource_type"`
				ResourceID   string `json:"resource_id"`
				Timestamp    string `json:"timestamp"`
				Summary      string `json:"summary"`
			}
			if err := json.Unmarshal(resp.Data, &items); err != nil {
				output.PrintError(fmt.Sprintf("Failed to parse response: %v", err))
				return nil
			}

			if len(items) == 0 {
				fmt.Println("No recent activity.")
				return nil
			}

			rows := make([][]string, len(items))
			for i, item := range items {
				rows[i] = []string{item.Timestamp, item.ResourceType, item.Summary}
			}
			output.PrintTable([]string{"TIMESTAMP", "TYPE", "SUMMARY"}, rows)

			if resp.Pagination != nil && resp.Pagination.HasMore {
				fmt.Printf("\nShowing %d of %d items. Use --json for full data.\n", len(items), resp.Pagination.Total)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&since, "since", "", "ISO 8601 start time (default: 24h ago)")
	cmd.Flags().StringVar(&types, "types", "", "Comma-separated types (posts,activity,insights)")
	cmd.Flags().StringVar(&actorID, "actor", "", "Filter by actor/author ID")

	return cmd
}
