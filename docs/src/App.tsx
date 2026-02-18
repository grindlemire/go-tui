import { useState, useEffect, useRef, useCallback, createContext, useContext } from "react";
import { Routes, Route, Link, useLocation, useParams, useNavigate, Navigate } from "react-router-dom";
import { type Theme, palette, ThemeContext, useTheme } from "./lib/theme.ts";
import { tailwindClasses } from "./content/projectInfo.ts";
import { loadGuide, loadReference } from "./lib/markdown.ts";
import Markdown from "./components/Markdown.tsx";
import TableOfContents from "./components/TableOfContents.tsx";
import CodeShowcase from "./components/CodeShowcase.tsx";
import SearchModal from "./components/SearchModal.tsx";

const SearchContext = createContext<{ openSearch: () => void }>({ openSearch: () => {} });
function useSearch() { return useContext(SearchContext); }

/* ─── Scroll to top on route change ─── */

function ScrollToTop() {
  const { pathname } = useLocation();
  useEffect(() => {
    if (pathname === "/") {
      history.scrollRestoration = "manual";
      window.scrollTo(0, 0);
    } else {
      history.scrollRestoration = "auto";
    }
  }, [pathname]);
  return null;
}

/* ─── Global Styles ─── */

function GlobalStyles() {
  return (
    <style>{`
      @import url('https://fonts.googleapis.com/css2?family=Fira+Code:wght@300;400;500;600;700&family=IBM+Plex+Sans:wght@300;400;500;600;700&display=swap');

      @keyframes blink {
        0%, 100% { opacity: 1; }
        50% { opacity: 0; }
      }

      @keyframes fadeInUp {
        from { opacity: 0; transform: translateY(10px); }
        to { opacity: 1; transform: translateY(0); }
      }
      .fade-in {
        animation: fadeInUp 0.35s ease-out forwards;
      }

      .custom-scroll::-webkit-scrollbar {
        width: 6px;
        height: 6px;
      }
      @keyframes scanDrift {
        from { transform: translateY(0); }
        to { transform: translateY(-80px); }
      }

      .custom-scroll::-webkit-scrollbar-track {
        background: transparent;
      }
      .custom-scroll::-webkit-scrollbar-thumb {
        background: #49483e;
        border-radius: 3px;
      }
      .custom-scroll::-webkit-scrollbar-thumb:hover {
        background: #75715e;
      }

      .neon-select ::selection {
        background: #66d9ef33;
        color: #66d9ef;
      }
      .light-theme .neon-select ::selection {
        background: #2f9eb833;
        color: #2f9eb8;
      }

      @keyframes syntaxPulse {
        0%, 100% { filter: brightness(1); }
        50% { filter: brightness(1.4); }
      }
      .syntax-active-token {
        animation: syntaxPulse 2s ease-in-out infinite;
      }

      @keyframes comparisonReveal {
        from { opacity: 0; transform: translateY(6px); }
        to { opacity: 1; transform: translateY(0); }
      }
      .comparison-row-animate {
        animation: comparisonReveal 0.3s ease-out both;
      }

      @keyframes cellReveal {
        from { opacity: 0; }
        to { opacity: 1; }
      }

      @keyframes lineIn {
        from { opacity: 0; transform: translateY(4px); }
        to   { opacity: 1; transform: translateY(0); }
      }

      .tl {
        opacity: 0;
        animation: lineIn 0.1s ease-out forwards;
      }

      @keyframes scrollBounce {
        0%, 100% { transform: translateY(0); opacity: 0.6; }
        50% { transform: translateY(6px); opacity: 1; }
      }

      @keyframes navSlideDown {
        from { transform: translateY(-100%); }
        to { transform: translateY(0); }
      }

      *, *::before, *::after {
        transition: color 0.3s ease, background-color 0.3s ease, border-color 0.3s ease, box-shadow 0.3s ease, fill 0.3s ease, stroke 0.3s ease;
      }
      /* Don't let theme transition interfere with existing animations */
      .tl, [style*="animation"] {
        transition: color 0.3s ease, background-color 0.3s ease, border-color 0.3s ease, fill 0.3s ease, stroke 0.3s ease;
      }
    `}</style>
  );
}

/* ─── Page Background (scan lines + glow) ─── */

function PageBackground({ theme }: { theme: Theme }) {
  const isDark = theme === "dark";
  const lineAlpha = isDark ? "0.025" : "0.018";
  const lineRgb = isDark ? "248,248,242" : "39,40,34";
  const glowColor = isDark
    ? "rgba(166,226,46,0.03)"
    : "rgba(212,37,104,0.02)";

  return (
    <div
      className="absolute inset-0 overflow-hidden pointer-events-none"
      aria-hidden="true"
    >
      {/* Scan lines */}
      <div
        className="absolute inset-0"
        style={{
          top: "-80px",
          bottom: "0",
          backgroundImage: `repeating-linear-gradient(0deg, transparent, transparent 3px, rgba(${lineRgb},${lineAlpha}) 3px, rgba(${lineRgb},${lineAlpha}) 4px)`,
          animation: "scanDrift 12s linear infinite",
          willChange: "transform",
        }}
      />
      {/* Warm radial glow */}
      <div
        className="absolute inset-0"
        style={{
          background: `radial-gradient(ellipse at 25% 40%, ${glowColor} 0%, transparent 55%)`,
        }}
      />
    </div>
  );
}

/* ─── Nav ─── */

