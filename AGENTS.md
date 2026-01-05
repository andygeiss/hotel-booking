# AGENTS.md

## Overview

This repository uses specialized AI agents defined in `.github/agents/` to maintain documentation, code quality, and consistency. Each agent has a specific responsibility and operates according to ground-truth documents.

This file serves as an index for AI tools (Zed, VS Code Copilot, MCP servers, or other assistants) to understand which agent to invoke for a given task.

---

## Agent Definitions

All agent definition files are located in `.github/agents/`.

| Agent | File | Role |
|-------|------|------|
| coding-assistant | `.github/agents/coding-assistant.md` | Implements code changes following architecture and conventions |
| CONTEXT-maintainer | `.github/agents/CONTEXT-maintainer.md` | Maintains `CONTEXT.md` with accurate architectural documentation |
| README-maintainer | `.github/agents/README-maintainer.md` | Maintains `README.md` as the human-first project introduction |
| VENDOR-maintainer | `.github/agents/VENDOR-maintainer.md` | Maintains `VENDOR.md` with vendor library documentation |
| AGENTS-maintainer | `.github/agents/AGENTS-maintainer.md` | Maintains this file (`AGENTS.md`) |

---

## Agent Responsibilities

### coding-assistant

**Purpose:** Senior software engineer that implements and evolves code while strictly following documented architecture, conventions, and vendor-usage rules.

**When to use:**
- Implementing new features or fixing bugs
- Refactoring existing code
- Adding new bounded contexts or adapters
- Creating tests
- Any code modification task

**Ground truth documents:**
1. `CONTEXT.md` — architecture, conventions, directory structure
2. `README.md` — project purpose and positioning
3. `VENDOR.md` — approved libraries and usage patterns

**Key behaviors:**
- Reads `CONTEXT.md` before significant work
- Prefers reusing vendor utilities over custom implementations
- Follows existing patterns and naming conventions
- Updates documentation when architecture evolves

---

### CONTEXT-maintainer

**Purpose:** Senior software architect that creates and maintains `CONTEXT.md` as the authoritative architectural reference for AI agents and developers.

**When to use:**
- Architecture changes that affect project structure
- New conventions or coding standards
- New bounded contexts or layers added
- Directory structure modifications
- After significant refactoring

**Ground truth documents:**
1. Actual codebase structure and code
2. `README.md` — for alignment on project purpose
3. `VENDOR.md` — for technology stack accuracy

**Key behaviors:**
- Scans repository to discover actual structure
- Documents only what actually exists (no invention)
- Optimizes for signal per token
- Maintains rules and contracts for new code

---

### README-maintainer

**Purpose:** Documentation specialist that maintains `README.md` as the human-first introduction to the project.

**When to use:**
- New features that affect usage
- Changed setup or installation steps
- Updated commands or workflows
- Project description or positioning changes
- After `CONTEXT.md` updates that affect human-facing docs

**Ground truth documents:**
1. Actual codebase and working commands
2. `CONTEXT.md` — architectural truth (README must not contradict)
3. Configured CI/CD and tooling

**Key behaviors:**
- Verifies every claim against actual code
- Tests all documented commands
- Aligns with `CONTEXT.md` on architecture
- Avoids marketing fluff; focuses on accuracy

---

### VENDOR-maintainer

**Purpose:** Vendor documentation specialist that maintains `VENDOR.md` with guidance on external libraries and their usage patterns.

**When to use:**
- Adding or removing vendor dependencies
- New integration patterns with existing vendors
- Version upgrades with breaking changes
- Clarifying when to use which vendor
- Deprecating vendor usage patterns

**Ground truth documents:**
1. `CONTEXT.md` — architecture and layering constraints
2. `README.md` — project positioning
3. Actual dependency manifests (`go.mod`, etc.)
4. Vendor official documentation

