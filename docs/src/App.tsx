import { useState, useEffect, useRef } from "react";
import { Routes, Route, Link, useLocation } from "react-router-dom";
import { type Theme, palette, ThemeContext, useTheme } from "./lib/theme.ts";
import { projectInfo, tailwindClasses } from "./content/projectInfo.ts";
import { loadGuide, loadReference } from "./lib/markdown.ts";
import { getHighlighter, highlight } from "./lib/highlighter.ts";
import Markdown from "./components/Markdown.tsx";

/* ─── Global Styles ─── */

function GlobalStyles() {
  return (
    <style>{`
      @import url('https://fonts.googleapis.com/css2?family=Fira+Code:wght@300;400;500;600;700&family=IBM+Plex+Sans:wght@300;400;500;600;700&display=swap');

      .neon-glow {
        text-shadow: 0 0 7px #00ffff, 0 0 20px #00ffff44, 0 0 40px #00ffff22;
      }
      .light-theme .neon-glow {
        text-shadow: none;
      }

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
      .custom-scroll::-webkit-scrollbar-track {
        background: transparent;
      }
      .custom-scroll::-webkit-scrollbar-thumb {
        background: #1a1a3a;
        border-radius: 3px;
      }
      .custom-scroll::-webkit-scrollbar-thumb:hover {
        background: #2a2a5a;
      }

      .neon-select ::selection {
        background: #00ffff22;
        color: #00ffff;
      }
      .light-theme .neon-select ::selection {
        background: #0088aa22;
        color: #0088aa;
      }
    `}</style>
  );
}

/* ─── Matrix Rain (hero only, subtle) ─── */

function MatrixRain({ theme }: { theme: Theme }) {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    let animId: number;
    const chars = "abcdefghijklmnopqrstuvwxyz0123456789{}[]<>";
    const fontSize = 14;
    let columns: number;
    let drops: number[];

    const resize = () => {
      canvas.width = canvas.offsetWidth;
      canvas.height = canvas.offsetHeight;
      columns = Math.floor(canvas.width / fontSize);
      drops = Array(columns)
        .fill(0)
        .map(() => Math.random() * -100);
    };

    resize();
    window.addEventListener("resize", resize);

    const draw = () => {
      const fadeColor =
        theme === "dark"
          ? "rgba(5, 5, 16, 0.06)"
          : "rgba(240, 244, 248, 0.08)";
      ctx.fillStyle = fadeColor;
      ctx.fillRect(0, 0, canvas.width, canvas.height);

      const charColor = theme === "dark" ? "#00ffff" : "#0088aa";
      ctx.fillStyle = charColor;
      ctx.font = `${fontSize}px 'Fira Code', monospace`;
      ctx.globalAlpha = theme === "dark" ? 0.08 : 0.04;

      for (let i = 0; i < columns; i++) {
        const char = chars[Math.floor(Math.random() * chars.length)];
        ctx.fillText(char, i * fontSize, drops[i] * fontSize);

        if (drops[i] * fontSize > canvas.height && Math.random() > 0.985) {
          drops[i] = 0;
        }
        drops[i] += 0.3 + Math.random() * 0.2;
      }

      ctx.globalAlpha = 1;
      animId = requestAnimationFrame(draw);
    };

    draw();

    return () => {
      window.removeEventListener("resize", resize);
      cancelAnimationFrame(animId);
    };
  }, [theme]);

  return (
    <canvas
      ref={canvasRef}
      className="absolute inset-0 w-full h-full"
      style={{ opacity: theme === "dark" ? 0.5 : 0.3 }}
    />
  );
}

/* ─── Typing Effect ─── */

function useTypingEffect(text: string, speed = 45) {
  const [displayed, setDisplayed] = useState("");
  const [done, setDone] = useState(false);

  useEffect(() => {
    setDisplayed("");
    setDone(false);
    let i = 0;
    const interval = setInterval(() => {
      if (i < text.length) {
        setDisplayed(text.slice(0, i + 1));
        i++;
      } else {
        setDone(true);
        clearInterval(interval);
      }
    }, speed);
    return () => clearInterval(interval);
  }, [text, speed]);

  return { displayed, done };
}

/* ─── Nav ─── */

