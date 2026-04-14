# ProjectAccelerator

Every new project starts with the same question: how do we want to build this? Not which framework or which database — those are implementation details. The real question is about the principles that shape how components compose, how abstractions layer, how the codebase communicates intent to the next person who reads it.

This repository is the answer to that question, distilled from two production systems: a Python/FastAPI analytics platform and a C#/.NET multi-tenant healthcare backend. Both projects evolved independently, yet converged on the same architectural principles — which is how we know those principles are worth codifying.

## What This Is

ProjectAccelerator is a reference repository. It captures a development philosophy, enforces it through cursor rules and AI agent instructions, and provides structural skeletons for new projects. The goal: when starting a new backend (C# or Python) or frontend (React Router 7), copy the relevant pieces and begin with every convention already in place — including AI agents that understand how you build software.

This is not a template generator. There is no CLI, no scaffolding tool, no `npx create-project`. The value is in the ideas and the structure, not in automation. You read the philosophy, copy the skeleton, merge the cursor rules, and adapt the CLAUDE.md template. The manual step is intentional — it forces you to understand what you are adopting.

## Repository Structure

```
ProjectAccelerator/
├── docs/philosophy/                    # The "why" — design principles with examples
├── .cursor/rules/                      # Universal rules (any stack)
├── stacks/
│   ├── csharp/
│   │   ├── .cursor/rules/              # C#-specific rules
│   │   ├── skeleton/                   # Project structure + templates
│   │   └── CLAUDE.md.template          # AI agent instructions
│   ├── python/
│   │   ├── .cursor/rules/              # Python-specific rules
│   │   ├── skeleton/                   # Project structure + templates
│   │   └── CLAUDE.md.template          # AI agent instructions
│   └── frontend/
│       └── .cursor/rules/              # Frontend-specific rules
└── templates/                          # ADR template, docs index template
```

## How to Use This for a New Project

Starting a new project involves three steps, and the order matters.

**First, read the philosophy.** Open `docs/philosophy/design-principles.md` and internalize the ten principles. These are not abstract ideals — they are specific, battle-tested patterns with concrete examples. If a principle does not make sense for your project, skip it deliberately, not accidentally.

**Second, copy the structure.** Pick the skeleton that matches your stack (`stacks/csharp/skeleton/` or `stacks/python/skeleton/`). Copy it into your new repository. Rename the `{Project}` placeholders. Read each `.template` file — they show the pattern, not runnable code. Delete them once you have written your first real feature in that slot.

**Third, merge the rules.** Copy `.cursor/rules/` (universal) into your project's `.cursor/rules/`. Then merge the stack-specific rules from `stacks/{stack}/.cursor/rules/`. Finally, copy the `CLAUDE.md.template`, rename it to `CLAUDE.md`, and customize it for your project.

The trade-off with this manual approach: it takes ten minutes instead of one command. The upside: you understand every convention you are adopting, and you can selectively omit what does not apply. That understanding is worth the ten minutes.

## What Each Piece Does

**`docs/philosophy/design-principles.md`** is the conceptual foundation. It documents ten principles — single decision point, semantic registration, three-layer abstraction, pipeline-first design, pure domain kernels, result types, fail-safe transitions, observable-by-design, and more — with side-by-side examples from C# and Python.

**`.cursor/rules/`** are AI agent rules that auto-inject into Cursor IDE (and are referenced from CLAUDE.md for Claude Code). They enforce conventions at development time: TDD discipline, structured logging, feature vocabulary, pagination patterns, documentation style. Universal rules are stack-agnostic. Stack-specific rules live under `stacks/{stack}/.cursor/rules/`.

**`stacks/{stack}/skeleton/`** is the starting folder structure for a new project. It embodies the architecture — vertical slices, kernel purity, thin routers, orchestrator layer — through directory layout and template files.

**`CLAUDE.md.template`** gives AI agents the context to work within these conventions. When an agent reads your CLAUDE.md, it knows about dependency direction, testing expectations, documentation style, and the feature vocabulary your project uses.

**`templates/`** holds the MADR ADR format and a docs index template. Both are copied into new projects to bootstrap documentation.
