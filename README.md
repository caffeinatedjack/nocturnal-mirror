![NOCTURNAL](docs/nocturnal.png)

# Nocturnal

**Specification-driven development with AI agent integration**

Nocturnal is a CLI tool and MCP server that brings structure to software development when using AI by managing specifications, proposals, and maintenance tasks in a systematic workflow. 

## Why?

Nocturnal addresses common development challenges:
- **Spec drift** - Requirements scattered across tickets, docs, and comments
- **Context loss** - AI agents lack project architecture and constraint knowledge
- **Maintenance debt** - Recurring tasks (updates, audits) forgotten or inconsistent
- **Knowledge silos** - Critical decisions undocumented
- **AI Context bloat** - Designed around the idea that AIs context get bloated, resulting in high cost and lower quality code.
- **Outdated AI data** - When a AI is trained on a older version of a library, it can result in a lot of issues during development. This comes with the ability to easily include this context to give the AI the information it needs off the bat.

Provides a structured workspace with proposals (spec + design + implementation), persistent rules, frequency-tracked maintenance, and MCP integration for AI agents.

## Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              NOCTURNAL CLI                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  nocturnal                                                                  │
│  ├── spec                    Specification management                       │
│  │   ├── init                  Initialize workspace                         │
│  │   ├── view                  Show workspace overview                      │
│  │   ├── stats                 Project statistics                           │
│  │   ├── config                Configuration management                     │
│  │   ├── proposal              Proposal lifecycle                           │
│  │   │   ├── add                 Create new proposal                        │
│  │   │   ├── list                List all proposals                         │
│  │   │   ├── activate            Set as active                              │
│  │   │   ├── deactivate          Unset active                               │
│  │   │   ├── current             Show active proposal                       │
│  │   │   ├── validate            Check against guidelines                   │
│  │   │   ├── complete            Promote to section/                        │
│  │   │   ├── abandon             Archive without promoting                  │
│  │   │   ├── remove              Delete proposal                            │
│  │   │   └── graph               Show dependency graph                      │
│  │   ├── precursor (experimental) Reusable proposal templates               │
│  │   │   ├── init                Create new precursor bundle                │
│  │   │   ├── validate            Check precursor structure                  │
│  │   │   ├── pack                Create zip from directory                  │
│  │   │   └── unpack              Extract zip to directory                   │
│  │   ├── rule                  Project-wide rules                           │
│  │   │   ├── add                 Create new rule                            │
│  │   │   └── show                Display all rules                          │
│  │   └── maintenance          Recurring operational tasks                   │
│  │       ├── add                 Create maintenance item                    │
│  │       ├── list                Show all items with due counts             │
│  │       ├── show                Display item content                       │
│  │       ├── due                 Show due requirements                      │
│  │       ├── actioned            Mark requirement as completed              │
│  │       └── remove              Delete maintenance item                    │
│  ├── agent                   Agent context commands                         │
│  │   ├── current               Show active proposal context                 │
│  │   ├── project               Show rules + project design                  │
│  │   └── specs                 Show completed specifications                │
│  ├── docs                    Third-party documentation                      │
│  │   ├── list                  List all doc components                      │
│  │   └── search                Search docs by name                          │
│  └── mcp                     MCP server for AI agents                       │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                              MCP SERVER                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Tools (callable by AI agents)                                              │
│  ├── context                 Get rules, design, active proposal/maintenance │
│  │                           (use maintenance_slug param for maintenance)   │
│  ├── tasks                   Get current phase tasks or maintenance reqs    │
│  │                           (use maintenance_slug param for maintenance)   │
│  ├── task_complete           Mark task/requirement done by ID               │
│  │                           (use maintenance_slug param for maintenance)   │
│  ├── docs_list               List documentation components                  │
│  ├── docs_search             Search documentation by name                   │
│  └── maintenance_list        List maintenance items with due counts         │
│                                                                             │
│  Prompts (implementation workflows)                                         │
│  ├── elaborate-spec          Guide comprehensive proposal elaboration with  │
│  │                           design, steps, phases, and dependencies        │
│  ├── start-implementation    Methodical: investigate → plan tests →         │
│  │                           implement → validate → test (fail-fast)        │
│  ├── lazy                    Fast: implement quickly, move past blockers,   │
│  │                           document incomplete items                      │
│  ├── start-maintenance       Execute due requirements for maintenance item  │
│  ├── add-third-party-docs    Generate condensed library documentation       │
│  └── populate-spec-sections  Write comprehensive specs for new project      │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                           PROPOSAL LIFECYCLE                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌─────────┐    ┌──────────┐    ┌─────────┐    ┌──────────────────────┐    │
│   │   ADD   │───►│ ACTIVATE │───►│ DEVELOP │───►│ COMPLETE / ABANDON   │    │
│   └─────────┘    └──────────┘    └─────────┘    └──────────────────────┘    │
│        │              │               │                    │                │
│        ▼              ▼               ▼                    ▼                │
│   proposal/      Set as          Edit docs:          ┌─────────────┐        │
│   <slug>/        primary         spec, design,       │ COMPLETE:   │        │
│   ├─ spec.md     active          implementation      │ → section/  │        │
│   ├─ design.md                                       │ → archive/  │        │
│   └─ impl.md                                         ├─────────────┤        │
│                                                      │ ABANDON:    │        │
│                                                      │ → archive/  │        │
│                                                      └─────────────┘        │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Features

