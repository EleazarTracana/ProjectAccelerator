# Sentinel Shared Protocol — v2.0

Every Sentinel skill reads this file first. Cross-cutting concerns for ALL pipeline agents.

---

## 1. Engram Memory Protocol

### Available Tools

| Tool | Purpose |
|------|---------|
| `mem_save` | Persist observations with topic keys (upserts on same key) |
| `mem_search` | Find prior observations by keyword or topic key |
| `mem_get_observation` | Retrieve FULL untruncated content of an observation (search results are truncated) |
| `mem_context` | Recover recent session history (fast, cheap — use after compaction) |
| `mem_session_summary` | Persist end-of-session summary (used by report phase) |

### On START — Check Prior Work

Before doing any work, search for prior results: `mem_search(query: "{your-topic-key}", project: "{project}")`. If found, call `mem_get_observation(id)` for full content. This gives you prior findings, validation context, and failed approaches to avoid.

### On COMPLETION — Persist Output

After producing output, persist BEFORE returning: `mem_save(title: "{Agent}: {description}", type: "{type}", scope: "project", project: "{project}", topic_key: "{your-topic-key}", content: "<output>")`.

**ALWAYS call `mem_save` before returning.** If engram is unavailable, include full output inline and warn the orchestrator.

### On COMPACTION — Recovery

1. `mem_context` → recover recent session history
2. `mem_search(query: "sdd/{change-name}", project: "{project}")` → find partial work
3. Resume — do NOT restart from scratch

### Topic Key Convention

| Artifact | Topic Key | Owner |
|----------|-----------|-------|
| Product Model | `sdd/{change-name}/product-model` | sentinel-init |
| DB scan results | `sdd/{change-name}/scan-db` | sentinel-scan-db |
| API scan results | `sdd/{change-name}/scan-api` | sentinel-scan-api |
| UI scan results | `sdd/{change-name}/scan-ui` | sentinel-scan-ui |
| Arch scan results | `sdd/{change-name}/scan-arch` | sentinel-scan-arch |
| Council verdict | `sdd/{change-name}/council-{task-id}` | sentinel-council |
| Heal result | `sdd/{change-name}/heal-{finding-id}` | sentinel-heal |
| Apply progress | `sdd/{change-name}/apply-progress` | sdd-apply |
| Verify report | `sdd/{change-name}/verify-report` | sdd-verify |
| Final report | `sdd/{change-name}/report` | sentinel-report |

Use `topic_key` for upsert behavior — same key overwrites previous content. Different findings or tasks use different keys (e.g., `council-task-1` vs `council-task-2`).

---

## 2. Pipeline Overview

```
Phase 1: Initialize
  El Arquelogo (Opus) → reads ALL docs, builds Product Model JSON
  Output: Product Model saved to engram — the SINGLE source of truth for all agents

Phase 2: Per-task loop (dependency order from SDD tasks)
  2a. sdd-apply         → implement task from specs
  2b. Docker deploy     → build + health check
  2c. 4 Scanners IN PARALLEL:
      - DB   (El Minero / Haiku)      → functional data validation via SQL
      - API  (El Rastreador / Sonnet)  → contract validation via HTTP
      - UI   (El Observador / Sonnet)  → visual evidence via Playwright
      - Arch (El Centinela / Opus)     → boundary compliance via static analysis
  2d. Council (3 judges, 2 model families):
      - El Fiscal   (Kimi K2.5)  → adversarial: reasons NOT to fix
      - El Perito   (Sonnet)     → forensic: evidence-only analysis
      - El Guardian (Opus)       → product alignment check
  2e. Verdict routing:
      APPROVED  → commit + save progress + next task
      ITERATE   → heal → redeploy → re-scan → re-council (max from config)
      ESCALATE  → flag for human + save to report + next task
  2f. Heal (El Cirujano / Sonnet) — if ITERATE
  2g. sdd-verify — post-fix verification against specs

Phase 3: Report
  El Cartografo (Haiku) → HTML report + GitHub PR
```

