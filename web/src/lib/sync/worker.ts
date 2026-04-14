import { db } from "../db/schema";
import { drain, markFailed, readyToFlush } from "./outbox";
import type {
  BatchPayload,
  BatchResponse,
  OutboxKind,
  OutboxRow,
} from "./types";

// Module-level single-flight guard.
let inflight: Promise<void> | null = null;

/** Drain the outbox by POSTing grouped batches to /api/v1/sync. Single-flight. */
export function flushOutbox(): Promise<void> {
  if (inflight) return inflight;
  inflight = doFlush().finally(() => {
    inflight = null;
  });
  return inflight;
}

async function doFlush(): Promise<void> {
  const rows = await readyToFlush(50);
  if (rows.length === 0) return;

  const batch: BatchPayload = {};
  const byKey = new Map<string, OutboxRow>();
  for (const row of rows) {
    const bucket = pluralOf(row.kind);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (batch as any)[bucket] ||= [];
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (batch as any)[bucket].push(JSON.parse(row.payload_json));
    byKey.set(`${row.kind}:${row.client_id}`, row);
  }

  let resp: Response;
  try {
    resp = await fetch("/api/v1/sync", {
      method: "POST",
      credentials: "same-origin",
      headers: {
        "Content-Type": "application/json",
        "X-CSRF-Token": readCsrfToken(),
      },
      body: JSON.stringify(batch),
    });
  } catch (e) {
    // Network failure — every row gets a fail bump.
    await Promise.all(rows.map((r) => markFailed(r.id!, "network_error")));
    return;
  }

  if (resp.status === 401) {
    window.dispatchEvent(new CustomEvent("pesa:auth-lost"));
    return;
  }
  if (resp.status === 413) {
    // Block rows so they don't keep retrying a too-big payload.
    await Promise.all(
      rows.map((r) =>
        db.outbox.update(r.id!, { blocked: true, last_error: "payload_too_large" }),
      ),
    );
    return;
  }
  if (!resp.ok) {
    await Promise.all(rows.map((r) => markFailed(r.id!, `http_${resp.status}`)));
    return;
  }

  const body = (await resp.json()) as BatchResponse;
  const toDrain: number[] = [];

  for (const result of body.results) {
    const row = byKey.get(`${result.kind}:${result.client_id}`);
    if (!row) continue;
    if (result.status === "applied" || result.status === "conflict_skipped") {
      toDrain.push(row.id!);
      if (result.status === "conflict_skipped") {
        window.dispatchEvent(
          new CustomEvent("pesa:conflict", {
            detail: { kind: result.kind, reason: result.reason },
          }),
        );
      }
    } else {
      await markFailed(row.id!, result.reason ?? "error");
    }
  }
  if (toDrain.length) await drain(toDrain);

  if (body.server_cursor > 0) {
    const prev = await db.meta.get("sync_cursor");
    const prevVal = Number(prev?.value ?? 0);
    if (body.server_cursor > prevVal) {
      await db.meta.put({ key: "sync_cursor", value: String(body.server_cursor) });
    }
  }
}

function pluralOf(kind: OutboxKind): keyof BatchPayload {
  switch (kind) {
    case "meal":
      return "meals";
    case "meal_item":
      return "meal_items";
    case "recipe":
      return "recipes";
    case "recipe_item":
      return "recipe_items";
    case "meal_template":
      return "meal_templates";
    case "favorite":
      return "favorites";
  }
}

function readCsrfToken(): string {
  if (typeof document === "undefined") return "";
  // Pattern established in Phase 1: CSRF token mirrored in a non-HttpOnly cookie.
  const match = /(?:^|;\s*)pesa_csrf=([^;]+)/.exec(document.cookie);
  return match ? decodeURIComponent(match[1]) : "";
}
