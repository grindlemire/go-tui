import { useState, useEffect, useRef } from "react";
import { palette, useTheme } from "../lib/theme.ts";
import CodeShowcase from "./CodeShowcase.tsx";
import Divider from "./Divider.tsx";
import PageBackground from "./PageBackground.tsx";
import DxCapability from "./DxCapability.tsx";
import EditorSimulation from "./EditorSimulation.tsx";
import ComparisonSection, { type ComparisonFeature } from "./ComparisonSection.tsx";

/* ─── Shared sub-components ─── */

function useValueProps() {
  const { theme } = useTheme();
  const t = palette[theme];
  const isDark = theme === "dark";
  return [
    {
      tag: ".gsx",
      label: "templ-inspired syntax",
      summary: "HTML-like templates that compile to type-safe Go",
      detail: ".gsx files with if, for, := control flow intermingled with regular Go code that compile to type-safe Go.",
      color: t.tertiary,
    },
    {
      tag: "tw",
      label: "Tailwind for terminals",
      summary: "Utility classes compile to Go layout options",
      detail: "Familiar patterns: border-rounded, p-2, text-cyan, flex-col, gap-1. Classes compile directly to Go option functions at build time.",
      color: isDark ? "#e6db74" : "#998a00",
    },
    {
      tag: "lsp",
      label: "Editor tooling",
      summary: "LSP, tree-sitter grammar, auto-formatter",
      detail: "Language server with completions, inline diagnostics, go-to-definition across .gsx and .go files. Tree-sitter grammar for syntax highlighting.",
      color: isDark ? "#ae81ff" : "#7c5cb8",
    },
    {
      tag: "state",
      label: "Reactive State[T]",
      summary: "Generic state with Bind() callbacks",
      detail: "State[T] with Bind() callbacks that trigger re-renders. Batch() coalesces multiple updates into a single frame.",
      color: t.secondary,
    },
    {
      tag: "layout",
      label: "Pure Go flexbox",
      summary: "Full layout engine, pure Go",
      detail: "Row, column, justify, align, gap, padding, margin, min/max constraints, percentage and auto sizing. Cross-compiles to any target.",
      color: t.accent,
    },
  ] as const;
}

/* ─── Comparison Features (HomePageExplore variant) ─── */

