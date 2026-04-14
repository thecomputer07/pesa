import "fake-indexeddb/auto";
import { beforeEach, describe, expect, it } from "vitest";
import { db } from "../../db/schema";
import { toggleFavorite } from "../favorites";

beforeEach(async () => {
  await db.favorites.clear();
  await db.outbox.clear();
});

describe("toggleFavorite", () => {
  it("first call creates a favorite and enqueues outbox", async () => {
    const on = await toggleFavorite(42, 1);
    expect(on).toBe(true);
    const rows = await db.favorites.toArray();
    expect(rows).toHaveLength(1);
    expect(rows[0].deleted_at_ms).toBeFalsy();
    expect(await db.outbox.count()).toBe(1);
  });

  it("second call soft-deletes the favorite (keeps same client_id)", async () => {
    await toggleFavorite(42, 1);
    const first = await db.favorites.toArray();
    const on = await toggleFavorite(42, 1);
    expect(on).toBe(false);
    const rows = await db.favorites.toArray();
    expect(rows[0].client_id).toBe(first[0].client_id);
    expect(rows[0].deleted_at_ms).toBeGreaterThan(0);
  });

  it("third call resurrects with the same client_id", async () => {
    await toggleFavorite(42, 1);
    await toggleFavorite(42, 1);
    const before = await db.favorites.toArray();
    const on = await toggleFavorite(42, 1);
    expect(on).toBe(true);
    const after = await db.favorites.toArray();
    expect(after[0].client_id).toBe(before[0].client_id);
    expect(after[0].deleted_at_ms).toBeNull();
  });

  // TODO: useIsFavorited reactive boolean flips across toggle cycles
  // TODO: multiple foods coexist (separate client_ids, separate outbox rows)
});
