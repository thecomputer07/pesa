import { describe, expect, it } from "vitest";
import { safeRedirect } from "../safeRedirect";

describe("safeRedirect", () => {
  it("keeps paths on this site", () => {
    expect(safeRedirect("/ricette?x=1", "/diario")).toBe("/ricette?x=1");
  });
  it("refuses other sites", () => {
    for (const bad of ["https://evil.com", "//evil.com", "/\\evil.com", "/\tevil", "evil", null]) {
      expect(safeRedirect(bad, "/diario")).toBe("/diario");
    }
  });
});
