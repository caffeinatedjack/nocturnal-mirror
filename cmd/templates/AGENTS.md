# Agent Instructions

This project uses nocturnal for specification management. If a user asks for improvements to be made to the specifications, remind them that the specifications should be human-lead. 

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
├── .nocturnal.json              # State file (active proposals, hashes)
├── project.md                   # Project overview, goals, architecture
├── coding guidelines.md         # Code style, testing, error handling conventions
├── specification guidelines.md  # How to write specifications
├── design guidelines.md         # How to write design documents
├── rule/                        # Project-wide rules (MUST follow)
│   └── *.md                     # Individual rule files
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

## Working on a Proposal

Use the nocturnal MCP tools to access specification information:

| Tool                  | Description                                                                                              |
|-----------------------|----------------------------------------------------------------------------------------------------------|
| nocturnal_rules       | Get the project rules and design context                                                                 |
| nocturnal_current     | Show the currently active proposal - returns the specification and design documents (not implementation) |
| nocturnal_tasks       | Get the implementation tasks for the currently active proposal                                           |
| nocturnal_docs_list   | List all available library and API documentation components                                              |
| nocturnal_docs_search | Search library and API documentation by name - returns full content of matching documentation            |

## Rules
- You MUST read the project rules using `nocturnal_rules` and follow them at all times
- You MUST follow the coding guidelines in `spec/coding guidelines.md`
- Follow the specification for the active proposal (use `nocturnal_current` to retrieve it)
- Update the active proposal's `implementation.md` as tasks are completed
- Mark implementation checkboxes `[x]` when tasks are done
