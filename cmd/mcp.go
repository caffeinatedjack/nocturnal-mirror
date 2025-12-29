package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

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
		server.WithPromptCapabilities(true),
	)

	registerRulesTool(s)
	registerCurrentTool(s)
	registerTasksTool(s)
	registerDocsListTool(s)
	registerDocsSearchTool(s)

	registerAddThirdPartyDocsPrompt(s)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
		os.Exit(1)
	}
}

func registerRulesTool(s *server.MCPServer) {
	tool := mcp.NewTool("rules",
		mcp.WithDescription("Get the project rules and design context."),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		specPath, err := checkSpecWorkspace()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		content, err := readRulesAndProject(specPath)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		if content == "" {
			return mcp.NewToolResultText("No project context found (no rules or project.md)"), nil
		}

		return mcp.NewToolResultText(content), nil
	})
}

func registerCurrentTool(s *server.MCPServer) {
	tool := mcp.NewTool("current",
		mcp.WithDescription("Show the currently active proposal. Returns the specification and design documents (not implementation)."),
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
		docs, err := readProposalSpecAndDesign(proposalPath)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(header + docs), nil
	})
}

func registerTasksTool(s *server.MCPServer) {
	tool := mcp.NewTool("tasks",
		mcp.WithDescription("Get the implementation tasks for the currently active proposal."),
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

		implPath := filepath.Join(proposalPath, "implementation.md")
		content, err := os.ReadFile(implPath)
		if err != nil {
			if os.IsNotExist(err) {
				return mcp.NewToolResultText(fmt.Sprintf("No implementation.md found for proposal '%s'", slug)), nil
			}
			return mcp.NewToolResultError(fmt.Sprintf("Failed to read implementation.md: %v", err)), nil
		}

		header := fmt.Sprintf("Implementation tasks for: %s\n\n", slug)
		return mcp.NewToolResultText(header + string(content)), nil
	})
}

// readProposalSpecAndDesign reads only the specification and design documents (not implementation)
func readProposalSpecAndDesign(proposalPath string) (string, error) {
	var buf bytes.Buffer

	specDesignDocs := []struct {
		Name string
		File string
	}{
		{"Specification", "specification.md"},
		{"Design", "design.md"},
	}

	for i, doc := range specDesignDocs {
		filePath := filepath.Join(proposalPath, doc.File)
		content, err := os.ReadFile(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			continue
		}

		if i > 0 {
			buf.WriteString("\n---\n\n")
		}

		buf.WriteString(fmt.Sprintf("## %s\n\n", doc.Name))
		buf.Write(content)
	}

	return buf.String(), nil
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

func registerAddThirdPartyDocsPrompt(s *server.MCPServer) {
	prompt := mcp.NewPrompt("add-third-party-docs",
		mcp.WithPromptDescription("Generate condensed documentation for third-party libraries"),
		mcp.WithArgument("urls",
			mcp.ArgumentDescription("Comma-separated list of documentation URLs to process"),
			mcp.RequiredArgument(),
		),
	)

	s.AddPrompt(prompt, func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		urls, _ := request.Params.Arguments["urls"]

		promptText := fmt.Sprintf(`You'll write a condensed version of the documentation to ~/.docs. If there are any key references missing, fetch them from the web as well. The goal is to develop a solid understanding of the library. These docs are intended to provide an AI agent with a clear overview of the library or technology, including its usage and where to find additional information. Be as concise as possible to avoid overwhelming the AI's context.

Separate each logical section with \n---\n, and immediately after the separator, include a header marked with #. Whenever possible, include direct links to the relevant documentation alongside any components or classes.

Documentation URLs to process:
%s`, urls)

		return &mcp.GetPromptResult{
			Description: "Instructions for creating condensed third-party documentation",
			Messages: []mcp.PromptMessage{
				{
					Role: mcp.RoleUser,
					Content: mcp.TextContent{
						Type: "text",
						Text: promptText,
					},
				},
			},
		}, nil
	})
}
