# PWA raster icons

The single source of truth is `web/public/icon.svg`. Raster PNGs are produced
at build time (or checked in by the maintainer) from that SVG.

## Generate PNGs with ImageMagick

Requires ImageMagick ≥ 7 with `rsvg` or `inkscape` delegate available.

```sh
cd web/public
# 192×192
magick -background none -density 768 icon.svg -resize 192x192 icons/icon-192.png
# 512×512
magick -background none -density 1536 icon.svg -resize 512x512 icons/icon-512.png
# 512×512 maskable (inner content padded to safe zone ~80%)
magick -background "#1d4ed8" -density 1536 icon.svg -resize 410x410 \
  -gravity center -extent 512x512 icons/icon-maskable-512.png
```

## Generate PNGs with rsvg-convert (alternative, no MagickWand needed)

```sh
rsvg-convert -w 192 -h 192 icon.svg -o icons/icon-192.png
rsvg-convert -w 512 -h 512 icon.svg -o icons/icon-512.png
rsvg-convert -w 410 -h 410 icon.svg -o /tmp/inner.png
magick -size 512x512 xc:"#1d4ed8" /tmp/inner.png -gravity center -composite icons/icon-maskable-512.png
```

## Output files (expected by `manifest.webmanifest`)

- `icons/icon-192.png` — any-purpose, standard density
- `icons/icon-512.png` — any-purpose, high density
- `icons/icon-maskable-512.png` — purpose=maskable, padded to Android safe zone

## Why the single-SVG pattern?

One source, zero drift. The PNGs are derived artefacts; `icon.svg` is the only
file that should be hand-edited. The `Makefile` (root) may wire a `make icons`
target to automate the conversion on CI.
