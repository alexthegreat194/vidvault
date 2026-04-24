# VIDVAULT design system

This document describes the visual language used in the web UI. The **source of truth** is `src/web/styles.css` (`:root` variables and component rules). Favicon art in `src/web/favicon.svg` matches the same accent colors.

## Principles

- **Dark-first**: Near-black background with layered grays for surfaces. High contrast for primary text; secondary text is deliberately subdued.
- **Editorial / technical**: A display face for the wordmark, monospace for everything else—fits a file/media tool, not a marketing page.
- **Accents are functional**: Chartreuse and coral draw attention to actions, selection, and warnings; mint is reserved for success and drop targets.

## Color

CSS custom properties (see `:root` in `styles.css`):

| Token      | Default   | Role |
|------------|-----------|------|
| `--bg`     | `#0a0a0b` | Page background |
| `--surface`| `#111114` | Cards, inputs, main panels |
| `--surface2`| `#18181c` | Elevated UI (context menu, toasts, select bar) |
| `--border` | `#222228` | Default borders and dividers |
| `--muted`  | `#3a3a44` | Muted chrome, scrollbars, inactive borders |
| `--text`   | `#c8c8d4` | Primary text |
| `--dim`    | `#5a5a6a` | Secondary text, labels, placeholders |
| `--accent` | `#e8ff47` | Primary accent (lime/chartreuse): focus rings, active states, selection, primary CTAs |
| `--accent2`| `#ff6b47` | Secondary accent (coral): wordmark second half, warnings, empty-state hints |
| `--green`  | `#47ffb2` | Success, drag-over affordance, completed uploads |

**Ad hoc colors** in the UI (not tokenized but consistent): pure black video areas (`#000`), very dark thumb background (`#0e0e11`), error red (`#ff5555`) for destructive hovers and error toasts. Modal overlays use `rgba(0,0,0,0.85–0.92)` with `backdrop-filter: blur(4px)`.

## Typography

- **Display / logo**: Bebas Neue — `VID` in `--accent`, `VAULT` in `--accent2`, wide letter-spacing, ~32px in the header.
- **UI**: DM Mono at 13px body; 11–12px for controls and metadata; 10px for small labels (e.g. sidebar section title, path hints), often uppercase with increased letter-spacing for section labels.
- **Weights**: 300/400/500 as loaded from Google Fonts; active buttons and emphasis use medium where needed.

## Layout and shape

- **Default radius**: `--radius: 4px` for inputs, buttons, cards, and small elements.
- **Larger panels**: Modals and floating bars (select bar, some modal shells) use **6px** radius.
- **Shell**: Sticky header; two-column main area—sidebar **220px** + fluid content; gallery uses a responsive grid (`minmax(260px, 1fr)`) with **12px** gap; list view is a single column with **4px** row gap.
- **Spacing**: Header padding 14px × 24px; main content 20px × 24px; consistent 8px / 12px / 16px gaps inside components.

## Components (patterns)

- **Buttons (`.btn`)**: Surface + border, dim text; hover lightens text and border. **Active** state uses `accent` fill and black text. **Upload** is an accent-outlined variant with subtle hover tint.
- **Inputs**: Same surface/border as buttons; **focus** uses `accent` border. Search field has an embedded icon using muted color.
- **Sidebar folders**: Text-only row with left border; active row uses accent text, light yellow tint, accent left border. Drag-over uses `green` tint and border.
- **Cards**: Bordered surface; hover lifts slightly (`translateY(-2px)`), border moves toward `--muted`. Thumbnails 16:9. List view compacts to a row with a fixed thumb height (60px).
- **Overlays**: Context menu and modals use `surface2` or `surface` with a strong shadow; entry animations are short scale/fade (e.g. `modalIn`, `fadeIn`).
- **Toasts**: Bottom-right; success uses green border/text, error uses `#ff5555`.
- **Scrollbars**: Thin (6px), rounded thumb, muted/dim on hover.

## Motion

- Short transitions (~0.1–0.2s) for borders, opacity, and transforms.
- Staggered feel: `fadeIn` for cards and menus; `modalIn` for dialogs.
- Avoiding excessive motion: hover lift is subtle; dragging reduces opacity/scale on cards.

## Icon and brand mark

- **Favicon** (`favicon.svg`): Dark `#1a1a1a` field, top-left **lime** `#e8ff00` (aligns with `--accent`), top-right **coral** `#ff6b47` (aligns with `--accent2`), central white play mark—consistent with the in-app play affordance on thumbnails.

## Maintenance

- Prefer extending **`styles.css` variables** for new features so palette and type stay consistent.
- When adding new screens, match existing **modal** / **form** / **list** padding and the **4px / 6px** radius split.
