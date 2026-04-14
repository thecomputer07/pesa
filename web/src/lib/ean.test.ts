import { describe, it, expect } from "vitest";
import { normalizeEAN } from "./ean";

describe("normalizeEAN", () => {
  it("rejects too-short input", () => {
    expect(normalizeEAN("123")).toBeNull();
  });
  it("zero-pads 8-digit EAN to 13", () => {
    expect(normalizeEAN("12345678")).toBe("0000012345678");
  });
  it("returns 13-digit EAN unchanged", () => {
    expect(normalizeEAN("8001505005707")).toBe("8001505005707");
  });
  it("strips one leading zero from GTIN-14", () => {
    expect(normalizeEAN("08001505005707")).toBe("8001505005707");
  });
  it("rejects 14-digit input that is not 0-prefixed", () => {
    expect(normalizeEAN("18001505005707")).toBeNull();
  });
  it("trims surrounding whitespace", () => {
    expect(normalizeEAN(" 8001505005707 ")).toBe("8001505005707");
  });
  it("rejects non-digit characters", () => {
    expect(normalizeEAN("80015x5005707")).toBeNull();
  });
});
