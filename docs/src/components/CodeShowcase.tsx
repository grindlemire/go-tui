import { useState, useEffect, useRef, useCallback, useMemo } from "react";
import { useTheme, palette } from "../lib/theme.ts";
import { getHighlighter, highlight } from "../lib/highlighter.ts";

/* ─── Code Showcase ───
 * A large annotated code editor with an integrated annotation rail.
 * Annotations float alongside their code sections on the right.
 * Click pills or hover annotations to spotlight sections.
 */

const gsxCode = `package main

import (
\t"fmt"
\t"time"
\ttui "github.com/grindlemire/go-tui"
)

type counterApp struct {
\tcount   *tui.State[int]
\telapsed *tui.State[int]
\tpeak    *tui.State[int]
}

func Counter() *counterApp {
\treturn &counterApp{
\t\tcount:   tui.NewState(0),
\t\telapsed: tui.NewState(0),
\t\tpeak:    tui.NewState(0),
\t}
}

func (c *counterApp) KeyMap() tui.KeyMap {
\treturn tui.KeyMap{
\t\ttui.On(tui.Rune('+'), func(ke tui.KeyEvent) {
\t\t\tc.count.Update(func(v int) int { return v + 1 })
\t\t}),
\t\ttui.On(tui.Rune('-'), func(ke tui.KeyEvent) {
\t\t\tc.count.Update(func(v int) int { return v - 1 })
\t\t}),
\t\ttui.On(tui.Rune('r'), func(ke tui.KeyEvent) { c.count.Set(0) }),
\t\ttui.On(tui.Rune('q'), func(ke tui.KeyEvent) { ke.App().Stop() }),
\t}
}

func (c *counterApp) Watchers() []tui.Watcher {
\treturn []tui.Watcher{
\t\ttui.OnTimer(time.Second, func() {
\t\t\tc.elapsed.Update(func(v int) int { return v + 1 })
\t\t}),
\t\ttui.OnChange(c.count, func(v int) {
\t\t\tif v > c.peak.Get() {
\t\t\t\tc.peak.Set(v)
\t\t\t}
\t\t}),
\t}
}

func formatTime(seconds int) string {
\treturn fmt.Sprintf("%d:%02d", seconds/60, seconds%60)
}

templ Badge(label string, value string, color string) {
\t<div class="flex gap-1">
\t\t<span class="font-dim">{label}</span>
\t\t<span class={"font-bold " + color}>{value}</span>
\t</div>
}

templ Card(title string) {
\t<div class="flex-col border-rounded p-1 gap-1" flexGrow={1.0}>
\t\t<span class="font-bold text-cyan">{title}</span>
\t\t<hr />
\t\t{children...}
\t</div>
}

templ (c *counterApp) Render() {
\t<div class="flex-col border-rounded p-1 gap-1">
\t\t<div class="flex justify-between">
\t\t\t<span class="font-bold text-cyan">Counter</span>
\t\t\t<div class="flex gap-2">
\t\t\t\t@Badge("peak:", fmt.Sprintf("%d", c.peak.Get()), "text-magenta")
\t\t\t\t@Badge("uptime:", formatTime(c.elapsed.Get()), "text-yellow")
\t\t\t</div>
\t\t</div>
\t\t<hr />
\t\t<div class="flex gap-2">
\t\t\t@Card("Count") {
\t\t\t\t<span class="text-cyan font-bold">
\t\t\t\t\t{fmt.Sprintf("%d", c.count.Get())}
\t\t\t\t</span>
\t\t\t}
\t\t\t@Card("Status") {
\t\t\t\tif c.count.Get() > 0 {
\t\t\t\t\t<span class="text-green font-bold">Positive</span>
\t\t\t\t} else if c.count.Get() < 0 {
\t\t\t\t\t<span class="text-red font-bold">Negative</span>
\t\t\t\t} else {
\t\t\t\t\t<span class="text-blue font-bold">Zero</span>
\t\t\t\t}
\t\t\t}
\t\t</div>
\t\t<div class="flex gap-1 justify-center">
\t\t\t<span class="font-dim">+/- count \u00b7 0 reset \u00b7 </span>
\t\t</div>
\t</div>
}`;

const mainGoCode = `package main

import (
\t"fmt"
\t"os"
\ttui "github.com/grindlemire/go-tui"
)

func main() {
\tapp, err := tui.NewApp(
\t\ttui.WithRootComponent(Counter()),
\t)
\tif err != nil {
\t\tfmt.Fprintf(os.Stderr, "%v\\n", err)
\t\tos.Exit(1)
\t}
\tdefer app.Close()
\tif err := app.Run(); err != nil {
\t\tfmt.Fprintf(os.Stderr, "%v\\n", err)
\t\tos.Exit(1)
\t}
}`;

interface FileDef {
  name: string;
  code: string;
  language: string;
}

const files: FileDef[] = [
  { name: "counter.gsx", code: gsxCode, language: "gsx" },
  { name: "main.go", code: mainGoCode, language: "go" },
  { name: "Terminal", code: "", language: "terminal" },
];

/* ─── Tutorial Steps ─── */
type Step = {
  id: string;
  label: string;
  description: string;
  lines: number[]; /* 1-indexed */
  color: string;
};

