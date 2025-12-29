# Rule Management

Rules are project-wide guidelines, constraints, and standards that persist across all proposals. Unlike proposals which have a lifecycle, rules are permanent and apply to all development work.

## Overview

Rules define:
- **Coding Standards** - Style guides, naming conventions, patterns
- **Architectural Constraints** - Technology choices, design patterns
- **Business Rules** - Domain logic, validation rules
- **Security Requirements** - Authentication, authorization, data handling
- **Testing Standards** - Coverage requirements, test patterns

Rules are stored as individual markdown files in `spec/rule/` and are automatically included when agents request project context.

## Commands

### spec rule add

Create a new rule document.

```bash
nocturnal spec rule add <rule-name>
```

**Arguments:**
- `<rule-name>` - Name of the rule (converted to a slug)

**What it does:**
- Creates `spec/rule/<slug>.md` file
- Generates a template with the rule name
- Uses the same slug conversion as proposals (lowercase, hyphenated)

**Slug examples:**
- "Naming Conventions" → `naming-conventions.md`
- "API Design Standards" → `api-design-standards.md`
- "Error_Handling" → `error-handling.md`

**Example:**
```bash
nocturnal spec rule add naming-conventions
```

**Output:**
```
Created rule 'naming-conventions'
Location: spec/rule/naming-conventions.md
```

**Template structure:**
```markdown
# Naming Conventions

<!-- Describe the rule and its purpose -->

## Rationale

<!-- Why this rule exists -->

## Guidelines

<!-- Specific guidelines and examples -->

## Examples

<!-- Good and bad examples -->
```

**When to create rules:**
- Project initialization (define core standards)
- When a pattern emerges across multiple proposals
- When team agreements need documentation
- Before starting complex proposals (establish constraints)

---

### spec rule show

Display all project rules.

```bash
nocturnal spec rule show
```

**What it displays:**
- Count of total rules
- Full content of each rule file
- Separator lines between rules

**Output format:**
```
Rules (3)

# Naming Conventions

All functions must use camelCase...

---

# API Design Standards

REST endpoints must follow...

---

# Error Handling

All errors must be wrapped with context...
```

**Purpose:**
- Review all project standards at once
- Verify rule consistency
- Share with team members
- Reference during code reviews

**No rules scenario:**
```bash
nocturnal spec rule show
```

**Output:**
```
No rules found
Use 'nocturnal spec rule add <rule-name>' to add a rule
```

---

## Rule Content Guidelines

### Structure

A well-written rule should include:

1. **Clear Title** - Descriptive heading (# Rule Name)
2. **Purpose Statement** - What this rule governs
3. **Rationale** - Why the rule exists
4. **Specific Guidelines** - Clear, actionable requirements
5. **Examples** - Both correct and incorrect usage
6. **Exceptions** - When the rule doesn't apply

### Normative Language

Rules should use normative keywords:
- **MUST** - Absolute requirement
- **MUST NOT** - Absolute prohibition
- **SHOULD** - Recommended but not required
- **SHOULD NOT** - Not recommended but not prohibited
- **MAY** - Optional

### Example Rule

```markdown
# API Error Responses

All API endpoints MUST return consistent error response formats.

## Rationale

Consistent error responses allow clients to handle errors uniformly,
reducing integration complexity and improving user experience.

## Response Format

All error responses MUST include:
- `status`: HTTP status code (number)
- `error`: Error type (string)
- `message`: Human-readable description (string)
- `details`: Additional context (object, optional)

## Example

### Correct

```json
{
  "status": 404,
  "error": "NotFound",
  "message": "User with ID 12345 not found",
  "details": {
    "resource": "user",
    "id": "12345"
  }
}
```

### Incorrect

```json
{
  "error": "User not found"
}
```

## Exceptions

Health check endpoints MAY return simplified responses.
```

---

## Rule Categories

### Coding Standards

Define code style and organization:
- Naming conventions (variables, functions, classes)
- File organization and module structure
- Comment and documentation requirements
- Code formatting (handled by automated tools)

**Example:**
```bash
nocturnal spec rule add naming-conventions
nocturnal spec rule add file-organization
nocturnal spec rule add documentation-standards
```

### Architectural Constraints

Establish system design principles:
- Technology stack (languages, frameworks, libraries)
- Design patterns (MVC, repository pattern, etc.)
- Communication protocols (REST, gRPC, message queues)
- Data storage patterns

**Example:**
```bash
nocturnal spec rule add technology-stack
nocturnal spec rule add architecture-patterns
nocturnal spec rule add api-design
```

### Security Requirements

Define security standards:
- Authentication mechanisms
- Authorization patterns
- Data encryption requirements
- Input validation and sanitization
- Secure communication

**Example:**
```bash
nocturnal spec rule add authentication
nocturnal spec rule add data-protection
nocturnal spec rule add input-validation
```

### Testing Standards

Establish testing requirements:
- Test coverage minimums
- Testing strategies (unit, integration, e2e)
- Test naming conventions
- Mock and fixture patterns

**Example:**
```bash
nocturnal spec rule add test-coverage
nocturnal spec rule add test-patterns
```

---

## Rules vs Proposals

| Aspect         | Rules                  | Proposals                      |
|----------------|------------------------|--------------------------------|
| **Lifecycle**  | Permanent              | Temporary (completed)          |
| **Scope**      | Project-wide           | Feature-specific               |
| **Purpose**    | Guidelines/constraints | Feature implementation         |
| **Location**   | `spec/rule/`           | `spec/proposal/`               |
| **Documents**  | Single markdown file   | Three files (spec/design/impl) |
| **Activation** | Always active          | One at a time                  |

---

## Integration with Agent Commands

Rules are automatically included in agent context:

```bash
nocturnal agent project
```

**Output includes:**
```markdown
# Rules

# Naming Conventions
[full rule content]

# API Design Standards
[full rule content]

---

# Project Design
[project.md content]
```

This ensures AI coding agents always have project standards in context when making changes.

---

## Working with Rules

### Creating Rules

1. **Identify the need** - Repeated patterns or team discussions
2. **Add the rule** - `nocturnal spec rule add <name>`
3. **Write clear guidelines** - Use normative language
4. **Provide examples** - Show good and bad practices
5. **Review with team** - Ensure agreement

### Updating Rules

Rules are just markdown files - edit directly:

```bash
$EDITOR spec/rule/naming-conventions.md
```

**When to update:**
- Standards evolve
- Exceptions are discovered
- Clarification needed
- Technology changes

### Removing Rules

Delete the file manually if a rule is no longer applicable:

```bash
rm spec/rule/obsolete-rule.md
```

**Consider instead:**
- Adding an "Obsolete" note at the top
- Moving to an archive directory
- Updating with new approach

---

## Rule Template

```markdown
# [Rule Name]

[Brief description of what this rule governs]

## Rationale

[Why this rule exists - the problem it solves]

## Guidelines

[Specific requirements using MUST/SHOULD/MAY]

1. Requirement one
2. Requirement two
3. Requirement three

## Examples

### Correct Usage

[Example demonstrating the rule]

### Incorrect Usage

[Example violating the rule]

## Exceptions

[When this rule doesn't apply, if any]

## References

[Links to external style guides, RFCs, or documentation]
```

