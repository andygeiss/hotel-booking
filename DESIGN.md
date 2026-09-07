---
version: alpha
name: Hotel Booking
description: A hotel reservation and payment system, styled as frosted glass over a dark gradient.
colors:
  bg: "#000000"
  bg-gradient: "linear-gradient(160deg, #1c1c1e 0%, #2c2c2e 50%, #1c1c1e 100%)"
  surface: "rgba(255, 255, 255, 0.08)"
  surface-solid: "rgba(255, 255, 255, 0.12)"
  text: "#f5f5f7"
  text-muted: "#8e8e93"
  accent: "#bf5af2"
  primary: "#0a84ff"
  primary-hover: "#409cff"
  primary-text: "#ffffff"
  error: "#ff453a"
  success: "#30d158"
  warning: "#ff9f0a"
  border: "rgba(255, 255, 255, 0.18)"
  border-light: "rgba(255, 255, 255, 0.3)"
  bg-dark: "#0f0f1a"
  bg-gradient-dark: "linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%)"
  surface-dark: "rgba(255, 255, 255, 0.08)"
  surface-solid-dark: "rgba(255, 255, 255, 0.12)"
  text-dark: "rgba(255, 255, 255, 0.95)"
  text-muted-dark: "rgba(255, 255, 255, 0.6)"
  border-dark: "rgba(255, 255, 255, 0.2)"
  accent-dark: "#bf5af2"
  error-dark: "#ff453a"
typography:
  body:
    fontFamily: '"Inter", "SF Pro Display", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif'
  mono:
    fontFamily: '"JetBrains Mono", "Fira Code", "SF Mono", monospace'
rounded:
  radius: "12px"
components:
  button:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.primary-text}"
    rounded: "{rounded.radius}"
---

# Design: Hotel Booking

## Overview

This file is the project's design system: the theme values, contrast floors, and component
inventory every page is styled from. It is written for whoever styles a page — usually an
AI coding agent — and design tools parse the same file.

The theme is **glassmorphism**: translucent white panels blurred over a dark gradient, with
one saturated blue for anything interactive. It is Apple-derived, and it is dark in both
colour schemes — see *Colors*.

Styles live in three layers, loaded in this order, rather than in one `app.css`:

| Layer | File | Holds |
|---|---|---|
| 1 | `cmd/server/assets/static/css/base.css` | Reset only. No colours, no fonts, no theme values. |
| 2 | `cmd/server/assets/static/css/theme.css` | Every token below, as custom properties on `:root`. |
| 3 | `cmd/server/assets/static/css/styles.css` | Components and layout. Reads tokens — with nine exceptions, listed under *Colors*. |

Every CSS value in this file is identical to the same value in `theme.css` — the gradients
are the only ones rewritten, onto one line, because `theme.css` wraps them across five. The
two files change in the same commit.

## Colors

Almost every colour `styles.css` writes comes from the roles above (nine literals are the
exception — see *Off-token colours* below). `primary` (`#0a84ff`) colours everything
interactive: links, primary buttons, the nav's active state. `accent` (`#bf5af2`) is decorative only — glow shadows and gradient
blobs. `surface` is the frosted panel: white at 8 % opacity over the page gradient, which is
what every card, the nav bar, and the mobile action bar are made of. `error`, `success` and
`warning` appear only on status badges and validation messages. Hover shades are the one
exception to the no-extra-tokens rule: `primary-hover` is stored, because the glass
compositing makes `color-mix()` at the use site unpredictable.

**There is no light scheme.** `:root` is already the Apple Dark palette, and the
`@media (prefers-color-scheme: dark)` block swaps it for a *different* dark palette — a blue
gradient instead of a grey one. A user whose OS is set to light gets the grey dark theme, not
a light one. That is the current behaviour, recorded here as fact rather than as intent.

### Off-token colours

Nine literal `rgba()` values sit in `styles.css`, against the rule below. Three duplicate a
token that already exists, and three name a colour this theme does not have:

