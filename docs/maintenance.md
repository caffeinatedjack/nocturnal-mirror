# Maintenance Management

Maintenance items are recurring operational tasks like dependency upgrades, security audits, and periodic reviews. Nocturnal helps you track these requirements with frequency-based due dates and state tracking.

## Overview

Maintenance items are stored as markdown files in `spec/maintenance/` with a simple structure:
- Each item contains a list of requirements
- Requirements have unique IDs for tracking
- Optional frequency tags (daily, weekly, monthly, etc.)
- State is tracked in `spec/.nocturnal.json` with last-actioned timestamps
- Due status is computed automatically based on frequency intervals

Unlike proposals (which have a lifecycle), maintenance items are permanent operational tasks that recur over time.

## Use Cases

### Dependency Management
Track updates for dependencies, toolchains, and infrastructure:
```markdown
- Update Go toolchain [id=go-update] [freq=monthly]
- Review npm dependencies for vulnerabilities [id=npm-audit] [freq=weekly]
- Upgrade database client library [id=db-client] [freq=quarterly]
```

### Security & Compliance
Schedule periodic security reviews and compliance checks:
```markdown
- Run security audit on authentication [id=auth-audit] [freq=quarterly]
- Review access control policies [id=access-review] [freq=monthly]
- Rotate API keys [id=key-rotation] [freq=yearly]
```

### Code Quality
Maintain code health with recurring reviews:
```markdown
- Review test coverage [id=test-coverage]
- Check for outdated TODO comments [id=todo-review] [freq=monthly]
- Update documentation for API changes [id=doc-sync] [freq=biweekly]
```

### Infrastructure
Keep infrastructure current and monitored:
```markdown
- Review CI/CD pipeline performance [id=ci-perf] [freq=quarterly]
- Check disk space usage [id=disk-check] [freq=weekly]
- Update SSL certificates [id=ssl-update] [freq=yearly]
```

## Commands

### spec maintenance add

Create a new maintenance item.

```bash
nocturnal spec maintenance add <name-or-slug>
```

**Arguments:**
- `<name-or-slug>` - Name of the maintenance item (converted to slug)

**What it does:**
- Creates `spec/maintenance/<slug>.md` file
- Generates a template with examples
- Sets up the structure for adding requirements

**Slug conversion:**
Same as proposals: lowercase, hyphens for spaces, special characters removed.

**Example:**
```bash
nocturnal spec maintenance add "Go Dependencies"
```

**Output:**
```
Created maintenance item 'go-dependencies'
Location: spec/maintenance/go-dependencies.md
```

**Template structure:**
```markdown
# Maintenance: Go Dependencies

**Slug**: go-dependencies

## Requirements

<!-- Each requirement must have an [id=...] tag. Frequency is optional. -->
<!-- Allowed frequencies: daily, weekly, biweekly, monthly, quarterly, yearly -->
<!-- If freq is omitted, the requirement is always due. -->

<!-- Example:
- Update Go toolchain in CI [id=go-toolchain] [freq=monthly]
- Run security audit [id=sec-audit] [freq=quarterly]
- Review dependencies [id=dep-review]
-->
```

---

### spec maintenance list

List all maintenance items with due counts.

```bash
nocturnal spec maintenance list
```

**What it displays:**
- All maintenance items (by slug)
- Count of due requirements vs total requirements
- Visual highlighting for items with due requirements

**Example output:**
```
Maintenance Items (3)

  go-dependencies  2/4 due
  security-audits  0/3 due
  documentation    1/1 due
```

**Color coding:**
- Items with due requirements are highlighted in yellow/warning style
- Items with no due requirements are shown in dimmed text

**Use case:**
Run this regularly to see what maintenance tasks need attention.

---

### spec maintenance show

Display the full content of a maintenance item.

```bash
nocturnal spec maintenance show <slug>
```

**Arguments:**
- `<slug>` - Name of the maintenance item

**What it displays:**
- Complete markdown file content
- All requirements with their ID and frequency tags
- Any notes or documentation in the file

**Example:**
```bash
nocturnal spec maintenance show go-dependencies
```

