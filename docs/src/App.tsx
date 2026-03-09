import { useState, useEffect, useRef, useCallback, createContext, useContext } from "react";
import { Routes, Route, Link, useLocation, useParams, useNavigate, Navigate, Outlet } from "react-router-dom";
import { type Theme, palette, ThemeContext, useTheme } from "./lib/theme.ts";
import { VERSION } from "./version.ts";
import { loadGuide, loadReference, loadLLMDoc } from "./lib/markdown.ts";
import Markdown from "./components/Markdown.tsx";
import TableOfContents from "./components/TableOfContents.tsx";
import SearchModal from "./components/SearchModal.tsx";
import HomePageExplore from "./components/HomePageExplore.tsx";

const SearchContext = createContext<{ openSearch: () => void }>({ openSearch: () => { } });
function useSearch() { return useContext(SearchContext); }

/* ─── Scroll to top on route change ─── */

function ScrollToTop() {
  const { pathname } = useLocation();
  useEffect(() => {
    history.scrollRestoration = "manual";
    window.scrollTo(0, 0);
  }, [pathname]);
  return null;
}

/* ─── Nav ─── */

function Nav() {
  const { theme, setTheme } = useTheme();
  const { openSearch: onOpenSearch } = useSearch();
  const t = palette[theme];
  const location = useLocation();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [starCount, setStarCount] = useState<string | null>(() => {
    try {
      const raw = localStorage.getItem("gh-stars");
      if (raw) {
        const { value, expiry } = JSON.parse(raw);
        if (Date.now() < expiry) return value;
      }
    } catch {}
    return null;
  });

  // Fetch GitHub star count (cached for 15 minutes to avoid rate limits)
  useEffect(() => {
    try {
      const raw = localStorage.getItem("gh-stars");
      if (raw) {
        const { expiry } = JSON.parse(raw);
        if (Date.now() < expiry) return;
      }
    } catch {}
    fetch("https://api.github.com/repos/grindlemire/go-tui")
      .then((r) => {
        if (!r.ok) return null;
        return r.json();
      })
      .then((data) => {
        if (data?.stargazers_count != null) {
          const count = data.stargazers_count;
          const formatted = count >= 1000 ? `${(count / 1000).toFixed(1)}k` : String(count);
          localStorage.setItem("gh-stars", JSON.stringify({ value: formatted, expiry: Date.now() + 900000 }));
          setStarCount(formatted);
        }
      })
      .catch(() => {});
  }, []);

  // Close mobile menu on navigation
  useEffect(() => {
    setMobileOpen(false);
  }, [location.pathname]);

  const isActive = (path: string) => {
    if (path === "/")
      return location.pathname === "/" || location.pathname === "";
    return location.pathname.startsWith(path);
  };

  const links = [
    { to: "/", label: "home" },
    { to: "/guide", label: "guide" },
    { to: "/reference", label: "reference" },
  ];

  return (
    <nav
      className="sticky top-0 left-0 right-0 z-40 backdrop-blur-md"
      style={{
        background:
          theme === "dark"
            ? "rgba(39, 40, 34, 0.92)"
            : "rgba(250, 250, 248, 0.92)",
        borderBottom: `1px solid ${t.border}`,
      }}
    >
      <div className="max-w-[1100px] mx-auto px-4 sm:px-6 h-12 flex items-center justify-between">
        <Link
          to="/"
          className="flex items-center"
          onClick={() => window.scrollTo({ top: 0, behavior: location.pathname === "/" ? "smooth" : "instant" })}
        >
          <img
            src={theme === "dark" ? "/go-tui-logo.svg" : "/go-tui-logo-light-bg.svg"}
            alt="go-tui"
            style={{ height: 32 }}
          />
        </Link>

        {/* Desktop links */}
        <div className="hidden sm:flex items-center">
          {links.map((link) => {
            const active = isActive(link.to);
            return (
              <Link
                key={link.to}
                to={link.to}
                className="font-['Fira_Code',monospace] text-xs px-1.5 py-1.5 rounded transition-all duration-200"
                style={{
                  color: active ? t.accent : t.textMuted,
                  background: active
                    ? theme === "dark"
                      ? "#66d9ef0a"
                      : "#2f9eb80a"
                    : "transparent",
                  border: `1px solid ${active ? (theme === "dark" ? "#66d9ef33" : "#2f9eb833") : "transparent"}`,
                  textShadow: "none",
                }}
                onClick={link.to === "/" ? () => window.scrollTo({ top: 0, behavior: location.pathname === "/" ? "smooth" : "instant" }) : undefined}
                onMouseEnter={(e) => {
                  if (!active) e.currentTarget.style.color = t.accent;
                }}
                onMouseLeave={(e) => {
                  if (!active) e.currentTarget.style.color = t.textMuted;
                }}
              >
                {link.label}
              </Link>
            );
          })}

          {/* Search bar */}
          <button
            onClick={onOpenSearch}
            className="font-['Fira_Code',monospace] text-xs rounded transition-all duration-200 flex items-center gap-2 ml-2"
            style={{
              color: t.textDim,
              background: theme === "dark" ? "rgba(62, 61, 50, 0.4)" : "rgba(232, 232, 227, 0.5)",
              border: `1px solid ${t.border}`,
              cursor: "pointer",
              padding: "5px 8px 5px 10px",
              minWidth: 160,
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.borderColor = theme === "dark" ? "#66d9ef44" : "#2f9eb844";
              e.currentTarget.style.background = theme === "dark" ? "rgba(62, 61, 50, 0.7)" : "rgba(232, 232, 227, 0.8)";
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.borderColor = t.border;
              e.currentTarget.style.background = theme === "dark" ? "rgba(62, 61, 50, 0.4)" : "rgba(232, 232, 227, 0.5)";
            }}
          >
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" style={{ flexShrink: 0, opacity: 0.7 }}>
              <circle cx="11" cy="11" r="8" />
              <line x1="21" y1="21" x2="16.65" y2="16.65" />
            </svg>
            <span style={{ flex: 1, textAlign: "left" }}>
              search...
            </span>
            <kbd
              style={{
                fontSize: 10,
                color: t.textDim,
                background: theme === "dark" ? "rgba(62, 61, 50, 0.6)" : "rgba(216, 216, 208, 0.6)",
                border: `1px solid ${theme === "dark" ? "#49483e" : "#d0d0c8"}`,
                borderRadius: 4,
                padding: "1px 5px",
                lineHeight: 1.4,
                flexShrink: 0,
                fontFamily: "'Fira Code', monospace",
              }}
            >
              {typeof navigator !== "undefined" && /Mac|iPhone|iPad/.test(navigator.userAgent) ? "\u2318K" : "Ctrl K"}
            </kbd>
          </button>

          <div
            className="mx-3"
            style={{ width: 1, height: 20, background: t.border }}
          />

          <a
            href="https://pkg.go.dev/github.com/grindlemire/go-tui"
            target="_blank"
            rel="noopener noreferrer"
            className="font-['Fira_Code',monospace] text-[10px] px-2 py-1 rounded transition-all duration-200 mr-1"
            style={{
              color: t.secondary,
              background: `${t.secondary}0a`,
              border: `1px solid ${t.secondary}22`,
            }}
            title={`v${VERSION} — view on pkg.go.dev`}
            onMouseEnter={(e) => {
              e.currentTarget.style.borderColor = `${t.secondary}55`;
              e.currentTarget.style.background = `${t.secondary}14`;
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.borderColor = `${t.secondary}22`;
              e.currentTarget.style.background = `${t.secondary}0a`;
            }}
          >
            v{VERSION}
          </a>

          <a
            href="https://github.com/grindlemire/go-tui"
            target="_blank"
            rel="noopener noreferrer"
            className="p-1.5 rounded transition-all duration-200 flex items-center gap-1.5 mr-1"
            style={{
              color: t.textMuted,
              border: `1px solid transparent`,
            }}
            title="View on GitHub"
            onMouseEnter={(e) => {
              e.currentTarget.style.color = t.accent;
              e.currentTarget.style.borderColor = t.border;
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.color = t.textMuted;
              e.currentTarget.style.borderColor = "transparent";
            }}
          >
            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor" aria-label="GitHub">
              <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z" />
            </svg>
            {starCount && (
              <span className="font-['Fira_Code',monospace] text-[10px] flex items-center gap-0.5">
                <svg width="10" height="10" viewBox="0 0 16 16" fill="currentColor">
                  <path d="M8 .25a.75.75 0 0 1 .673.418l1.882 3.815 4.21.612a.75.75 0 0 1 .416 1.279l-3.046 2.97.719 4.192a.75.75 0 0 1-1.088.791L8 12.347l-3.766 1.98a.75.75 0 0 1-1.088-.79l.72-4.194L.818 6.374a.75.75 0 0 1 .416-1.28l4.21-.611L7.327.668A.75.75 0 0 1 8 .25z" />
                </svg>
                {starCount}
              </span>
            )}
          </a>

          <button
            onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
            className="font-['Fira_Code',monospace] text-xs p-1.5 rounded transition-all duration-300"
            style={{
              color: theme === "dark" ? t.secondary : t.tertiary,
              background: "transparent",
              border: `1px solid ${t.border}`,
              cursor: "pointer",
              lineHeight: 1,
            }}
            title={
              theme === "dark" ? "Switch to light mode" : "Switch to dark mode"
            }
          >
            {theme === "dark" ? "light" : "dark"}
          </button>
        </div>

        {/* Mobile hamburger */}
        <div className="flex sm:hidden items-center gap-2">
          <button
            onClick={onOpenSearch}
            className="p-1.5 rounded flex items-center"
            style={{ color: t.textMuted, background: "transparent", border: "none", cursor: "pointer" }}
            title="Search docs"
          >
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <circle cx="11" cy="11" r="8" />
              <line x1="21" y1="21" x2="16.65" y2="16.65" />
            </svg>
          </button>
          <a
            href="https://github.com/grindlemire/go-tui"
            target="_blank"
            rel="noopener noreferrer"
            className="p-1.5 rounded flex items-center gap-1"
            style={{ color: t.textMuted }}
            title="View on GitHub"
          >
            <svg width="15" height="15" viewBox="0 0 16 16" fill="currentColor" aria-label="GitHub">
              <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z" />
            </svg>
            {starCount && (
              <span className="font-['Fira_Code',monospace] text-[10px] flex items-center gap-0.5">
                <svg width="10" height="10" viewBox="0 0 16 16" fill="currentColor">
                  <path d="M8 .25a.75.75 0 0 1 .673.418l1.882 3.815 4.21.612a.75.75 0 0 1 .416 1.279l-3.046 2.97.719 4.192a.75.75 0 0 1-1.088.791L8 12.347l-3.766 1.98a.75.75 0 0 1-1.088-.79l.72-4.194L.818 6.374a.75.75 0 0 1 .416-1.28l4.21-.611L7.327.668A.75.75 0 0 1 8 .25z" />
                </svg>
                {starCount}
              </span>
            )}
          </a>
          <button
            onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
            className="font-['Fira_Code',monospace] text-[10px] p-1.5 rounded"
            style={{
              color: theme === "dark" ? t.secondary : t.tertiary,
              background: "transparent",
              border: `1px solid ${t.border}`,
              cursor: "pointer",
            }}
          >
            {theme === "dark" ? "light" : "dark"}
          </button>
          <button
            onClick={() => setMobileOpen(!mobileOpen)}
            className="font-['Fira_Code',monospace] text-sm p-1.5"
            style={{
              color: t.textMuted,
              background: "transparent",
              border: "none",
              cursor: "pointer",
            }}
          >
            {mobileOpen ? "\u2715" : "\u2630"}
          </button>
        </div>
      </div>

      {/* Mobile dropdown */}
      {mobileOpen && (
        <div
          className="sm:hidden px-4 pb-3 flex flex-col gap-1"
          style={{
            borderTop: `1px solid ${t.border}`,
            background:
              theme === "dark"
                ? "rgba(39, 40, 34, 0.95)"
                : "rgba(250, 250, 248, 0.98)",
          }}
        >
          {links.map((link) => {
            const active = isActive(link.to);
            return (
              <Link
                key={link.to}
                to={link.to}
                className="font-['Fira_Code',monospace] text-sm px-3 py-2 rounded"
                style={{
                  color: active ? t.accent : t.textMuted,
                  background: active
                    ? theme === "dark"
                      ? "#66d9ef0a"
                      : "#2f9eb80a"
                    : "transparent",
                }}
                onClick={link.to === "/" ? () => window.scrollTo({ top: 0, behavior: location.pathname === "/" ? "smooth" : "instant" }) : undefined}
              >
                {link.label}
              </Link>
            );
          })}
          <div
            className="h-px my-1"
            style={{ background: t.border }}
          />
          <a
            href="https://pkg.go.dev/github.com/grindlemire/go-tui"
            target="_blank"
            rel="noopener noreferrer"
            className="font-['Fira_Code',monospace] text-sm px-3 py-2 rounded flex items-center gap-2"
            style={{ color: t.textMuted }}
          >
            <span
              className="text-[10px] px-1.5 py-0.5 rounded"
              style={{
                color: t.secondary,
                background: `${t.secondary}0a`,
                border: `1px solid ${t.secondary}22`,
              }}
            >
              v{VERSION}
            </span>
            pkg.go.dev
          </a>
        </div>
      )}
    </nav>
  );
}

