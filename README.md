![NOCTURNAL](docs/nocturnal.png)

A CLI tool for spec-driven development and agent tooling.

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
│  │   └── rule                  Project-wide rules                           │
│  │       ├── add                 Create new rule                            │
│  │       └── show                Display all rules                          │
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
│  ├── context                 Get rules, project design, active proposal     │
│  ├── tasks                   Get current phase tasks with IDs               │
│  ├── task_complete           Mark task done by ID (e.g., "1.1")             │
│  ├── docs_list               List documentation components                  │
│  └── docs_search             Search documentation by name                   │
│                                                                             │
│  Prompts (implementation workflows)                                         │
│  ├── start-implementation    Methodical: investigate → plan tests →         │
│  │                           implement → validate → test (fail-fast)        │
│  ├── lazy                    Fast: implement quickly, move past blockers,   │
│  │                           document incomplete items                      │
│  └── add-third-party-docs    Generate condensed library documentation       │
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

- **Proposals** - Draft changes with specification, design, and implementation documents
- **Validation** - Check proposals against documentation guidelines before completion
- **Rules** - Project-wide constraints that persist across all proposals
- **MCP Server** - Expose tools to AI assistants (Claude, OpenCode, Cursor, etc.)
- **Third-party Docs** - Store and search API/library documentation for AI context

Read the [full documentation](/docs/index.md) for detailed usage.

## Installation

```bash
make build
make install
```

Or download the binary from releases.
