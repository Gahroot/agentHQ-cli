package commands

import (
	"encoding/json"
	"fmt"

	"github.com/Gahroot/agentHQ-cli/internal/common/client"
	"github.com/Gahroot/agentHQ-cli/pkg/output"
	"github.com/spf13/cobra"
)

func NewNotificationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notifications",
		Short: "Notification management commands",
	}

	cmd.AddCommand(newNotificationsListCmd())
	cmd.AddCommand(newNotificationsUnreadCmd())
	cmd.AddCommand(newNotificationsReadCmd())
	cmd.AddCommand(newNotificationsReadAllCmd())

	return cmd
}

func newNotificationsListCmd() *cobra.Command {
	var notificationType, readStatus string
	var verbose bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List notifications",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			query := map[string]string{}
			if notificationType != "" {
				query["type"] = notificationType
			}
			if readStatus != "" {
				query["read"] = readStatus
			}

			resp, err := c.Get("/api/v1/notifications", query)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to list notifications: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(json.RawMessage(resp.Data))
				return nil
			}

			var notifications []struct {
				ID        string `json:"id"`
				Type      string `json:"type"`
				Read      bool   `json:"read"`
				Title     string `json:"title"`
				Body      string `json:"body"`
				CreatedAt string `json:"created_at"`
			}
			if err := json.Unmarshal(resp.Data, &notifications); err != nil {
				output.PrintError(fmt.Sprintf("Failed to parse response: %v", err))
				return nil
			}

			if len(notifications) == 0 {
				fmt.Println("No notifications found.")
				return nil
			}

			maxLen := 50
			if verbose {
				maxLen = 200
			}

			rows := make([][]string, len(notifications))
			for i, n := range notifications {
				readStatus := " "
				if n.Read {
					readStatus = "✓"
				}
				// Use title as primary display, body as fallback
				displayText := n.Title
				if displayText == "" && n.Body != "" {
					displayText = n.Body
				}
				// Truncate for table display
				if len(displayText) > maxLen {
					displayText = displayText[:maxLen-3] + "..."
				}
				rows[i] = []string{n.ID[:8], n.Type, readStatus, displayText}
			}
			output.PrintTable([]string{"ID", "TYPE", "READ", "TITLE"}, rows)

			if resp.Pagination != nil && resp.Pagination.HasMore {
				fmt.Printf("\nShowing %d of %d notifications. Use --json for full data.\n", len(notifications), resp.Pagination.Total)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&notificationType, "type", "", "Filter by notification type")
	cmd.Flags().StringVar(&readStatus, "read", "", "Filter by read status (true/false)")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Show longer content in table display")

	return cmd
}

func newNotificationsUnreadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unread",
		Short: "Show unread notification count",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			resp, err := c.Get("/api/v1/notifications/unread-count", nil)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to get unread count: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(json.RawMessage(resp.Data))
				return nil
			}

			var result struct {
				Count int `json:"count"`
			}
			if err := json.Unmarshal(resp.Data, &result); err != nil {
				output.PrintError(fmt.Sprintf("Failed to parse response: %v", err))
				return nil
			}

			if result.Count == 0 {
				fmt.Println("No unread notifications.")
			} else {
				fmt.Printf("You have %d unread notification%s.\n", result.Count, plural(result.Count))
			}

			return nil
		},
	}

	return cmd
}

func newNotificationsReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read <id>",
		Short: "Mark notification as read",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			id := args[0]
			_, err = c.Patch(fmt.Sprintf("/api/v1/notifications/%s/read", id), nil)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to mark notification as read: %v", err))
				return nil
			}

			output.PrintSuccess(fmt.Sprintf("Notification %s marked as read", id))
			return nil
		},
	}

	return cmd
}

func newNotificationsReadAllCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read-all",
		Short: "Mark all notifications as read",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			_, err = c.Post("/api/v1/notifications/read-all", nil)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to mark all as read: %v", err))
				return nil
			}

			output.PrintSuccess("All notifications marked as read")
			return nil
		},
	}

	return cmd
}

// plural returns "s" for n != 1, empty string otherwise.
func plural(n int) string {
	if n != 1 {
		return "s"
	}
	return ""
}