/* ─── Footer ─── */

function Footer() {
  const { theme } = useTheme();
  const t = palette[theme];
  return (
    <footer
      className="mt-20"
      style={{ borderTop: `1px solid ${t.border}` }}
    >
      <div className="py-6 flex flex-col items-center justify-center gap-2 font-['Fira_Code',monospace] text-[11px]">
        <div className="flex items-center gap-4">
          <a
            href="https://github.com/grindlemire/go-tui"
            target="_blank"
            rel="noopener noreferrer"
            className="transition-colors duration-200"
            style={{ color: t.textDim }}
            onMouseEnter={(e) => (e.currentTarget.style.color = t.accent)}
            onMouseLeave={(e) => (e.currentTarget.style.color = t.textDim)}
          >
            GitHub
          </a>
          <span style={{ color: t.border }}>·</span>
          <a
            href="https://pkg.go.dev/github.com/grindlemire/go-tui"
            target="_blank"
            rel="noopener noreferrer"
            className="transition-colors duration-200"
            style={{ color: t.textDim }}
            onMouseEnter={(e) => (e.currentTarget.style.color = t.accent)}
            onMouseLeave={(e) => (e.currentTarget.style.color = t.textDim)}
          >
            pkg.go.dev
          </a>
        </div>
        <span style={{ color: t.textDim }}>
          &copy; {new Date().getFullYear()} Joel Holsteen. All rights reserved.
        </span>
      </div>
    </footer>
  );
}

