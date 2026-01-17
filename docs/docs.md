# Documentation Management

Nocturnal provides tools for managing third-party library and API documentation. This documentation is stored in `spec/third` and can be accessed by AI agents through the MCP server or CLI commands.

## Overview

One of the biggest problems I’ve had with AI agents is when they simply don’t have the right information in their context. To fix this, I added a few tools to the MCP.

The MCP server exposes tools the agent can call to pull library information that’s already stored locally. If there aren’t any docs to parse, there’s also a prompt you can trigger from the website that lets the agent scrape the info and save it locally.

This keeps the heavy lifting out of the prompt and avoids stuffing the context with a bunch of unnecessary data. You can grab library information _before_ you need it, separately from the rest of your development flow.

THis does mean you need to have the libraries in a specific format so the search functionality works, but I feel like this is a good trade off. 

## File Format

Documentation files are stored in `spec/third` as text or markdown files. Each file can contain multiple **components** separated by `---` dividers, with headers starting with `#`.

```markdown
---
# Component Name

This is the content for the component.
It can span multiple lines and include code examples.

---
# Another Component

More content here.
```

**Format rules:**

- Components are separated by `---` on its own line
- Each component starts with `# Component Name` header
- Component names are used for searching
- Content continues until the next `---` separator or end of file

## CLI Commands

### docs list

List all documentation components from all files.

```bash
nocturnal docs list
```

**Output:**

- Component names with their source file
- Preview of each component's content
- Total count of components found

**Example output:**
```
Found 5 component(s)

# cobra-basics
  from go-libs.md
  Command-line interface framework for Go...

# lipgloss-styles
  from go-libs.md
  Terminal styling library...
```

---

### docs search

Search documentation by component name.

```bash
nocturnal docs search <query>
```

**Arguments:**

- `<query>` - Search string to match against component names (case-insensitive)

**What it does:**

- Finds all components whose names contain the query
- Displays full content of matching components
- Shows source file for each match

**Example:**
```bash
nocturnal docs search "cobra"
```

**Output:**
```
Found 2 result(s)

# cobra-basics
  from go-libs.md

  Cobra is a library for creating CLI applications...
  [full content]

# cobra-flags
  from go-libs.md

  Adding flags to commands...
  [full content]
```

## MCP Tools

The MCP server exposes two documentation tools that can be called by AI agents.

### docs_list

List all available library and API documentation components.

**Parameters:** None

**Returns:** Formatted list of all documentation components with:
- Component count
- Component names
- Source files
- Content previews

**Example response:**
```
Found 3 component(s)

# mcp-go-server
  from mcp.md
  Server creation and configuration...

# mcp-go-tools
  from mcp.md
  Defining and registering tools...
```

---

### docs_search

Search library and API documentation by name. Returns full content of matching documentation.

**Parameters:**

| Name  | Type   | Required | Description                                   |
|-------|--------|----------|-----------------------------------------------|
| query | string | Yes      | Search query to match against component names |

**Returns:** Full content of all matching components, including:
- Match count
- Component names and source files
- Complete component content

**Error cases:**
- Returns error if query parameter is missing or not a string
- Returns message if no documentation found
- Returns message if no components match the query

## MCP Prompts

### add-third-party-docs

Generate condensed documentation for third-party libraries.

**Arguments:**

| Name | Type   | Required | Description                                           |
|------|--------|----------|-------------------------------------------------------|
| urls | string | Yes      | Comma-separated list of documentation URLs to process |

**What it does:**

Generates a prompt instructing the AI to:
1. Fetch documentation from the provided URLs
2. Create condensed, AI-friendly documentation
3. Save to `spec/third` directory
4. Structure content with `---` separators and `#` headers
5. Include links to original documentation

**Example usage:**

When using an MCP client, invoke the prompt with:
```
urls: "https://pkg.go.dev/github.com/spf13/cobra, https://cobra.dev/docs"
```

The AI will then create a condensed documentation file covering the library's key concepts, APIs, and usage patterns.

---

### start-implementation

Methodical, fail-fast implementation with multiple validation checkpoints.

**Arguments:**

| Name | Type   | Required | Description                                                |
|------|--------|----------|------------------------------------------------------------|
| goal | string | No       | Short description of what you want to implement (optional) |

**What it does:**

For each task, runs through 5 phases as separate subagents:

1. **Investigation** - Analyze codebase, create implementation plan, identify blocking questions
2. **Test Planning** - Define acceptance criteria and test cases BEFORE implementation
3. **Implementation** - Write code changes following the plan
4. **Validation** - Verify implementation against test plan and specification
5. **Testing** - Run unit tests, linters, and integration tests

**Philosophy:**
- Fail fast: If any phase fails, STOP and ask the user for guidance
- Ask questions: When uncertain, ask rather than guess
- Quality over speed: Each phase must pass before proceeding

---

### lazy

Fast, autonomous implementation that prioritizes speed over perfection.

**Arguments:**

| Name | Type   | Required | Description                                                |
|------|--------|----------|------------------------------------------------------------|
| goal | string | No       | Short description of what you want to implement (optional) |

**What it does:**

1. Implements tasks immediately without extensive planning
2. If blocked, partially completes the task, adds TODO comments, and moves on
3. Spawns quick validation subagent after each task
4. Runs tests at end of each phase but proceeds even if they fail
5. Documents all partial completions and skipped items

**Philosophy:**
- Speed over perfection: Get something working, then iterate
- Move past blockers: Don't get stuck on any single task
- Document incompleteness: Always note what was skipped or partially done

Autonomous implementation loop that uses the project spec workspace to drive task-by-task execution.

**Arguments:**

| Name | Type   | Required | Description                                                |
|------|--------|----------|------------------------------------------------------------|
| goal | string | No       | Short description of what you want to implement (optional) |

**What it does:**

Returns a prompt instructing the AI to:
1. Call `context` and treat it as the source of truth
2. Call `tasks` to get the first incomplete phase
3. Implement tasks one at a time
4. After each task, call `task_complete(id: "X.Y")` and update the agent's internal todo/task list
5. After finishing a phase, spawn a validation subagent; if validation fails, iterate
6. Loop until `tasks` reports "All phases complete!"

**Important behavior notes:**
- `tasks` only returns the **current phase**, not all phases.
- If `context` returns an integrity warning about changed proposal files, the agent should stop until the user confirms.
- The agent is reminded to keep its internal todo/task list updated (separate from the Nocturnal task tracking).

Autonomous implementation loop (“ralphing”) that uses the project spec workspace to drive task-by-task execution.

**Arguments:**

| Name | Type   | Required | Description                                                |
|------|--------|----------|------------------------------------------------------------|
| goal | string | No       | Short description of what you want to implement (optional) |

**What it does:**

Returns a prompt instructing the AI to:
1. Call `context` and treat it as the source of truth
2. Call `tasks` to get the first incomplete phase
3. Implement tasks one at a time
4. After each task, call `task_complete(id: "X.Y")`
5. After finishing a phase, spawn a validation subagent; if validation fails, iterate
6. Loop until `tasks` reports “All phases complete!”

**Important behavior notes:**
- `tasks` only returns the **current phase**, not all phases.
- If `context` returns an integrity warning about changed proposal files, the agent should stop until the user confirms.

