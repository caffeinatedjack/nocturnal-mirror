```
 ______   ______   ______ _______  _    _   ______   ______   ______   _
| |  \ \ / |  | \ | |       | |   | |  | | | |  | \ | |  \ \ | |  | | | |
| |  | | | |  | | | |       | |   | |  | | | |__| | | |  | | | |__| | | |   _
|_|  |_| \_|__|_/ |_|____   |_|   \_|__|_| |_|  \_\ |_|  |_| |_|  |_| |_|__|_|
```

A CLI tool for agent-assisted coding with spec-driven development and documentation management.

## Features

### Specification Management

Create and manage structured specifications for your project. Initialize a workspace, then use proposals to develop new features through a defined lifecycle:

- **Proposals** - Draft changes with specification, design, and implementation documents
- **Validation** - Check proposals against documentation guidelines before completion
- **Promotion** - Complete proposals to archive designs and promote specs to the main section

```bash
nocturnal spec init                      # Initialize workspace
nocturnal spec proposal add my-feature   # Create a new proposal
nocturnal spec proposal activate my-feature
nocturnal spec proposal validate my-feature
nocturnal spec proposal complete my-feature
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
nocturnal agent todoread  # Read TODO.md
nocturnal agent todowrite # Write structured tasks to TODO.md
```

### MCP Server

Expose agent tools via the Model Context Protocol for integration with AI assistants:

```bash
nocturnal mcp
```

Available tools:
- `todoread` / `todowrite` - Task management
- `current` - Active proposal access
- `docs_list` / `docs_search` - Documentation lookup from `~/.docs/`

### Documentation Management

Store and search API/library documentation for agent reference:

```bash
nocturnal agent docs list
nocturnal agent docs search <query>
```

## Installation

```bash
make build
make install  # Installs to ~/.local/bin
```

## License

MIT