const gsxStepDefs: Omit<Step, "color">[] = [
  {
    id: "state",
    label: "State",
    description:
      "Reactive State[T] values live on the component struct. The constructor initializes them with NewState.",
    lines: [9, 10, 11, 12, 13, 15, 16, 17, 18, 19, 20, 21],
  },
  {
    id: "events",
    label: "Keyboard Events",
    description:
      "KeyMap binds keys to actions. On with Rune() matches character keys; Update and Set mutate state and trigger re-renders.",
    lines: [23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34],
  },
  {
    id: "watchers",
    label: "Watchers",
    description:
      "Side effects that react to state changes or run on intervals, like React's useEffect. OnChange watches a State value and fires when it changes; OnTimer fires on a fixed interval.",
    lines: [36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47],
  },
  {
    id: "components",
    label: "Components",
    description:
      "Reusable templ functions that take props and return elements. Card accepts {children...} as a content slot.",
    lines: [53, 54, 55, 56, 57, 58, 60, 61, 62, 63, 64, 65, 66],
  },
  {
    id: "template",
    label: "Composition",
    description:
      "The Render template ties it together: @-calls nest components and if/else adds conditionals.",
    lines: [
      68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79, 80, 81, 82, 83, 84,
      85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 96, 97, 98,
    ],
  },
];

const mainGoStepDefs: Omit<Step, "color">[] = [
  {
    id: "entry",
    label: "Entry point",
    description:
      "Pass your component to WithRootComponent. The framework handles the event loop, rendering, and cleanup.",
    lines: [10, 11, 12],
  },
];

/* Colors per step per file, indexed by theme */
const gsxStepColors = {
  dark: ["#66d9ef", "#a6e22e", "#e6db74", "#ae81ff", "#f92672"],
  light: ["#217f96", "#507009", "#7a6e00", "#6e5dc6", "#c01f5c"],
};

const mainGoStepColors = {
  dark: ["#66d9ef"],
  light: ["#217f96"],
};

/* Per-file step config */
const fileSteps: Record<number, { defs: Omit<Step, "color">[]; colors: { dark: string[]; light: string[] } }> = {
  0: { defs: gsxStepDefs, colors: gsxStepColors },
  1: { defs: mainGoStepDefs, colors: mainGoStepColors },
};

const LINE_H = 20;
const CODE_PAD_Y = 16;
const ANNOTATION_W = 272;

/* Subtle scrollbar styles */
const scrollbarCSS = `
.cshowcase-scroll::-webkit-scrollbar { width: 5px; height: 5px; }
.cshowcase-scroll::-webkit-scrollbar-track { background: transparent; }
.cshowcase-scroll::-webkit-scrollbar-thumb { background: rgba(128,128,128,0.18); border-radius: 4px; }
.cshowcase-scroll::-webkit-scrollbar-thumb:hover { background: rgba(128,128,128,0.32); }
.cshowcase-scroll { scrollbar-width: thin; scrollbar-color: rgba(128,128,128,0.18) transparent; }
`;

