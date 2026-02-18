package commands

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Gahroot/agentHQ-cli/internal/common/client"
	"github.com/Gahroot/agentHQ-cli/pkg/output"
	"github.com/spf13/cobra"
)

func NewTaskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Task management commands",
	}

	cmd.AddCommand(newTaskListCmd())
	cmd.AddCommand(newTaskCreateCmd())
	cmd.AddCommand(newTaskGetCmd())
	cmd.AddCommand(newTaskUpdateCmd())
	cmd.AddCommand(newTaskDeleteCmd())

	return cmd
}

// Task represents a task in the system
type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	AssignedTo  string     `json:"assigned_to"`
	AssignedType string    `json:"assigned_type"`
	ChannelID   string     `json:"channel_id"`
	DueDate     *time.Time `json:"due_date"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

func newTaskListCmd() *cobra.Command {
	var status, priority, assignedTo, channel string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			query := map[string]string{}
			if status != "" {
				query["status"] = status
			}
			if priority != "" {
				query["priority"] = priority
			}
			if assignedTo != "" {
				query["assigned_to"] = assignedTo
			}
			if channel != "" {
				query["channel"] = channel
			}

			resp, err := c.Get("/api/v1/tasks", query)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to list tasks: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(json.RawMessage(resp.Data))
				return nil
			}

			var tasks []Task
			if err := json.Unmarshal(resp.Data, &tasks); err != nil {
				output.PrintError(fmt.Sprintf("Failed to parse response: %v", err))
				return nil
			}

			if len(tasks) == 0 {
				output.PrintSuccess("No tasks found")
				return nil
			}

			rows := make([][]string, len(tasks))
			for i, t := range tasks {
				dueDate := ""
				if t.DueDate != nil {
					dueDate = t.DueDate.Format("2006-01-02")
				}
				rows[i] = []string{t.ID, t.Title, t.Status, t.Priority, dueDate}
			}
			output.PrintTable([]string{"ID", "TITLE", "STATUS", "PRIORITY", "DUE DATE"}, rows)
			return nil
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "Filter by status")
	cmd.Flags().StringVar(&priority, "priority", "", "Filter by priority")
	cmd.Flags().StringVar(&assignedTo, "assigned-to", "", "Filter by assigned agent")
	cmd.Flags().StringVar(&channel, "channel", "", "Filter by channel")

	return cmd
}

func newTaskCreateCmd() *cobra.Command {
	var description, status, priority, assignedTo, assignedType, channel, dueDate string

	cmd := &cobra.Command{
		Use:   "create --title <title>",
		Short: "Create a task",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			title, err := cmd.Flags().GetString("title")
			if err != nil || title == "" {
				output.PrintError("--title is required")
				return nil
			}

			body := map[string]interface{}{
				"title": title,
			}
			if description != "" {
				body["description"] = description
			}
			if status != "" {
				body["status"] = status
			}
			if priority != "" {
				body["priority"] = priority
			}
			if assignedTo != "" {
				body["assigned_to"] = assignedTo
			}
			if assignedType != "" {
				body["assigned_type"] = assignedType
			}
			if channel != "" {
				body["channel"] = channel
			}
			if dueDate != "" {
				body["due_date"] = dueDate
			}

			resp, err := c.Post("/api/v1/tasks", body)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to create task: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(json.RawMessage(resp.Data))
				return nil
			}

			var task Task
			if err := json.Unmarshal(resp.Data, &task); err != nil {
				output.PrintError(fmt.Sprintf("Failed to parse response: %v", err))
				return nil
			}
			output.PrintSuccess(fmt.Sprintf("Task created: %s (%s)", task.Title, task.ID))
			return nil
		},
	}

	cmd.Flags().String("title", "", "Task title (required)")
	cmd.Flags().StringVar(&description, "description", "", "Task description")
	cmd.Flags().StringVar(&status, "status", "", "Task status")
	cmd.Flags().StringVar(&priority, "priority", "", "Task priority")
	cmd.Flags().StringVar(&assignedTo, "assigned-to", "", "Assigned agent ID")
	cmd.Flags().StringVar(&assignedType, "assigned-type", "", "Assignment type")
	cmd.Flags().StringVar(&channel, "channel", "", "Channel ID")
	cmd.Flags().StringVar(&dueDate, "due-date", "", "Due date (ISO 8601)")

	cmd.MarkFlagRequired("title")

	return cmd
}

func newTaskGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a task by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			resp, err := c.Get(fmt.Sprintf("/api/v1/tasks/%s", args[0]), nil)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to get task: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(json.RawMessage(resp.Data))
				return nil
			}

			var task Task
			if err := json.Unmarshal(resp.Data, &task); err != nil {
				output.PrintError(fmt.Sprintf("Failed to parse response: %v", err))
				return nil
			}

			dueDate := "none"
			if task.DueDate != nil {
				dueDate = task.DueDate.Format("2006-01-02 15:04:05")
			}
			completedAt := "none"
			if task.CompletedAt != nil {
				completedAt = task.CompletedAt.Format("2006-01-02 15:04:05")
			}

			rows := [][]string{
				{"ID", task.ID},
				{"Title", task.Title},
				{"Description", task.Description},
				{"Status", task.Status},
				{"Priority", task.Priority},
				{"Assigned To", task.AssignedTo},
				{"Assigned Type", task.AssignedType},
				{"Channel ID", task.ChannelID},
				{"Due Date", dueDate},
				{"Created At", task.CreatedAt.Format("2006-01-02 15:04:05")},
				{"Completed At", completedAt},
			}
			output.PrintTable([]string{"FIELD", "VALUE"}, rows)
			return nil
		},
	}
}

func newTaskUpdateCmd() *cobra.Command {
	var title, description, status, priority, assignedTo, assignedType, channel, dueDate string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			body := map[string]interface{}{}
			if title != "" {
				body["title"] = title
			}
			if description != "" {
				body["description"] = description
			}
			if status != "" {
				body["status"] = status
			}
			if priority != "" {
				body["priority"] = priority
			}
			if assignedTo != "" {
				body["assigned_to"] = assignedTo
			}
			if assignedType != "" {
				body["assigned_type"] = assignedType
			}
			if channel != "" {
				body["channel"] = channel
			}
			if dueDate != "" {
				body["due_date"] = dueDate
			}

			resp, err := c.Patch(fmt.Sprintf("/api/v1/tasks/%s", args[0]), body)
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to update task: %v", err))
				return nil
			}

			if output.JSONMode {
				output.PrintJSON(json.RawMessage(resp.Data))
				return nil
			}

			var task Task
			if err := json.Unmarshal(resp.Data, &task); err != nil {
				output.PrintError(fmt.Sprintf("Failed to parse response: %v", err))
				return nil
			}
			output.PrintSuccess(fmt.Sprintf("Task updated: %s (%s)", task.Title, task.ID))
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Task title")
	cmd.Flags().StringVar(&description, "description", "", "Task description")
	cmd.Flags().StringVar(&status, "status", "", "Task status")
	cmd.Flags().StringVar(&priority, "priority", "", "Task priority")
	cmd.Flags().StringVar(&assignedTo, "assigned-to", "", "Assigned agent ID")
	cmd.Flags().StringVar(&assignedType, "assigned-type", "", "Assignment type")
	cmd.Flags().StringVar(&channel, "channel", "", "Channel ID")
	cmd.Flags().StringVar(&dueDate, "due-date", "", "Due date (ISO 8601)")

	return cmd
}

func newTaskDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New()
			if err != nil {
				output.PrintError(err.Error())
				return nil
			}

			_, err = c.Delete(fmt.Sprintf("/api/v1/tasks/%s", args[0]))
			if err != nil {
				output.PrintError(fmt.Sprintf("Failed to delete task: %v", err))
				return nil
			}

			output.PrintSuccess(fmt.Sprintf("Task deleted: %s", args[0]))
			return nil
		},
	}
}
