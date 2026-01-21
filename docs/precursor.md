# Proposal Precursors (Experimental)

**Status:** Experimental - API and format may change

> ⚠️ **Experimental Feature**: The precursor system is under active development. The manifest format, template syntax, and CLI commands may change in future releases.

Proposal precursors are reusable, parameterized templates for creating proposals. They enable you to package common proposal patterns (database migrations, service creation, API integrations, etc.) and share them across projects or teams.

## Quick Start

```bash
# 1. Create a precursor
nocturnal precursor init my-template --out ./my-template.zip

# 2. Edit the generated precursor.yaml and templates

# 3. Create a proposal from the precursor
nocturnal spec proposal add my-feature --precursor-path ./my-template.zip

# 4. Fill in the generated precursor-answers.yaml

# 5. Regenerate with answers
nocturnal spec proposal add my-feature --precursor-path ./my-template.zip --overwrite
```

## Overview

A precursor is a bundle (directory or zip file) containing:
- **Manifest** (`precursor.yaml`) - Defines inputs and metadata
- **Templates** (optional) - Custom spec/design/implementation templates
- **Third-party docs** (optional) - Bundled API documentation

When creating a proposal from a precursor, you provide values for the defined inputs, and Nocturnal generates customized proposal documents.

## Use Cases

- **Database migrations**: Parameterize source/target databases and tooling
- **Microservice creation**: Template service name, port, dependencies
- **API integrations**: Define provider, auth method, endpoints
- **Infrastructure setup**: Specify cloud provider, region, resources
- **Common patterns**: Any repeatable proposal scenario

## Precursor Structure

### Directory Layout

```
my-precursor/
├── precursor.yaml                      # Manifest (required)
├── templates/                          # Custom templates (optional)
│   ├── specification.md.tmpl
│   ├── design.md.tmpl
│   └── implementation.md.tmpl
└── third/                              # Bundled docs (optional)
    ├── library-a.md
    └── library-b.md
```

### Manifest Format (`precursor.yaml`)

```yaml
version: 1
id: database-migration-template
description: Template for database migration proposals
inputs:
  - key: source_db
    prompt: "Source database type (e.g., PostgreSQL, MySQL)"
    required: true
  - key: target_db
    prompt: "Target database type (e.g., PostgreSQL, MySQL)"
    required: true
  - key: migration_tool
    prompt: "Migration tool to use (e.g., Flyway, Liquibase)"
    required: true
  - key: rollback_strategy
    prompt: "Rollback strategy"
    required: false
```

### Template Files

Templates use Go's `text/template` syntax with access to:
- `{{.Name}}` - Proposal name
- `{{.Slug}}` - Proposal slug
- `{{.Inputs.key_name}}` - Input values from answers

**Example: `templates/specification.md.tmpl`**

```markdown
# Specification: {{.Name}}

## Abstract

This specification defines the migration from {{.Inputs.source_db}} to {{.Inputs.target_db}} 
using {{.Inputs.migration_tool}}.

## Requirements

- The system MUST support migration from {{.Inputs.source_db}} to {{.Inputs.target_db}}
- The migration MUST use {{.Inputs.migration_tool}} as the primary tool
{{if .Inputs.rollback_strategy}}- The system MUST implement {{.Inputs.rollback_strategy}} for rollback
{{end}}
```

**Template Functions Available:**
- `{{if .Inputs.key}}...{{end}}` - Conditional rendering
- `{{range .Inputs.list}}...{{end}}` - Iterate over lists
- Standard Go template functions

If a template is not provided in the precursor, Nocturnal falls back to the embedded default template.

## Workflow

### 1. Create a Precursor

```bash
# Initialize a new precursor (creates directory or zip)
nocturnal precursor init database-migration --out ./db-migration.zip

# Or create as a directory
nocturnal precursor init database-migration --out ./db-migration
```

This creates a scaffold with:
- `precursor.yaml` with sample inputs
- `templates/` directory with example templates
- `third/` directory for documentation

Edit the manifest to define your inputs, and customize the templates as needed.

### 2. Validate Precursor Structure

```bash
# Check that manifest is valid and templates parse correctly
nocturnal precursor validate --path ./db-migration.zip
```

### 3. Create a Proposal from Precursor

```bash
# First attempt - creates questionnaire if inputs are missing
nocturnal spec proposal add migrate-prod --precursor-path ./db-migration.zip
```

**Output:**
```
Proposal 'migrate-prod' created but requires input
Please fill in the following required fields in: spec/proposal/migrate-prod/precursor-answers.yaml

  • source_db: Source database type (e.g., PostgreSQL, MySQL)
  • target_db: Target database type (e.g., PostgreSQL, MySQL)
  • migration_tool: Migration tool to use (e.g., Flyway, Liquibase)

After filling in the answers, run:
  nocturnal spec proposal add migrate-prod --precursor-path ./db-migration.zip --overwrite
```

### 4. Fill in Answers

Edit `spec/proposal/migrate-prod/precursor-answers.yaml`:

```yaml
version: 1
precursor_path: ./db-migration.zip
inputs:
  source_db:
    required: true
    prompt: Source database type (e.g., PostgreSQL, MySQL)
    value: "MySQL"
  target_db:
    required: true
    prompt: Target database type (e.g., PostgreSQL, MySQL)
    value: "PostgreSQL"
  migration_tool:
    required: true
    prompt: Migration tool to use (e.g., Flyway, Liquibase)
    value: "Flyway"
  rollback_strategy:
    required: false
    prompt: Rollback strategy
    value: "Restore from backup snapshot"
```

