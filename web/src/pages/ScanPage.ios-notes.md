# iOS (Safari / standalone PWA) — Scanner caveats

This annex documents the platform-specific quirks that shape `/scan`.

## Permission gesture requirement

Safari — including the standalone PWA — requires a user gesture before
`navigator.mediaDevices.getUserMedia({ video: true })` can prompt. The
permission UI on `/scan` therefore always begins in the `idle` state and
never auto-requests.

## First-denial recovery

Once the user denies camera permission, iOS Safari remembers the choice and
will reject subsequent calls without re-prompting. Recovery path:

1. Open **Settings → Safari → Camera** (or **Settings → Pesa** when running as
   a home-screen PWA on iOS 16.4+).
2. Change the entry to **Ask** or **Allow**.
3. Return to Pesa and tap **Riprova**.

## Torch

`MediaTrackCapabilities.torch` is not exposed on iOS Safari below iOS 18. The
component hides the torch button when the capability is missing rather than
surfacing a control that does nothing.

## Background lifecycle

When the app is backgrounded, the OS suspends the page. The scanner stops
camera tracks on `visibilitychange=hidden` as a belt-and-braces measure so the
recording indicator never persists after the user leaves the page.

## Standalone vs. in-browser

The standalone PWA runs in its own WebView; the camera permission is scoped to
the PWA origin, separate from Safari's. Users who granted permission in Safari
will still see the prompt on first use inside the installed PWA.