**Output:**
```markdown
# Maintenance: Go Dependencies

**Slug**: go-dependencies

## Requirements

- Update Go toolchain in CI [id=go-toolchain] [freq=monthly]
- Review go.mod for outdated dependencies [id=go-mod-review] [freq=weekly]
- Run `go vet` on all packages [id=go-vet] [freq=daily]
- Security scan with gosec [id=gosec] [freq=weekly]
```

---

### spec maintenance due

Show only the requirements that are currently due.

```bash
nocturnal spec maintenance due <slug>
```

**Arguments:**
- `<slug>` - Name of the maintenance item

**What it displays:**
- Requirements that are currently due
- Requirement ID, text, frequency
- Last actioned timestamp (if any)

**Due criteria:**
A requirement is due if:
1. It has never been actioned, OR
2. The frequency interval has elapsed since last actioned, OR
3. It has no frequency tag (always due)

**Example:**
```bash
nocturnal spec maintenance due go-dependencies
```

**Output:**
```
Due Requirements: go-dependencies

  [go-toolchain]  Update Go toolchain in CI
      freq: monthly
      last: 2025-12-19T10:30:00Z

  [go-vet]  Run `go vet` on all packages
      freq: daily
      last: 2026-01-18T09:00:00Z
```

**Use case:**
Before executing maintenance tasks, check what's currently due to prioritize work.

---

### spec maintenance actioned

Mark a requirement as actioned (completed).

```bash
nocturnal spec maintenance actioned <slug> <id>
```

**Arguments:**
- `<slug>` - Name of the maintenance item
- `<id>` - Requirement ID (from the `[id=...]` tag)

**What it does:**
- Records current timestamp in `spec/.nocturnal.json`
- Updates the requirement's last-actioned time
- Resets the frequency counter for due date calculation

**Example:**
```bash
nocturnal spec maintenance actioned go-dependencies go-toolchain
```

**Output:**
```
Marked 'go-toolchain' as actioned
Update Go toolchain in CI
```

**State tracking:**
The state is stored in `spec/.nocturnal.json`:
```json
{
  "maintenance": {
    "go-dependencies": {
      "go-toolchain": {
        "last_actioned": "2026-01-19T10:15:00Z"
      }
    }
  }
}
```

**Use case:**
After completing a maintenance task, mark it as actioned so the due date is recalculated.

---

### spec maintenance remove

Remove a maintenance item and its tracking state.

```bash
nocturnal spec maintenance remove <slug>
```

**Arguments:**
- `<slug>` - Name of the maintenance item to remove

**What it does:**
- Deletes the maintenance file from `spec/maintenance/`
- Removes tracking state from `spec/.nocturnal.json`
- Cannot be undone

**Example:**
```bash
nocturnal spec maintenance remove go-dependencies
```

**Output:**
```
Removed maintenance item 'go-dependencies'
```

**When to use:**
- Maintenance item is no longer relevant
- Requirements have been moved to another item
- Consolidating multiple items

---

## Frequency Options

Maintenance requirements support the following frequency intervals:

| Frequency    | Interval     | Use Case                                    |
|--------------|--------------|---------------------------------------------|
| `daily`      | 1 day        | Daily checks, logs, monitoring              |
| `weekly`     | 7 days       | Weekly reviews, minor updates               |
| `biweekly`   | 14 days      | Bi-weekly sprints, mid-month reviews        |
| `monthly`    | 1 month      | Monthly updates, reports, minor versions    |
| `quarterly`  | 3 months     | Quarterly audits, major reviews             |
| `yearly`     | 1 year       | Annual reviews, major upgrades, certificates|
| *(omitted)*  | Always due   | One-time tasks, always show until actioned  |

**Note:** Frequency is optional. If omitted, the requirement is always shown as due until actioned.

---

## MCP Integration

Maintenance items are exposed to AI agents via the MCP server, allowing automated execution of maintenance tasks.

### MCP Tools

#### `maintenance_list`

Returns all maintenance items with due/total counts.

**Response:**
```
Maintenance Items:

go-dependencies: 2/4 due
security-audits: 0/3 due
documentation: 1/1 due
```

