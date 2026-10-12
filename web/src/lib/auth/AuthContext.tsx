import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { ApiError, apiFetch } from "../api";
import { clearMyFoods, syncMyFoods } from "../catalog/sync";

export interface AuthUser {
  id: number;
  email: string;
  email_verified: boolean;
}

export interface AuthContextShape {
  user: AuthUser | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, localeConsent: boolean) => Promise<void>;
  logout: () => Promise<void>;
  refresh: () => Promise<void>;
}

const AuthContext = createContext<AuthContextShape | undefined>(undefined);

const LS_KEY = "pesa.auth.user";

function loadCachedUser(): AuthUser | null {
  try {
    const raw = localStorage.getItem(LS_KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw);
    if (
      parsed &&
      typeof parsed.id === "number" &&
      typeof parsed.email === "string" &&
      typeof parsed.email_verified === "boolean"
    ) {
      return parsed as AuthUser;
    }
  } catch {
    /* ignore */
  }
  return null;
}

function saveCachedUser(user: AuthUser | null) {
  try {
    if (user) localStorage.setItem(LS_KEY, JSON.stringify(user));
    else localStorage.removeItem(LS_KEY);
  } catch {
    /* ignore */
  }
}

function hasCsrfCookie(): boolean {
  if (typeof document === "undefined") return false;
  return document.cookie.split("; ").some((c) => c.startsWith("pesa_csrf="));
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [loading, setLoading] = useState(true);
  const mountedRef = useRef(true);

  const refresh = useCallback(async () => {
    try {
      const res = await apiFetch("/me", { method: "GET" });
      if (res.ok) {
        const data = (await res.json()) as AuthUser;
        if (mountedRef.current) {
          setUser(data);
          saveCachedUser(data);
        }
        return;
      }
      if (res.status === 401 || res.status === 403) {
        if (mountedRef.current) {
          setUser(null);
          saveCachedUser(null);
        }
        return;
      }
      // Endpoint not yet available (404/5xx): fall back to cached user iff
      // the CSRF cookie — set at login, cleared at logout — is still present.
      const cached = loadCachedUser();
      if (cached && hasCsrfCookie()) {
        if (mountedRef.current) setUser(cached);
      } else {
        if (mountedRef.current) {
          setUser(null);
          saveCachedUser(null);
        }
      }
    } catch {
      const cached = loadCachedUser();
      if (cached && hasCsrfCookie() && mountedRef.current) {
        setUser(cached);
      } else if (mountedRef.current) {
        setUser(null);
      }
    }
  }, []);

  useEffect(() => {
    mountedRef.current = true;
    void refresh().finally(() => {
      if (mountedRef.current) setLoading(false);
    });
    return () => {
      mountedRef.current = false;
    };
  }, [refresh]);

  // The user's own foods live on the device only while they are signed in.
  const userId = user?.id;
  useEffect(() => {
    if (loading) return;
    if (userId) void syncMyFoods(userId);
    else void clearMyFoods();
  }, [userId, loading]);

  const login = useCallback(
    async (email: string, password: string) => {
      const res = await apiFetch("/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });
      if (!res.ok) {
        let detail: string | undefined;
        let type: string | undefined;
        try {
          const body = await res.json();
          detail = body.detail ?? body.message;
          type = body.type;
        } catch {
          /* ignore */
        }
        throw new ApiError(res.status, `HTTP ${res.status}`, type, detail);
      }
      await refresh();
    },
    [refresh],
  );

  const register = useCallback(
    async (email: string, password: string, localeConsent: boolean) => {
      if (!localeConsent) {
        throw new ApiError(422, "privacy_consent_required", "privacy_consent_required");
      }
      const res = await apiFetch("/auth/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          email,
          password,
          privacy_consent: true,
          locale: "it",
        }),
      });
      if (!res.ok) {
        let detail: string | undefined;
        let type: string | undefined;
        try {
          const body = await res.json();
          detail = body.detail ?? body.message;
          type = body.type;
        } catch {
          /* ignore */
        }
        throw new ApiError(res.status, `HTTP ${res.status}`, type, detail);
      }
    },
    [],
  );

  const logout = useCallback(async () => {
    try {
      await apiFetch("/auth/logout", { method: "POST" });
    } catch {
      /* ignore — clear client state anyway */
    }
    if (mountedRef.current) {
      setUser(null);
      saveCachedUser(null);
    }
  }, []);

  const value = useMemo<AuthContextShape>(
    () => ({ user, loading, login, register, logout, refresh }),
    [user, loading, login, register, logout, refresh],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextShape {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used inside <AuthProvider>");
  return ctx;
}
