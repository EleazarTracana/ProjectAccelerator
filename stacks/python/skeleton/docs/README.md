# Project Documentation

This directory holds all living documentation for the project. The goal is not exhaustive prose but useful, maintainable reference material that stays close to the code.

## Structure

| Directory | Purpose |
|-----------|---------|
| `adl/` | Architecture Decision Log — MADR-format records capturing the "why" behind structural choices |
| Root files | Narrative guides, onboarding docs, operational runbooks |

## When Working On...

| Task | Consult |
|------|---------|
| Adding a new feature | `adl/` for prior architectural constraints, then the skeleton README for layer conventions |
| Changing infrastructure | Relevant ADRs plus `src/shared/infrastructure/` templates |
| Debugging a job | `adl/` for job registry decisions, `src/app/job_bootstrap.py` for registration flow |
| Onboarding a new developer | Start with the skeleton README, then this file, then ADRs |

## Conventions

Documentation follows the same principle as code: small, focused, single-responsibility. Each ADR covers one decision. Guides cover one workflow. If a document tries to explain everything, it explains nothing.
