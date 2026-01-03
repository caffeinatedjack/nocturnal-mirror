package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

const integrityWarning = `WARNING: Proposal files have changed since activation.

Changed files: %s

User confirmation is required before continuing. Please ask the user to either:
1. Re-activate the proposal to update the file hashes: nocturnal spec proposal activate %s
2. Or confirm they want to proceed with the modified files

Do not proceed with implementation until the user confirms.`

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

	// Tools (breaking change by design): keep only context + docs.
	registerContextTool(s)
	registerDocsListTool(s)
	registerDocsSearchTool(s)

	// Prompts
	registerAddThirdPartyDocsPrompt(s)
	registerStartImplementationPrompt(s)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
		os.Exit(1)
	}
}

func registerContextTool(s *server.MCPServer) {
	tool := mcp.NewTool("context",
		mcp.WithDescription("Get the minimum project context needed to implement the active proposal (rules, proposal docs, and tasks)."),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		specPath, err := checkSpecWorkspace()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		var sections []string

		// Rules + project design (constraints)
		content, err := readRulesAndProject(specPath)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if content != "" {
			sections = append(sections, content)
		}

		slug, proposalPath, err := getPrimaryProposal(specPath)
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		if slug == "" {
			if len(sections) == 0 {
				return mcp.NewToolResultText("No project context found (no rules or project.md)\n\nNo active proposal"), nil
			}
			sections = append(sections, "# Active Proposal\n\nNo active proposal")
			return mcp.NewToolResultText(strings.Join(sections, "\n\n---\n\n")), nil
		}

		// Integrity check: if proposal changed since activation, stop.
		changed, requiresConfirmation, err := checkProposalIntegrity(specPath, slug)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to verify proposal integrity: %v", err)), nil
		}
		if requiresConfirmation {
			return mcp.NewToolResultText(fmt.Sprintf(integrityWarning, strings.Join(changed, ", "), slug)), nil
		}

		activeHeader := fmt.Sprintf("# Active Proposal: %s\n\nLocation: %s\n", slug, proposalPath)

		// Proposal specification + design
		docs, err := readProposalDocsFiltered(proposalPath, []string{"specification.md", "design.md"})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if docs != "" {
			sections = append(sections, activeHeader+"\n"+docs)
		} else {
			sections = append(sections, activeHeader+"\n(No specification.md or design.md found)")
		}

		// Tasks (open tasks only, to keep context lean)
		implPath := filepath.Join(proposalPath, "implementation.md")
		implContent, err := os.ReadFile(implPath)
		if err != nil {
			if os.IsNotExist(err) {
				sections = append(sections, "# Implementation Tasks\n\nNo implementation.md found")
				return mcp.NewToolResultText(strings.Join(sections, "\n\n---\n\n")), nil
			}
			return mcp.NewToolResultError(fmt.Sprintf("Failed to read implementation.md: %v", err)), nil
		}

		total, completed := getProposalProgress(proposalPath)
		openTasks := extractOpenTasks(string(implContent))

		tasksHeader := fmt.Sprintf("# Implementation Tasks\n\nProgress: %d/%d complete\n\n", completed, total)
		if len(openTasks) == 0 {
			sections = append(sections, tasksHeader+"No open tasks found")
		} else {
			sections = append(sections, tasksHeader+strings.Join(openTasks, "\n"))
		}

		return mcp.NewToolResultText(strings.Join(sections, "\n\n---\n\n")), nil
	})
}

func extractOpenTasks(content string) []string {
	var tasks []string
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- [ ]") {
			tasks = append(tasks, trimmed)
		}
	}
	return tasks
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

		promptText := fmt.Sprintf(`You'll write a condensed version of the documentation to spec/third. If there are any key references missing, fetch them from the web as well. The goal is to develop a solid understanding of the library. These docs are intended to provide an AI agent with a clear overview of the library or technology, including its usage and where to find additional information. Be as concise as possible to avoid overwhelming the AI's context.

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

func registerStartImplementationPrompt(s *server.MCPServer) {
	prompt := mcp.NewPrompt("start-implementation",
		mcp.WithPromptDescription("Start an implementation using Nocturnal context tools"),
		mcp.WithArgument("goal",
			mcp.ArgumentDescription("Short description of what you want to implement (optional)"),
		),
	)

	s.AddPrompt(prompt, func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		goalStr := ""
		goal := strings.TrimSpace(request.Params.Arguments["goal"])
		if goal != "" {
			goalStr = fmt.Sprintf("Goal: %s\n\n", goal)
		}

		promptText := fmt.Sprintf(`%sYou will implement the active proposal in this repository.

1) Call the MCP tool "context".
2) Read the returned context carefully and treat it as the source of truth.
3) If the tool returns an integrity warning about modified proposal files, STOP and ask the user to either re-activate the proposal or confirm proceeding.
4) Identify the key constraints (rules + non-goals) and the MUST-level requirements from the specification.
5) Implement changes in small, reviewable steps. Do not introduce unrelated refactors.
6) If you need third-party API details, use docs_search (and only fetch what you need).
7) When you finish, summarize what you changed and what requirements you believe are satisfied.
8) Ask the user to run tests/linters and confirm results (do not claim you ran anything).
`, goalStr)

		return &mcp.GetPromptResult{
			Description: "Bootstrap instructions for implementing using Nocturnal context",
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