Know WHERE you fit: what ran before you (your input), what runs after you (your output format), who your siblings are (parallel phases).

---

## 3. Common Output Contract

Every skill returns structured JSON. These fields are REQUIRED in every response:

| Field | Type | Description |
|-------|------|-------------|
| `status` | string | `success`, `failure`, `partial`, `blocked`, `escalated`, or role-specific values |
| `schema_version` | string | Must be `"1.0"` — see contract version below |
| `engram_saved` | boolean | Whether the agent successfully persisted its output to engram |
| `skill_resolution` | string | `injected` if project standards were in the prompt; else `fallback-registry`, `fallback-path`, or `none` |

**Current contract version: `1.0`**

All agents MUST include `"schema_version": "1.0"` in their JSON response.
If the orchestrator receives a response with a missing or mismatched `schema_version`, treat it as `status: "failure"` with `error: "schema_version_mismatch"` and escalate.
Do NOT implement version negotiation — all agents share the same protocol version from the same skill files.

### Findings Format

Scanners produce findings. Each finding is a standalone object:

```json
{
  "id": "scan-db-001",
  "severity": "CRITICAL | HIGH | MEDIUM | LOW | SUGGESTION",
  "scanner": "db | api | ui | arch",
  "description": "Human-readable description of what is wrong",
  "evidence": {
    "expected": "What should happen",
    "actual": "What actually happened",
    "reproduction": "Command to reproduce (curl, SQL, file:line)"
  },
  "task_ref": "task-id this finding relates to"
}
```

> The `evidence` block (`expected`, `actual`, `reproduction`) conforms to ISO 29119-3 defect report requirements. Do not remove or merge these fields.

### Severity Levels

| Level | Criteria | Council? |
|-------|----------|----------|
| CRITICAL | Feature non-functional AND reproducible in ≤2 steps AND affects all users OR blocks the task under test | Yes |
| HIGH | Significant defect AND reproducible AND affects most users OR data integrity at risk | Yes |
| MEDIUM | Incorrect behavior, reproducible, not blocking; subset of users or edge path affected | Yes |
| LOW | Cosmetic, rare reproduction, or single-user edge case | Yes |
| SUGGESTION | Improvement opportunity, not a defect | No — straight to report |

Scanners apply these criteria mechanically. When evidence supports two levels, choose the higher one and note the ambiguity in `description`.
Domain-specific names (e.g., `BROKEN`, `INCOHERENT`, `DATA_MISSING`) map to the levels above — the scanner's SKILL.md defines the mapping.

---

## 4. Project-Agnostic Rules

You are a ROLE, not a project expert. All project-specific knowledge comes from external sources.

### Context Sources (priority order)

| Source | Provides | Read by |
|--------|----------|---------|
| `sentinel.config.yaml` | Services, ports, DBs, models, guardrails, blocked paths | Orchestrator (passes to you) |
| Product Model JSON | Stack, architecture, entities, endpoints, ADRs, anti-patterns | Built by Phase 1, distributed to all |
| Engram | Prior results, decisions, discoveries | You (on START) |
| Orchestrator prompt | Task context, iteration number, prior findings | Passed directly |

### Hard Rules

1. **NEVER hardcode project-specific technology.** Your SKILL.md defines your role. The Product Model tells you about the project.
2. **Product Model IS your project knowledge.** Do not assume anything beyond what it contains.
3. **Orchestrator injects context.** Do not search for config or SDD artifacts unless your SKILL.md explicitly instructs it.
4. **Blocked paths are absolute.** ESCALATE if your work requires touching one.
5. **NEVER modify tests to make them pass.** Fix the implementation. If the test is wrong, ESCALATE.
6. **NEVER push to main/master or force-push any branch.**
7. **NEVER write to the Product Model topic key.** It is built once by El Arquelogo in Phase 1 and is read-only for all subsequent agents. If you believe it contains an error, include a `product_model_discrepancy` note in your output and ESCALATE — do not patch it.

---

## 5. Communication Protocol

### Execution Order (every agent follows this)

