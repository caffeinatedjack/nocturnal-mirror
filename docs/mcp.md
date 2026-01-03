# MCP Server

The Model Context Protocol (MCP) server exposes Nocturnal's agent tools to AI assistants like Claude, OpenCode, and other MCP-compatible clients. This allows AI assistants to directly access your project's specifications, rules, and documentation.

## Overview

The MCP server runs via standard input/output (stdio), making it easy to integrate with AI assistant configurations. It provides three tools and two prompts for accessing project context.

**Tools:**

- `context` - Rules/project design + active proposal docs + open tasks (single, correctness-oriented context payload)
- `docs_list` - List available documentation components
- `docs_search` - Search documentation components by name

**Prompts:**

- `start-implementation` - Bootstrap instructions for implementing using Nocturnal context
- `add-third-party-docs` - Generate condensed documentation for third-party libraries

## Configuration

### OpenCode

Add to `~/.opencode/config.json` or project `.opencode/config.json`:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "nocturnal": {
      "type": "local",
      "enabled": true,
      "command": ["nocturnal", "mcp"]
    }
  }
}
```

### Cursor

Add to Cursor's MCP settings (unverified):

```json
{
  "mcp": {
    "nocturnal": {
      "command": "nocturnal",
      "args": ["mcp"],
      "cwd": "/path/to/your/project"
    }
  }
}
```