### 5. Regenerate Proposal with Answers

```bash
# Generate proposal documents with filled inputs
nocturnal spec proposal add migrate-prod --precursor-path ./db-migration.zip --overwrite
```

**Output:**
```
Regenerated proposal 'migrate-prod' from precursor
Location: spec/proposal/migrate-prod/
Precursor: database-migration-template
```

Now your proposal contains fully rendered documents with all template variables substituted.

## Commands

### `nocturnal precursor init <name>`

Initialize a new precursor bundle.

**Flags:**
- `--out <path>` - Output path (required, .zip for zip file, directory otherwise)

**Example:**
```bash
nocturnal precursor init microservice-template --out ./templates/microservice.zip
```

### `nocturnal precursor validate`

Validate a precursor bundle structure and templates.

**Flags:**
- `--path <path>` - Path to precursor (required, directory or .zip)

**Example:**
```bash
nocturnal precursor validate --path ./microservice.zip
```

### `nocturnal precursor pack`

Pack a precursor directory into a zip file.

**Flags:**
- `--in <directory>` - Input directory (required)
- `--out <file.zip>` - Output zip file (required)

**Example:**
```bash
nocturnal precursor pack --in ./my-precursor --out ./my-precursor.zip
```

### `nocturnal precursor unpack`

Unpack a precursor zip into a directory.

**Flags:**
- `--in <file.zip>` - Input zip file (required)
- `--out <directory>` - Output directory (required)

**Example:**
```bash
nocturnal precursor unpack --in ./my-precursor.zip --out ./my-precursor-edit
```

### `nocturnal spec proposal add <name> --precursor-path <path>`

Create a proposal from a precursor bundle.

**Flags:**
- `--precursor-path <path>` - Path to precursor (directory or .zip)
- `--overwrite` - Allow regenerating existing proposal and overwrite conflicting third-party docs

**Behavior:**
1. **First run** (missing inputs): Creates `precursor-answers.yaml` and exits
2. **Subsequent run** (with `--overwrite`): Generates proposal documents from templates
3. **Third-party docs**: Installs bundled docs to `spec/third/`, overwrites if `--overwrite` is set

**Example:**
```bash
# Initial creation (generates questionnaire)
nocturnal spec proposal add api-integration --precursor-path ./templates/api.zip

# Fill in spec/proposal/api-integration/precursor-answers.yaml

# Regenerate with answers
nocturnal spec proposal add api-integration --precursor-path ./templates/api.zip --overwrite
```

## Third-Party Documentation

Precursors can bundle relevant third-party documentation in the `third/` directory. When creating a proposal:

- **Without `--overwrite`**: Skips existing files, shows warning
- **With `--overwrite`**: Replaces existing files with precursor versions

This allows precursors to ship with relevant API docs for the pattern being implemented.

## Template Fallback

If a precursor doesn't provide a specific template (spec, design, or implementation), Nocturnal automatically falls back to the embedded default template. This allows you to:

- Override just the specification template
- Customize only the implementation plan
- Use all default templates with custom inputs

## Advanced: Comma-Separated Lists

The answers-to-template-data converter automatically detects comma-separated values and converts them to arrays for template iteration:

**Answers:**
```yaml
inputs:
  endpoints:
    value: "/users, /posts, /comments"
```

**Template:**
```markdown
## Endpoints
{{range .Inputs.endpoints}}
- {{.}}
{{end}}
```

**Result:**
```markdown
## Endpoints
- /users
- /posts
- /comments
```

## Sharing Precursors

Precursors can be shared as:
- **Zip files**: Self-contained, easy to distribute
- **Git repositories**: Version-controlled, collaborative editing
- **Internal packages**: Stored in artifact repositories

## Limitations (Experimental)

- Templates use basic Go `text/template` syntax (no advanced functions)
- No validation of input values (type checking, regex, etc.)
- Precursor versions not tracked after proposal creation
- No precursor update/migration mechanism
- Third-party doc conflicts require manual resolution

## Best Practices

1. **Use descriptive input keys**: `source_database` not `src_db`
2. **Provide helpful prompts**: Include examples in prompt text
3. **Mark truly required inputs**: Only set `required: true` for essential inputs
4. **Test templates**: Use `precursor validate` before distributing
5. **Document your precursor**: Add a README.md explaining the use case
6. **Version your precursors**: Use git tags or version in ID field
7. **Keep templates focused**: Don't try to handle too many scenarios in one precursor

## Examples

See the [precursor examples repository](#) for ready-to-use precursors:
- Database migration templates
- Microservice creation
- API integration patterns
- Infrastructure as code proposals

## Future Enhancements (Planned)

- Input validation (types, regex, enums)
- Precursor registry/marketplace
- Version tracking and migration
- Interactive input prompts (CLI wizard)
- Dependency between inputs
- More template helper functions

## Design Rationale

The precursor system was designed with the following principles:

1. **Questionnaire-first workflow**: Don't generate partial documents. Create a questionnaire, let users fill it completely, then generate full documents.
2. **Graceful fallback**: Missing templates fall back to embedded defaults, allowing partial customization.
3. **Portable bundles**: Support both directory and zip formats for easy sharing and version control.
4. **Third-party doc bundling**: Enable precursors to ship with relevant API documentation.
5. **Simple manifest**: YAML-based configuration that's easy to read and edit.

This approach balances flexibility with simplicity, making precursors useful for common patterns while keeping the system maintainable.

