---
name: component-reuse
description: >
  Frontend component reuse and code quality principles. Apply when writing, editing,
  or reviewing any frontend component, hook, or utility. Prevents duplication,
  promotes composability, and enforces senior-level frontend thinking.
---

# Component Reuse — Think Like a Senior Frontend Dev

## Core principle

Before writing a single line of UI or logic, ask: **does this already exist?**

1. Search the codebase (`components/`, `lib/`, `hooks/`) for existing components, hooks, and utilities that solve the same problem.
2. If something is 80%+ the same, extend or compose it — don't duplicate it.
3. Only create a new primitive when nothing close exists.

## Composition over duplication

- Prefer **props + slots** to forking a component.
- Extract shared logic into a custom hook (`use-*.ts`) rather than copy-pasting stateful code.
- Shared pure helpers (formatters, validators, transformers) live in `lib/` — never inline them in more than one place.
- A component that does two unrelated things should be split; a component that's copy-pasted three times should become one parameterised component.

## When duplication is acceptable

- The two things share a name but serve genuinely different domains (e.g. a `Button` in a design system vs. a one-off `SubmitButton` with business rules baked in).
- The abstraction would require more props/conditionals than just writing a small, focused component.
- You are writing a test double or a story — not production code.

## Checklist before creating anything new

- [ ] Searched `components/` for an existing match or near-match.
- [ ] Checked `lib/` for utilities I can import instead of re-implement.
- [ ] Checked `hooks/` (or `lib/use-*.ts`) for reusable state logic.
- [ ] Confirmed the new component has a single, clear responsibility.
- [ ] No prop is added "just in case" — YAGNI applies to component APIs too.

## Code quality non-negotiables

- Name things for what they **are**, not what they **do right now** (`UserAvatar`, not `CircleImageThing`).
- Co-locate: types, styles, and helpers that only one component uses live next to that component, not in a global barrel.
- Avoid index-barrel re-exports that hide where things come from.
- Keep component files under ~200 lines; extract if they grow beyond that.
- Never pass raw DOM event handlers through multiple layers — lift state or use context instead.