**Key behaviors:**
- Documents approved libraries and patterns
- Discourages duplicate implementations
- Links vendor docs to project structure
- Maintains "Vendors to Avoid" guidance

---

### AGENTS-maintainer

**Purpose:** Agent orchestrator that maintains `AGENTS.md` as the index of all agents and their collaboration patterns.

**When to use:**
- New agent definitions added
- Agent responsibilities change
- Agent collaboration patterns evolve
- External tools need agent discovery

**Ground truth documents:**
1. `.github/agents/*.md` — actual agent definitions
2. `CONTEXT.md` — architecture and conventions
3. `README.md` — project description
4. `VENDOR.md` — vendor usage rules

**Key behaviors:**
- Keeps `AGENTS.md` in sync with actual agent files
- Documents when to use each agent
- Describes agent collaboration patterns
- Does not invent agents that don't exist

---

## Agent Collaboration

Agents work together to maintain repository consistency:

```
┌─────────────────────────────────────────────────────────────┐
│                    coding-assistant                         │
│            (implements changes, follows rules)              │
└─────────────────────────┬───────────────────────────────────┘
                          │ triggers doc updates
        ┌─────────────────┼─────────────────┐
        ▼                 ▼                 ▼
┌───────────────┐ ┌───────────────┐ ┌───────────────┐
│   CONTEXT-    │ │    README-    │ │    VENDOR-    │
│  maintainer   │ │   maintainer  │ │   maintainer  │
│ (architecture)│ │ (human intro) │ │  (libraries)  │
└───────┬───────┘ └───────┬───────┘ └───────┬───────┘
        │                 │                 │
        └─────────────────┼─────────────────┘
                          ▼
                  ┌───────────────┐
                  │    AGENTS-    │
                  │   maintainer  │
                  │  (this index) │
                  └───────────────┘
```

### Typical Workflows

**Adding a new feature:**
1. `coding-assistant` implements the feature
2. `CONTEXT-maintainer` updates architecture docs if structure changes
3. `README-maintainer` updates usage docs if user-facing behavior changes
4. `VENDOR-maintainer` updates vendor docs if new dependencies added

**Adding a new vendor dependency:**
1. `coding-assistant` adds dependency and integration code
2. `VENDOR-maintainer` documents the library and usage patterns
3. `CONTEXT-maintainer` updates technology stack if significant
4. `README-maintainer` updates installation if user action required

**Architectural refactoring:**
1. `coding-assistant` implements the refactoring
2. `CONTEXT-maintainer` updates directory structure and conventions
3. `README-maintainer` aligns human-facing docs with new structure
4. `AGENTS-maintainer` updates if agent responsibilities change

---

## Document Hierarchy

When conflicts arise between documents:

| Concern | Authoritative Document |
|---------|----------------------|
| Architecture, conventions, directory structure | `CONTEXT.md` |
| Human-facing description, setup, usage | `README.md` |
| Vendor APIs, capabilities, integration details | `VENDOR.md` |
| Agent definitions and behaviors | `.github/agents/*.md` |

`CONTEXT.md` takes precedence for architectural decisions. `README.md` takes precedence for human-facing messaging. `VENDOR.md` clarifies vendor usage but must not contradict the others.

---

## For External Tools

### Discovery

Agent definitions are located at:
```
.github/agents/
├── AGENTS-maintainer.md
├── CONTEXT-maintainer.md
├── README-maintainer.md
├── VENDOR-maintainer.md
└── coding-assistant.md
```

### Invocation

To use an agent:
1. Read the agent's definition file from `.github/agents/`
2. Load the agent's ground-truth documents
3. Apply the agent's workflow and constraints
4. Produce output according to the agent's output rules

### Adding New Agents

To add a new agent:
1. Create `.github/agents/<agent-name>.md` with:
   - Purpose and core identity
   - Ground-truth documents
   - Workflow and responsibilities
   - Output rules
2. Run `AGENTS-maintainer` to update this index
