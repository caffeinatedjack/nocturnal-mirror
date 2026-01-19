# MCP Server

The Model Context Protocol (MCP) server exposes Nocturnal's agent tools to AI assistants like Claude, OpenCode, and other MCP-compatible clients. This allows AI assistants to directly access your project's specifications, rules, and documentation.

## Overview

The MCP server runs via standard input/output (stdio), making it easy to integrate with AI assistant configurations.

It exposes **tools** for reading project/proposal context, tracking phased tasks, and marking tasks complete; and **prompts** that guide an agent through either a normal implementation flow or an autonomous loop.

## Exposed Tools

### `context`

Returns:
- Project rules (`spec/rule/*.md`)
- Project design (`spec/project.md`)
- Active proposal documents: `specification.md` and `design.md`

Behavior notes:
- Performs a proposal integrity check using file hashes captured at activation. If proposal files changed since activation, it returns a warning and the agent should stop until the user confirms.
- May also include an "Affected Files" section (file contents) if enabled in `spec/nocturnal.yaml`.

### `tasks`

Returns the **current phase** from the active proposal's `implementation.md`.

Behavior notes:
- Only the **first incomplete phase** is returned.
- Tasks get IDs like `1.1`, `1.2`, `2.1` based on their phase number and order within the phase.

### `task_complete`

Marks a task complete by ID (e.g. `task_complete(id: "1.1")`) by updating the checkbox in `implementation.md`.

### `task_snapshot`

Creates a git snapshot before starting work on a task (if git integration is enabled in config).

### `docs_list`

Lists documentation components found in `spec/third/`.

### `docs_search`

Searches documentation components by name and returns full matching content.

### `maintenance_list`

Lists all maintenance items with due/total requirement counts.

Returns items showing how many requirements are currently due based on frequency and last-actioned time.

### `maintenance_context`

Gets requirements for a specific maintenance item, showing which are currently due.

Parameters:
- `slug` - Maintenance item slug

Returns full file content with parsed requirements, due status, and instructions for marking items as actioned.

### `maintenance_actioned`

Marks a maintenance requirement as completed.

Parameters:
- `slug` - Maintenance item slug
- `id` - Requirement ID

Records current timestamp and resets the frequency counter for due date calculation.

## Exposed Prompts

### `elaborate-spec`

Guides comprehensive elaboration of a proposal specification with a thorough 9-step process:

1. **Identify the Proposal** - Asks user which proposal to elaborate on
2. **Gather Requirements** - Functional requirements, technical constraints, dependencies
3. **Third-Party Documentation Check** - Verifies docs exist, highlights missing ones using `docs_search`
4. **Design Elaboration** - Architecture, components, technical decisions, testing strategy
5. **Implementation Planning** - Creates phased approach with specific, actionable tasks
6. **Specification Update** - Ensures proper RFC 2119 requirements, success criteria, scope
7. **Dependency Documentation** - Documents proposal, third-party, and system dependencies
8. **Validation** - 8-point checklist before finalizing
9. **Summary** - Presents overview with readiness assessment

Philosophy:
- **98% confidence threshold** - Agent must ask user if not highly confident about any aspect
- **Comprehensive documentation** - Detailed enough for any developer to implement
- **Dependency awareness** - Explicitly asks about and documents all dependencies
- **Third-party documentation** - Highlights missing docs and suggests using `add-third-party-docs`

Use this prompt BEFORE implementation to ensure the proposal is thoroughly designed and planned.

### `start-implementation`

A methodical, fail-fast implementation approach with multiple validation checkpoints. Each task goes through 5 phases, each run as a separate subagent:

1. **Investigation** - Analyze the codebase and create an implementation plan. If there are blocking questions, stop and ask the user.
2. **Test Planning** - Define acceptance criteria and test cases BEFORE writing code.
3. **Implementation** - Write the code changes following the plan.
4. **Validation** - Verify the implementation against the test plan and specification.
5. **Testing** - Run unit tests, linters, and integration tests.

If any phase fails, the agent stops and asks the user for guidance on next steps.

Accepts an optional `goal` argument.

### `lazy`

A fast, autonomous implementation loop that prioritizes speed over perfection:

1. Implements tasks immediately without extensive planning
2. If a task is difficult or blocked, partially completes it, adds TODO comments, and moves on
3. Spawns a quick validation subagent after each task
4. Runs tests at the end of each phase but proceeds even if they fail
5. Documents all partial completions and skipped items

Philosophy: Get working code quickly, document what's incomplete, let the user decide on follow-up.

Accepts an optional `goal` argument.

### `start-maintenance`

Guides an agent through executing all due requirements for a maintenance item.

Parameters:
- `slug` - Maintenance item slug

Workflow:
1. Agent calls `maintenance_context` to get due requirements
2. For each due requirement:
   - Analyze what needs to be done
   - Execute the task (update dependencies, run audits, etc.)
   - Verify completion
   - Call `maintenance_actioned(slug, id)` to mark as done
3. Call `maintenance_context` again to confirm no requirements are still due
4. Report summary to user

The agent autonomously executes all due maintenance tasks, committing changes as appropriate.

### `add-third-party-docs`

Instructions for generating condensed third-party docs and saving them into `spec/third/`.

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
