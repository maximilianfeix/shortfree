# shortfree

## Rationale

shortfree presents domain availability as dependable public infrastructure, with the restraint of a developer tool built to last.
The domain dot and waiting cursor turn a familiar fragment of syntax into an ownable mark.
An ink rounded square gives that fragment a stable silhouette across repositories, websites, and MCP clients.
Green means free, while paper and ink carry the identity without decorative effects.
The wordmark uses original filled vector outlines on an 8-unit construction grid, with 32-unit letter cells and a 40-unit advance.
Its shared stems, open counters, and compact rhythm connect the identity to code without depending on a particular font.
System typography, explicit status words, and generous spacing keep the data readable in both themes.

## Assets

| File | Canvas | Use |
| --- | --- | --- |
| `mark.svg` | 512 × 512 | Full-color master; transparent outside the rounded square |
| `mark-on-dark.svg` | 512 × 512 | Full-color mark with a 1 px inner border for dark backgrounds |
| `mark-mono.svg` | 512 × 512 | Single `currentColor` silhouette; dot and cursor are transparent knockouts |
| `favicon.svg` | 32 × 32 | Integer-coordinate favicon; also resolves cleanly at 16 × 16 |
| `wordmark-dark.svg` | 368 × 80 | Paper wordmark on transparency, for dark backgrounds |
| `wordmark-light.svg` | 368 × 80 | Ink wordmark on transparency, for light backgrounds |
| `lockup-dark.svg` | 480 × 96 | Horizontal lockup on transparency, for dark backgrounds |
| `lockup-light.svg` | 480 × 96 | Horizontal lockup on transparency, for light backgrounds |
| `banner-dark.svg` | 1200 × 300 | README banner with opaque ink background |
| `banner-light.svg` | 1200 × 300 | README banner with opaque paper background |
| `social-preview.svg` | 1280 × 640 | Dark social card; all content stays at least 64 px from the sides |

Dark/light suffixes describe the intended background. Wordmarks contain only path geometry, plus grouping and accessibility metadata; there are no font dependencies. Each SVG is self-contained. The three banner domains are illustrative examples supplied for the layout, not live availability claims; do not treat them as verified results or infer a dropping prediction from this artwork.

## Color roles

| Token | Dark | Light | Role |
| --- | --- | --- | --- |
| `bg` | `#0B0C0E` | `#FAFAF7` | Page background |
| `surface` | `#16181C` | `#FAFAF7` | Panels; light panels share the page color and use rules or spacing |
| `surface-2` | `#24272D` | `#C9CDD3` | Selected or inset surfaces |
| `border` | `#3A3F47` | `#C9CDD3` | Decorative separators only |
| `control-border` | `#8A919C` | `#3A3F47` | Required control outlines and meaningful boundaries |
| `text` | `#FAFAF7` | `#0B0C0E` | Primary text |
| `muted` | `#8A919C` | `#3A3F47` | Secondary text; never reduce opacity |
| `accent` / `free` | `#1FC16B` | `#1FC16B` | Free state, accent fills, mark dot |
| `free-text-light` | — | `#0E7A43` | Free status text directly on paper; 5.17:1 |
| `warning-text-light` | — | `#8A5A00` | Dropping status text directly on paper; 5.67:1 |
| `taken` | `#E5484D` | `#E5484D` | Taken or negative state |
| `warning` | `#F5A524` | `#F5A524` | Dropping, stale, or uncertain state |
| `focus-ring` | `#1FC16B` | `#0B0C0E` | Keyboard focus outline on ordinary surfaces |
| `status-bg` | `#0B0C0E` | `#0B0C0E` | Inverse chips for colored status text |
| `on-accent` | `#0B0C0E` | `#0B0C0E` | Text on solid green, red, or amber fills |
| `inverse-text` | `#FAFAF7` | `#FAFAF7` | Text on ink surfaces |

The banner's extra-quiet dark table rules use `surface-2`; they are decorative. Green, red, and amber retain their fixed values in both themes. On light surfaces, use `free-text-light` or `warning-text-light` for free or dropping text directly on paper. Other status text uses an ink chip, or an ink label on a solid status fill. Never use paper text on a status fill. Always write the state as a word; color is an additional cue.

## Measured text contrast

Ratios below use WCAG relative luminance in sRGB: linearize each channel (`c / 12.92` for `c ≤ 0.04045`, otherwise `((c + 0.055) / 1.055)^2.4`), calculate `L = 0.2126 R + 0.7152 G + 0.0722 B`, then `(Llighter + 0.05) / (Ldarker + 0.05)`. Values are rounded to two decimals after calculation. Every recommended normal-text combination passes 4.5:1; no large-text exception is needed.

