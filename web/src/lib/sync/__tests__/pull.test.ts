import "fake-indexeddb/auto";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { db } from "../../db/schema";
import { pullDelta } from "../pull";

const originalFetch = globalThis.fetch;

afterEach(() => {
  globalThis.fetch = originalFetch;
});

beforeEach(async () => {
  await db.meals.clear();
  await db.meta.clear();
});

describe("pullDelta", () => {
  it("exits cleanly when the server returns zero rows", async () => {
    globalThis.fetch = vi.fn(async () =>
      new Response(JSON.stringify({ server_cursor: 0 }), { status: 200 }),
    ) as unknown as typeof fetch;
    await pullDelta();
    expect(await db.meals.count()).toBe(0);
  });

  it("applies meals to Dexie and advances the cursor", async () => {
    let call = 0;
    globalThis.fetch = vi.fn(async () => {
      call++;
      if (call === 1) {
        return new Response(
          JSON.stringify({
            meals: [
              {
                client_id: "m1",
                occurred_at: "2026-04-13T09:30:00Z",
                slot: "colazione",
                updated_at_ms: 100,
                deleted_at_ms: null,
              },
            ],
            server_cursor: 500,
          }),
          { status: 200 },
        );
      }
      // Subsequent page: no more rows.
      return new Response(JSON.stringify({ server_cursor: 500 }), { status: 200 });
    }) as unknown as typeof fetch;

    await pullDelta();
    expect(await db.meals.count()).toBe(1);
    expect((await db.meta.get("sync_cursor"))?.value).toBe("500");
  });

  // TODO: pagination loop — returns 2 pages, cursor advances 500 → 900 → stop
  // TODO: single-flight — two concurrent pullDelta() calls issue the same inflight promise
  // TODO: non-ok responses short-circuit without throwing
});
