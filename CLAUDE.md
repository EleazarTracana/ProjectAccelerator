# ProjectAccelerator

This repository is a reference accelerator for new projects. It contains design philosophy documentation, cursor rules, skeleton architectures, and CLAUDE.md templates.

## Rules

- All documentation (`.md` files) must follow the narrative writing style defined in `docs-writing-style.mdc`: conceptual prose, trade-offs explicit, "the key insight..." phrasing, no dense bullet lists when paragraphs read better.
- Cursor rules (`.mdc` files) follow their own style: short, focused, one concern per rule, guide judgment not replace it.
- No project-specific references (no "polymarket", "referrals", "OneCall", "PhysicalTherapy") in universal rules or philosophy docs. Use generic examples.
- Template files use `.template` extension to signal "copy and rename."
- Skeletons are structural, not runnable. No `.csproj`, no `setup.py`, no real dependency resolution.

## Structure

- `docs/philosophy/` — Design principles that transcend stack choice
- `.cursor/rules/` — Universal cursor rules (stack-agnostic)
- `stacks/csharp/` — C# cursor rules, skeleton, CLAUDE.md template
- `stacks/python/` — Python cursor rules, skeleton, CLAUDE.md template
- `stacks/frontend/` — Frontend cursor rules
- `templates/` — ADR, docs index, Engram config, and bootstrap templates

## When Editing

When modifying cursor rules, ensure:
1. Valid frontmatter (`description`, and either `globs` or `alwaysApply`)
2. No project-specific references leak into universal rules
3. Examples use generic domain concepts (e.g., "orders", "users", "items")

When modifying philosophy docs, ensure:
1. Each principle has examples from both C# and Python
2. Trade-offs are named, not hidden
3. Narrative flows: problem → choice → rationale → nuances → trade-offs
