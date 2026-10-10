# Cookie Policy — Pesa

**Version:** 1.0
**Last updated:** 2026-04-13
**Authoritative language:** Italian

> *This is a convenience translation. In case of conflict with the Italian version (`cookie-policy.it.md`), the Italian version prevails.*

Reference compliance: **Italian Data Protection Authority (Garante), decision no. 231 of 10 June 2021 — "Guidelines on cookies and other tracking tools"**, in addition to Regulation (EU) 2016/679 (GDPR) and Directive 2002/58/EC ("ePrivacy") as implemented in Italy by Legislative Decree 196/2003 and Legislative Decree 101/2018.

## 1. Preamble

A **cookie** is a typically small text string that a website sends to the user's browser, which the browser stores and later sends back to the site on subsequent visits. By extension, "cookie" includes — under the Garante 2021 guidelines — **analogous tracking tools**: identifiers stored in `localStorage` / `sessionStorage` / IndexedDB, browser/device fingerprints, pixel tags, and equivalent technologies used to identify the user or the device.

Pesa uses **only technical cookies** strictly necessary for delivering the service or ensuring its security. Under **Garante 2021 §4.1** and Art. 122(1) Legislative Decree 196/2003, for technical cookies **the user's consent is not required**: this **notice** is sufficient. No consent banner is therefore shown (it would be misleading and superfluous).

Pesa **does not use**: profiling cookies, marketing cookies, advertising cookies, social-plugin cookies, third-party cookies, tracking pixels, fingerprinting, social login, CDNs that store user data. **Zero third-party trackers. Zero banner.**

## 2. List of cookies used

| Name | Type | Purpose | Duration | Attributes | Legal basis |
|------|------|---------|----------|------------|-------------|
| `pesa_session` | **Technical — session** | Maintain the user's authenticated session across HTTP requests. Contains an **opaque** identifier associated server-side with an `scs` store on PostgreSQL; no personal data is stored in the cookie itself. | 30-day rolling (renewed on each authenticated request); immediate revocation at logout | `HttpOnly`, `Secure`, `SameSite=Lax`, `Path=/` | **Art. 6(1)(b) GDPR** — performance of a contract; Garante 2021 §4.1 — technical cookie, no consent |
| `pesa_csrf` | **Technical — security** | Protection against Cross-Site Request Forgery via double-submit token pattern (cookie value must match `X-CSRF-Token` header sent by the frontend on mutating requests) | Session duration | `Secure`, `SameSite=Lax`, `Path=/`, **NOT** `HttpOnly` (must be readable by the frontend) | **Art. 6(1)(f) GDPR** — legitimate interest in security |

**No other cookie is set by the service.** No persistent profiling cookie, no third-party cookie, no advertising tracker, no first-party or third-party identifier for marketing purposes.

### Additional local storage (non-cookie)

For strictly technical purposes, the app may use:

- **IndexedDB / Dexie** to cache the **food dataset** locally (offline-first). Contains no personal identifier; it can be cleared by the user via browser settings or via the "clear cache" feature of the app when available.
- **Service Worker cache** (Workbox 7) for the PWA shell and static assets. No personal data.

These storages are not "cookies" in the strict sense of Art. 122 of the Privacy Code but — consistent with Garante 2021 §2 ("other tracking tools") — are disclosed here for transparency. Being **strictly necessary for delivering** the offline-first service, they fall under the same consent exemption as technical cookies.

## 3. Self-hosted analytics (Plausible Community Edition)

Where activated (INFRA-14), the service uses **Plausible Community Edition**, installed on the Provider's own infrastructure (Docker container alongside the Pesa binary and the PostgreSQL database). Plausible CE is a first-party analytics tool with the following properties:

- **Does not use cookies** — no identifier is stored in the user's browser;
- **Does not use localStorage / sessionStorage / IndexedDB** for analytics;
- **Does not fingerprint** device or browser;
- **Does not store the IP address** in clear (it uses a rotated daily hash to count unique visitors within the daily window, without persisting the identifier);
- **Does not cross-site track** (first-party only);
- **Aggregates** only: pageview count, generic referrer, device type (desktop / mobile / tablet), browser family, country at aggregate level.

**No user identifier.** **No profiling.** **No linking across visits.**

**Legal basis:** Art. 6(1)(f) GDPR — legitimate interest of the Provider in aggregate knowledge of service usage. Compliance with **Recital 26 GDPR** (fully anonymous data falls outside the GDPR scope) and **Garante 2021 §4.2** (first-party anonymized analytics that do not make the visitor identifiable are treated as technical cookies and **do not require consent**).

The user can however **object** to future counting of their visits by writing to `privacy@pesadia.com`: from that point onward, Plausible will be configured server-side to exclude interactions from the indicated IP/environment (technical note: since historic aggregates contain no identifiers, it is not possible to retroactively remove specific interactions from aggregates because they are already anonymous and non-isolable).

Plausible dashboard: access reserved exclusively to the Provider via authentication; no third party has access.

## 4. How to disable cookies

The user can at any time disable or clear cookies via their browser settings. Links to official guides:

- **Google Chrome:** https://support.google.com/chrome/answer/95647
- **Mozilla Firefox:** https://support.mozilla.org/en-US/kb/enable-and-disable-cookies-website-preferences
- **Safari (macOS):** https://support.apple.com/guide/safari/sfri11471/mac
- **Safari (iOS):** https://support.apple.com/en-us/HT201265
- **Microsoft Edge:** https://support.microsoft.com/en-us/microsoft-edge/delete-cookies-in-microsoft-edge-63947406-40ac-c3b8-57b9-2a946a29ae09

**Warning:** by disabling the technical cookies `pesa_session` and `pesa_csrf`, authentication will not work and mutating requests will be blocked by CSRF protection; the service remains usable in **demo mode** (food search + carb computation without account and without synced diary), which requires no authenticated session.

## 5. Changes

Any substantive change to the list of cookies (addition of a new cookie, change of purpose or duration) will be published on this page with a **version bump** and, where it involves non-technical cookies, with an in-app notice and a consent prompt. **Introducing any profiling cookie** would require a consent banner under Garante 2021 §2-§3 — such introduction is **not planned** in the MVP and would be an explicit decision of the Provider documented in `PROJECT.md` Key Decisions.

Version history:

- **v1.0** — 2026-04-13 — first publication (Phase 0), exclusively technical cookies `pesa_session` and `pesa_csrf`.

## 6. Contacts

- **Controller:** Lorenzo Di Mare
- **Privacy requests email:** `privacy@pesadia.com`
- **General email:** lorenzo.dimare@icloud.com
- **Certified email (PEC):** —
- **Supervisory authority:** Garante per la protezione dei dati personali — Piazza Venezia 11, 00187 Rome, Italy — `garante@gpdp.it` — `protocollo@pec.gpdp.it` — https://www.garanteprivacy.it
