package commands

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Gahroot/agentHQ-cli/internal/common/client"
	"github.com/Gahroot/agentHQ-cli/pkg/output"
	"github.com/spf13/cobra"
)

func NewInsightsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "insights",
		Short: "Insight management commands",
	}

	cmd.AddCommand(newInsightsGenerateCmd())
	cmd.AddCommand(newInsightsListCmd())

	return cmd
}

func newInsightsGenerateCmd() *cobra.Command {
	var insightType string
	var title string
	var content string
	var confidence float64

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate insight",
		RunE: func(cmd *cobra.Command, args []string) error {
			if insightType == "" {
				output.PrintError("--type is required (trend/performance/recommendation/summary/anomaly)")
				return nil
			}
			if title == "" {
				output.PrintError("--title is required")
				return nil
			}
			if content == "" {
				output.PrintError("--content is required")
				return nil
			}

			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			body := map[string]interface{}{
				"type":    insightType,
				"title":   title,
				"content": content,
			}
			if confidence > 0 {
				body["confidence"] = confidence
			}

			resp, err := c.Post("/api/v1/insights/generate", body)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to generate insight: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(json.RawMessage(resp.Data))
				return nil
			}

			var insight struct {
				ID        string  `json:"id"`
				Type      string  `json:"type"`
				Title     string  `json:"title"`
				Confidence float64 `json:"confidence"`
			}
			if err := json.Unmarshal(resp.Data, &insight); err != nil {
				output.PrintError(fmt.Sprintf("Failed to parse response: %v", err))
				return nil
			}
			output.PrintSuccess(fmt.Sprintf("Insight generated: %s (%s)", insight.Title, insight.ID))
			return nil
		},
	}

	cmd.Flags().StringVar(&insightType, "type", "", "Insight type (trend/performance/recommendation/summary/anomaly)")
	cmd.Flags().StringVar(&title, "title", "", "Insight title")
	cmd.Flags().StringVar(&content, "content", "", "Insight content")
	cmd.Flags().Float64Var(&confidence, "confidence", 0, "Confidence score (0-1)")

	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("content")

	return cmd
}

func newInsightsListCmd() *cobra.Command {
	var insightType string
	var since string

	return &cobra.Command{
		Use:   "list",
		Short: "List insights",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			params := make(map[string]string)
			if insightType != "" {
				params["type"] = insightType
			}
			if since != "" {
				params["since"] = since
			}

			resp, err := c.Get("/api/v1/insights", params)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to list insights: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(json.RawMessage(resp.Data))
				return nil
			}

			var insights []struct {
				ID        string  `json:"id"`
				Type      string  `json:"type"`
				Title     string  `json:"title"`
				Confidence float64 `json:"confidence"`
			}
			if err := json.Unmarshal(resp.Data, &insights); err != nil {
				output.PrintError(fmt.Sprintf("Failed to parse response: %v", err))
				return nil
			}

			rows := make([][]string, len(insights))
			for i, insight := range insights {
				confidenceStr := ""
				if insight.Confidence > 0 {
					confidenceStr = strconv.FormatFloat(insight.Confidence, 'f', 2, 64)
				}
				rows[i] = []string{insight.ID, insight.Type, insight.Title, confidenceStr}
			}
			output.PrintTable([]string{"ID", "TYPE", "TITLE", "CONFIDENCE"}, rows)
			return nil
		},
	}
}
