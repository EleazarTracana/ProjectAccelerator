# Python/FastAPI Project Skeleton

This skeleton captures the structural conventions of a production Python/FastAPI system. It is not a framework and not a library. It is a starting point — a set of directories, templates, and rules that encode decisions so you do not have to re-derive them on every new project.

## The Key Insight: Vertical Slices, Not Entity Buckets

The most common mistake in FastAPI projects is organizing code around entities: a `services/` folder with `OrderService`, `UserService`, `MetricsService`, each growing into a god class with twenty methods. This skeleton takes a different approach.

A feature is a **verb** — `create_order`, `get_user_detail`, `send_notification`. Each verb gets its own folder with at most four files: a router, a service, schemas, and optionally a job definition. The area folder above it (`orders/`, `users/`) is a grouping prefix that carries no behavior.

The trade-off is more directories. The benefit is that every feature is independently deployable, testable, and deletable. When a feature is retired, you delete one folder. When a feature is added, you create one folder. No god classes, no merge conflicts on shared service files.

## Layer Architecture

Dependencies always point inward, toward the domain kernel.

**Shared Kernel** (`src/shared/kernel/`) contains pure dataclasses (entities) and abstract repository protocols. It has zero external dependencies — no FastAPI, no Motor, no HTTP. This is the vocabulary of the domain, and it must stay clean.

**Infrastructure** (`src/shared/infrastructure/`) implements kernel protocols using real tools: MongoDB via Motor, HTTP clients, external API adapters. This is the only layer that knows about the outside world.

**Feature Services** (`src/features/<area>/<verb>/service.py`) contain use-case logic. They depend on kernel protocols (not implementations) and are instantiated fresh per request via constructor injection. A service handles one use case. If it needs to coordinate multiple features, it belongs in the orchestrators layer instead.

**Routers** (`src/features/<area>/<verb>/router.py`) are thin HTTP bindings. They extract the request, call one service method, and return the result. No business logic, no data assembly, no conditional branching. The moment you see an `if` in a router that is not about HTTP concerns, move it to the service.

**Orchestrators** (`src/orchestrators/`) handle cross-domain workflows that compose two or more feature services. They are the glue layer for batch processing, scheduled jobs, and multi-step business processes.

## Composition Root

All dependency wiring happens in one place: `src/app/container.py`. This is the composition root. No other file instantiates services or repositories. The container is built at startup and attached to `app.state`, making it available to every router via `request.app.state.container`.

The trade-off is explicit wiring (you list every dependency by hand). The benefit is that every dependency relationship is visible in one file, and testing with mocks requires no framework — just pass different implementations to the constructor.

## Job Registry

Scheduled jobs use a factory pattern with DB-driven configuration. Factory functions register via `@JobFactoryRegistry.register("name")` and are discovered through side-effect imports in `job_bootstrap.py`. Job schedules and configs live in a MongoDB collection, allowing deployment of new jobs without code restarts.

## Testing Strategy

Unit and service tests live in `tests/` and run by default with `pytest`. They test services in isolation using mock repositories.

Integration tests live in `integration-test/` and require external resources (a real MongoDB instance, typically via Docker). They are separated because they are slower, require infrastructure, and should not block the fast feedback loop of unit tests. Integration tests for MongoDB include `explain()` assertions on hot queries to verify index usage.

## How to Use This Skeleton

1. Copy this skeleton into your new project root.
2. Remove `.template` extensions from files you want to activate.
3. Merge cursor rules from both the universal set (ProjectAccelerator root) and the Python-specific set.
4. Replace `{project}` placeholders with your project name.
5. Start with one feature verb to establish the pattern, then add more as needed.

The templates are commented-out code showing structure and patterns. They are not runnable as-is — they exist to show you what belongs where and why.
