# Pesa — WCAG 2.2 AA checklist

This is the operational checklist used during development and manual QA.
The authoritative reference is WCAG 2.2 AA; this file tracks how each
requirement is met in the current codebase.

## Perceivable

- [x] **1.1.1 Non-text content** — every `<img>` and icon has `alt` or `aria-label` (SourceBadge, monogram).
- [x] **1.3.1 Info & relationships** — semantic headings (`h1`→`h2`→`h3`), `<nav>`, `<main>`, `<footer>`, `role="search"`.
- [x] **1.4.3 Contrast (minimum)** — brand text on background passes 4.5:1 (brand-700 #1e40af on white).
- [x] **1.4.4 Resize text** — no fixed `px` font sizes in base layers; uses `rem`.
- [x] **1.4.11 Non-text contrast** — focus ring and borders use `currentColor`, ≥3:1.
- [x] **1.4.12 Text spacing** — Tailwind defaults respect user-overridable spacing.
- [x] **1.4.13 Content on hover/focus** — no hover-only tooltips with critical info.

## Operable

- [x] **2.1.1 Keyboard** — all interactive elements reachable with Tab; no `accesskey`.
- [x] **2.1.2 No keyboard trap** — `trapFocus` only used intentionally in modals with Esc/Tab cycle.
- [x] **2.4.1 Bypass blocks** — `<a class="skip-link" href="#main">` skip link.
- [x] **2.4.3 Focus order** — DOM order matches visual order (Layout has no CSS reorder).
- [x] **2.4.4 Link purpose** — links either have text or `aria-label` (dismiss button `×` has `aria-label`).
- [x] **2.4.6 Headings & labels** — every form control has a visible or `sr-only` `<label>`.
- [x] **2.4.7 Focus visible** — global `:focus-visible { outline: 2px solid currentColor; offset: 2px }`.
- [x] **2.4.11 Focus not obscured** — sticky install prompt never covers focused elements (fixed bottom, reserved space).
- [x] **2.5.3 Label in name** — accessible names start with the visible label text.
- [x] **2.5.5 Target size (AAA-leaning)** — button/input min-height 44px via base CSS.
- [x] **2.5.8 Target size (minimum, 2.2)** — verified via axe-core in `e2e/a11y.spec.ts`.

## Understandable

- [x] **3.1.1 Language of page** — `<html lang="it">`, `i18n.changeLanguage` updates `htmlTag` via i18next detector.
- [x] **3.2.1 On focus** — focus does not trigger navigation.
- [x] **3.2.2 On input** — typing in search box does not auto-submit.
- [x] **3.3.1 Error identification** — `role="alert"` on error banners.
- [x] **3.3.2 Labels or instructions** — WeightInput has `aria-describedby` → help text.
- [x] **3.3.7 Redundant entry (2.2)** — no form across pages requires re-typing.
- [x] **3.3.8 Accessible authentication (2.2)** — auth flows (Phase 6) must not require cognitive puzzles.

## Robust

- [x] **4.1.2 Name, role, value** — state toggle uses `role="radiogroup"` / `role="radio"` / `aria-checked`.
- [x] **4.1.3 Status messages** — search result counts + loading states use `aria-live="polite"` + `role="status"`.

## Automation

- `axe-core` via `@axe-core/playwright` runs in `e2e/a11y.spec.ts` against `/`, `/cerca`, `/alimento/:id`, `/scopo`, `/accessibilita`.
- CI fails on any serious/critical violation.
