package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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

	// Tools
	registerContextTool(s)
	registerTasksTool(s)
	registerTaskCompleteTool(s)
	registerTaskSnapshotTool(s)
	registerDocsListTool(s)
	registerDocsSearchTool(s)
	registerMaintenanceListTool(s)
	registerMaintenanceContextTool(s)
	registerMaintenanceActionedTool(s)

	// Prompts
	registerAddThirdPartyDocsPrompt(s)
	registerStartImplementationPrompt(s)
	registerLazyPrompt(s)
	registerStartMaintenancePrompt(s)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
		os.Exit(1)
	}
}

func registerContextTool(s *server.MCPServer) {
	tool := mcp.NewTool("context",
		mcp.WithDescription("Get project context for implementing the active proposal: rules, project design, specification, and design docs."),
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

		// Include affected files if configured
		config := loadConfigOrDefault(specPath)
		if config.Context.IncludeAffectedFiles {
			affectedFiles, err := getAffectedFiles(proposalPath)
			if err == nil && len(affectedFiles) > 0 {
				affectedSection := buildAffectedFilesSection(affectedFiles, config.Context.MaxFileLines)
				if affectedSection != "" {
					sections = append(sections, affectedSection)
				}
			}
		}

		return mcp.NewToolResultText(strings.Join(sections, "\n\n---\n\n")), nil
	})
}

