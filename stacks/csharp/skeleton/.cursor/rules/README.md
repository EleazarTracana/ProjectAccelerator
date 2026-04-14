# Cursor Rules

Rules in this directory come from two sources, and both should be merged into your project:

**Universal rules** live at the ProjectAccelerator root under `.cursor/rules/`. These are stack-agnostic conventions: TDD discipline, observability, documentation style, feature vocabulary, dead code cleanup, and delegation-ready plans. They apply to any project regardless of language.

**C#-specific rules** live at `stacks/csharp/.cursor/rules/` in the ProjectAccelerator. These cover conventions specific to the C#/.NET stack: endpoint thinness, unit testing style, enum storage and contracts, Dapper vs stored procedures, FluentValidation usage, architectural test guidance, comment discipline, and pipeline step logging.

To set up a new project, copy both sets of `.mdc` files into your project's `.cursor/rules/` directory. Review each rule — if a convention does not apply to your project, omit it deliberately rather than copying everything blindly.

After copying, you will likely add project-specific rules as your codebase develops its own conventions. Keep rules short, focused, and one concern per rule. The `guardrail-cursor-rule-writing-style.mdc` universal rule describes the expected format.
