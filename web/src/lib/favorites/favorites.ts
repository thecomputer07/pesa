import { useLiveQuery } from "dexie-react-hooks";
import { db } from "../db/schema";
import { enqueue } from "../sync/outbox";
import type { FavoriteDTO, LocalFavorite } from "../sync/types";

/**
 * Toggle the favorite state for (userId, foodId). Atomic Dexie write + outbox
 * enqueue. Returns the NEW favorited state (true = now favorited).
 *
 * Soft-delete + resurrect use the SAME client_id so the server's
 * (user_id, client_id) upsert updates in place (clearing deleted_at_ms),
 * avoiding the (user_id, food_id) UNIQUE collision path entirely on happy-path.
 */
export async function toggleFavorite(foodId: number, userId: number): Promise<boolean> {
  const now = Date.now();
  return db.transaction("rw", db.favorites, db.outbox, async (tx) => {
    const existing = await tx
      .table<LocalFavorite, string>("favorites")
      .where("[user_id+food_id]")
      .equals([userId, foodId])
      .first();
    if (existing) {
      const nowDeleted = !existing.deleted_at_ms; // toggling
      const next: LocalFavorite = {
        ...existing,
        deleted_at_ms: nowDeleted ? now : null,
        updated_at_ms: now,
      };
      await tx.table("favorites").put(next);
      const dto: FavoriteDTO = {
        client_id: existing.client_id,
        updated_at_ms: now,
        deleted_at_ms: next.deleted_at_ms,
        food_id: foodId,
      };
      await enqueue("favorite", existing.client_id, dto, tx);
      return !nowDeleted;
    }
    const clientId = crypto.randomUUID();
    const row: LocalFavorite = {
      client_id: clientId,
      user_id: userId,
      food_id: foodId,
      updated_at_ms: now,
      deleted_at_ms: null,
    };
    await tx.table("favorites").put(row);
    const dto: FavoriteDTO = {
      client_id: clientId,
      updated_at_ms: now,
      deleted_at_ms: null,
      food_id: foodId,
    };
    await enqueue("favorite", clientId, dto, tx);
    return true;
  });
}

export function useIsFavorited(foodId: number, userId: number): boolean | undefined {
  return useLiveQuery(async () => {
    const row = await db.favorites
      .where("[user_id+food_id]")
      .equals([userId, foodId])
      .first();
    return !!row && !row.deleted_at_ms;
  }, [foodId, userId]);
}

export function useFavorites(userId: number) {
  return useLiveQuery(async () => {
    const rows = await db.favorites
      .where("user_id")
      .equals(userId)
      .and((r) => !r.deleted_at_ms)
      .toArray();
    return Promise.all(
      rows.map(async (r) => ({ ...r, food: await db.foods.get(r.food_id) })),
    );
  }, [userId]);
}