#### `maintenance_context`

Gets requirements for a specific maintenance item, showing which are due.

**Parameters:**
- `slug` - Maintenance item slug

**Response:**
Includes:
- Full file content
- Parsed requirements with due status
- Last actioned timestamps
- Instruction to call `maintenance_actioned` after completing tasks

**Example usage (by agent):**
```
maintenance_context(slug="go-dependencies")
```

#### `maintenance_actioned`

Marks a requirement as completed.

**Parameters:**
- `slug` - Maintenance item slug
- `id` - Requirement ID

**Example usage (by agent):**
```
maintenance_actioned(slug="go-dependencies", id="go-toolchain")
```

### MCP Prompt: `start-maintenance`

A specialized prompt that guides an agent through executing all due requirements for a maintenance item.

**Parameters:**
- `slug` - Maintenance item slug

**Workflow:**
1. Agent calls `maintenance_context` to get due requirements
2. For each due requirement:
   - Analyze what needs to be done
   - Execute the task (update dependencies, run audits, etc.)
   - Verify completion
   - Call `maintenance_actioned(slug, id)` to mark as done
3. Call `maintenance_context` again to confirm no requirements are still due
4. Report summary to user

**Example invocation (in MCP client like Claude):**
```
Use the start-maintenance prompt with slug="go-dependencies"
```

**Agent behavior:**
The agent will autonomously execute all due requirements, committing changes as appropriate, and marking each task as actioned when complete.

---

## Workflow Example

### 1. Create a maintenance item

```bash
nocturnal spec maintenance add "Security Audits"
```

Edit `spec/maintenance/security-audits.md`:
```markdown
# Maintenance: Security Audits

**Slug**: security-audits

## Requirements

- Review authentication flows for vulnerabilities [id=auth-review] [freq=quarterly]
- Scan dependencies for known CVEs [id=cve-scan] [freq=weekly]
- Rotate API keys and secrets [id=key-rotation] [freq=yearly]
- Review access logs for anomalies [id=log-review] [freq=monthly]
```

### 2. Check what's due

```bash
nocturnal spec maintenance due security-audits
```

Output shows which requirements are currently due based on their frequency and last-actioned time.

### 3. Execute maintenance tasks

Manually execute each due task, or use the MCP server to have an agent do it:

```
Use the start-maintenance prompt with slug="security-audits"
```

The agent will:
- Get the list of due requirements
- Execute each task (run scans, review logs, etc.)
- Mark each as actioned when complete

### 4. Verify completion

```bash
nocturnal spec maintenance list
```

Should show `security-audits` with 0 due requirements.

### 5. Regular monitoring

Add to your workflow:
```bash
# Daily/weekly check
nocturnal spec maintenance list

# Execute due tasks
nocturnal spec maintenance due <slug>
# ... perform tasks ...
nocturnal spec maintenance actioned <slug> <id>
```

---

## Best Practices

### Requirement Writing

**DO:**
- ✓ Use clear, actionable language
- ✓ Include the specific action (update, review, scan, rotate)
- ✓ Set appropriate frequencies based on criticality
- ✓ Use unique, descriptive IDs (kebab-case recommended)

**DON'T:**
- ✗ Use vague descriptions ("maintain dependencies")
- ✗ Duplicate IDs within the same file
- ✗ Set overly aggressive frequencies that cause alert fatigue

**Examples:**

Good:
```markdown
- Update Go toolchain to latest patch version [id=go-update] [freq=monthly]
- Run gosec security scanner on codebase [id=gosec-scan] [freq=weekly]
```

Poor:
```markdown
- Check stuff [id=check] [freq=daily]
- Dependencies [id=deps]
```

### Frequency Selection

Choose frequency based on:
- **Criticality** - Security items more frequent than documentation
- **Rate of change** - Fast-moving dependencies need more frequent checks
- **Effort** - Balance maintenance overhead with value
- **Risk** - Higher risk areas need more frequent attention

**Guidelines:**
- Security: weekly to quarterly
- Dependencies: weekly to monthly
- Documentation: monthly to quarterly
- Infrastructure: quarterly to yearly
- Certificates/credentials: yearly (with advance warning)