/* ─── Page Wrapper ─── */

function PageShell({ children }: { children: React.ReactNode }) {
  const { theme } = useTheme();
  const t = palette[theme];
  return (
    <div
      className={`${theme === "dark" ? "dark-theme" : "light-theme"} neon-select overflow-x-clip flex flex-col`}
      style={{
        background: t.bg,
        color: t.text,
        minHeight: "100vh",
        fontFamily: "'IBM Plex Sans', sans-serif",
      }}
    >
      <Nav />
      <div className="flex-1">
        {children}
      </div>
      <Footer />
    </div>
  );
}

function PageLayout() {
  return (
    <PageShell>
      <Outlet />
    </PageShell>
  );
}

/* ============================================================
   Pages
   ============================================================ */



/* ─── Prev / Next Navigation ─── */

function PrevNextNav({
  pages,
  activeIndex,
  basePath,
}: {
  pages: { slug: string; title: string }[];
  activeIndex: number;
  basePath: string;
}) {
  const { theme } = useTheme();
  const t = palette[theme];

  return (
    <div
      className="mt-10 sm:mt-12 pt-6 sm:pt-8 flex justify-between"
      style={{ borderTop: `1px solid ${t.border}` }}
    >
      {activeIndex > 0 ? (
        <Link
          to={`${basePath}/${pages[activeIndex - 1].slug}`}
          className="font-['Fira_Code',monospace] text-xs sm:text-sm transition-colors duration-200"
          style={{
            color: t.textMuted,
            textDecoration: "none",
          }}
          onMouseEnter={(e) => (e.currentTarget.style.color = t.accent)}
          onMouseLeave={(e) => (e.currentTarget.style.color = t.textMuted)}
        >
          &larr; {pages[activeIndex - 1].title}
        </Link>
      ) : (
        <div />
      )}
      {activeIndex < pages.length - 1 ? (
        <Link
          to={`${basePath}/${pages[activeIndex + 1].slug}`}
          className="font-['Fira_Code',monospace] text-xs sm:text-sm transition-colors duration-200"
          style={{
            color: t.textMuted,
            textDecoration: "none",
          }}
          onMouseEnter={(e) => (e.currentTarget.style.color = t.accent)}
          onMouseLeave={(e) => (e.currentTarget.style.color = t.textMuted)}
        >
          {pages[activeIndex + 1].title} &rarr;
        </Link>
      ) : (
        <div />
      )}
    </div>
  );
}

