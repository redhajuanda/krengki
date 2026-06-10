---
name: nextjs-clean-architecture
description: >
  Solo-friendly Clean Architecture for this Next.js app. Feature-sliced,
  three layers per feature (domain → data → ui), no DI container, no
  separate controllers/presenters. Defines exact folders, file names,
  and what goes where so a single developer can move fast without
  spaghetti. Read this BEFORE adding any feature, data fetch, mutation,
  or business rule. Pure UI / token / shadcn work uses
  `frontend-conventions` instead.
---

# Next.js Clean Architecture (solo edition)

## 0. Read first

- This is the **lite** version: small enough for one developer, strict
  enough to stop the codebase rotting. We drop DI containers,
  controllers, presenters, and the four-layer textbook split.
- Before using any Next.js API (Server Components, Server Actions, Route
  Handlers, `cookies()`, `headers()`, `revalidatePath`, etc.), open the
  matching guide in `node_modules/next/dist/docs/`. The repo's Next.js
  diverges from training data — verify the API there.
- Defers to: `frontend-conventions` (CSS/theming), `component-reuse`
  (composition), `caveman` (Caveman primitives), `shadcn` (shadcn
  workflows). Pure visual work skips this skill.

---

## 1. Mental model — three layers per feature

```
┌──────────────────────────────────────────────┐
│  ui/         React components, hooks         │  ← outermost
│  ↓ uses use-cases or actions                  │
├──────────────────────────────────────────────┤
│  use-cases   Business rules (pure functions)  │
│  ↓ uses data + domain                         │
├──────────────────────────────────────────────┤
│  data/       Fetch / DB / SDK calls           │
│  ↓ uses domain                                │
├──────────────────────────────────────────────┤
│  domain/     Types, zod schemas, errors       │  ← innermost
└──────────────────────────────────────────────┘
```

**The one rule:** dependencies point **inward**. `domain` knows nothing.
`data` may import `domain`. `use-cases` may import `domain` and `data`.
`ui` may import any of them. Never the reverse.

If a use case is trivial (one fetch, no rules) — skip it and let `ui`
call `data` directly. Add the use case the moment you have:
- a second caller, OR
- a real business rule (validation, permission, branching, multi-step).

This is the “rule of two” — don't pre-build layers you don't need.

---

## 2. Folder structure (exact paths)

```
woah-fe/
├── app/                             # Next.js App Router (UI layer)
│   ├── [username]/
│   ├── studio/
│   ├── api/                         # Route handlers (thin)
│   ├── layout.tsx
│   └── page.tsx
│
├── src/
│   ├── features/                    # All feature code lives here
│   │   └── <feature>/               # e.g. profile, blocks, auth, links
│   │       ├── domain.ts            # types + zod schemas + errors
│   │       ├── data.ts              # API/DB calls (the "repository")
│   │       ├── use-cases.ts         # business logic (optional)
│   │       ├── actions.ts           # 'use server' actions (optional)
│   │       └── ui/                  # feature-only components/hooks
│   │
│   └── shared/                      # cross-feature pieces
│       ├── domain.ts                # shared types/errors (e.g. AppError)
│       ├── http.ts                  # fetch wrapper, axios instance
│       └── auth.ts                  # session helpers
│
├── components/                      # presentation primitives (existing)
│   ├── ui/                          # shadcn-generated
│   ├── studio/ui/                   # project component library
│   └── …
│
└── lib/                             # framework-side helpers (existing)
    ├── utils.ts                     # cn(), formatters
    ├── *-store.ts                   # client stores
    └── use-*.ts                     # generic React hooks
```

### Reconciling with existing layout

- Keep `components/` and `lib/` as-is. They are part of the **ui** layer.
- New domain logic goes under `src/features/<feature>/`.
- Don't bulk-move existing files. Migrate one helper at a time when you
  next touch it (Section 7).

### When does a "feature" deserve its own folder?

Create `src/features/<feature>/` when **any** of these is true:
- It owns its own data shape (a noun the user thinks about: profile,
  block, link, workspace).
- It has more than ~50 lines of non-UI logic.
- It will have at least two call sites (a page + an action, etc.).

Otherwise: a one-off helper next to the page is fine.

---

## 3. Layer specifications