function Nav({ hideUntilScroll = false }: { hideUntilScroll?: boolean }) {
  const { theme, setTheme } = useTheme();
  const { openSearch: onOpenSearch } = useSearch();
  const t = palette[theme];
  const location = useLocation();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [navVisible, setNavVisible] = useState(!hideUntilScroll);

  // Close mobile menu on navigation
  useEffect(() => {
    setMobileOpen(false);
  }, [location.pathname]);

  // Show nav after scrolling past hero
  useEffect(() => {
    if (!hideUntilScroll) { setNavVisible(true); return; }
    const onScroll = () => {
      setNavVisible(window.scrollY > window.innerHeight * 0.4);
    };
    window.addEventListener("scroll", onScroll, { passive: true });
    onScroll();
    return () => window.removeEventListener("scroll", onScroll);
  }, [hideUntilScroll]);

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
      className={`${hideUntilScroll ? "fixed" : "sticky"} top-0 left-0 right-0 z-40 backdrop-blur-md`}
      style={{
        background:
          theme === "dark"
            ? "rgba(39, 40, 34, 0.92)"
            : "rgba(250, 250, 248, 0.92)",
        borderBottom: `1px solid ${t.border}`,
        transform: navVisible ? "translateY(0)" : "translateY(-100%)",
        transition: "transform 0.3s ease-out",
        pointerEvents: navVisible ? "auto" : "none",
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
            className="mx-2"
            style={{ width: 1, height: 20, background: t.border }}
          />

          <a
            href="https://pkg.go.dev/github.com/grindlemire/go-tui"
            target="_blank"
            rel="noopener noreferrer"
            className="font-['Fira_Code',monospace] text-[10px] px-2 py-1 rounded transition-all duration-200"
            style={{
              color: t.secondary,
              background: `${t.secondary}0a`,
              border: `1px solid ${t.secondary}22`,
            }}
            title="v0.1.0 — view on pkg.go.dev"
            onMouseEnter={(e) => {
              e.currentTarget.style.borderColor = `${t.secondary}55`;
              e.currentTarget.style.background = `${t.secondary}14`;
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.borderColor = `${t.secondary}22`;
              e.currentTarget.style.background = `${t.secondary}0a`;
            }}
          >
            v0.1.0
          </a>

          <a
            href="https://github.com/grindlemire/go-tui"
            target="_blank"
            rel="noopener noreferrer"
            className="p-1.5 rounded transition-all duration-200 flex items-center"
            style={{
              color: t.textMuted,
              border: `1px solid transparent`,
            }}
            title="View on GitHub — open source"
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
            className="p-1.5 rounded flex items-center"
            style={{ color: t.textMuted }}
            title="View on GitHub"
          >
            <svg width="15" height="15" viewBox="0 0 16 16" fill="currentColor" aria-label="GitHub">
              <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z" />
            </svg>
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
              v0.1.0
            </span>
            pkg.go.dev
          </a>
        </div>
      )}
    </nav>
  );
}

/* ─── Editor Simulation (DX Section) ─── */