/* ─── Mobile Page Picker ─── */

function MobilePicker({
  pages,
  activeIndex,
  onSelect,
}: {
  pages: { slug: string; title: string }[];
  activeIndex: number;
  onSelect: (index: number) => void;
}) {
  const { theme } = useTheme();
  const t = palette[theme];
  const [open, setOpen] = useState(false);

  return (
    <div className="md:hidden mb-6 sm:mb-8 relative">
      <button
        onClick={() => setOpen(!open)}
        className="font-['Fira_Code',monospace] w-full rounded-lg px-4 py-3 text-[13px] text-left flex items-center justify-between gap-2"
        style={{
          background: t.bgSecondary,
          color: t.text,
          border: `1px solid ${t.border}`,
        }}
      >
        <div className="flex items-center gap-2.5 min-w-0">
          <span
            className="text-[10px] px-1.5 py-0.5 rounded shrink-0"
            style={{
              background: `${t.accent}15`,
              color: t.accent,
              border: `1px solid ${t.accent}30`,
            }}
          >
            {String(activeIndex + 1).padStart(2, "0")}
          </span>
          <span className="truncate">{pages[activeIndex].title}</span>
        </div>
        <svg
          width="12"
          height="12"
          viewBox="0 0 12 12"
          fill="none"
          className="shrink-0 transition-transform duration-200"
          style={{
            transform: open ? "rotate(180deg)" : "rotate(0deg)",
            color: t.textDim,
          }}
        >
          <path d="M2.5 4.5L6 8L9.5 4.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      </button>

      {open && (
        <>
          <div className="fixed inset-0 z-30" onClick={() => setOpen(false)} />
          <div
            className="absolute top-full left-0 right-0 mt-1 rounded-lg overflow-hidden z-40"
            style={{
              background: t.bgSecondary,
              border: `1px solid ${t.border}`,
              boxShadow: theme === "dark"
                ? "0 8px 24px rgba(0,0,0,0.5)"
                : "0 8px 24px rgba(0,0,0,0.12)",
            }}
          >
            {pages.map((page, i) => {
              const active = i === activeIndex;
              return (
                <button
                  key={page.slug}
                  onClick={() => {
                    onSelect(i);
                    setOpen(false);
                  }}
                  className="font-['Fira_Code',monospace] w-full text-left px-4 py-2.5 text-[12px] flex items-center gap-2.5 transition-colors duration-100"
                  style={{
                    color: active ? t.accent : t.textMuted,
                    background: active
                      ? `${t.accent}0a`
                      : "transparent",
                    borderBottom: i < pages.length - 1 ? `1px solid ${t.border}` : "none",
                  }}
                >
                  <span
                    className="text-[10px] w-5 text-center shrink-0"
                    style={{ color: active ? t.accent : t.textDim }}
                  >
                    {String(i + 1).padStart(2, "0")}
                  </span>
                  {page.title}
                </button>
              );
            })}
          </div>
        </>
      )}
    </div>
  );
}