func registerTasksTool(s *server.MCPServer) {
	tool := mcp.NewTool("tasks",
		mcp.WithDescription("Get the current phase tasks for the active proposal. Shows only the first incomplete phase. Use task_snapshot before starting a task (optional, if git integration enabled) and task_complete to mark tasks done."),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		specPath, err := checkSpecWorkspace()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		slug, proposalPath, err := getPrimaryProposal(specPath)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if slug == "" {
			return mcp.NewToolResultError("No active proposal"), nil
		}

		implPath := filepath.Join(proposalPath, "implementation.md")
		implContent, err := os.ReadFile(implPath)
		if err != nil {
			if os.IsNotExist(err) {
				return mcp.NewToolResultText("No implementation.md found"), nil
			}
			return mcp.NewToolResultError(fmt.Sprintf("Failed to read implementation.md: %v", err)), nil
		}

		total, completed := getProposalProgress(proposalPath)
		phases := extractPhases(string(implContent))

		var result strings.Builder
		result.WriteString(fmt.Sprintf("# Tasks for: %s\n\n", slug))
		result.WriteString(fmt.Sprintf("Progress: %d/%d tasks complete\n\n", completed, total))

		if len(phases) == 0 {
			result.WriteString("No phases found in implementation.md")
			return mcp.NewToolResultText(result.String()), nil
		}

		// Find current phase number
		currentPhaseNum := 0
		for i, p := range phases {
			if !p.Complete {
				currentPhaseNum = i + 1
				break
			}
		}

		currentPhase := getCurrentPhase(phases)
		if currentPhase == nil {
			result.WriteString("All phases complete!")
		} else {
			result.WriteString(formatPhaseForContext(currentPhase, currentPhaseNum, len(phases)))
			result.WriteString("\n\n> Once all tasks in this phase are complete, the next phase will appear.")
		}

		return mcp.NewToolResultText(result.String()), nil
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

// Phase represents a phase from implementation.md with its tasks.
type Phase struct {
	Name      string
	Goal      string
	Tasks     []Task
	Milestone string
	Complete  bool
}

// Task represents a single task checkbox.
type Task struct {
	ID       string // Task ID in format "phase.task" (e.g., "1.1", "2.3")
	Text     string
	Complete bool
	Line     int // Line number in the file (1-indexed)
}

// extractPhases parses implementation.md content and returns all phases with their tasks.
func extractPhases(content string) []Phase {
	var phases []Phase
	var currentPhase *Phase
	phaseNum := 0
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for phase header (### Phase N: Name)
		if strings.HasPrefix(trimmed, "### Phase") {
			// Save previous phase if exists
			if currentPhase != nil {
				currentPhase.Complete = isPhaseComplete(currentPhase)
				phases = append(phases, *currentPhase)
			}

			phaseNum++

			// Extract phase name
			name := trimmed
			if colonIdx := strings.Index(trimmed, ":"); colonIdx != -1 {
				name = strings.TrimSpace(trimmed[colonIdx+1:])
			}

			currentPhase = &Phase{Name: name}
			continue
		}

		// Only process if we're inside a phase
		if currentPhase == nil {
			continue
		}

		// Check for Goal
		if strings.HasPrefix(trimmed, "**Goal**:") {
			currentPhase.Goal = strings.TrimSpace(strings.TrimPrefix(trimmed, "**Goal**:"))
			continue
		}

		// Check for Milestone
		if strings.HasPrefix(trimmed, "**Milestone**:") {
			currentPhase.Milestone = strings.TrimSpace(strings.TrimPrefix(trimmed, "**Milestone**:"))
			continue
		}

		// Check for task checkboxes
		if strings.HasPrefix(trimmed, "- [ ]") {
			taskNum := len(currentPhase.Tasks) + 1
			currentPhase.Tasks = append(currentPhase.Tasks, Task{
				ID:       fmt.Sprintf("%d.%d", phaseNum, taskNum),
				Text:     strings.TrimSpace(strings.TrimPrefix(trimmed, "- [ ]")),
				Complete: false,
				Line:     i + 1,
			})
		} else if strings.HasPrefix(trimmed, "- [x]") || strings.HasPrefix(trimmed, "- [X]") {
			text := trimmed
			if strings.HasPrefix(trimmed, "- [x]") {
				text = strings.TrimPrefix(trimmed, "- [x]")
			} else {
				text = strings.TrimPrefix(trimmed, "- [X]")
			}
			taskNum := len(currentPhase.Tasks) + 1
			currentPhase.Tasks = append(currentPhase.Tasks, Task{
				ID:       fmt.Sprintf("%d.%d", phaseNum, taskNum),
				Text:     strings.TrimSpace(text),
				Complete: true,
				Line:     i + 1,
			})
		}

		// Stop at next major section (## header that's not a phase)
		if strings.HasPrefix(trimmed, "## ") && !strings.HasPrefix(trimmed, "### ") {
			break
		}
	}

	// Don't forget the last phase
	if currentPhase != nil {
		currentPhase.Complete = isPhaseComplete(currentPhase)
		phases = append(phases, *currentPhase)
	}

	return phases
}

// isPhaseComplete returns true if all tasks in the phase are complete.
func isPhaseComplete(p *Phase) bool {
	if len(p.Tasks) == 0 {
		return false
	}
	for _, task := range p.Tasks {
		if !task.Complete {
			return false
		}
	}
	return true
}

// getCurrentPhase returns the first incomplete phase, or nil if all complete.
func getCurrentPhase(phases []Phase) *Phase {
	for i := range phases {
		if !phases[i].Complete {
			return &phases[i]
		}
	}
	return nil
}

// formatPhaseForContext formats a phase for the context output.
func formatPhaseForContext(phase *Phase, phaseNum int, totalPhases int) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("### Current Phase: %d of %d - %s\n\n", phaseNum, totalPhases, phase.Name))

	if phase.Goal != "" {
		buf.WriteString(fmt.Sprintf("**Goal**: %s\n\n", phase.Goal))
	}

	buf.WriteString("**Tasks** (use task_complete with the ID to mark done):\n")
	for _, task := range phase.Tasks {
		if task.Complete {
			buf.WriteString(fmt.Sprintf("- [x] `%s` %s\n", task.ID, task.Text))
		} else {
			buf.WriteString(fmt.Sprintf("- [ ] `%s` %s\n", task.ID, task.Text))
		}
	}

	if phase.Milestone != "" {
		buf.WriteString(fmt.Sprintf("\n**Milestone**: %s\n", phase.Milestone))
	}

	return buf.String()
}

