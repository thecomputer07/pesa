/**
 * Pure state machine driving the camera-permission UX on /scan.
 *
 * The flow is intentionally explicit so users understand that the camera is
 * only engaged after a user gesture (required on iOS Safari standalone).
 */
export type PermissionState =
  | "idle"
  | "prompting"
  | "granted"
  | "denied"
  | "unavailable";

export type PermissionEvent =
  | { type: "REQUEST" }
  | { type: "GRANTED" }
  | { type: "DENIED" }
  | { type: "UNAVAILABLE" }
  | { type: "RESET" };

export function nextState(
  s: PermissionState,
  e: PermissionEvent,
): PermissionState {
  if (e.type === "RESET") return "idle";
  switch (s) {
    case "idle":
      if (e.type === "REQUEST") return "prompting";
      return s;
    case "prompting":
      if (e.type === "GRANTED") return "granted";
      if (e.type === "DENIED") return "denied";
      if (e.type === "UNAVAILABLE") return "unavailable";
      return s;
    case "granted":
    case "denied":
    case "unavailable":
      return s;
    default:
      return s;
  }
}