/* ─── Guide Page ─── */

function GuideRedirect() {
  const pages = loadGuide();
  if (pages.length === 0) return null;
  return <Navigate to={`/guide/${pages[0].slug}`} replace />;
}

function GuidePage() {
  const { theme } = useTheme();
  const t = palette[theme];
  const { slug } = useParams();
  const navigate = useNavigate();
  const location = useLocation();

  const pages = loadGuide();
  const activeSection = Math.max(0, pages.findIndex((p) => p.slug === slug));

  // Deep link: scroll to hash target on mount / page change
  useEffect(() => {
    const hash = location.hash.replace("#", "");
    if (!hash) return;
    // Wait for markdown to render
    const timer = setTimeout(() => {
      const el = document.getElementById(hash);
      if (el) {
        el.scrollIntoView({ behavior: "smooth", block: "start" });
      }
    }, 100);
    return () => clearTimeout(timer);
  }, [slug, location.hash]);

  return (
    <div className="max-w-[1100px] xl:max-w-[1360px] mx-auto px-4 sm:px-6 pt-10 sm:pt-16 pb-16 sm:pb-24">
      <h1
        className="text-3xl sm:text-5xl font-bold tracking-tight mb-8 sm:mb-12"
        style={{ color: t.heading }}
      >
        Guide
      </h1>

      <div className="flex gap-8 sm:gap-10">
        {/* Desktop Sidebar */}
        <div className="w-48 shrink-0 hidden md:block">
          <div className="sticky top-16">
            <div
              className="font-['Fira_Code',monospace] text-[10px] tracking-[0.15em] uppercase mb-4"
              style={{ color: t.textDim }}
            >
              chapters
            </div>

            {pages.map((page, i) => {
              const active = activeSection === i;
              return (
                <Link
                  key={page.slug}
                  to={`/guide/${page.slug}`}
                  className="block w-full text-left font-['Fira_Code',monospace] text-[12px] py-1.5 px-3 rounded transition-all duration-200"
                  style={{
                    color: active ? t.accent : t.textMuted,
                    background: active
                      ? theme === "dark"
                        ? "#66d9ef0d"
                        : "#2f9eb80d"
                      : "transparent",
                    textDecoration: "none",
                    borderLeft: `2px solid ${active ? t.accent : "transparent"}`,
                  }}
                  onMouseEnter={(e) => {
                    if (!active) e.currentTarget.style.color = t.accent;
                  }}
                  onMouseLeave={(e) => {
                    if (!active) e.currentTarget.style.color = t.textMuted;
                  }}
                >
                  {page.title}
                </Link>
              );
            })}

            <SidebarLLMButton label="copy all as markdown" />
          </div>
        </div>

        {/* Main content */}
        <div className="flex-1 min-w-0">
          <MobilePicker
            pages={pages}
            activeIndex={activeSection}
            onSelect={(i) => navigate(`/guide/${pages[i].slug}`)}
          />

          <RawMarkdownButton body={pages[activeSection].body} />

          <div className="fade-in" key={slug}>
            <Markdown content={pages[activeSection].body} />
          </div>

          <PrevNextNav
            pages={pages}
            activeIndex={activeSection}
            basePath="/guide"
          />
        </div>

        {/* On-page TOC */}
        <TableOfContents content={pages[activeSection].body} key={`toc-${slug}`} />
      </div>
    </div>
  );
}

