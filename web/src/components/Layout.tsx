import { useEffect, useRef } from "react";
import { Link, NavLink, Outlet, useLocation, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import Footer from "./Footer";
import ScopeBanner from "./ScopeBanner";
import InstallPrompt from "./InstallPrompt";
import Icon, { type IconName } from "./ui/Icon";
import { useAuth } from "../lib/auth/AuthContext";

export default function Layout() {
  const { t } = useTranslation();
  const { user } = useAuth();

  const primary: { to: string; label: string; icon: IconName; end?: boolean }[] = [
    { to: "/", label: t("nav.home"), icon: "home", end: true },
    { to: "/cerca", label: t("nav.search"), icon: "search" },
    { to: "/scan", label: t("nav.scan"), icon: "scan" },
    ...(user
      ? [
          { to: "/diario", label: t("nav.diary"), icon: "diary" as IconName },
          { to: "/ricette", label: t("nav.recipes"), icon: "recipe" as IconName },
        ]
      : [{ to: "/scopo", label: t("nav.scope"), icon: "info" as IconName }]),
  ];

  const pill = ({ isActive }: { isActive: boolean }) =>
    [
      "relative px-3.5 py-2 rounded-full text-sm font-medium no-underline transition-colors",
      isActive
        ? "bg-[color:var(--color-ink)] text-white hover:text-white"
        : "text-[color:var(--color-ink-muted)] hover:text-[color:var(--color-ink)] hover:bg-black/[0.04]",
    ].join(" ");

  return (
    <div className="min-h-dvh flex flex-col">
      <a href="#main" className="skip-link">
        {t("a11y.skipToContent")}
      </a>

      <header
        role="banner"
        className="sticky top-0 z-30 border-b border-black/[0.06] bg-[color:var(--color-canvas)]/75 backdrop-blur-xl backdrop-saturate-150"
      >
        <div className="max-w-6xl mx-auto px-4 md:px-6 h-16 flex items-center gap-6">
          <Link to="/" className="flex items-center gap-2.5 no-underline shrink-0">
            <img
              src="/icon.svg"
              alt=""
              aria-hidden="true"
              width={34}
              height={34}
              className="w-[34px] h-[34px] rounded-[10px] shadow-[var(--shadow-brand)]"
            />
            <span className="font-semibold text-[1.15rem] tracking-[-0.02em] text-[color:var(--color-ink)]">
              Pesa
            </span>
          </Link>

          <nav aria-label={t("a11y.primaryNav")} className="hidden md:flex items-center gap-1">
            {primary.map((item) => (
              <NavLink key={item.to} to={item.to} end={item.end} className={pill}>
                {item.label}
              </NavLink>
            ))}
          </nav>

          <div className="ml-auto flex items-center gap-2">
            {user ? (
              <AccountMenu email={user.email} />
            ) : (
              <>
                <Link
                  to="/accedi"
                  className="px-3.5 py-2 rounded-full text-sm font-medium no-underline text-[color:var(--color-ink)] hover:bg-black/[0.04]"
                >
                  {t("auth.login.title")}
                </Link>
                <Link to="/registrati" className="btn-primary !py-2 !px-4 text-sm hidden sm:inline-flex">
                  {t("auth.register.title")}
                </Link>
              </>
            )}
          </div>
        </div>
      </header>

      <ScopeBanner />

      <main
        id="main"
        tabIndex={-1}
        className="flex-1 w-full max-w-6xl mx-auto px-4 md:px-6 pt-6 md:pt-10 pb-28 md:pb-16 outline-none"
      >
        <Outlet />
      </main>

      <Footer />
      <InstallPrompt />

      {/* Mobile tab bar: thumb-reachable primary navigation, like a native app. */}
      <nav
        aria-label={t("a11y.primaryNav")}
        className="md:hidden fixed bottom-0 inset-x-0 z-30 border-t border-black/[0.06] bg-white/85 backdrop-blur-xl backdrop-saturate-150 pb-[env(safe-area-inset-bottom)]"
      >
        <ul className="grid grid-cols-4 h-16">
          {primary.slice(0, 4).map((item) => (
            <li key={item.to}>
              <NavLink
                to={item.to}
                end={item.end}
                className={({ isActive }) =>
                  [
                    "h-full flex flex-col items-center justify-center gap-1 text-[11px] font-medium no-underline transition-colors",
                    isActive
                      ? "text-[color:var(--color-brand-600)]"
                      : "text-[color:var(--color-ink-faint)] hover:text-[color:var(--color-ink)]",
                  ].join(" ")
                }
              >
                {({ isActive }) => (
                  <>
                    <span
                      className={[
                        "grid place-items-center w-12 h-7 rounded-full transition-colors",
                        isActive ? "bg-[color:var(--color-brand-50)]" : "",
                      ].join(" ")}
                    >
                      <Icon name={item.icon} size={22} strokeWidth={isActive ? 2 : 1.75} />
                    </span>
                    {item.label}
                  </>
                )}
              </NavLink>
            </li>
          ))}
        </ul>
      </nav>
    </div>
  );
}

/** Account popover: secondary destinations live here, out of the main bar. */
function AccountMenu({ email }: { email: string }) {
  const { t } = useTranslation();
  const { logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const ref = useRef<HTMLDetailsElement>(null);

  // Close on navigation and on outside click.
  useEffect(() => {
    if (ref.current) ref.current.open = false;
  }, [location.pathname]);
  useEffect(() => {
    function onDoc(e: MouseEvent) {
      if (ref.current?.open && !ref.current.contains(e.target as Node)) ref.current.open = false;
    }
    document.addEventListener("click", onDoc);
    return () => document.removeEventListener("click", onDoc);
  }, []);

  async function onLogout() {
    await logout();
    navigate("/", { replace: true });
  }

  const item =
    "flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm no-underline text-[color:var(--color-ink)] hover:bg-[color:var(--color-surface-muted)] hover:text-[color:var(--color-ink)]";

  return (
    <details ref={ref} className="relative group">
      <summary
        className="list-none [&::-webkit-details-marker]:hidden flex items-center gap-2 h-10 pl-1.5 pr-3 rounded-full cursor-pointer hover:bg-black/[0.04] select-none"
        aria-label={t("auth.greeting", { email })}
      >
        <span className="grid place-items-center w-8 h-8 rounded-full bg-[color:var(--color-ink)] text-white text-sm font-semibold uppercase">
          {email.slice(0, 1)}
        </span>
        <span className="hidden sm:block text-sm font-medium max-w-[16ch] truncate">{email}</span>
      </summary>
      <div className="absolute right-0 mt-2 w-64 surface p-2 shadow-[var(--shadow-lift)] z-40 origin-top-right rise">
        <p className="px-3 pt-2 pb-3 text-xs text-[color:var(--color-ink-muted)] truncate border-b border-[color:var(--color-border)] mb-1">
          {email}
        </p>
        <Link to="/preferiti" className={item}>
          <Icon name="heart" size={18} /> {t("nav.favorites")}
        </Link>
        <Link to="/miei-alimenti" className={item}>
          <Icon name="book" size={18} /> {t("nav.myFoods")}
        </Link>
        <Link to="/template" className={item}>
          <Icon name="diary" size={18} /> {t("nav.templates")}
        </Link>
        <Link to="/ricette" className={`${item} md:hidden`}>
          <Icon name="recipe" size={18} /> {t("nav.recipes")}
        </Link>
        <Link to="/io/privacy" className={item}>
          <Icon name="user" size={18} /> {t("nav.privacy")}
        </Link>
        <Link to="/scopo" className={item}>
          <Icon name="info" size={18} /> {t("nav.scope")}
        </Link>
        <div className="border-t border-[color:var(--color-border)] mt-1 pt-1">
          <button type="button" onClick={onLogout} className={`${item} w-full text-left !min-h-0`}>
            <Icon name="arrow" size={18} /> {t("auth.logout")}
          </button>
        </div>
      </div>
    </details>
  );
}