| Theme / foreground | Background | Ratio |
| --- | --- | ---: |
| Dark primary `#FAFAF7` | bg `#0B0C0E` | 18.71:1 |
| Dark primary `#FAFAF7` | surface `#16181C` | 17.00:1 |
| Dark primary `#FAFAF7` | surface-2 `#24272D` | 14.31:1 |
| Dark muted `#8A919C` | bg `#0B0C0E` | 6.16:1 |
| Dark muted `#8A919C` | surface `#16181C` | 5.60:1 |
| Dark muted `#8A919C` | surface-2 `#24272D` | 4.71:1 |
| Either theme free `#1FC16B` | status-bg `#0B0C0E` | 8.29:1 |
| Either theme taken `#E5484D` | status-bg `#0B0C0E` | 5.00:1 |
| Either theme warning `#F5A524` | status-bg `#0B0C0E` | 9.59:1 |
| Light free-text-light `#0E7A43` | bg / surface `#FAFAF7` | 5.17:1 |
| Light warning-text-light `#8A5A00` | bg / surface `#FAFAF7` | 5.67:1 |
| Light primary `#0B0C0E` | bg / surface `#FAFAF7` | 18.71:1 |
| Light primary `#0B0C0E` | surface-2 `#C9CDD3` | 12.26:1 |
| Light muted `#3A3F47` | bg / surface `#FAFAF7` | 10.13:1 |
| Light muted `#3A3F47` | surface-2 `#C9CDD3` | 6.64:1 |
| Either theme on-accent `#0B0C0E` | free fill `#1FC16B` | 8.29:1 |
| Either theme on-accent `#0B0C0E` | taken fill `#E5484D` | 5.00:1 |
| Either theme on-accent `#0B0C0E` | warning fill `#F5A524` | 9.59:1 |
| Either theme inverse-text `#FAFAF7` | ink `#0B0C0E` | 18.71:1 |

Use the fixed status colors only on `status-bg`; all three pass there. The light-theme text tokens above also pass directly on paper. In dark layouts, `status-bg` may be a larger ink region instead of an individual chip. Do not move red text onto `surface-2`: it measures 3.82:1. Green, red, and amber on paper measure only 2.26:1, 3.74:1, and 1.95:1 respectively, so none is recommended for body text on paper. Border tokens are not text colors.

For non-text controls, dark `control-border` against bg / surface / surface-2 measures 6.16 / 5.60 / 4.71:1; light `control-border` against bg / surface-2 measures 10.13 / 6.64:1. Dark green focus rings measure 8.29 / 7.53 / 6.34:1 on the three dark surfaces; light ink rings measure 18.71 / 12.26:1 on the two light surfaces. Use a 2 px outline with 4 px offset so the ring lies against the surrounding page or panel, including around filled actions. Decorative rules must never be the sole means of identifying an interactive control.

## Typography, spacing, and geometry

- Body and UI: `system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif`. Default 16 px / 24 px; secondary UI 14 px / 20 px. No web fonts.
- Domains, counts, timestamps, and TLDs: `ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace`. Use tabular numbers, left-aligned domains, and right-aligned numeric columns. Status text stays lowercase.
- Spacing: 4 px base; use 4, 8, 12, 16, 24, 32, 40, 48, 64, and 96 px. Default panel padding 24 px; section spacing 48–64 px.
- UI radii: 0 px for rules and data tables, 4 px for chips, 8 px for inputs, 12 px for panels. Avoid gratuitous pills.
- Mark radius: nominally 22% of its side, snapped to 112 px at 512 and 7 px at 32 (21.875%). The dot is 64 px in diameter at 512; the cursor is 160 × 32 px. Dot center `(160, 256)`; cursor origin `(256, 256)`. Both finish at y = 288.
- Letterforms: ascender y = 0, x-height top y = 16, baseline y = 64; 8-unit stems, 32-unit cells, 8-unit inter-cell space. The `t` ascender starts at y = 4 for optical balance. These are filled shapes, not stroked text.

## Usage

Clear space around a mark is at least one quarter of its displayed side. Around a standalone wordmark, reserve one quarter of its 64-unit glyph height; around a lockup, reserve one quarter of the mark side, measured from the visible artwork. Transparent canvas padding is not the full required clear space. The master lockup uses an 80-unit mark and a 24-unit gap before the wordmark; keep that relationship when scaling.

