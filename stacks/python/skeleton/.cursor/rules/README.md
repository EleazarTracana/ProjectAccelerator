# Cursor Rules

Rules in this directory come from two sources:

1. **Universal rules** from the ProjectAccelerator root `.cursor/rules/` — these apply to all stacks (TDD, ADR templates, observability, documentation standards).
2. **Python-specific rules** from `stacks/python/.cursor/rules/` — these encode conventions specific to the Python/FastAPI stack (vertical slices, MongoDB guardrails, job registry, clean architecture).

When bootstrapping a new project, merge both sets into this directory. Universal rules provide the cross-cutting standards; stack-specific rules provide the implementation patterns.

If a universal rule and a stack rule conflict, the stack rule takes precedence for implementation details (e.g., how to structure a test), while the universal rule takes precedence for process (e.g., when to write an ADR).