1. Read your SKILL.md
2. Read this shared protocol
3. Search engram for prior work (On START)
3a. **Verify required inputs exist.** If your SKILL.md declares required inputs (e.g., Council requires ≥1 scanner finding), check engram before executing. If required inputs are missing, return immediately with `status: "blocked"`, list the missing keys in the `error` field, and do NOT attempt to proceed. The orchestrator handles blocked agents — do not improvise.
4. Execute your role-specific steps
5. Save to engram (On COMPLETION) — MANDATORY
6. Return structured JSON to the orchestrator

### Failure Handling

- `status: "failure"` + `error` field — do NOT crash silently
- `status: "partial"` — include what you accomplished and what remains
- `status: "escalated"` — include the specific reason
- ALWAYS save to engram before returning, even on failure

### Emergency Brake

Two levels of failure response:

**Agent-level (isolate, continue)**: If a scanner or healer fails due to tool unavailability, timeout, or internal error, return `status: "failure"` with a full `error` field. The orchestrator skips this agent's contribution and continues the pipeline. The finding is preserved in the report as `agent_failure` with no Council verdict.

**Pipeline-level (full stop)**: STOP and return immediately if you observe:
- Container restart loops (>3 in 5 min) — infrastructure is unstable, proceeding produces garbage
- Same file modified >5 times in the current iteration — fix is diverging
- A fix that makes scan results measurably worse than the pre-fix baseline
- Critically low disk space preventing tool execution

Pipeline-level stops are rare. When in doubt, use agent-level isolation.

---

## 6. Scanner-Specific Conventions

If you are a scanner (DB, API, UI, Arch), these additional rules apply:

- You run AFTER deployment — you validate LIVE runtime behavior (except Arch, which is static)
- You produce a findings array — one object per issue, never merge or summarize
- Every finding MUST include a reproduction command (curl, SQL query, file:line reference)
- You do NOT fix anything — you are a scanner, not a healer
- If no issues found, return an empty findings array `[]`
- Your findings feed the Council — quality of evidence determines quality of verdict

## 7. Council-Specific Conventions

If you are the council clerk or a council judge:

- Judges receive an IDENTICAL brief — no judge-specific additions
- The orchestrator MUST include in the Council brief: current iteration number (`iteration: N`) and all prior Council verdicts for this finding (`prior_verdicts: [...]`). Judges use this to detect fix regressions and avoid re-ITERATE loops when the root cause has not changed.
- If `iteration > 1` and the fix did not change the finding evidence, El Fiscal SHOULD recommend ESCALATE rather than ITERATE again.
- Zero information sharing between judges — they are dispatched in parallel
- Consensus resolution is DETERMINISTIC — apply rules mechanically, no discretion
- Guardian NEEDS_HUMAN is an absolute override — no combination of other verdicts can override it
- Every council session produces a full decision log, even if unanimous

## 8. Healer-Specific Conventions

If you are the healer (El Cirujano):

- MINIMUM VIABLE FIX only — one cut, one stitch, done
- Check engram for prior heal attempts before touching code — never repeat a failed approach
- Max 2 files per fix — if more needed, ESCALATE
- Max 3 iterations per finding — after that, FORCE STOP and escalate
- Revert on regression: `git revert HEAD --no-edit` and escalate. Never fix the fix.
- Save EVERY attempt to engram (success AND failure) — the next iteration depends on it
- On escalation (max iterations reached): save a final `heal-{finding-id}-summary` entry containing:
  ```json
  {
    "finding_id": "scan-db-001",
    "total_attempts": 3,
    "attempt_history": [
      { "attempt": 1, "approach": "...", "result": "regressed", "files_changed": ["x.go"] },
      { "attempt": 2, "approach": "...", "result": "partial", "files_changed": ["y.go"] },
      { "attempt": 3, "approach": "...", "result": "failed", "files_changed": [] }
    ],
    "escalation_reason": "All approaches exhausted without convergence"
  }
  ```
  The Report phase reads this key — not the individual attempt entries — to generate the escalation summary.
