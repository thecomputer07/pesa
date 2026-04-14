import "fake-indexeddb/auto";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { db } from "../../db/schema";
import { flushOutbox } from "../worker";

// Lightweight fetch stub — avoids pulling msw into this PR's devDeps path
// (msw is declared in package.json for future integration). TODO(msw): migrate
// these cases to msw + request/response handlers for richer assertions.
const originalFetch = globalThis.fetch;

afterEach(() => {
  globalThis.fetch = originalFetch;
});

beforeEach(async () => {
  await db.outbox.clear();
  await db.meta.clear();
  Object.defineProperty(globalThis, "document", {
    configurable: true,
    value: { cookie: "pesa_csrf=tok" },
  });
});

describe("flushOutbox", () => {
  it("no-ops when the outbox is empty (no fetch call)", async () => {
    const spy = vi.fn();
    globalThis.fetch = spy as unknown as typeof fetch;
    await flushOutbox();
    expect(spy).not.toHaveBeenCalled();
  });

  it("drains rows on status:applied and advances sync_cursor", async () => {
    await db.outbox.add({
      kind: "meal",
      client_id: "aaa",
      payload_json: JSON.stringify({ client_id: "aaa" }),
      created_at: Date.now(),
      next_attempt_at: 0,
      attempts: 0,
    });
    globalThis.fetch = vi.fn(async () =>
      new Response(
        JSON.stringify({
          results: [{ kind: "meal", client_id: "aaa", status: "applied", server_id: 1 }],
          server_cursor: 1234,
        }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      ),
    ) as unknown as typeof fetch;

    await flushOutbox();
    expect(await db.outbox.count()).toBe(0);
    expect((await db.meta.get("sync_cursor"))?.value).toBe("1234");
  });

  it("drains conflict_skipped rows AND dispatches pesa:conflict on window", async () => {
    await db.outbox.add({
      kind: "meal",
      client_id: "bbb",
      payload_json: JSON.stringify({ client_id: "bbb" }),
      created_at: Date.now(),
      next_attempt_at: 0,
      attempts: 0,
    });
    globalThis.fetch = vi.fn(async () =>
      new Response(
        JSON.stringify({
          results: [{ kind: "meal", client_id: "bbb", status: "conflict_skipped", reason: "older_timestamp" }],
          server_cursor: 0,
        }),
        { status: 200 },
      ),
    ) as unknown as typeof fetch;

    const evts: Event[] = [];
    const handler = (e: Event) => evts.push(e);
    window.addEventListener("pesa:conflict", handler);
    await flushOutbox();
    window.removeEventListener("pesa:conflict", handler);

    expect(await db.outbox.count()).toBe(0);
    expect(evts).toHaveLength(1);
  });

  // TODO: 401 → dispatches pesa:auth-lost and does not drain
  // TODO: 413 → rows marked blocked=true
  // TODO: 500 → markFailed increments attempts
  // TODO: single-flight coalescing (two concurrent calls issue one fetch)
});
