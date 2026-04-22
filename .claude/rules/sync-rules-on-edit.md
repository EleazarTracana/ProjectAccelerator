---
paths:
  - .claude/rules/**/*.md
  - .claude/skills/**/SKILL.md
  - CLAUDE.md
description: When Claude-side rules or CLAUDE.md always-apply content changes, mirror into the Cursor rule set
---

# Auto-sync Claude Rules → Cursor Rules (Claude Code)

This repo maintains BOTH `.cursor/rules/` (for Cursor) AND `.claude/rules/` + `.claude/skills/` + `CLAUDE.md` (for Claude Code). Both sides MUST stay in sync because the team uses both editors.

**Scope:** Claude Code agents and humans edit **only** the Claude side (`.claude/rules`, `.claude/skills`, and the always-apply block in `CLAUDE.md`). This rule fires when **those** files change. It does **not** apply when someone edits `.cursor/rules` elsewhere—that workflow lives in [`.cursor/rules/sync-rules-on-edit.mdc`](.cursor/rules/sync-rules-on-edit.mdc) for Cursor.

This slug is excluded from `scripts/sync-rules.mjs` mirroring (`EXCLUDED_SYNC_SLUGS`) so Cursor and Claude each keep their own copy of these instructions.

## When this rule fires

You created or modified a file under `.claude/rules/**/*.md` or `.claude/skills/**/SKILL.md`, or you changed the always-apply section in `CLAUDE.md` (between `<!-- ALWAYS-APPLY-START -->` and `<!-- ALWAYS-APPLY-END -->`).

## What you MUST do (immediately, in the same turn)

1. Run the sync script from the repo root:

   ```bash
   node scripts/sync-rules.mjs --from claude
   ```

2. Inspect the script output. It will report which Cursor-side files were created/updated (`.mdc` under `.cursor/rules/`, and outputs derived from always-apply rules in `CLAUDE.md`).

3. If you are about to commit, stage the regenerated `.cursor/rules/*.mdc` (and any other outputs) together with your Claude-side edits. They MUST land in the same commit so the two sides never diverge in `main`.

4. Do NOT manually edit files marked with `<!-- generated-from: ... -->`. Edit the Claude-side source instead and re-run the sync.

## Why this matters

Without sync, the team gets ghost behavior: a rule shows up in Claude Code but not in Cursor (or the reverse), so different teammates get different agent guidance for the same code. That is worse than having no rules at all.
