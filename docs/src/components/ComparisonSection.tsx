import { useState, useEffect, useRef } from "react";
import { palette, useTheme } from "../lib/theme.ts";

type CellValue = {
  summary: string;
  detail: string;
};

export type ComparisonFeature = {
  label: string;
  values: Record<string, CellValue>;
};

export const comparisonLibraries = ["go-tui", "Bubble Tea", "tview", "gocui"] as const;

export const defaultComparisonFeatures: ComparisonFeature[] = [
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
        detail: "OOP style. Create widget objects, configure via methods, compose in layout containers. Implements the Primitive interface",
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
        detail: "lipgloss provides box model styling (padding, margin, borders) and JoinHorizontal/JoinVertical for composition. No flexbox (open issue since 2023)",
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
        detail: "No pre-built widgets. Views provide text I/O and keybindings, but widgets like tables or lists must be built from scratch",
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
        detail: "Inline is the default. Fullscreen requires opting in with tea.WithAltScreen(). Supports tea.Println() for output above",
      },
      tview: {
        summary: "Fullscreen only",
        detail: "Architectural limitation: tcell takes over the entire screen. No inline rendering support",
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
        detail: "All code is plain Go with full gopls support out of the box. No additional tooling needed or available",
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
          {/* Label cell spacer */}
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

export default function ComparisonSection({ features }: { features?: ComparisonFeature[] }) {
  const compFeatures = features ?? defaultComparisonFeatures;
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
        versus
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
        Different trade-offs, side by side.{" "}
        <span
          className="hidden lg:inline font-['Fira_Code',monospace] text-[11px]"
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
          {compFeatures.map((feature, rowIdx) => {
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
                    rowIdx < compFeatures.length - 1
                      ? `1px solid ${t.border}`
                      : "none",
                  animationDelay: visible ? `${rowIdx * 50}ms` : "0ms",
                }}
              >
                {/* Summary row */}
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

                {/* Detail panel */}
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
                {compFeatures.map((feature, fIdx) => {
                  const val = feature.values[lib];
                  return (
                    <div
                      key={feature.label}
                      className="py-2.5"
                      style={{
                        borderBottom:
                          fIdx < compFeatures.length - 1
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