/* ─── Main ─── */
export default function CodeShowcase() {
  const { theme } = useTheme();
  const t = palette[theme];
  const [ready, setReady] = useState(false);
  const [revealed, setRevealed] = useState(false);
  const [activeFile, setActiveFile] = useState(0);
  const [hoveredStep, setHoveredStep] = useState<number | null>(null);
  const [activeStep, setActiveStep] = useState<number | null>(null);
  const [isWide, setIsWide] = useState(
    typeof window !== "undefined" ? window.innerWidth > 920 : true,
  );
  const [termCount, setTermCount] = useState(0);
  const [termElapsed, setTermElapsed] = useState(0);
  const [termPeak, setTermPeak] = useState(0);
  const wideScrollRef = useRef<HTMLDivElement>(null);
  const codeScrollRef = useRef<HTMLDivElement>(null);
  const gutterScrollRef = useRef<HTMLDivElement>(null);

  const isTerminal = activeFile === 2;

  useEffect(() => {
    getHighlighter().then(() => setReady(true));
  }, []);

  useEffect(() => {
    const timer = setTimeout(() => setRevealed(true), 100);
    return () => clearTimeout(timer);
  }, []);

  useEffect(() => {
    const handler = () => setIsWide(window.innerWidth > 920);
    window.addEventListener("resize", handler);
    return () => window.removeEventListener("resize", handler);
  }, []);

  /* Terminal interactivity: timer */
  useEffect(() => {
    if (!isTerminal) {
      setTermElapsed(0);
      setTermCount(0);
      setTermPeak(0);
      return;
    }
    const interval = setInterval(() => setTermElapsed((e) => e + 1), 1000);
    return () => clearInterval(interval);
  }, [isTerminal]);

  /* Terminal interactivity: keyboard */
  useEffect(() => {
    if (!isTerminal) return;
    const handler = (e: KeyboardEvent) => {
      if (
        e.target instanceof HTMLInputElement ||
        e.target instanceof HTMLTextAreaElement
      )
        return;
      if (e.key === "+" || e.key === "=") {
        e.preventDefault();
        setTermCount((c) => {
          const next = c + 1;
          setTermPeak((p) => Math.max(p, next));
          return next;
        });
      } else if (e.key === "-") {
        e.preventDefault();
        setTermCount((c) => c - 1);
      } else if (e.key === "0") {
        e.preventDefault();
        setTermCount(0);
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [isTerminal]);

  const file = files[activeFile];
  const activeCode = file.code;
  const activeLines = activeCode.split("\n");

  /* ─── Full-code syntax highlighting ───
   * Highlight the entire file at once so shiki has full grammar context
   * (function bodies, struct fields, etc.), then split into per-line HTML. */
  const highlightedLines = useMemo(() => {
    if (!ready) return null;
    if (file.language === "text") return null;
    if (file.language === "terminal") return null;
    const fullHtml = highlight(activeCode, file.language, theme);
    if (!fullHtml) return null;
    /* Strip <pre> and <code> wrappers */
    const inner = fullHtml
      .replace(/<\/?pre[^>]*>/g, "")
      .replace(/<\/?code[^>]*>/g, "");
    /* Each source line is wrapped in <span class="line">…</span> */
    return inner
      .split("\n")
      .map((l) =>
        l.replace(/^<span class="line">/, "").replace(/<\/span>$/, ""),
      );
  }, [ready, theme, activeFile]);

  /* Sync gutter scroll with code scroll (narrow mode) */
  const onCodeScroll = useCallback(() => {
    if (codeScrollRef.current && gutterScrollRef.current) {
      gutterScrollRef.current.scrollTop = codeScrollRef.current.scrollTop;
    }
  }, []);

  /* Derive steps for the active file */
  const isDark = theme === "dark";
  const fileStepConfig = fileSteps[activeFile];
  const stepDefs = fileStepConfig?.defs ?? [];
  const stepColorSet = fileStepConfig?.colors ?? { dark: [], light: [] };
  const colors = stepColorSet[theme];
  const steps: Step[] = stepDefs.map((s, i) => ({ ...s, color: colors[i] }));
  const hasSteps = steps.length > 0;

  /* Build line → step mapping */
  const lineToStep = new Map<number, number>();
  steps.forEach((step, i) => {
    step.lines.forEach((ln) => lineToStep.set(ln, i));
  });

  /* Calculate annotation y positions (aligned to first line of each range) */
  const annotationYs = steps.map((step) => {
    const firstLine = Math.min(...step.lines);
    return (firstLine - 1) * LINE_H + CODE_PAD_Y;
  });

  const focusStep = hoveredStep ?? activeStep;

  /* Scroll code view to center a given step (called directly from pill handlers) */
  const scrollToStep = useCallback((idx: number) => {
    if (idx >= stepDefs.length) return;
    const scrollEl = isWide ? wideScrollRef.current : codeScrollRef.current;
    if (!scrollEl) return;
    const step = stepDefs[idx];
    const firstLine = Math.min(...step.lines);
    const lastLine = Math.max(...step.lines);
    const midLine = (firstLine + lastLine) / 2;
    const targetScroll =
      (midLine - 1) * LINE_H - scrollEl.clientHeight / 2 + CODE_PAD_Y;
    scrollEl.scrollTo({
      top: Math.max(0, targetScroll),
      behavior: "smooth",
    });
  }, [stepDefs, isWide]);

  /* Theme palette */
  const borderColor = isDark ? "#49483e" : "#d8d8d0";
  const gutterBg = isDark ? "#1e1f1b" : "#ededea";
  const gutterBorder = isDark
    ? "rgba(73,72,62,0.4)"
    : "rgba(216,216,208,0.6)";
  const gutterText = isDark ? "#5c5c50" : "#9e9b8c";
  const titleBarBg = isDark ? "#1e1f1b" : "#ededea";
  const tabActiveBg = isDark ? "#23241e" : "#f5f5f1";
  const tabActiveText = isDark ? "#f8f8f2" : "#3d3c34";
  const tabInactiveText = isDark ? "#5c5c50" : "#767260";
  const dotColors = isDark
    ? ["#ff5f57", "#febc2e", "#28c840"]
    : ["#ff6159", "#ffbf2f", "#2bc840"];
  const statusBarBg = isDark ? "#1a1b17" : "#e8e8e3";

  const maxCodeH = 580;

  /* In wide mode: hover overrides click for spotlight */
  const wideFocusStep = focusStep;
  const focusColor =
    wideFocusStep !== null ? steps[wideFocusStep].color : null;

  /* ─── Shared sub-elements ─── */

  const titleBar = (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        height: 38,
        padding: "0 14px",
        background: titleBarBg,
        borderBottom: `1px solid ${borderColor}`,
        userSelect: "none",
        gap: 12,
      }}
    >
      <div style={{ display: "flex", gap: 7, flexShrink: 0 }}>
        {dotColors.map((c, i) => (
          <div
            key={i}
            style={{
              width: 11,
              height: 11,
              borderRadius: "50%",
              background: c,
              opacity: isDark ? 0.85 : 0.9,
            }}
          />
        ))}
      </div>

      <div
        style={{
          display: "flex",
          alignItems: "center",
          flex: 1,
          gap: 0,
          marginLeft: 4,
        }}
      >
        {files.map((f, fi) => {
          const isActive = activeFile === fi;
          return (
            <div
              key={f.name}
              onClick={() => {
                setActiveFile(fi);
                setActiveStep(null);
                setHoveredStep(null);
              }}
              style={{
                display: "flex",
                alignItems: "center",
                gap: 6,
                padding: isActive ? "5px 14px 5px 10px" : "5px 12px",
                background: isActive ? tabActiveBg : "transparent",
                borderRadius: isActive ? "6px 6px 0 0" : 0,
                marginBottom: isActive ? -1 : 0,
                fontFamily: "'Fira Code', monospace",
                fontSize: isActive ? 11.5 : 11,
                fontWeight: isActive ? 500 : 400,
                color: isActive ? tabActiveText : tabInactiveText,
                letterSpacing: "0.01em",
                cursor: "pointer",
                transition: "color 0.2s",
              }}
            >
              {isActive && fi === 0 && (
                <svg width="14" height="14" viewBox="0 0 16 16" fill="none">
                  <path
                    d="M4 1h5.5L13 4.5V14a1 1 0 01-1 1H4a1 1 0 01-1-1V2a1 1 0 011-1z"
                    stroke={t.accent}
                    strokeWidth="1.2"
                    fill="none"
                  />
                  <path
                    d="M9 1v4h4"
                    stroke={t.accent}
                    strokeWidth="1.2"
                    fill="none"
                  />
                </svg>
              )}
              {isActive && fi === 2 && (
                <svg width="14" height="14" viewBox="0 0 16 16" fill="none">
                  <rect
                    x="1" y="2" width="14" height="12" rx="1.5"
                    stroke={t.accent}
                    strokeWidth="1.2"
                    fill="none"
                  />
                  <path
                    d="M4 6l2.5 2L4 10"
                    stroke={t.accent}
                    strokeWidth="1.2"
                    fill="none"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  />
                  <path
                    d="M8 10h4"
                    stroke={t.accent}
                    strokeWidth="1.2"
                    fill="none"
                    strokeLinecap="round"
                  />
                </svg>
              )}
              {f.name}
            </div>
          );
        })}
      </div>

      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 5,
          fontFamily: "'Fira Code', monospace",
          fontSize: 10,
          color: tabInactiveText,
          flexShrink: 0,
        }}
      >
        <span>src</span>
        <span style={{ opacity: 0.4 }}>/</span>
        <span style={{ color: t.accent, opacity: 0.7 }}>{file.name}</span>
      </div>
    </div>
  );

  const statusBar = (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        height: 24,
        padding: "0 12px",
        background: statusBarBg,
        borderTop: `1px solid ${borderColor}`,
        fontFamily: "'Fira Code', monospace",
        fontSize: 10,
        color: tabInactiveText,
        userSelect: "none",
        letterSpacing: "0.02em",
      }}
    >
      <div style={{ display: "flex", alignItems: "center", gap: 14 }}>
        <span style={{ color: t.accent, fontWeight: 500 }}>
          {isTerminal ? "TERMINAL" : file.language === "text" ? "MOD" : file.language.toUpperCase()}
        </span>
        {!isTerminal && <span>UTF-8</span>}
        {isTerminal ? <span>80×24</span> : <span>{activeLines.length} lines</span>}
      </div>
      <div style={{ display: "flex", alignItems: "center", gap: 14 }}>
        {!isTerminal && <span>Spaces: 4</span>}
        <span style={{ color: t.secondary, fontWeight: 500 }}>go-tui</span>
      </div>
    </div>
  );

  /* ─── Terminal view (rendered TUI output matching counter.gsx) ─── */
  function renderTerminalView(maxH: number) {
    const termBg = "#111210";
    const termHr = "#49483e";
    const termDim = "#75715e";
    const termCyan = "#66d9ef";
    const termYellow = "#e6db74";
    const termMagenta = "#ae81ff";
    const termGreen = "#a6e22e";
    const termRed = "#f92672";
    const termBlue = "#6796e6";
    const gradient = "linear-gradient(90deg, #66d9ef, #f92672)";
    const gradientCyanBlue = "linear-gradient(90deg, #66d9ef, #6796e6)";
    const borderGradient =
      "linear-gradient(135deg, #66d9ef, #ae81ff 50%, #f92672)";

    const formatTime = (s: number) =>
      `${Math.floor(s / 60)}:${String(s % 60).padStart(2, "0")}`;

    const statusColor =
      termCount > 0 ? termGreen : termCount < 0 ? termRed : termBlue;
    const statusText =
      termCount > 0 ? "Positive" : termCount < 0 ? "Negative" : "Zero";

    const gradientText = (bg: string): React.CSSProperties => ({
      background: bg,
      WebkitBackgroundClip: "text",
      WebkitTextFillColor: "transparent",
      backgroundClip: "text",
    });

    const hr = (
      <div
        style={{
          color: termHr,
          overflow: "hidden",
          whiteSpace: "nowrap",
          lineHeight: "20px",
        }}
      >
        {"─".repeat(200)}
      </div>
    );

    return (
      <div
        style={{
          padding: "12px 16px",
          height: maxH,
          boxSizing: "border-box",
          display: "flex",
          flexDirection: "column",
        }}
      >
        {/* border-double border-gradient-cyan-magenta */}
        <div
          style={{
            flex: 1,
            minWidth: 0,
            background: borderGradient,
            borderRadius: 7,
            padding: 1.5,
            display: "flex",
          }}
        >
          <div
            style={{
              flex: 1,
              minWidth: 0,
              background: termBg,
              borderRadius: 6,
              padding: 1.5,
              display: "flex",
            }}
          >
            <div
              style={{
                flex: 1,
                minWidth: 0,
                background: borderGradient,
                borderRadius: 5,
                padding: 1.5,
                display: "flex",
              }}
            >
              {/* Content area — flex-col p-1 gap-1 */}
              <div
                style={{
                  flex: 1,
                  minWidth: 0,
                  background: termBg,
                  borderRadius: 4,
                  padding: "8px 10px",
                  fontFamily: "'Fira Code', monospace",
                  fontSize: 13,
                  lineHeight: "24px",
                  display: "flex",
                  flexDirection: "column",
                  gap: 8,
                  overflow: "hidden",
                }}
              >
                {/* Header — flex justify-between items-center */}
                <div
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "center",
                  }}
                >
                  {/* text-gradient-cyan-magenta font-bold */}
                  <span
                    style={{
                      fontWeight: 700,
                      ...gradientText(gradient),
                    }}
                  >
                    Counter
                  </span>
                  {/* @Badge("peak:", ...) @Badge("uptime:", ...) */}
                  <div
                    style={{
                      display: "flex",
                      gap: 16,
                    }}
                  >
                    <div style={{ display: "flex", gap: 8 }}>
                      <span style={{ color: termDim }}>peak:</span>
                      <span style={{ color: termMagenta, fontWeight: 700 }}>
                        {termPeak}
                      </span>
                    </div>
                    <div style={{ display: "flex", gap: 8 }}>
                      <span style={{ color: termDim }}>uptime:</span>
                      <span style={{ color: termYellow, fontWeight: 700 }}>
                        {formatTime(termElapsed)}
                      </span>
                    </div>
                  </div>
                </div>

                {/* <hr /> */}
                {hr}

                {/* Cards — flex gap-2 */}
                <div style={{ display: "flex", gap: 16, minWidth: 0 }}>
                  {/* @Card("Count") — flex-col border-rounded border-cyan p-1 gap-1 flexGrow=1 */}
                  <div
                    style={{
                      flex: 1,
                      minWidth: 0,
                      display: "flex",
                      flexDirection: "column",
                      gap: 8,
                      border: `1px solid ${termCyan}`,
                      borderRadius: 5,
                      padding: "8px 10px",
                    }}
                  >
                    {/* font-bold text-gradient-cyan-blue */}
                    <span
                      style={{
                        fontWeight: 700,
                        ...gradientText(gradientCyanBlue),
                      }}
                    >
                      Count
                    </span>
                    {hr}
                    {/* text-cyan font-bold */}
                    <span style={{ color: termCyan, fontWeight: 700 }}>
                      {termCount}
                    </span>
                  </div>

                  {/* @Card("Status") — flex-col border-rounded border-cyan p-1 gap-1 flexGrow=1 */}
                  <div
                    style={{
                      flex: 1,
                      minWidth: 0,
                      display: "flex",
                      flexDirection: "column",
                      gap: 8,
                      border: `1px solid ${termCyan}`,
                      borderRadius: 5,
                      padding: "8px 10px",
                    }}
                  >
                    {/* font-bold text-gradient-cyan-blue */}
                    <span
                      style={{
                        fontWeight: 700,
                        ...gradientText(gradientCyanBlue),
                      }}
                    >
                      Status
                    </span>
                    {hr}
                    {/* Conditional: text-green/red/blue font-bold */}
                    <span style={{ color: statusColor, fontWeight: 700 }}>
                      {statusText}
                    </span>
                  </div>
                </div>

                {/* Help — flex gap-1 justify-center */}
                <div
                  style={{
                    display: "flex",
                    gap: 8,
                    justifyContent: "center",
                  }}
                >
                  <span style={{ color: termDim }}>
                    +/-count·0 reset
                  </span>
                </div>

                {/* Fill remaining space */}
                <div style={{ flex: 1 }} />

                {/* Disclaimer */}
                <div
                  style={{
                    fontFamily: "'IBM Plex Sans', sans-serif",
                    fontSize: 11,
                    lineHeight: 1.4,
                    color: termDim,
                    textAlign: "center",
                    padding: "6px 0 2px",
                    opacity: 0.7,
                  }}
                >
                  This is a React recreation, run the real thing in{" "}
                  <a
                    href="https://github.com/grindlemire/go-tui/tree/main/examples/docs-example"
                    target="_blank"
                    rel="noopener noreferrer"
                    style={{
                      fontFamily: "'Fira Code', monospace",
                      color: termCyan,
                      opacity: 0.8,
                      textDecoration: "none",
                      borderBottom: `1px solid ${termCyan}40`,
                    }}
                    onMouseEnter={(e) => { e.currentTarget.style.opacity = "1"; e.currentTarget.style.borderBottomColor = termCyan; }}
                    onMouseLeave={(e) => { e.currentTarget.style.opacity = "0.8"; e.currentTarget.style.borderBottomColor = `${termCyan}40`; }}
                  >
                    examples/docs-example
                  </a>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  /* ─── Step pills (shared between wide + narrow) ─── */
  function renderPills(opts: {
    highlightIdx: number | null;
    onToggle: (i: number) => void;
    onHover?: (i: number) => void;
    onLeave?: () => void;
  }) {
    return (
      <div
        style={{
          display: "flex",
          gap: 6,
          marginBottom: 12,
          flexWrap: "wrap",
          /* keep stable height when steps are empty (terminal tab) */
          minHeight: steps.length === 0 ? 29 : undefined,
        }}
      >
        {steps.map((step, i) => {
          const isLit = opts.highlightIdx === i;
          return (
            <button
              key={step.id}
              onClick={() => opts.onToggle(i)}
              onMouseEnter={opts.onHover ? () => opts.onHover!(i) : undefined}
              onMouseLeave={opts.onLeave}
              style={{
                fontFamily: "'Fira Code', monospace",
                fontSize: 11,
                padding: "6px 14px",
                borderRadius: 6,
                border: `1px solid ${isLit ? step.color + "60" : borderColor}`,
                background: isLit ? step.color + "15" : "transparent",
                color: isLit ? step.color : t.textDim,
                fontWeight: isLit ? 600 : 400,
                cursor: "pointer",
                transition: "all 0.2s",
                letterSpacing: "0.02em",
              }}
            >
              {step.label}
            </button>
          );
        })}
      </div>
    );
  }

  /* ─── Step description (shared between wide + narrow) ─── */
  function renderDescription(idx: number | null) {
    const color = idx !== null ? steps[idx].color : null;
    return (
      <div
        style={{
          minHeight: 28,
          marginBottom: 14,
          transition: "opacity 0.25s",
          opacity: idx !== null ? 1 : 0,
        }}
      >
        {idx !== null && (
          <p
            style={{
              fontFamily: "'IBM Plex Sans', sans-serif",
              fontSize: 13,
              lineHeight: 1.5,
              color: t.textMuted,
              margin: 0,
              paddingLeft: 2,
            }}
          >
            <span style={{ color: color!, fontWeight: 600, marginRight: 6 }}>
              {steps[idx].label}
            </span>
            {steps[idx].description}
          </p>
        )}
      </div>
    );
  }

  /* ─── Code line renderer (uses full-file highlighted HTML) ─── */
  function renderCodeLine(
    line: string,
    i: number,
    opts: {
      focusIdx: number | null;
      showAllHighlights: boolean;
      paddingRight: number;
    },
  ) {
    const ln = i + 1;
    const stepIdx = lineToStep.get(ln);
    const isHL = stepIdx !== undefined;
    const color = isHL ? steps[stepIdx].color : null;

    let isActive: boolean;
    let isDimmed: boolean;

    if (opts.focusIdx !== null) {
      isActive = isHL && opts.focusIdx === stepIdx;
      isDimmed = !isActive;
    } else if (opts.showAllHighlights) {
      isActive = isHL;
      isDimmed = false;
    } else {
      isActive = false;
      isDimmed = false;
    }

    const lineHtml = highlightedLines?.[i] ?? "";
    const isEmpty = line.trim() === "";

    return (
      <div
        key={i}
        style={{
          display: "flex",
          fontFamily: "'Fira Code', monospace",
          fontSize: 12,
          lineHeight: `${LINE_H}px`,
          height: LINE_H,
          paddingLeft: 18,
          paddingRight: opts.paddingRight,
          whiteSpace: "pre",
          background: isHL
            ? `${color}${isActive ? "10" : "06"}`
            : "transparent",
          borderLeft: isHL
            ? `2px solid ${color}${isActive ? "" : "30"}`
            : "2px solid transparent",
          opacity: isDimmed ? 0.2 : 1,
          transition: "opacity 0.3s, background 0.3s, border-color 0.3s",
        }}
      >
        {isEmpty ? (
          <span>{"\u00a0"}</span>
        ) : highlightedLines ? (
          <span dangerouslySetInnerHTML={{ __html: lineHtml }} />
        ) : (
          <span style={{ color: t.text }}>{line}</span>
        )}
      </div>
    );
  }

  /* ─── Gutter line renderer ─── */
  function renderGutterLine(
    i: number,
    opts: { focusIdx: number | null; showAllHighlights: boolean },
  ) {
    const ln = i + 1;
    const stepIdx = lineToStep.get(ln);
    const isHL = stepIdx !== undefined;
    const color = isHL ? steps[stepIdx].color : gutterText;

    let isActive: boolean;
    let isDimmed: boolean;

    if (opts.focusIdx !== null) {
      isActive = isHL && opts.focusIdx === stepIdx;
      isDimmed = !isActive;
    } else if (opts.showAllHighlights) {
      isActive = isHL;
      isDimmed = false;
    } else {
      isActive = false;
      isDimmed = false;
    }

    return (
      <div
        key={i}
        style={{
          fontFamily: "'Fira Code', monospace",
          fontSize: 11,
          lineHeight: `${LINE_H}px`,
          height: LINE_H,
          paddingRight: 16,
          paddingLeft: 12,
          color: isActive ? color : gutterText,
          textAlign: "right",
          opacity: isDimmed ? 0.12 : isHL && isActive ? 0.9 : 0.45,
          fontWeight: 400,
          letterSpacing: "0.02em",
          transition: "color 0.3s, opacity 0.3s",
        }}
      >
        {ln}
      </div>
    );
  }

  /* ═══════════════════════════════════════════════════════
   * WIDE LAYOUT — Pills + Code + annotation rail
   * ═══════════════════════════════════════════════════════ */
  if (isWide) {
    return (
      <div style={{ margin: "0 auto" }}>
        <style>{scrollbarCSS}</style>

        {/* ─── Step pills above editor ─── */}
        {renderPills({
          highlightIdx: isTerminal ? null : wideFocusStep,
          onToggle: (i) => { if (!isTerminal) { setActiveStep(activeStep === i ? null : i); scrollToStep(i); } },
          onHover: isTerminal ? undefined : (i) => { setHoveredStep(i); scrollToStep(i); },
          onLeave: isTerminal ? undefined : () => setHoveredStep(null),
        })}

        {/* ─── Step description ─── */}
        {renderDescription(isTerminal ? null : wideFocusStep)}

        {/* ─── Editor frame ─── */}
        <div
          style={{
            position: "relative",
            opacity: revealed ? 1 : 0,
            transform: revealed ? "translateY(0)" : "translateY(12px)",
            transition:
              "opacity 0.6s cubic-bezier(0.16,1,0.3,1), transform 0.6s cubic-bezier(0.16,1,0.3,1)",
          }}
        >
          {/* Atmospheric glow */}
          <div
            style={{
              position: "absolute",
              inset: -1,
              borderRadius: 11,
              background: focusColor
                ? `linear-gradient(135deg, ${focusColor}10, transparent 50%)`
                : `linear-gradient(135deg, ${isDark ? "rgba(102,217,239,0.06)" : "rgba(47,158,184,0.06)"}, transparent 40%, ${isDark ? "rgba(166,226,46,0.04)" : "rgba(99,139,12,0.04)"}, transparent 70%)`,
              transition: "background 0.4s",
              pointerEvents: "none",
            }}
          />

          <div
            style={{
              position: "relative",
              background: t.bgCode,
              border: `1px solid ${borderColor}`,
              borderRadius: 10,
              overflow: "hidden",
              boxShadow: isDark
                ? "0 25px 60px rgba(0,0,0,0.5), 0 8px 20px rgba(0,0,0,0.3), inset 0 1px 0 rgba(255,255,255,0.03)"
                : "0 25px 60px rgba(0,0,0,0.08), 0 8px 20px rgba(0,0,0,0.04), inset 0 1px 0 rgba(255,255,255,0.8)",
            }}
          >
            {titleBar}

            {/* ─── Code + annotation area (or terminal view) ─── */}
            {isTerminal ? renderTerminalView(maxCodeH) : <div
              ref={wideScrollRef}
              className="cshowcase-scroll"
              style={{
                display: "flex",
                maxHeight: maxCodeH,
                overflowY: "auto",
                overflowX: "hidden",
              }}
            >
              {/* Gutter */}
              <div
                style={{
                  display: "flex",
                  flexDirection: "column",
                  alignItems: "flex-end",
                  padding: `${CODE_PAD_Y}px 0`,
                  background: isDark
                    ? "rgba(30,31,27,0.5)"
                    : "rgba(237,237,234,0.4)",
                  borderRight: `1px solid ${isDark ? "rgba(73,72,62,0.25)" : "rgba(216,216,208,0.4)"}`,
                  flexShrink: 0,
                  userSelect: "none",
                  minWidth: 50,
                }}
              >
                {activeLines.map((_, i) =>
                  renderGutterLine(i, {
                    focusIdx: wideFocusStep,
                    showAllHighlights: hasSteps,
                  }),
                )}
              </div>

              {/* Code lines */}
              <div
                style={{
                  flex: 1,
                  minWidth: 0,
                  padding: `${CODE_PAD_Y}px 0`,
                }}
              >
                {activeLines.map((line, i) =>
                  renderCodeLine(line, i, {
                    focusIdx: wideFocusStep,
                    showAllHighlights: hasSteps,
                    paddingRight: 24,
                  }),
                )}
              </div>

              {/* Annotation rail */}
              {hasSteps && <div
                style={{
                  width: ANNOTATION_W,
                  flexShrink: 0,
                  position: "relative",
                  borderLeft: `1px solid ${isDark ? "rgba(73,72,62,0.25)" : "rgba(216,216,208,0.4)"}`,
                  background: isDark
                    ? "rgba(26,27,23,0.4)"
                    : "rgba(245,245,241,0.5)",
                }}
              >
                {steps.map((step, i) => {
                  const isHovered = hoveredStep === i;
                  const isClicked = activeStep === i;
                  const isFocused = isHovered || isClicked;
                  const isDimmedCard =
                    wideFocusStep !== null && wideFocusStep !== i;

                  return (
                    <div
                      key={step.id}
                      onMouseEnter={() => {
                        setActiveStep(null);
                        setHoveredStep(i);
                      }}
                      onMouseLeave={() => setHoveredStep(null)}
                      onClick={() =>
                        setActiveStep(activeStep === i ? null : i)
                      }
                      style={{
                        position: "absolute",
                        top: annotationYs[i],
                        left: 8,
                        right: 8,
                        padding: "8px 12px",
                        borderRadius: 6,
                        borderLeft: `2px solid ${step.color}${isFocused ? "" : isDimmedCard ? "20" : "60"}`,
                        background: isFocused
                          ? isDark
                            ? `${step.color}14`
                            : `${step.color}0c`
                          : "transparent",
                        opacity: isDimmedCard ? 0.25 : 1,
                        transition:
                          "all 0.3s cubic-bezier(0.16,1,0.3,1)",
                        cursor: "pointer",
                      }}
                    >
                      {/* Step label */}
                      <div
                        style={{
                          display: "flex",
                          alignItems: "baseline",
                          gap: 6,
                          marginBottom: 4,
                        }}
                      >
                        <span
                          style={{
                            fontFamily: "'Fira Code', monospace",
                            fontSize: 9,
                            fontWeight: 500,
                            color: step.color,
                            opacity: isFocused ? 0.7 : 0.4,
                            letterSpacing: "0.04em",
                            transition: "opacity 0.3s",
                          }}
                        >
                          {String(i + 1).padStart(2, "0")}
                        </span>
                        <span
                          style={{
                            fontFamily: "'Fira Code', monospace",
                            fontSize: 11,
                            fontWeight: 600,
                            color: isFocused ? step.color : isDark ? t.textMuted : t.textDim,
                            letterSpacing: "0.01em",
                            transition: "color 0.3s",
                          }}
                        >
                          {step.label}
                        </span>
                      </div>

                      {/* Description */}
                      <div
                        style={{
                          fontFamily: "'IBM Plex Sans', sans-serif",
                          fontSize: 11,
                          lineHeight: 1.55,
                          color: isFocused ? t.textMuted : t.textDim,
                          opacity: isFocused ? 1 : 0.7,
                          transition: "color 0.3s, opacity 0.3s",
                        }}
                      >
                        {step.description}
                      </div>
                    </div>
                  );
                })}
              </div>}
            </div>}

            {statusBar}
          </div>
        </div>
      </div>
    );
  }

  /* ═══════════════════════════════════════════════════════
   * NARROW LAYOUT — Step pills + code (mobile / tablet)
   * ═══════════════════════════════════════════════════════ */

  const narrowActiveLineSet =
    hasSteps && activeStep !== null && activeStep < steps.length ? new Set(steps[activeStep].lines) : null;
  const narrowActiveColor =
    hasSteps && activeStep !== null && activeStep < steps.length ? steps[activeStep].color : null;
  const narrowMaxH = 480;

  return (
    <div style={{ maxWidth: 880, margin: "0 auto" }}>
      <style>{scrollbarCSS}</style>

      {/* ─── Step pills ─── */}
      {renderPills({
        highlightIdx: isTerminal ? null : activeStep,
        onToggle: (i) => { if (!isTerminal) { setActiveStep(activeStep === i ? null : i); scrollToStep(i); } },
      })}

      {/* ─── Step description ─── */}
      {renderDescription(isTerminal ? null : activeStep)}

      {/* ─── Editor frame ─── */}
      <div
        style={{
          position: "relative",
          opacity: revealed ? 1 : 0,
          transform: revealed ? "translateY(0)" : "translateY(12px)",
          transition:
            "opacity 0.6s cubic-bezier(0.16,1,0.3,1), transform 0.6s cubic-bezier(0.16,1,0.3,1)",
        }}
      >
        {/* Atmospheric glow */}
        <div
          style={{
            position: "absolute",
            inset: -1,
            borderRadius: 11,
            background: narrowActiveColor
              ? `linear-gradient(135deg, ${narrowActiveColor}08, transparent 50%)`
              : `linear-gradient(135deg, ${isDark ? "rgba(102,217,239,0.06)" : "rgba(47,158,184,0.06)"}, transparent 40%, ${isDark ? "rgba(166,226,46,0.04)" : "rgba(99,139,12,0.04)"}, transparent 70%)`,
            transition: "background 0.4s",
            pointerEvents: "none",
          }}
        />

        <div
          style={{
            position: "relative",
            background: t.bgCode,
            border: `1px solid ${borderColor}`,
            borderRadius: 10,
            overflow: "hidden",
            boxShadow: isDark
              ? "0 25px 60px rgba(0,0,0,0.5), 0 8px 20px rgba(0,0,0,0.3), inset 0 1px 0 rgba(255,255,255,0.03)"
              : "0 25px 60px rgba(0,0,0,0.08), 0 8px 20px rgba(0,0,0,0.04), inset 0 1px 0 rgba(255,255,255,0.8)",
          }}
        >
          {titleBar}

          {/* ─── Code area (or terminal view) ─── */}
          {isTerminal ? renderTerminalView(narrowMaxH) : <div
            style={{
              display: "flex",
              position: "relative",
              maxHeight: narrowMaxH,
            }}
          >
            {/* Gutter */}
            <div
              ref={gutterScrollRef}
              style={{
                display: "flex",
                flexDirection: "column",
                alignItems: "flex-end",
                padding: `${CODE_PAD_Y}px 0`,
                background: gutterBg,
                borderRight: `1px solid ${gutterBorder}`,
                flexShrink: 0,
                userSelect: "none",
                minWidth: 48,
                overflow: "hidden",
              }}
            >
              {activeLines.map((_, i) => {
                const ln = i + 1;
                const isHL =
                  narrowActiveLineSet !== null && narrowActiveLineSet.has(ln);
                const isDimmedLine =
                  narrowActiveLineSet !== null && !narrowActiveLineSet.has(ln);
                return (
                  <div
                    key={i}
                    style={{
                      fontFamily: "'Fira Code', monospace",
                      fontSize: 12,
                      lineHeight: `${LINE_H}px`,
                      height: LINE_H,
                      paddingRight: 14,
                      paddingLeft: 10,
                      color: isHL ? narrowActiveColor! : gutterText,
                      textAlign: "right",
                      opacity: isDimmedLine ? 0.25 : 1,
                      fontWeight: isHL ? 600 : 400,
                      transition: "color 0.25s, opacity 0.25s",
                    }}
                  >
                    {ln}
                  </div>
                );
              })}
            </div>

            {/* Code lines */}
            <div
              ref={codeScrollRef}
              onScroll={onCodeScroll}
              className="cshowcase-scroll"
              style={{
                flex: 1,
                overflowY: "auto",
                overflowX: "auto",
                padding: `${CODE_PAD_Y}px 0`,
                position: "relative",
                minWidth: 0,
              }}
            >
              <div style={{ minWidth: "fit-content" }}>
              {activeLines.map((line, i) => {
                const ln = i + 1;
                const isHL =
                  narrowActiveLineSet !== null && narrowActiveLineSet.has(ln);
                const isDimmedLine =
                  narrowActiveLineSet !== null && !narrowActiveLineSet.has(ln);

                const lineHtml = highlightedLines?.[i] ?? "";
                const isEmpty = line.trim() === "";

                return (
                  <div
                    key={i}
                    style={{
                      display: "flex",
                      fontFamily: "'Fira Code', monospace",
                      fontSize: 12,
                      lineHeight: `${LINE_H}px`,
                      height: LINE_H,
                      paddingLeft: 18,
                      paddingRight: 70,
                      whiteSpace: "pre",
                      opacity: isDimmedLine ? 0.25 : 1,
                      background: isHL
                        ? `${narrowActiveColor}0c`
                        : "transparent",
                      borderLeft: isHL
                        ? `2px solid ${narrowActiveColor}`
                        : "2px solid transparent",
                      transition:
                        "opacity 0.25s, background 0.25s, border-color 0.25s",
                    }}
                  >
                    {isEmpty ? (
                      <span>{"\u00a0"}</span>
                    ) : highlightedLines ? (
                      <span
                        dangerouslySetInnerHTML={{ __html: lineHtml }}
                      />
                    ) : (
                      <span style={{ color: t.text }}>{line}</span>
                    )}
                  </div>
                );
              })}
              </div>
            </div>
          </div>}

          {statusBar}
        </div>
      </div>
    </div>
  );
}
