package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server exposing agent tools",
	Run:   runMCP,
}

func init() {
	mcpCmd.Long = helpText("mcp")
	rootCmd.AddCommand(mcpCmd)
}

func runMCP(cmd *cobra.Command, args []string) {
	s := server.NewMCPServer(
		"nocturnal",
		Version,
		server.WithToolCapabilities(true),
	)

	registerTodoWriteTool(s)
	registerTodoReadTool(s)
	registerCurrentTool(s)
	registerDocsListTool(s)
	registerDocsSearchTool(s)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
		os.Exit(1)
	}
}

func registerTodoWriteTool(s *server.MCPServer) {
	tool := mcp.NewTool("todowrite",
		mcp.WithDescription("Write tasks to TODO.md in the current directory. Takes a JSON object with a 'todos' array."),
		mcp.WithString("todos",
			mcp.Required(),
			mcp.Description("JSON string containing the todo list. Format: {\"todos\": [{\"id\": \"task-1\", \"content\": \"Task description\", \"status\": \"pending\", \"priority\": \"high\", \"children\": []}]}. Status values: pending, in_progress, completed, cancelled. Priority values: high, medium, low."),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		todosJSON, ok := request.Params.Arguments["todos"].(string)
		if !ok {
			return mcp.NewToolResultError("todos parameter must be a string"), nil
		}

		var todoList TodoList
		if err := json.Unmarshal([]byte(todosJSON), &todoList); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to parse JSON: %v", err)), nil
		}

		if len(todoList.Todos) == 0 {
			return mcp.NewToolResultText("No todos found in input"), nil
		}

		todoPath, count, err := writeTodoFile(todoList)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Wrote %d tasks to %s", count, todoPath)), nil
	})
}

func registerTodoReadTool(s *server.MCPServer) {
	tool := mcp.NewTool("todoread",
		mcp.WithDescription("Read and display TODO.md from the current directory."),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		content, err := readTodoFile()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(content), nil
	})
}

func registerCurrentTool(s *server.MCPServer) {
	tool := mcp.NewTool("current",
		mcp.WithDescription("Show the currently active proposal. Returns the active proposal content including specification, design, and implementation documents."),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		specPath, err := checkSpecWorkspace()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		slug, proposalPath, err := getActiveProposal(specPath)
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		if slug == "" {
			return mcp.NewToolResultText("No active proposal"), nil
		}

		header := fmt.Sprintf("Active proposal: %s\nLocation: %s\n\n", slug, proposalPath)
		docs, err := readProposalDocs(proposalPath)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(header + docs), nil
	})
}

func registerDocsListTool(s *server.MCPServer) {
	tool := mcp.NewTool("docs_list",
		mcp.WithDescription("List all available library and API documentation components."),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		components, err := loadDocs()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load docs: %v", err)), nil
		}

		if len(components) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("No documentation found. Create %s directory and add documentation files.", docsPath)), nil
		}

		return mcp.NewToolResultText(formatDocsListOutput(components)), nil
	})
}

func registerDocsSearchTool(s *server.MCPServer) {
	tool := mcp.NewTool("docs_search",
		mcp.WithDescription("Search library and API documentation by name. Returns full content of matching documentation."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query to match against component names"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, ok := request.Params.Arguments["query"].(string)
		if !ok {
			return mcp.NewToolResultError("query parameter must be a string"), nil
		}

		components, err := loadDocs()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load docs: %v", err)), nil
		}

		if len(components) == 0 {
			return mcp.NewToolResultText("No documentation found"), nil
		}

		matches := searchDocs(components, query)
		if len(matches) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("No components found matching '%s'. Use docs_list to see all available components.", query)), nil
		}

		return mcp.NewToolResultText(formatDocsSearchOutput(matches)), nil
	})
}
