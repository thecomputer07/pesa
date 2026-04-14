import type Dexie from "dexie";
import { db } from "../db/schema";
import type { OutboxKind, OutboxRow } from "./types";
import { registerBackgroundSync } from "./bg-sync-register";

// Backoff schedule: 5s, 30s, 2m, 10m, 1h (capped). Cap at 10 attempts OR 24h age.
const BACKOFF_MS = [5_000, 30_000, 120_000, 600_000, 3_600_000];
const MAX_ATTEMPTS = 10;
const MAX_AGE_MS = 24 * 60 * 60 * 1000;

/**
 * Enqueue an outbox row. If called inside an existing Dexie transaction `tx`
 * that includes the `outbox` table, the enqueue is atomic with the caller's
 * entity write (both commit or both roll back). Otherwise opens its own tx.
 */
export async function enqueue(
  kind: OutboxKind,
  clientId: string,
  payload: unknown,
  tx?: Dexie.Transaction,
): Promise<void> {
  const row: OutboxRow = {
    kind,
    client_id: clientId,
    payload_json: JSON.stringify(payload),
    created_at: Date.now(),
    next_attempt_at: 0,
    attempts: 0,
  };
  if (tx) {
    await tx.table<OutboxRow, number>("outbox").add(row);
  } else {
    await db.outbox.add(row);
  }
  // Best-effort: ask the SW to schedule a Background Sync flush on Chromium.
  // Fire-and-forget; no await — must not block the write path.
  void registerBackgroundSync();
}

/** Fetch outbox rows whose `next_attempt_at` has elapsed, oldest first. */
export async function readyToFlush(limit = 50): Promise<OutboxRow[]> {
  const now = Date.now();
  return db.outbox
    .where("next_attempt_at")
    .belowOrEqual(now)
    .and((r) => !r.blocked)
    .limit(limit)
    .toArray();
}

/** Bulk-delete applied/conflict_skipped rows. */
export async function drain(ids: number[]): Promise<void> {
  if (ids.length === 0) return;
  await db.outbox.bulkDelete(ids);
}

/**
 * Increment attempts + schedule next retry per backoff curve.
 * Once `attempts >= 10` OR age exceeds 24h the row flips to `blocked=true`
 * (surfaces the "Sync bloccato" banner per CONTEXT retry policy).
 */
export async function markFailed(id: number, err: string): Promise<void> {
  const row = await db.outbox.get(id);
  if (!row) return;
  const attempts = row.attempts + 1;
  const age = Date.now() - row.created_at;
  const blocked = attempts >= MAX_ATTEMPTS || age > MAX_AGE_MS;
  const backoffIdx = Math.min(attempts - 1, BACKOFF_MS.length - 1);
  const backoff = BACKOFF_MS[Math.max(0, backoffIdx)];
  // last_error intentionally stores only a short reason code, never a response body
  // (T-04.03-03 information-disclosure mitigation).
  const safeErr = err.slice(0, 64);
  await db.outbox.update(id, {
    attempts,
    last_error: safeErr,
    next_attempt_at: Date.now() + backoff,
    blocked,
  });
}

/** Convenience for tests. */
export const _internals = { BACKOFF_MS, MAX_ATTEMPTS, MAX_AGE_MS };
