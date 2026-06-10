<!--BEGIN:Behavioral Guidelines -->
# Behavioral Guidelines

Behavioral guidelines to reduce common LLM coding mistakes.

**Tradeoff:** These guidelines bias toward caution over speed. For trivial tasks, use judgment.

## 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

## 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

## 3. Surgical Changes
/
**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

## 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.

---

**These guidelines are working if:** fewer unnecessary changes in diffs, fewer rewrites due to overcomplication, and clarifying questions come before implementation rather than after mistakes.
<!--END:Behavioral Guidelines -->

<!-- BEGIN:nextjs-agent-rules -->
# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` before writing any code. Heed deprecation notices.
<!-- END:nextjs-agent-rules -->

<!-- BEGIN:frontend-skills -->
## Frontend Skills

Skills are located at `.agents/skills/<skill-name>/SKILL.md`.

**Always read `component-reuse` and `frontend-conventions` before writing or editing any UI code.**

| Skill | When to read |
|-------|-------------|
| `component-reuse` | Before creating or editing any component, hook, or shared utility — enforces reuse-first thinking and prevents duplication |
| `frontend-conventions` | Before writing any UI code — covers the CSS utility system, component library (`components/editor/ui/`), dark-mode theming, and rules for avoiding style duplication |
| `nextjs-clean-architecture` | Before adding any feature, data fetch, mutation, or business rule — defines the entities/application/interface-adapters/infrastructure layers, exact folder paths under `src/`, and how they connect to `app/` |
<!-- END:frontend-skills -->

<!-- BEGIN:code-review-graph MCP tools -->
# Code Review Graph

This project has a knowledge graph (auto-updated on file changes). **Prefer the
`code-review-graph` MCP tools over Grep/Glob/Read for exploring code, tracing
relationships, and reviewing changes** — they're cheaper and give structural
context (callers, dependents, test coverage). Fall back to file scanning only
when the graph doesn't cover what you need.

| Tool | Use when |
|------|----------|
| `semantic_search_nodes` | Finding functions/classes by name or keyword |
| `query_graph` | Tracing callers/callees/imports/tests/dependencies |
| `detect_changes` + `get_review_context` | Reviewing changes (risk-scored, token-efficient) |
| `get_impact_radius` / `get_affected_flows` | Understanding blast radius of a change |
| `get_architecture_overview` | High-level codebase structure |
| `refactor_tool` | Planning renames, finding dead code |
<!-- END:code-review-graph MCP tools -->

<!-- BEGIN:rtk-instructions v2 -->
# RTK (Rust Token Killer) - Token-Optimized Commands

**Always prefix shell commands with `rtk`** — it applies a compact filter when one
exists and otherwise passes through unchanged, so it's always safe. Apply it to
every command in a chain too (e.g. `rtk git add . && rtk git commit -m "msg"`).

RTK has dedicated filters for: build/compile (`cargo`, `tsc`, `lint`, `prettier`,
`next build`), tests (`cargo test`, `vitest`, `playwright`), `git` (all
subcommands), `gh` (pr/run/issue/api), package managers (`pnpm`, `npm`, `npx`,
`prisma`), files/search (`ls`, `read`, `grep`, `find`), infra (`docker`,
`kubectl`), and network (`curl`, `wget`). Run `rtk gain` for savings stats.
<!-- END:rtk-instructions -->