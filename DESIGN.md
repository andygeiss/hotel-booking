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
  text-muted: "#aeaeb2"
  accent: "#bf5af2"
  primary: "#0a84ff"
  primary-hover: "#409cff"
  primary-text: "#000000"
  link: "#d9a0f7"
  link-hover: "#ff9aae"
  focus: "#ffd60a"
  error: "#ff8a80"
  error-solid: "#ff453a"
  success: "#30d158"
  warning: "#ff9f0a"
  border: "rgba(255, 255, 255, 0.35)"
  border-light: "rgba(255, 255, 255, 0.5)"
  bg-dark: "#0f0f1a"
  bg-gradient-dark: "linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%)"
  surface-dark: "rgba(255, 255, 255, 0.08)"
  surface-solid-dark: "rgba(255, 255, 255, 0.12)"
  text-dark: "rgba(255, 255, 255, 0.95)"
  text-muted-dark: "rgba(255, 255, 255, 0.6)"
  border-dark: "rgba(255, 255, 255, 0.4)"
  accent-dark: "#bf5af2"
  error-dark: "#ff8a80"
typography:
  body:
    fontFamily: '"Inter", "SF Pro Display", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif'
  mono:
    fontFamily: '"JetBrains Mono", "Fira Code", "SF Mono", monospace'
rounded:
  radius: "12px"
components:
  button:
    backgroundColor: "{colors.accent}"
    textColor: "{colors.primary-text}"
    rounded: "{rounded.radius}"
---

# Design: Hotel Booking

## Overview

This file is the project's design system: the theme values, contrast floors, and component
inventory every page is styled from. It is written for whoever styles a page — usually an
AI coding agent — and design tools parse the same file.

The theme is **glassmorphism**: translucent white panels blurred over a dark gradient, with
a purple-to-pink gradient on anything filled and interactive. It is Apple-derived, and it
is dark in both colour schemes — see *Colors*.

Styles live in three layers, loaded in this order, rather than in one `app.css`:

| Layer | File | Holds |
|---|---|---|
| 1 | `cmd/server/assets/static/css/base.css` | Reset only. No colours, no fonts, no theme values. |
| 2 | `cmd/server/assets/static/css/theme.css` | Every token below, as custom properties on `:root`. |
| 3 | `cmd/server/assets/static/css/styles.css` | Components and layout. Reads tokens — with five exceptions, listed under *Colors*. |

Every CSS value in this file is identical to the same value in `theme.css` — the gradients
are the only ones rewritten, onto one line, because `theme.css` wraps them across five. The
two files change in the same commit.

## Colors

Almost every colour `styles.css` writes comes from the roles above (five literals are the
exception — see *Off-token colours*).

`primary` (`#0a84ff`) is the semantic blue. `accent` (`#bf5af2`) with `#ff375f` forms the
gradient every filled button wears — `.btn-primary` is that gradient, not a flat `primary`
fill. `surface` is the frosted panel: white at 8 % opacity over the page gradient, which is
what every card, the nav bar and the mobile action bar are made of. `link` and `link-hover`
are lightened members of the same purple-to-pink family, chosen so link text clears its
floor on a card. `error`, `success` and `warning` carry meaning and appear only on badges,
alerts and validation messages.

Three roles exist to keep contrast honest and are easy to misuse:

- **`primary-text` is black, not white.** It is the text colour for a *filled* surface —
  `.btn-primary`, `.btn-secondary`, `.btn-danger`. Those gradients are light enough that
  white on them measures 3.52:1 and black 5.96:1. It is not a general "bright text"
  colour; the glass buttons and the nav use `text` and `#ffffff`.
- **`error-solid` (`#ff453a`) never carries text.** It is the fill for `.btn-danger` and
  the tint under a danger badge. `error` (`#ff8a80`) is the text role.
- **`focus` (`#ffd60a`) is used by exactly one thing**, the focus ring. It is the only hue
  nothing else in the theme uses, which is what makes the ring read as focus.