| `styles.css` | Value | Problem |
|---|---|---|
| :411 | `rgba(255, 255, 255, 0.18)` | duplicates `border` |
| :418 | `rgba(255, 255, 255, 0.3)` | duplicates `border-light` |
| :642 | `rgba(0, 0, 0, 0.95)` | near-duplicate of `bg` |
| :307, :375 | `rgba(168, 85, 247, …)` | a purple that is **not** `accent` (`#bf5af2` is `rgb(191, 90, 242)`) |
| :544 | `rgba(239, 68, 68, 0.5)` | a red that is **not** `error` |
| :549 | `rgba(16, 185, 129, 0.5)` | a green that is **not** `success` |
| :130, :446 | `rgba(0, 0, 0, 0.2)`, `rgba(31, 38, 135, 0.37)` | shadow tints with no token |

The three colour-role impostors are the ones that matter: change `accent`, `error` or
`success` in `theme.css` and those glows keep the old hue.

### Measured contrast (2026-09-07)

Measured with the WCAG 2.x relative-luminance formula, compositing every translucent value
over the ground beneath it. The page ground is a gradient, so each row gives the darkest stop
(`#1c1c1e`) and the lightest (`#2c2c2e`); text sits on both.

| Pair | On `#1c1c1e` | On `#2c2c2e` | On `surface` | Floor | |
|---|---|---|---|---|---|
| `text` | 15.63:1 | 12.80:1 | 12.45:1 | 4.5:1 | pass |
| `success` | 8.42:1 | 6.89:1 | 6.70:1 | 4.5:1 | pass |
| `warning` | 8.28:1 | 6.78:1 | 6.59:1 | 4.5:1 | pass |
| `text-muted` | 5.22:1 | **4.27:1** | **4.16:1** | 4.5:1 | **fails on the lighter half** |
| `error` | 4.99:1 | **4.09:1** | **3.98:1** | 4.5:1 | **fails on the lighter half** |
| `accent` | 4.83:1 | **3.96:1** | **3.85:1** | 4.5:1 | **fails on the lighter half** |
| `border` | **1.78:1** | **1.79:1** | — | 3:1 | **fails everywhere** |
| `primary-text` on `primary` | **3.65:1** | — | — | 4.5:1 | **fails everywhere** |

In the `prefers-color-scheme: dark` palette, `text-dark` measures 15.52:1 / 14.43:1 / 11.43:1
across the three gradient stops and `text-muted-dark` 6.81:1 / 6.53:1 / 5.52:1 — both pass on
all three.

**Four floors are not met**, and they are design decisions to make, not facts to leave here
quietly:

1. **The primary button** — white on `#0a84ff` is 3.65:1. Every primary action in the app is
   below AA. The baseline's own theme holds this pair at ≥ 7.4:1.
2. **`border` at 1.78:1** — glass borders are nearly invisible against the page. WCAG 1.4.11
   asks 3:1 for the boundary of a UI component, so card and input edges are not perceivable.
3. **`text-muted`** — passes only on the darkest gradient stop. Most muted copy in the app
   sits on a card, where it measures 4.16:1.
4. **`error` and `accent`** as text — same shape, and `error` is the one role that must never
   be missed.

After any colour change, re-measure and update this table.

## Typography

Two stacks, no web fonts and no font files in the repository: `body` for text, `mono` for
code. `"Inter"` and `"JetBrains Mono"` lead the stacks but are **not shipped** — the CSP is
`default-src 'self'`, so nothing is fetched from a font CDN. On a machine without them
installed the stacks fall through to `-apple-system` / `SF Mono`, which is the intended
result on Apple hardware and a plain system font everywhere else.

The size scale is fixed, not fluid — nine steps from `--font-size-xs: 0.75rem` to
`--font-size-4xl: 2.5rem`, with `--font-size-base: 1rem` as the body size. Weights run
`300`–`700`; line heights are `1.2` tight, `1.5` base, `1.75` relaxed.

## Layout

Spacing is a ten-step rem scale, `--space-1: 0.25rem` through `--space-16: 4rem`, all
multiples of `0.25rem`. Three fixed layout values frame the page: `--header-height: 64px`,
`--nav-width: 280px`, `--container-max: 1200px`.

Layout is mobile-first: the base styles are the narrow layout, and wider screens add columns.
Touch targets never go below `--touch-target-min: 44px`. On small screens the nav collapses
behind a checkbox toggle and a fixed `.action-bar` appears at the bottom of the viewport —
there is no JavaScript in either, per the no-inline-script rule.

