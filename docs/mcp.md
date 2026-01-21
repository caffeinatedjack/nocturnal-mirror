# MCP Server

The Model Context Protocol (MCP) server exposes Nocturnal's agent tools to AI assistants like Claude, OpenCode, and other MCP-compatible clients. This allows AI assistants to directly access your project's specifications, rules, and documentation.

## Overview

The MCP server runs via standard input/output (stdio), making it easy to integrate with AI assistant configurations.

It exposes **tools** for reading project/proposal context, tracking phased tasks, and marking tasks complete; and **prompts** that guide an agent through either a normal implementation flow or an autonomous loop.

**Note**: All MCP tools are unified - optional parameters allow switching between proposal and maintenance contexts.

## Exposed Tools

### `context`

Returns:
- Project rules (`spec/rule/*.md`)
- Project design (`spec/project.md`)
- Active proposal documents: `specification.md` and `design.md`
- OR maintenance item requirements (when `maintenance_slug` parameter is provided)

**Parameters**:
- `maintenance_slug` (optional): Pass a maintenance item slug to get maintenance context instead of proposal context

Behavior notes:
- Performs a proposal integrity check using file hashes captured at activation. If proposal files changed since activation, it returns a warning and the agent should stop until the user confirms.
- May also include an "Affected Files" section (file contents) if enabled in `spec/nocturnal.yaml`.

**Examples**:
```
context()                                    # Get active proposal context
context(maintenance_slug="dependencies")     # Get maintenance item context
```

### `tasks`

Returns the **current phase** from the active proposal's `implementation.md`, or due maintenance requirements.

**Parameters**:
- `maintenance_slug` (optional): Pass a maintenance item slug to get maintenance tasks instead of proposal tasks

Behavior notes:
- For proposals: Only the **first incomplete phase** is returned.
- For proposals: Tasks get IDs like `1.1`, `1.2`, `2.1` based on their phase number and order within the phase.
- For maintenance: Returns all requirements that are currently due based on frequency and last-actioned time.

**Examples**:
```
tasks()                                      # Get current proposal phase tasks
tasks(maintenance_slug="dependencies")       # Get due maintenance requirements
```

### `task_complete`

Marks a task or maintenance requirement as complete.

**Parameters**:
- `id` (required): Task ID (e.g., "1.1") or requirement ID
- `maintenance_slug` (optional): Required when marking maintenance requirements as actioned

For proposals:
- Updates the checkbox in `implementation.md`
- If `git.auto_commit` is enabled, automatically commits all changes

For maintenance:
- Records current timestamp
- Resets the frequency counter for due date calculation

**Examples**:
```
task_complete(id="1.1")                                          # Mark proposal task complete
task_complete(id="REQ-1", maintenance_slug="dependencies")       # Mark maintenance requirement actioned
```

### `docs_list`

Lists documentation components found in `spec/third/`.

### `docs_search`

Searches documentation components by name and returns full matching content.

**Parameters**:
- `query` (required): Search term to match against component names

### `maintenance_list`

Lists all maintenance items with due/total requirement counts.

Returns items showing how many requirements are currently due based on frequency and last-actioned time.

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
1. Agent calls `context(maintenance_slug=slug)` to get due requirements
2. For each due requirement:
   - Analyze what needs to be done
   - Execute the task (update dependencies, run audits, etc.)
   - Verify completion
   - Call `task_complete(id, maintenance_slug=slug)` to mark as done
3. Call `context(maintenance_slug=slug)` again to confirm no requirements are still due
4. Report summary to user

The agent autonomously executes all due maintenance tasks, committing changes as appropriate.

### `add-third-party-docs`

Instructions for generating condensed third-party docs and saving them into `spec/third/`.

### `populate-spec-sections`

A comprehensive guide for writing complete specification sections for a new project. This prompt helps AI agents create full, RFC-formatted specifications for all features.

**Workflow:**

1. **Understand the Project** - Gather high-level description, feature areas, technologies, and constraints
2. **Identify Specification Sections** - Propose logical groupings of features (e.g., "authentication", "API", "workflows")
3. **For Each Section**:
   - Gather detailed functional and technical requirements through questions
   - Write comprehensive spec following IETF RFC format with numbered sections
   - Include: Introduction, Requirements Notation, Terminology, Concepts, Core Technical Sections, Examples, Security Considerations, Error Handling
   - Validate completeness before moving to next section
4. **Cross-Reference Check** - Review all specs for inconsistencies, missing dependencies, gaps
5. **Final Validation** - Present summary with coverage assessment and readiness check

**Philosophy:**
- **Complete specifications** - Every feature, API, behavior, and constraint documented
- **RFC format** - IETF RFC/Internet-Draft structure with normative language (MUST/SHOULD/MAY)
- **Ask questions** - When clarification needed about features or technical details
- **Technical rigor** - Precise, unambiguous, and testable requirements
- **One spec per domain** - Group related functionality into logical documents

**Output:** Creates markdown files in `spec/section/[section-slug].md` following the project's specification guidelines. These become the foundation for all development.

Use this prompt when starting a new project to establish comprehensive specifications before creating any proposals or writing code.

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
