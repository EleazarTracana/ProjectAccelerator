# Documentation

This directory holds guides, architecture decision records, and reference material for the project. The goal is not to document everything — it is to document the things that are hard to learn from the code alone.

## When Working On...

| Topic | Read First |
|-------|------------|
| New feature | `docs/adl/` for related ADRs |
| Pipeline / batch job | `docs/pipeline/` + relevant ADRs |
| Infrastructure / deploy | `docs/operations/` |
| Architecture decision | `docs/adl/` + this index |

For **non-trivial** changes (new behavior, pipeline work, cross-cutting refactors), use this table to load the relevant guides. Typically 1-2 guides plus any linked ADRs. Do not read the full `docs/` tree by default.

## Structure

```
docs/
├── adl/              # Architecture Decision Records (MADR format)
├── operations/       # Runbooks, deploy guides, incident logs
├── pipeline/         # Data flow, job constraints, throughput notes
└── README.md         # This file
```

## ADRs

Architecture decisions live in `docs/adl/`. Each follows the MADR format. Statuses: PROPOSED, ACCEPTED, RETIRED. Do not act on RETIRED ADRs. Confirm before acting on PROPOSED ones.

If the change records a lasting architecture decision, add or update an ADR in `docs/adl/`, not a duplicate design dump at the docs root.