### 3.1 `domain.ts` — types, schemas, errors

Pure TypeScript. **No** imports from `next/*`, `react`, `axios`, fetch
wrappers, or anything else with side effects.

```ts
// src/features/profile/domain.ts
import { z } from 'zod';

export const ProfileSchema = z.object({
  id: z.string(),
  username: z.string().min(2),
  displayName: z.string(),
  isPublished: z.boolean(),
});
export type Profile = z.infer<typeof ProfileSchema>;

export class ProfileNotFoundError extends Error {
  readonly code = 'PROFILE_NOT_FOUND';
  constructor(public id: string) { super(`Profile ${id} not found`); }
}
```

Rule: errors carry a stable string `code`. The UI layer maps codes to
toasts / status; nothing else needs `instanceof` chains.

If two features share types, lift them to `src/shared/domain.ts`.

---

### 3.2 `data.ts` — the only place that does I/O

The only layer allowed to import `axios`, `fetch`, SDKs, Drizzle, etc.
Functions are named after what they do, not how:

```ts
// src/features/profile/data.ts
import { http } from '@/src/shared/http';
import { Profile, ProfileSchema, ProfileNotFoundError } from './domain';

export async function getProfileById(id: string): Promise<Profile> {
  const res = await http.get(`/profiles/${id}`);
  if (res.status === 404) throw new ProfileNotFoundError(id);
  return ProfileSchema.parse(res.data);
}

export async function publishProfile(id: string): Promise<Profile> {
  const res = await http.post(`/profiles/${id}/publish`);
  return ProfileSchema.parse(res.data);
}
```

Rules:
- Always parse external responses through a zod schema. Don't trust the
  network.
- Catch transport errors and rethrow as domain errors from `domain.ts`.
- One file per feature is normally enough. Split into `data/` folder
  only when it crosses ~300 lines.

---

### 3.3 `use-cases.ts` — business rules (optional)

Add this file the moment a function does more than "fetch and return".
Each use case is a **plain async function**. No DI container, no class.
Compose by importing.

```ts
// src/features/profile/use-cases.ts
import { getProfileById, publishProfile } from './data';
import { ProfileNotFoundError } from './domain';

export async function publishMyProfile(input: { profileId: string; userId: string }) {
  const profile = await getProfileById(input.profileId);
  if (profile.id !== input.profileId || profile.ownerId !== input.userId) {
    throw new ProfileNotFoundError(input.profileId);
  }
  return publishProfile(profile.id);
}
```

Rules:
- Validate input with a zod schema from `domain.ts` if it comes from the
  outside. Internal callers can pass typed args.
- Throw domain errors. Never throw `new Error('...')` across a layer
  boundary.
- A use case should not call another use case. If it needs to, the
  shared logic belongs in `domain.ts` or a small helper inside
  `use-cases.ts`.
- If you ever need to swap `data.ts` for tests, refactor the use case to
  accept the data fns as a parameter — but only when that day comes.

---

### 3.4 `actions.ts` — Next.js Server Actions (optional)

Thin wrapper that converts an unknown payload from the client into a
typed call to a use case (or directly to `data` for trivial cases).

```ts
// src/features/profile/actions.ts
'use server';
import { revalidatePath } from 'next/cache';
import { getSession } from '@/src/shared/auth';
import { PublishProfileInput } from './domain';
import { publishMyProfile } from './use-cases';
import { mapError } from '@/src/shared/errors';

export async function publishProfileAction(raw: unknown) {
  try {
    const session = await getSession();
    if (!session) throw new UnauthorizedError();
    const input = PublishProfileInput.parse(raw);
    const profile = await publishMyProfile({ profileId: input.profileId, userId: session.userId });
    revalidatePath('/studio/profile');
    return { ok: true as const, data: profile };
  } catch (e) {
    return mapError(e);
  }
}
```

Rules:
- This is the **only** non-UI place allowed to import `next/*`.
- Always parse the raw payload with zod before calling inward.
- Always return `{ ok: true, data } | { ok: false, error: { code, message } }`.
  No throwing across the action boundary — clients can't catch typed errors.

---

### 3.5 `ui/` — feature components

Anything React/Next-flavoured for this feature. Imports from:
- `./domain` for types
- `./use-cases` or `./actions` for behaviour
- `@/components/*` for primitives
- `@/lib/*` for hooks/utilities