const homeComparisonFeatures: ComparisonFeature[] = [
  {
    label: "Approach",
    values: {
      "go-tui": { summary: "Declarative .gsx templates", detail: ".gsx files use HTML-like syntax with Tailwind-style classes and compile to type-safe Go via tui generate" },
      "Bubble Tea": { summary: "Elm architecture", detail: "Functional Model → Update → View cycle. State changes are centralized in Update(), messages drive updates, View returns a string" },
      tview: { summary: "Imperative widget toolkit", detail: "OOP style: create widget objects, configure via methods, compose in layout containers" },
      gocui: { summary: "View manager", detail: "Create named rectangular views with absolute coordinates. Views implement io.ReadWriter for content" },
    },
  },
  {
    label: "Layout",
    values: {
      "go-tui": { summary: "CSS flexbox", detail: "Full flexbox: grow, shrink, justify, align, gap, padding, margin, min/max constraints, percentage and auto sizing" },
      "Bubble Tea": { summary: "String joins via lipgloss", detail: "lipgloss provides box model styling and JoinHorizontal/JoinVertical for composition. No flexbox" },
      tview: { summary: "Basic Flex and Grid", detail: "Flex supports direction and proportional sizing. Grid adds row/column spans and gap. No align-items" },
      gocui: { summary: "Manual coordinates", detail: "Views positioned with absolute (x0, y0, x1, y1) coordinates. Responsive sizing requires manual calculation" },
    },
  },
  {
    label: "Widgets",
    values: {
      "go-tui": { summary: "HTML-style primitives", detail: "Built-in: div, span, p, ul, li, button, input, table, progress, hr, br. Composable via .gsx components" },
      "Bubble Tea": { summary: "14+ via Bubbles", detail: "Separate Bubbles library: text input, viewport, list, table, spinner, progress, file picker, and more" },
      tview: { summary: "16+ built-in", detail: "Richest widget set: TextView, Table, TreeView, List, Form, Modal, InputField, DropDown, and more" },
      gocui: { summary: "Views only", detail: "No pre-built widgets. Views provide text I/O and keybindings, so widgets must be built from scratch" },
    },
  },
  {
    label: "State",
    values: {
      "go-tui": { summary: "Reactive State[T]", detail: "Generic State[T] with Bind() callbacks triggers re-render on change. Batch() coalesces updates" },
      "Bubble Tea": { summary: "Elm update cycle", detail: "Messages flow through Update() returning a new model and optional commands" },
      tview: { summary: "Manual redraw", detail: "Mutate widget state via setter methods, then call app.Draw() or app.QueueUpdateDraw()" },
      gocui: { summary: "Manual redraw", detail: "Write content directly to views. Manage goroutines manually for async updates" },
    },
  },
  {
    label: "Styling",
    values: {
      "go-tui": { summary: "Tailwind classes", detail: "Utility classes (border-rounded, p-2, text-cyan) compile to Go options. True color support" },
      "Bubble Tea": { summary: "lipgloss API", detail: "Fluent Go API: lipgloss.NewStyle().Bold(true).Foreground(...). Rich but verbose for complex layouts" },
      tview: { summary: "tcell styles + tags", detail: "Inline [color]text[/] tags in strings. Styles applied via tcell.Style" },
      gocui: { summary: "Basic ANSI", detail: "Attribute type with 8 named colors and bold/underline/reverse via termbox-go. 256-color requires raw ANSI escapes" },
    },
  },
  {
    label: "Tooling",
    values: {
      "go-tui": { summary: "LSP + tree-sitter + fmt", detail: "Language server with completions, diagnostics, go-to-definition. Tree-sitter grammar. Auto-formatter" },
      "Bubble Tea": { summary: "Standard Go", detail: "Standard Go tooling applies. No custom language features needed" },
      tview: { summary: "Standard Go", detail: "Standard Go tooling. Demo application included for exploration" },
      gocui: { summary: "Standard Go", detail: "Standard Go tooling. ~20 examples in repository covering layout, colors, mouse, goroutines, and more" },
    },
  },
];

/* ─── Tooling Section ─── */

