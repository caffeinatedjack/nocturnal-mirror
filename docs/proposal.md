# Proposal Management

Proposals are the core workflow mechanism in Nocturnal. Each proposal represents a feature, change, or enhancement being developed, containing specification, design, and implementation documents.

## Overview

The proposal lifecycle:

1. **Add** - Create a new proposal with template documents
2. **Activate** - Set as the current working proposal
3. **Develop** - Edit documents and implement the feature
4. **Validate** - Check documents against guidelines
5. **Complete** - Archive design/implementation and promote specification

## Commands

### spec init

Initialize a specification workspace in the current directory.

```bash
nocturnal spec init
```

**What it does:**
- Creates `spec/` directory structure
- Generates subdirectories: `rule/`, `proposal/`, `archive/`, `section/`
- Copies template files: `project.md`, `AGENTS.md`, and guideline documents
- Sets up the workspace for proposal management

**When to use:**
- First-time setup in a new project
- After cloning a repository that uses Nocturnal

**Output:**
- Success message with workspace location
- Error if workspace already exists

---

### spec view

View an overview of the specification workspace.

```bash
nocturnal spec view
```

**What it displays:**
- **Specifications** - List of completed specs with requirement counts
- **Active Proposal** - Current working proposal with progress bar
- **Other Proposals** - All non-active proposals with completion percentages

**Features:**
- Shows task completion percentage for each proposal
- Displays dependency information (which proposals depend on others)
- Visual progress bar for active proposal
- Requirement counts using normative language (MUST, SHALL)

**Example output:**
```
Specifications

  authentication  (12 requirements)
  data-validation  (8 requirements)

Active Proposal

  rate-limiting  [████████████░░░░░░░░] 60% (6/10 tasks)
  depends on: authentication

Other Proposals

  logging  (25% complete, depends on: rate-limiting)
  metrics  (0% complete)
```

---

### spec proposal add

Create a new proposal with template documents.

```bash
nocturnal spec proposal add <change-slug>
```

**Arguments:**
- `<change-slug>` - Name of the proposal (converted to lowercase with hyphens)

**What it does:**
- Creates `spec/proposal/<slug>/` directory
- Generates three template files:
  - `specification.md` - Requirements document template
  - `design.md` - Design decision template
  - `implementation.md` - Implementation plan template
- Fills templates with proposal name and slug

**Slug conversion:**
- Converts spaces and underscores to hyphens
- Converts to lowercase
- Removes special characters
- Examples:
  - "User Authentication" → "user-authentication"
  - "API_Rate_Limiting" → "api-rate-limiting"

**Example:**
```bash
nocturnal spec proposal add user-authentication
```

**Output:**
```
Created proposal 'user-authentication'
Location: spec/proposal/user-authentication/
```

---

### spec proposal activate

Set a proposal as the currently active one for development.

```bash
nocturnal spec proposal activate <change-slug>
```

**Arguments:**
- `<change-slug>` - Name of the proposal to activate

**What it does:**
- Creates/updates `spec/current` symlink to point to the proposal
- Removes any existing symlink to a different proposal
- Validates that no other proposals depend on this one
- Makes the proposal the default for agent commands

**Dependency check:**
- Prevents activating a proposal if other proposals depend on it
- Shows list of dependent proposals
- Ensures logical development order

**Why symlinks:**
- Provides a stable path for tools/agents to access current work
- Allows quick switching between proposals
- Shell-friendly for `cd spec/current`

**Example:**
```bash
nocturnal spec proposal activate user-authentication
```

**Output:**
```
Activated proposal 'user-authentication'
```

**Error cases:**
- Proposal doesn't exist
- Other proposals depend on it (must complete dependents first)

---

### spec proposal validate

Validate proposal documents against documentation guidelines.

```bash
nocturnal spec proposal validate <change-slug>
```

**Arguments:**
- `<change-slug>` - Name of the proposal to validate

**What it checks:**

**For specification.md:**
- Required sections: Abstract, Introduction, Requirements
- Recommended sections: Examples, Security Considerations, Error Handling
- Use of normative language (MUST/SHOULD/MAY)
- Unfilled template comments

**For design.md:**
- Required sections: Context, Goals and Non-Goals, Options Considered, Decision, Detailed Design, Cross-Cutting Concerns, Implementation Plan
- Recommended sections: Open Questions
- Metadata: Title, Status, Specification Reference
- At least 2 design options documented
- Unfilled template comments

