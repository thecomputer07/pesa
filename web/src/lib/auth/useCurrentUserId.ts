import { useAuth } from "./AuthContext";

/**
 * Returns the current authenticated user's numeric id, or null if not logged
 * in. Callers inside <RequireAuth> branches can safely treat `null` as a
 * transient boundary (parent gate will redirect before render settles).
 *
 * Kept as a tiny convenience hook so existing call-sites (DiaryPage,
 * RecipeEditorPage, FavoritesPage, FavoriteToggle) don't
 * each need to destructure `useAuth()` for a single field.
 */
export function useCurrentUserId(): number | null {
  // Vitest harness hook: tests set window.__pesaTestUserId to bypass auth.
  const testId = (globalThis as unknown as { __pesaTestUserId?: number }).__pesaTestUserId;
  if (typeof testId === "number") return testId;
  const { user } = useAuth();
  return user?.id ?? null;
}
