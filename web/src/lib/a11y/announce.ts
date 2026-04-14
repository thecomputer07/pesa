/**
 * Off-screen aria-live region for programmatic announcements.
 * A single region instance is reused across calls. Useful when a state
 * change happens outside a React tree (e.g. route navigation completion).
 */

let politeNode: HTMLElement | null = null;
let assertiveNode: HTMLElement | null = null;

function ensure(kind: "polite" | "assertive"): HTMLElement {
  const existing = kind === "polite" ? politeNode : assertiveNode;
  if (existing && document.body.contains(existing)) return existing;

  const el = document.createElement("div");
  el.setAttribute("role", "status");
  el.setAttribute("aria-live", kind);
  el.setAttribute("aria-atomic", "true");
  // visually hidden
  el.style.cssText = [
    "position:absolute",
    "width:1px",
    "height:1px",
    "padding:0",
    "margin:-1px",
    "overflow:hidden",
    "clip:rect(0,0,0,0)",
    "white-space:nowrap",
    "border:0",
  ].join(";");
  document.body.appendChild(el);
  if (kind === "polite") politeNode = el;
  else assertiveNode = el;
  return el;
}

export function announce(message: string, kind: "polite" | "assertive" = "polite"): void {
  if (typeof document === "undefined" || !message) return;
  const el = ensure(kind);
  // Clear first to force SR re-announcement even on identical text.
  el.textContent = "";
  window.setTimeout(() => {
    el.textContent = message;
  }, 50);
}
