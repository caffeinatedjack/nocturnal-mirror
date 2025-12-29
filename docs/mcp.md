# MCP Server

The Model Context Protocol (MCP) server exposes Nocturnal's agent tools to AI assistants like Claude, OpenCode, and other MCP-compatible clients. This allows AI assistants to directly access your project's specifications, rules, and documentation.

## Overview

The MCP server runs via standard input/output (stdio), making it easy to integrate with AI assistant configurations. It provides five tools and one prompt for accessing project context.

**What it does:**

Exposes a few helper tools and prompt to make integrating with nocturnal easier.

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
