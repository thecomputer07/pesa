import { deleteCustomFood } from "./api";
import { queryClient } from "./queryClient";
import { db } from "./db/schema";

/**
 * Deletes one of the user's custom foods and forgets every local copy, so it
 * disappears at once: the "my foods" list, search results, the food detail,
 * the offline catalogue and local favourites (the server cascades those).
 * Resolves with null on success, or a message to show.
 */
export async function removeCustomFood(id: number): Promise<string | null> {
  const err = await deleteCustomFood(id);
  if (err) return err;
  try {
    await db.foods.delete(id);
    await db.favorites.where("food_id").equals(id).delete();
  } catch {
    /* local cache unavailable: the server state is what matters */
  }
  queryClient.removeQueries({ queryKey: ["foods", id] });
  await Promise.all([
    queryClient.invalidateQueries({ queryKey: ["foods-mine"] }),
    queryClient.invalidateQueries({ queryKey: ["foods", "search"] }),
  ]);
  return null;
}
