/**
 * normalizeEAN trims and validates a barcode string, zero-padding up to 13 digits.
 *
 * Rules:
 *   - whitespace is trimmed
 *   - must contain only digits after trim
 *   - length 8–13 → left-pad with zeros to 13
 *   - length 14 → only accepted when the leading digit is "0" (GTIN-14 → EAN-13)
 *   - anything else → null
 */
export function normalizeEAN(raw: string): string | null {
  if (typeof raw !== "string") return null;
  const s = raw.trim();
  if (!/^\d+$/.test(s)) return null;
  if (s.length === 14) {
    if (s[0] !== "0") return null;
    return s.slice(1);
  }
  if (s.length >= 8 && s.length <= 13) {
    return s.padStart(13, "0");
  }
  return null;
}