**There is no light scheme.** `:root` is already the Apple Dark palette, and the
`@media (prefers-color-scheme: dark)` block swaps it for a *different* dark palette — a blue
gradient instead of a grey one. A user whose OS is set to light gets the grey dark theme, not
a light one. That is the current behaviour, recorded here as fact rather than as intent.

### Off-token colours

Five literal `rgba()` values remain in `styles.css`, against the rule below. All five are
shadows or decorative tints; none is text, a border, or anything a contrast floor applies to:

| `styles.css` | Value | Note |
|---|---|---|
| :130 | `rgba(0, 0, 0, 0.2)` | brand text-shadow |
| :307, :375 | `rgba(168, 85, 247, …)` | a purple that is **not** `accent` (`#bf5af2` is `rgb(191, 90, 242)`) — decorative glow only |
| :446 | `rgba(31, 38, 135, 0.37)` | action-bar shadow tint |
| :815 | `rgba(0, 0, 0, 0.95)` | mobile nav backdrop |

The two `rgba(168, 85, 247, …)` glows are the ones worth fixing: change `accent` in
`theme.css` and they keep the old hue.

### Measured contrast (2026-09-07)

Measured with the WCAG 2.x relative-luminance formula, compositing every translucent value
over the ground beneath it. **The worst ground in the app is a glass card over the lighter
gradient stop** — white at 8 % over `#2c2c2e` — so that is the column that decides a pass.
Text floor 4.5:1; borders, focus rings and filled shapes 3:1.

**Text**

| Role | On the page (`#2c2c2e`) | On a card | |
|---|---|---|---|
| `text` `#f5f5f7` | 12.80:1 | 9.95:1 | pass |
| `text-muted` `#aeaeb2` | 6.30:1 | 4.90:1 | pass |
| `success` `#30d158` | 6.89:1 | 5.36:1 | pass |
| `error` `#ff8a80` | 6.10:1 | 4.75:1 | pass |
| `link` `#d9a0f7` | 6.85:1 | 5.33:1 | pass |
| `link-hover` `#ff9aae` | 6.96:1 | 5.41:1 | pass |

**Filled surfaces, all carrying `primary-text` `#000000`**

| Component | Fill | Text on it | Fill vs page |
|---|---|---|---|
| `.btn-primary` | `#bf5af2` → `#ff375f` | 5.96:1 at both stops | 3.96:1 |
| `.btn-secondary` | `#64d2ff` → `#0a84ff` | 12.20:1 → 5.76:1 | 8.10:1 → 3.82:1 |
| `.btn-danger` | `#ff453a` | 6.16:1 | 4.09:1 |

**Non-text**

| | On `#1c1c1e` | On `#2c2c2e` | |
|---|---|---|---|
| `border` white @ 0.35 | 3.20:1 | 3.07:1 | pass |
| `focus` ring `#ffd60a` | — | 9.87:1 (7.68:1 on a card) | pass |

**Badges and alerts** put `text` on a hue tinted 20–25 % over the card, which measures
between 6.29:1 (`badge-info`) and 8.30:1 (`.alert-danger`). The hue reads from the fill; the
legibility comes from near-white text, because the hues themselves do not clear 4.5:1 at
badge size.

**`prefers-color-scheme: dark`** — `text-dark` measures 12.41 / 11.49 / 9.16:1 across the
three gradient stops, `text-muted-dark` 5.98 / 5.59 / 4.72:1, and `border-dark` at 0.40
measures 3.80 / 3.70 / 3.30:1. That border is `0.40` rather than the default scheme's `0.35`
because this palette's lightest stop (`#0f3460`) is lighter, and `0.35` measures 2.87:1
there.

Every pair above meets its floor. After any colour change, re-measure and update these
tables.

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
`--radius-full: 9999px` for pills and circles — badges use that last one. Larger panels take
larger radii: the corner grows with the box, which is what keeps a blurred pane from looking
pinched.

