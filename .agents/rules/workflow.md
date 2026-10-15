# Workflow & Engineering Rules

This document defines the development workflow, branching, commit discipline, and PR guidelines for **pdf2md**.

## 1. Branching & PR Strategy
- `main`: Protected, always stable, source of truth.
- Feature branches: `feat/<index>-<desc>` (e.g. `feat/01-repo-scaffold`, `feat/02-core-pdf-engine`).
- Keep local branches after merging to `main` for developmental study and historical traceability.
- Every branch merged into `main` must go through a documented Pull Request on GitHub.

## 2. Commit Standards (Grouped by Concern, Subissue, Item, Step)
Commits must be atomic and grouped by concern:
```text
<type>(<scope>): <concise imperative title>

- Concern: <module/layer: Architecture / Core / Cleaner / CLI / GUI / Tooling / Docs>
- Subissue/Item: <Item reference>
- Step: <Step X/Y: brief description of step>

<Optional details: rationale, edge cases, or ponytail notes>
```

Allowed types: `feat`, `fix`, `refactor`, `perf`, `test`, `build`, `ci`, `docs`, `chore`.

## 3. Engineering Directives
- **Language**: Go (v1.26+). Idiomatic, minimal dependencies, standard library first.
- **Cross-Platform**: Windows (`pwsh`) and Unix/Linux compatibility across all build targets.
- **Documentation**: All documentation, comments, commit messages, and PR descriptions in professional English.
