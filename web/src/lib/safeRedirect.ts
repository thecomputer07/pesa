/**
 * Where to go after login: only a path on this site. "//host" and "/\host"
 * start with "/" but browsers read them as another site (open redirect).
 */
export function safeRedirect(from: string | null, fallback: string): string {
  if (!from || !from.startsWith("/") || from.startsWith("//") || from.startsWith("/\\")) return fallback;
  if (/[\u0000-\u001f]/.test(from)) return fallback;
  return from;
}
