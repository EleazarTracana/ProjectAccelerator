# Sentinel QA Agent — Setup Guide

Sentinel is an autonomous QA agent that deploys your project via Docker, validates across DB/API/UI/Architecture layers, convenes a multi-model adversarial council, and produces an HTML report.

## Installation

### 1. Copy agent and skills to your project

Copy the following structure into your project's `.claude/` directory:

```
.claude/
├── agents/
│   └── sentinel.md              # Orchestrator agent
├── docs/
│   └── sentinel-reference.md    # Reference material (schemas, templates)
└── skills/
    ├── _shared/
    │   └── sentinel-protocol.md # Cross-cutting protocol
    ├── sentinel-init/SKILL.md
    ├── sentinel-scan-db/SKILL.md
    ├── sentinel-scan-api/SKILL.md
    ├── sentinel-scan-ui/SKILL.md
    ├── sentinel-scan-arch/SKILL.md
    ├── sentinel-council/SKILL.md
    ├── sentinel-heal/SKILL.md
    └── sentinel-report/SKILL.md
```

### 2. Create your project config

Copy `sentinel.config.template.yaml` to your project root as `sentinel.config.yaml` and customize:

```bash
cp /path/to/ProjectAccelerator/sentinel.config.template.yaml ./sentinel.config.yaml
```

**What to customize:**

- **`services`**: List every service in your docker-compose with ports, health paths, and base URLs
- **`deploy.command`**: Your docker compose up command
- **`implementation.blocked_paths`**: Paths Sentinel should never modify (DB scripts, CI, etc.)
- **`models`**: Adjust model assignments if needed (e.g., add a diversity model for council)
- **`council`**: Configure which providers each judge uses

### 3. Prerequisites

- Docker and Docker Compose
- Playwright MCP server (for UI scanning)
- Engram MCP server (for persistent memory, optional but recommended)

### 4. Run Sentinel

From your project root with Claude Code:

```
sentinel run
```

Or invoke directly: `/sentinel`

## Config Reference

| Section | Purpose |
|---------|---------|
| `deploy` | How to start/stop the project and health check timeout |
| `services` | Service topology — ports, health endpoints, DB names |
| `validation` | Guardrails — max cycles, duration, cost |
| `models` | Model stack — which Claude model for each tier |
| `personas` | Agent personas — instructions for each scanner/judge/healer |
| `council` | Consensus rules for adversarial review |
| `implementation` | Scope limits and blocked paths for healing |
| `evidence` | Screenshot and trace settings |
| `report` | Output format and PR creation |
