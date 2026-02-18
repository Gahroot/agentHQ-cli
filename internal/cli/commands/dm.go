package commands

import (
	"encoding/json"
	"fmt"

	"github.com/Gahroot/agentHQ-cli/internal/common/client"
	"github.com/Gahroot/agentHQ-cli/pkg/output"
	"github.com/spf13/cobra"
)

func NewDMCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dm",
		Short: "DM conversation commands",
	}

	cmd.AddCommand(newDMListCmd())
	cmd.AddCommand(newDMStartCmd())

	return cmd
}

func newDMListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List DM conversations",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			resp, err := c.Get("/api/v1/dm", nil)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to list DMs: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(json.RawMessage(resp.Data))
				return nil
			}

			var dms []struct {
				ID         string `json:"id"`
				Name       string `json:"name"`
				MemberID   string `json:"member_id"`
				MemberType string `json:"member_type"`
			}
			if err := json.Unmarshal(resp.Data, &dms); err != nil {
				output.PrintError(fmt.Sprintf("Failed to parse response: %v", err))
				return nil
			}

			rows := make([][]string, len(dms))
			for i, dm := range dms {
				rows[i] = []string{dm.ID, dm.Name, dm.MemberID, dm.MemberType}
			}
			output.PrintTable([]string{"ID", "NAME", "MEMBER_ID", "MEMBER_TYPE"}, rows)
			return nil
		},
	}
}

func newDMStartCmd() *cobra.Command {
	var memberType string

	cmd := &cobra.Command{
		Use:   "start <member-id>",
		Short: "Start DM conversation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			if memberType == "" {
				output.PrintError("--member-type is required")
				return nil
			}

			body := map[string]string{
				"member_id":   args[0],
				"member_type": memberType,
			}

			resp, err := c.Post("/api/v1/dm", body)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to start DM: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(json.RawMessage(resp.Data))
				return nil
			}

			var dm struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}
			if err := json.Unmarshal(resp.Data, &dm); err != nil {
				output.PrintError(fmt.Sprintf("Failed to parse response: %v", err))
				return nil
			}
			output.PrintSuccess(fmt.Sprintf("DM started: %s (%s)", dm.Name, dm.ID))
			return nil
		},
	}

	cmd.Flags().StringVar(&memberType, "member-type", "", "Member type (required)")

	return cmd
}
