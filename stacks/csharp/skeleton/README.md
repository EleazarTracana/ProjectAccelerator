# C# / .NET Skeleton

This skeleton is not a runnable project. There is no `.csproj`, no NuGet restore, no `dotnet build` that produces a binary. What it provides is more valuable than that: a structural blueprint that encodes how features, infrastructure, and tests relate to each other in a production C#/.NET backend.

The directory layout embodies the architecture. When you copy this skeleton into a new project, you are adopting a set of conventions about where code lives, how dependencies flow, and what each layer is responsible for. The `.template` files inside each directory show the pattern through commented examples — read them, understand them, then replace them with your first real implementation.

## How to Use This

Replace every `{Project}` placeholder with your actual project name. For example, if your project is called `Inventory`, rename `{Project}.Api` to `Inventory.Api` and `{Project}.Api.Tests` to `Inventory.Api.Tests`.

Read each `.template` file before deleting it. The templates are not boilerplate to copy verbatim — they are illustrated conventions. Once you have written your first real endpoint, service, test, and repository in the corresponding folders, the templates have served their purpose and can be removed.

## What Each Layer Does

**`src/{Project}.Api/Features/`** is where all user-facing behavior lives. Features are organized as vertical slices grouped by domain area. Each area (`_ExampleArea` is a placeholder) contains concrete slices — an endpoint, its service, its request/response records — and optionally a `Shared/` folder for validators, mappers, or helpers that serve multiple slices within the same area. Features do not reference each other.

**`src/{Project}.Api/Common/`** holds cross-cutting infrastructure that multiple features depend on: the global exception handler, base repository abstractions, middleware, and shared configuration. Code moves here only when it is genuinely needed by more than one feature area. If something is only used by one area, it stays in that area's `Shared/` folder.

**`src/{Project}.Api/Program.cs`** is the single decision point for application composition. All dependency registration, middleware ordering, and configuration binding happens here. No other file instantiates services or wires dependencies. This is the composition root.

**`src/{Project}.Api.Tests/Unit/`** contains unit tests that verify service logic in isolation, using mocks for external dependencies. Tests follow the conventions in the unit testing cursor rule: mandatory DisplayName, no section comments, quality over quantity.

**`src/{Project}.Api.Tests/Architecture/`** contains architectural tests that enforce structural invariants: dependency direction, feature isolation, forbidden references. These tests run with the unit test suite but protect project-wide rules rather than individual behaviors.

**`src/libs/`** is where shared libraries live — but only when they are genuinely reusable across multiple projects. A helper class used by one project does not belong here. See the libs README for guidance.

**`docs/`** holds project documentation. The `adl/` subfolder contains Architecture Decision Records in MADR format. The docs README describes when and how to add documentation.

**`.cursor/rules/`** is where your project's cursor rules live. You merge rules from two sources: universal rules from the ProjectAccelerator root and C#-specific rules from `stacks/csharp/.cursor/rules/`. See the rules README for details.
