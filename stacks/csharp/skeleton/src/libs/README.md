# Shared Libraries

This directory is for code that is genuinely reusable across multiple independent projects. The key word is "independent" — if the code is only used by one project, it does not belong here regardless of how generic it feels.

The temptation to extract a shared library early is strong, especially when a utility class or a base abstraction looks like it could serve other projects someday. Resist that temptation. Premature extraction creates coupling between projects that did not need to be coupled, and the shared library becomes a coordination bottleneck where changes require synchronized deployments.

A good candidate for extraction into `libs/` has these characteristics: it is used by at least two projects today (not hypothetically), it has a stable interface that changes infrequently, and it does not depend on application-specific configuration or domain concepts.

Examples of good shared libraries: a structured logging configuration package, a Dapper extension for multi-tenant connection resolution, a common health check implementation. Examples of things that should stay in the project: a helper class used by three services within the same API, a mapper that converts between two DTOs specific to one feature area.

When you do extract a library here, give it its own project file and treat it as an internal NuGet package — even if you never publish it to a feed. The discipline of a package boundary prevents accidental coupling back to the consuming project.