/* ─── Sidebar LLM Copy Button ─── */

function SidebarLLMButton({ label }: { label: string }) {
  const { theme } = useTheme();
  const t = palette[theme];
  const [copied, setCopied] = useState(false);

  return (
    <button
      onClick={() => {
        navigator.clipboard.writeText(loadLLMDoc());
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
      }}
      className="font-['Fira_Code',monospace] text-[10px] flex items-center gap-1.5 mt-6 px-3 py-1.5 rounded transition-all duration-200 w-full"
      style={{
        color: copied ? t.secondary : t.textDim,
        background: "transparent",
        border: "none",
        cursor: "pointer",
        borderTop: `1px solid ${t.border}`,
        paddingTop: "12px",
      }}
      onMouseEnter={(e) => {
        if (!copied) e.currentTarget.style.color = t.accent;
      }}
      onMouseLeave={(e) => {
        if (!copied) e.currentTarget.style.color = t.textDim;
      }}
      title="Copy all docs as a single LLM-optimized markdown file"
    >
      {copied ? (
        <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
          <polyline points="20 6 9 17 4 12" />
        </svg>
      ) : (
        <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
          <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
        </svg>
      )}
      {copied ? "copied!" : label}
    </button>
  );
}

/* ─── Doc Action Buttons ─── */

