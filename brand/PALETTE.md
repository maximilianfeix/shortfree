# shortfree

The mark is a domain dot beside a closing cursor bracket: a short name held in a terminal selection. Its two solid shapes remain distinct at 16 px. Cobalt is the identity color; green is reserved for availability. The symbol is hand-drawn geometry, with no registrar artwork, external assets, or web fonts.

## Theme colors

| Role | Dark | Light | Use |
| --- | --- | --- | --- |
| Background | `#0C1426` | `#F7F9FD` | Page and banner canvas |
| Surface | `#152139` | `#FFFFFF` | Domain rows and status chips |
| Border | `#637590` | `#76849B` | Meaningful outlines and controls |
| Text | `#F3F6FC` | `#14213A` | Headings, names, primary copy |
| Muted | `#A4B2C8` | `#526079` | Supporting text and metadata |
| Accent | `#91ACFF` | `#315BEE` | Links, focus indicators, emphasis |
| Free / success | `#69E0AA` | `#146C49` | Available domain status |
| Taken | `#FF9EAA` | `#B32D49` | Registered domain status |
| Warning | `#FFD079` | `#8A5700` | Stale, uncertain, or pending checks |

The mark always uses `#FFFFFF` on `#315BEE`, including in dark mode. The dark accent is a lighter companion for accessible text; it does not replace the mark's cobalt tile.

## Contrast and accessibility

Ratios use the WCAG relative-luminance formula for opaque sRGB colors. The values below give the lower contrast against either the theme background or surface. All text tokens meet WCAG AA for normal text (4.5:1); borders meet the 3:1 non-text threshold against both backgrounds.

| Foreground | Dark minimum | Light minimum |
| --- | ---: | ---: |
| Text | 14.83:1 | 15.23:1 |
| Muted | 7.47:1 | 6.02:1 |
| Accent | 7.30:1 | 5.16:1 |
| Free / success | 9.82:1 | 6.09:1 |
| Taken | 8.19:1 | 5.88:1 |
| Warning | 11.12:1 | 5.79:1 |

White on the cobalt mark has 5.44:1 contrast. Use text tokens directly on the background or surface; adding opacity or arbitrary status fills requires another contrast check. Background and surface alone do not distinguish essential boundaries: use the border token where that distinction matters. Pair status colors with explicit words such as “free”, “taken”, or “pending”; never rely on color alone.

SVG text remains selectable and uses system fonts, so glyph shapes and widths vary by operating system. The banners leave room for those differences. At small embedded sizes, use the standalone mark or favicon instead of shrinking the full banner's text.

## Typography

System sans, for taglines, explanations, and interface copy:

```css
font-family: system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
```

Monospace, for the wordmark, domain names, and registry metadata:

```css
font-family: ui-monospace, 'SF Mono', Menlo, Consolas, monospace;
```

Use bold lowercase monospace for `shortfree`, with modest tight tracking at display sizes. Use regular monospace for domain names and regular sans for explanatory copy. No font downloads are required.

## Asset usage

- `mark.svg`: 512 × 512 primary symbol, on its own cobalt tile.
- `favicon.svg`: matching 32 × 32 geometry; also suitable at 16 × 16.
- `banner-dark.svg` and `banner-light.svg`: 1200 × 300 README banners.
- `social-preview.svg`: 1280 × 640 social image with illustrative domain statuses. The examples are not live availability claims.

Keep the mark's proportions and colors intact. Allow at least one dot-width of clear space around the tile when composing new layouts. The banners and preview include their own spacing. All SVG artwork is static, self-contained, and hand-written.
