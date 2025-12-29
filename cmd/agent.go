package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type TodoItem struct {
	ID       string     `json:"id"`
	Content  string     `json:"content"`
	Status   string     `json:"status"`
	Priority string     `json:"priority"`
	Children []TodoItem `json:"children,omitempty"`
}

type TodoList struct {
	Todos []TodoItem `json:"todos"`
}

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Agent commands for TODO.md, proposals, and documentation",
}

var todoWriteCmd = &cobra.Command{
	Use:   "todowrite",
	Short: "Write tasks to TODO.md in the current directory",
	Run:   runTodoWrite,
}

var todoReadCmd = &cobra.Command{
	Use:   "todoread",
	Short: "Read and display TODO.md from the current directory",
	Run:   runTodoRead,
}

func init() {
	agentCmd.Long = helpText("agent")
	todoWriteCmd.Long = helpText("agent-todowrite")
	todoReadCmd.Long = helpText("agent-todoread")

	rootCmd.AddCommand(agentCmd)

	agentCmd.AddCommand(todoWriteCmd)
	agentCmd.AddCommand(todoReadCmd)

	RegisterDocsCommand(agentCmd)
}

func runTodoWrite(cmd *cobra.Command, args []string) {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		printError(fmt.Sprintf("Failed to read from stdin: %v", err))
		return
	}

	if len(input) == 0 {
		printError("No input provided. Please provide JSON via stdin.")
		printDim("Example: echo '{\"todos\":[...]}' | nocturnal agent todowrite")
		return
	}

	var todoList TodoList
	if err := json.Unmarshal(input, &todoList); err != nil {
		printError(fmt.Sprintf("Failed to parse JSON: %v", err))
		return
	}

	if len(todoList.Todos) == 0 {
		printDim("No todos found in input")
		return
	}

	todoPath, count, err := writeTodoFile(todoList)
	if err != nil {
		printError(fmt.Sprintf("Failed to write TODO.md: %v", err))
		return
	}

	printSuccess(fmt.Sprintf("Wrote %d tasks to %s", count, todoPath))
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

	checkbox := "[ ]"
	switch todo.Status {
	case "completed":
		checkbox = "[x]"
	case "in_progress":
		checkbox = "[-]"
	case "cancelled":
		checkbox = "[~]"
	default:
		checkbox = "[ ]"
	}

	var lines []string

	lines = append(lines, fmt.Sprintf("%s- %s %s {#%s}", prefix, checkbox, todo.Content, todo.ID))

	for _, child := range todo.Children {
		lines = append(lines, formatTodoItem(child, indent+1)...)
	}

	return lines
}

func runTodoRead(cmd *cobra.Command, args []string) {
	content, err := readTodoFile()
	if err != nil {
		printError(err.Error())
		if strings.Contains(err.Error(), "not found") {
			printDim("Run 'nocturnal agent todowrite' to create one")
		}
		return
	}

	fmt.Println(content)
}
