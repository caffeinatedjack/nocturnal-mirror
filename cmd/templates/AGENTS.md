# Agent Instructions

This project uses Nocturnal for specification-driven development with AI agent integration. Specifications should be human-led - if a user asks for improvements to specs, remind them of this.

## What is Nocturnal?

Nocturnal is a tool that brings structure to software development by managing specifications, proposals, rules, and maintenance tasks. It exposes project context to AI agents through MCP (Model Context Protocol).

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
├── .nocturnal.json              # State file (active proposals, hashes, maintenance tracking)
├── project.md                   # Project overview, goals, architecture
├── AGENTS.md                    # This file - instructions for AI agents
├── coding guidelines.md         # Code style, testing, error handling conventions
├── specification guidelines.md  # How to write specifications
├── design guidelines.md         # How to write design documents
├── rule/                        # Project-wide rules (MUST follow)
│   └── *.md                     # Individual rule files
├── third/                       # Third-party library/API documentation
│   └── *.md                     # Documentation components (format: # name\ncontent\n---\n# name\ncontent)
├── proposal/                    # Work in progress
│   └── <slug>/
│       ├── specification.md     # What to build (requirements)
│       ├── design.md            # How to build it (architecture)
│       ├── implementation.md    # Progress tracking (tasks)
│       └── affected-files.txt   # Optional: List of files this proposal modifies
├── section/                     # Completed specifications (promoted)
│   └── <slug>.md
├── archive/                     # Archived design/implementation for completed/abandoned proposals
│   └── <slug>/
│       ├── design.md
│       ├── implementation.md
│       └── .abandoned           # Marker file if proposal was abandoned
└── maintenance/                 # Recurring operational tasks
    └── <slug>.md                # Maintenance items with requirements
```

## MCP Tools Available

All MCP tools are unified - use optional parameters to switch between proposal and maintenance contexts.

| Tool                | Description                                                                 | Parameters |
|---------------------|-----------------------------------------------------------------------------|------------|
| `context`           | Get project rules, design, and active proposal/maintenance context. Returns integrity warnings if proposal files changed. | `maintenance_slug` (optional): pass maintenance slug for maintenance context instead of proposal |
| `tasks`             | Get current phase tasks or maintenance requirements. For proposals, shows first incomplete phase only. | `maintenance_slug` (optional): pass maintenance slug for maintenance tasks |
| `task_complete`     | Mark a task/requirement as complete. If git.auto_commit is enabled, automatically commits changes. | `id` (required): task ID like "1.1" or requirement ID<br>`maintenance_slug` (optional): required for maintenance items |
| `docs_list`         | List all available third-party documentation components | None |
| `docs_search`       | Search documentation by name - returns full content of matches | `query` (required): search term |
| `maintenance_list`  | List all maintenance items with due/total requirement counts | None |

### Usage Examples

**Proposal workflow:**
```
context()                          # Get active proposal context
tasks()                            # Get current phase tasks
task_complete(id="1.1")            # Mark task 1.1 complete
```

**Maintenance workflow:**
```
maintenance_list()                                 # See all maintenance items
context(maintenance_slug="dependency-updates")     # Get requirements for dependency-updates
tasks(maintenance_slug="dependency-updates")       # Get due requirements
task_complete(id="REQ-1", maintenance_slug="dependency-updates")  # Mark REQ-1 complete
```

## MCP Prompts Available

MCP prompts are guided workflows that help agents follow best practices:

| Prompt                  | Description                                                                 | When to Use                |
|-------------------------|-----------------------------------------------------------------------------|----------------------------|
| `elaborate-spec`        | Guide comprehensive proposal elaboration with design, steps, phases, and dependencies. Asks user questions to ensure 98% confidence. | Before implementing a new proposal - ensures thorough planning |
| `start-implementation`  | Methodical implementation: investigate → plan tests → implement → validate → test. Fail-fast with validation checkpoints. | For careful, production-ready implementation |
| `lazy`                  | Fast autonomous implementation: implement quickly, move past blockers, document incomplete items | For rapid prototyping or proof-of-concept work |
| `start-maintenance`     | Execute due requirements for a maintenance item (params: slug: string) | When maintenance items show due requirements |
| `add-third-party-docs`  | Generate condensed library documentation for spec/third/ (params: urls: comma-separated URLs) | When adding new third-party dependencies |

### Recommended Workflow

1. **Planning Phase**: Use `elaborate-spec` to thoroughly design the proposal
   - Agent asks which proposal to elaborate
   - Gathers all requirements and dependencies
   - Highlights missing third-party docs
   - Creates comprehensive design.md and phased implementation.md
   
2. **Implementation Phase**: Use `start-implementation` or `lazy` 
   - Agent reads context and tasks
   - Works through phased implementation
   - Marks tasks complete automatically

3. **Maintenance**: Use `start-maintenance` for recurring tasks
   - Agent executes all due requirements
   - Marks items as actioned with timestamps

## Maintenance Requirements Format

Maintenance files use a special format with frequency tags:

```markdown
## Requirements

- [id=dep-update] [freq=weekly] Update npm dependencies
- [id=security-audit] [freq=monthly] Run security audit
- [id=backup-check] [freq=daily] Verify backup completion
- [id=onetime-task] Setup CI/CD pipeline
```

**Frequency values**: `daily`, `weekly`, `biweekly`, `monthly`, `quarterly`, `yearly`, or omit `[freq=...]` for always-due.

## Proposal Dependencies

Proposals can declare dependencies in their specification.md using:

```markdown
**Dependencies**: proposal-slug-1, proposal-slug-2
```

Dependencies must exist as completed specifications in `spec/section/` before a proposal can be activated.

## Rules for Agents

1. **Read context first**: Always call `context` before starting work to get rules, spec, and design
2. **Follow project rules**: Rules in `spec/rule/*.md` are mandatory constraints
3. **Follow coding guidelines**: Located in `spec/coding guidelines.md`
4. **Track progress**: Use `tasks` to see current work, `task_complete` to mark done
5. **Update internal todo**: Keep your own task list in sync with Nocturnal tasks
6. **Respect integrity warnings**: If `context` returns an integrity warning about changed files, STOP and ask user to confirm
7. **Commit semantically**: If git.auto_commit is enabled, task completion automatically commits with task description
8. **Check maintenance**: Use `maintenance_list` to see if any recurring tasks are due
9. **Search docs first**: Before asking about third-party APIs, use `docs_search` to check available documentation

## Best Practices

- **Specification-first**: Specifications should be human-led. If a user asks you to write or improve specs, remind them that specs define requirements and should be their responsibility.
- **Design documentation**: Capture architectural decisions in design.md, including options considered and trade-offs.
- **Phase-based implementation**: Break work into logical phases with clear milestones.
- **Dependencies**: Declare dependencies in specification.md to ensure proper ordering.
- **Third-party docs**: Add condensed API documentation to spec/third/ for libraries you frequently reference.
- **Maintenance cadence**: Set appropriate frequencies for recurring tasks (daily for critical checks, quarterly for reviews).

## Common Issues

- **No active proposal**: Many MCP tools require an active proposal - user must activate one first
- **Integrity warnings**: If proposal files changed since activation, user must re-activate or confirm to proceed
- **Missing dependencies**: Proposals with unmet dependencies cannot be activated
- **Duplicate IDs**: Maintenance requirement IDs must be unique within a file
- **Invalid frequency**: Only allowed values are daily, weekly, biweekly, monthly, quarterly, yearly