### Proposals - Structured Feature Development
Draft changes with three interconnected documents:
- **Specification** - Formal requirements using normative language (MUST/SHOULD/MAY)
- **Design** - Technical decisions, options considered, and rationale
- **Implementation** - Phased task breakdown with progress tracking

Proposals follow a complete lifecycle: add → activate → develop → validate → complete (archive design/impl, promote spec).

### Rules - Project-Wide Standards
Define persistent constraints that apply across all development:
- Coding standards and conventions
- Architectural patterns and technology choices
- Security requirements
- Testing standards

Rules are always included in agent context, ensuring consistency.

### Maintenance - Recurring Operational Tasks
Track periodic maintenance with frequency-based due dates:
- Dependency updates (weekly, monthly, quarterly)
- Security audits and compliance checks
- Documentation reviews
- Infrastructure maintenance

Agents can easily work on maintenance tasks, based on a schedule.

### AI Agent Integration
Expose project context and tools to AI assistants:
- **Tools**: Read specifications, track tasks, mark completions, manage maintenance
- **Prompts**: Guided workflows for implementation and maintenance execution
- **Integrity checks**: Warn agents when proposal files change
- **Compatible with**: Claude Desktop, OpenCode, Cursor, and other MCP clients

### Validation - Documentation Quality
Check proposals against guidelines before completion:
- Required and recommended sections
- Use of normative language
- Unfilled template placeholders
- Task structure and formatting

### Third-Party Docs - API Context
Store condensed documentation for libraries and APIs:
- Keep relevant API docs in `spec/third/`
- Search by component name
- Provide to agents for implementation context

### Proposal Precursors - Reusable Templates (Experimental)
Create shareable proposal templates for common scenarios:
- **Parameterized templates**: Define custom inputs for proposals
- **Custom documents**: Override spec, design, or implementation templates
- **Bundled documentation**: Include relevant third-party docs
- **Portable**: Share as directory or zip file across projects

Example use cases: database migrations, service creation, API integration patterns.

Read the [full documentation](/docs/index.md) for detailed usage.

## Quick Start

```bash
# Initialize a specification workspace
nocturnal spec init

# Add a project rule
nocturnal spec rule add coding-standards

# Create a feature proposal
nocturnal spec proposal add user-authentication

# Activate it for development
nocturnal spec proposal activate user-authentication

# View workspace status
nocturnal spec view

# Validate documentation quality
nocturnal spec proposal validate user-authentication

# Complete and promote
nocturnal spec proposal complete user-authentication

# Create recurring maintenance tasks
nocturnal spec maintenance add dependencies
nocturnal spec maintenance list

# (Experimental) Create a reusable precursor template
nocturnal precursor init database-migration --out ./db-migration.zip

# Use precursor to create a proposal
nocturnal spec proposal add migrate-to-postgres --precursor-path ./db-migration.zip
# Fill in the generated precursor-answers.yaml, then regenerate:
nocturnal spec proposal add migrate-to-postgres --precursor-path ./db-migration.zip --overwrite
```

## Installation

```bash
# From source
make build
make install

# Binary will be installed to ~/.local/bin/nocturnal
```

## MCP Integration Setup

Nocturnal exposes its functionality via Model Context Protocol (MCP), allowing AI assistants to access your project context.

### OpenCode

Add to `~/.opencode/config.json`:

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

### VS Code with GitHub Copilot

Add to `.vscode/settings.json` in your project:

```json
{
  "github.copilot.advanced": {
    "mcp": {
      "servers": {
        "nocturnal": {
          "command": "nocturnal",
          "args": ["mcp"],
          "cwd": "${workspaceFolder}"
        }
      }
    }
  }
}
```

See [docs/mcp.md](docs/mcp.md) for full configuration details and usage.

## Future plans
- Effective TUI to show the project current state 
- Further experimentation with automation, embedding it further into AI tools

## License
MIT
