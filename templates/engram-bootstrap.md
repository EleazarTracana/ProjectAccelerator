# Engram Bootstrap Protocol

When starting a new project from the accelerator, run these observations in the first session to seed the assistant's memory. Replace placeholders with actual values.

## Required Bootstrap Observations

### 1. Project Identity

```
topic_key: architecture/overview
type: architecture
title: {Project} — stack and architecture overview
content: |
  **What**: {Project} is a {stack} project following {architecture style}.
  **Why**: {One sentence on why this stack and architecture were chosen}.
  **Where**: Entry point at {entry point path}. Features at {features path}.
  **Learned**: {Any initial constraints — deployment target, team size, compliance requirements}.
```

### 2. Testing Strategy

```
topic_key: pattern/testing
type: pattern
title: Testing strategy and conventions
content: |
  **What**: {Test framework} for unit tests, {integration approach} for integration.
  **Why**: {Why this testing approach — speed, reliability, team preference}.
  **Where**: Unit tests at {test path}. Integration tests at {integration path}.
  **Learned**: {Any testing constraints — CI timeout, required coverage, external dependencies}.
```

### 3. Database and Persistence

```
topic_key: architecture/persistence
type: architecture
title: Database choice and data access patterns
content: |
  **What**: {Database} accessed via {ORM/driver/pattern}.
  **Why**: {Why this database — data model fit, team expertise, scaling needs}.
  **Where**: {Repository/infrastructure path}.
  **Learned**: {Gotchas — connection pooling, migration strategy, enum storage}.
```

### 4. Deployment Model

```
topic_key: config/deployment
type: config
title: Deployment target and constraints
content: |
  **What**: Deploys to {target} via {mechanism}.
  **Why**: {Why this deployment model}.
  **Where**: {CI/CD config path if applicable}.
  **Learned**: {Constraints — cold start limits, memory caps, environment variable patterns}.
```

## Optional Bootstrap Observations

Add these when applicable:

- `architecture/auth` — Authentication and authorization model
- `config/environment` — Environment variable conventions and secrets management
- `pattern/error-handling` — Error handling strategy (Result types, exceptions, error codes)
- `decision/api-style` — API design conventions (REST, GraphQL, RPC, versioning)
- `pattern/logging` — Observability conventions (structured logging, correlation IDs, log levels)

## Verification

After bootstrap, run `mem_search` with the project name to confirm all observations are saved. A well-bootstrapped project should have 4-6 observations covering the fundamentals above.