function Nav() {
  const { theme, setTheme } = useTheme();
  const t = palette[theme];
  const location = useLocation();
  const [mobileOpen, setMobileOpen] = useState(false);

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
      className="sticky top-0 z-40 backdrop-blur-md"
      style={{
        background:
          theme === "dark"
            ? "rgba(5, 5, 16, 0.9)"
            : "rgba(240, 244, 248, 0.92)",
        borderBottom: `1px solid ${t.border}`,
      }}
    >
      <div className="max-w-[1100px] mx-auto px-4 sm:px-6 h-12 flex items-center justify-between">
        <Link
          to="/"
          className="font-['Fira_Code',monospace] text-sm font-bold tracking-tight flex items-center gap-1"
          style={{
            color: t.accent,
            textShadow: theme === "dark" ? t.accentGlowSubtle : "none",
          }}
        >
          go-tui
          <span
            className="inline-block w-1.5 h-4 ml-0.5"
            style={{
              background: t.accent,
              animation: "blink 1s step-end infinite",
              opacity: 0.6,
            }}
          />
        </Link>

        {/* Desktop links */}
        <div className="hidden sm:flex items-center gap-1">
          {links.map((link) => {
            const active = isActive(link.to);
            return (
              <Link
                key={link.to}
                to={link.to}
                className="font-['Fira_Code',monospace] text-xs px-3 py-1.5 rounded transition-all duration-200"
                style={{
                  color: active ? t.accent : t.textMuted,
                  background: active
                    ? theme === "dark"
                      ? "#00ffff0a"
                      : "#0088aa0a"
                    : "transparent",
                  border: `1px solid ${active ? (theme === "dark" ? "#00ffff33" : "#0088aa33") : "transparent"}`,
                  textShadow:
                    active && theme === "dark" ? t.accentGlowSubtle : "none",
                }}
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

          <div
            className="mx-2"
            style={{ width: 1, height: 20, background: t.border }}
          />

          <button
            onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
            className="font-['Fira_Code',monospace] text-xs p-1.5 rounded transition-all duration-300"
            style={{
              color: theme === "dark" ? t.secondary : t.tertiary,
              background: "transparent",
              border: `1px solid ${theme === "dark" ? "#39ff1433" : "#aa00aa33"}`,
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
            onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
            className="font-['Fira_Code',monospace] text-[10px] p-1.5 rounded"
            style={{
              color: theme === "dark" ? t.secondary : t.tertiary,
              background: "transparent",
              border: `1px solid ${theme === "dark" ? "#39ff1433" : "#aa00aa33"}`,
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
                ? "rgba(5, 5, 16, 0.95)"
                : "rgba(240, 244, 248, 0.98)",
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
                      ? "#00ffff0a"
                      : "#0088aa0a"
                    : "transparent",
                }}
              >
                {link.label}
              </Link>
            );
          })}
        </div>
      )}
    </nav>
  );
}

/* ─── Code Block (used on HomePage) ─── */

function CodeBlock({
  code,
  title,
  language,
}: {
  code: string;
  title?: string;
  language?: string;
}) {
  const { theme } = useTheme();
  const t = palette[theme];
  const [ready, setReady] = useState(false);

  useEffect(() => {
    getHighlighter().then(() => setReady(true));
  }, []);

  const html = ready && language ? highlight(code, language, theme) : "";

  return (
    <div
      className="rounded-lg overflow-hidden"
      style={{
        background: t.bgCode,
        border: `1px solid ${t.border}`,
        boxShadow:
          theme === "dark"
            ? `0 0 12px ${t.borderGlow}, inset 0 0 20px rgba(0,0,0,0.3)`
            : "0 1px 3px rgba(0,0,0,0.08)",
      }}
    >
      <div
        className="flex items-center justify-between px-4 py-2"
        style={{ borderBottom: `1px solid ${t.border}` }}
      >
        <div className="flex items-center gap-2">
          <div className="flex gap-1.5">
            <div
              className="w-2.5 h-2.5 rounded-full"
              style={{ background: "#ff5f57" }}
            />
            <div
              className="w-2.5 h-2.5 rounded-full"
              style={{ background: "#febc2e" }}
            />
            <div
              className="w-2.5 h-2.5 rounded-full"
              style={{ background: "#28c840" }}
            />
          </div>
          {title && (
            <span
              className="font-['Fira_Code',monospace] text-[10px] ml-2"
              style={{ color: t.textDim }}
            >
              {title}
            </span>
          )}
        </div>
        {language && (
          <span
            className="font-['Fira_Code',monospace] text-[10px]"
            style={{ color: t.accentDim }}
          >
            {language}
          </span>
        )}
      </div>
      <div className="px-4 py-3.5 overflow-x-auto custom-scroll">
        {html ? (
          <div
            className="[&_pre]:!bg-transparent [&_pre]:!m-0 [&_pre]:!p-0 [&_code]:!text-[12px] [&_code]:!sm:text-[13px] [&_code]:!leading-[1.7] [&_code]:font-['Fira_Code',monospace]"
            dangerouslySetInnerHTML={{ __html: html }}
          />
        ) : (
          <pre className="font-['Fira_Code',monospace] text-[12px] sm:text-[13px] leading-[1.7]">
            <code style={{ color: t.text }}>{code}</code>
          </pre>
        )}
      </div>
    </div>
  );
}