function ToolingSection() {
  const { theme } = useTheme();
  const t = palette[theme];
  const [dxFeature, setDxFeature] = useState(0);
  const dxPausedRef = useRef(false);

  useEffect(() => {
    const interval = setInterval(() => {
      if (!dxPausedRef.current) setDxFeature((prev) => (prev + 1) % 5);
    }, 4000);
    return () => clearInterval(interval);
  }, []);

  const capabilities = [
    { title: "Syntax highlighting", description: "Tree-sitter grammar with distinct tokens for keywords, elements, Go, and Tailwind classes.", color: t.accent, editorIdx: 0 },
    { title: "Code completions", description: "Component suggestions with type signatures as you type.", color: t.secondary, editorIdx: 1 },
    { title: "Inline diagnostics", description: "See errors in your editor before you compile.", color: t.tertiary, editorIdx: 2 },
    { title: "Go-to-definition", description: "Jump to definitions across .gsx and Go files.", color: theme === "dark" ? "#e6db74" : "#998a00", editorIdx: 3 },
    { title: "Auto-formatting", description: "Indentation, alignment, and imports. On save or via CLI.", color: theme === "dark" ? "#ae81ff" : "#7c5cb8", editorIdx: 4 },
  ] as const;

  return (
    <section className="max-w-[1100px] mx-auto px-4 sm:px-6 py-10 sm:py-12">
      <div className="flex items-center gap-3 mb-3">
        <div className="font-['Fira_Code',monospace] text-[10px] tracking-[0.2em] uppercase" style={{ color: t.tertiaryDim }}>developer experience</div>
        <div className="h-px flex-1" style={{ background: theme === "dark" ? "linear-gradient(to right, #f9267218, transparent)" : `linear-gradient(to right, ${t.border}, transparent)` }} />
      </div>
      <h2 className="text-2xl sm:text-3xl font-bold tracking-tight mb-3" style={{ color: t.heading }}>Built-in editor tooling</h2>
      <p className="text-[14px] sm:text-[15px] mb-4 max-w-[600px]" style={{ color: t.textMuted }}>
        Your editor knows .gsx. Completions, diagnostics, and go-to-definition work out of the box.
      </p>
      <div className="flex flex-wrap items-center gap-3 mb-8 sm:mb-10 font-['Fira_Code',monospace] text-[12px]">
        <a
          href="https://marketplace.visualstudio.com/items?itemName=grindlemire.go-tui"
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md transition-colors duration-150"
          style={{ color: t.accent, background: `${t.accent}10`, border: `1px solid ${t.accent}25` }}
        >
          VS Code Marketplace
          <svg width="10" height="10" viewBox="0 0 10 10" fill="none" style={{ opacity: 0.6 }}><path d="M3 1h6v6M9 1L1 9" stroke="currentColor" strokeWidth="1.2" strokeLinecap="round" strokeLinejoin="round" /></svg>
        </a>
        <a
          href="https://open-vsx.org/extension/grindlemire/go-tui"
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md transition-colors duration-150"
          style={{ color: t.secondary, background: `${t.secondary}10`, border: `1px solid ${t.secondary}25` }}
        >
          Open VSX
          <svg width="10" height="10" viewBox="0 0 10 10" fill="none" style={{ opacity: 0.6 }}><path d="M3 1h6v6M9 1L1 9" stroke="currentColor" strokeWidth="1.2" strokeLinecap="round" strokeLinejoin="round" /></svg>
        </a>
      </div>

      <div className="grid lg:grid-cols-[1fr_340px] gap-5 sm:gap-8 items-stretch">
        <EditorSimulation
          activeFeature={dxFeature}
          onSetFeature={(i) => { setDxFeature(i); dxPausedRef.current = true; }}
          pausedRef={dxPausedRef}
        />
        <div className="flex flex-col gap-1 lg:gap-0 lg:justify-between">
          {capabilities.map((cap, i) => (
            <DxCapability
              key={cap.title}
              title={cap.title}
              description={cap.description}
              color={cap.color}
              delay={i * 60}
              active={dxFeature === cap.editorIdx}
              onHover={() => { dxPausedRef.current = true; setDxFeature(cap.editorIdx); }}
              onLeave={() => { dxPausedRef.current = false; }}
            />
          ))}
        </div>
      </div>
    </section>
  );
}

/* ─── Value Prop Carousel ─── */

function ValuePropCarousel() {
  const { theme } = useTheme();
  const t = palette[theme];
  const props = useValueProps();
  const [active, setActive] = useState(0);
  const pausedRef = useRef(false);

  useEffect(() => {
    const id = setInterval(() => {
      if (!pausedRef.current) setActive((p) => (p + 1) % props.length);
    }, 3500);
    return () => clearInterval(id);
  }, [props.length]);

  return (
    <div
      className="relative overflow-hidden"
      onMouseEnter={() => { pausedRef.current = true; }}
      onMouseLeave={() => { pausedRef.current = false; }}
    >
      {/* Indicator dots */}
      <div className="flex items-center gap-2 mb-4 sm:mb-3">
        {props.map((p, i) => (
          <button
            key={p.tag}
            onClick={() => { setActive(i); pausedRef.current = true; }}
            className="cursor-pointer rounded-full"
            style={{
              width: active === i ? 24 : 10,
              height: 10,
              background: active === i ? p.color : t.textDim,
              opacity: active === i ? 1 : 0.25,
              border: "none",
              padding: 0,
              transition: "all 0.35s cubic-bezier(0.4, 0, 0.2, 1)",
            }}
          />
        ))}
      </div>

      {/* Sliding content — taller on mobile for text wrap */}
      <div className="relative carousel-content">
        {props.map((p, i) => (
          <div
            key={p.tag}
            className="absolute inset-0 flex flex-col justify-start"
            style={{
              opacity: active === i ? 1 : 0,
              transform: active === i ? "translateY(0)" : "translateY(6px)",
              transition: "opacity 0.35s ease, transform 0.35s ease",
              pointerEvents: active === i ? "auto" : "none",
            }}
          >
            <span
              className="font-['Fira_Code',monospace] text-[12px] sm:text-[12px] font-bold"
              style={{ color: p.color }}
            >
              {p.label}
            </span>
            <p
              className="text-[13px] mt-1.5 sm:mt-1 leading-[1.55] max-w-[520px]"
              style={{ color: t.textMuted, fontFamily: "'IBM Plex Sans', sans-serif" }}
            >
              {p.detail}
            </p>
          </div>
        ))}
      </div>
    </div>
  );
}

