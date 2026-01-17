![NOCTURNAL](docs/nocturnal.png)

A CLI tool for spec-driven development and agent tooling.

## Features

### Specification Management

Create and manage structured specifications for your project. Initialize a workspace, then use proposals to develop new features through a defined lifecycle:

- **Proposals** - Draft changes with specification, design, and implementation documents
- **Validation** - Check proposals against documentation guidelines before completion
- **Promotion** - Complete proposals to archive designs and promote specs to the main section

## Proposal Lifecycle
```
  ┌─────────┐      ┌─────────┐      ┌──────────┐
  │   ADD   │ ───► │ DEVELOP │ ───► │ COMPLETE │
  └─────────┘      └─────────┘      └──────────┘
       │                │                 │
       ▼                ▼                 ▼
   proposal/        Edit spec,       ┌────────────────┐
   <slug>/          design,          │ archive/       │
   created          impl docs        │   design.md    │
                                     │   impl.md      │
                                     ├────────────────┤
                                     │ section/       │
                                     │   <slug>.md    │
                                     │   (promoted)   │
                                     └────────────────┘
```
Read the [docs](/docs/index.md) for the cli commands.

## Installation

```bash
make build
make install
```
Alternatively you can download the executable from the artifacts.

