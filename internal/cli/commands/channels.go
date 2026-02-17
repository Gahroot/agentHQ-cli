package commands

import (
	"encoding/json"
	"fmt"

	"github.com/agenthq/cli/internal/common/client"
	"github.com/agenthq/cli/pkg/output"
	"github.com/spf13/cobra"
)

func NewChannelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "channel",
		Short: "Channel management commands",
	}

	cmd.AddCommand(newChannelListCmd())
	cmd.AddCommand(newChannelCreateCmd())

	return cmd
}

func newChannelListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List channels",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			resp, err := c.Get("/api/v1/channels", nil)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to list channels: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(json.RawMessage(resp.Data))
				return nil
			}

			var channels []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Type string `json:"type"`
			}
			json.Unmarshal(resp.Data, &channels)

			rows := make([][]string, len(channels))
			for i, ch := range channels {
				rows[i] = []string{ch.ID, ch.Name, ch.Type}
			}
			output.PrintTable([]string{"ID", "NAME", "TYPE"}, rows)
			return nil
		},
	}
}

func newChannelCreateCmd() *cobra.Command {
	var description string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a channel",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			body := map[string]string{
				"name": args[0],
			}
			if description != "" {
				body["description"] = description
			}

			resp, err := c.Post("/api/v1/channels", body)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to create channel: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(json.RawMessage(resp.Data))
				return nil
			}

			var ch struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}
			json.Unmarshal(resp.Data, &ch)
			output.PrintSuccess(fmt.Sprintf("Channel created: %s (%s)", ch.Name, ch.ID))
			return nil
		},
	}

	cmd.Flags().StringVar(&description, "description", "", "Channel description")

	return cmd
}