Minimum mark size is 16 × 16 px; prefer 32 px and use `favicon.svg` at 16 or 32 so the dot is 2 or 4 px wide and the cursor is 1 or 2 px high. Render favicons at integer positions. Minimum standalone wordmark canvas width is 138 px (24 px visible glyph height); minimum lockup canvas width is 180 px (24 px visible glyph height, 30 px mark). For smaller slots, use the mark alone. Prefer complete banners at 900 px or wider so the secondary table remains readable; use the lockup for narrow placements. Keep social artwork uncropped at 2:1; its outermost content is inset 64 px horizontally and at least 80 px vertically, exceeding the required 40 px safety margin.

Use `mark-on-dark.svg` on dark backgrounds and `mark.svg` on light backgrounds. Dark lockups and banners include the same #3A3F47 inner border, 1 px at each asset’s native size; the tile’s outer geometry stays unchanged. For monochrome reproduction, embed `mark-mono.svg` inline and set `color` on its container; an SVG loaded through `<img>` does not inherit the page's `currentColor` and defaults to black. Use paper on dark backgrounds or ink on light backgrounds; the knocked-out dot and cursor reveal the host surface. Keep the color master distinct from the reversed monochrome stencil.

Do not stretch, rotate, add outlines beyond the supplied dark-background tile border, add shadows, recolor individual letters, redraw the punctuation, change the mark's radius, or substitute typed text for the wordmark. Do not add gradients, glows, sparkles, emoji, generic technology symbols, or ornamental circuitry. Do not animate availability into a promise: daily scans describe a verification cadence, and names can change state between scans. Keep every status coupled to its word and, in live product UI, its verification time.

## Ready-to-paste CSS

```css
:root {
  color-scheme: dark;
  --bg: #0B0C0E;
  --surface: #16181C;
  --surface-2: #24272D;
  --border: #3A3F47;
  --control-border: #8A919C;
  --text: #FAFAF7;
  --muted: #8A919C;
  --accent: #1FC16B;
  --free: var(--accent);
  --free-text-light: #0E7A43;
  --warning-text-light: #8A5A00;
  --taken: #E5484D;
  --warning: #F5A524;
  --focus-ring: #1FC16B;
  --status-bg: #0B0C0E;
  --on-accent: #0B0C0E;
  --inverse-text: #FAFAF7;
  --font-body: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  --font-mono: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas,
    "Liberation Mono", monospace;
  --space-1: 4px;
  --space-2: 8px;
  --space-3: 12px;
  --space-4: 16px;
  --space-6: 24px;
  --space-8: 32px;
  --space-10: 40px;
  --space-12: 48px;
  --space-16: 64px;
  --space-24: 96px;
  --radius-none: 0;
  --radius-sm: 4px;
  --radius-md: 8px;
  --radius-lg: 12px;
}

@media (prefers-color-scheme: light) {
  :root {
    color-scheme: light;
    --bg: #FAFAF7;
    --surface: #FAFAF7;
    --surface-2: #C9CDD3;
    --border: #C9CDD3;
    --control-border: #3A3F47;
    --text: #0B0C0E;
    --muted: #3A3F47;
    --accent: #1FC16B;
    --free: var(--accent);
    --taken: #E5484D;
    --warning: #F5A524;
    --focus-ring: #0B0C0E;
    --status-bg: #0B0C0E;
    --on-accent: #0B0C0E;
    --inverse-text: #FAFAF7;
  }
}

body {
  margin: 0;
  background: var(--bg);
  color: var(--text);
  font-family: var(--font-body);
  font-size: 16px;
  line-height: 1.5;
}

.domain, .number, .tld, .status {
  font-family: var(--font-mono);
  font-variant-numeric: tabular-nums;
}

.status {
  display: inline-block;
  padding: 4px 12px;
  border-radius: var(--radius-sm);
  background: var(--status-bg);
  color: var(--inverse-text);
}
.status[data-state="free"] { color: var(--free); }
.status[data-state="taken"] { color: var(--taken); }
.status[data-state="dropping"] { color: var(--warning); }

@media (prefers-color-scheme: light) {
  .status[data-state="free"],
  .status[data-state="dropping"] {
    padding: 0;
    background: transparent;
  }
  .status[data-state="free"] { color: var(--free-text-light); }
  .status[data-state="dropping"] { color: var(--warning-text-light); }
}

:focus-visible {
  outline: 2px solid var(--focus-ring);
  outline-offset: 4px;
}
```
