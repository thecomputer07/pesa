import { describe, it, expect } from "vitest";
import { nextState } from "./permissionMachine";

describe("permissionMachine.nextState", () => {
  it("idle + REQUEST -> prompting", () => {
    expect(nextState("idle", { type: "REQUEST" })).toBe("prompting");
  });
  it("prompting + GRANTED -> granted", () => {
    expect(nextState("prompting", { type: "GRANTED" })).toBe("granted");
  });
  it("prompting + DENIED -> denied", () => {
    expect(nextState("prompting", { type: "DENIED" })).toBe("denied");
  });
  it("prompting + UNAVAILABLE -> unavailable", () => {
    expect(nextState("prompting", { type: "UNAVAILABLE" })).toBe(
      "unavailable",
    );
  });
  it("any + RESET -> idle", () => {
    expect(nextState("granted", { type: "RESET" })).toBe("idle");
    expect(nextState("denied", { type: "RESET" })).toBe("idle");
    expect(nextState("unavailable", { type: "RESET" })).toBe("idle");
  });
  it("illegal transitions return the same state", () => {
    expect(nextState("idle", { type: "GRANTED" })).toBe("idle");
    expect(nextState("granted", { type: "REQUEST" })).toBe("granted");
  });
});
