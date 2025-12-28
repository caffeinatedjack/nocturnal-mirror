package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/nocturnal/pkg/ui"
)

// TodoItem represents a single todo item
type TodoItem struct {
	ID       string     `json:"id"`
	Content  string     `json:"content"`
	Status   string     `json:"status"`   // "pending", "in_progress", "completed", "cancelled"
	Priority string     `json:"priority"` // "high", "medium", "low"
	Children []TodoItem `json:"children,omitempty"`
}

// TodoList represents the structure expected from stdin
type TodoList struct {
	Todos []TodoItem `json:"todos"`
}

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Agent commands for working with TODO.md files",
	Long: `Agent commands for reading and writing todo lists to TODO.md.

Commands:
    todowrite  Write tasks to TODO.md in the current directory
    todoread   Read and display TODO.md from the current directory

Examples:
    nocturnal agent todowrite < todos.json
    nocturnal agent todoread`,
}

var todoWriteCmd = &cobra.Command{
	Use:   "todowrite",
	Short: "Write tasks to TODO.md in the current directory",
	Long: `Write tasks to TODO.md in the current directory.

This command reads a JSON todo list from stdin and writes it to TODO.md
in the current working directory.

Input format (JSON):
{
  "todos": [
    {
      "id": "task-1",
      "content": "Task description",
      "status": "pending",
      "priority": "high"
    }
  ]
}

Status values: pending, in_progress, completed, cancelled
Priority values: high, medium, low

Examples:
    nocturnal agent todowrite < todos.json
    echo '{"todos":[{"id":"1","content":"Test","status":"pending","priority":"high"}]}' | nocturnal agent todowrite`,
	Run: runTodoWrite,
}

var todoReadCmd = &cobra.Command{
	Use:   "todoread",
	Short: "Read and display TODO.md from the current directory",
	Long: `Read and display TODO.md from the current directory.

This command reads and displays the TODO.md file in the current working directory.

Examples:
    nocturnal agent todoread`,
	Run: runTodoRead,
}

func init() {
	rootCmd.AddCommand(agentCmd)

	// Add subcommands to agentCmd
	agentCmd.AddCommand(todoWriteCmd)
	agentCmd.AddCommand(todoReadCmd)
}

func runTodoWrite(cmd *cobra.Command, args []string) {
	// Read JSON from stdin
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to read from stdin: %v", err))
		return
	}

	if len(input) == 0 {
		ui.Error("No input provided. Please provide JSON via stdin.")
		ui.PrintDim("Example: echo '{\"todos\":[...]}' | nocturnal agent todowrite")
		return
	}

	// Parse JSON
	var todoList TodoList
	if err := json.Unmarshal(input, &todoList); err != nil {
		ui.Error(fmt.Sprintf("Failed to parse JSON: %v", err))
		return
	}

	if len(todoList.Todos) == 0 {
		ui.PrintDim("No todos found in input")
		return
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to get current directory: %v", err))
		return
	}
	todoPath := filepath.Join(cwd, "TODO.md")

	// Write TODO.md
	content := generateTodoContent(todoList.Todos)
	if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
		ui.Error(fmt.Sprintf("Failed to write TODO.md: %v", err))
		return
	}

	ui.Success(fmt.Sprintf("Wrote %d tasks to %s", len(todoList.Todos), todoPath))
}

func generateTodoContent(todos []TodoItem) string {
	var lines []string
	lines = append(lines, "# TODO")
	lines = append(lines, "")

	for _, todo := range todos {
		lines = append(lines, formatTodoItem(todo, 0)...)
	}

	return strings.Join(lines, "\n")
}

func formatTodoItem(todo TodoItem, indent int) []string {
	prefix := strings.Repeat("  ", indent)

	// Determine checkbox based on status
	checkbox := "[ ]"
	switch todo.Status {
	case "completed":
		checkbox = "[x]"
	case "in_progress":
		checkbox = "[-]"
	case "cancelled":
		checkbox = "[~]"
	default: // "pending" or unknown
		checkbox = "[ ]"
	}

	var lines []string

	// Task line with ID
	lines = append(lines, fmt.Sprintf("%s- %s %s {#%s}", prefix, checkbox, todo.Content, todo.ID))

	// Children/subtasks
	for _, child := range todo.Children {
		lines = append(lines, formatTodoItem(child, indent+1)...)
	}

	return lines
}

func runTodoRead(cmd *cobra.Command, args []string) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to get current directory: %v", err))
		return
	}
	todoPath := filepath.Join(cwd, "TODO.md")

	// Read TODO.md
	content, err := os.ReadFile(todoPath)
	if err != nil {
		if os.IsNotExist(err) {
			ui.Error("TODO.md not found in current directory")
			ui.PrintDim("Run 'nocturnal agent todowrite' to create one")
		} else {
			ui.Error(fmt.Sprintf("Failed to read TODO.md: %v", err))
		}
		return
	}

	// Display contents
	fmt.Println(string(content))
}