function CopyButton({
  text,
  label,
  copiedLabel,
  title,
}: {
  text: string;
  label: string;
  copiedLabel: string;
  title: string;
}) {
  const { theme } = useTheme();
  const t = palette[theme];
  const [copied, setCopied] = useState(false);

  return (
    <button
      onClick={() => {
        navigator.clipboard.writeText(text);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
      }}
      className="font-['Fira_Code',monospace] text-[11px] flex items-center gap-1.5 px-2.5 py-1 rounded transition-all duration-200"
      style={{
        color: copied ? t.secondary : t.textDim,
        background: "transparent",
        border: `1px solid ${copied ? t.secondary + "40" : "transparent"}`,
        cursor: "pointer",
      }}
      onMouseEnter={(e) => {
        if (!copied) {
          e.currentTarget.style.color = t.accent;
          e.currentTarget.style.borderColor = t.accent + "30";
        }
      }}
      onMouseLeave={(e) => {
        if (!copied) {
          e.currentTarget.style.color = t.textDim;
          e.currentTarget.style.borderColor = "transparent";
        }
      }}
      title={title}
    >
      {copied ? (
        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
          <polyline points="20 6 9 17 4 12" />
        </svg>
      ) : (
        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
          <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
        </svg>
      )}
      {copied ? copiedLabel : label}
    </button>
  );
}

function RawMarkdownButton({ body }: { body: string }) {
  return (
    <div className="flex justify-end mb-3">
      <CopyButton
        text={body}
        label="raw markdown"
        copiedLabel="copied!"
        title="Copy raw markdown to clipboard"
      />
    </div>
  );
}

/* ─── Raw Guide Page (plain markdown for AI consumption) ─── */

function RawGuidePage() {
  const { slug } = useParams();
  const pages = loadGuide();
  const page = pages.find((p) => p.slug === slug);

  if (!page) return <pre>Guide not found.</pre>;

  return (
    <pre
      style={{
        margin: 0,
        padding: "1rem",
        whiteSpace: "pre-wrap",
        wordBreak: "break-word",
        fontFamily: "'Fira Code', monospace",
        fontSize: "13px",
        lineHeight: 1.6,
        background: "#1a1a2e",
        color: "#e0e0e0",
        minHeight: "100vh",
      }}
    >
      {page.body}
    </pre>
  );
}

/* ─── Reference Page ─── */

function ReferenceRedirect() {
  const pages = loadReference();
  if (pages.length === 0) return null;
  return <Navigate to={`/reference/${pages[0].slug}`} replace />;
}

