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
- `docs_list` / `docs_search` - Documentation lookup 

### Documentation Management

Store and search API/library documentation for agent reference:

```bash
nocturnal agent docs list
nocturnal agent docs search <query>
```

The documentation manager expects documentation to be inside `~/.docs/` folder. Each library should have its own markdown file and inside this file each component should be seperated with `---`. This allows the lookup tool to search and get whole components and send it back to the agent. 

I recomend generating these docs with AI using a prompt like this: 
```
You’ll write a condensed version of the documentation to ~/.docs. If there are any key references missing, fetch them from the web as well. The goal is to develop a solid understanding of the library. These docs are intended to provide an AI agent with a clear overview of the library or technology, including its usage and where to find additional information. Be as concise as possible to avoid overwhelming the AI's context.
Separate each logical section with \n---\n, and immediately after the separator, include a header marked with #. Whenever possible, include direct links to the relevant documentation alongside any components or classes.
```
Make sure you include individual links to every component for good reliability. This tool is really useful if you are working with a later version then what the AI is trained on. 

## Installation

```bash
make build
make install  # Installs to ~/.local/bin
```

## License

MIT

## Why I made this?

Like a lot of developers, I’ve been experimenting with AI in development, trying to keep things consistent, reduce context switching, and improve quality. I’ve tried out a few spec-driven tools like Speckit, but honestly, I think specs should still be something people drive, not the AI. That’s why I built this tool. You can use AI to help with writing specs if you want, but the real creation process is still up to you.

I also added a persistent to-do manager and a documentation tool, since they fit well with how this tool works.