function EditorSimulation({
  activeFeature,
  onSetFeature,
  pausedRef,
}: {
  activeFeature: number;
  onSetFeature: (i: number) => void;
  pausedRef: React.RefObject<boolean>;
}) {
  const { theme } = useTheme();
  const t = palette[theme];
  const [showCompletions, setShowCompletions] = useState(false);
  const [showDiagnostic, setShowDiagnostic] = useState(false);
  const [highlightLine, setHighlightLine] = useState<number | null>(null);
  const [showFormatted, setShowFormatted] = useState(false);

  const fmtColor = theme === "dark" ? "#ae81ff" : "#7c5cb8";
  const features = [
    { id: "syntax", label: "syntax highlighting", icon: "\u2726", color: t.accent },
    { id: "completions", label: "completions", icon: "\u00bb", color: t.secondary },
    { id: "diagnostics", label: "diagnostics", icon: "\u26a0", color: t.tertiary },
    { id: "goto", label: "go-to-definition", icon: "\u2192", color: theme === "dark" ? "#e6db74" : "#998a00" },
    { id: "format", label: "auto-format", icon: "\u2261", color: fmtColor },
  ];

  useEffect(() => {
    setShowCompletions(false);
    setShowDiagnostic(false);
    setHighlightLine(null);
    setShowFormatted(false);

    const timer = setTimeout(() => {
      if (activeFeature === 1) setShowCompletions(true);
      if (activeFeature === 2) setShowDiagnostic(true);
      if (activeFeature === 3) setHighlightLine(10);
      if (activeFeature === 4) setShowFormatted(true);
    }, 300);

    return () => clearTimeout(timer);
  }, [activeFeature]);

  const editorLines = [
    { num: 1, tokens: [{ text: "package", color: t.codeKeyword }, { text: " dashboard", color: t.text }] },
    { num: 2, tokens: [{ text: "", color: t.text }] },
    { num: 3, tokens: [{ text: "import", color: t.codeKeyword }, { text: " (", color: t.codePunct }] },
    { num: 4, tokens: [{ text: '  "fmt"', color: t.codeString }] },
    { num: 5, tokens: [{ text: ")", color: t.codePunct }] },
    { num: 6, tokens: [{ text: "", color: t.text }] },
    { num: 7, tokens: [{ text: "templ", color: t.codeKeyword }, { text: " ", color: t.text }, { text: "Dashboard", color: t.codeFunc }, { text: "(", color: t.codePunct }, { text: "title ", color: t.text }, { text: "string", color: t.codeKeyword }, { text: ") {", color: t.codePunct }] },
    { num: 8, tokens: [{ text: '  <', color: t.codePunct }, { text: 'div', color: t.codeKeyword }, { text: ' class=', color: t.codePunct }, { text: '"flex-col h-full"', color: t.codeString }, { text: '>', color: t.codePunct }] },
    { num: 9, tokens: [{ text: "    @", color: t.codeDirective }, { text: "Header", color: t.codeFunc }, { text: "(title)", color: t.codePunct }] },
    { num: 10, tokens: [{ text: "    @", color: t.codeDirective }, { text: "Sidebar", color: t.codeFunc }, { text: "()", color: t.codePunct }] },
    { num: 11, tokens: [{ text: "    @", color: t.codeDirective }, { text: "MainContent", color: t.codeFunc }, { text: "()", color: t.codePunct }] },
    { num: 12, tokens: [{ text: "  </", color: t.codePunct }, { text: "div", color: t.codeKeyword }, { text: ">", color: t.codePunct }] },
    { num: 13, tokens: [{ text: "}", color: t.codePunct }] },
    { num: 14, tokens: [{ text: "", color: t.text }] },
    { num: 15, tokens: [{ text: "templ", color: t.codeKeyword }, { text: " ", color: t.text }, { text: "Header", color: t.codeFunc }, { text: "(", color: t.codePunct }, { text: "title ", color: t.text }, { text: "string", color: t.codeKeyword }, { text: ") {", color: t.codePunct }] },
    { num: 16, tokens: [{ text: '  <', color: t.codePunct }, { text: 'div', color: t.codeKeyword }, { text: ' class=', color: t.codePunct }, { text: '"border-single p-1"', color: t.codeString }, { text: '>', color: t.codePunct }] },
    { num: 17, tokens: [{ text: '    <', color: t.codePunct }, { text: 'span', color: t.codeKeyword }, { text: ' class=', color: t.codePunct }, { text: '"font-bold text-cyan"', color: t.codeString }, { text: '>', color: t.codePunct }] },
    { num: 18, tokens: [{ text: "      {", color: t.codePunct }, { text: "fmt", color: t.text }, { text: ".", color: t.codePunct }, { text: "Sprintf", color: t.codeFunc }, { text: "(", color: t.codePunct }, { text: '"%s"', color: t.codeString }, { text: ", title)", color: t.codePunct }, { text: "}", color: t.codePunct }] },
    { num: 19, tokens: [{ text: "    </", color: t.codePunct }, { text: "span", color: t.codeKeyword }, { text: ">", color: t.codePunct }] },
    { num: 20, tokens: [{ text: "  </", color: t.codePunct }, { text: "div", color: t.codeKeyword }, { text: ">", color: t.codePunct }] },
    { num: 21, tokens: [{ text: "}", color: t.codePunct }] },
  ];

  // Lines with bad indentation (for the format demo)
  // Only certain lines differ — we track which line nums changed
  const fmtChangedLines = new Set([9, 10, 11, 17, 18, 19]);
  const messyLines: typeof editorLines = editorLines.map((line) => {
    if (line.num === 9) return { num: 9, tokens: [{ text: "  @", color: t.codeDirective }, { text: "Header", color: t.codeFunc }, { text: "(title)", color: t.codePunct }] };
    if (line.num === 10) return { num: 10, tokens: [{ text: "      @", color: t.codeDirective }, { text: "Sidebar", color: t.codeFunc }, { text: "()", color: t.codePunct }] };
    if (line.num === 11) return { num: 11, tokens: [{ text: "   @", color: t.codeDirective }, { text: "MainContent", color: t.codeFunc }, { text: "()", color: t.codePunct }] };
    if (line.num === 17) return { num: 17, tokens: [{ text: "  <", color: t.codePunct }, { text: "span", color: t.codeKeyword }, { text: " class=", color: t.codePunct }, { text: '"font-bold text-cyan"', color: t.codeString }, { text: ">", color: t.codePunct }] };
    if (line.num === 18) return { num: 18, tokens: [{ text: "        {", color: t.codePunct }, { text: "fmt", color: t.text }, { text: ".", color: t.codePunct }, { text: "Sprintf", color: t.codeFunc }, { text: "(", color: t.codePunct }, { text: '"%s"', color: t.codeString }, { text: ", title)", color: t.codePunct }, { text: "}", color: t.codePunct }] };
    if (line.num === 19) return { num: 19, tokens: [{ text: "      </", color: t.codePunct }, { text: "span", color: t.codeKeyword }, { text: ">", color: t.codePunct }] };
    return line;
  });

  const completionItems = [
    { label: "Sidebar", detail: "() *Element" },
    { label: "SearchBar", detail: "(query string) *Element" },
    { label: "StatusLine", detail: "() *Element" },
  ];

  return (
    <div
      className="rounded-lg overflow-hidden"
      style={{
        background: t.bgCode,
        border: `1px solid ${t.border}`,
        boxShadow: theme === "dark"
          ? "0 4px 24px rgba(0,0,0,0.4)"
          : "0 2px 12px rgba(0,0,0,0.08)",
      }}
      onMouseEnter={() => { pausedRef.current = true; }}
      onMouseLeave={() => { pausedRef.current = false; }}
    >
      {/* Editor title bar */}
      <div
        className="flex items-center justify-between px-4 py-2"
        style={{ borderBottom: `1px solid ${t.border}` }}
      >
        <div className="flex items-center gap-2">
          <div className="flex gap-1.5">
            <div className="w-2.5 h-2.5 rounded-full" style={{ background: "#ff5f57" }} />
            <div className="w-2.5 h-2.5 rounded-full" style={{ background: "#febc2e" }} />
            <div className="w-2.5 h-2.5 rounded-full" style={{ background: "#28c840" }} />
          </div>
          <span className="font-['Fira_Code',monospace] text-[10px] ml-2" style={{ color: t.textDim }}>
            dashboard.gsx
          </span>
        </div>
        <div className="flex items-center gap-1.5">
          <span
            className="font-['Fira_Code',monospace] text-[9px] px-1.5 py-0.5 rounded"
            style={{
              color: t.secondary,
              background: `${t.secondary}12`,
              border: `1px solid ${t.secondary}25`,
            }}
          >
            LSP
          </span>
          <span
            className="font-['Fira_Code',monospace] text-[9px] px-1.5 py-0.5 rounded"
            style={{
              color: t.accent,
              background: `${t.accent}12`,
              border: `1px solid ${t.accent}25`,
            }}
          >
            tree-sitter
          </span>
        </div>
      </div>

      {/* Feature selector tabs */}
      <div
        className="flex items-center gap-1 px-3 py-1.5 overflow-x-auto custom-scroll"
        style={{
          borderBottom: `1px solid ${t.border}`,
          background: theme === "dark" ? "#1e1f1a" : "#eeeee8",
        }}
      >
        {features.map((f, i) => (
          <button
            key={f.id}
            onClick={() => onSetFeature(i)}
            className="font-['Fira_Code',monospace] text-[10px] sm:text-[11px] px-2.5 py-1 rounded transition-all duration-200 flex items-center gap-1.5 whitespace-nowrap shrink-0"
            style={{
              color: activeFeature === i ? f.color : t.textDim,
              background: activeFeature === i ? `${f.color}10` : "transparent",
              border: `1px solid ${activeFeature === i ? `${f.color}30` : "transparent"}`,
              cursor: "pointer",
            }}
          >
            <span style={{ fontSize: "9px" }}>{f.icon}</span>
            {f.label}
          </button>
        ))}
      </div>

      {/* Editor body */}
      <div className="relative px-0 py-3 font-['Fira_Code',monospace] text-[11px] sm:text-[12px] leading-[1.8] overflow-x-auto custom-scroll">
        {(activeFeature === 4 && !showFormatted ? messyLines : editorLines).map((line) => {
          const isGotoTarget = activeFeature === 3 && highlightLine === line.num;
          const hasDiagnostic = activeFeature === 2 && showDiagnostic && line.num === 10;
          const isFmtChanged = activeFeature === 4 && showFormatted && fmtChangedLines.has(line.num);

          return (
            <div
              key={line.num}
              className="flex transition-all duration-300"
              style={{
                background: isGotoTarget
                  ? `${features[3].color}10`
                  : hasDiagnostic
                    ? `${t.tertiary}08`
                    : isFmtChanged
                      ? `${fmtColor}08`
                      : "transparent",
                borderLeft: isGotoTarget
                  ? `2px solid ${features[3].color}`
                  : isFmtChanged
                    ? `2px solid ${fmtColor}`
                    : "2px solid transparent",
              }}
            >
              <span
                className="inline-block w-8 sm:w-10 text-right pr-3 sm:pr-4 select-none shrink-0"
                style={{ color: t.textDim, opacity: 0.5 }}
              >
                {line.num}
              </span>
              <span className="whitespace-pre">
                {line.tokens.map((tok, j) => {
                  const isSyntaxHighlighted = activeFeature === 0
                    && tok.text.trim().length > 0
                    && tok.color !== t.text
                    && tok.color !== t.codePunct;
                  return (
                    <span
                      key={j}
                      className={isSyntaxHighlighted ? "syntax-active-token" : ""}
                      style={{
                        color: tok.color,
                        animationDelay: isSyntaxHighlighted ? `${j * 120}ms` : undefined,
                      }}
                    >{tok.text}</span>
                  );
                })}
              </span>
              {hasDiagnostic && (
                <span className="ml-4 text-[10px] flex items-center gap-1.5" style={{ color: t.tertiary }}>
                  <span className="opacity-80">undefined: Sidebar</span>
                </span>
              )}
            </div>
          );
        })}

        {/* Completions popup */}
        {activeFeature === 1 && showCompletions && (
          <div
            className="absolute rounded-md overflow-hidden"
            style={{
              top: "calc(1.8em * 10 + 12px)",
              left: "calc(10px + 8ch)",
              background: theme === "dark" ? "#3e3d32" : "#ffffff",
              border: `1px solid ${t.border}`,
              boxShadow: theme === "dark"
                ? "0 4px 16px rgba(0,0,0,0.5)"
                : "0 4px 16px rgba(0,0,0,0.12)",
              zIndex: 10,
              animation: "fadeInUp 0.2s ease-out forwards",
              minWidth: "220px",
            }}
          >
            {completionItems.map((item, i) => (
              <div
                key={i}
                className="flex items-center gap-2 px-2.5 py-1.5 text-[11px]"
                style={{
                  background: i === 0 ? `${t.accent}15` : "transparent",
                  borderLeft: i === 0 ? `2px solid ${t.accent}` : "2px solid transparent",
                }}
              >
                <span
                  className="px-1 py-0.5 rounded text-[9px]"
                  style={{ background: `${t.secondary}18`, color: t.secondary }}
                >
                  C
                </span>
                <span style={{ color: i === 0 ? t.accent : t.text }}>{item.label}</span>
                <span className="ml-auto" style={{ color: t.textDim, fontSize: "10px" }}>
                  {item.detail}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Status bar */}
      <div
        className="flex items-center justify-between px-4 py-1.5 font-['Fira_Code',monospace] text-[10px]"
        style={{
          borderTop: `1px solid ${t.border}`,
          background: theme === "dark" ? "#1e1f1a" : "#eeeee8",
        }}
      >
        <div className="flex items-center gap-3">
          <span style={{ color: t.textDim }}>Ln 10, Col 5</span>
          <span style={{ color: t.textDim }}>GSX</span>
        </div>
        <div className="flex items-center gap-2">
          {activeFeature === 2 && showDiagnostic && (
            <span className="flex items-center gap-1" style={{ color: t.tertiary }}>
              <svg width="10" height="10" viewBox="0 0 10 10" fill="currentColor">
                <circle cx="5" cy="5" r="4" />
              </svg>
              1 error
            </span>
          )}
          {activeFeature === 4 && showFormatted && (
            <span className="flex items-center gap-1" style={{ color: fmtColor }}>
              6 lines formatted
            </span>
          )}
          <span style={{ color: t.secondary }}>tui lsp</span>
        </div>
      </div>
    </div>
  );
}

/* ─── DX Capability Row ─── */

function DxCapability({
  title,
  description,
  color,
  delay,
  active,
  onHover,
  onLeave,
}: {
  title: string;
  description: string;
  color: string;
  delay: number;
  active?: boolean;
  onHover?: () => void;
  onLeave?: () => void;
}) {
  const { theme } = useTheme();
  const t = palette[theme];
  const highlighted = active ?? false;

  return (
    <div
      className="py-3 px-4 rounded-lg transition-all duration-200 cursor-default"
      style={{
        background: highlighted ? `${color}06` : "transparent",
        borderLeft: `2px solid ${highlighted ? color : "transparent"}`,
        animation: `fadeInUp 0.4s ease-out ${delay}ms both`,
      }}
      onMouseEnter={onHover}
      onMouseLeave={onLeave}
    >
      <div className="flex items-center gap-2 mb-1">
        <div
          className="w-1.5 h-1.5 rounded-full shrink-0"
          style={{ background: color }}
        />
        <div
          className="font-['Fira_Code',monospace] text-[13px] font-medium"
          style={{ color: t.heading }}
        >
          {title}
        </div>
      </div>
      <div
        className="text-[12px] sm:text-[13px] leading-relaxed pl-3.5"
        style={{ color: t.textMuted }}
      >
        {description}
      </div>
    </div>
  );
}

/* ─── Divider ─── */

function Divider() {
  const { theme } = useTheme();
  const t = palette[theme];
  return (
    <div className="max-w-[1100px] mx-auto px-4 sm:px-6 py-8">
      <div
        className="h-px"
        style={{
          background:
            theme === "dark"
              ? "linear-gradient(to right, transparent, #66d9ef18, #f9267218, transparent)"
              : `linear-gradient(to right, transparent, ${t.border}, transparent)`,
        }}
      />
    </div>
  );
}

/* ─── Footer ─── */

function Footer() {
  const { theme } = useTheme();
  const t = palette[theme];
  return (
    <footer
      className="mt-20"
      style={{
        borderTop: `1px solid ${t.border}`,
        background: t.bgSecondary,
      }}
    >
      <div className="max-w-[1100px] mx-auto px-4 sm:px-6 py-8">
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 font-['Fira_Code',monospace] text-[11px]">
          <span style={{ color: t.textDim }}>
            go-tui &mdash; reactive terminal UIs in Go
          </span>
          <a
            href="https://github.com/grindlemire/go-tui"
            target="_blank"
            rel="noopener noreferrer"
            className="transition-colors duration-200"
            style={{ color: t.textMuted }}
            onMouseEnter={(e) => (e.currentTarget.style.color = t.accent)}
            onMouseLeave={(e) => (e.currentTarget.style.color = t.textMuted)}
          >
            github.com/grindlemire/go-tui
          </a>
        </div>
      </div>
    </footer>
  );
}

/* ─── Page Wrapper ─── */

function Page({ children, hideNavUntilScroll = false }: { children: React.ReactNode; hideNavUntilScroll?: boolean }) {
  const { theme } = useTheme();
  const t = palette[theme];
  return (
    <div
      className={`${theme === "dark" ? "dark-theme" : "light-theme"} neon-select overflow-x-clip`}
      style={{
        background: t.bg,
        color: t.text,
        minHeight: "100vh",
        fontFamily: "'IBM Plex Sans', sans-serif",
      }}
    >
      <Nav hideUntilScroll={hideNavUntilScroll} />
      {children}
      <Footer />
    </div>
  );
}

/* ─── Comparison Section ─── */

type CellValue = { summary: string; detail: string };
type ComparisonFeature = {
  label: string;
  values: Record<string, CellValue>;
};

const comparisonLibraries = ["go-tui", "Bubble Tea", "tview", "gocui"] as const;

const comparisonFeatures: ComparisonFeature[] = [
  {
    label: "Approach",
    values: {
      "go-tui": {
        summary: "Declarative .gsx templates",
        detail: ".gsx files use HTML-like syntax with Tailwind-style classes and compile to type-safe Go via tui generate",
      },
      "Bubble Tea": {
        summary: "Elm architecture",
        detail: "Functional Model → Update → View cycle. State is immutable, messages drive updates, View returns a string",
      },
      tview: {
        summary: "Imperative widget toolkit",
        detail: "OOP style — create widget objects, configure via methods, compose in layout containers. Implements the Primitive interface",
      },
      gocui: {
        summary: "View manager",
        detail: "Create named rectangular views with absolute coordinates. Views implement io.ReadWriter for content",
      },
    },
  },
  {
    label: "Layout",
    values: {
      "go-tui": {
        summary: "CSS flexbox",
        detail: "Full flexbox: grow, shrink, justify, align, gap, padding, margin, min/max constraints, percentage and auto sizing",
      },
      "Bubble Tea": {
        summary: "String joins via lipgloss",
        detail: "lipgloss provides box model styling (padding, margin, borders) and JoinHorizontal/JoinVertical for composition. No flexbox — open issue since 2023",
      },
      tview: {
        summary: "Basic Flex and Grid",
        detail: "Flex supports direction and proportional sizing. Grid adds row/column spans. Neither has gap, justify-content, or align-items",
      },
      gocui: {
        summary: "Manual coordinates",
        detail: "Views positioned with absolute (x0, y0, x1, y1) coordinates in a Layout function. Responsive sizing requires manual calculation",
      },
    },
  },
  {
    label: "Widgets",
    values: {
      "go-tui": {
        summary: "HTML-style primitives",
        detail: "Built-in elements: div, span, p, ul, li, button, input, table, progress, hr, br. Composable via .gsx components",
      },
      "Bubble Tea": {
        summary: "14+ via Bubbles",
        detail: "Separate Bubbles library: text input, text area, viewport, list, table, spinner, progress, file picker, paginator, help, and more",
      },
      tview: {
        summary: "16+ built-in",
        detail: "Richest widget set: TextView, TextArea, Table, TreeView, List, Form, Modal, InputField, DropDown, Checkbox, Button, Image, and more",
      },
      gocui: {
        summary: "Views only",
        detail: "No pre-built widgets. Views provide text I/O and keybindings — widgets like tables or lists must be built from scratch",
      },
    },
  },
  {
    label: "State",
    values: {
      "go-tui": {
        summary: "Reactive State[T]",
        detail: "Generic State[T] with Bind() callbacks triggers re-render on change. Batch() coalesces multiple updates. Global dirty tracking",
      },
      "Bubble Tea": {
        summary: "Elm update cycle",
        detail: "Messages flow through Update() which returns a new model and optional commands. Predictable data flow, but requires message routing boilerplate for nested components",
      },
      tview: {
        summary: "Manual redraw",
        detail: "Mutate widget state directly via setter methods, then call app.Draw() or app.QueueUpdateDraw() for thread-safe re-rendering",
      },
      gocui: {
        summary: "Manual redraw",
        detail: "Write content to views via io.Writer. The Layout() manager function is called each iteration to reposition views",
      },
    },
  },
  {
    label: "Inline mode",
    values: {
      "go-tui": {
        summary: "Supported",
        detail: "Enabled via WithInlineHeight(). PrintAbove() outputs to scrollback above the widget",
      },
      "Bubble Tea": {
        summary: "Default mode",
        detail: "Inline is the default — fullscreen requires opting in with tea.WithAltScreen(). Supports tea.Println() for output above",
      },
      tview: {
        summary: "Fullscreen only",
        detail: "Architectural limitation — tcell takes over the entire screen. No inline rendering support",
      },
      gocui: {
        summary: "Fullscreen only",
        detail: "Same tcell limitation. The GUI manager controls the full terminal screen",
      },
    },
  },
  {
    label: "Mouse",
    values: {
      "go-tui": {
        summary: "SGR mode + ref hit testing",
        detail: "SGR extended mouse protocol. HandleClicks() provides automatic ref-based hit testing for named elements",
      },
      "Bubble Tea": {
        summary: "SGR mode + modifiers",
        detail: "SGR extended mouse with click, release, wheel, motion events. v2 splits into typed MouseClickMsg, MouseWheelMsg, etc.",
      },
      tview: {
        summary: "Click, drag, wheel",
        detail: "Opt-in via EnableMouse(). Supports click, double-click, drag, and wheel events through tcell",
      },
      gocui: {
        summary: "Click + motion",
        detail: "Mouse support in awesome-gocui fork. Click, motion, and modifier key detection",
      },
    },
  },
  {
    label: "Tooling",
    values: {
      "go-tui": {
        summary: "LSP + formatter + tree-sitter",
        detail: "Custom language server for .gsx files with completions, hover, go-to-definition, and diagnostics. Formatter and tree-sitter grammar included. Generated Go code uses standard gopls",
      },
      "Bubble Tea": {
        summary: "Standard Go (gopls)",
        detail: "All code is plain Go — full gopls support out of the box. No additional tooling needed or available",
      },
      tview: {
        summary: "Standard Go (gopls)",
        detail: "Plain Go code with full gopls support. No domain-specific tooling",
      },
      gocui: {
        summary: "Standard Go (gopls)",
        detail: "Plain Go code with full gopls support. No domain-specific tooling",
      },
    },
  },
  {
    label: "Build",
    values: {
      "go-tui": {
        summary: "tui generate + go build",
        detail: "Requires running tui generate to compile .gsx templates to Go before go build. Extra step in the build chain",
      },
      "Bubble Tea": {
        summary: "Standard go build",
        detail: "No code generation or extra build steps. Standard Go compilation",
      },
      tview: {
        summary: "Standard go build",
        detail: "No code generation or extra build steps. Standard Go compilation",
      },
      gocui: {
        summary: "Standard go build",
        detail: "No code generation or extra build steps. Standard Go compilation",
      },
    },
  },
];

/* ─── Comparison Row Detail Panel ─── */

function ComparisonDetailPanel({
  feature,
  expanded,
  libColors,
}: {
  feature: ComparisonFeature;
  expanded: boolean;
  libColors: Record<string, string>;
}) {
  const { theme } = useTheme();
  const t = palette[theme];
  const panelRef = useRef<HTMLDivElement>(null);
  const [measuredHeight, setMeasuredHeight] = useState(0);

  useEffect(() => {
    if (!panelRef.current) return;
    const ro = new ResizeObserver((entries) => {
      for (const entry of entries) {
        setMeasuredHeight(entry.contentRect.height);
      }
    });
    ro.observe(panelRef.current);
    return () => ro.disconnect();
  }, []);

  return (
    <div
      className="overflow-hidden"
      style={{
        maxHeight: expanded ? measuredHeight + 1 : 0,
        opacity: expanded ? 1 : 0,
        transition: "max-height 0.28s cubic-bezier(0.4, 0, 0.2, 1), opacity 0.22s ease",
      }}
    >
      <div ref={panelRef}>
        <div
          className="grid gap-0"
          style={{
            gridTemplateColumns: "140px repeat(4, 1fr)",
            borderTop: `1px solid ${theme === "dark" ? "rgba(255,255,255,0.04)" : "rgba(0,0,0,0.04)"}`,
            background: theme === "dark" ? "rgba(0,0,0,0.2)" : "rgba(0,0,0,0.02)",
          }}
        >
          {/* Label cell — empty spacer */}
          <div className="px-4 py-3" />
          {comparisonLibraries.map((lib, colIdx) => {
            const isGoTui = colIdx === 0;
            const color = libColors[lib];
            return (
              <div
                key={lib}
                className="px-4 py-3"
                style={{
                  borderLeft: `1px solid ${isGoTui ? `${t.accent}20` : `${t.border}60`}`,
                }}
              >
                <div className="flex items-center gap-1.5 mb-1.5">
                  <div
                    className="w-[5px] h-[5px] rounded-full shrink-0"
                    style={{ background: color, opacity: 0.7 }}
                  />
                  <div
                    className="font-['Fira_Code',monospace] text-[9px] tracking-[0.1em] uppercase"
                    style={{ color: t.textDim }}
                  >
                    {lib}
                  </div>
                </div>
                <div
                  className="font-['IBM_Plex_Sans',sans-serif] text-[11.5px] leading-[1.55]"
                  style={{ color: t.textMuted }}
                >
                  {feature.values[lib].detail}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}

function ComparisonSection() {
  const { theme } = useTheme();
  const t = palette[theme];
  const [hoveredRow, setHoveredRow] = useState<number | null>(null);
  const [expandedRow, setExpandedRow] = useState<number | null>(null);
  const [visible, setVisible] = useState(false);
  const sectionRef = useRef<HTMLDivElement>(null);

  const goTuiColIdx = 0;
  const accentTint = theme === "dark" ? `${t.accent}14` : `${t.accent}0c`;

  const libColors: Record<string, string> = {
    "go-tui": t.accent,
    "Bubble Tea": t.secondary,
    tview: theme === "dark" ? "#e6db74" : "#998a00",
    gocui: t.tertiary,
  };

  useEffect(() => {
    const el = sectionRef.current;
    if (!el) return;
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setVisible(true);
          observer.disconnect();
        }
      },
      { threshold: 0.1 }
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, []);

  return (
    <section
      ref={sectionRef}
      className="max-w-[1100px] mx-auto px-4 sm:px-6 py-10 sm:py-12"
    >
      <div
        className="font-['Fira_Code',monospace] text-[10px] tracking-[0.2em] uppercase mb-3"
        style={{ color: t.accentDim }}
      >
        landscape
      </div>
      <h2
        className="text-2xl sm:text-3xl font-bold tracking-tight mb-3"
        style={{ color: t.heading }}
      >
        Go TUI libraries
      </h2>
      <p
        className="text-[14px] sm:text-[15px] mb-8 sm:mb-10 max-w-[640px]"
        style={{ color: t.textMuted }}
      >
        Different trade-offs, side by side.
        Pure Go, zero CGO.{" "}
        <span
          className="font-['Fira_Code',monospace] text-[11px]"
          style={{ color: t.textDim }}
        >
          Click a row to expand.
        </span>
      </p>

      {/* Desktop table */}
      <div className="hidden lg:block overflow-x-auto custom-scroll">
        <div
          className="rounded-lg overflow-hidden"
          style={{
            border: `1px solid ${t.border}`,
            background: t.bgCard,
            boxShadow:
              theme === "dark"
                ? "0 2px 16px rgba(0,0,0,0.5)"
                : "0 1px 6px rgba(0,0,0,0.07)",
          }}
        >
          {/* Header */}
          <div
            className="grid items-end gap-0"
            style={{
              gridTemplateColumns: "140px repeat(4, 1fr)",
              borderBottom: `1px solid ${t.border}`,
              background: theme === "dark" ? "#23241e" : "#f5f5f1",
            }}
          >
            <div
              className="px-4 py-4 font-['Fira_Code',monospace] text-[10px] tracking-[0.15em] uppercase"
              style={{ color: t.textDim }}
            />
            {comparisonLibraries.map((lib, colIdx) => {
              const isGoTui = colIdx === goTuiColIdx;
              const color = libColors[lib];
              return (
                <div
                  key={lib}
                  className="px-4 py-4 text-center"
                  style={{
                    background: isGoTui
                      ? theme === "dark" ? `${t.accent}1a` : `${t.accent}10`
                      : "transparent",
                    borderLeft: `1px solid ${t.border}`,
                  }}
                >
                  <div
                    className="font-['Fira_Code',monospace] text-[12px] font-semibold"
                    style={{ color: isGoTui ? t.accent : t.heading }}
                  >
                    {lib}
                  </div>
                  <div
                    className="mt-1.5 mx-auto h-[2px] rounded-full"
                    style={{
                      width: isGoTui ? "60%" : "0%",
                      background: color,
                      opacity: isGoTui ? 1 : 0,
                      transition: "width 0.3s ease, opacity 0.3s ease",
                    }}
                  />
                </div>
              );
            })}
          </div>

          {/* Rows */}
          {comparisonFeatures.map((feature, rowIdx) => {
            const isExpanded = expandedRow === rowIdx;
            const isHovered = hoveredRow === rowIdx;
            const isEvenRow = rowIdx % 2 === 0;
            const stripeBg = isEvenRow
              ? "transparent"
              : theme === "dark" ? "rgba(255,255,255,0.012)" : "rgba(0,0,0,0.012)";

            return (
              <div
                key={feature.label}
                className={visible ? "comparison-row-animate" : "opacity-0"}
                style={{
                  borderBottom:
                    rowIdx < comparisonFeatures.length - 1
                      ? `1px solid ${t.border}`
                      : "none",
                  animationDelay: visible ? `${rowIdx * 50}ms` : "0ms",
                }}
              >
                {/* Summary row — fixed height, no reflow */}
                <div
                  className="grid items-stretch gap-0 cursor-pointer select-none"
                  style={{
                    gridTemplateColumns: "140px repeat(4, 1fr)",
                    background: isExpanded
                      ? theme === "dark" ? "rgba(255,255,255,0.03)" : "rgba(0,0,0,0.02)"
                      : isHovered
                        ? theme === "dark" ? "rgba(255,255,255,0.02)" : "rgba(0,0,0,0.015)"
                        : stripeBg,
                    transition: "background 0.15s ease",
                  }}
                  onClick={() =>
                    setExpandedRow(isExpanded ? null : rowIdx)
                  }
                  onMouseEnter={() => setHoveredRow(rowIdx)}
                  onMouseLeave={() => setHoveredRow(null)}
                >
                  <div className="px-4 py-3.5 flex items-center gap-2">
                    <svg
                      width="8"
                      height="8"
                      viewBox="0 0 8 8"
                      className="shrink-0"
                      style={{
                        transform: isExpanded ? "rotate(90deg)" : "rotate(0deg)",
                        transition: "transform 0.2s cubic-bezier(0.4, 0, 0.2, 1)",
                        opacity: isHovered || isExpanded ? 0.8 : 0.3,
                      }}
                    >
                      <path
                        d="M2 1L6 4L2 7"
                        fill="none"
                        stroke={t.textMuted}
                        strokeWidth="1.5"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      />
                    </svg>
                    <div
                      className="font-['Fira_Code',monospace] text-[11px] font-semibold"
                      style={{ color: t.text }}
                    >
                      {feature.label}
                    </div>
                  </div>
                  {comparisonLibraries.map((lib, colIdx) => {
                    const isGoTui = colIdx === goTuiColIdx;
                    const val = feature.values[lib];
                    return (
                      <div
                        key={lib}
                        className="px-4 py-3.5"
                        style={{
                          borderLeft: isGoTui
                            ? `1px solid ${t.accent}30`
                            : `1px solid ${t.border}`,
                          borderRight: isGoTui && colIdx < comparisonLibraries.length - 1
                            ? `1px solid ${t.accent}30`
                            : undefined,
                          background: isGoTui ? accentTint : "transparent",
                        }}
                      >
                        <span
                          className="font-['Fira_Code',monospace] text-[11px] leading-snug"
                          style={{
                            color: isGoTui ? t.accent : t.text,
                          }}
                        >
                          {val.summary}
                        </span>
                      </div>
                    );
                  })}
                </div>

                {/* Detail panel — slides open, no reflow */}
                <ComparisonDetailPanel
                  feature={feature}
                  expanded={isExpanded}
                  libColors={libColors}
                />
              </div>
            );
          })}
        </div>
      </div>

      {/* Mobile / tablet: card per library */}
      <div className="lg:hidden flex flex-col gap-5">
        {comparisonLibraries.map((lib, libIdx) => {
          const isGoTui = libIdx === goTuiColIdx;
          const color = libColors[lib];
          return (
            <div
              key={lib}
              className={`rounded-lg overflow-hidden ${visible ? "comparison-row-animate" : "opacity-0"}`}
              style={{
                border: `1px solid ${isGoTui ? `${t.accent}55` : t.border}`,
                background: t.bgCard,
                animationDelay: visible ? `${libIdx * 60}ms` : "0ms",
                boxShadow: isGoTui && theme === "dark"
                  ? `0 0 12px ${t.accent}08`
                  : undefined,
              }}
            >
              <div
                className="px-4 py-3 flex items-center gap-3"
                style={{
                  borderBottom: `1px solid ${isGoTui ? `${t.accent}55` : t.border}`,
                  background: isGoTui
                    ? theme === "dark" ? `${t.accent}0a` : `${t.accent}08`
                    : theme === "dark" ? "#23241e" : "#f5f5f1",
                }}
              >
                <div
                  className="w-2 h-2 rounded-full shrink-0"
                  style={{ background: color }}
                />
                <div
                  className="font-['Fira_Code',monospace] text-[13px] font-semibold"
                  style={{ color: isGoTui ? t.accent : t.heading }}
                >
                  {lib}
                </div>
              </div>
              <div className="px-4 py-2">
                {comparisonFeatures.map((feature, fIdx) => {
                  const val = feature.values[lib];
                  return (
                    <div
                      key={feature.label}
                      className="py-2.5"
                      style={{
                        borderBottom:
                          fIdx < comparisonFeatures.length - 1
                            ? `1px solid ${t.border}33`
                            : "none",
                      }}
                    >
                      <div className="flex items-baseline gap-2 mb-0.5">
                        <div
                          className="font-['Fira_Code',monospace] text-[10px] uppercase tracking-wider shrink-0"
                          style={{ color: t.textDim }}
                        >
                          {feature.label}
                        </div>
                        <div
                          className="font-['Fira_Code',monospace] text-[11px]"
                          style={{ color: t.text }}
                        >
                          {val.summary}
                        </div>
                      </div>
                      <div
                        className="font-['IBM_Plex_Sans',sans-serif] text-[11px] leading-relaxed"
                        style={{ color: t.textDim }}
                      >
                        {val.detail}
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          );
        })}
      </div>
    </section>
  );
}

/* ============================================================
   Pages
   ============================================================ */

function HomePage() {
  const { theme, setTheme } = useTheme();
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

  // Interactive prompt state
  const [promptInput, setPromptInput] = useState("");
  const promptRef = useRef<HTMLInputElement>(null);

  return (
    <Page hideNavUntilScroll>
      <div className="relative">
        <PageBackground theme={theme} />
        <div className="relative z-10">
          {/* Hero — Man Page Terminal */}
          <section className="relative" style={{ minHeight: "100vh" }}>
            {/* Subtle top-right controls */}
            <div
              className="tl absolute top-0 right-0 z-20 flex items-center gap-2 font-['Fira_Code',monospace]"
              style={{
                padding: "16px 20px",
                animationDelay: "800ms",
                fontSize: "10px",
                opacity: 0.6,
              }}
              onMouseEnter={(e) => { e.currentTarget.style.opacity = "1"; }}
              onMouseLeave={(e) => { e.currentTarget.style.opacity = "0.6"; }}
            >
              <a
                href="https://pkg.go.dev/github.com/grindlemire/go-tui"
                target="_blank"
                rel="noopener noreferrer"
                className="transition-colors duration-200"
                style={{ color: t.textDim }}
                title="v0.1.0 — view on pkg.go.dev"
                onMouseEnter={(e) => { e.currentTarget.style.color = t.secondary; }}
                onMouseLeave={(e) => { e.currentTarget.style.color = t.textDim; }}
              >
                v0.1.0
              </a>
              <span style={{ color: t.textDim }}>·</span>
              <Link
                to="/guide"
                className="transition-colors duration-200"
                style={{ color: t.textDim }}
                title="Guides"
                onMouseEnter={(e) => { e.currentTarget.style.color = t.secondary; }}
                onMouseLeave={(e) => { e.currentTarget.style.color = t.textDim; }}
              >
                guide
              </Link>
              <span style={{ color: t.textDim }}>·</span>
              <Link
                to="/reference"
                className="transition-colors duration-200"
                style={{ color: t.textDim }}
                title="Reference"
                onMouseEnter={(e) => { e.currentTarget.style.color = t.secondary; }}
                onMouseLeave={(e) => { e.currentTarget.style.color = t.textDim; }}
              >
                reference
              </Link>
              <span style={{ color: t.textDim }}>·</span>
              <a
                href="https://github.com/grindlemire/go-tui"
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center transition-colors duration-200"
                style={{ color: t.textDim }}
                title="View on GitHub"
                onMouseEnter={(e) => { e.currentTarget.style.color = t.accent; }}
                onMouseLeave={(e) => { e.currentTarget.style.color = t.textDim; }}
              >
                <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor" aria-label="GitHub">
                  <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z" />
                </svg>
              </a>
              <span style={{ color: t.textDim }}>·</span>
              <button
                onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
                className="transition-colors duration-200"
                style={{
                  color: t.textDim,
                  background: "none",
                  border: "none",
                  cursor: "pointer",
                  font: "inherit",
                  fontSize: "inherit",
                  padding: 0,
                }}
                title={theme === "dark" ? "Switch to light mode" : "Switch to dark mode"}
                onMouseEnter={(e) => { e.currentTarget.style.color = theme === "dark" ? t.secondary : t.tertiary; }}
                onMouseLeave={(e) => { e.currentTarget.style.color = t.textDim; }}
              >
                {theme === "dark" ? "light" : "dark"}
              </button>
            </div>

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

                {/* Interactive prompt — desktop only */}
                <div
                  className="tl hidden sm:flex items-center mt-8 text-[13px] cursor-text"
                  style={{ animationDelay: "230ms" }}
                  onClick={() => promptRef.current?.focus()}
                >
                  <span style={{ color: t.secondary }}>$</span>
                  <span className="ml-2 relative flex-1">
                    <input
                      ref={promptRef}
                      type="text"
                      value={promptInput}
                      onChange={(e) => setPromptInput(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === "Enter" || e.key === " ") {
                          e.preventDefault();
                          e.stopPropagation();
                          scrollToNextSection();
                          setPromptInput("");
                        }
                      }}
                      className="bg-transparent border-none outline-none font-['Fira_Code',monospace] text-[13px] w-full caret-transparent"
                      style={{ color: t.heading, padding: 0, margin: 0 }}
                      spellCheck={false}
                      autoComplete="off"
                    />
                    {/* Blinking block cursor */}
                    <span
                      className="absolute top-1/2 -translate-y-1/2 pointer-events-none"
                      style={{
                        left: `${promptInput.length}ch`,
                        width: "0.6ch",
                        height: "1.15em",
                        background: t.secondary,
                        animation: "blink 1s step-end infinite",
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
              First-class editor support
            </h2>
            <p
              className="text-[14px] sm:text-[15px] mb-8 sm:mb-10 max-w-[600px]"
              style={{ color: t.textMuted }}
            >
              LSP, tree-sitter grammar, and formatter included.
              Real IDE features, not just syntax coloring.
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
                  { title: "Intelligent completions", description: "Component suggestions with type signatures as you type.", color: t.secondary, editorIdx: 1 },
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
    <Page>
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
            </div>
          </div>

          {/* Main content */}
          <div className="flex-1 min-w-0">
            <MobilePicker
              pages={pages}
              activeIndex={activeSection}
              onSelect={(i) => navigate(`/guide/${pages[i].slug}`)}
            />

            <div className="flex justify-end mb-3">
              <Link
                to={`/guide/${pages[activeSection].slug}/raw`}
                className="font-['Fira_Code',monospace] text-[11px] flex items-center gap-1.5 px-2.5 py-1 rounded transition-colors duration-200"
                style={{ color: t.textDim, textDecoration: "none" }}
                onMouseEnter={(e) => { e.currentTarget.style.color = t.accent; }}
                onMouseLeave={(e) => { e.currentTarget.style.color = t.textDim; }}
                title="View raw markdown (for AI consumption)"
              >
                <span style={{ fontSize: "13px" }}>{ "\u2913" }</span>
                raw markdown
              </Link>
            </div>

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
    </Page>
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
    <Page>
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
            </div>
          </div>

          {/* Main content */}
          <div className="flex-1 min-w-0">
            <MobilePicker
              pages={pages}
              activeIndex={activeCategory}
              onSelect={(i) => navigate(`/reference/${pages[i].slug}`)}
            />

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
    </Page>
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
        <GlobalStyles />
        <ScrollToTop />
        <SearchModal open={searchOpen} onClose={() => setSearchOpen(false)} />
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/guide" element={<GuideRedirect />} />
          <Route path="/guide/:slug" element={<GuidePage />} />
          <Route path="/guide/:slug/raw" element={<RawGuidePage />} />
          <Route path="/reference" element={<ReferenceRedirect />} />
          <Route path="/reference/:slug" element={<ReferencePage />} />
        </Routes>
      </SearchContext.Provider>
    </ThemeContext.Provider>
  );
}
