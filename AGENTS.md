# AGENTS.md

## Overview

This document indexes the AI agent definitions available in this repository. Each agent is a specialized assistant with a specific role, ground-truth documents, and responsibilities. External tools (Zed, MCP servers, GitHub Copilot, or other AI assistants) should reference this file to understand which agent to invoke for a given task.

Agent definitions live in `.github/agents/`.

---

## Agent Index

| Agent | File | Role | When to Use |
|-------|------|------|-------------|
| **coding-assistant** | `.github/agents/coding-assistant.md` | Senior engineer for code changes | Implementing features, fixing bugs, refactoring |
| **CONTEXT-maintainer** | `.github/agents/CONTEXT-maintainer.md` | Architecture documentation | Updating `CONTEXT.md` after structural changes |
| **README-maintainer** | `.github/agents/README-maintainer.md` | User-facing documentation | Updating `README.md` for new features or usage |
| **VENDOR-maintainer** | `.github/agents/VENDOR-maintainer.md` | Dependency documentation | Updating `VENDOR.md` when dependencies change |
| **AGENTS-maintainer** | `.github/agents/AGENTS-maintainer.md` | Agent index maintenance | Updating this file when agents change |

---

## Agent Details

### coding-assistant

- **File**: `.github/agents/coding-assistant.md`
- **Role**: Senior software engineer and coding agent
- **Responsibilities**:
  - Implement features, fix bugs, and refactor code
  - Follow architecture and conventions from `CONTEXT.md`
  - Prefer vendor utilities from `VENDOR.md` over custom implementations
  - Write tests following project patterns
- **Ground Truth**:
  1. `CONTEXT.md` — Architecture and conventions
  2. `README.md` — Project purpose and positioning
  3. `VENDOR.md` — Approved vendor libraries and patterns
- **When to Call**: For any code implementation, modification, or review task

---

### CONTEXT-maintainer

- **File**: `.github/agents/CONTEXT-maintainer.md`
- **Role**: Senior architect and context engineer
- **Responsibilities**:
  - Create and maintain `CONTEXT.md`
  - Document architecture, directory structure, and conventions
  - Ensure `CONTEXT.md` reflects actual codebase state
  - Optimize for signal per token (concise, accurate, no fluff)
- **Ground Truth**:
  - Actual repository structure and code
  - Existing `CONTEXT.md` (to update, not replace blindly)
- **When to Call**: After architectural changes, new patterns, or structural refactors

---

### README-maintainer

- **File**: `.github/agents/README-maintainer.md`
- **Role**: Documentation specialist for user-facing docs
- **Responsibilities**:
  - Create and maintain `README.md`
  - Ensure accuracy with actual codebase
  - Keep consistent with `CONTEXT.md` architecture
  - Write for humans (developers, contributors, users)
- **Ground Truth**:
  1. `CONTEXT.md` — Architectural source of truth
  2. Actual repository structure and code
- **When to Call**: After new features, changed commands, or usage updates

---

### VENDOR-maintainer

- **File**: `.github/agents/VENDOR-maintainer.md`
- **Role**: Vendor documentation specialist
- **Responsibilities**:
  - Create and maintain `VENDOR.md`
  - Document external libraries and their usage patterns
  - Guide agents and developers to prefer existing vendors
  - Track dependency versions and migration guidance
- **Ground Truth**:
  1. `CONTEXT.md` — Architecture boundaries
  2. `README.md` — Project positioning
  3. Vendor official documentation
  4. `go.mod` / package manifests
- **When to Call**: After adding, removing, or upgrading dependencies

---

### AGENTS-maintainer

- **File**: `.github/agents/AGENTS-maintainer.md`
- **Role**: Agent orchestrator and index maintainer
- **Responsibilities**:
  - Maintain this `AGENTS.md` file
  - Keep agent index in sync with `.github/agents/` files
  - Document agent collaboration patterns
- **Ground Truth**:
  1. `.github/agents/*.md` — Actual agent definitions
  2. `CONTEXT.md`, `README.md`, `VENDOR.md` — Project docs
- **When to Call**: After adding, modifying, or removing agent definitions

---

## Agent Collaboration

Agents work together to maintain documentation consistency:

```
┌─────────────────────────────────────────────────────────────────┐
│                     coding-assistant                             │
│              (implements code changes)                           │
└──────────────┬──────────────────────────────────────────────────┘
               │ triggers updates to
               ▼
┌──────────────────────────────────────────────────────────────────┐
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │CONTEXT-maintainer│  │README-maintainer │  │VENDOR-maintainer │  │
│  │ (architecture)  │  │ (user docs)     │  │ (dependencies)  │  │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘  │
│           │                    │                    │           │
│           ▼                    ▼                    ▼           │
│      CONTEXT.md            README.md            VENDOR.md       │
└──────────────────────────────────────────────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────────────────────────────┐
│                     AGENTS-maintainer                            │
│              (updates AGENTS.md when agents change)              │
└──────────────────────────────────────────────────────────────────┘
```

### Workflow Example

1. **Feature implementation**: `coding-assistant` adds a new tool
2. **Architecture update**: `CONTEXT-maintainer` updates `CONTEXT.md` with new patterns
3. **User docs update**: `README-maintainer` updates `README.md` with usage instructions
4. **Dependency added**: `VENDOR-maintainer` updates `VENDOR.md` if new library used

---

## Document Hierarchy

When there are conflicts between documents:

| Concern | Authoritative Source |
|---------|---------------------|
| Architecture, conventions, patterns | `CONTEXT.md` |
| User-facing description, positioning | `README.md` |
| Vendor API details, integration | `VENDOR.md` |
| Agent definitions, capabilities | `.github/agents/*.md` |

---

## For External Tools

### Zed Agent / MCP Servers

To use these agents:

1. Agent definitions are in `.github/agents/`
2. Each `.md` file contains the full system prompt for that agent
3. Select the appropriate agent based on the task (see Agent Index above)
4. The agent will reference `CONTEXT.md`, `README.md`, and `VENDOR.md` as needed

### Adding New Agents

1. Create a new `.md` file in `.github/agents/`
2. Define the agent's role, responsibilities, and ground-truth documents
3. Run `AGENTS-maintainer` to update this index

---

## Related Documentation

- [CONTEXT.md](CONTEXT.md) — Architecture and conventions
- [README.md](README.md) — User-facing documentation
- [VENDOR.md](VENDOR.md) — Vendor library documentation