/* ─── Terminal Block ─── */

function TerminalBlock({
  command,
  title = "terminal",
}: {
  command: string;
  title?: string;
}) {
  const { theme } = useTheme();
  const t = palette[theme];
  return (
    <div
      className="rounded-lg overflow-hidden font-['Fira_Code',monospace]"
      style={{
        background: t.bgCode,
        border: `1px solid ${t.border}`,
        boxShadow:
          theme === "dark"
            ? `0 0 12px ${t.borderGlow}`
            : "0 1px 3px rgba(0,0,0,0.08)",
      }}
    >
      <div
        className="flex items-center gap-2 px-4 py-2"
        style={{ borderBottom: `1px solid ${t.border}` }}
      >
        <div className="flex gap-1.5">
          <div
            className="w-2.5 h-2.5 rounded-full"
            style={{ background: "#ff5f57" }}
          />
          <div
            className="w-2.5 h-2.5 rounded-full"
            style={{ background: "#febc2e" }}
          />
          <div
            className="w-2.5 h-2.5 rounded-full"
            style={{ background: "#28c840" }}
          />
        </div>
        <span className="text-[10px] ml-2" style={{ color: t.textDim }}>
          {title}
        </span>
      </div>
      <div className="px-4 py-3.5 flex items-center gap-2 text-[13px] sm:text-[14px] overflow-x-auto custom-scroll">
        <span style={{ color: t.secondary }}>$</span>
        <span style={{ color: t.text }}>{command}</span>
      </div>
    </div>
  );
}

/* ─── Feature Card ─── */