## Elevation & Depth

Depth is the point of this theme, not something it avoids. Two families of shadow:

- **Plain shadows** — `--shadow-sm` through `--shadow-xl`, black at 10–25 % opacity, for
  ordinary raised surfaces.
- **Glass shadows** — `--glass-shadow` (`0 8px 32px rgba(0, 0, 0, 0.4)`) with its `sm` and
  `lg` steps, plus `--glass-inset`, a 1 px white top edge that reads as the lit rim of a pane.

Three **glow** shadows (`--shadow-glow`, `--shadow-glow-pink`, `--shadow-glow-blue`) are
decorative only and never carry meaning.

The glass itself is `backdrop-filter` blur at four steps: `--glass-blur-sm: 10px`,
`--glass-blur: 20px`, `--glass-blur-lg: 30px`, `--glass-blur-xl: 50px`. A browser without
`backdrop-filter` falls back to the flat translucent fill, which is legible but not frosted.

## Shapes

Six radii, not one: `--radius-sm: 8px`, `--radius-md: 12px` (the default and the one in the
frontmatter), `--radius-lg: 16px`, `--radius-xl: 24px`, `--radius-2xl: 32px`, and
`--radius-full: 9999px` for pills and circles. Larger panels take larger radii — the corner
grows with the box, which is what keeps a blurred pane from looking pinched.

## Components

Every component composes the roles above. `.sr-only` is present, so text alternatives have
somewhere to go.

**Styled and shipping:**

- **Button** — `.btn`, with `.btn-primary` (filled `primary`), `.btn-secondary`,
  `.btn-outline`, `.btn-ghost`, and `.btn-sm` / `.btn-lg` sizes.
- **Card** — `.card` with `__header`, `__body`, `__footer`. The frosted panel; the main unit
  of the whole UI.
- **Nav** — `.nav` with `__brand`, `__links`, `__link`, and the checkbox-driven
  `__toggle` / `__toggle-label` / `__hamburger` for narrow screens.
- **Action bar** — `.action-bar` with `__item`. Fixed bottom nav, mobile only.
- **Form field** — `.form-group`, `.form-label`, `.form-input`, `.form-row`, `.form-actions`.
- **Utilities** — `.container`, `.flex-center`, `.flex-column`, spacing (`.mb-4`, `.mt-4`,
  `.my-4`, `.p-4`, `.px-4`, `.py-4`, `.gap-1`/`2`/`4`/`6`), and text helpers
  (`.text-center`, `.text-muted`, `.text-success`, `.text-error`, `.sr-only`).

**Used by the templates but not styled at all** — these classes appear in the markup and
match nothing in any stylesheet, so they render with `base.css`'s reset and nothing more:

- **`.table` / `.badge`** — the reservations list. `base.css` strips `border-collapse`
  spacing and list markers, so the table renders as unformatted rows and each status badge
  as bare text.
- **`.alert` / `.alert-danger`** — the error page's message.

**States.** `:hover` is styled in ten places and `:active` in one. **`:focus-visible` appears
in no stylesheet**, so keyboard focus has no visible ring anywhere in the app; `:disabled` is
unstyled too. The design-system rule is that every interactive state is styled — hover,
focus-visible, active, disabled, loading — and three of the five are missing.

## Do's and Don'ts

- Do take every colour from the roles above. Don't write a literal colour value outside
  `theme.css` — nine already exist in `styles.css` and are listed above; don't add a tenth.
- Do keep `base.css` free of colour, font and theme values — it is a reset, and layer 2 is
  where values live.
- Don't add a theme toggle. The scheme follows the OS, even though both branches are
  currently dark.
- Do re-measure the contrast table after any colour change, and update the numbers here in
  the same commit.
- Don't put behaviour in an inline `<script>` or rules in an inline `<style>`. The CSP
  carries no `'unsafe-inline'`, so both are dead on arrival; interactivity is htmx
  attributes and CSS.
- Do give every new interactive element a `:focus-visible` ring, even though the existing
  ones lack one. Don't copy that gap forward.
- Do keep glow shadows decorative. Don't use one to signal state — `success`, `warning` and
  `error` carry meaning, glows do not.
