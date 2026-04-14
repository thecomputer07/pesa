import type { ReactNode } from "react";
import { Navigate, useLocation } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useAuth } from "./AuthContext";

export interface RequireAuthProps {
  children: ReactNode;
}

/**
 * Gate wrapper: redirects unauthenticated visitors to /accedi, preserving
 * the attempted path as `?from=...` so the login page can bounce them back
 * after success.
 */
export default function RequireAuth({ children }: RequireAuthProps) {
  const { user, loading } = useAuth();
  const location = useLocation();
  const { t } = useTranslation();

  if (loading) {
    return (
      <div role="status" aria-live="polite" className="py-12 text-center text-sm opacity-70">
        {t("a11y.loading")}
      </div>
    );
  }

  if (!user) {
    const from = encodeURIComponent(location.pathname + location.search);
    return <Navigate to={`/accedi?from=${from}`} replace />;
  }

  return <>{children}</>;
}
