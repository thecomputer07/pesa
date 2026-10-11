# PWA raster icons

The single source of truth is `web/public/icon.svg` (kitchen scale weighing a
wheat ear, on the brand blue). Every PNG here, plus `public/favicon-32.png`, is
derived from it — never edit the PNGs by hand.

## Regenerate

```sh
cd web
node scripts/render-icons.mjs .            # writes the PNGs below
node scripts/render-icons.mjs . /tmp/prev  # optional: preview sheet only
```

The script rasterizes the SVG with Chromium via Playwright (`channel: "msedge"`,
so the Edge already on Windows is used and no browser download is needed).

## Output files

- `icons/icon-192.png`, `icons/icon-512.png` — manifest, purpose `any`
- `icons/icon-maskable-512.png` — purpose `maskable`: full-bleed brand colour,
  artwork inside the ~80% Android safe zone
- `icons/apple-touch-icon.png` — 180×180, opaque background (iOS paints
  transparent corners black and applies its own rounding)
- `../favicon-32.png` — PNG favicon fallback for browsers without SVG favicons
