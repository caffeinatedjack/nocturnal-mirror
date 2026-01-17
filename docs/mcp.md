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

### `docs_list`

Lists documentation components found in `spec/third/`.

### `docs_search`

Searches documentation components by name and returns full matching content.

## Exposed Prompts

### `start-implementation`

A methodical, fail-fast implementation approach with multiple validation checkpoints. Each task goes through 5 phases, each run as a separate subagent:

1. **Investigation** - Analyze the codebase and create an implementation plan. If there are blocking questions, stop and ask the user.
2. **Test Planning** - Define acceptance criteria and test cases BEFORE writing code.
3. **Implementation** - Write the code changes following the plan.
4. **Validation** - Verify the implementation against the test plan and specification.
5. **Testing** - Run unit tests, linters, and integration tests.

If any phase fails, the agent stops and asks the user for guidance on next steps.

Accepts an optional `goal` argument.

### `add-third-party-docs`

Instructions for generating condensed third-party docs and saving them into `spec/third/`.

### `lazy`

A fast, autonomous implementation loop that prioritizes speed over perfection:

1. Implements tasks immediately without extensive planning
2. If a task is difficult or blocked, partially completes it, adds TODO comments, and moves on
3. Spawns a quick validation subagent after each task
4. Runs tests at the end of each phase but proceeds even if they fail
5. Documents all partial completions and skipped items

Philosophy: Get working code quickly, document what's incomplete, let the user decide on follow-up.

Accepts an optional `goal` argument.

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