Never reaches into another feature's `data.ts` or `use-cases.ts` directly.
If you need cross-feature data, that's a sign the data belongs in
`src/shared/` or you're slicing features wrong.

---

## 4. Where does X go?

| Question | Answer |
|----------|--------|
| Pure data shape | `<feature>/domain.ts` |
| Validation rule for input | `<feature>/domain.ts` (zod schema) |
| Typed error to throw | `<feature>/domain.ts` |
| Talks to API / DB / SDK | `<feature>/data.ts` |
| “When user does Y, system does Z” | `<feature>/use-cases.ts` |
| `cookies()`, `redirect()`, `revalidatePath` | `app/` page or `<feature>/actions.ts` |
| Reusable React hook (no business rule) | `lib/use-*.ts` |
| Reusable hook that calls a server action | `<feature>/ui/use-*.ts` |
| shadcn / studio UI primitive | `components/editor/ui/` (see `frontend-conventions`) |
| One-off component for a single page | next to that `page.tsx` |
| Cross-feature type/helper | `src/shared/` |

**Tie-breakers**
- Imports `react` or `next/*` → ui layer.
- Imports a third-party I/O lib → `data.ts` only.
- Can be tested without mocks → `domain.ts`.
- Orchestrates multiple data calls or has a real rule → `use-cases.ts`.

---

## 5. Adding a feature — step-by-step

1. **Name the noun.** Create `src/features/<noun>/`.
2. **Model it.** `domain.ts` — types, zod schemas, errors.
3. **Talk to the world.** `data.ts` — one function per API call, parse
   responses through schemas.
4. **Add rules if needed.** Skip `use-cases.ts` until you have a real
   rule or a second caller.
5. **Expose to the client.** Server action in `actions.ts` if it's a
   mutation; server component imports `data.ts` directly for reads.
6. **Build UI.** `ui/` for feature components; pages in `app/` import
   them.
7. **Lint + build.** `pnpm lint && pnpm build`. Fix layering violations
   instead of suppressing them.

---

## 6. Anti-patterns (don't do these)

- **`fetch()` inside a Server Component.** Move it to `data.ts` and
  import it back. One-line change, huge dividends.
- **Importing `axios` from a use case or component.** Infrastructure
  leak. Wrap in `data.ts`.
- **Catching `Error` and rethrowing `Error`.** Use typed errors with
  `code` strings.
- **`'use client'` at the top of a route's `page.tsx`.** Push it to the
  smallest interactive leaf.
- **Cross-feature reach-around.** `features/blocks/ui/` importing
  `features/profile/data.ts` is a smell. Lift the shared piece up to
  `src/shared/` or call a server action.
- **`lib/` junk drawer for domain helpers.** Anything tied to a noun
  (profile, block, link) goes under `src/features/<noun>/`, not `lib/`.
- **Premature use-cases / actions / interfaces.** If `domain.ts` and
  `data.ts` are enough, ship that. Add layers when pain shows up, not
  before.

---

## 7. Migrating existing code

`lib/` currently mixes domain code (`auth-store`, `profile-store`,
`fetch-link-metadata`, `instagram-oauth`) with framework helpers
(`utils.ts`, `use-theme`). Don't bulk-migrate. Move one helper when you
next touch it for an unrelated reason:

1. Identify the noun.
2. Create `src/features/<noun>/domain.ts` and lift types/errors there.
3. Lift API calls to `src/features/<noun>/data.ts`.
4. Replace import sites; delete the old file when nothing references it.
5. No compatibility shims — change call sites in the same commit.

`utils.ts`, generic hooks, and client stores can stay in `lib/`.

---

## 8. Pre-PR checklist

- [ ] No file in `src/features/*/domain.ts` imports `next/*`, `react`,
      or any I/O lib.
- [ ] No file in `src/features/*/use-cases.ts` imports `next/*` or
      third-party I/O libs (only `./data`, `./domain`, `src/shared/*`).
- [ ] Every external response in `data.ts` is parsed through a zod
      schema.
- [ ] Errors thrown across layer boundaries are typed and carry a
      `code`.
- [ ] No new business helper landed in `lib/`.
- [ ] `pnpm lint && pnpm build` clean.
