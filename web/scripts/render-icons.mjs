// Rasterize web/public/icon.svg with Chromium (Playwright).
// usage (from web/): node scripts/render-icons.mjs . [preview-dir]
// Uses the installed Edge, so no Playwright browser download is needed.
import { chromium } from "playwright";
import fs from "node:fs";
import path from "node:path";

const [webDir, prevDir] = process.argv.slice(2);
const svg = fs.readFileSync(path.join(webDir, "public/icon.svg"), "utf8");
const toUrl = (src) => "data:image/svg+xml;base64," + Buffer.from(src).toString("base64");
const dataUrl = toUrl(svg);
// Full-bleed variant: square background, no rounded corners. The OS applies
// its own mask (Android maskable, iOS), and the artwork already sits inside
// the ~80% safe zone.
const bleedUrl = toUrl(svg.replace(/(<rect width="512" height="512") rx="\d+"/, "$1"));
if (bleedUrl === dataUrl) throw new Error("icon.svg: background rect not found");

const browser = await chromium.launch({ channel: "msedge" });
const page = await browser.newPage();

async function shot(size, out, { bg = "transparent", inner = size, src = dataUrl } = {}) {
  await page.setViewportSize({ width: size, height: size });
  await page.setContent(
    `<html><body style="margin:0;background:${bg};display:grid;place-items:center;width:${size}px;height:${size}px">` +
      `<img src="${src}" width="${inner}" height="${inner}"></body></html>`,
  );
  await page.screenshot({ path: out, omitBackground: bg === "transparent" });
}

if (prevDir) {
  // Preview sheet: big, home-screen size, favicon sizes, on light and dark.
  await page.setViewportSize({ width: 900, height: 360 });
  const sizes = [256, 120, 64, 32, 16];
  const row = (bg) =>
    `<div style="display:flex;gap:28px;align-items:center;padding:24px;background:${bg}">` +
    sizes.map((s) => `<img src="${dataUrl}" width="${s}" height="${s}">`).join("") +
    `<div style="font:600 28px system-ui;color:${bg === "#fff" ? "#0f172a" : "#fff"};display:flex;align-items:center;gap:10px"><img src="${dataUrl}" width="32" height="32">Pesa</div></div>`;
  await page.setContent(`<html><body style="margin:0">${row("#fff")}${row("#0f172a")}</body></html>`);
  await page.screenshot({ path: path.join(prevDir, "icon-preview.png"), fullPage: true });
} else {
  const out = path.join(webDir, "public/icons");
  fs.mkdirSync(out, { recursive: true });
  await shot(192, path.join(out, "icon-192.png"));
  await shot(512, path.join(out, "icon-512.png"));
  // iOS fills transparent corners with black and rounds the icon itself.
  await shot(180, path.join(out, "apple-touch-icon.png"), { src: bleedUrl });
  await shot(512, path.join(out, "icon-maskable-512.png"), { src: bleedUrl });
  await shot(32, path.join(webDir, "public/favicon-32.png"));
}
await browser.close();
