import { useState, useEffect, useRef, useCallback, createContext, useContext } from "react";
import { Routes, Route, Link, useLocation, useParams, useNavigate, Navigate, Outlet } from "react-router-dom";
import { type Theme, palette, ThemeContext, useTheme } from "./lib/theme.ts";
import { tailwindClasses } from "./content/projectInfo.ts";
import { VERSION } from "./version.ts";
import { loadGuide, loadReference, loadLLMDoc } from "./lib/markdown.ts";
import Markdown from "./components/Markdown.tsx";
import TableOfContents from "./components/TableOfContents.tsx";
import CodeShowcase from "./components/CodeShowcase.tsx";
import SearchModal from "./components/SearchModal.tsx";
import HomePageExplore from "./components/HomePageExplore.tsx";
import Divider from "./components/Divider.tsx";
import PageBackground from "./components/PageBackground.tsx";
import DxCapability from "./components/DxCapability.tsx";
import EditorSimulation from "./components/EditorSimulation.tsx";
import ComparisonSection from "./components/ComparisonSection.tsx";

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

function Page({ children }: { children: React.ReactNode }) {
  return <PageShell>{children}</PageShell>;
}


/* ============================================================
   Pages
   ============================================================ */


function HomePage() {
  const { theme } = useTheme();
  const t = palette[theme];

  // Shared DX feature state — editor + capability list both read/write this
  const [dxFeature, setDxFeature] = useState(0);
  const dxPausedRef = useRef(false);

  // Auto-cycle when not paused
  useEffect(() => {
    const interval = setInterval(() => {
      if (!dxPausedRef.current) {
        setDxFeature((prev) => (prev + 1) % 5);
      }
    }, 4000);
    return () => clearInterval(interval);
  }, []);

  // Scroll so the next section's label + heading are visible below the nav
  const scrollToNextSection = () => {
    const navHeight = 48; // h-12
    const sections = document.querySelectorAll<HTMLElement>("section:has(> h2)");
    const scrollY = window.scrollY;
    for (const s of sections) {
      const top = s.getBoundingClientRect().top + scrollY - navHeight;
      if (top > scrollY + 20) {
        window.scrollTo({ top, behavior: "smooth" });
        return;
      }
    }
    // Already past last section — scroll to bottom
    window.scrollTo({ top: document.body.scrollHeight, behavior: "smooth" });
  };

  // Desktop: Enter/Space scrolls to next section heading
  useEffect(() => {
    const isMobile = () => window.matchMedia("(max-width: 639px)").matches;
    const onKeyDown = (e: KeyboardEvent) => {
      if (isMobile()) return;
      const tag = (e.target as HTMLElement).tagName;
      if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return;
      if (e.metaKey || e.ctrlKey || e.altKey) return;
      if (e.key === "Enter" || e.key === " ") {
        e.preventDefault();
        scrollToNextSection();
      }
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, []);


  return (
    <Page>
      <div className="relative">
        <PageBackground theme={theme} />
        <div className="relative z-10">
          {/* Hero — Man Page Terminal */}
          <section className="relative" style={{ minHeight: "100vh" }}>

            <div
              className="flex flex-col"
              style={{
                minHeight: "100vh",
                background: theme === "dark" ? "#1e1f1a" : "#f0f0ec",
              }}
            >
              {/* Terminal body */}
              <div
                className="flex-1 flex flex-col justify-center overflow-auto font-['Fira_Code',monospace] text-[13px] leading-[1.6]"
                style={{ padding: "24px 32px 24px clamp(32px, 12vw, 200px)" }}
              >
                <div className="w-full max-w-[720px]">
                  {/* Prompt line */}
                  <div className="tl mb-4 text-[13px]" style={{ animationDelay: "10ms" }}>
                    <span style={{ color: t.secondary }}>$</span>{" "}
                    <span style={{ color: t.heading }}>man tui</span>
                  </div>

                  {/* ASCII Art — REACTIVE */}
                  {[
                    " ██████╗ ███████╗ █████╗  ██████╗████████╗██╗██╗   ██╗███████╗",
                    " ██╔══██╗██╔════╝██╔══██╗██╔════╝╚══██╔══╝██║██║   ██║██╔════╝",
                    " ██████╔╝█████╗  ███████║██║        ██║   ██║██║   ██║█████╗",
                    " ██╔══██╗██╔══╝  ██╔══██║██║        ██║   ██║╚██╗ ██╔╝██╔══╝",
                    " ██║  ██║███████╗██║  ██║╚██████╗   ██║   ██║ ╚████╔╝ ███████╗",
                    " ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝ ╚═════╝   ╚═╝   ╚═╝  ╚═══╝  ╚══════╝",
                  ].map((line, i) => (
                    <div
                      key={`r${i}`}
                      className="tl whitespace-pre leading-[1.15] overflow-hidden"
                      style={{
                        animationDelay: `${30 + i * 5}ms`,
                        color: t.heading,
                        fontSize: "clamp(7px, 1.15vw, 13px)",
                        letterSpacing: 0,
                      }}
                    >
                      {line}
                    </div>
                  ))}

                  <div className="tl h-[2px]" style={{ animationDelay: "60ms" }} />

                  {/* ASCII Art — TERMINAL */}
                  {[
                    " ████████╗███████╗██████╗ ███╗   ███╗██╗███╗   ██╗ █████╗ ██╗",
                    " ╚══██╔══╝██╔════╝██╔══██╗████╗ ████║██║████╗  ██║██╔══██╗██║",
                    "    ██║   █████╗  ██████╔╝██╔████╔██║██║██╔██╗ ██║███████║██║",
                    "    ██║   ██╔══╝  ██╔══██╗██║╚██╔╝██║██║██║╚██╗██║██╔══██║██║",
                    "    ██║   ███████╗██║  ██║██║ ╚═╝ ██║██║██║ ╚████║██║  ██║███████╗",
                    "    ╚═╝   ╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝╚══════╝",
                  ].map((line, i) => (
                    <div
                      key={`t${i}`}
                      className="tl whitespace-pre leading-[1.15] overflow-hidden"
                      style={{
                        animationDelay: `${65 + i * 5}ms`,
                        color: t.accent,
                        fontSize: "clamp(7px, 1.15vw, 13px)",
                        letterSpacing: 0,
                      }}
                    >
                      {line}
                    </div>
                  ))}

                  <div className="tl h-[2px]" style={{ animationDelay: "100ms" }} />

                  {/* ASCII Art — UIs  in  Go — "s" and "in" as regular text */}
                  <div
                    className="tl flex items-end gap-0"
                    style={{ animationDelay: "105ms" }}
                  >
                    {/* UI block letters */}
                    <div className="whitespace-pre leading-[1.15] overflow-hidden" style={{ fontSize: "clamp(7px, 1.15vw, 13px)", letterSpacing: 0 }}>
                      {[
                        " ██╗   ██╗██╗",
                        " ██║   ██║██║",
                        " ██║   ██║██║",
                        " ██║   ██║██║",
                        " ╚██████╔╝██║",
                        "  ╚═════╝ ╚═╝",
                      ].map((line, i) => (
                        <div key={`u${i}`} style={{ color: t.secondary }}>{line}</div>
                      ))}
                    </div>
                    {/* "s" as regular text, quarter height of block letters */}
                    <span
                      className="font-['Fira_Code',monospace] font-bold self-end"
                      style={{
                        color: t.secondary,
                        fontSize: "clamp(14px, 2.3vw, 26px)",
                        lineHeight: 1,
                        paddingBottom: "clamp(1px, 0.15vw, 2px)",
                      }}
                    >
                      s
                    </span>
                    {/* "in" as regular text, larger */}
                    <span
                      className="font-['Fira_Code',monospace] font-light"
                      style={{
                        color: t.textDim,
                        fontSize: "clamp(20px, 3.5vw, 40px)",
                        lineHeight: 1,
                        padding: "0 clamp(8px, 1.5vw, 18px)",
                        paddingBottom: "clamp(0px, 0.1vw, 1px)",
                      }}
                    >
                      in
                    </span>
                    {/* Go block letters */}
                    <div className="whitespace-pre leading-[1.15] overflow-hidden" style={{ fontSize: "clamp(7px, 1.15vw, 13px)", letterSpacing: 0 }}>
                      {[
                        " ██████╗  ██████╗",
                        "██╔════╝ ██╔═══██╗",
                        "██║  ███╗██║   ██║",
                        "██║   ██║██║   ██║",
                        "╚██████╔╝╚██████╔╝",
                        " ╚═════╝  ╚═════╝",
                      ].map((line, i) => (
                        <div key={`g${i}`} style={{ color: t.tertiary }}>{line}</div>
                      ))}
                    </div>
                  </div>

                  {/* Man page sections — tight and punchy */}
                  <div className="tl mt-[14px] text-[12px]" style={{ animationDelay: "145ms" }}>
                    <div className="font-bold tracking-[0.04em]" style={{ color: t.heading }}>NAME</div>
                    <div className="pl-5 mt-1 leading-[1.7]" style={{ color: t.textMuted }}>
                      <span className="font-semibold" style={{ color: t.heading }}>go-tui</span>
                      {" "}&mdash; declarative terminal UIs in{" "}
                      <span
                        className="inline-block text-[10px] px-2 py-[2px] rounded-[3px] font-bold"
                        style={{
                          color: t.tertiary,
                          background: theme === "dark" ? "rgba(249,38,114,0.1)" : "rgba(212,37,104,0.08)",
                          border: `1px solid ${theme === "dark" ? "rgba(249,38,114,0.25)" : "rgba(212,37,104,0.2)"}`,
                        }}
                      >
                        Go
                      </span>
                    </div>
                  </div>

                  <div className="tl mt-[14px] text-[12px]" style={{ animationDelay: "170ms" }}>
                    <div className="font-bold tracking-[0.04em]" style={{ color: t.heading }}>SYNOPSIS</div>
                    <div className="pl-5 mt-1 leading-[1.7] whitespace-pre" style={{ color: t.textMuted }}>
                      <span style={{ color: t.secondary }}>$</span> go get github.com/grindlemire/go-tui{"\n"}
                      <span style={{ color: t.secondary }}>$</span> tui generate ./...
                    </div>
                  </div>

                  <div className="tl mt-[14px] text-[12px]" style={{ animationDelay: "200ms" }}>
                    <div className="font-bold tracking-[0.04em]" style={{ color: t.heading }}>SEE ALSO</div>
                    <div className="pl-5 mt-1 leading-[1.7]" style={{ color: t.textMuted }}>
                      {[
                        { label: "guide(7)", href: "/guide", external: false },
                        { label: "reference(3)", href: "/reference", external: false },
                        { label: "examples(7)", href: "https://github.com/grindlemire/go-tui/tree/main/examples", external: true },
                        { label: "github(1)", href: "https://github.com/grindlemire/go-tui", external: true },
                      ].map((link, i, arr) => (
                        <span key={link.label}>
                          {link.external ? (
                            <a
                              href={link.href}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="transition-all duration-150"
                              style={{
                                color: t.accent,
                                textDecoration: "underline",
                                textUnderlineOffset: "3px",
                                textDecorationColor: theme === "dark" ? "rgba(102,217,239,0.3)" : "rgba(47,158,184,0.3)",
                              }}
                              onMouseEnter={(e) => (e.currentTarget.style.textDecorationColor = t.accent)}
                              onMouseLeave={(e) => (e.currentTarget.style.textDecorationColor = theme === "dark" ? "rgba(102,217,239,0.3)" : "rgba(47,158,184,0.3)")}
                            >
                              {link.label}
                            </a>
                          ) : (
                            <Link
                              to={link.href}
                              className="transition-all duration-150"
                              style={{
                                color: t.accent,
                                textDecoration: "underline",
                                textUnderlineOffset: "3px",
                                textDecorationColor: theme === "dark" ? "rgba(102,217,239,0.3)" : "rgba(47,158,184,0.3)",
                              }}
                              onMouseEnter={(e) => (e.currentTarget.style.textDecorationColor = t.accent)}
                              onMouseLeave={(e) => (e.currentTarget.style.textDecorationColor = theme === "dark" ? "rgba(102,217,239,0.3)" : "rgba(47,158,184,0.3)")}
                            >
                              {link.label}
                            </Link>
                          )}
                          {i < arr.length - 1 && ",  "}
                        </span>
                      ))}
                    </div>
                  </div>

                  {/* Static prompt — desktop only */}
                  <div
                    className="tl hidden sm:flex items-center mt-8 text-[13px]"
                    style={{ animationDelay: "230ms" }}
                  >
                    <span style={{ color: t.secondary }}>$</span>
                    <span className="ml-2 relative">
                      {/* Blinking block cursor */}
                      <span
                        style={{
                          display: "inline-block",
                          width: "0.6ch",
                          height: "1.15em",
                          background: t.secondary,
                          animation: "blink 1s step-end infinite",
                          verticalAlign: "middle",
                        }}
                      />
                    </span>
                  </div>

                </div>
              </div>
            </div>

            {/* Scroll indicator */}
            <div
              className="absolute bottom-6 left-1/2 -translate-x-1/2 flex flex-col items-center gap-1"
              style={{ animation: "scrollBounce 2s ease-in-out infinite" }}
            >
              <span className="font-['Fira_Code',monospace] text-[10px] tracking-[0.15em] uppercase" style={{ color: t.textDim }}>
                scroll<span className="hidden sm:inline"> (enter/space)</span>
              </span>
              <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke={t.textDim} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                <path d="M4 6l4 4 4-4" />
              </svg>
            </div>
          </section>

          <Divider />

          {/* How it works */}
          <section className="max-w-[1100px] mx-auto px-4 sm:px-6 py-10 sm:py-12">
            <div
              className="font-['Fira_Code',monospace] text-[10px] tracking-[0.2em] uppercase mb-3"
              style={{ color: t.accentDim }}
            >
              what is it
            </div>
            <h2
              className="text-2xl sm:text-3xl font-bold tracking-tight mb-3"
              style={{ color: t.heading }}
            >
              Go meets declarative templates
            </h2>
            <p
              className="text-[14px] sm:text-[15px] mb-8 sm:mb-10 max-w-[600px]"
              style={{ color: t.textMuted }}
            >
              Inspired by{" "}
              <a
                href="https://templ.guide"
                target="_blank"
                rel="noopener noreferrer"
                style={{ color: t.accent, textDecoration: "underline", textUnderlineOffset: "3px" }}
              >
                templ
              </a>
              , but built for the terminal with reactive state management.
              Define state and event handlers as Go methods, write the UI in
              a <code style={{ color: t.accent }}>.gsx</code> template.
              The compiler does the rest.
            </p>
            <CodeShowcase />
          </section>

          <Divider />

          {/* Developer Experience */}
          <section className="max-w-[1100px] mx-auto px-4 sm:px-6 py-10 sm:py-12">
            <div className="flex items-center gap-3 mb-3">
              <div
                className="font-['Fira_Code',monospace] text-[10px] tracking-[0.2em] uppercase"
                style={{ color: t.tertiaryDim }}
              >
                developer experience
              </div>
              <div
                className="h-px flex-1"
                style={{
                  background: theme === "dark"
                    ? "linear-gradient(to right, #f9267218, transparent)"
                    : `linear-gradient(to right, ${t.border}, transparent)`,
                }}
              />
            </div>
            <h2
              className="text-2xl sm:text-3xl font-bold tracking-tight mb-3"
              style={{ color: t.heading }}
            >
              Editor tooling included
            </h2>
            <p
              className="text-[14px] sm:text-[15px] mb-8 sm:mb-10 max-w-[600px]"
              style={{ color: t.textMuted }}
            >
              LSP, tree-sitter grammar, and formatter included.
              Completions, diagnostics, formatting, and go-to-definition for .gsx files.
            </p>

            <div className="grid lg:grid-cols-[1fr_340px] gap-6 sm:gap-8 items-stretch">
              {/* Editor simulation */}
              <EditorSimulation
                activeFeature={dxFeature}
                onSetFeature={(i) => { setDxFeature(i); dxPausedRef.current = true; }}
                pausedRef={dxPausedRef}
              />

              {/* Capabilities list — stretches to match editor height */}
              <div className="flex flex-col justify-between">
                {([
                  { title: "Syntax highlighting", description: "Tree-sitter grammar with distinct tokens for keywords, elements, Go, and Tailwind classes.", color: t.accent, editorIdx: 0 },
                  { title: "Code completions", description: "Component suggestions with type signatures as you type.", color: t.secondary, editorIdx: 1 },
                  { title: "Inline diagnostics", description: "Errors surface in your editor before you compile.", color: t.tertiary, editorIdx: 2 },
                  { title: "Go-to-definition", description: "Jump to definitions across .gsx and Go files.", color: theme === "dark" ? "#e6db74" : "#998a00", editorIdx: 3 },
                  { title: "Auto-formatting", description: "Indentation, alignment, and imports. On save or via CLI.", color: theme === "dark" ? "#ae81ff" : "#7c5cb8", editorIdx: 4 },
                ] as const).map((cap, i) => (
                  <DxCapability
                    key={cap.title}
                    title={cap.title}
                    description={cap.description}
                    color={cap.color}
                    delay={i * 60}
                    active={dxFeature === cap.editorIdx}
                    onHover={() => {
                      dxPausedRef.current = true;
                      setDxFeature(cap.editorIdx);
                    }}
                    onLeave={() => {
                      dxPausedRef.current = false;
                    }}
                  />
                ))}
              </div>
            </div>
          </section>

          <Divider />

          {/* Comparison */}
          <ComparisonSection />

          <Divider />

          {/* Tailwind Preview */}
          <section className="max-w-[1100px] mx-auto px-4 sm:px-6 py-10 sm:py-12">
            <div
              className="font-['Fira_Code',monospace] text-[10px] tracking-[0.2em] uppercase mb-3"
              style={{ color: t.secondaryDim }}
            >
              styling
            </div>
            <h2
              className="text-2xl sm:text-3xl font-bold tracking-tight mb-3"
              style={{ color: t.heading }}
            >
              Tailwind-style classes
            </h2>
            <p
              className="text-[14px] sm:text-[15px] mb-8 sm:mb-10 max-w-[560px]"
              style={{ color: t.textMuted }}
            >
              Utility classes that compile to Go options.
            </p>

            <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-x-8 gap-y-0">
              {tailwindClasses.slice(0, 18).map((tc, i) => (
                <div
                  key={i}
                  className="flex items-baseline gap-3 py-2.5 transition-colors duration-150"
                  style={{ borderBottom: `1px solid ${t.border}` }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.background = t.bgTertiary;
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.background = "transparent";
                  }}
                >
                  <code
                    className="font-['Fira_Code',monospace] text-[12px] shrink-0"
                    style={{ color: t.accent }}
                  >
                    {tc.class}
                  </code>
                  <span
                    className="text-[11px] truncate"
                    style={{ color: t.textDim }}
                  >
                    {tc.description}
                  </span>
                </div>
              ))}
            </div>
            <Link
              to="/reference"
              className="font-['Fira_Code',monospace] inline-flex items-center gap-2 mt-6 sm:mt-8 text-sm transition-colors duration-200"
              style={{ color: t.textMuted }}
              onMouseEnter={(e) => (e.currentTarget.style.color = t.accent)}
              onMouseLeave={(e) => (e.currentTarget.style.color = t.textMuted)}
            >
              view all &rarr;
            </Link>
          </section>
        </div>
      </div>
    </Page>
  );
}

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
          <Route path="/legacy" element={<HomePage />} />
          <Route path="/guide/:slug/raw" element={<RawGuidePage />} />
        </Routes>
      </SearchContext.Provider>
    </ThemeContext.Provider>
  );
}
