# Persistent Memory in AI-Assisted Development

Software projects accumulate knowledge that lives nowhere in the code. The reason behind an architectural decision, the gotcha that burned two hours of debugging, the convention the team agreed on verbally but never wrote down. When an AI assistant starts a fresh session, all of that context evaporates. The assistant reads the code, infers what it can, and guesses the rest. Sometimes it guesses wrong.

Persistent memory solves this by giving the AI a structured, searchable record of observations across sessions. Not a chat log. Not a brain dump. A curated set of decisions, discoveries, patterns, and conventions that inform future work without polluting the context window.

## Why This Matters More Than You Think

Consider what happens without persistent memory. You spend a session refactoring your authentication layer and explain to the assistant that you chose JWTs over sessions because your deployment is multi-instance. Next week, you ask the assistant to add a new endpoint behind auth. It has no idea about your JWT decision. It might suggest session-based middleware. You correct it. Again.

Now multiply that across every architectural decision, every team convention, every hard-won debugging insight. Without memory, every session starts from zero. The assistant becomes a brilliant amnesiac — capable but perpetually uninformed.

The key insight is that persistent memory transforms the assistant from a stateless tool into a collaborator that accumulates institutional knowledge. It remembers why you chose MongoDB over Postgres, that your CI pipeline breaks if tests take longer than 3 minutes, that the team agreed to use vertical slice architecture with verb-based folders.

## What Memory Is Not

Memory is not documentation. Documentation lives in the codebase and targets humans. Memory lives in the assistant's recall system and targets future AI sessions. The overlap is intentional but the audience is different.

Memory is not a replacement for good commit messages or ADRs. If a decision is significant enough to record, it probably deserves both an ADR for the team and a memory observation for the assistant. They serve different retrieval patterns — humans browse docs, the assistant searches memory semantically.

Memory is also not a chat log. Storing raw conversation history is wasteful and noisy. What matters is the distilled observation: what was decided, why, where it applies, and what surprised you.

## The Engram Protocol

This accelerator standardizes on Engram as the persistent memory system. Engram is an MCP server that stores typed observations with semantic search, topic-based upsert, and cross-project scope. The choice was deliberate.

Engram's topic key system is fundamental to how the SDD workflow operates. Each planning phase writes to a stable topic key (`sdd/{change}/proposal`, `sdd/{change}/spec`), and the next phase reads from it. This creates a chain of artifacts that survives context compaction and session boundaries. No other memory system currently offers this upsert-by-key capability natively.

The trade-off is that Engram requires explicit saves — the assistant must be instructed to call `mem_save` at the right moments. This is handled by protocol conventions in the project's CLAUDE.md, which define what to save, when, and with what topic key structure.

## Topic Key Architecture

A topic key is a stable identifier that allows observations to be updated rather than duplicated. When the assistant saves an observation about your authentication architecture with topic key `architecture/auth`, a future observation with the same key replaces the previous one. This keeps memory current without accumulating stale entries.

The accelerator defines a standard topic key schema that all projects follow:

```
architecture/{component}     — Structural decisions (auth, database, caching)
decision/{topic}             — Non-architectural choices (library picks, conventions)
pattern/{name}               — Established patterns (naming, error handling, testing)
bugfix/{issue}               — Root cause analyses worth remembering
discovery/{topic}            — Non-obvious findings about the codebase or tools
config/{what}                — Environment and tooling configuration
sdd/{change}/{phase}         — SDD workflow artifacts (proposal, spec, design, tasks)
```

This schema is deliberately flat. Two levels of nesting is enough to be navigable without becoming a taxonomy exercise. The first segment is the category, the second is the topic. No deeper.

## The Bootstrap Problem

When a project is born from this accelerator, the assistant's memory for that project is empty. The first session is the most important one to get right, because every observation saved during bootstrap becomes the foundation for all future sessions.

The bootstrap protocol defines what to save in the first session: the stack choice, the architectural style, the testing strategy, the deployment model, and any constraints that would not be obvious from reading the code alone. These observations use the standard topic keys and establish the memory's initial shape.

This is not busywork. A well-bootstrapped memory means the second session starts with context that would otherwise take 15 minutes of re-explanation. By the fifth session, the assistant knows the project's conventions, quirks, and boundaries as well as a team member who has been on the project for weeks.

## What to Save and What to Skip

The protocol is opinionated about what belongs in memory and what does not.

**Save**: architectural decisions with rationale, team conventions not obvious from code, bug root causes that reveal system behavior, tool and library choices with trade-offs, deployment constraints, performance boundaries, anything that made you say "I wish the assistant already knew this."

**Skip**: anything derivable from reading the current code, git history and blame data, debugging steps that led to a fix (save the root cause instead), anything already in CLAUDE.md or documentation, ephemeral task state that only matters for the current session.

The goal is a memory that is small, current, and high-signal. A hundred precise observations are worth more than a thousand noisy ones.
