import { useState, useEffect, useRef } from "react";
import { palette, useTheme } from "../lib/theme.ts";
import CodeShowcase from "./CodeShowcase.tsx";
import Divider from "./Divider.tsx";
import PageBackground from "./PageBackground.tsx";

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
      detail: ".gsx files with @if, @for, @let control flow intermingled with regular Go code that compile to type-safe Go.",
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

/* ─── Comparison Section ─── */

const comparisonLibraries = ["go-tui", "Bubble Tea", "tview", "gocui"] as const;

type ComparisonFeature = {
  label: string;
  values: Record<string, { summary: string; detail: string }>;
};

const comparisonFeatures: ComparisonFeature[] = [
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
      tview: { summary: "15+ built-in", detail: "Richest widget set: TextView, Table, TreeView, List, Form, Modal, InputField, DropDown, and more" },
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

function ComparisonDetailPanel({ feature, expanded }: { feature: ComparisonFeature; expanded: boolean }) {
  const { theme } = useTheme();
  const t = palette[theme];
  return (
    <div style={{ maxHeight: expanded ? 300 : 0, opacity: expanded ? 1 : 0, overflow: "hidden", transition: "max-height 0.3s ease, opacity 0.2s ease" }}>
      <div className="grid gap-0" style={{ gridTemplateColumns: "140px repeat(4, minmax(0, 1fr))", background: theme === "dark" ? "rgba(255,255,255,0.015)" : "rgba(0,0,0,0.01)", borderTop: `1px solid ${t.border}` }}>
        <div className="px-4 py-3" />
        {comparisonLibraries.map((lib, colIdx) => {
          const val = feature.values[lib];
          const isGoTui = lib === "go-tui";
          return (
            <div key={lib} className="px-4 py-3 min-w-0" style={{ borderLeft: isGoTui ? `1px solid ${t.accent}30` : `1px solid ${t.border}`, borderRight: isGoTui && colIdx < comparisonLibraries.length - 1 ? `1px solid ${t.accent}30` : undefined, background: isGoTui ? (theme === "dark" ? `${t.accent}14` : `${t.accent}0c`) : "transparent" }}>
              <p className="text-[11px] leading-[1.6] m-0" style={{ color: t.textDim, fontFamily: "'IBM Plex Sans', sans-serif", overflowWrap: "break-word" }}>{val.detail}</p>
            </div>
          );
        })}
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
  const libColors: Record<string, string> = { "go-tui": t.accent, "Bubble Tea": t.secondary, tview: theme === "dark" ? "#e6db74" : "#998a00", gocui: t.tertiary };

  useEffect(() => {
    const el = sectionRef.current;
    if (!el) return;
    const observer = new IntersectionObserver(([entry]) => { if (entry.isIntersecting) { setVisible(true); observer.disconnect(); } }, { threshold: 0.1 });
    observer.observe(el);
    return () => observer.disconnect();
  }, []);

  return (
    <section ref={sectionRef} className="max-w-[1100px] mx-auto px-4 sm:px-6 py-10 sm:py-12">
      <div className="font-['Fira_Code',monospace] text-[10px] tracking-[0.2em] uppercase mb-3" style={{ color: t.accentDim }}>versus</div>
      <h2 className="text-2xl sm:text-3xl font-bold tracking-tight mb-3" style={{ color: t.heading }}>Go TUI libraries</h2>
      <p className="text-[14px] sm:text-[15px] mb-8 sm:mb-10 max-w-[640px]" style={{ color: t.textMuted }}>
        Different trade-offs, side by side.{" "}
        <span className="hidden lg:inline font-['Fira_Code',monospace] text-[11px]" style={{ color: t.textDim }}>Click a row to expand.</span>
      </p>

      <div className="hidden lg:block overflow-x-auto custom-scroll">
        <div className="rounded-lg overflow-hidden" style={{ border: `1px solid ${t.border}`, background: t.bgCard, boxShadow: theme === "dark" ? "0 2px 16px rgba(0,0,0,0.5)" : "0 1px 6px rgba(0,0,0,0.07)" }}>
          <div className="grid items-end gap-0" style={{ gridTemplateColumns: "140px repeat(4, 1fr)", borderBottom: `1px solid ${t.border}`, background: theme === "dark" ? "#23241e" : "#f5f5f1" }}>
            <div className="px-4 py-4" />
            {comparisonLibraries.map((lib, colIdx) => {
              const isGoTui = colIdx === goTuiColIdx;
              return (
                <div key={lib} className="px-4 py-4 text-center" style={{ background: isGoTui ? (theme === "dark" ? `${t.accent}1a` : `${t.accent}10`) : "transparent", borderLeft: `1px solid ${t.border}` }}>
                  <div className="font-['Fira_Code',monospace] text-[12px] font-semibold" style={{ color: isGoTui ? t.accent : t.heading }}>{lib}</div>
                  <div className="mt-1.5 mx-auto h-[2px] rounded-full" style={{ width: isGoTui ? "60%" : "0%", background: libColors[lib], opacity: isGoTui ? 1 : 0 }} />
                </div>
              );
            })}
          </div>
          {comparisonFeatures.map((feature, rowIdx) => {
            const isExpanded = expandedRow === rowIdx;
            const isHovered = hoveredRow === rowIdx;
            const stripeBg = rowIdx % 2 === 0 ? "transparent" : theme === "dark" ? "rgba(255,255,255,0.012)" : "rgba(0,0,0,0.012)";
            return (
              <div key={feature.label} className={visible ? "comparison-row-animate" : "opacity-0"} style={{ borderBottom: rowIdx < comparisonFeatures.length - 1 ? `1px solid ${t.border}` : "none", animationDelay: visible ? `${rowIdx * 50}ms` : "0ms" }}>
                <div className="grid items-stretch gap-0 cursor-pointer select-none" style={{ gridTemplateColumns: "140px repeat(4, 1fr)", background: isExpanded ? (theme === "dark" ? "rgba(255,255,255,0.03)" : "rgba(0,0,0,0.02)") : isHovered ? (theme === "dark" ? "rgba(255,255,255,0.02)" : "rgba(0,0,0,0.015)") : stripeBg, transition: "background 0.15s ease" }} onClick={() => setExpandedRow(isExpanded ? null : rowIdx)} onMouseEnter={() => setHoveredRow(rowIdx)} onMouseLeave={() => setHoveredRow(null)}>
                  <div className="px-4 py-3.5 flex items-center gap-2">
                    <svg width="8" height="8" viewBox="0 0 8 8" className="shrink-0" style={{ transform: isExpanded ? "rotate(90deg)" : "rotate(0deg)", transition: "transform 0.2s cubic-bezier(0.4, 0, 0.2, 1)", opacity: isHovered || isExpanded ? 0.8 : 0.3 }}>
                      <path d="M2 1L6 4L2 7" fill="none" stroke={t.textMuted} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
                    </svg>
                    <div className="font-['Fira_Code',monospace] text-[11px] font-semibold" style={{ color: t.text }}>{feature.label}</div>
                  </div>
                  {comparisonLibraries.map((lib, colIdx) => {
                    const isGoTui = colIdx === goTuiColIdx;
                    return (
                      <div key={lib} className="px-4 py-3.5" style={{ borderLeft: isGoTui ? `1px solid ${t.accent}30` : `1px solid ${t.border}`, borderRight: isGoTui && colIdx < comparisonLibraries.length - 1 ? `1px solid ${t.accent}30` : undefined, background: isGoTui ? accentTint : "transparent" }}>
                        <span className="font-['Fira_Code',monospace] text-[11px] leading-snug" style={{ color: isGoTui ? t.accent : t.text }}>{feature.values[lib].summary}</span>
                      </div>
                    );
                  })}
                </div>
                <ComparisonDetailPanel feature={feature} expanded={isExpanded} />
              </div>
            );
          })}
        </div>
      </div>

      <div className="lg:hidden flex flex-col gap-4">
        {comparisonLibraries.map((lib, libIdx) => {
          const isGoTui = libIdx === goTuiColIdx;
          const color = libColors[lib];
          return (
            <div key={lib} className={`rounded-lg overflow-hidden ${visible ? "comparison-row-animate" : "opacity-0"}`} style={{ border: `1px solid ${isGoTui ? `${t.accent}55` : t.border}`, background: t.bgCard, animationDelay: visible ? `${libIdx * 60}ms` : "0ms" }}>
              <div className="px-3 sm:px-4 py-2.5 sm:py-3 flex items-center gap-2.5" style={{ borderBottom: `1px solid ${isGoTui ? `${t.accent}55` : t.border}`, background: isGoTui ? (theme === "dark" ? `${t.accent}0a` : `${t.accent}08`) : theme === "dark" ? "#23241e" : "#f5f5f1" }}>
                <div className="w-2 h-2 rounded-full shrink-0" style={{ background: color }} />
                <div className="font-['Fira_Code',monospace] text-[12px] sm:text-[13px] font-semibold" style={{ color: isGoTui ? t.accent : t.heading }}>{lib}</div>
              </div>
              <div className="px-3 sm:px-4 py-1.5 sm:py-2">
                {comparisonFeatures.map((feature, fIdx) => (
                  <div key={feature.label} className="py-2" style={{ borderBottom: fIdx < comparisonFeatures.length - 1 ? `1px solid ${t.border}33` : "none" }}>
                    <div className="flex flex-col sm:flex-row sm:items-baseline gap-0.5 sm:gap-2 mb-0.5">
                      <div className="font-['Fira_Code',monospace] text-[10px] uppercase tracking-wider shrink-0" style={{ color: t.textDim }}>{feature.label}</div>
                      <div className="font-['Fira_Code',monospace] text-[11px]" style={{ color: t.text }}>{feature.values[lib].summary}</div>
                    </div>
                    <div className="font-['IBM_Plex_Sans',sans-serif] text-[11px] leading-relaxed" style={{ color: t.textDim }}>{feature.values[lib].detail}</div>
                  </div>
                ))}
              </div>
            </div>
          );
        })}
      </div>
    </section>
  );
}

/* ─── Editor Simulation ─── */

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
        boxShadow: theme === "dark" ? "0 4px 24px rgba(0,0,0,0.4)" : "0 2px 12px rgba(0,0,0,0.08)",
      }}
      onMouseEnter={() => { pausedRef.current = true; }}
      onMouseLeave={() => { pausedRef.current = false; }}
    >
      {/* Editor title bar */}
      <div className="flex items-center justify-between px-4 py-2" style={{ borderBottom: `1px solid ${t.border}` }}>
        <div className="flex items-center gap-2">
          <div className="flex gap-1.5">
            <div className="w-2.5 h-2.5 rounded-full" style={{ background: "#ff5f57" }} />
            <div className="w-2.5 h-2.5 rounded-full" style={{ background: "#febc2e" }} />
            <div className="w-2.5 h-2.5 rounded-full" style={{ background: "#28c840" }} />
          </div>
          <span className="font-['Fira_Code',monospace] text-[10px] ml-2" style={{ color: t.textDim }}>dashboard.gsx</span>
        </div>
        <div className="flex items-center gap-1.5">
          <span className="font-['Fira_Code',monospace] text-[9px] px-1.5 py-0.5 rounded" style={{ color: t.secondary, background: `${t.secondary}12`, border: `1px solid ${t.secondary}25` }}>LSP</span>
          <span className="font-['Fira_Code',monospace] text-[9px] px-1.5 py-0.5 rounded" style={{ color: t.accent, background: `${t.accent}12`, border: `1px solid ${t.accent}25` }}>tree-sitter</span>
        </div>
      </div>

      {/* Feature selector tabs */}
      <div className="flex items-center gap-1 px-3 py-1.5 overflow-x-auto custom-scroll" style={{ borderBottom: `1px solid ${t.border}`, background: theme === "dark" ? "#1e1f1a" : "#eeeee8" }}>
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
                background: isGotoTarget ? `${features[3].color}10` : hasDiagnostic ? `${t.tertiary}08` : isFmtChanged ? `${fmtColor}08` : "transparent",
                borderLeft: isGotoTarget ? `2px solid ${features[3].color}` : isFmtChanged ? `2px solid ${fmtColor}` : "2px solid transparent",
              }}
            >
              <span className="inline-block w-8 sm:w-10 text-right pr-3 sm:pr-4 select-none shrink-0" style={{ color: t.textDim, opacity: 0.5 }}>{line.num}</span>
              <span className="whitespace-pre">
                {line.tokens.map((tok, j) => {
                  const isSyntaxHighlighted = activeFeature === 0 && tok.text.trim().length > 0 && tok.color !== t.text && tok.color !== t.codePunct;
                  return (
                    <span
                      key={j}
                      className={isSyntaxHighlighted ? "syntax-active-token" : ""}
                      style={{ color: tok.color, animationDelay: isSyntaxHighlighted ? `${j * 120}ms` : undefined }}
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
              boxShadow: theme === "dark" ? "0 4px 16px rgba(0,0,0,0.5)" : "0 4px 16px rgba(0,0,0,0.12)",
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
                <span className="px-1 py-0.5 rounded text-[9px]" style={{ background: `${t.secondary}18`, color: t.secondary }}>C</span>
                <span style={{ color: i === 0 ? t.accent : t.text }}>{item.label}</span>
                <span className="ml-auto" style={{ color: t.textDim, fontSize: "10px" }}>{item.detail}</span>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Status bar */}
      <div className="flex items-center justify-between px-4 py-1.5 font-['Fira_Code',monospace] text-[10px]" style={{ borderTop: `1px solid ${t.border}`, background: theme === "dark" ? "#1e1f1a" : "#eeeee8" }}>
        <div className="flex items-center gap-3">
          <span style={{ color: t.textDim }}>Ln 10, Col 5</span>
          <span style={{ color: t.textDim }}>GSX</span>
        </div>
        <div className="flex items-center gap-2">
          {activeFeature === 2 && showDiagnostic && (
            <span className="flex items-center gap-1" style={{ color: t.tertiary }}>
              <svg width="10" height="10" viewBox="0 0 10 10" fill="currentColor"><circle cx="5" cy="5" r="4" /></svg>
              1 error
            </span>
          )}
          {activeFeature === 4 && showFormatted && (
            <span className="flex items-center gap-1" style={{ color: fmtColor }}>6 lines formatted</span>
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
      className="py-2.5 sm:py-3 px-3 sm:px-4 rounded-lg transition-all duration-200 cursor-default"
      style={{
        background: highlighted ? `${color}06` : "transparent",
        borderLeft: `2px solid ${highlighted ? color : "transparent"}`,
        animation: `fadeInUp 0.4s ease-out ${delay}ms both`,
      }}
      onMouseEnter={onHover}
      onMouseLeave={onLeave}
    >
      <div className="flex items-center gap-2 mb-0.5 sm:mb-1">
        <div className="w-1.5 h-1.5 rounded-full shrink-0" style={{ background: color }} />
        <div className="font-['Fira_Code',monospace] text-[12px] sm:text-[13px] font-medium" style={{ color: t.heading }}>{title}</div>
      </div>
      <div className="text-[12px] sm:text-[13px] leading-relaxed pl-3.5" style={{ color: t.textMuted }}>{description}</div>
    </div>
  );
}

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
          <ComparisonSection />
          <Divider />
          <ToolingSection />
        </div>
      </div>
    </>
  );
}
