/**
 * Inline content of the /scopo page (EN).
 * Canonical source: `.planning/legal/scope/scope.en.md`.
 * TODO(phase-4): replace with `vite-plugin-md` import to avoid drift.
 */
export const scopoEN = `# About Pesa / Scope

> *Pesa's primary language is Italian; this is a convenience translation. In case of conflict, the Italian version prevails.*

## Who we are

Pesa is an open-source project created and maintained by Lorenzo Di Mare, an independent developer. The source code is published under the **AGPL v3** license in a public repository: https://github.com/thecomputer07/pesa. There is no company, no investor, no advertising, no paid tier: Pesa is built for people who count carbohydrates and want a simple, transparent tool grounded in certified Italian data.

## What Pesa is for

Pesa computes the **carbohydrate content** of any food item based on its weight in grams. Type the name of a food, enter the weight, and get the carbohydrate value — along with fiber, sugars and kilocalories as contextual information. The data comes from certified Italian sources, not from translated international databases.

You can use Pesa to:

- count carbohydrates in a meal for nutritional reasons (sports, personal dietary regimen, curiosity);
- build a daily and weekly food diary, offline-first;
- create composed recipes from multiple ingredients and reuse them as "composite foods" in the diary;
- scan the barcode of a packaged product at the supermarket.

Pesa works **offline**: once the food catalog has been cached on your device, you can search and compute without a network connection. The diary syncs when you come back online.

## What Pesa is NOT for

Pesa **is not a clinical application**. It does not provide health advice, and it **does not replace consultation with a physician, nutritionist, or dietitian**. It shows public nutritional values — it does not interpret your personal situation.

If you have specific clinical needs — for example a form of diabetes, phenylketonuria, a celiac condition, renal insufficiency, or a dietary regimen prescribed by a professional — your attending physician is the right person to discuss food choices and goals with. Pesa can sit alongside you as a generic calculation tool, much like a kitchen scale or a paper nutritional chart sits alongside anyone: it does not make decisions for you, and the people who follow you clinically should know what you are using.

## Data sources

Pesa integrates two sources, each visibly credited on every food item:

- **CREA — Alimenti e Nutrizione** (alimentinutrizione.it): Italian governmental standard, about 900 base foods. License: CC BY-NC-SA.
- **Open Food Facts** (openfoodfacts.org): collaborative database of packaged products with barcodes, ODbL license. Used by the scanner.

For every search result and every food detail page, Pesa shows a badge with the data source and a link to the original entry. A dedicated "Sources" page lists all credits and applicable licenses.

## Privacy in two lines

- No third-party trackers, no Google Analytics, no advertising pixels.
- No health data: Pesa does not ask for, and does not store, information about your health.
- Only the necessary data: email, password (always hashed), food diary that you chose to save.
- Hosting in Italy (OVH Milan), backup in the EU (Hetzner Finland/Germany).

Read the [Privacy Policy](pesadia.com/privacy), the [Terms of Service](pesadia.com/terms) and the [Cookie Policy](pesadia.com/cookies) for the details. You can export or delete your account at any time.

## Contact

For GDPR requests (access, rectification, erasure, portability): \`privacy@pesadia.com\`. The full procedure is documented in \`subject-rights-procedure.md\`, with a **30-day** response window.

For bugs, suggestions, code contributions: the public repository https://github.com/thecomputer07/pesa with issue tracker. For security reports: \`pesadia.com/.well-known/security.txt\`.

For general inquiries: lorenzo.dimare@icloud.com.

---

*Last updated: 2026-04-13 — v1.0.*
`;
