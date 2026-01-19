# Agent Instructions

This repository contains **nocturnal**, a Go CLI tool for specification-driven development with AI agent integration via MCP (Model Context Protocol).

Nocturnal manages specifications, proposals, rules, maintenance tasks, and third-party documentation in a structured workspace. It exposes project context to AI agents through an MCP server.

## Quick Reference

```bash
# Build & Run
make build                    # Build binary
make install                  # Build and install to ~/.local/bin
go run . <command>            # Run without building

# Testing & Quality
make test                     # Run all tests with race detection
go test -v ./...              # Run all tests
go test -v -run TestName ./.. # Run a single test by name

# Code Quality
make fmt                      # Format code (go fmt)
make lint                     # Run linter (go vet)
make deps                     # Download and tidy dependencies
```

## Project Structure

```
nocturnal/
├── main.go              # Entry point, sets version/build vars
├── cmd/                 # CLI commands (Cobra-based)
│   ├── root.go          # Root command and shell completion
│   ├── agent.go         # Agent context commands (current, project, specs)
│   ├── docs.go          # Third-party documentation management
│   ├── mcp.go           # MCP server with tools and prompts
│   ├── spec.go          # Specification and proposal commands
│   ├── maintenance.go   # Recurring maintenance task management
│   ├── stats.go         # Project statistics and metrics
│   ├── graph.go         # Proposal dependency graph visualization
│   ├── config.go        # Configuration management
│   ├── state.go         # State persistence (active proposals, etc.)
│   ├── git.go           # Git snapshot and commit management
│   ├── ui.go            # Terminal output styling
│   ├── util.go          # Helper functions
│   └── templates/       # Embedded templates and help text
├── docs/                # Project documentation
└── spec/                # Workspace (created per-project via `nocturnal spec init`)
    ├── proposal/        # Active proposals
    ├── section/         # Completed specifications
    ├── archive/         # Archived design/implementation docs
    ├── rule/            # Project-wide rules
    ├── maintenance/     # Recurring maintenance items
    ├── third/           # Third-party library documentation
    ├── project.md       # Project design overview
    ├── nocturnal.yaml   # Configuration file
    └── .nocturnal.json  # State file (active proposals, hashes, etc.)
```

## Code Style Guidelines

### Imports

Order imports in three groups, separated by blank lines:
1. Standard library
2. External dependencies
3. Internal packages

```go
import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)
```

### Formatting

- Use `go fmt` (via `make fmt`) before committing
- Use tabs for indentation (Go standard)
- Line length: no strict limit, but prefer readable lines (~100 chars)

### Naming Conventions

| Element      | Convention                     | Example                      |
|--------------|--------------------------------|------------------------------|
| Packages     | lowercase, short               | `cmd`, `ui`                  |
| Files        | lowercase, underscores ok      | `format.go`, `agent.go`      |
| Functions    | PascalCase (exported)          | `Execute()`, `Success()`     |
| Functions    | camelCase (unexported)         | `runTodoWrite()`, `copyFile` |
| Variables    | camelCase                      | `specPath`, `proposalPath`   |
| Constants    | camelCase (unexported)         | `specDir`, `ruleDir`         |
| Structs      | PascalCase                     | `DocComponent`               |
| Interfaces   | PascalCase, -er suffix if verb | `Reader`, `Formatter`        |

### Error Handling

- Always check errors immediately after function calls
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Use `os.IsNotExist(err)` for file existence checks
- Return early on errors (avoid deep nesting)

```go
// Good
content, err := os.ReadFile(path)
if err != nil {
    if os.IsNotExist(err) {
        return "", fmt.Errorf("file not found: %s", path)
    }
    return "", fmt.Errorf("failed to read file: %w", err)
}

// For CLI commands, use ui.Error() and return
if err != nil {
    ui.Error(fmt.Sprintf("Failed to read: %v", err))
    return
}
```

### Struct Definitions

- Use JSON tags for serializable structs
- Include inline comments for field documentation

```go
type TodoItem struct {
    ID       string     `json:"id"`
    Content  string     `json:"content"`
    Status   string     `json:"status"`   // "pending", "in_progress", "completed", "cancelled"
    Priority string     `json:"priority"` // "high", "medium", "low"
    Children []TodoItem `json:"children,omitempty"`
}
```

### CLI Commands (Cobra)

Each command should have:
- `Use`: command name and argument placeholders
- `Short`: one-line description
- `Long`: detailed description with examples
- `Run` or `RunE`: handler function
- Optional: `Args`, `ValidArgsFunction` for completion

```go
var myCmd = &cobra.Command{
    Use:   "mycommand <arg>",
    Short: "Brief description",
    Long: `Detailed description.

Examples:
    nocturnal mycommand example`,
    Args: cobra.ExactArgs(1),
    Run:  runMyCommand,
}
```

### UI Output

Use the cmd package's output helpers for consistent terminal output:

```go
printSuccess("Operation completed")
printError("Something went wrong")
printWarning("Be careful")
printInfo("FYI")
printDim("Secondary info")
```

## Working with Nocturnal Specifications

This project uses its own specification management. When working on proposals:

1. Read project rules: `nocturnal agent project`
2. Check active proposal: `nocturnal agent current`
3. View completed specs: `nocturnal agent specs`

### Proposal Workflow

```bash
nocturnal spec proposal add my-feature     # Create proposal
nocturnal spec proposal activate my-feature # Set as active
# ... implement the feature ...
nocturnal spec proposal complete my-feature # Archive and promote
```

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `github.com/mark3labs/mcp-go` - MCP server protocol

## Testing Guidelines

- Test files: `*_test.go` in the same package
- Run single test: `go test -v -run TestFunctionName ./...`
- Use table-driven tests for multiple cases
- Use `t.Helper()` in test helper functions

## Key Implementation Notes

- Templates are embedded using `//go:embed` directive in `spec.go`
- MCP server runs via stdio (standard input/output)
- Specification workspace is created in `./spec/`
- Documentation components are read from `spec/third/`
- Active proposal state is stored in `spec/.nocturnal.json`

## Common Tasks

### Adding a New Command

1. Create handler function: `func runMyCmd(cmd *cobra.Command, args []string) {...}`
2. Define cobra.Command with Use, Short, Long, Run
3. Register in `init()`: `parentCmd.AddCommand(myCmd)`

### Adding an MCP Tool

1. Create `registerMyTool(s *server.MCPServer)` function in `mcp.go`
2. Define tool with `mcp.NewTool()` and handler
3. Register in `runMCP()`: `registerMyTool(s)`

### Adding an MCP Prompt

1. Create `registerMyPrompt(s *server.MCPServer)` function in `mcp.go`
2. Define prompt with `mcp.NewPrompt()` and handler
3. Register in `runMCP()`: `registerMyPrompt(s)`
4. Update `cmd/templates/help/mcp.txt` to document the new prompt

Example:
```go
func registerMyPrompt(s *server.MCPServer) {
    prompt := mcp.NewPrompt("my-prompt",
        mcp.WithPromptDescription("Brief description of what this prompt does"),
        mcp.WithArgument("arg1",
            mcp.ArgumentDescription("Description of argument"),
            mcp.RequiredArgument(), // Optional
        ),
    )

    s.AddPrompt(prompt, func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
        promptText := `Instructions for the AI agent...`

        return &mcp.GetPromptResult{
            Description: "Brief description",
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
```
