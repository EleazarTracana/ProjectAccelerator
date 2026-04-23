# Sentinel — Reference Material

> Templates, schemas, and detailed specifications. Loaded on demand by the sentinel agent.
> Do NOT read this file on activation — only when a specific section is needed.

---

## 1. Product Model Schema

```json
{
  "project_name": "string",
  "purpose": "one-sentence product purpose",
  "change_name": "SDD change being implemented",
  "tasks": [
    {
      "id": "task-1",
      "description": "what to implement",
      "spec_requirements": ["req-1", "req-2"],
      "affected_endpoints": ["/api/..."],
      "affected_routes": ["/dashboard"],
      "dependencies": ["task-0"],
      "status": "pending | completed | escalated"
    }
  ],
  "architecture_decisions": [
    {
      "id": "ADR-001",
      "decision": "what was decided",
      "constraint": "what this means for code",
      "violation_pattern": "what a violation looks like"
    }
  ],
  "services": "read from sentinel.config.yaml → services",
  "anti_patterns": "see §5 below + custom from config"
}
```

---

## 2. HTML Report Template

Self-contained single HTML file. Dark theme. No external dependencies.

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Sentinel Report — {change-name} — {date}</title>
  <style>
    :root {
      --bg: #0f172a;
      --surface: #1e293b;
      --border: #334155;
      --text: #f1f5f9;
      --text-muted: #94a3b8;
      --accent: #3b82f6;
      --approved: #22c55e;
      --escalated: #f97316;
      --broken: #ef4444;
      --suggestion: #3b82f6;
    }
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
      background: var(--bg); color: var(--text);
      line-height: 1.6; padding: 2rem; max-width: 1200px; margin: 0 auto;
    }
    h1 { font-size: 1.8rem; margin-bottom: 0.5rem; }
    h2 { font-size: 1.3rem; margin: 2rem 0 1rem; border-bottom: 1px solid var(--border); padding-bottom: 0.5rem; }
    h3 { font-size: 1.1rem; margin: 1rem 0 0.5rem; }
    .meta { color: var(--text-muted); font-size: 0.9rem; margin-bottom: 2rem; }

    /* Dashboard */
    .dashboard { display: grid; grid-template-columns: repeat(auto-fit, minmax(160px, 1fr)); gap: 1rem; margin-bottom: 2rem; }
    .stat { background: var(--surface); border-radius: 8px; padding: 1.2rem; text-align: center; border-left: 4px solid var(--accent); }
    .stat .count { font-size: 2rem; font-weight: 700; display: block; }
    .stat .label { color: var(--text-muted); font-size: 0.85rem; text-transform: uppercase; letter-spacing: 0.05em; }
    .stat[data-type="approved"] { border-color: var(--approved); }
    .stat[data-type="escalated"] { border-color: var(--escalated); }
    .stat[data-type="broken"] { border-color: var(--broken); }

    /* Timeline */
    .timeline { position: relative; padding-left: 2rem; }
    .timeline::before { content: ''; position: absolute; left: 8px; top: 0; bottom: 0; width: 2px; background: var(--border); }
    .timeline-item { position: relative; margin-bottom: 1.5rem; }
    .timeline-item::before { content: ''; position: absolute; left: -2rem; top: 6px; width: 12px; height: 12px; border-radius: 50%; background: var(--accent); border: 2px solid var(--bg); }
    .timeline-item[data-status="approved"]::before { background: var(--approved); }
    .timeline-item[data-status="escalated"]::before { background: var(--escalated); }
    .timeline-item time { color: var(--text-muted); font-size: 0.8rem; font-family: monospace; }

    /* Task cards */
    details { background: var(--surface); border-radius: 8px; margin-bottom: 0.75rem; border: 1px solid var(--border); }
    details[open] { border-color: var(--accent); }
    summary { padding: 1rem 1.2rem; cursor: pointer; list-style: none; display: flex; align-items: center; gap: 0.75rem; }
    summary::-webkit-details-marker { display: none; }
    summary::before { content: '▸'; transition: transform 0.2s; flex-shrink: 0; }
    details[open] > summary::before { transform: rotate(90deg); }
    .finding-body { padding: 0 1.2rem 1.2rem; }

    /* Badges */
    .badge { display: inline-block; padding: 0.15rem 0.6rem; border-radius: 4px; font-size: 0.75rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em; }
    .badge--approved { background: rgba(34,197,94,0.15); color: var(--approved); }
    .badge--escalated { background: rgba(249,115,22,0.15); color: var(--escalated); }
    .badge--broken { background: rgba(239,68,68,0.15); color: var(--broken); }
    .badge--suggestion { background: rgba(59,130,246,0.15); color: var(--suggestion); }

    /* Council */
    .council { display: grid; grid-template-columns: repeat(3, 1fr); gap: 0.75rem; margin: 1rem 0; }
    .council-agent { background: var(--bg); border-radius: 6px; padding: 0.8rem; border: 1px solid var(--border); }
    .council-agent .name { font-weight: 600; font-size: 0.85rem; margin-bottom: 0.3rem; }
    .council-agent .verdict { font-weight: 700; }

    /* Screenshots */
    .screenshot { max-width: 100%; border-radius: 6px; border: 1px solid var(--border); margin: 0.5rem 0; }
    .before-after { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
    .before-after img { width: 100%; }
    .before-after .label { text-align: center; color: var(--text-muted); font-size: 0.8rem; margin-bottom: 0.3rem; }

    /* Code */
    pre { background: var(--bg); padding: 1rem; border-radius: 6px; overflow-x: auto; font-size: 0.85rem; border: 1px solid var(--border); }
    code { font-family: 'SF Mono', 'Fira Code', monospace; }

    @media print { body { background: white; color: black; } details { display: block; border: 1px solid #ccc; } details > * { display: block; } }
    @media (max-width: 640px) { body { padding: 1rem; } .council { grid-template-columns: 1fr; } .before-after { grid-template-columns: 1fr; } }
  </style>
</head>
<body>

  <header>
    <h1>Sentinel Report</h1>
    <p class="meta">
      Change: <strong>{change-name}</strong> ·
      Run: {run-id} · {date} · Duration: {duration} · Cost: ${cost}
    </p>
  </header>

  <!-- DASHBOARD -->
  <section class="dashboard">
    <div class="stat" data-type="tasks"><span class="count">{total}</span><span class="label">Tasks</span></div>
    <div class="stat" data-type="approved"><span class="count">{approved}</span><span class="label">Approved</span></div>
    <div class="stat" data-type="escalated"><span class="count">{escalated}</span><span class="label">Escalated</span></div>
    <div class="stat" data-type="broken"><span class="count">{iterations}</span><span class="label">Iterations</span></div>
  </section>

  <!-- TIMELINE -->
  <h2>Timeline</h2>
  <section class="timeline">
    <!-- {for each event} -->
    <div class="timeline-item" data-status="{status}">
      <time>{timestamp}</time>
      <strong>{event}</strong> — {description}
    </div>
    <!-- {/for} -->
  </section>

  <!-- TASK RESULTS -->
  <h2>Task Results</h2>
  <!-- {for each task} -->
  <details data-status="{status}" open>
    <summary>
      <span class="badge badge--{status}">{STATUS}</span>
      <strong>Task {id}</strong>: {description}
    </summary>
    <div class="finding-body">
      <h3>Implementation</h3>
      <p>{summary}</p>
      <pre><code>{files changed}</code></pre>

      <h3>Validation</h3>
      <p>{scanner results}</p>

      <h3>Council Verdict</h3>
      <div class="council">
        <div class="council-agent">
          <div class="name">El Fiscal (Opus)</div>
          <div class="verdict">{VERDICT}</div>
          <p>{reasoning}</p>
        </div>
        <div class="council-agent">
          <div class="name">El Perito (Kimi K2.5)</div>
          <div class="verdict">{VERDICT}</div>
          <p>{reasoning}</p>
        </div>
        <div class="council-agent">
          <div class="name">El Guardián (Kimi K2.5)</div>
          <div class="verdict">{VERDICT}</div>
          <p>{reasoning}</p>
        </div>
      </div>

      <h3>Evidence</h3>
      <div class="before-after">
        <div><div class="label">Before</div><img class="screenshot" src="data:image/jpeg;base64,{b64}" alt="Before"></div>
        <div><div class="label">After</div><img class="screenshot" src="data:image/jpeg;base64,{b64}" alt="After"></div>
      </div>
    </div>
  </details>
  <!-- {/for} -->

  <!-- ESCALATED -->
  <h2>Escalated (Needs Human Review)</h2>
  <!-- {for each escalated} -->
  <details data-status="escalated">
    <summary>
      <span class="badge badge--escalated">ESCALATED</span>
      <strong>Task {id}</strong>: {description}
    </summary>
    <div class="finding-body">
      <h3>Why Escalated</h3>
      <p>{reason}</p>
      <h3>Council Opinions</h3>
      <div class="council"><!-- same grid --></div>
      <h3>Suggested Approach</h3>
      <p>{recommendation}</p>
    </div>
  </details>
  <!-- {/for} -->

  <!-- SUGGESTIONS -->
  <h2>Suggestions</h2>
  <!-- {for each suggestion} -->
  <details>
    <summary><span class="badge badge--suggestion">SUGGESTION</span> {description}</summary>
    <div class="finding-body">
      <p>{anti-pattern violated}</p>
      <img class="screenshot" src="data:image/jpeg;base64,{b64}" alt="Screenshot">
    </div>
  </details>
  <!-- {/for} -->

  <!-- RAW LOG -->
  <details>
    <summary>Raw Execution Log (JSON)</summary>
    <pre id="raw-log"></pre>
  </details>

  <script type="application/json" id="report-data">{full_run_data_json}</script>
  <script>
    const d = JSON.parse(document.getElementById('report-data').textContent);
    document.getElementById('raw-log').textContent = JSON.stringify(d, null, 2);
  </script>
</body>
</html>
```

---

## 3. Deploy Failure Classification

When `docker compose --profile app up --build --wait` fails:

| Category | Detection | Agent action |
|----------|-----------|-------------|
| **Build failure** | stderr: `failed to solve`, `failed to build` | Fix Dockerfile or code, retry |
| **Crash loop** | Container exits immediately | `docker compose logs {service} --tail=50` |
| **Health timeout** | Container running but `unhealthy` | `docker inspect --format='{{json .State.Health.Log}}' {container}` |
| **Dependency failure** | Service B stuck because A unhealthy | Fix upstream first |
| **Port conflict** | stderr: `address already in use` | `lsof -i :{port}`, resolve |

Max 3 retries. After 3 → STOP and generate failure report.

---

## 4. Scanner Output Schemas

### DB Scanner (El Minero)
```json
[{
  "finding": "string",
  "severity": "BROKEN|INCONSISTENT|DEGRADED|SUGGESTION",
  "evidence": {
    "query": "string",
    "expected": "string",
    "actual": "string",
    "reproduction": "docker compose exec command"
  }
}]
```

### API Scanner (El Rastreador)
```json
[{
  "endpoint": "GET /api/...",
  "finding": "string",
  "expected": "status/schema",
  "actual": "status/schema",
  "severity": "string",
  "evidence": {
    "curl_command": "string",
    "response_body": "truncated to 500 chars",
    "response_time_ms": 123
  }
}]
```

### UI Scanner (El Observador)
```json
[{
  "route": "/dashboard",
  "finding": "string",
  "screenshot_path": "string",
  "severity": "string",
  "evidence": {
    "console_errors": [],
    "accessibility_issues": [],
    "anti_pattern": "name from checklist"
  }
}]
```

### Architecture Scanner (El Centinela)
```json
[{
  "decision_id": "ADR-001",
  "finding": "string",
  "file": "src/...",
  "line": 42,
  "severity": "string",
  "evidence": {
    "import_chain": "A -> B -> C",
    "violation_type": "layer_crossing|naming|hardcoded"
  }
}]
```

---

## 5. Anti-Pattern Checklist (Default)

### UX Anti-Patterns

| # | Name | Description |
|---|------|------------|
| 1 | **Dead End** | Page/modal with no navigation out |
| 2 | **Silent Empty State** | List shows nothing with no message |
| 3 | **Missing Loading State** | Data-dependent UI blank during fetch |
| 4 | **Orphaned Action** | Button that does nothing |
| 5 | **Raw Error Display** | User sees stack trace or "undefined" |
| 6 | **Infinite Spinner** | Loading indicator never resolves |
| 7 | **Missing Feedback** | Action completes with no confirmation |
| 8 | **Broken Navigation** | Nav link that 404s |
| 9 | **State Desync** | UI shows stale data after action |
| 10 | **Inaccessible Action** | Element not keyboard-reachable |

### API Anti-Patterns

| # | Name | Description |
|---|------|------------|
| 11 | **Stack Trace Leak** | Error response has internal paths |
| 12 | **Inconsistent Status Codes** | Same error, different codes |
| 13 | **Missing CORS** | Frontend gets CORS error |
| 14 | **Schema Mismatch** | Response doesn't match contract |
| 15 | **Silent Failure** | Returns 200 but op failed |
| 16 | **N+1 Response** | List needs N extra fetches |

### Architecture Anti-Patterns

| # | Name | Description |
|---|------|------------|
| 17 | **Layer Violation** | Import crosses boundary |
| 18 | **Hardcoded Config** | Secrets in source |
| 19 | **God Service** | Module >500 lines or >10 deps |
| 20 | **Missing Interface** | No port/adapter abstraction |
| 21 | **Test Modifying Source** | Test changes prod state |
| 22 | **Circular Dependency** | A imports B imports A |

### Data Anti-Patterns

| # | Name | Description |
|---|------|------------|
| 23 | **Orphaned Reference** | FK to non-existent record |
| 24 | **Missing Index** | Query field has no index |
| 25 | **Schema Drift** | DB doesn't match entities |
| 26 | **Stale Seed Data** | Fixtures reference deleted entities |

---

## 6. Report Data Schema

The `report-data` JSON embedded in HTML:

```json
{
  "run_id": "sentinel-run-20260422-023000",
  "change_name": "add-feature-x",
  "start_time": "2026-04-22T02:30:00Z",
  "end_time": "2026-04-22T05:12:00Z",
  "duration_seconds": 9720,
  "cost_usd": 12.50,
  "tasks": [
    {
      "id": "task-1",
      "description": "...",
      "status": "approved | escalated",
      "iterations": 1,
      "implementation": {
        "files_changed": ["src/..."],
        "commit_sha": "abc123"
      },
      "validation": {
        "findings": [
          { "scanner": "api", "finding": "...", "severity": "BROKEN", "evidence": "..." }
        ]
      },
      "council": {
        "fiscal": { "verdict": "ALLOW", "reasoning": "..." },
        "perito": { "verdict": "FIX_VALID", "reasoning": "...", "evidence_cited": ["file:line"] },
        "guardian": { "verdict": "ALIGNED", "reasoning": "..." },
        "consensus": "APPROVED"
      },
      "screenshots": {
        "before": "base64...",
        "after": "base64..."
      }
    }
  ],
  "escalated": [],
  "suggestions": []
}
```

---

## 7. Screenshot Embedding

Convert to base64 JPEG for HTML:

```bash
# ImageMagick
convert screenshot.png -quality 75 jpeg:- | base64 -w0

# Or with sips (macOS native)
sips -s format jpeg -s formatOptions 75 screenshot.png --out /tmp/ss.jpg
base64 -i /tmp/ss.jpg
```

Size: ~80KB JPEG → ~107KB base64 per screenshot. 30 screenshots ≈ 3.2MB total.
