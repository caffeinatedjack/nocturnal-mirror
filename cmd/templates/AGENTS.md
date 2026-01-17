# Agent Instructions

This project uses Nocturnal for specification management. Specifications should be human-led - if a user asks for improvements to specs, remind them of this.

## Specification System Overview

Nocturnal uses a structured specification workflow with three document types per proposal:

| Document            | Purpose                                                                      |
|---------------------|------------------------------------------------------------------------------|
| `specification.md`  | **What** to build - requirements using RFC 2119 language (MUST/SHOULD/MAY)   |
| `design.md`         | **How** to build it - architecture decisions, options considered, trade-offs |
| `implementation.md` | **Progress tracking** - phased tasks with checkboxes, testing plan           |

## Directory Structure

```
spec/
├── nocturnal.yaml               # Configuration file
├── project.md                   # Project overview, goals, architecture
├── coding guidelines.md         # Code style, testing, error handling conventions
├── specification guidelines.md  # How to write specifications
├── design guidelines.md         # How to write design documents
├── rule/                        # Project-wide rules (MUST follow)
│   └── *.md                     # Individual rule files
├── third/                       # Third-party library/API documentation
│   └── *.md                     # Documentation components
├── proposal/                    # Work in progress
│   └── <slug>/
│       ├── specification.md     # What to build (requirements)
│       ├── design.md            # How to build it (architecture)
│       └── implementation.md    # Progress tracking (tasks)
├── section/                     # Completed specifications (promoted)
│   └── <slug>.md
└── archive/                     # Archived design/implementation for completed/abandoned proposals
    └── <slug>/
        ├── design.md
        └── implementation.md
```

## MCP Tools

Use these MCP tools to access specification information:

| Tool            | Description                                                                 |
|-----------------|-----------------------------------------------------------------------------|
| `context`       | Get project rules, design, and active proposal's spec + design docs         |
| `tasks`         | Get current phase tasks with IDs (e.g., "1.1", "1.2") - shows first incomplete phase only |
| `task_complete` | Mark a task complete by ID (e.g., `task_complete(id: "1.1")`)               |
| `docs_list`     | List all available third-party documentation components                     |
| `docs_search`   | Search documentation by name - returns full content of matches              |

## MCP Prompts

Two implementation workflows are available:

| Prompt                 | Description                                                              |
|------------------------|--------------------------------------------------------------------------|
| `start-implementation` | **Methodical**: 5-phase process (investigate → plan tests → implement → validate → test). Fails fast - stops and asks user on any failure. |
| `lazy`                 | **Fast**: Implements quickly, moves past blockers, documents incomplete items. Prioritizes progress over perfection. |

## Rules for Agents

1. **Read context first**: Always call `context` before starting work to get rules, spec, and design
2. **Follow project rules**: Rules in `spec/rule/*.md` are mandatory constraints
3. **Follow coding guidelines**: Located in `spec/coding guidelines.md`
4. **Track progress**: Use `tasks` to see current work, `task_complete` to mark done
5. **Update internal todo**: Keep your own task list in sync with Nocturnal tasks
6. **Respect integrity warnings**: If `context` returns an integrity warning about changed files, STOP and ask user to confirm