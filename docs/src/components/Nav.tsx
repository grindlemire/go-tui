import { useState, useEffect } from "react";
import { Link, useLocation } from "react-router-dom";
import { palette, useTheme } from "../lib/theme.ts";
import { VERSION } from "../version.ts";
import { useSearch } from "../lib/search-context.ts";

export default function Nav() {
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
