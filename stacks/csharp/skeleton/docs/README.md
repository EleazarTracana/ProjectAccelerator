# Project Documentation

This directory is the home for all project documentation that does not live in code comments or cursor rules. The goal is not comprehensive documentation of every class and method — the code and tests handle that. The goal is to capture the decisions, context, and guides that a developer needs when the code alone is not enough.

## Structure

- **`adl/`** — Architecture Decision Log. Each ADR captures a significant architectural choice: what was decided, why, what alternatives were considered, and what trade-offs were accepted. Use the MADR format from the ADR template.

## When to Add Documentation

| Situation | What to write | Where |
|-----------|--------------|-------|
| Significant architectural choice | ADR | `docs/adl/` |
| Non-obvious deployment or environment setup | Operations guide | `docs/` |
| Complex domain workflow that spans multiple features | Domain guide | `docs/` |
| API integration with external systems | Integration guide | `docs/` |

Do not document things that are obvious from the code. Do not create documentation preemptively for things that do not exist yet. Write documentation when its absence would cost the next developer real time.