/* ─── Homepage ─── */

export default function HomePageExplore() {
  const { theme } = useTheme();
  const t = palette[theme];

  return (
    <>
      <style>{`
        @keyframes vSlideUp {
          from { opacity: 0; transform: translateY(6px); }
          to { opacity: 1; transform: translateY(0); }
        }
        .v-in { opacity: 0; animation: vSlideUp 0.35s ease-out forwards; }

        @keyframes seeItNudge {
          0%, 100% { transform: translateX(0); }
          50% { transform: translateX(2px); }
        }
        .see-it-cursor { animation: seeItNudge 2.5s ease-in-out infinite; }

        .carousel-content { height: 90px; }
        @media (min-width: 640px) { .carousel-content { height: 60px; } }

        @keyframes fadeInUp {
          from { opacity: 0; transform: translateY(10px); }
          to { opacity: 1; transform: translateY(0); }
        }

        @keyframes syntaxPulse {
          0%, 100% { filter: brightness(1); }
          50% { filter: brightness(1.4); }
        }
        .syntax-active-token {
          animation: syntaxPulse 2s ease-in-out infinite;
        }
      `}</style>

      <div className="relative">
        <PageBackground theme={theme} />
        <div className="relative z-10">
          <section className="max-w-[1100px] mx-auto px-4 sm:px-6" style={{ paddingTop: "clamp(28px, 5vh, 56px)" }}>
            <div className="v-in" style={{ animationDelay: "0ms" }}>
              <h1 className="leading-[1.08] tracking-tight font-bold" style={{ color: t.heading, fontSize: "clamp(24px, 5vw, 42px)", margin: 0 }}>
                Declarative terminal UIs in <span style={{ color: t.tertiary }}>Go</span>
              </h1>
            </div>

            {/* Value prop carousel */}
            <div className="v-in mt-5 sm:mt-5" style={{ animationDelay: "50ms" }}>
              <ValuePropCarousel />
            </div>

            {/* Inline "see it in action" nudge + CodeShowcase */}
            <div className="v-in mt-5 sm:mt-6" style={{ animationDelay: "90ms" }}>
              <div className="flex items-center gap-2 mb-3">
                <svg width="12" height="12" viewBox="0 0 14 14" fill="none" className="see-it-cursor" style={{ opacity: 0.45 }}>
                  <path d="M2 1l8.5 5L7 7.5l-1 4.5L2 1z" fill={t.accent} fillOpacity="0.6" stroke={t.accent} strokeWidth="0.8" />
                </svg>
                <span className="font-['Fira_Code',monospace] text-[10px] tracking-[0.1em] uppercase" style={{ color: t.textDim }}>
                  Hover the annotations to explore
                </span>
              </div>
              <CodeShowcase />
            </div>
          </section>
          <Divider />
          <ComparisonSection features={homeComparisonFeatures} />
          <Divider />
          <ToolingSection />
        </div>
      </div>
    </>
  );
}
