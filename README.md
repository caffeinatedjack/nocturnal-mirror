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

# Why I made this?

Like a lot of developers, I’ve been experimenting with AI in development, trying to keep things consistent, reduce context switching, and improve quality. I’ve tried out a few spec-driven tools like Speckit, but honestly, I think specs should still be something people drive, not the AI. That’s why I built this tool. You can use AI to help with writing specs if you want, but the real creation process is still up to you.
I also added a persistent to-do manager and a documentation tool, since they fit well with how this tool works.

# License
MIT