// buildAffectedFilesSection creates a section with affected file contents.
func buildAffectedFilesSection(files []string, maxLines int) string {
	var buf strings.Builder
	buf.WriteString("# Affected Files\n\n")

	foundAny := false
	for _, filePath := range files {
		// Try to read the file
		content, truncated, err := readAffectedFileContent(filePath, maxLines)
		if err != nil {
			if os.IsNotExist(err) {
				buf.WriteString(fmt.Sprintf("## %s\n\n(file not found)\n\n", filePath))
			} else {
				buf.WriteString(fmt.Sprintf("## %s\n\n(error reading: %v)\n\n", filePath, err))
			}
			continue
		}

		foundAny = true
		buf.WriteString(fmt.Sprintf("## %s\n\n", filePath))

		// Determine language for code fence
		ext := filepath.Ext(filePath)
		lang := ""
		switch ext {
		case ".go":
			lang = "go"
		case ".js":
			lang = "javascript"
		case ".ts":
			lang = "typescript"
		case ".py":
			lang = "python"
		case ".rs":
			lang = "rust"
		case ".rb":
			lang = "ruby"
		case ".java":
			lang = "java"
		case ".c", ".h":
			lang = "c"
		case ".cpp", ".hpp", ".cc":
			lang = "cpp"
		case ".md":
			lang = "markdown"
		case ".yaml", ".yml":
			lang = "yaml"
		case ".json":
			lang = "json"
		case ".sh", ".bash":
			lang = "bash"
		}

		buf.WriteString(fmt.Sprintf("```%s\n", lang))
		buf.WriteString(content)
		if !strings.HasSuffix(content, "\n") {
			buf.WriteString("\n")
		}
		buf.WriteString("```\n")

		if truncated {
			buf.WriteString(fmt.Sprintf("\n(truncated to %d lines)\n", maxLines))
		}
		buf.WriteString("\n")
	}

	if !foundAny {
		return ""
	}

	return buf.String()
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

func registerTaskCompleteTool(s *server.MCPServer) {
	tool := mcp.NewTool("task_complete",
		mcp.WithDescription("Mark a task as complete in the active proposal's implementation.md using its task ID (e.g., '1.1', '2.3'). If git.auto_commit is enabled, automatically commits all changes."),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("The task ID to mark as complete (e.g., '1.1' for phase 1, task 1)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		taskID, ok := request.Params.Arguments["id"].(string)
		if !ok {
			return mcp.NewToolResultError("id parameter must be a string"), nil
		}
		taskID = strings.TrimSpace(taskID)

		specPath, err := checkSpecWorkspace()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		slug, proposalPath, err := getPrimaryProposal(specPath)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if slug == "" {
			return mcp.NewToolResultError("No active proposal"), nil
		}

		implPath := filepath.Join(proposalPath, "implementation.md")
		content, err := os.ReadFile(implPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to read implementation.md: %v", err)), nil
		}

		// Parse phases to find the task by ID
		phases := extractPhases(string(content))
		var targetTask *Task
		for _, phase := range phases {
			for i := range phase.Tasks {
				if phase.Tasks[i].ID == taskID {
					targetTask = &phase.Tasks[i]
					break
				}
			}
			if targetTask != nil {
				break
			}
		}

		if targetTask == nil {
			return mcp.NewToolResultError(fmt.Sprintf("Task ID not found: %s", taskID)), nil
		}

		if targetTask.Complete {
			return mcp.NewToolResultError(fmt.Sprintf("Task %s is already complete", taskID)), nil
		}

		// Check if git auto-commit is enabled
		config := loadConfigOrDefault(specPath)
		if config.Git.AutoCommit {
			// Create git snapshot manager
			gitMgr := NewGitSnapshotManager(specPath, slug, taskID)

			// Commit all changes for this task
			if err := gitMgr.CommitChanges(targetTask.Text); err != nil {
				// Log warning but don't fail the task completion
				fmt.Fprintf(os.Stderr, "Warning: failed to commit changes: %v\n", err)
			}
		}

		// Replace the task at the specific line
		lines := strings.Split(string(content), "\n")
		lineIdx := targetTask.Line - 1 // Convert to 0-indexed

		if lineIdx < 0 || lineIdx >= len(lines) {
			return mcp.NewToolResultError("Internal error: invalid line number"), nil
		}

		line := lines[lineIdx]
		// Preserve indentation and replace checkbox
		indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
		lines[lineIdx] = indent + "- [x] " + targetTask.Text

		// Write back
		newContent := strings.Join(lines, "\n")
		if err := os.WriteFile(implPath, []byte(newContent), 0644); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to write implementation.md: %v", err)), nil
		}

		// Get updated progress
		total, completed := getProposalProgress(proposalPath)
		updatedPhases := extractPhases(newContent)
		currentPhase := getCurrentPhase(updatedPhases)

		var result strings.Builder
		result.WriteString(fmt.Sprintf("Task %s marked complete: %s\n\n", taskID, targetTask.Text))

		if config.Git.AutoCommit {
			result.WriteString("✓ Changes committed to git\n\n")
		}

		result.WriteString(fmt.Sprintf("Progress: %d/%d tasks complete\n", completed, total))

		if currentPhase == nil {
			result.WriteString("\nAll phases complete!")
		} else {
			// Count remaining tasks in current phase
			remaining := 0
			for _, t := range currentPhase.Tasks {
				if !t.Complete {
					remaining++
				}
			}
			result.WriteString(fmt.Sprintf("Current phase: %s (%d tasks remaining)", currentPhase.Name, remaining))
		}

		return mcp.NewToolResultText(result.String()), nil
	})
}

