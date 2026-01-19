# Nocturnal Documentation

Nocturnal is a CLI tool for spec-driven development and agent tooling. It helps you manage project specifications, rules, and proposals through a structured workflow, while providing tools for AI coding agents to access project context.

## Quick Start

```bash
# Initialize a specification workspace
nocturnal spec init
# Create a new feature proposal
nocturnal spec proposal add my-feature
# Activate the proposal to work on it
nocturnal spec proposal activate my-feature
# View workspace overview
nocturnal spec view
# Complete and promote the proposal
nocturnal spec proposal complete my-feature
```

## Core Concepts

### Specifications
Formal requirements documents that define what a feature must do. Specifications use normative language (MUST, SHOULD, MAY) and include sections for abstract, introduction, requirements, examples, and considerations.

### Proposals
Development workspaces containing three documents:
- **specification.md** - The formal requirements
- **design.md** - Technical design decisions and architecture
- **implementation.md** - Phased implementation plan with tasks

Active proposals are tracked in `spec/.nocturnal.json`. When activated, file hashes are computed to detect modifications - MCP tools will warn agents if proposal files change, requiring user confirmation before proceeding.

### Rules
Project-wide constraints and guidelines that persist across all proposals. Rules define coding standards, architectural patterns, or business constraints.

### Maintenance
Recurring operational tasks with frequency-based tracking. Maintenance items contain requirements that are marked as due based on their frequency (daily, weekly, monthly, etc.) and last-actioned timestamp. Examples include dependency updates, security audits, and periodic reviews.

### Archive
Completed proposals are archived - their design and implementation documents are preserved for reference, while specifications are promoted to the main section.

## Command Categories

- **[Specification Management](./proposal.md)** - Create and manage proposals through their lifecycle
- **[Rules Management](./rule.md)** - Define project-wide guidelines and constraints
- **[Maintenance Management](./maintenance.md)** - Track recurring operational tasks with frequency-based due dates
- **[MCP Server](./mcp.md)** - Expose tools to AI assistants via Model Context Protocol
- **[Documentation Management](./docs.md)** - Store and search API/library documentation

## Shell Completion

Generate shell completion scripts for faster command entry:

```bash
# Bash
nocturnal completion bash > /etc/bash_completion.d/nocturnal
# Zsh
nocturnal completion zsh > "${fpath[1]}/_nocturnal"
# Fish
nocturnal completion fish > ~/.config/fish/completions/nocturnal.fish
# PowerShell
nocturnal completion powershell > nocturnal.ps1
```

