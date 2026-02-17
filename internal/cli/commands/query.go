package commands

import (
	"encoding/json"
	"fmt"

	"github.com/agenthq/cli/internal/common/client"
	"github.com/agenthq/cli/pkg/output"
	"github.com/spf13/cobra"
)

func NewQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query the hub",
	}

	cmd.AddCommand(newQueryAskCmd())

	return cmd
}

func newQueryAskCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ask <question>",
		Short: "Ask a natural language question",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			resp, err := c.Post("/api/v1/query", map[string]string{
				"question": args[0],
			})
			if err != nil {
				output.PrintError(fmt.Sprintf("Query failed: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(json.RawMessage(resp.Data))
				return nil
			}

			var data struct {
				Answer  string `json:"answer"`
				Sources []struct {
					ID    string `json:"id"`
					Title string `json:"title"`
				} `json:"sources"`
			}
			json.Unmarshal(resp.Data, &data)

			fmt.Println(data.Answer)
			if len(data.Sources) > 0 {
				fmt.Println("\nSources:")
				for _, s := range data.Sources {
					title := s.Title
					if title == "" {
						title = s.ID
					}
					fmt.Printf("  - %s (%s)\n", title, s.ID)
				}
			}
			return nil
		},
	}
}