function ReferencePage() {
  const { theme } = useTheme();
  const t = palette[theme];
  const { slug } = useParams();
  const navigate = useNavigate();
  const location = useLocation();

  const pages = loadReference();
  const activeCategory = Math.max(0, pages.findIndex((p) => p.slug === slug));

  // Deep link: scroll to hash target on mount / page change
  useEffect(() => {
    const hash = location.hash.replace("#", "");
    if (!hash) return;
    const timer = setTimeout(() => {
      const el = document.getElementById(hash);
      if (el) {
        el.scrollIntoView({ behavior: "smooth", block: "start" });
      }
    }, 100);
    return () => clearTimeout(timer);
  }, [slug, location.hash]);

  return (
    <div className="max-w-[1100px] xl:max-w-[1360px] mx-auto px-4 sm:px-6 pt-10 sm:pt-16 pb-16 sm:pb-24">
      <h1
        className="text-3xl sm:text-5xl font-bold tracking-tight mb-8 sm:mb-12"
        style={{ color: t.heading }}
      >
        API Reference
      </h1>

      <div className="flex gap-8 sm:gap-10">
        {/* Desktop Sidebar */}
        <div className="w-48 shrink-0 hidden md:block">
          <div className="sticky top-16">
            <div
              className="font-['Fira_Code',monospace] text-[10px] tracking-[0.15em] uppercase mb-4"
              style={{ color: t.textDim }}
            >
              categories
            </div>

            {pages.map((page, i) => {
              const active = activeCategory === i;
              const isTailwind = page.slug === "tailwind-classes";
              const activeColor = isTailwind ? t.secondary : t.accent;
              return (
                <Link
                  key={page.slug}
                  to={`/reference/${page.slug}`}
                  className="block w-full text-left font-['Fira_Code',monospace] text-[12px] py-1.5 px-3 rounded transition-all duration-200"
                  style={{
                    color: active ? activeColor : t.textMuted,
                    background: active
                      ? isTailwind
                        ? theme === "dark"
                          ? "#a6e22e08"
                          : "#638b0c08"
                        : theme === "dark"
                          ? "#66d9ef0d"
                          : "#2f9eb80d"
                      : "transparent",
                    textDecoration: "none",
                    borderLeft: `2px solid ${active ? activeColor : "transparent"}`,
                    textShadow:
                      active && theme === "dark" && !isTailwind
                        ? t.accentGlowSubtle
                        : "none",
                  }}
                  onMouseEnter={(e) => {
                    if (!active) e.currentTarget.style.color = isTailwind ? t.secondary : t.accent;
                  }}
                  onMouseLeave={(e) => {
                    if (!active) e.currentTarget.style.color = t.textMuted;
                  }}
                >
                  {page.title}
                </Link>
              );
            })}

            <SidebarLLMButton label="copy all as markdown" />
          </div>
        </div>

        {/* Main content */}
        <div className="flex-1 min-w-0">
          <MobilePicker
            pages={pages}
            activeIndex={activeCategory}
            onSelect={(i) => navigate(`/reference/${pages[i].slug}`)}
          />

          <RawMarkdownButton body={pages[activeCategory].body} />

          <div className="fade-in" key={slug}>
            <Markdown content={pages[activeCategory].body} />
          </div>

          <PrevNextNav
            pages={pages}
            activeIndex={activeCategory}
            basePath="/reference"
          />
        </div>

        {/* On-page TOC */}
        <TableOfContents content={pages[activeCategory].body} key={`toc-${slug}`} />
      </div>
    </div>
  );
}

/* ─── Main Export ─── */

export default function Design2() {
  const [theme, setThemeState] = useState<Theme>(() => {
    const saved = localStorage.getItem("go-tui-theme");
    return saved === "light" || saved === "dark" ? saved : "dark";
  });
  const setTheme = (t: Theme) => {
    localStorage.setItem("go-tui-theme", t);
    setThemeState(t);
  };

  const [searchOpen, setSearchOpen] = useState(false);
  const openSearch = useCallback(() => setSearchOpen(true), []);

  // Global Cmd+K / Ctrl+K shortcut
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        setSearchOpen(true);
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, []);

  return (
    <ThemeContext.Provider value={{ theme, setTheme }}>
      <SearchContext.Provider value={{ openSearch }}>
        <ScrollToTop />
        <SearchModal open={searchOpen} onClose={() => setSearchOpen(false)} />
        <Routes>
          <Route element={<PageLayout />}>
            <Route path="/" element={<HomePageExplore />} />
            <Route path="/guide" element={<GuideRedirect />} />
            <Route path="/guide/:slug" element={<GuidePage />} />
            <Route path="/reference" element={<ReferenceRedirect />} />
            <Route path="/reference/:slug" element={<ReferencePage />} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Route>
          <Route path="/guide/:slug/raw" element={<RawGuidePage />} />
        </Routes>
      </SearchContext.Provider>
    </ThemeContext.Provider>
  );
}
