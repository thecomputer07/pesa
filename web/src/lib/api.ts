import { z } from "zod";

/** Base URL. Proxied via Vite in dev; same-origin in prod. */
export const API_BASE = "/api/v1";

export class ApiError extends Error {
  status: number;
  type?: string;
  detail?: string;
  constructor(status: number, message: string, type?: string, detail?: string) {
    super(message);
    this.status = status;
    this.type = type;
    this.detail = detail;
  }
}

function readCookie(name: string): string | undefined {
  if (typeof document === "undefined") return undefined;
  const prefix = `${name}=`;
  for (const c of document.cookie.split("; ")) {
    if (c.startsWith(prefix)) return decodeURIComponent(c.slice(prefix.length));
  }
  return undefined;
}

/**
 * Thin fetch wrapper that:
 *  - prepends {@link API_BASE}
 *  - always sends session cookies (`credentials: 'include'`)
 *  - attaches the `X-CSRF-Token` header on state-changing requests, read from
 *    the non-HttpOnly `pesa_csrf` cookie (double-submit pattern).
 * Does NOT throw on non-2xx — callers handle status themselves.
 */
export function apiFetch(path: string, init: RequestInit = {}): Promise<Response> {
  const method = (init.method ?? "GET").toUpperCase();
  const headers = new Headers(init.headers ?? {});
  if (!headers.has("Accept")) headers.set("Accept", "application/json");
  if (method !== "GET" && method !== "HEAD" && method !== "OPTIONS") {
    const csrf = readCookie("pesa_csrf");
    if (csrf && !headers.has("X-CSRF-Token")) headers.set("X-CSRF-Token", csrf);
  }
  return fetch(`${API_BASE}${path}`, {
    ...init,
    method,
    headers,
    credentials: "include",
  });
}

async function request<T>(path: string, schema: z.ZodType<T>, init?: RequestInit): Promise<T> {
  const res = await apiFetch(path, init);

  if (!res.ok) {
    let type: string | undefined;
    let detail: string | undefined;
    try {
      const body = await res.json();
      type = body.type ?? body.title;
      detail = body.detail ?? body.message;
    } catch {
      /* ignore */
    }
    throw new ApiError(res.status, `HTTP ${res.status}`, type, detail);
  }

  const json = await res.json();
  return schema.parse(json);
}

// ─── Schemas (mirror Go DTOs in internal/http/handlers/foods.go) ───

export const FoodSourceSchema = z.object({
  code: z.string(),
  license: z.string(),
  license_url: z.string().url().optional(),
});
export type FoodSource = z.infer<typeof FoodSourceSchema>;

export const FoodSearchResultSchema = z.object({
  id: z.number().int().positive(),
  name_it: z.string(),
  name_en: z.string().optional(),
  category: z.string().optional(),
  state: z.string(),
  kcal: z.number().optional(),
  carb_g: z.number().optional(),
  protein_g: z.number().optional(),
  fat_g: z.number().optional(),
  sugar_g: z.number().optional(),
  fiber_g: z.number().optional(),
  thumb_url: z.string().optional(),
  source: FoodSourceSchema,
});
export type FoodSearchResult = z.infer<typeof FoodSearchResultSchema>;

export const FoodSearchResponseSchema = z.object({
  results: z.array(FoodSearchResultSchema),
  total: z.number().int().nonnegative(),
  page: z.number().int().positive(),
  limit: z.number().int().positive(),
});
export type FoodSearchResponse = z.infer<typeof FoodSearchResponseSchema>;

export const FoodImageSchema = z.object({
  url: z.string(),
  attribution: z.string().optional(),
  license: z.string().optional(),
  is_external: z.boolean(),
});

export const FoodDetailSchema = z.object({
  id: z.number().int().positive(),
  name_it: z.string(),
  name_en: z.string().optional(),
  category: z.string().optional(),
  state: z.string(),
  edible_pct: z.number(),
  kcal: z.number().optional(),
  protein_g: z.number().optional(),
  fat_g: z.number().optional(),
  carb_g: z.number().optional(),
  sugar_g: z.number().optional(),
  fiber_g: z.number().optional(),
  water_g: z.number().optional(),
  ash_g: z.number().optional(),
  alcohol_g: z.number().optional(),
  salt_g: z.number().optional(),
  attribution: z.string(),
  license: z.string(),
  license_url: z.string().optional(),
  source_url: z.string().optional(),
  last_imported_at: z.string().optional(),
  images: z.array(FoodImageSchema).default([]),
  thumb_url: z.string().optional(),
  siblings: z
    .array(
      z.object({
        id: z.number().int().positive(),
        name_it: z.string(),
        state: z.string(),
      }),
    )
    .default([]),
});
export type FoodDetail = z.infer<typeof FoodDetailSchema>;

// ─── API calls ───

/** Search source filter: each value returns rows of that source only. */
export type SearchSource = "crea" | "ciqual" | "off" | "mine";

export function searchFoods(
  q: string,
  opts: { limit?: number; page?: number; source?: SearchSource; signal?: AbortSignal } = {},
): Promise<FoodSearchResponse> {
  const params = new URLSearchParams({ q });
  if (opts.source) params.set("source", opts.source);
  if (opts.limit) params.set("limit", String(opts.limit));
  if (opts.page) params.set("page", String(opts.page));
  return request(`/foods/search?${params.toString()}`, FoodSearchResponseSchema, {
    signal: opts.signal,
  });
}

export function getFood(id: number, opts: { signal?: AbortSignal } = {}): Promise<FoodDetail> {
  return request(`/foods/${id}`, FoodDetailSchema, { signal: opts.signal });
}
