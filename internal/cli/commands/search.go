package commands

import (
	"encoding/json"
	"fmt"

	"github.com/Gahroot/agentHQ-cli/internal/common/client"
	"github.com/Gahroot/agentHQ-cli/pkg/output"
	"github.com/spf13/cobra"
)

func NewSearchCmd() *cobra.Command {
	var types string

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search across posts, insights, and agents",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			query := map[string]string{"q": args[0]}
			if types != "" {
				query["types"] = types
			}

			resp, err := c.Get("/api/v1/search", query)
			if err != nil {
				output.PrintError(fmt.Sprintf("Search failed: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(json.RawMessage(resp.Data))
				return nil
			}

			var data struct {
				Posts []struct {
					ID      string `json:"id"`
					Title   string `json:"title"`
					Content string `json:"content"`
				} `json:"posts"`
				Insights []struct {
					ID    string `json:"id"`
					Title string `json:"title"`
					Type  string `json:"type"`
				} `json:"insights"`
				Agents []struct {
					ID     string `json:"id"`
					Name   string `json:"name"`
					Status string `json:"status"`
				} `json:"agents"`
			}
			if err := json.Unmarshal(resp.Data, &data); err != nil {
				output.PrintError(fmt.Sprintf("Failed to parse response: %v", err))
				return nil
			}

			if len(data.Posts) > 0 {
				fmt.Println("Posts:")
				rows := make([][]string, len(data.Posts))
				for i, p := range data.Posts {
					title := p.Title
					if title == "" {
						title = truncate(p.Content, 50)
					}
					rows[i] = []string{p.ID, title}
				}
				output.PrintTable([]string{"ID", "TITLE"}, rows)
				fmt.Println()
			}

			if len(data.Insights) > 0 {
				fmt.Println("Insights:")
				rows := make([][]string, len(data.Insights))
				for i, ins := range data.Insights {
					rows[i] = []string{ins.ID, ins.Type, ins.Title}
				}
				output.PrintTable([]string{"ID", "TYPE", "TITLE"}, rows)
				fmt.Println()
			}

			if len(data.Agents) > 0 {
				fmt.Println("Agents:")
				rows := make([][]string, len(data.Agents))
				for i, a := range data.Agents {
					rows[i] = []string{a.ID, a.Name, a.Status}
				}
				output.PrintTable([]string{"ID", "NAME", "STATUS"}, rows)
			}

			if len(data.Posts) == 0 && len(data.Insights) == 0 && len(data.Agents) == 0 {
				fmt.Println("No results found.")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&types, "types", "", "Comma-separated resource types to search (posts,insights,agents)")

	return cmd
}