**For implementation.md:**
- Phase structure (Phase 1, Phase 2, etc.)
- Task checkboxes (- [ ] for tracking)
- Unfilled template comments

**Output:**
- ✓ for documents that pass
- ⚠ for warnings (recommended sections missing)
- ✗ for errors (required sections missing)
- Summary with total error and warning counts

**Example:**
```bash
nocturnal spec proposal validate user-authentication
```

**Output:**
```
Validating proposal: user-authentication

✓ specification.md

⚠ design.md
    ⚠ Missing recommended section: Open Questions - List unresolved items

✓ implementation.md

---
Validation complete: 0 error(s), 1 warning(s)
```

---

### spec proposal complete

Complete a proposal, archiving design/implementation and promoting specification.

```bash
nocturnal spec proposal complete <change-slug>
```

**Arguments:**
- `<change-slug>` - Name of the proposal to complete

**What it does:**
1. Validates proposal exists and has specification.md
2. Creates `spec/archive/<slug>/` directory
3. Copies `design.md` and `implementation.md` to archive
4. Copies `specification.md` to `spec/section/<slug>.md`
5. Removes the proposal directory
6. Clears the `current` symlink if this was the active proposal

**Archive structure:**
```
spec/archive/user-authentication/
├── design.md           # Historical design decisions
└── implementation.md   # Completed implementation tasks
```

**Promoted specification:**
```
spec/section/user-authentication.md  # Now part of the main spec
```

**Why this workflow:**
- Specifications become permanent project requirements
- Design decisions are preserved for historical reference
- Implementation tasks are archived but not needed for ongoing development
- Keeps proposal directory clean for active work

**Example:**
```bash
nocturnal spec proposal complete user-authentication
```

**Output:**
```
Completed proposal 'user-authentication'
Specification promoted to section/user-authentication.md
Design/implementation archived to archive/user-authentication/
```

---

### spec proposal remove

Remove a proposal and its documents.

```bash
nocturnal spec proposal remove <change-slug>
nocturnal spec proposal remove <change-slug> --force
```

**Arguments:**
- `<change-slug>` - Name of the proposal to remove

**Flags:**
- `--force`, `-f` - Remove even if proposal is currently active

**What it does:**
- Deletes the proposal directory and all its documents
- Checks if proposal is active (prevents accidental deletion)
- Removes `current` symlink if this was the active proposal

**Safety features:**
- Requires `--force` flag if proposal is active
- Prevents accidental deletion of current work
- Cannot be undone (no recycle bin)

**Example:**
```bash
# Remove inactive proposal
nocturnal spec proposal remove experimental-feature

# Remove active proposal (dangerous!)
nocturnal spec proposal remove user-authentication --force
```

**Output:**
```
Removed proposal 'user-authentication'
```

**Error cases:**
- Proposal doesn't exist
- Proposal is active and --force not provided

---

## Proposal Document Templates

### specification.md

Contains:
- Metadata (title, status, depends on)
- Abstract (2-4 sentence summary)
- Introduction (context and motivation)
- Requirements (using MUST/SHOULD/MAY)
- Examples (concrete usage examples)
- Security Considerations
- Error Handling
- Future Extensions

### design.md

Contains:
- Metadata (title, status, specification reference)
- Context (technical landscape)
- Goals and Non-Goals
- Options Considered (at least 2 alternatives)
- Decision (chosen approach and rationale)
- Detailed Design (architecture, components, APIs)
- Cross-Cutting Concerns (security, performance, testing)
- Implementation Plan (phases and milestones)
- Open Questions

### implementation.md

Contains:
- Phased implementation breakdown
- Task checkboxes for tracking (- [ ] incomplete, - [x] complete)
- Dependencies and ordering
- Testing requirements
- Rollout strategy

## Progress Tracking

Nocturnal automatically tracks proposal progress:

1. **Task Counting** - Parses `- [ ]` and `- [x]` from implementation.md
2. **Percentage Calculation** - Completed / Total tasks
3. **Visual Progress Bar** - Shown in `spec view`
4. **Requirement Counting** - Counts MUST/SHALL in specifications

## Dependencies

Proposals can depend on other proposals:

```markdown
**Depends on**: authentication, rate-limiting
```

**Effects:**
- Dependent proposals cannot be activated before dependencies
- Shows dependency chains in `spec view`
- Prevents activation of proposals that others depend on
- Helps maintain logical development order

