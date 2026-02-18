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
	cmd.AddCommand(newPostGetCmd())
	cmd.AddCommand(newPostListCmd())
	cmd.AddCommand(newPostSearchCmd())
	cmd.AddCommand(newPostReplyCmd())
	cmd.AddCommand(newPostEditCmd())
	cmd.AddCommand(newPostDeleteCmd())
	cmd.AddCommand(newPostReactionsCmd())

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

func newPostGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a single post with thread",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			resp, err := c.Get("/api/v1/posts/"+args[0], nil)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to get post: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(json.RawMessage(resp.Data))
				return nil
			}

			var result struct {
				Post    struct {
					ID      string `json:"id"`
					Type    string `json:"type"`
					Title   string `json:"title"`
					Content string `json:"content"`
				} `json:"post"`
				Thread []struct {
					ID      string `json:"id"`
					Content string `json:"content"`
				} `json:"thread"`
			}
			if err := json.Unmarshal(resp.Data, &result); err != nil {
				output.PrintError(fmt.Sprintf("Failed to parse response: %v", err))
				return nil
			}

			fmt.Printf("ID: %s\n", result.Post.ID)
			fmt.Printf("Type: %s\n", result.Post.Type)
			if result.Post.Title != "" {
				fmt.Printf("Title: %s\n", result.Post.Title)
			}
			fmt.Printf("Content: %s\n", result.Post.Content)

			if len(result.Thread) > 0 {
				fmt.Printf("\nThread (%d replies):\n", len(result.Thread))
				for i, reply := range result.Thread {
					fmt.Printf("  %d. %s: %s\n", i+1, reply.ID, truncate(reply.Content, 60))
				}
			}

			return nil
		},
	}

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

func newPostReplyCmd() *cobra.Command {
	var content, channelID string

	cmd := &cobra.Command{
		Use:   "reply <id>",
		Short: "Reply to a post",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			body := map[string]string{
				"parent_id": args[0],
				"content":   content,
			}
			if channelID != "" {
				body["channel_id"] = channelID
			}

			resp, err := c.Post("/api/v1/posts", body)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to create reply: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(json.RawMessage(resp.Data))
				return nil
			}

			var reply struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(resp.Data, &reply); err != nil {
				output.PrintError(fmt.Sprintf("Failed to parse response: %v", err))
				return nil
			}
			output.PrintSuccess(fmt.Sprintf("Reply created: %s", reply.ID))
			return nil
		},
	}

	cmd.Flags().StringVar(&content, "content", "", "Reply content")
	cmd.Flags().StringVar(&channelID, "channel", "", "Channel ID (optional, defaults to parent's channel)")
	cmd.MarkFlagRequired("content")

	return cmd
}

func newPostEditCmd() *cobra.Command {
	var title, content string

	cmd := &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit a post",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			if title == "" && content == "" {
				output.PrintError("At least one of --title or --content is required")
				return nil
			}

			body := map[string]string{}
			if title != "" {
				body["title"] = title
			}
			if content != "" {
				body["content"] = content
			}

			resp, err := c.Patch("/api/v1/posts/"+args[0], body)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to edit post: %v", err))
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
			output.PrintSuccess(fmt.Sprintf("Post updated: %s", post.ID))
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "New title")
	cmd.Flags().StringVar(&content, "content", "", "New content")

	return cmd
}

func newPostDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a post",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			_, err = c.Delete("/api/v1/posts/" + args[0])
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to delete post: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(map[string]interface{}{"status": "deleted", "id": args[0]})
				return nil
			}

			output.PrintSuccess(fmt.Sprintf("Post deleted: %s", args[0]))
			return nil
		},
	}

	return cmd
}

func newPostReactionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reaction",
		Short: "Post reaction commands",
	}

	cmd.AddCommand(newPostReactionAddCmd())
	cmd.AddCommand(newPostReactionRemoveCmd())
	cmd.AddCommand(newPostReactionListCmd())

	return cmd
}

func newPostReactionAddCmd() *cobra.Command {
	var emoji string

	cmd := &cobra.Command{
		Use:   "add <id>",
		Short: "Add a reaction to a post",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			body := map[string]string{
				"emoji": emoji,
			}

			resp, err := c.Post("/api/v1/posts/"+args[0]+"/reactions", body)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to add reaction: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(json.RawMessage(resp.Data))
				return nil
			}

			var reaction struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(resp.Data, &reaction); err != nil {
				output.PrintError(fmt.Sprintf("Failed to parse response: %v", err))
				return nil
			}
			output.PrintSuccess(fmt.Sprintf("Reaction added: %s", reaction.ID))
			return nil
		},
	}

	cmd.Flags().StringVar(&emoji, "emoji", "", "Emoji to add")
	cmd.MarkFlagRequired("emoji")

	return cmd
}

func newPostReactionRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <id> <emoji>",
		Short: "Remove a reaction from a post",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			_, err = c.Delete("/api/v1/posts/" + args[0] + "/reactions/" + args[1])
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to remove reaction: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(map[string]interface{}{
					"status":  "removed",
					"post_id": args[0],
					"emoji":   args[1],
				})
				return nil
			}

			output.PrintSuccess(fmt.Sprintf("Reaction removed: %s from post %s", args[1], args[0]))
			return nil
		},
	}

	return cmd
}

func newPostReactionListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <id>",
		Short: "List reactions on a post",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			resp, err := c.Get("/api/v1/posts/"+args[0]+"/reactions", nil)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to list reactions: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(json.RawMessage(resp.Data))
				return nil
			}

			var reactions []struct {
				Emoji  string `json:"emoji"`
				Count  int    `json:"count"`
				UserID string `json:"user_id"`
			}
			if err := json.Unmarshal(resp.Data, &reactions); err != nil {
				output.PrintError(fmt.Sprintf("Failed to parse response: %v", err))
				return nil
			}

			if len(reactions) == 0 {
				output.PrintSuccess("No reactions found")
				return nil
			}

			rows := make([][]string, len(reactions))
			for i, r := range reactions {
				rows[i] = []string{r.Emoji, fmt.Sprintf("%d", r.Count), r.UserID}
			}
			output.PrintTable([]string{"EMOJI", "COUNT", "USER_ID"}, rows)
			return nil
		},
	}

	return cmd
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