### Organization

**Group by domain:**
```bash
nocturnal spec maintenance add go-dependencies
nocturnal spec maintenance add npm-dependencies
nocturnal spec maintenance add security-audits
nocturnal spec maintenance add documentation-sync
nocturnal spec maintenance add infrastructure
```

**Avoid overly granular items:**
Instead of creating 10 items with 1 requirement each, create 3-4 items with related requirements grouped together.

**Split when items grow too large:**
If a maintenance item has >10 requirements, consider splitting by subdomain.

### State Management

The state file (`spec/.nocturnal.json`) tracks when requirements were last actioned:
- Committed to version control for team synchronization
- Automatically updated by `maintenance actioned` command
- Read by `maintenance due` to compute due status
- Used by MCP tools for agent execution

**Team coordination:**
- Commit state changes after actioning requirements
- Pull latest state before checking due items
- Resolve conflicts by keeping latest timestamps

---

## Integration with CI/CD

You can integrate maintenance checks into your CI/CD pipeline:

### Check for due maintenance

```bash
#!/bin/bash
# scripts/check-maintenance.sh

# Get list of maintenance items
items=$(nocturnal spec maintenance list 2>&1 | grep "due" | grep -v "0/")

if [ -n "$items" ]; then
  echo "::warning::Maintenance items have due requirements:"
  echo "$items"
  # Optional: fail the build
  # exit 1
fi
```

### Automated reminders

Use a scheduled job (cron, GitHub Actions) to check maintenance:

```yaml
# .github/workflows/maintenance-reminder.yml
name: Maintenance Reminder
on:
  schedule:
    - cron: '0 9 * * 1'  # Every Monday at 9am

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Check maintenance
        run: |
          nocturnal spec maintenance list
```

---

## Maintenance vs Proposals

| Aspect          | Maintenance              | Proposals                      |
|-----------------|--------------------------|--------------------------------|
| **Purpose**     | Recurring operations     | Feature development            |
| **Lifecycle**   | Permanent, repeating     | Temporary, completed once      |
| **Tracking**    | Frequency-based due dates| Task checkboxes                |
| **Location**    | `spec/maintenance/`      | `spec/proposal/`               |
| **Structure**   | Simple requirement list  | Three documents (spec/design/impl)|
| **Activation**  | Always accessible        | One active at a time           |
| **Completion**  | Mark actioned, resets    | Archive and promote            |

**When to use maintenance:**
- Recurring operational tasks
- Periodic checks and updates
- Time-based requirements
- Continuous operational hygiene

**When to use proposals:**
- New features or changes
- One-time implementations
- Complex development efforts
- Requirements that lead to archived specifications

---

## FAQ

### Can I have multiple maintenance items with the same requirement?

Yes, but it's better to organize requirements into logical groupings. If the same action applies to multiple domains, consider whether it should be one item or multiple.

### What happens if I delete a maintenance file but keep state in .nocturnal.json?

The state is harmless and will be ignored. You can manually edit `.nocturnal.json` to remove orphaned state, or leave it.

### Can I manually edit .nocturnal.json to change last-actioned timestamps?

Yes, but use the `maintenance actioned` command instead for safety. Manual edits should use RFC3339 format: `2026-01-19T10:15:00Z`.

### How do I handle maintenance that doesn't fit a regular frequency?

Omit the `[freq=...]` tag to make it always due. This is useful for:
- One-time migration tasks
- Items pending investigation
- Tasks that should always be visible until completed

### Can I have sub-items or nested requirements?

No, the current implementation only supports flat lists. Use clear requirement text and group related items under one maintenance slug.

### What's the difference between maintenance and CI/CD checks?

- **Maintenance** tracks operational tasks that may require human decision-making or context
- **CI/CD checks** are automated tests/lints that block merges

Maintenance items might trigger CI/CD checks (e.g., "update dependencies" → triggers tests), but they represent higher-level operational work.

---

## See Also

- [Proposal Management](./proposal.md) - Feature development workflow
- [Rule Management](./rule.md) - Project-wide constraints
- [MCP Server](./mcp.md) - Agent integration and automation