## Components

Every component composes the roles above.

- **Button** — `.btn` (glass), with `.btn-primary` and `.btn-secondary` (gradient fills),
  `.btn-danger` (solid `error-solid`), `.btn-outline`, `.btn-ghost`, and `.btn-sm` /
  `.btn-lg` sizes. Every filled variant carries `primary-text`; the glass ones carry `text`.
- **Card** — `.card` with `__header`, `__body`, `__footer`. The frosted panel; the main unit
  of the whole UI. Its border reads `border` from the token, so a contrast change reaches it.
- **Nav** — `.nav` with `__brand`, `__links`, `__link`, and the checkbox-driven
  `__toggle` / `__toggle-label` / `__hamburger` for narrow screens.
- **Action bar** — `.action-bar` with `__item`. Fixed bottom nav, mobile only.
- **Table** — the bare `table`, `th` and `td` elements are styled: a glass panel with a
  darkened header row and `border` rules between rows. The `.table` **class** in the markup
  matches nothing and is inert; the element selectors do the work.
- **Badge** — `.badge` plus `badge-primary` / `-secondary` / `-success` / `-danger` /
  `-warning` / `-info`, one per reservation status. A pill-radius tinted well with `text` on
  it; the tint and its border are `color-mix()`ed at the use site rather than stored as
  tokens.
- **Alert** — `.alert` and `.alert-danger`, the same idea at paragraph size. The markup
  carries `role="alert"`, so colour is reinforcement and never the only signal.
- **Form field** — `.form`, `.form-group`, `.form-label`, `.form-input`, `.form-row`,
  `.form-actions`, and `.form-inline` for a form that must sit inside a table cell.
- **Detail grid** — `.detail-grid` and `.detail-item`, an auto-fitting label-over-value grid
  for the reservation detail page.
- **Utilities** — `.container`, `.flex-center`, `.flex-column`, spacing (`.mb-2`, `.mb-4`,
  `.mt-4`, `.my-4`, `.p-4`, `.px-4`, `.py-4`, `.gap-1`/`2`/`4`/`6`), and text helpers
  (`.text-center`, `.text-muted`, `.text-success`, `.text-error`, `.sr-only`).

**States.** `:hover` and `:active` are styled. `:focus-visible` puts a 3 px `focus` ring with
a 2 px offset on every focusable element, restated on buttons, inputs, links and nav items so
no component can quietly drop it, and moved onto the visible label for the visually hidden
nav-toggle checkbox. `:disabled` dims to 50 % and removes the lift and shadow. There is no
loading indicator, because no interaction in this app is slow enough to need one.

## Do's and Don'ts

- Do take every colour from the roles above. Don't write a literal colour value outside
  `theme.css` — five remain in `styles.css` and are listed above; don't add a sixth.
- Do read borders from `--glass-border` rather than retyping the rgba. `.card` had its own
  copy, so a contrast fix to the token did not reach the app's most-used component until the
  copy was removed.
- Do keep `base.css` free of colour, font and theme values — it is a reset, and layer 2 is
  where values live.
- Don't add a theme toggle. The scheme follows the OS, even though both branches are
  currently dark.
- Do re-measure the contrast tables after any colour change, and update the numbers here in
  the same commit. Measure against a glass card over the **lighter** gradient stop; it is the
  worst ground in the app, and measuring against the darker one flatters every value.
- Do check what actually paints before trusting a token name. `.btn-primary` is a gradient of
  `accent` and `#ff375f`, not a `primary` fill — measuring the pair the token name suggests
  gives a number for a combination that never renders.
- Don't put behaviour in an inline `<script>` or rules in an inline `<style>`. The CSP
  carries no `'unsafe-inline'`, so both are dead on arrival; interactivity is htmx
  attributes and CSS.
- Do give every new interactive element a `:focus-visible` ring.
- Do keep glow shadows decorative. Don't use one to signal state — `success`, `warning` and
  `error` carry meaning, glows do not.
