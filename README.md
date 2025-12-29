```
 ______   ______   ______ _______  _    _   ______   ______   ______   _
| |  \ \ / |  | \ | |       | |   | |  | | | |  | \ | |  \ \ | |  | | | |
| |  | | | |  | | | |       | |   | |  | | | |__| | | |  | | | |__| | | |   _
|_|  |_| \_|__|_/ |_|____   |_|   \_|__|_| |_|  \_\ |_|  |_| |_|  |_| |_|__|_|
```

A CLI tool for spec-driven development and agent tooling.

## Features

### Specification Management

Create and manage structured specifications for your project. Initialize a workspace, then use proposals to develop new features through a defined lifecycle:

- **Proposals** - Draft changes with specification, design, and implementation documents
- **Validation** - Check proposals against documentation guidelines before completion
- **Promotion** - Complete proposals to archive designs and promote specs to the main section

```bash
nocturnal spec init                         # Initialize workspace
nocturnal spec view                         # View workspace overview
nocturnal spec proposal add my-feature      # Create a new proposal
nocturnal spec proposal activate my-feature # Set as active proposal
nocturnal spec proposal validate my-feature # Validate against guidelines
nocturnal spec proposal complete my-feature # Archive and promote to specs
nocturnal spec proposal remove my-feature   # Remove a proposal
```

### Project Rules

Define project-wide rules that persist across proposals. Rules provide consistent constraints and guidelines for development.

```bash
nocturnal spec rule add naming-conventions
nocturnal spec rule show
```

### Agent Commands

Commands designed for AI coding agents to access project context:

```bash
nocturnal agent current   # Get the active proposal with all documents
nocturnal agent project   # Get project rules and design context
nocturnal agent specs     # Get completed specifications
nocturnal agent docs list           # List available documentation
nocturnal agent docs search <query> # Search documentation by name
```

### MCP Server

Expose agent tools via the Model Context Protocol for integration with AI assistants:

```bash
nocturnal mcp
```

Available tools:
- `rules` - Get project rules and design context
- `current` - Show active proposal (specification and design only)
- `tasks` - Get implementation tasks for active proposal
- `docs_list` / `docs_search` - Documentation lookup

### Documentation Management

Store and search API/library documentation for agent reference. The documentation manager expects documentation to be inside `~/.docs/` folder. Each library should have its own markdown file and inside this file each component should be separated with `---`. This allows the lookup tool to search and get whole components and send it back to the agent.

The MCP server exposes a prompt for you to use to generate these reference files.

## Installation

```bash
make build
make install
```
Alternatively you can download the executable from the artifacts.

### Installing MCP server in OpenCode

Add MCP to opencode
```json
{
    "$schema": "https://opencode.ai/config.json",
    "mcp" : {
        "nocturnal": {
            "type": "local",
            "enabled": true,
            "command": ["nocturnal", "mcp"]
        },
    }
}
```

## License

MIT

## Why I made this?

Like a lot of developers, I’ve been experimenting with AI in development, trying to keep things consistent, reduce context switching, and improve quality. I’ve tried out a few spec-driven tools like Speckit, but honestly, I think specs should still be something people drive, not the AI. That’s why I built this tool. You can use AI to help with writing specs if you want, but the real creation process is still up to you.

I also added a persistent to-do manager and a documentation tool, since they fit well with how this tool works.
