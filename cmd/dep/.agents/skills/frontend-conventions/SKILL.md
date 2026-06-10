---
name: frontend-conventions
description: >
  How to write frontend code in this project. Covers the CSS utility system,
  component library, dark-mode theming, and the decision rules for avoiding
  style duplication. Read before writing or editing any UI code.
---

# Frontend Conventions

## Goal

Every style decision should have **one source of truth**. When a designer changes the hover color for picker cards, it should update everywhere automatically — not require a grep-and-replace across 10 files.

---

## 1. CSS Variables — the theming layer

All design tokens live in `app/globals.css` as CSS custom properties.

```
:root   → light mode values
.dark   → dark mode overrides
@theme  → maps CSS vars to Tailwind color names (e.g. --color-border → border-border)
```

Because Tailwind utilities reference these vars at runtime, **every Tailwind color class is automatically dark-mode-aware**. You do not need `dark:` variants for standard token-based colors (`border-border`, `bg-muted`, `text-foreground`, etc.).

Only use `dark:` for values that are not design tokens — e.g. a hardcoded `#fff` that needs a different hardcoded value in dark mode.

**Key tokens to know:**

| Token | Use |
|-------|-----|
| `border-border` | Default card/input borders |
| `bg-background` | Page background |
| `bg-muted` | Subtle fills, hover backgrounds |
| `bg-muted-tint` | Very subtle primary tint (selected card background) |
| `text-foreground` | Primary text |
| `text-muted-foreground` | Secondary / label text |
| `border-primary` / `text-primary` | Selected / active state |

---

## 2. Global CSS Utilities — `app/globals.css`

Recurring Tailwind class groups that appear on 3+ elements belong in `@layer utilities` in `globals.css`, **not** inline in components.

Current utilities:

### `selectable-card`
Non-selectable hover cards (block list, picker cards). Applies `border-2 border-border transition-all` + hover state.

```tsx
// ✅ correct
<div className="selectable-card rounded-xl p-4">...</div>

// ❌ wrong — duplicates the pattern inline
<div className="border-2 border-border transition-all hover:border-muted-foreground/30 hover:bg-muted/30 rounded-xl p-4">...</div>
```

**When to add a new utility:** When the same 3+ Tailwind classes appear together in 3+ unrelated components and represent a single visual concept (not just a coincidence). Add it to `@layer utilities` with a comment naming the concept.

```css
@layer utilities {
  .my-new-pattern {
    @apply /* base styles */;
  }
  .my-new-pattern:hover {
    @apply /* hover styles */;
  }
}
```

---

## 3. Component Library — `components/editor/ui/`

Always check this directory before writing a new component or reaching for inline Tailwind. Import from `@/components/editor/ui`.

### Selectable tile components

These handle the full selected/unselected/hover state cycle. Use them instead of writing border logic inline.

| Component | Shape | Use for |
|-----------|-------|---------|
| `OptionTile` | Vertical: preview + label | Style/mode pickers (grid of square tiles) |
| `OptionChoiceCard` | Vertical: preview + title + description + radio dot | Layout choices with descriptions |
| `OptionMediaTile` | Square image frame + label | Wallpaper style, pattern, gradient pickers |
| `OptionListRow` | Horizontal: left icon + label + end adornment | Font picker, list-style selectors |
| `OptionSegmentButton` | Pill/segment | Two-option toggles (small/large, text/logo) |
| `OptionTileSection` | Grid wrapper + title | Wraps any of the above in a `p-5` section with a heading |

**Selected state is handled by the component** — you only pass `selected` and `onSelect`:

```tsx
// ✅ correct
<OptionTile
  selected={value === opt.id}
  onSelect={() => setValue(opt.id)}
  label={opt.label}
  preview={<MyIcon className={value === opt.id ? "text-primary" : "text-muted-foreground"} />}
/>

// ❌ wrong — reinventing the wheel inline
<button className={cn(
  "flex flex-col items-center gap-2 py-3 rounded-xl border-2 transition-all",
  active ? "border-primary bg-muted-tint" : "border-border hover:border-muted-foreground/30 hover:bg-muted/30"
)}>
  ...
</button>
```

### Other components

| Component | Use for |
|-----------|---------|
| `Toggle` | Boolean on/off toggles (e.g. Noise, feature flags) |
| `DashedButton` | Upload / file-select actions with dashed border |
| `MediaUrlUploadCard` | Combined URL input + file upload UI |
| `ColorInput` | Label + hex text field + color swatch picker |
| `FontPicker` | Font selection dropdown |
| `SettingsCardHeader` | Section header with icon + title in settings panels |

---

## 4. Decision tree — inline vs utility vs component

```
Need a style on one element, one place only?
  → Inline Tailwind. Fine.

Same 3+ Tailwind classes appear together in 3+ places?
  → Extract to @layer utilities in globals.css.

Same visual + interactive pattern (e.g. selectable card) with multiple elements?
  → Extract to a component in components/editor/ui/.

Pattern already exists as a component?
  → Use the component. Never rewrite it inline.
```

---

## 5. Anti-patterns to avoid

**Inline duplication of the selectable card pattern**
```tsx
// ❌ — every occurrence must be changed manually when the style evolves
<button className={cn("... border-2 transition-all", active ? "border-primary bg-muted-tint" : "border-border hover:border-muted-foreground/30 ...")}>
```
→ Use `OptionTile` / `OptionChoiceCard` / `selectable-card`.

**Hardcoded colors instead of tokens**
```tsx
// ❌
<div className="bg-[#f2f4f9] dark:bg-[#1a1a1a]">
// ✅
<div className="bg-muted">
```

**Unnecessary `dark:` overrides for token-based colors**
```tsx
// ❌ — border-border already adapts to dark mode via CSS vars
<div className="border-border dark:border-white/10">
// ✅
<div className="border-border">
```

**Skipping OptionTileSection for grouped tiles**
```tsx
// ❌ — duplicates the p-5 + title + grid pattern
<div className="p-5">
  <p className="text-sm font-bold text-foreground mb-3">Style</p>
  <div className="grid grid-cols-3 gap-2">...</div>
</div>
// ✅
<OptionTileSection title="Style" columns={3}>...</OptionTileSection>
```

---

## 6. Adding a new design token

When a new color/spacing/radius concept is needed project-wide:

1. Add the CSS variable to both `:root` and `.dark` in `globals.css`.
2. Map it in `@theme inline` so Tailwind generates a utility class for it.
3. Use the Tailwind class everywhere — never reference `var(--my-token)` directly in JSX.

```css
/* globals.css */
:root  { --my-new-color: #e0f2fe; }
.dark  { --my-new-color: #0c4a6e; }

@theme inline {
  --color-my-new-color: var(--my-new-color);
}
```

```tsx
/* usage */
<div className="bg-my-new-color text-foreground">
```
