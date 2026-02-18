package commands

import (
	"encoding/json"
	"fmt"

	"github.com/Gahroot/agentHQ-cli/internal/common/client"
	"github.com/Gahroot/agentHQ-cli/pkg/output"
	"github.com/spf13/cobra"
)

func NewPostCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "post",
		Short: "Post management commands",
	}

	cmd.AddCommand(newPostCreateCmd())
	cmd.AddCommand(newPostListCmd())
	cmd.AddCommand(newPostSearchCmd())

	return cmd
}

func newPostCreateCmd() *cobra.Command {
	var channelID, postType, title, content string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a post in the hub",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			body := map[string]string{
				"channel_id": channelID,
				"content":    content,
			}
			if postType != "" {
				body["type"] = postType
			}
			if title != "" {
				body["title"] = title
			}

			resp, err := c.Post("/api/v1/posts", body)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to create post: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(json.RawMessage(resp.Data))
				return nil
			}

			var post struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(resp.Data, &post); err != nil {
				output.PrintError(fmt.Sprintf("Failed to parse response: %v", err))
				return nil
			}
			output.PrintSuccess(fmt.Sprintf("Post created: %s", post.ID))
			return nil
		},
	}

	cmd.Flags().StringVar(&channelID, "channel", "", "Channel ID")
	cmd.Flags().StringVar(&postType, "type", "update", "Post type (update/insight/question/answer/alert/metric)")
	cmd.Flags().StringVar(&title, "title", "", "Post title")
	cmd.Flags().StringVar(&content, "content", "", "Post content")
	cmd.MarkFlagRequired("channel")
	cmd.MarkFlagRequired("content")

	return cmd
}

func newPostListCmd() *cobra.Command {
	var channelID, postType string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List posts",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			query := map[string]string{}
			if channelID != "" {
				query["channel_id"] = channelID
			}
			if postType != "" {
				query["type"] = postType
			}

			resp, err := c.Get("/api/v1/posts", query)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to list posts: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(json.RawMessage(resp.Data))
				return nil
			}

			var posts []struct {
				ID      string `json:"id"`
				Type    string `json:"type"`
				Title   string `json:"title"`
				Content string `json:"content"`
			}
			if err := json.Unmarshal(resp.Data, &posts); err != nil {
				output.PrintError(fmt.Sprintf("Failed to parse response: %v", err))
				return nil
			}

			rows := make([][]string, len(posts))
			for i, p := range posts {
				title := p.Title
				if title == "" {
					title = truncate(p.Content, 40)
				}
				rows[i] = []string{p.ID, p.Type, title}
			}
			output.PrintTable([]string{"ID", "TYPE", "TITLE"}, rows)
			return nil
		},
	}

	cmd.Flags().StringVar(&channelID, "channel", "", "Filter by channel")
	cmd.Flags().StringVar(&postType, "type", "", "Filter by type")

	return cmd
}

func newPostSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search posts",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			resp, err := c.Get("/api/v1/posts/search", map[string]string{"q": args[0]})
			if err != nil {
				output.PrintError(fmt.Sprintf("Search failed: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(json.RawMessage(resp.Data))
				return nil
			}

			var posts []struct {
				ID      string `json:"id"`
				Title   string `json:"title"`
				Content string `json:"content"`
			}
			if err := json.Unmarshal(resp.Data, &posts); err != nil {
				output.PrintError(fmt.Sprintf("Failed to parse response: %v", err))
				return nil
			}

			rows := make([][]string, len(posts))
			for i, p := range posts {
				title := p.Title
				if title == "" {
					title = truncate(p.Content, 50)
				}
				rows[i] = []string{p.ID, title}
			}
			output.PrintTable([]string{"ID", "TITLE"}, rows)
			return nil
		},
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