func registerTaskSnapshotTool(s *server.MCPServer) {
	tool := mcp.NewTool("task_snapshot",
		mcp.WithDescription("Create a git snapshot before starting work on a task. This captures the current state before making changes."),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("The task ID that you're about to start working on (e.g., '1.1', '2.3')"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		taskID, ok := request.Params.Arguments["id"].(string)
		if !ok {
			return mcp.NewToolResultError("id parameter must be a string"), nil
		}
		taskID = strings.TrimSpace(taskID)

		specPath, err := checkSpecWorkspace()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Check if git auto-snapshot is enabled
		config := loadConfigOrDefault(specPath)
		if !config.Git.AutoSnapshot {
			return mcp.NewToolResultText(fmt.Sprintf("Git auto-snapshot is disabled. Task %s ready to start.\n\nTo enable: Set git.auto_snapshot: true in spec/nocturnal.yaml", taskID)), nil
		}

		slug, _, err := getPrimaryProposal(specPath)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if slug == "" {
			return mcp.NewToolResultError("No active proposal"), nil
		}

		// Create git snapshot
		gitMgr := NewGitSnapshotManager(specPath, slug, taskID)
		if err := gitMgr.CreateSnapshot(); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create snapshot: %v", err)), nil
		}

		// Save snapshot reference in state
		state, err := loadState(specPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load state: %v", err)), nil
		}

		snapshotRef := gitMgr.GetSnapshotRef()
		if snapshotRef != "" {
			state.GitSnapshots[taskID] = GitSnapshotState{
				SnapshotRef: snapshotRef,
				TaskID:      taskID,
				Timestamp:   time.Now().Format(time.RFC3339),
			}

			if err := saveState(specPath, state); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to save state: %v", err)), nil
			}
		}

		var result strings.Builder
		result.WriteString(fmt.Sprintf("Git snapshot created for task %s\n\n", taskID))
		if snapshotRef != "" {
			result.WriteString(fmt.Sprintf("Snapshot ref: %s\n", snapshotRef[:8]))
		} else {
			result.WriteString("No uncommitted changes to snapshot (working directory clean)\n")
		}
		result.WriteString("\nYou can now proceed with implementing the task. When complete, use task_complete to commit all changes.")

		return mcp.NewToolResultText(result.String()), nil
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
		mcp.WithPromptDescription("Methodical implementation with investigation, planning, testing, and validation phases"),
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

		promptText := fmt.Sprintf(`%sYou will implement the active proposal using a methodical, fail-fast approach with multiple validation checkpoints.

## Philosophy

- **Fail fast**: If any phase fails, STOP and ask the user for guidance before proceeding
- **Ask questions**: When uncertain, ask the user for clarification rather than guessing
- **Quality over speed**: Each phase must complete successfully before moving to the next
- **Subagent isolation**: Each phase runs as a separate subagent for clean separation of concerns

## Setup

1) Call the MCP tool "context" to get the specification and design.
2) Call the MCP tool "tasks" to get the current phase tasks.
3) If there's an integrity warning about modified proposal files, STOP and ask the user to confirm.
4) Update your internal todo/task list with the tasks from the current phase.

## Implementation Lifecycle

For EACH task, run through these phases sequentially. If any phase fails, STOP and ask the user for guidance.

### Phase 1: Investigation (Subagent)

Spawn a subagent with this task:

---
INVESTIGATION TASK:

You are an investigation agent. Analyze the codebase and create an implementation plan.

**Task to implement:** [Task description from tasks tool]
**Specification:** [Include relevant spec sections]
**Design:** [Include relevant design sections]

**Instructions:**
1. Identify all files that need to be modified or created
2. Identify dependencies and potential conflicts
3. List any ambiguities or questions about the requirements
4. Create a step-by-step implementation plan

**Output format:**
INVESTIGATION REPORT
====================
Files to modify: [list]
Files to create: [list]
Dependencies: [list]
Questions/Ambiguities: [list - if any, mark as BLOCKING or NON-BLOCKING]
Implementation plan:
1. [step]
2. [step]
...
---

If the investigation report contains BLOCKING questions, STOP and ask the user to resolve them.

### Phase 2: Test Planning (Subagent)

Spawn a subagent with this task:

---
TEST PLANNING TASK:

You are a test planning agent. Define the testing requirements BEFORE implementation.

**Task to implement:** [Task description]
**Implementation plan:** [From Phase 1]
**Specification requirements:** [Relevant MUST/SHOULD/MAY items]

**Instructions:**
1. Define acceptance criteria based on the specification
2. List specific test cases that must pass
3. Identify edge cases and error conditions
4. Specify any integration test requirements

**Output format:**
TEST PLAN
=========
Acceptance criteria:
- [ ] [criterion 1]
- [ ] [criterion 2]
...

Test cases:
- [ ] [test case 1]: [expected behavior]
- [ ] [test case 2]: [expected behavior]
...

Edge cases:
- [ ] [edge case 1]: [expected behavior]
...

Integration requirements:
- [any cross-component testing needs]
---

Review the test plan. If it seems incomplete or incorrect, ask the user for input.

### Phase 3: Implementation (Subagent)

Spawn a subagent with this task:

---
IMPLEMENTATION TASK:

You are an implementation agent. Write the code changes.

**Task to implement:** [Task description]
**Implementation plan:** [From Phase 1]
**Test plan:** [From Phase 2]

**Instructions:**
1. Follow the implementation plan step by step
2. Keep changes minimal and focused on the task
3. Do not introduce unrelated refactors
4. Add appropriate comments for complex logic
5. If you need third-party API details, use docs_search

**Output format:**
IMPLEMENTATION REPORT
=====================
Files modified: [list with brief description of changes]
Files created: [list with brief description]
Notes: [any implementation decisions or deviations from plan]
---

### Phase 4: Validation (Subagent)

Spawn a subagent with this task:

---
VALIDATION TASK:

You are a validation agent. Verify the implementation against the test plan and specification.

**Task implemented:** [Task description]
**Test plan:** [From Phase 2]
**Specification requirements:** [Relevant sections]

**Instructions:**
1. Review the code changes
2. Verify each acceptance criterion is met
3. Check each test case can pass
4. Verify edge cases are handled
5. Check for specification compliance

**Output format:**
VALIDATION REPORT
=================
Status: PASS | FAIL

Acceptance criteria:
- [x] or [ ] [criterion]: [explanation]
...

Test case coverage:
- [x] or [ ] [test case]: [explanation]
...

Specification compliance:
- [x] or [ ] [requirement]: [explanation]
...

Issues found (if FAIL):
- [issue description and location]
...
---

If validation FAILS, STOP and ask the user whether to:
a) Fix the issues and re-run from Phase 3
b) Adjust the requirements/plan
c) Skip this task and document it as incomplete

### Phase 5: Testing (Subagent)

Spawn a subagent with this task:

---
TESTING TASK:

You are a testing agent. Execute all available tests.

**Instructions:**
1. Run the project's unit tests
2. Run any linters or formatters
3. Run integration tests if available
4. Report all results

**Output format:**
TEST RESULTS
============
Unit tests: PASS | FAIL
  [summary or failure details]

Linting: PASS | FAIL
  [summary or issues]

Integration tests: PASS | FAIL | N/A
  [summary or failure details]

Overall: PASS | FAIL
---

If tests FAIL, STOP and ask the user for guidance.

## After Each Task

1. If all phases PASS: Call task_complete with the task ID
2. Update your internal todo/task list
3. Call tasks to get the next task
4. Repeat the lifecycle for the next task

## Completion

When all tasks are complete:
1. Summarize all changes made
2. List which specification requirements are satisfied
3. Note any deferred items or known limitations
`, goalStr)

		return &mcp.GetPromptResult{
			Description: "Methodical implementation with investigation, planning, testing, and validation phases",
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

func registerLazyPrompt(s *server.MCPServer) {
	prompt := mcp.NewPrompt("lazy",
		mcp.WithPromptDescription("Fast autonomous implementation - implements quickly, moves on from blockers"),
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

		promptText := fmt.Sprintf(`%sYou will implement the active proposal as quickly as possible, moving past blockers.

## Philosophy

- **Speed over perfection**: Get something working, then iterate
- **Move past blockers**: If a task is hard, partially complete it and move on
- **Autonomous**: Don't ask for user input unless completely stuck
- **Document incompleteness**: Always note what was skipped or partially done

## Setup

1) Call the MCP tool "context" to get the specification and design.
2) Call the MCP tool "tasks" to get the current phase tasks.
3) If there's an integrity warning, STOP and ask the user to confirm.
4) Update your internal todo/task list with tasks from the current phase.

## Implementation Loop

REPEAT for each task until all phases are complete:

### Step 1: Implement Immediately

- Read the task description
- Implement it directly without extensive planning
- If you need third-party API details, use docs_search
- Time-box yourself: if stuck for more than a few attempts, go to Step 2

### Step 2: Handle Blockers

If a task is difficult or blocked:
1. Implement what you CAN do
2. Add a code comment: // TODO: [what remains to be done]
3. Note the partial completion in your summary
4. Mark the task complete anyway and move on

### Step 3: Quick Validation (Subagent)

After completing each task, spawn a quick validation subagent:

---
QUICK VALIDATION:

Task: [task description]
Changes made: [brief summary]

Check:
1. Does the code compile/parse without errors?
2. Are there obvious bugs or missing pieces?
3. Does it roughly satisfy the task intent?

Reply: GOOD (proceed) | ISSUES: [brief list]
---

- If GOOD: Continue to next task
- If ISSUES: Make a quick fix attempt, then move on regardless

### Step 4: Mark Complete

- Call task_complete with the task ID
- Update your internal todo/task list
- Call tasks to get the next task or phase

## End of Phase Testing (Subagent)

After completing ALL tasks in a phase, spawn a test subagent:

---
TEST TASK:

Run all project tests:
1. Unit tests
2. Linting/formatting
3. Integration tests (if available)

Report: PASS | FAIL with summary
---

- If PASS: Proceed to next phase
- If FAIL: Make ONE quick fix attempt, then proceed anyway and note the failures

## Completion

When all phases are complete:
1. List all changes made
2. List any partial completions or skipped items
3. List test failures if any
4. Ask user to review and decide on follow-up

## Important Notes

- Do NOT get stuck on any single task - time-box and move on
- Do NOT ask for user input during the loop
- ALWAYS document what was skipped or partially done
- Prefer working code with TODOs over perfect but incomplete implementation
- Update your internal todo/task list after each task
`, goalStr)

		return &mcp.GetPromptResult{
			Description: "Fast autonomous implementation - implements quickly, moves on from blockers",
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

func registerMaintenanceListTool(s *server.MCPServer) {
	tool := mcp.NewTool("maintenance_list",
		mcp.WithDescription("List all maintenance items with due/total requirement counts."),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		specPath, err := checkSpecWorkspace()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		slugs, err := listMaintenanceFiles(specPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list maintenance items: %v", err)), nil
		}

		if len(slugs) == 0 {
			return mcp.NewToolResultText("No maintenance items found"), nil
		}

		state, err := loadState(specPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load state: %v", err)), nil
		}

		var result strings.Builder
		result.WriteString(fmt.Sprintf("# Maintenance Items (%d)\n\n", len(slugs)))

		for _, slug := range slugs {
			filePath := filepath.Join(specPath, maintenanceDir, slug+".md")
			reqs, err := parseMaintenanceFile(filePath, state, slug)
			if err != nil {
				result.WriteString(fmt.Sprintf("- %s: error parsing (%v)\n", slug, err))
				continue
			}

			dueCount := 0
			for _, req := range reqs {
				if req.Due {
					dueCount++
				}
			}

			result.WriteString(fmt.Sprintf("- **%s**: %d/%d due\n", slug, dueCount, len(reqs)))
		}

		return mcp.NewToolResultText(result.String()), nil
	})
}

func registerMaintenanceContextTool(s *server.MCPServer) {
	tool := mcp.NewTool("maintenance_context",
		mcp.WithDescription("Get requirements for a maintenance item, showing which are currently due."),
		mcp.WithString("slug",
			mcp.Required(),
			mcp.Description("Maintenance item slug"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slug, ok := request.Params.Arguments["slug"].(string)
		if !ok {
			return mcp.NewToolResultError("slug parameter must be a string"), nil
		}

		specPath, err := checkSpecWorkspace()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		filePath := filepath.Join(specPath, maintenanceDir, slug+".md")
		if !fileExists(filePath) {
			return mcp.NewToolResultError(fmt.Sprintf("Maintenance item '%s' does not exist", slug)), nil
		}

		state, err := loadState(specPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load state: %v", err)), nil
		}

		reqs, err := parseMaintenanceFile(filePath, state, slug)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to parse maintenance file: %v", err)), nil
		}

		var result strings.Builder
		result.WriteString(fmt.Sprintf("# Maintenance: %s\n\n", slug))

		// Separate due and not-due requirements
		var dueReqs, notDueReqs []MaintenanceRequirement
		for _, req := range reqs {
			if req.Due {
				dueReqs = append(dueReqs, req)
			} else {
				notDueReqs = append(notDueReqs, req)
			}
		}

		// Due requirements (these should be actioned)
		result.WriteString(fmt.Sprintf("## Due Requirements (%d)\n\n", len(dueReqs)))
		if len(dueReqs) == 0 {
			result.WriteString("No requirements are currently due.\n\n")
		} else {
			result.WriteString("These requirements should be addressed:\n\n")
			for _, req := range dueReqs {
				result.WriteString(fmt.Sprintf("- **[%s]** %s\n", req.ID, req.Text))
				if req.Freq != "" {
					result.WriteString(fmt.Sprintf("  - Frequency: %s\n", req.Freq))
				}
				if req.LastActioned != "" {
					result.WriteString(fmt.Sprintf("  - Last actioned: %s\n", req.LastActioned))
				}
				result.WriteString("\n")
			}
		}

		// Not due requirements (informational only)
		if len(notDueReqs) > 0 {
			result.WriteString(fmt.Sprintf("## Not Due Yet (%d)\n\n", len(notDueReqs)))
			result.WriteString("These requirements do not need action right now:\n\n")
			for _, req := range notDueReqs {
				result.WriteString(fmt.Sprintf("- **[%s]** %s", req.ID, req.Text))
				if req.Freq != "" {
					result.WriteString(fmt.Sprintf(" (freq: %s)", req.Freq))
				}
				if req.LastActioned != "" {
					result.WriteString(fmt.Sprintf(" - last: %s", req.LastActioned))
				}
				result.WriteString("\n")
			}
			result.WriteString("\n")
		}

		result.WriteString("## Instructions\n\n")
		result.WriteString("For each due requirement you action, call `maintenance_actioned` with the slug and requirement ID.\n")

		return mcp.NewToolResultText(result.String()), nil
	})
}

func registerMaintenanceActionedTool(s *server.MCPServer) {
	tool := mcp.NewTool("maintenance_actioned",
		mcp.WithDescription("Mark a requirement as actioned (records current timestamp)."),
		mcp.WithString("slug",
			mcp.Required(),
			mcp.Description("Maintenance item slug"),
		),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Requirement ID to mark as actioned"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slug, ok := request.Params.Arguments["slug"].(string)
		if !ok {
			return mcp.NewToolResultError("slug parameter must be a string"), nil
		}

		id, ok := request.Params.Arguments["id"].(string)
		if !ok {
			return mcp.NewToolResultError("id parameter must be a string"), nil
		}

		specPath, err := checkSpecWorkspace()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		filePath := filepath.Join(specPath, maintenanceDir, slug+".md")
		if !fileExists(filePath) {
			return mcp.NewToolResultError(fmt.Sprintf("Maintenance item '%s' does not exist", slug)), nil
		}

		state, err := loadState(specPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to load state: %v", err)), nil
		}

		// Parse file to validate ID exists
		reqs, err := parseMaintenanceFile(filePath, state, slug)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to parse maintenance file: %v", err)), nil
		}

		found := false
		var reqText string
		for _, req := range reqs {
			if req.ID == id {
				found = true
				reqText = req.Text
				break
			}
		}

		if !found {
			return mcp.NewToolResultError(fmt.Sprintf("Requirement ID '%s' not found in maintenance item '%s'", id, slug)), nil
		}

		// Update state
		if state.Maintenance == nil {
			state.Maintenance = make(map[string]map[string]MaintenanceState)
		}
		if state.Maintenance[slug] == nil {
			state.Maintenance[slug] = make(map[string]MaintenanceState)
		}

		timestamp := time.Now().Format(time.RFC3339)
		state.Maintenance[slug][id] = MaintenanceState{
			LastActioned: timestamp,
		}

		if err := saveState(specPath, state); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to save state: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Marked '%s' as actioned at %s\n\n%s", id, timestamp, reqText)), nil
	})
}

func registerStartMaintenancePrompt(s *server.MCPServer) {
	prompt := mcp.NewPrompt("start-maintenance",
		mcp.WithPromptDescription("Execute maintenance requirements for a maintenance item"),
		mcp.WithArgument("slug",
			mcp.ArgumentDescription("Maintenance item slug"),
			mcp.RequiredArgument(),
		),
	)

	s.AddPrompt(prompt, func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		slug := strings.TrimSpace(request.Params.Arguments["slug"])

		promptText := fmt.Sprintf(`You will execute maintenance requirements for: %s

## Setup

1. Call the MCP tool "maintenance_context" with slug="%s" to get the requirements.
2. Focus ONLY on the "Due Requirements" section. Ignore "Not Due Yet" requirements.
3. Review your internal todo/task list and add the due requirements.

## Execution Flow

For EACH due requirement:

### Step 1: Understand the Requirement
- Read the requirement text carefully.
- If it references commands, files, or processes, locate them first.
- If unclear, ask the user for clarification.

### Step 2: Execute the Requirement
- Perform the required action (run tests, update deps, review files, etc.).
- Use appropriate tools: bash for commands, read/edit for files, docs_search for APIs.
- Document what you did.

### Step 3: Verify Success
- Confirm the action completed successfully.
- If it failed, document the failure and ask the user how to proceed.
- Do NOT mark as actioned if it failed.

### Step 4: Mark as Actioned
- If the requirement was successfully completed, call:
  maintenance_actioned(slug="%s", id="<requirement-id>")
- This records the timestamp so the requirement won't be due again until its frequency interval elapses.

## Important Notes

- Do NOT action requirements that are "Not Due Yet" - only focus on "Due Requirements".
- If a requirement has no frequency tag, it's always due and will appear every time until completed.
- Ask the user if you're unsure about any requirement's intent.
- After actioning all due requirements, summarize what was done.

## Completion

When all due requirements are actioned:
1. List which requirements were completed.
2. Note any requirements that failed or were skipped.
3. Call maintenance_context again to confirm no requirements are still due.
`, slug, slug, slug)

		return &mcp.GetPromptResult{
			Description: "Execute maintenance requirements",
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
