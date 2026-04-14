import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import Footer from "./Footer";
import ScopeBanner from "./ScopeBanner";
import InstallPrompt from "./InstallPrompt";
import { useAuth } from "../lib/auth/AuthContext";

export default function Layout() {
  const { t } = useTranslation();
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  async function onLogout() {
    await logout();
    navigate("/", { replace: true });
  }

  const navClass = ({ isActive }: { isActive: boolean }) =>
    [
      "px-3 py-2 rounded-md text-sm font-medium",
      isActive
        ? "bg-[color:var(--color-brand-50)] text-[color:var(--color-brand-700)]"
        : "text-[color:var(--color-ink-muted)] hover:text-[color:var(--color-ink)]",
    ].join(" ");

  return (
    <div className="min-h-dvh flex flex-col">
      <a href="#main" className="skip-link">
        {t("a11y.skipToContent")}
      </a>

      <header
        role="banner"
        className="bg-[color:var(--color-surface)] border-b border-[color:var(--color-border)]"
      >
        <div className="max-w-5xl mx-auto px-4 py-3 flex items-center gap-6">
          <NavLink to="/" className="flex items-center gap-2 no-underline">
            <span
              aria-hidden="true"
              className="inline-block w-8 h-8 rounded-lg bg-[color:var(--color-brand-600)] text-white grid place-items-center font-bold"
            >
              P
            </span>
            <span className="font-semibold text-lg text-[color:var(--color-ink)]">Pesa</span>
          </NavLink>

          <nav aria-label={t("a11y.primaryNav")} className="flex items-center gap-1 ml-auto flex-wrap">
            <NavLink to="/" end className={navClass}>
              {t("nav.home")}
            </NavLink>
            <NavLink to="/cerca" className={navClass}>
              {t("nav.search")}
            </NavLink>
            <NavLink to="/scan" className={navClass}>
              {t("nav.scan")}
            </NavLink>
            {user ? (
              <>
                <NavLink to="/diario" className={navClass}>
                  {t("nav.diary")}
                </NavLink>
                <NavLink to="/ricette" className={navClass}>
                  {t("nav.recipes")}
                </NavLink>
                <NavLink to="/preferiti" className={navClass}>
                  {t("nav.favorites")}
                </NavLink>
                <NavLink to="/miei-alimenti" className={navClass}>
                  I miei alimenti
                </NavLink>
              </>
            ) : null}
            <NavLink to="/scopo" className={navClass}>
              {t("nav.scope")}
            </NavLink>
            {user ? (
              <>
                <NavLink to="/io/privacy" className={navClass}>
                  {t("nav.privacy")}
                </NavLink>
                <span
                  className="px-2 py-2 text-xs text-[color:var(--color-ink-muted)] truncate max-w-[14ch]"
                  title={user.email}
                  aria-label={t("auth.greeting", { email: user.email })}
                >
                  {user.email}
                </span>
                <button
                  type="button"
                  onClick={onLogout}
                  className="px-3 py-2 rounded-md text-sm font-medium text-[color:var(--color-ink-muted)] hover:text-[color:var(--color-ink)]"
                >
                  {t("auth.logout")}
                </button>
              </>
            ) : (
              <>
                <NavLink to="/accedi" className={navClass}>
                  {t("auth.login.title")}
                </NavLink>
                <NavLink to="/registrati" className={navClass}>
                  {t("auth.register.title")}
                </NavLink>
              </>
            )}
          </nav>
        </div>
      </header>

      <ScopeBanner />

      <main id="main" tabIndex={-1} className="flex-1 w-full max-w-5xl mx-auto px-4 py-6">
        <Outlet />
      </main>

      <Footer />
      <InstallPrompt />
    </div>
  );
}
