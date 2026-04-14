# Skill Registry

**Delegator use only.** Any agent that launches sub-agents reads this registry to resolve compact rules, then injects them directly into sub-agent prompts. Sub-agents do NOT read this registry or individual SKILL.md files.

See `_shared/skill-resolver.md` for the full resolution protocol.

## User Skills

| Trigger | Skill | Path |
|---------|-------|------|
| When writing Go tests, using teatest, or adding test coverage | go-testing | ~/.claude/skills/go-testing/SKILL.md |
| When user asks to create a new skill, add agent instructions, or document patterns for AI | skill-creator | ~/.claude/skills/skill-creator/SKILL.md |
| When user says "judgment day", "dual review", "doble review", "juzgar" | judgment-day | ~/.claude/skills/judgment-day/SKILL.md |
| When creating a pull request, opening a PR, or preparing changes for review | branch-pr | ~/.claude/skills/branch-pr/SKILL.md |
| When creating a GitHub issue, reporting a bug, or requesting a feature | issue-creation | ~/.claude/skills/issue-creation/SKILL.md |

## Compact Rules

Pre-digested rules per skill. Delegators copy matching blocks into sub-agent prompts as `## Project Standards (auto-resolved)`.

### go-testing
- Always use table-driven tests with named subtests (`t.Run(tt.name, ...)`)
- Use `t.Helper()` in test helpers for correct error line reporting
- For Bubbletea TUI: use `teatest.NewModel()`, send messages via `tm.Send()`, assert with `teatest.WaitFor()`
- Golden files: use `teatest.RequireEqualOutput()` with `-update` flag for regeneration
- Integration tests: use build tags `//go:build integration` to separate from unit tests
- Never mock interfaces you don't own — wrap them first

### skill-creator
- Skills live in `skills/{skill-name}/SKILL.md` with YAML frontmatter (name, description with Trigger:, license)
- Structure: Critical Patterns (actionable rules) > Examples (minimal) > Anti-patterns (what NOT to do)
- Keep skills focused on ONE concern — split if it covers multiple domains
- Trigger line in description must be specific enough for auto-loading
- Include decision trees for complex choices, not just rules

### judgment-day
- Launch TWO independent blind judge sub-agents in parallel — neither knows about the other
- Resolve skills from registry BEFORE launching judges — inject compact rules into both
- Synthesize verdicts: deduplicate, classify (CRITICAL/WARNING/SUGGESTION), assign to fix agent
- Max 2 iterations: if both judges don't pass after 2 rounds, escalate to user
- Fix agent gets combined findings — never fix between judge rounds without synthesis

### branch-pr
- Every PR MUST link an approved issue (with `status:approved` label)
- Every PR MUST have exactly one `type:*` label
- Branch naming: `^(feat|fix|chore|docs|style|refactor|perf|test|build|ci|revert)\/[a-z0-9._-]+$`
- Conventional commits required — match branch type to commit prefix
- Run shellcheck on modified scripts before opening PR

### issue-creation
- MUST use a template (bug report or feature request) — blank issues disabled
- Every issue gets `status:needs-review` automatically on creation
- A maintainer MUST add `status:approved` before any PR can be opened
- Questions go to Discussions, not issues
- Search existing issues for duplicates before creating

## Project Conventions

| File | Path | Notes |
|------|------|-------|
| CLAUDE.md | /Users/eleazar/ProjectAccelerator/CLAUDE.md | Project instructions — narrative docs style, no project-specific refs in universal rules |

Read the convention files listed above for project-specific patterns and rules. All referenced paths have been extracted — no need to read index files to discover more.
