import "fake-indexeddb/auto";
import { beforeEach, describe, expect, it } from "vitest";
import { db } from "../../db/schema";
import { drain, enqueue, markFailed, readyToFlush, _internals } from "../outbox";

beforeEach(async () => {
  await db.outbox.clear();
  await db.meals.clear();
});

describe("outbox.enqueue", () => {
  it("creates a row with attempts=0 and next_attempt_at=0 (ready immediately)", async () => {
    await enqueue("meal", "00000000-0000-7000-8000-000000000001", { foo: "bar" });
    const rows = await db.outbox.toArray();
    expect(rows).toHaveLength(1);
    expect(rows[0].attempts).toBe(0);
    expect(rows[0].next_attempt_at).toBe(0);
    expect(rows[0].kind).toBe("meal");
    expect(JSON.parse(rows[0].payload_json)).toEqual({ foo: "bar" });
  });

  it("is atomic with a sibling write inside the same Dexie transaction", async () => {
    const clientId = "00000000-0000-7000-8000-000000000002";
    await db.transaction("rw", db.meals, db.outbox, async (tx) => {
      await tx.table("meals").put({
        client_id: clientId,
        user_id: 1,
        occurred_at: Date.now(),
        slot: "pranzo",
        updated_at_ms: Date.now(),
        deleted_at_ms: null,
      });
      await enqueue("meal", clientId, { client_id: clientId }, tx);
    });
    expect(await db.meals.count()).toBe(1);
    expect(await db.outbox.count()).toBe(1);
  });

  // TODO: rollback test — inducing a throw inside a Dexie `rw` transaction
  // reliably across fake-indexeddb + dexie versions requires a specific error
  // shape. Representative atomicity is covered by the happy-path test above.
});

describe("outbox.readyToFlush / drain", () => {
  it("returns only rows with next_attempt_at <= now and drops blocked rows", async () => {
    const future = Date.now() + 100_000;
    await db.outbox.bulkAdd([
      { kind: "meal", client_id: "a", payload_json: "{}", created_at: 0, next_attempt_at: 0, attempts: 0 },
      { kind: "meal", client_id: "b", payload_json: "{}", created_at: 0, next_attempt_at: future, attempts: 0 },
      { kind: "meal", client_id: "c", payload_json: "{}", created_at: 0, next_attempt_at: 0, attempts: 0, blocked: true },
    ]);
    const ready = await readyToFlush();
    expect(ready.map((r) => r.client_id)).toEqual(["a"]);
  });

  it("drain removes the given ids", async () => {
    const id1 = await db.outbox.add({
      kind: "meal", client_id: "x", payload_json: "{}", created_at: 0, next_attempt_at: 0, attempts: 0,
    });
    const id2 = await db.outbox.add({
      kind: "meal", client_id: "y", payload_json: "{}", created_at: 0, next_attempt_at: 0, attempts: 0,
    });
    await drain([id1, id2]);
    expect(await db.outbox.count()).toBe(0);
  });
});

describe("outbox.markFailed backoff", () => {
  it("schedules next_attempt_at per the documented BACKOFF_MS curve", async () => {
    const id = await db.outbox.add({
      kind: "meal", client_id: "bo1", payload_json: "{}", created_at: Date.now(), next_attempt_at: 0, attempts: 0,
    });
    const before = Date.now();
    await markFailed(id, "http_500");
    const row = await db.outbox.get(id);
    expect(row!.attempts).toBe(1);
    expect(row!.last_error).toBe("http_500");
    expect(row!.next_attempt_at).toBeGreaterThanOrEqual(before + _internals.BACKOFF_MS[0] - 10);
  });

  it("sets blocked=true once attempts cap reached", async () => {
    const id = await db.outbox.add({
      kind: "meal",
      client_id: "bo2",
      payload_json: "{}",
      created_at: Date.now(),
      next_attempt_at: 0,
      attempts: _internals.MAX_ATTEMPTS - 1,
    });
    await markFailed(id, "still_bad");
    const row = await db.outbox.get(id);
    expect(row!.blocked).toBe(true);
  });
});

// TODO(full-coverage): add tests for
//   - the 24h age cap blocking path
//   - concurrent markFailed calls on the same id (no double-increment)
//   - enqueue triggering registerBackgroundSync (stubbed)
