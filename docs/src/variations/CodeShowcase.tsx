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
\tdisplay *tui.Ref
}

func Counter() *counterApp {
\treturn &counterApp{
\t\tcount:   tui.NewState(0),
\t\telapsed: tui.NewState(0),
\t\tdisplay: tui.NewRef(),
\t}
}

func (c *counterApp) KeyMap() tui.KeyMap {
\treturn tui.KeyMap{
\t\ttui.OnRune('+', func(ke tui.KeyEvent) {
\t\t\tc.count.Update(func(v int) int { return v + 1 })
\t\t}),
\t\ttui.OnRune('-', func(ke tui.KeyEvent) {
\t\t\tc.count.Update(func(v int) int { return v - 1 })
\t\t}),
\t\ttui.OnRune('r', func(ke tui.KeyEvent) { c.count.Set(0) }),
\t\ttui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
\t}
}

func (c *counterApp) Watchers() []tui.Watcher {
\treturn []tui.Watcher{
\t\ttui.OnTimer(time.Second, func() {
\t\t\tc.elapsed.Update(func(v int) int { return v + 1 })
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
\t\t\t@Badge("uptime:", formatTime(c.elapsed.Get()), "text-yellow")
\t\t</div>
\t\t<hr />
\t\t<div class="flex gap-2">
\t\t\t@Card("Count") {
\t\t\t\t<span ref={c.display} class="text-cyan font-bold">
\t\t\t\t\t{fmt.Sprintf("%d", c.count.Get())}
\t\t\t\t</span>
\t\t\t}
\t\t\t@Card("Status") {
\t\t\t\t@if c.count.Get() > 0 {
\t\t\t\t\t<span class="text-green font-bold">Positive</span>
\t\t\t\t} @else @if c.count.Get() < 0 {
\t\t\t\t\t<span class="text-red font-bold">Negative</span>
\t\t\t\t} @else {
\t\t\t\t\t<span class="text-blue font-bold">Zero</span>
\t\t\t\t}
\t\t\t}
\t\t</div>
\t\t<div class="flex gap-1 justify-center">
\t\t\t<span class="font-dim">+/- count \u00b7 r reset \u00b7 q quit</span>
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
    label: "State & Refs",
    description:
      "Reactive State[T] values and Refs live on the component struct. The constructor initializes them with NewState and NewRef.",
    lines: [9, 10, 11, 12, 13, 15, 16, 17, 18, 19, 20, 21],
  },
  {
    id: "events",
    label: "Keyboard Events",
    description:
      "KeyMap binds keys to actions. OnRune matches character keys \u2014 Update and Set mutate state and trigger re-renders.",
    lines: [23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34],
  },
  {
    id: "watchers",
    label: "Watchers",
    description:
      "Background tasks that run independently. OnTimer fires a callback at a fixed interval \u2014 here it ticks once per second.",
    lines: [36, 37, 38, 39, 40, 41, 42],
  },
  {
    id: "components",
    label: "Components",
    description:
      "Reusable templ functions that take props and return elements. Card accepts {children...} as a content slot.",
    lines: [48, 49, 50, 51, 52, 53, 55, 56, 57, 58, 59, 60, 61],
  },
  {
    id: "template",
    label: "Composition",
    description:
      "The Render template ties it together: @-calls nest components, ref targets elements, and @if/@else adds conditionals.",
    lines: [
      63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79,
      80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90,
    ],
  },
];

const mainGoStepDefs: Omit<Step, "color">[] = [
  {
    id: "entry",
    label: "Entry point",
    description:
      "Pass your component to WithRootComponent \u2014 the framework owns the event loop, rendering, and cleanup.",
    lines: [10, 11, 12],
  },
];

/* Colors per step per file, indexed by theme */
const gsxStepColors = {
  dark: ["#66d9ef", "#a6e22e", "#e6db74", "#ae81ff", "#f92672"],
  light: ["#2f9eb8", "#638b0c", "#998a00", "#6e5dc6", "#d42568"],
};

const mainGoStepColors = {
  dark: ["#66d9ef"],
  light: ["#2f9eb8"],
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
  const wideScrollRef = useRef<HTMLDivElement>(null);
  const codeScrollRef = useRef<HTMLDivElement>(null);
  const gutterScrollRef = useRef<HTMLDivElement>(null);

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

  const file = files[activeFile];
  const isGsx = activeFile === 0;
  const activeCode = file.code;
  const activeLines = activeCode.split("\n");

  /* ─── Full-code syntax highlighting ───
   * Highlight the entire file at once so shiki has full grammar context
   * (function bodies, struct fields, etc.), then split into per-line HTML. */
  const highlightedLines = useMemo(() => {
    if (!ready) return null;
    if (file.language === "text") return null;
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

  /* Scroll to highlighted region when activeStep changes */
  useEffect(() => {
    if (activeStep === null) return;
    if (activeStep >= stepDefs.length) return;
    const scrollEl = isWide ? wideScrollRef.current : codeScrollRef.current;
    if (!scrollEl) return;
    const step = stepDefs[activeStep];
    const firstLine = Math.min(...step.lines);
    const lastLine = Math.max(...step.lines);
    const midLine = (firstLine + lastLine) / 2;
    const targetScroll =
      (midLine - 1) * LINE_H - scrollEl.clientHeight / 2 + CODE_PAD_Y;
    scrollEl.scrollTo({
      top: Math.max(0, targetScroll),
      behavior: "smooth",
    });
  }, [activeStep, isWide, activeFile]);

  /* Theme palette */
  const borderColor = isDark ? "#49483e" : "#d8d8d0";
  const gutterBg = isDark ? "#1e1f1b" : "#ededea";
  const gutterBorder = isDark
    ? "rgba(73,72,62,0.4)"
    : "rgba(216,216,208,0.6)";
  const gutterText = isDark ? "#5c5c50" : "#b8b8ad";
  const titleBarBg = isDark ? "#1e1f1b" : "#ededea";
  const tabActiveBg = isDark ? "#23241e" : "#f5f5f1";
  const tabActiveText = isDark ? "#f8f8f2" : "#49483e";
  const tabInactiveText = isDark ? "#5c5c50" : "#a6a68a";
  const dotColors = isDark
    ? ["#ff5f57", "#febc2e", "#28c840"]
    : ["#ff6159", "#ffbf2f", "#2bc840"];
  const annotationRailBg = isDark
    ? "rgba(26,27,23,0.6)"
    : "rgba(240,240,236,0.6)";
  const statusBarBg = isDark ? "#1a1b17" : "#e8e8e3";

  const maxCodeH = 580;

  /* In wide mode: hover overrides click for spotlight */
  const wideFocusStep = hoveredStep ?? activeStep;
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
          {file.language === "text" ? "MOD" : file.language.toUpperCase()}
        </span>
        <span>UTF-8</span>
        <span>{activeLines.length} lines</span>
      </div>
      <div style={{ display: "flex", alignItems: "center", gap: 14 }}>
        <span>Spaces: 4</span>
        <span style={{ color: t.secondary, fontWeight: 500 }}>go-tui</span>
      </div>
    </div>
  );

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
          fontSize: 12,
          lineHeight: `${LINE_H}px`,
          height: LINE_H,
          paddingRight: 14,
          paddingLeft: 10,
          color: isActive ? color : gutterText,
          textAlign: "right",
          opacity: isDimmed ? 0.15 : isHL && isActive ? 1 : 0.6,
          fontWeight: isActive && isHL ? 600 : 400,
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
          highlightIdx: wideFocusStep,
          onToggle: (i) => setActiveStep(activeStep === i ? null : i),
          onHover: (i) => setHoveredStep(i),
          onLeave: () => setHoveredStep(null),
        })}

        {/* ─── Step description ─── */}
        {renderDescription(wideFocusStep)}

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

            {/* ─── Code + annotation area ─── */}
            <div
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
                  background: gutterBg,
                  borderRight: `1px solid ${gutterBorder}`,
                  flexShrink: 0,
                  userSelect: "none",
                  minWidth: 48,
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
                  borderLeft: `1px solid ${gutterBorder}`,
                  background: annotationRailBg,
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
                        left: 0,
                        right: 0,
                        padding: "10px 14px 10px 16px",
                        borderLeft: `3px solid ${step.color}${isFocused ? "" : isDimmedCard ? "30" : "80"}`,
                        background: isFocused
                          ? `${step.color}18`
                          : `${step.color}08`,
                        opacity: isDimmedCard ? 0.2 : 1,
                        transition:
                          "all 0.3s cubic-bezier(0.16,1,0.3,1)",
                        cursor: "pointer",
                      }}
                    >
                      {/* Step number + label */}
                      <div
                        style={{
                          display: "flex",
                          alignItems: "center",
                          gap: 7,
                          marginBottom: 5,
                        }}
                      >
                        <span
                          style={{
                            fontFamily: "'Fira Code', monospace",
                            fontSize: 9,
                            fontWeight: 600,
                            color: step.color,
                            opacity: isFocused ? 1 : 0.6,
                            letterSpacing: "0.05em",
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
                            color: step.color,
                            letterSpacing: "0.02em",
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
                          lineHeight: 1.5,
                          color: isFocused ? t.textMuted : t.textDim,
                          transition: "color 0.3s",
                        }}
                      >
                        {step.description}
                      </div>
                    </div>
                  );
                })}
              </div>}
            </div>

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
        highlightIdx: activeStep,
        onToggle: (i) => setActiveStep(activeStep === i ? null : i),
      })}

      {/* ─── Step description ─── */}
      {renderDescription(activeStep)}

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

          {/* ─── Code area ─── */}
          <div
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

          {statusBar}
        </div>
      </div>
    </div>
  );
}
