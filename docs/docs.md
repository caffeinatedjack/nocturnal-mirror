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

## MCP Prompt

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