function FeatureCard({
  feature,
  index,
}: {
  feature: { title: string; description: string; icon: string };
  index: number;
}) {
  const { theme } = useTheme();
  const t = palette[theme];
  const [hovered, setHovered] = useState(false);

  const iconMap: Record<string, string> = {
    code: "</>",
    layout: "[=]",
    zap: "/!/",
    box: "[ ]",
    edit: " ~ ",
    package: "{..}",
  };

  const neonColors = [t.accent, t.secondary, t.tertiary];
  const color = neonColors[index % 3];

  return (
    <div
      className="rounded-lg p-5 transition-all duration-300"
      style={{
        background: t.bgCard,
        border: `1px solid ${hovered ? color : t.border}`,
        boxShadow:
          hovered && theme === "dark"
            ? `0 0 18px ${color}33, 0 0 40px ${color}11`
            : theme === "dark"
              ? "0 0 5px #00000044"
              : "0 1px 3px rgba(0,0,0,0.06)",
      }}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
    >
      <div
        className="font-['Fira_Code',monospace] text-sm mb-3 inline-block px-2.5 py-1 rounded"
        style={{
          color: color,
          background: `${color}0a`,
          textShadow:
            hovered && theme === "dark" ? `0 0 5px ${color}88` : "none",
        }}
      >
        {iconMap[feature.icon] || ">>>"}
      </div>
      <h3
        className="font-['IBM_Plex_Sans',sans-serif] font-semibold text-[15px] mb-2"
        style={{ color: t.heading }}
      >
        {feature.title}
      </h3>
      <p
        className="font-['IBM_Plex_Sans',sans-serif] text-[13px] leading-relaxed"
        style={{ color: t.textMuted }}
      >
        {feature.description}
      </p>
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
              ? "linear-gradient(to right, transparent, #00ffff18, #ff00ff18, transparent)"
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
            go-tui &mdash; declarative terminal UI for Go
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

function Page({ children }: { children: React.ReactNode }) {
  const { theme } = useTheme();
  const t = palette[theme];
  return (
    <div
      className={`${theme === "dark" ? "dark-theme" : "light-theme"} neon-select`}
      style={{
        background: t.bg,
        color: t.text,
        minHeight: "100vh",
        fontFamily: "'IBM Plex Sans', sans-serif",
      }}
    >
      <Nav />
      {children}
      <Footer />
    </div>
  );
}

/* ============================================================
   Pages
   ============================================================ */

function HomePage() {
  const { theme } = useTheme();
  const t = palette[theme];
  const { displayed, done } = useTypingEffect(projectInfo.tagline, 40);

  return (
    <Page>
      {/* Hero */}
      <section className="relative overflow-hidden">
        <MatrixRain theme={theme} />
        <div className="relative z-10 max-w-[1100px] mx-auto px-4 sm:px-6 pt-16 sm:pt-24 pb-20 sm:pb-28">
          <h1 className="font-['IBM_Plex_Sans',sans-serif] text-3xl sm:text-5xl md:text-6xl font-bold tracking-tight leading-[1.1] max-w-[720px]">
            <span
              className={theme === "dark" ? "neon-glow" : ""}
              style={{ color: t.accent }}
            >
              {displayed}
            </span>
            {!done && (
              <span
                className="inline-block w-2 sm:w-3 h-8 sm:h-12 ml-1 align-middle"
                style={{
                  background: t.accent,
                  animation: "blink 0.7s step-end infinite",
                }}
              />
            )}
          </h1>

          <p
            className="text-base sm:text-lg mt-6 sm:mt-8 max-w-[560px] leading-relaxed"
            style={{ color: t.textMuted }}
          >
            {projectInfo.description}
          </p>

          <div className="mt-8 sm:mt-10 flex items-center gap-3 sm:gap-4 flex-wrap">
            <Link
              to="/guide"
              className="font-['Fira_Code',monospace] inline-flex items-center gap-2 px-4 sm:px-5 py-2.5 text-sm rounded transition-all duration-200"
              style={{
                background: theme === "dark" ? "#00ffff10" : t.accent,
                color: theme === "dark" ? t.accent : "#ffffff",
                border: `1px solid ${theme === "dark" ? "#00ffff44" : t.accent}`,
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.background =
                  theme === "dark" ? "#00ffff20" : t.accentDim;
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background =
                  theme === "dark" ? "#00ffff10" : t.accent;
              }}
            >
              get started
            </Link>
            <Link
              to="/reference"
              className="font-['Fira_Code',monospace] inline-flex items-center gap-2 px-4 sm:px-5 py-2.5 text-sm rounded transition-all duration-200"
              style={{
                color: t.textMuted,
                border: `1px solid ${t.border}`,
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.color = t.accent;
                e.currentTarget.style.borderColor =
                  theme === "dark" ? "#00ffff44" : t.accent;
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.color = t.textMuted;
                e.currentTarget.style.borderColor = t.border;
              }}
            >
              api reference
            </Link>
          </div>

          <div className="mt-8 sm:mt-10 max-w-[600px]">
            <TerminalBlock command={projectInfo.installCmd} />
          </div>
        </div>
      </section>

      <Divider />

      {/* Quick Example */}
      <section className="max-w-[1100px] mx-auto px-4 sm:px-6 py-10 sm:py-12">
        <div
          className="font-['Fira_Code',monospace] text-[10px] tracking-[0.2em] uppercase mb-3"
          style={{ color: t.accentDim }}
        >
          quick start
        </div>
        <h2
          className="text-2xl sm:text-3xl font-bold tracking-tight mb-3"
          style={{ color: t.heading }}
        >
          Write .gsx, run Go
        </h2>
        <p
          className="text-[14px] sm:text-[15px] mb-8 sm:mb-10 max-w-[560px]"
          style={{ color: t.textMuted }}
        >
          Define your UI in .gsx files with HTML-like syntax and Tailwind-style
          classes. It compiles to type-safe Go.
        </p>

        <div className="grid md:grid-cols-2 gap-4 sm:gap-6">
          <div>
            <div
              className="font-['Fira_Code',monospace] text-[10px] tracking-[0.15em] uppercase mb-3 flex items-center gap-2"
              style={{ color: t.textDim }}
            >
              <span style={{ color: t.accentDim }}>01</span>
              define
            </div>
            <CodeBlock
              title="dashboard.gsx"
              language="gsx"
              code={`templ Dashboard() {
  <div class="flex-col h-full">
    <div class="border-single p-1">
      <span class="font-bold text-cyan">
        Dashboard
      </span>
    </div>
    <div class="flex grow gap-2 p-1">
      @Sidebar()
      @MainContent()
    </div>
  </div>
}`}
            />
          </div>
          <div>
            <div
              className="font-['Fira_Code',monospace] text-[10px] tracking-[0.15em] uppercase mb-3 flex items-center gap-2"
              style={{ color: t.textDim }}
            >
              <span style={{ color: t.secondaryDim }}>02</span>
              run
            </div>
            <CodeBlock
              title="main.go"
              language="go"
              code={`package main

import (
  "fmt"
  "os"
  tui "github.com/grindlemire/go-tui"
)

func main() {
  app, err := tui.NewApp()
  if err != nil {
    fmt.Fprintf(os.Stderr, "%v\\n", err)
    os.Exit(1)
  }
  defer app.Close()
  app.SetRootComponent(Dashboard())
  app.Run()
}`}
            />
          </div>
        </div>
      </section>

      <Divider />

      {/* Features */}
      <section className="max-w-[1100px] mx-auto px-4 sm:px-6 py-10 sm:py-12">
        <div
          className="font-['Fira_Code',monospace] text-[10px] tracking-[0.2em] uppercase mb-3"
          style={{ color: t.tertiaryDim }}
        >
          features
        </div>
        <h2
          className="text-2xl sm:text-3xl font-bold tracking-tight mb-8 sm:mb-12"
          style={{ color: t.heading }}
        >
          What's in the box
        </h2>
        <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-5">
          {projectInfo.features.map((feature, i) => (
            <FeatureCard key={i} feature={feature} index={i} />
          ))}
        </div>
      </section>

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
          Style your terminal UI with familiar utility classes that compile to
          type-safe Go options.
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
    </Page>
  );
}

/* ─── Guide Page ─── */

function GuidePage() {
  const { theme } = useTheme();
  const t = palette[theme];

  const pages = loadGuide();
  const [activeSection, setActiveSection] = useState(0);

  return (
    <Page>
      <div className="max-w-[1100px] mx-auto px-4 sm:px-6 pt-10 sm:pt-16 pb-16 sm:pb-24">
        <h1
          className="text-3xl sm:text-5xl font-bold tracking-tight mb-8 sm:mb-12"
          style={{ color: t.heading }}
        >
          Guide
        </h1>

        {/* Tabs — scrollable on mobile */}
        <div
          className="flex gap-1 mb-8 sm:mb-10 overflow-x-auto custom-scroll pb-2 -mx-4 px-4 sm:mx-0 sm:px-0"
          style={{ borderBottom: `1px solid ${t.border}` }}
        >
          {pages.map((page, i) => {
            const active = activeSection === i;
            return (
              <button
                key={page.slug}
                onClick={() => setActiveSection(i)}
                className="font-['Fira_Code',monospace] text-[11px] sm:text-xs px-3 sm:px-4 py-2 whitespace-nowrap transition-all duration-200"
                style={{
                  color: active ? t.accent : t.textMuted,
                  background: "transparent",
                  border: "none",
                  borderBottom: `2px solid ${active ? t.accent : "transparent"}`,
                  marginBottom: -1,
                  cursor: "pointer",
                  textShadow:
                    active && theme === "dark" ? t.accentGlowSubtle : "none",
                }}
                onMouseEnter={(e) => {
                  if (!active) e.currentTarget.style.color = t.accent;
                }}
                onMouseLeave={(e) => {
                  if (!active) e.currentTarget.style.color = t.textMuted;
                }}
              >
                {page.title}
              </button>
            );
          })}
        </div>

        {/* Content */}
        <div className="max-w-[780px] fade-in" key={activeSection}>
          <h2
            className="text-2xl sm:text-3xl font-bold tracking-tight mb-6 sm:mb-8"
            style={{ color: t.heading }}
          >
            {pages[activeSection].title}
          </h2>

          <Markdown content={pages[activeSection].body} />
        </div>

        {/* Prev / Next */}
        <div
          className="mt-10 sm:mt-12 pt-6 sm:pt-8 flex justify-between"
          style={{ borderTop: `1px solid ${t.border}` }}
        >
          {activeSection > 0 ? (
            <button
              onClick={() => setActiveSection(activeSection - 1)}
              className="font-['Fira_Code',monospace] text-xs sm:text-sm transition-colors duration-200"
              style={{
                color: t.textMuted,
                background: "transparent",
                border: "none",
                cursor: "pointer",
              }}
              onMouseEnter={(e) => (e.currentTarget.style.color = t.accent)}
              onMouseLeave={(e) =>
                (e.currentTarget.style.color = t.textMuted)
              }
            >
              &larr; {pages[activeSection - 1].title}
            </button>
          ) : (
            <div />
          )}
          {activeSection < pages.length - 1 ? (
            <button
              onClick={() => setActiveSection(activeSection + 1)}
              className="font-['Fira_Code',monospace] text-xs sm:text-sm transition-colors duration-200"
              style={{
                color: t.textMuted,
                background: "transparent",
                border: "none",
                cursor: "pointer",
              }}
              onMouseEnter={(e) => (e.currentTarget.style.color = t.accent)}
              onMouseLeave={(e) =>
                (e.currentTarget.style.color = t.textMuted)
              }
            >
              {pages[activeSection + 1].title} &rarr;
            </button>
          ) : (
            <div />
          )}
        </div>
      </div>
    </Page>
  );
}

/* ─── Reference Page ─── */

function ReferencePage() {
  const { theme } = useTheme();
  const t = palette[theme];

  const pages = loadReference();
  const [activeCategory, setActiveCategory] = useState(0);

  return (
    <Page>
      <div className="max-w-[1100px] mx-auto px-4 sm:px-6 pt-10 sm:pt-16 pb-16 sm:pb-24">
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
                  <button
                    key={page.slug}
                    onClick={() => setActiveCategory(i)}
                    className="block w-full text-left font-['Fira_Code',monospace] text-[12px] py-1.5 px-3 rounded transition-all duration-200"
                    style={{
                      color: active ? activeColor : t.textMuted,
                      background: active
                        ? isTailwind
                          ? theme === "dark"
                            ? "#39ff1408"
                            : "#1a8a0a08"
                          : theme === "dark"
                            ? "#00ffff08"
                            : "#0088aa08"
                        : "transparent",
                      border: "none",
                      borderLeft: `2px solid ${active ? activeColor : "transparent"}`,
                      cursor: "pointer",
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
                  </button>
                );
              })}
            </div>
          </div>

          {/* Main content */}
          <div className="flex-1 min-w-0">
            {/* Mobile category picker */}
            <div className="md:hidden mb-6 sm:mb-8">
              <select
                value={activeCategory}
                onChange={(e) => setActiveCategory(Number(e.target.value))}
                className="font-['Fira_Code',monospace] w-full rounded px-3 py-2 text-sm"
                style={{
                  background: t.bgCard,
                  color: t.text,
                  border: `1px solid ${t.border}`,
                }}
              >
                {pages.map((page, i) => (
                  <option key={page.slug} value={i}>
                    {page.title}
                  </option>
                ))}
              </select>
            </div>

            <div className="fade-in" key={activeCategory}>
              <h2
                className="text-2xl sm:text-3xl font-bold tracking-tight mb-6 sm:mb-8"
                style={{ color: t.heading }}
              >
                {pages[activeCategory].title}
              </h2>

              <Markdown content={pages[activeCategory].body} />
            </div>
          </div>
        </div>
      </div>
    </Page>
  );
}

/* ─── Main Export ─── */

export default function Design2() {
  const [theme, setTheme] = useState<Theme>("dark");

  return (
    <ThemeContext.Provider value={{ theme, setTheme }}>
      <GlobalStyles />
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/guide" element={<GuidePage />} />
        <Route path="/reference" element={<ReferencePage />} />
      </Routes>
    </ThemeContext.Provider>
  );
}
